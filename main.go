package main

import "github.com/jhonnatangomes/golox/lox"

func main() {
	chunk := lox.NewChunk()
	constant := chunk.AddConstant(1.2)
	chunk.Write(byte(lox.OpConstant), 123)
	chunk.Write(byte(constant), 123)
	chunk.Write(byte(lox.OpReturn), 123)
	chunk.Disassemble("test chunk")
}
