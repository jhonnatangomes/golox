package main

import "github.com/jhonnatangomes/golox/lox"

func main() {
	chunk := lox.NewChunk()
	vm := lox.NewVm(chunk)
	constant := chunk.AddConstant(1.2)
	chunk.Write(byte(lox.OpConstant), 123)
	chunk.Write(byte(constant), 123)
	constant = chunk.AddConstant(3.4)
	chunk.Write(byte(lox.OpConstant), 123)
	chunk.Write(byte(constant), 123)

	chunk.Write(byte(lox.OpAdd), 123)

	constant = chunk.AddConstant(5.6)
	chunk.Write(byte(lox.OpConstant), 123)
	chunk.Write(byte(constant), 123)

	chunk.Write(byte(lox.OpDivide), 123)
	chunk.Write(byte(lox.OpNegate), 123)
	chunk.Write(byte(lox.OpReturn), 123)
	// chunk.Disassemble("test chunk")
	vm.Interpret()
}
