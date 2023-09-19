package main

import (
	"fmt"
	"sort"
	"testing"
)

func TestFunctionEval(t *testing.T) {
	type TestCase struct {
		input     string
		evaluated int
	}
	inputs := []TestCase{
		{input: "(def plus-two (a b) (+ a (+ b 2))) (def main () (plus-two 3 (plus-two 1 1)) 0)", evaluated: 0},
		{input: "(def plus_two(a) (+ a 2)) (def main() (plus_two 3) )", evaluated: 5},
		{input: "(def add_two(a b) (+ a (+ b 2))) (def main() (add_two 1 2))", evaluated: 5},
		{input: "(def main() (if (< 3 2) 1 0))", evaluated: 0},
		{input: "(def is_small (x) (if (< x 5) 1 0))(def main() (is_small 3)", evaluated: 1},
		{input: "(def is_small (x) (if (< x 5) 1 0))(def main() (is_small 6)", evaluated: 0},
		{input: "(def fib (n) (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2))))) (def main () (fib 8))", evaluated: 21},
	}
	for _, input := range inputs {
		parser := newParser(input.input)
		scope := &Scope{inner: make(map[string]ASTNode), outer: nil}
		var evaluated int
		expressions := parser.Parse()
		for _, expression := range expressions {
			evaluated = expression.eval(scope)
		}
		if evaluated != input.evaluated {
			t.Errorf(fmt.Sprintf("Expected 5, got %d", evaluated))
		}
	}
}

func TestParserIfExpr(t *testing.T) {
	input := "(def main() (if (< n 2) n (+ 1 n)))"
	parser := newParser(input)
	function, ok := parser.ParseExpression().(*FunctionNode)
	if !ok {
		panic("Expected function node")
	}
	ifNode, ok := function.body[0].(*IfNode)
	if !ok {
		panic("Expected if node")
	}
	if ifNode.condition.operand != "<" {
		t.Errorf("Expected < got %s", ifNode.condition.operand)
	}
	if len(ifNode.condition.arguments) != 2 {
		t.Errorf("Error")
	}
	identifierTrueExp, ok := ifNode.trueExpr.(*IdentifierNode)
	if !ok {
		t.Errorf("Identifier needed got something else")
	}
	if identifierTrueExp.name != "n" {
		t.Errorf("Expected n, got %s", identifierTrueExp.name)
	}
	identifierFalseExp, ok := ifNode.falseExpr.(*SExpr)
	if identifierFalseExp.operand != "+" {
		t.Errorf("Expected +, got %s", identifierFalseExp.operand)
	}
	if len(identifierFalseExp.arguments) != 2 {
		t.Errorf("Error")
	}
}

func TestParserFunctions(t *testing.T) {
	type FunctionTestCase struct {
		input      string
		name       string
		arguments  []string
		bodyLength int
	}
	inputs := []FunctionTestCase{
		{input: "(def main() (+ 1 2))", name: "main", arguments: []string{}, bodyLength: 1},
		{input: "(def plus_two (a) (+ a 2))", name: "plus_two", arguments: []string{"a"}, bodyLength: 1},
	}
	for _, input := range inputs {
		parser := newParser(input.input)
		node := parser.ParseExpression()
		functionNode, ok := node.(*FunctionNode)
		if !ok {
			t.Errorf("Expected function node, got another")
		}
		if functionNode.name != input.name {
			t.Errorf(fmt.Sprintf("Expected function name %s, got %s", input.name, functionNode.name))
		}
		if len(functionNode.arguments) != len(input.arguments) {
			t.Errorf(fmt.Sprintf("Expected %d arguments, got %d", len(input.arguments), len(functionNode.arguments)))
		}
		sort.Strings(functionNode.arguments)
		sort.Strings(input.arguments)
		for indx, argument := range input.arguments {
			if functionNode.arguments[indx] != argument {
				t.Errorf("Argument mismatch in test case")
			}
		}
	}
}

func TestParserSExpr(t *testing.T) {
	type TestCase struct {
		input  string
		output int
	}
	testCases := []TestCase{
		{
			input: "(+ 1 2 )", output: 3,
		},
		{
			input: "(+ 1 (* 3 4 ))", output: 13,
		},
		{input: "(+ (/ 7 2) (* 2 3) )", output: 9},
		{input: "(+ (/ 7 2) (* 2 3 4) )", output: 27},
	}
	for _, testCase := range testCases {
		fmt.Println("testing input: ", testCase)
		parser := newParser(testCase.input)
		scope := &Scope{inner: make(map[string]ASTNode), outer: nil}
		evaled := parser.ParseExpression().eval(scope)
		output := testCase.output
		if evaled != output {
			t.Errorf("Evaluation incorrect, expected %d, got %d\n", evaled, output)
		}
	}
}
