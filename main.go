package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/jhonnatangomes/golox/lox"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		repl()
	} else if len(args) == 1 {
		runFile(args[1])
	} else {
		fmt.Fprintf(os.Stderr, "Usage: golox [path]\n")
		os.Exit(64)
	}
}

func repl() {
	chunk := lox.NewChunk()
	vm := lox.NewVm(chunk)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			os.Exit(64)
		}
		vm.Interpret(input)
	}
}

func runFile(path string) {
	chunk := lox.NewChunk()
	vm := lox.NewVm(chunk)
	source := readFile(path)
	result := vm.Interpret(source)

	if result == lox.InterpretCompileError {
		os.Exit(65)
	}
	if result == lox.InterpretRuntimeError {
		os.Exit(70)
	}
}

func readFile(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file %s\n", path)
		os.Exit(74)
	}
	return string(file)
}
