package lox

import "fmt"

func (chunk *Chunk) Disassemble(name string) {
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < len(chunk.code); {
		offset = chunk.disassembleInstruction(offset)
	}
}

func (chunk *Chunk) disassembleInstruction(offset int) int {
	fmt.Printf("%04d ", offset)
	if offset > 0 && chunk.lines[offset] == chunk.lines[offset-1] {
		fmt.Printf("   | ")
	} else {
		fmt.Printf("%4d ", chunk.lines[offset])
	}

	instruction := chunk.code[offset]
	switch OpCode(instruction) {
	case OpReturn:
		return chunk.simpleInstruction("OP_RETURN", offset)
	case OpConstant:
		return chunk.constantInstruction("OP_CONSTANT", offset)
	case OpNegate:
		return chunk.simpleInstruction("OP_NEGATE", offset)
	case OpAdd:
		return chunk.simpleInstruction("OP_ADD", offset)
	case OpSubtract:
		return chunk.simpleInstruction("OP_SUBTRACT", offset)
	case OpMultiply:
		return chunk.simpleInstruction("OP_MULTIPLY", offset)
	case OpDivide:
		return chunk.simpleInstruction("OP_DIVIDE", offset)
	default:
		fmt.Printf("Unknown opcode %d\n", instruction)
		return offset + 1
	}
}

func (chunk *Chunk) simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func (chunk *Chunk) constantInstruction(name string, offset int) int {
	constant := chunk.code[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	printValue(chunk.constants[constant])
	fmt.Printf("'\n")
	return offset + 2
}
