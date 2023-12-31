package utils

import (
	"bufio"
	"fmt"
	"lisp-compiler/core"
	"os"
	"os/exec"
	"strings"
)

func WriteLLVMAssembly(asm string) {
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

func LoadLispFileToString(fileName string) (string, error) {
	// Read the file into a byte slice
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(fileContents)), nil
}

func writeArmAssembly(parser *core.Parser) {
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

func runCommand(command []string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
