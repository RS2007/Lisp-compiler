package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runCommand(command []string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeArmAssembly(parser *Parser) {
	file, e := os.Create("output.s")
	if e != nil {
		panic(e)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	if _, err := writer.WriteString(""); err != nil {
		panic(err)
	}
	err := writer.Flush()
	if err != nil {
		panic(err)
	}

	asCommand := []string{"as", "-arch", "arm64", "output.s", "-o", "output.o"}
	ldCommand := []string{"ld", "-o", "output", "output.o", "-lSystem", "-syslibroot", `/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk`, "-e", "_start", "-arch", "arm64"}
	for _, command := range ldCommand {
		fmt.Printf("%s ", command)
	}

	if err := runCommand(asCommand); err != nil {
		fmt.Println("Error running 'as' command:", err)
		return
	}

	if err := runCommand(ldCommand); err != nil {
		fmt.Println("Error running 'ld' command:", err)
		return
	}
}

func writeLLVMAssembly(asm string) {
	file, e := os.Create("output.ll")
	if e != nil {
		panic(e)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	if _, err := writer.WriteString(asm); err != nil {
		panic(err)
	}
	err := writer.Flush()
	if err != nil {
		panic(err)
	}
	llvmCommand := []string{"llc", "-o", "output.s", "output.ll"}
	compileCommand := []string{"gcc", "-o", "output", "output.s"}

	if err := runCommand(llvmCommand); err != nil {
		fmt.Println("Error running 'as' command:", err)
		return
	}

	if err := runCommand(compileCommand); err != nil {
		fmt.Println("Error running 'ld' command:", err)
		return
	}
}

func loadLispFileToString(fileName string) (string, error) {
	// Read the file into a byte slice
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(fileContents)), nil
}

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
		input, err = loadLispFileToString(strings.TrimSpace(os.Args[2]))
	}
	if len(os.Args) == 2 {
		mode = "compile"
		input, err = loadLispFileToString(strings.TrimSpace(os.Args[1]))
	}
	if err != nil {
		panic(err)
	}
	parser := newParser(input)
	asm := ""
	symbol := "%sym1"
	parsed := parser.Parse()

	if mode == "interpret" {
		scope := newInterpreterScope(nil)
		var value int
		for _, parsedExpr := range parsed {
			value = parsedExpr.eval(scope)
		}
		fmt.Println(value)
		return
	} else {
		scope := newCompilerScope(nil)
		for _, parsedExpr := range parsed {
			parsedExpr.codegen(&asm, symbol, scope)
			asm += "\n"
		}
		writeLLVMAssembly(asm)
	}
}
