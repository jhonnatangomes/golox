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
	case OpNil:
		return chunk.simpleInstruction("OP_NIL", offset)
	case OpTrue:
		return chunk.simpleInstruction("OP_TRUE", offset)
	case OpFalse:
		return chunk.simpleInstruction("OP_FALSE", offset)
	case OpNot:
		return chunk.simpleInstruction("OP_NOT", offset)
	case OpEqual:
		return chunk.simpleInstruction("OP_EQUAL", offset)
	case OpNotEqual:
		return chunk.simpleInstruction("OP_NOT_EQUAL", offset)
	case OpGreater:
		return chunk.simpleInstruction("OP_GREATER", offset)
	case OpGreaterEqual:
		return chunk.simpleInstruction("OP_GREATER_EQUAL", offset)
	case OpLess:
		return chunk.simpleInstruction("OP_LESS", offset)
	case OpLessEqual:
		return chunk.simpleInstruction("OP_LESS_EQUAL", offset)
	case OpPrint:
		return chunk.simpleInstruction("OP_PRINT", offset)
	case OpPop:
		return chunk.simpleInstruction("OP_POP", offset)
	case OpDefineGlobal:
		return chunk.constantInstruction("OP_DEFINE_GLOBAL", offset)
	case OpGetGlobal:
		return chunk.constantInstruction("OP_GET_GLOBAL", offset)
	case OpSetGlobal:
		return chunk.constantInstruction("OP_SET_GLOBAL", offset)
	case OpGetLocal:
		return chunk.byteInstruction("OP_GET_LOCAL", offset)
	case OpSetLocal:
		return chunk.byteInstruction("OP_SET_LOCAL", offset)
	case OpJump:
		return chunk.jumpInstruction("OP_JUMP", 1, offset)
	case OpJumpIfFalse:
		return chunk.jumpInstruction("OP_JUMP_IF_FALSE", 1, offset)
	case OpLoop:
		return chunk.jumpInstruction("OP_LOOP", -1, offset)
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
	chunk.constants[constant].print()
	fmt.Printf("'\n")
	return offset + 2
}

func (chunk *Chunk) byteInstruction(name string, offset int) int {
	slot := chunk.code[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot)
	return offset + 2
}

func (chunk *Chunk) jumpInstruction(name string, sign int, offset int) int {
	jump := int(chunk.code[offset+1]) << 8
	jump |= int(chunk.code[offset+2])
	fmt.Printf("%-16s %4d -> %d\n", name, offset, offset+3+sign*jump)
	return offset + 3
}
