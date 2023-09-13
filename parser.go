package main

import (
	"fmt"
	"strconv"
	"unicode"
)

type ASTNode interface {
	eval(scope *Scope) int
	codegen(asm *string, symbol string, scope *Scope)
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
	body      ASTNode
	scope     *Scope
}

type IdentifierNode struct {
	name string
}

type IfNode struct {
	condition *SExpr
	trueExpr  ASTNode
	falseExpr ASTNode
}

var builtInOperations = []string{"+", "-", "*", "/", "<", ">"}

func Includes[T comparable](arr []T, target T) bool {
	for _, elem := range arr {
		if elem == target {
			return true
		}
	}
	return false
}

type Scope struct {
	inner map[string]ASTNode
	outer *Scope
}

func (i *IntegerNode) eval(scope *Scope) int {
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

func evalBuiltin(operand string, arguments []ASTNode, scope *Scope) int {
	evaluatedArgs := make([]int, 0)
	for _, arg := range arguments {
		evaluatedArg := arg.eval(scope)
		evaluatedArgs = append(evaluatedArgs, evaluatedArg)
	}
	return BuiltinFuncMap[operand](evaluatedArgs)
}

func applyFunction(function *FunctionNode, functionEnv *Scope) int {
	return function.body.eval(functionEnv)
}

func (s *SExpr) eval(scope *Scope) int {
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
	extendedEnv := newScope(scope)
	for indx := range s.arguments {
		extendedEnv.inner[function.arguments[indx]] = s.arguments[indx]
	}

	return applyFunction(function, extendedEnv)
}

func (f *FunctionNode) eval(scope *Scope) int {
	scope.inner[f.name] = f
	if f.name == "main" {
		return f.body.eval(scope)
	}
	return 0
}

func (s *Scope) get(variable string) ASTNode {
	if s.inner[variable] == nil {
		if s.outer == nil {
			return nil
		}
		return s.outer.get(variable)
	}
	return s.inner[variable]
}

func (i *IdentifierNode) eval(scope *Scope) int {
	if scope.get(i.name) == nil {
		panic(fmt.Sprintf("Compiler error: %s not found in scope", i.name))
	}
	return scope.get(i.name).eval(scope)
}

func (i *IntegerNode) codegen(asm *string, symbol string, scope *Scope) {
	*asm += fmt.Sprintf(`
	%s = add i32 %d,0
	`, symbol, i.value)
}

var (
	comparisionOps = []string{"<", ">"}
	arithmeticOps  = []string{"+", "-", "*", "/"}
)

func (i *IfNode) eval(scope *Scope) int {
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
		isCondTrue = i.condition.arguments[0].eval(scope) < i.condition.arguments[1].eval(scope)
		break
	default:
		panic("Error")
	}
	if isCondTrue {
		return i.trueExpr.eval(scope)
	} else {
		return i.falseExpr.eval(scope)
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

var generateNextSymbol = nextSymbolGenerator()

var operandFunctioanMap = map[string]string{
	"+": "add",
	"-": "sub",
	"*": "mul",
	"/": "udiv",
}

func (s *SExpr) codegen(asm *string, symbol string, scope *Scope) {
	if Includes(arithmeticOps, s.operand) {
		if len(s.arguments) == 1 {
			nextSymbol := generateNextSymbol()
			s.arguments[0].codegen(asm, nextSymbol, scope)
			*asm += fmt.Sprintf(`
	%s = add i32 0,%s
			`, symbol, nextSymbol)
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
	%s = %s i32 %s,%s
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
		default:
			panic("Error")
		}
		arg1Symbol := generateNextSymbol()
		arg2Symbol := generateNextSymbol()
		s.arguments[0].codegen(asm, arg1Symbol, scope)
		s.arguments[1].codegen(asm, arg2Symbol, scope)
		*asm += fmt.Sprintf(`
	%s = icmp %s i32 %s , %s
		`, symbol, compareInstr, arg1Symbol, arg2Symbol)
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
	fmt.Printf("%v", argumentStack)
	function, ok := scope.get(s.operand).(*FunctionNode)
	if !ok {
		panic("Give function node pls")
	}
	argumentString := "("
	for indx, arg := range argumentStack {
		argumentString += fmt.Sprintf("i32 %s", arg)
		if indx != len(function.arguments)-1 {
			argumentString += ","
		}
	}
	argumentString += ")"
	*asm += fmt.Sprintf(`
	%s = call i32 @%s%s
	`, currentSymbol, s.operand, argumentString)
}

func (f *FunctionNode) codegen(asm *string, symbol string, scope *Scope) {
	scope.inner[f.name] = f
	argumentString := "("
	generateNextSymbol = nextSymbolGenerator()
	symbol = generateNextSymbol()
	for indx, arg := range f.arguments {
		argumentString += ("i32 %" + arg)
		if indx != len(f.arguments)-1 {
			argumentString += ","
		}
	}
	argumentString += ")"
	loadArgumentInstructions := ""
	for _, arg := range f.arguments {
		loadArgumentInstructions += (fmt.Sprintf(`
	%s = add i32 `, symbol) + "%" + arg + ",0")
		symbol = generateNextSymbol()
	}
	*asm += fmt.Sprintf(`
define i32 @%s%s{
	`, f.name, argumentString)
	*asm += fmt.Sprintf(` 
	%s
	`, loadArgumentInstructions)
	f.body.codegen(asm, symbol, scope)
	*asm += fmt.Sprintf(`
	ret i32 %%sym%d
}
	`, len(f.arguments)+1)
}

func (i *IdentifierNode) codegen(asm *string, symbol string, scope *Scope) {
	*asm += fmt.Sprintf(`%s = add i32 %%`, symbol) + fmt.Sprintf(`%s,0
	`, i.name)
}

func (i *IfNode) codegen(asm *string, symbol string, scope *Scope) {
	allocVariable := generateNextSymbol()
	*asm += fmt.Sprintf(`
	%s = alloca i32, align 4
	`, allocVariable)
	conditionSymbol := generateNextSymbol()
	i.condition.codegen(asm, conditionSymbol, scope)
	*asm += fmt.Sprintf(`
	br i1 %s,label %%iftrue,label %%iffalse
	iftrue:
	`, conditionSymbol)
	trueSymbol := generateNextSymbol()
	falseSymbol := generateNextSymbol()
	i.trueExpr.codegen(asm, trueSymbol, scope)
	*asm += fmt.Sprintf(`
		store i32 %s, i32* %s, align 4
		br label %%ifresult
	`, trueSymbol, allocVariable)
	*asm += fmt.Sprintf(`
	iffalse:
	`)
	i.falseExpr.codegen(asm, falseSymbol, scope)
	*asm += fmt.Sprintf(`
		store i32 %s, i32* %s, align 4
		br label %%ifresult
	`, falseSymbol, allocVariable)
	*asm += fmt.Sprintf(`
	ifresult:
		%s = load i32,i32* %s, align 4
	`, symbol, allocVariable)
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
	for Includes(whiteSpaceChars, rune(p.currentChar)) && p.currentIndex < len(p.input) {
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

func (p *Parser) parseFunctionBody() ASTNode {
	peekedChar, ok := p.peekChar()
	if !ok {
		return nil
	}
	for (peekedChar != '(' && p.currentChar == ' ') || (p.currentChar == ')') {
		p.nextChar()
		peekedChar, ok = p.peekChar()
		if !ok {
			return nil
		}
	}
	peekedChar, ok = p.peekChar()
	if !ok {
		return nil
	}
	p.skipWhitespace()
	var expr ASTNode
	for peekedChar == '(' {
		switch parsed := p.ParseExpression().(type) {
		case *SExpr:
			{
				expr = parsed
				return parsed
			}
		case *IfNode:
			{
				expr = parsed
				return parsed
			}
		default:
			return nil
		}
	}
	return expr
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
		scope:     &Scope{inner: make(map[string]ASTNode), outer: nil},
	}
}

func newIdentifierNode(name string) *IdentifierNode {
	return &IdentifierNode{
		name: name,
	}
}

func newScope(outer *Scope) *Scope {
	return &Scope{inner: make(map[string]ASTNode), outer: outer}
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
				'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '_', '+', '-', '*', '/', '<', '>':
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
					falseExpr := p.ParseExpression()
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
				panic("Should be a ' ' here")
			}
			p.skipWhitespace()
			return newIdentifierNode(ident)
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
		if p.currentChar != ')' {
			panic(fmt.Sprintf("Expected ) got %s", string(p.currentChar)))
		}
		p.nextChar()
	}
	return astNodeArray
}
