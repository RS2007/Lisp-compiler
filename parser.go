package main

import (
	"fmt"
	"strconv"
)

type ASTNode interface {
	eval() int
}

type IntegerNode struct {
	value int
}

type SExpr struct {
	operand string
	left    ASTNode
	right   ASTNode
}

func (i *IntegerNode) eval() int {
	return i.value
}

func (s *SExpr) eval() int {
	switch s.operand {
	case "+":
		return s.left.eval() + s.right.eval()
	case "-":
		return s.left.eval() - s.right.eval()
	case "*":
		return s.left.eval() * s.right.eval()
	case "/":
		return s.left.eval() / s.right.eval()
	default:
		panic("Should not hit here")
	}
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
		currentIndex: -1,
		currentChar:  0,
	}
}

func (p *Parser) nextChar() bool {
	p.currentIndex += 1
	if p.currentIndex >= len(p.input) {
		return false
	}
	for p.input[p.currentIndex] == ' ' {
		p.currentIndex += 1
	}
	if p.currentIndex >= len(p.input) {
		return false
	}
	p.currentChar = p.input[p.currentIndex]
	return true
}

func newSExpr(operand string) *SExpr {
	return &SExpr{
		operand: operand,
		left:    nil,
		right:   nil,
	}
}

func newIntegerNode(value int) *IntegerNode {
	return &IntegerNode{
		value: value,
	}
}

func (p *Parser) Parse() ASTNode {
	var node ASTNode
	for p.nextChar() {
		switch p.currentChar {
		case '(':
			p.nextChar()
			switch string(p.currentChar) {
			case "+", "-", "*", "/":
				node = newSExpr(string(p.currentChar))
				break
			default:
				panic("Invalid operand " + string(p.currentChar))
			}
			sexpr, ok := node.(*SExpr)
			if !ok {
				panic("Error")
			}
			sexpr.left = p.Parse()
			sexpr.right = p.Parse()
			p.nextChar()
			return sexpr
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			var value, err = strconv.Atoi(string(p.currentChar))
			if err != nil {
				panic(err)
			}
			node = newIntegerNode(value)
			return node

		default:
			fmt.Println(string(p.currentChar))
			panic("Should'nt hit here")
		}
	}
	return node
}
