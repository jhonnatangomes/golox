package lox

type OpCode int

const (
	OpReturn OpCode = iota
	OpConstant
	OpNegate
	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
)

type Chunk struct {
	code      []byte
	constants []Value
	lines     []int
}

func NewChunk() *Chunk {
	return &Chunk{
		code:      make([]byte, 0),
		constants: make([]Value, 0),
		lines:     make([]int, 0),
	}
}

func (chunk *Chunk) Write(byte byte, line int) {
	chunk.code = append(chunk.code, byte)
	chunk.lines = append(chunk.lines, line)
}

func (chunk *Chunk) AddConstant(value Value) int {
	chunk.constants = append(chunk.constants, value)
	return len(chunk.constants) - 1
}
