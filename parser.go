package main

import (
	"fmt"
	"strconv"
	"unicode"
)

type ASTNode interface {
	eval(scope *InterpreterScope) int
	codegen(asm *string, symbol string, scope *CompilerScope)
}

type IntegerNode struct {
	value int
}

type SExpr struct {
	operand   string
	arguments []ASTNode
}

type FunctionNode struct {
	name      string
	arguments []string
	body      []ASTNode
	scope     *InterpreterScope
}

type IdentifierNode struct {
	name string
}

type IfNode struct {
	condition *SExpr
	trueExpr  ASTNode
	falseExpr ASTNode
}

type ReferenceNode struct {
	value ASTNode
}

type FunctionStore struct {
	store map[string]*FunctionNode
}

var globalFunctionStore = &FunctionStore{store: make(map[string]*FunctionNode)}

var builtInOperations = []string{"+", "-", "*", "/", "%", "<", ">", "=", "&", "sys_write"}

func Includes[T comparable](arr []T, target T) bool {
	for _, elem := range arr {
		if elem == target {
			return true
		}
	}
	return false
}

type InterpreterScope struct {
	inner map[string]ASTNode
	outer *InterpreterScope
}

type CompilerScope struct {
	inner map[string]string
	outer *CompilerScope
}

func (i *IntegerNode) eval(scope *InterpreterScope) int {
	return i.value
}

var BuiltinFuncMap = map[string]func([]int) int{
	"+": builtinAdd,
	"-": builtinSub,
	"*": builtinMul,
	"/": builtinDiv,
}

func builtinAdd(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum += number
	}
	return sum
}

func builtinSub(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum -= number
	}
	return sum
}

func builtinMul(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum *= number
	}
	return sum
}

func builtinDiv(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum /= number
	}
	return sum
}

func evalBuiltin(operand string, arguments []ASTNode, scope *InterpreterScope) int {
	evaluatedArgs := make([]int, 0)
	for _, arg := range arguments {
		evaluatedArg := arg.eval(scope)
		evaluatedArgs = append(evaluatedArgs, evaluatedArg)
	}
	return BuiltinFuncMap[operand](evaluatedArgs)
}

func applyFunction(function *FunctionNode, functionEnv *InterpreterScope) int {
	value := 0
	for _, expr := range function.body {
		value = expr.eval(functionEnv)
	}
	return value
}

func (s *SExpr) eval(scope *InterpreterScope) int {
	if Includes(builtInOperations, s.operand) {
		return evalBuiltin(s.operand, s.arguments, scope)
	}
	if scope.get(s.operand) == nil {
		panic(fmt.Sprintf("%s not in scope", s.operand))
	}
	astNode := scope.get(s.operand)
	function, ok := astNode.(*FunctionNode)
	if !ok {
		panic("Expected function node,got some nonsense")
	}
	extendedEnv := newInterpreterScope(scope)
	for indx := range s.arguments {
		extendedEnv.inner[function.arguments[indx]] = &IntegerNode{value: s.arguments[indx].eval(scope)}
	}

	return applyFunction(function, extendedEnv)
}

func (f *FunctionNode) eval(scope *InterpreterScope) int {
	scope.inner[f.name] = f
	value := 0
	if f.name == "main" {
		for _, expr := range f.body {
			value = expr.eval(scope)
		}
	}
	return value
}

func (r *ReferenceNode) eval(scope *InterpreterScope) int {
	panic("Interpreter does not support references")
}

func (s *InterpreterScope) get(variable string) ASTNode {
	if s.inner[variable] == nil {
		if s.outer == nil {
			return nil
		}
		return s.outer.get(variable)
	}
	return s.inner[variable]
}

func (s *CompilerScope) get(variable string) (string, error) {
	inner, innerOk := s.inner[variable]
	if !innerOk {
		if s.outer == nil {
			return "", fmt.Errorf("variable not defined")
		}
		return s.outer.get(variable)
	}
	return inner, nil
}

func (i *IdentifierNode) eval(scope *InterpreterScope) int {
	if scope.get(i.name) == nil {
		panic(fmt.Sprintf("Compiler error: %s not found in scope", i.name))
	}
	return scope.get(i.name).eval(scope)
}

func (i *IntegerNode) codegen(asm *string, symbol string, scope *CompilerScope) {
	*asm += fmt.Sprintf(`
	%s = add i64 %d,0
	`, symbol, i.value)
}

var (
	comparisionOps = []string{"<", ">", "="}
	arithmeticOps  = []string{"+", "-", "*", "/", "%"}
	systemCalls    = []string{"sys_write"}
)

func (i *IfNode) eval(scope *InterpreterScope) int {
	if !Includes(comparisionOps, i.condition.operand) {
		panic("Should have a comparision operator in if condition")
	}
	if len(i.condition.arguments) != 2 {
		panic("Conditional operators are binary")
	}
	var isCondTrue bool
	switch i.condition.operand {
	case "<":
		isCondTrue = i.condition.arguments[0].eval(scope) < i.condition.arguments[1].eval(scope)
		break
	case ">":
		isCondTrue = i.condition.arguments[0].eval(scope) > i.condition.arguments[1].eval(scope)
		break
	case "=":
		isCondTrue = i.condition.arguments[0].eval(scope) == i.condition.arguments[1].eval(scope)
		break
	default:
		panic("Error")
	}
	if isCondTrue {
		return i.trueExpr.eval(scope)
	} else {
		if i.falseExpr != nil {
			return i.falseExpr.eval(scope)
		}
		return 0
	}
}

func getMidpointIndex[T any](arr []T) int {
	if len(arr)%2 == 0 {
		return len(arr) / 2
	} else {
		return (len(arr) + 1) / 2
	}
}

func nextSymbolGenerator() func() string {
	count := 1
	return func() string {
		count += 1
		return fmt.Sprintf("%%sym%d", count-1)
	}
}

func ifLabelGenerator() func() [3]string {
	count := 1
	return func() [3]string {
		count += 1
		return [3]string{fmt.Sprintf("iftrue%d", count-1), fmt.Sprintf("iffalse%d", count-1), fmt.Sprintf("ifresult%d", count-1)}
	}
}

var generateNextSymbol = nextSymbolGenerator()
var generateNextIfLabel = ifLabelGenerator()

var operandFunctioanMap = map[string]string{
	"+": "add",
	"-": "sub",
	"*": "mul",
	"/": "udiv",
	"%": "urem",
}

func (s *SExpr) codegen(asm *string, symbol string, scope *CompilerScope) {
	if Includes(arithmeticOps, s.operand) {
		if len(s.arguments) == 1 {
			s.arguments[0].codegen(asm, symbol, scope)
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
		firstHalfSExpr.codegen(asm, firstHalfSymbol, scope)
		secondHalf := s.arguments[midpoint:]
		secondHalfSExpr := &SExpr{
			operand:   s.operand,
			arguments: secondHalf,
		}
		secondHalfSExpr.codegen(asm, secondHalfSymbol, scope)
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
		s.arguments[0].codegen(asm, arg1Symbol, scope)
		s.arguments[1].codegen(asm, arg2Symbol, scope)
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
		outFd.codegen(asm, outFdSymbol, scope)
		referenceNode.codegen(asm, referenceSymbol, scope)
		charNum.codegen(asm, charNumSymbol, scope)
		if !ok {
			panic("Expected Integer")
		}
		pointerToIntSymbol := generateNextSymbol()
		syscallNumSymbol := generateNextSymbol()
		syscallStatusSymbol := generateNextSymbol()
		checkIfSyscallSuccessSymbol := generateNextSymbol()
		symbolForOne := generateNextSymbol()
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
		return
	}
	currentSymbol := symbol
	symbol = generateNextSymbol()
	argumentStack := make([]string, 0)
	for _, arg := range s.arguments {
		argumentStack = append(argumentStack, symbol)
		arg.codegen(asm, symbol, scope)
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

func (f *FunctionNode) codegen(asm *string, symbol string, scope *CompilerScope) {
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
		expr.codegen(asm, symbolForExpression, scope)
	}
	*asm += fmt.Sprintf(`
	ret i64 %%sym%d
}
	`, len(f.arguments)+1)
}

func (i *IdentifierNode) codegen(asm *string, symbol string, scope *CompilerScope) {
	compilerSymbol, err := scope.get(i.name)
	if err != nil {
		panic(fmt.Sprintf("Symbol not in scope %s", i.name))
	}
	*asm += fmt.Sprintf(`%s = load i64 ,i64* %s,align 4`, symbol, compilerSymbol)
}

func (i *IfNode) codegen(asm *string, symbol string, scope *CompilerScope) {
	allocVariable := generateNextSymbol()
	*asm += fmt.Sprintf(`
	%s = alloca i64, align 4
	`, allocVariable)
	conditionSymbol := generateNextSymbol()
	ifLabel := generateNextIfLabel()
	i.condition.codegen(asm, conditionSymbol, scope)
	*asm += fmt.Sprintf(`
	br i1 %s,label %%%s,label %%%s
	%s:
	`, conditionSymbol, ifLabel[0], ifLabel[1], ifLabel[0])
	trueSymbol := generateNextSymbol()
	falseSymbol := generateNextSymbol()
	i.trueExpr.codegen(asm, trueSymbol, scope)
	*asm += fmt.Sprintf(`
		store i64 %s, i64* %s, align 4
    br label %%%s
	`, trueSymbol, allocVariable, ifLabel[2])
	*asm += fmt.Sprintf(`
	%s:
	`, ifLabel[1])
	if i.falseExpr != nil {
		i.falseExpr.codegen(asm, falseSymbol, scope)
		*asm += fmt.Sprintf(`
		store i64 %s, i64* %s, align 4
	`, falseSymbol, allocVariable)
	}
	*asm += fmt.Sprintf(`
		br label %%%s
	`, ifLabel[2])
	*asm += fmt.Sprintf(`
	%s:
		%s = load i64,i64* %s, align 4
	`, ifLabel[2], symbol, allocVariable)
}

func (r *ReferenceNode) codegen(asm *string, symbol string, scope *CompilerScope) {
	valueSymbol := generateNextSymbol()
	r.value.codegen(asm, valueSymbol, scope)
	*asm += fmt.Sprintf(`
	%s = alloca i64, align 4
	store i64 %s,i64* %s,align 4
	`, symbol, valueSymbol, symbol)
}

type Parser struct {
	input        string
	currentIndex int
	currentChar  byte
	AST          *ASTNode
}

func newParser(input string) *Parser {
	return &Parser{
		input:        input,
		currentIndex: 0,
		currentChar:  input[0],
	}
}

var whiteSpaceChars = []rune{'\n', '\r', '\t', ' '}

func (p *Parser) skipWhitespace() {
	for Includes(whiteSpaceChars, rune(p.currentChar)) && p.currentIndex < len(p.input)-1 {
		p.currentIndex += 1
		p.currentChar = p.input[p.currentIndex]
	}
}

func (p *Parser) isEndOfInput() bool {
	if p.currentIndex >= len(p.input) {
		return true
	}
	return false
}

func (p *Parser) nextChar() bool {
	p.currentIndex += 1
	if p.isEndOfInput() {
		return false
	}
	p.currentChar = p.input[p.currentIndex]
	return true
}

func (p *Parser) peekChar() (byte, bool) {
	if p.currentIndex+1 >= len(p.input) {
		return 0, false
	}
	return p.input[p.currentIndex+1], true
}

func (p *Parser) parseSExprArgs() []ASTNode {
	parsedArgs := make([]ASTNode, 0)
	for p.currentChar != ')' {
		arg := p.ParseExpression()
		parsedArgs = append(parsedArgs, arg)
		p.skipWhitespace()
	}
	return parsedArgs
}

func (p *Parser) parseFunctionBody() []ASTNode {
	bodyExpressions := make([]ASTNode, 0)
	p.nextChar() // To go from the closing parans of the arguments array
	// Can either be at a whitespace or at a (, skip whitespace and assert
	for p.currentChar != ')' {
		p.skipWhitespace()
		if p.currentChar == '(' {
			if p.currentChar != '(' {
				panic(fmt.Sprintf("Expected ( got %s", string(p.currentChar)))
			}
			bodyExpressions = append(bodyExpressions, p.ParseExpression())
			p.skipWhitespace()
		} else {
			bodyExpressions = append(bodyExpressions, p.ParseExpression())
		}
	}
	return bodyExpressions
}

func (p *Parser) parseFunctionArguments() []string {
	argArray := make([]string, 0)
	for p.currentChar != ')' {
		arg := p.readIdentifier()
		argArray = append(argArray, arg)
		p.skipWhitespace()
	}
	return argArray
}

func (p *Parser) parseCallArguments() []ASTNode {
	p.nextChar()
	p.nextChar()
	p.skipWhitespace()
	evaluatedArgs := make([]ASTNode, 0)
	for p.currentChar != ')' {
		arg := p.ParseExpression()
		if p.currentChar != ',' {
			panic(fmt.Sprintf("Expected , got %s", string(p.currentChar)))
		}
		evaluatedArgs = append(evaluatedArgs, arg)
	}
	return evaluatedArgs
}

func newSExpr(operand string) *SExpr {
	return &SExpr{
		operand:   operand,
		arguments: make([]ASTNode, 0),
	}
}

func newIntegerNode(value int) *IntegerNode {
	return &IntegerNode{
		value: value,
	}
}

func newFunctionNode(name string) *FunctionNode {
	return &FunctionNode{
		name:      name,
		arguments: nil,
		body:      nil,
		scope:     &InterpreterScope{inner: make(map[string]ASTNode), outer: nil},
	}
}

func newIdentifierNode(name string) *IdentifierNode {
	return &IdentifierNode{
		name: name,
	}
}

func newReferenceNode(value ASTNode) *ReferenceNode {
	return &ReferenceNode{
		value,
	}
}

func newInterpreterScope(outer *InterpreterScope) *InterpreterScope {
	return &InterpreterScope{inner: make(map[string]ASTNode), outer: outer}
}

func newCompilerScope(outer *CompilerScope) *CompilerScope {
	return &CompilerScope{inner: make(map[string]string), outer: outer}
}

func (p *Parser) readIdentifier() string {
	identifier := ""
	for unicode.IsLower(rune(p.currentChar)) || Includes(builtInOperations, string(p.currentChar)) || string(p.currentChar) == "_" {
		identifier += string(p.currentChar)
		peekedChar, ok := p.peekChar()
		if !ok {
			break
		}
		if peekedChar == ' ' {
			p.nextChar()
			break
		}
		p.nextChar()
	}
	return identifier
}

func (p *Parser) ParseExpression() ASTNode {
	var node ASTNode
	for !p.isEndOfInput() {
		p.skipWhitespace()
		switch p.currentChar {
		case '(':
			p.nextChar()
			p.skipWhitespace()
			// At the end of this you should be at a non-( and non-space character
			// An identifier is now read and then two subsequent s expressions are parsed
			switch p.currentChar {
			case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
				'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '_', '+', '-', '*', '/', '%', '<', '>', '=':
				identifier := p.readIdentifier()
				if identifier == "def" {
					// Implies this is a function defn
					// skip the def keyword, go to the function name add it to the node
					p.skipWhitespace()
					functionNode := newFunctionNode(p.readIdentifier())
					p.skipWhitespace()
					if p.currentChar != '(' {
						// Should start a () pair to store arguments
						panic(fmt.Sprintf("Expected ( got %s", string(p.currentChar)))
					}
					p.nextChar()
					// parse the arguments(a list of identifiers)
					functionNode.arguments = p.parseFunctionArguments()
					// parse the body(parse an s expression)
					functionNode.body = p.parseFunctionBody()
					p.nextChar()
					return functionNode
				}
				if identifier == "if" {
					p.skipWhitespace()
					if p.currentChar != '(' {
						panic(fmt.Sprintf("Expected ( got %s", string(p.currentChar)))
					}
					condition, ok := p.ParseExpression().(*SExpr)
					if !ok {
						panic("Error")
					}
					p.skipWhitespace()
					trueExpr := p.ParseExpression()
					p.skipWhitespace()
					var falseExpr ASTNode
					if p.currentChar == '(' || unicode.IsDigit(rune(p.currentChar)) {
						falseExpr = p.ParseExpression()
					} else {
						falseExpr = nil
					}
					if p.currentChar != ')' {
						panic(fmt.Sprintf("Expected ) got %s", string(p.currentChar)))
					}
					p.nextChar()
					return &IfNode{
						condition,
						trueExpr,
						falseExpr,
					}
				}
				sexpr := newSExpr(identifier)
				p.skipWhitespace()
				sexpr.arguments = p.parseSExprArgs()
				if p.currentChar != ')' {
					panic(fmt.Sprintf("Should be a ) here, got %s", string(p.currentChar)))
				}
				p.nextChar()
				return sexpr
			default:
				panic("Invalid operand " + string(p.currentChar))
			}
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			value, err := strconv.Atoi(string(p.currentChar))
			p.nextChar()
			for unicode.IsNumber(rune(p.currentChar)) {
				newValue, err := strconv.Atoi(string(p.currentChar))
				if err != nil {
					panic(err)
				}
				value = value*10 + newValue
				p.nextChar()
			}
			if err != nil {
				panic(err)
			}
			// Reaches the non numeric character, has to be a space/) can put an asssert here
			if p.currentChar != ' ' && p.currentChar != ')' {
				panic(fmt.Sprintf("Should be a ' ' or ) here got %s", string(p.currentChar)))
			}
			p.skipWhitespace()
			// Should reach an identifier or an s expression or a number
			node = newIntegerNode(value)
			return node
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '_':
			ident := p.readIdentifier()
			if p.currentChar != ' ' && p.currentChar != ')' {
				panic("Should be a ' ' or ) here")
			}
			p.skipWhitespace()
			return newIdentifierNode(ident)
		case '&':
			p.nextChar()
			p.skipWhitespace()
			return newReferenceNode(p.ParseExpression())
		default:
			fmt.Println(string(p.currentChar))
			panic("Should'nt hit here")
		}
	}
	return node
}

func (p *Parser) Parse() []ASTNode {
	astNodeArray := make([]ASTNode, 0)
	for !p.isEndOfInput() {
		astNodeArray = append(astNodeArray, p.ParseExpression())
		p.skipWhitespace()
	}
	return astNodeArray
}
