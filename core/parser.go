package core

import (
	"fmt"
	"strconv"
	"unicode"
)

// Lookup global variables
var builtInOperations = []string{"+", "-", "*", "/", "%", "<", ">", "=", "&", "sys_write"}
var operandFunctioanMap = map[string]string{
	"+": "add",
	"-": "sub",
	"*": "mul",
	"/": "udiv",
	"%": "urem",
}
var whiteSpaceChars = []rune{'\n', '\r', '\t', ' '}

// global variables
var generateNextSymbol = nextSymbolGenerator()
var generateNextIfLabel = ifLabelGenerator()
var basicBlockQueue = []string{}
var globalFunctionStore = &FunctionStore{store: make(map[string]*FunctionNode)}

func Includes[T comparable](arr []T, target T) bool {
	for _, elem := range arr {
		if elem == target {
			return true
		}
	}
	return false
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

func NewParser(input string) *Parser {
	return &Parser{
		input:        input,
		currentIndex: 0,
		currentChar:  input[0],
	}
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

func NewInterpreterScope(outer *InterpreterScope) *InterpreterScope {
	return &InterpreterScope{inner: make(map[string]ASTNode), outer: outer}
}

func NewCompilerScope(outer *CompilerScope) *CompilerScope {
	return &CompilerScope{inner: make(map[string]string), outer: outer}
}

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
					if p.currentChar == '(' || unicode.IsDigit(rune(p.currentChar)) || unicode.IsLetter(rune(p.currentChar)) {
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
