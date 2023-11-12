package main

import (
	"fmt"
	"lisp-compiler/core"
	"lisp-compiler/utils"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`
			Usage: lisp-compiler <mode> <input-path>
			mode: interpret,compile, default: compile
		`)
		return
	}
	var input string
	var err error
	var mode string
	if len(os.Args) == 3 {
		mode = os.Args[1]
		input, err = utils.LoadLispFileToString(strings.TrimSpace(os.Args[2]))
	}
	if len(os.Args) == 2 {
		mode = "compile"
		input, err = utils.LoadLispFileToString(strings.TrimSpace(os.Args[1]))
	}
	if err != nil {
		panic(err)
	}
	parser := core.NewParser(input)
	asm := ""
	symbol := "%sym1"
	parsed := parser.Parse()

	if mode == "interpret" {
		scope := core.NewInterpreterScope(nil)
		var value int
		for _, parsedExpr := range parsed {
			value = parsedExpr.Eval(scope)
		}
		fmt.Println(value)
		return
	} else {
		scope := core.NewCompilerScope(nil)
		for _, parsedExpr := range parsed {
			parsedExpr.Codegen(&asm, symbol, scope)
			asm += "\n"
		}
		utils.WriteLLVMAssembly(asm)
	}
}
