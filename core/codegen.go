package core

import (
	"fmt"
	"runtime"
)

var prefix = `
.global _start
.align 2

plus:
	add X2,X0,X1
	str X2,[sp,#-16]!
	ret

minus:
	sub X2,X0,X1
	str X2,[sp,#-16]!
	ret

multiply:
	mul X2,X0,X1
	str X2,[sp,#-16]!
	ret

divide:
	udiv X2,X0,X1
	str X2,[sp,#-16]!
	ret
_start:

`

var suffix = `
	mov X0,X2
	mov X16,#1
	svc #0x80
`

func (s *SExpr) Codegen(asm *string, symbol string, scope *CompilerScope) {
	if Includes(arithmeticOps, s.operand) {
		if len(s.arguments) == 1 {
			s.arguments[0].Codegen(asm, symbol, scope)
			return
		}
		midpoint := getMidpointIndex[ASTNode](s.arguments)
		firstHalf := s.arguments[:midpoint]
		firstHalfSymbol := generateNextSymbol()
		secondHalfSymbol := generateNextSymbol()
		firstHalfSExpr := &SExpr{
			operand:   s.operand,
			arguments: firstHalf,
		}
		firstHalfSExpr.Codegen(asm, firstHalfSymbol, scope)
		secondHalf := s.arguments[midpoint:]
		secondHalfSExpr := &SExpr{
			operand:   s.operand,
			arguments: secondHalf,
		}
		secondHalfSExpr.Codegen(asm, secondHalfSymbol, scope)
		*asm += fmt.Sprintf(`
	%s = %s i64 %s,%s
		`, symbol, operandFunctioanMap[s.operand], firstHalfSymbol, secondHalfSymbol)
		return
	}
	if Includes(comparisionOps, s.operand) {
		if len(s.arguments) != 2 {
			panic("Error: comparision operators can have only two arguments")
		}
		var compareInstr string
		switch s.operand {
		case "<":
			compareInstr = "slt"
			break
		case ">":
			compareInstr = "sgt"
			break
		case "=":
			compareInstr = "eq"
			break
		default:
			panic("Error")
		}
		arg1Symbol := generateNextSymbol()
		arg2Symbol := generateNextSymbol()
		s.arguments[0].Codegen(asm, arg1Symbol, scope)
		s.arguments[1].Codegen(asm, arg2Symbol, scope)
		*asm += fmt.Sprintf(`
	%s = icmp %s i64 %s , %s
		`, symbol, compareInstr, arg1Symbol, arg2Symbol)
		return
	}
	if Includes(systemCalls, s.operand) {
		outFd, ok := s.arguments[0].(*IntegerNode)
		outFdSymbol := generateNextSymbol()
		if !ok {
			panic("Expected Integer")
		}
		referenceNode, ok := s.arguments[1].(*ReferenceNode)
		referenceSymbol := generateNextSymbol()
		if !ok {
			panic("Expected Reference")
		}
		charNum, ok := s.arguments[2].(*IntegerNode)
		charNumSymbol := generateNextSymbol()
		outFd.Codegen(asm, outFdSymbol, scope)
		referenceNode.Codegen(asm, referenceSymbol, scope)
		charNum.Codegen(asm, charNumSymbol, scope)
		if !ok {
			panic("Expected Integer")
		}
		pointerToIntSymbol := generateNextSymbol()
		syscallNumSymbol := generateNextSymbol()
		syscallStatusSymbol := generateNextSymbol()
		checkIfSyscallSuccessSymbol := generateNextSymbol()
		symbolForOne := generateNextSymbol()
		if runtime.GOOS == "darwin" {
			*asm += fmt.Sprintf(`
				%s = ptrtoint i64* %s to i64
				%s = add i64 4,0
				%s = call i64 asm sideeffect "svc #0x80","=r,{x0},{x1},{x2},{x16}" (i64 %s,i64 %s,i64 %s,i64 %s)
				%s = add i64 1,0
				%s = icmp eq i64 %s,%s
				br i1 %s,label %%syscallSuccess,label %%syscallFail
				syscallSuccess:
					ret i64 0
				syscallFail:
					%s = add i64 %s,0
			`, pointerToIntSymbol, referenceSymbol, syscallNumSymbol, syscallStatusSymbol, outFdSymbol, pointerToIntSymbol, charNumSymbol, syscallNumSymbol, symbolForOne, checkIfSyscallSuccessSymbol, symbolForOne, syscallStatusSymbol, checkIfSyscallSuccessSymbol, symbol, syscallStatusSymbol)
		} else {
			*asm += fmt.Sprintf(`
				%s = ptrtoint i64* %s to i64
				%s = add i64 1,0
				%s = call i64 asm sideeffect "syscall","=r,{rax},{rdi},{rsi},{rdx}" (i64 %s,i64 %s,i64 %s,i64 %s)
				%s = add i64 1,0
				%s = icmp eq i64 %s,%s
				br i1 %s,label %%syscallSuccess,label %%syscallFail
				syscallSuccess:
					ret i64 0
				syscallFail:
					%s = add i64 %s,0
			`, pointerToIntSymbol, referenceSymbol, syscallNumSymbol, syscallStatusSymbol, syscallNumSymbol, outFdSymbol, pointerToIntSymbol, charNumSymbol, symbolForOne, checkIfSyscallSuccessSymbol, symbolForOne, syscallStatusSymbol, checkIfSyscallSuccessSymbol, symbol, syscallStatusSymbol)
		}
		return
	}
	currentSymbol := symbol
	symbol = generateNextSymbol()
	argumentStack := make([]string, 0)
	for _, arg := range s.arguments {
		argumentStack = append(argumentStack, symbol)
		arg.Codegen(asm, symbol, scope)
		symbol = generateNextSymbol()
	}
	function, ok := globalFunctionStore.store[s.operand]
	if !ok {
		panic(fmt.Sprintf("%s function not defined", s.operand))
	}
	argumentString := "("
	for indx, arg := range argumentStack {
		argumentString += fmt.Sprintf("i64 %s", arg)
		if indx != len(function.arguments)-1 {
			argumentString += ","
		}
	}
	argumentString += ")"
	*asm += fmt.Sprintf(`
	%s = call i64 @%s%s
	`, currentSymbol, s.operand, argumentString)
}

func (f *FunctionNode) Codegen(asm *string, symbol string, scope *CompilerScope) {
	globalFunctionStore.store[f.name] = f
	scope.inner[f.name] = f.name
	argumentString := "("
	generateNextIfLabel = ifLabelGenerator()
	generateNextSymbol = nextSymbolGenerator()
	symbol = generateNextSymbol()
	for indx, arg := range f.arguments {
		argumentString += ("i64 %" + arg)
		if indx != len(f.arguments)-1 {
			argumentString += ","
		}
	}
	argumentString += ")"
	loadArgumentInstructions := ""
	for _, arg := range f.arguments {
		loadArgumentInstructions += fmt.Sprintf(`
  %s = alloca i64, align 4
	store i64 %%%s, i64* %s, align 4
    `, symbol, arg, symbol)
		scope.inner[arg] = symbol
		symbol = generateNextSymbol()
	}
	*asm += fmt.Sprintf(`
define i64 @%s%s{
    entry:
	`, f.name, argumentString)
	*asm += fmt.Sprintf(` 
	%s
	`, loadArgumentInstructions)
	for i, expr := range f.body {
		var symbolForExpression string // symbol for each expression in the function body, the last statement should use the main symbol(cause thats what gets returned) and the subsidiaries should use a new symbol
		if i == len(f.body)-1 {
			symbolForExpression = symbol
		} else {
			symbolForExpression = generateNextSymbol()
		}
		expr.Codegen(asm, symbolForExpression, scope)
	}
	*asm += fmt.Sprintf(`
	ret i64 %%sym%d
}
	`, len(f.arguments)+1)
}

func (i *IdentifierNode) Codegen(asm *string, symbol string, scope *CompilerScope) {
	compilerSymbol, err := scope.get(i.name)
	if err != nil {
		panic(fmt.Sprintf("Symbol not in scope %s", i.name))
	}
	*asm += fmt.Sprintf(`%s = load i64 ,i64* %s,align 4`, symbol, compilerSymbol)
}

func (i *IfNode) Codegen(asm *string, symbol string, scope *CompilerScope) {
	allocVariable := generateNextSymbol()
	*asm += fmt.Sprintf(`
	%s = alloca i64, align 4
	`, allocVariable)

	conditionSymbol := generateNextSymbol()
	ifLabel := generateNextIfLabel()

	i.condition.Codegen(asm, conditionSymbol, scope)

	if i.falseExpr != nil {
		*asm += fmt.Sprintf(`
	br i1 %s,label %%%s,label %%%s
	%s:
	`, conditionSymbol, ifLabel[0], ifLabel[1], ifLabel[0])
	} else {
		*asm += fmt.Sprintf(`
	br i1 %s,label %%%s,label %%%s
	%s:
	`, conditionSymbol, ifLabel[0], ifLabel[2], ifLabel[0])
	}
	basicBlockQueue = append(basicBlockQueue, ifLabel[0])

	trueSymbol := generateNextSymbol()
	falseSymbol := generateNextSymbol()

	i.trueExpr.Codegen(asm, trueSymbol, scope)

	*asm += fmt.Sprintf(`
    br label %%%s
	`, ifLabel[2])

	if i.falseExpr != nil {
		*asm += fmt.Sprintf(`
	%s:
	`, ifLabel[1])
		i.falseExpr.Codegen(asm, falseSymbol, scope)
		*asm += fmt.Sprintf(`
      br label %%%s
  `, ifLabel[2])
		basicBlockQueue = append(basicBlockQueue, ifLabel[1])
	}

	if i.falseExpr == nil {

		if len(basicBlockQueue) == 0 {
			panic("block queue cannot be empty")
		}

		phiTrueLabel := basicBlockQueue[len(basicBlockQueue)-1]
		basicBlockQueue = basicBlockQueue[:len(basicBlockQueue)-1]
		var phiFalseLabel string
		if len(basicBlockQueue) == 0 {
			phiFalseLabel = "entry"
		} else {

			phiFalseLabel = basicBlockQueue[len(basicBlockQueue)-1]
			basicBlockQueue = basicBlockQueue[:len(basicBlockQueue)-1]
		}
		*asm += fmt.Sprintf(`
	%s:
    %s = phi i64 [%s,%%%s],[0,%%%s]
	`, ifLabel[2], symbol, trueSymbol, phiTrueLabel, phiFalseLabel)
	} else {
		if len(basicBlockQueue) == 0 {
			panic("block queue cannot be empty")
		}

		phiFalseLabel := basicBlockQueue[len(basicBlockQueue)-1]
		basicBlockQueue = basicBlockQueue[:len(basicBlockQueue)-1]
		if len(basicBlockQueue) == 0 {
			panic("block queue cannot be empty")
		}
		phiTrueLabel := basicBlockQueue[len(basicBlockQueue)-1]
		basicBlockQueue = basicBlockQueue[:len(basicBlockQueue)-1]

		*asm += fmt.Sprintf(`
	%s:
    %s = phi i64 [%s,%%%s],[%s,%%%s]
	`, ifLabel[2], symbol, trueSymbol, phiTrueLabel, falseSymbol, phiFalseLabel)
	}
	basicBlockQueue = append(basicBlockQueue, ifLabel[2])
}

func (r *ReferenceNode) Codegen(asm *string, symbol string, scope *CompilerScope) {
	valueSymbol := generateNextSymbol()
	r.value.Codegen(asm, valueSymbol, scope)
	*asm += fmt.Sprintf(`
	%s = alloca i64, align 4
	store i64 %s,i64* %s,align 4
	`, symbol, valueSymbol, symbol)
}

func (i *IntegerNode) Codegen(asm *string, symbol string, scope *CompilerScope) {
	*asm += fmt.Sprintf(`
	%s = add i64 %d,0
	`, symbol, i.value)
}
