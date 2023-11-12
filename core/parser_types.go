package core

type ASTNode interface {
	Eval(scope *InterpreterScope) int
	Codegen(asm *string, symbol string, scope *CompilerScope)
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

type InterpreterScope struct {
	inner map[string]ASTNode
	outer *InterpreterScope
}

type CompilerScope struct {
	inner map[string]string
	outer *CompilerScope
}

type Parser struct {
	input        string
	currentIndex int
	currentChar  byte
	AST          *ASTNode
}
