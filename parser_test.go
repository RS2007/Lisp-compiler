package main

import (
	"testing"
)

func TestParser(t *testing.T){
	type TestCase struct {
		input string
		output int
	}
	testCases := []TestCase{{
		input:"(+ 1 2 )",output: 3,
	},
	{
	input: "(+ 1 (* 3 4 ))",output:13},
	{ input: "(+ (/ 7 2) (* 2 3) )",output: 9},
	}
	for _,testCase := range testCases {
		parser := newParser(testCase.input)
		evaled := parser.Parse().eval()
		output := testCase.output
		if  evaled != output {
			t.Errorf("Evaluation incorrect, expected %d, got %d\n",	evaled,output)
		}
	}
}
