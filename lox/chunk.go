package lox

type OpCode int

const (
	Return OpCode = iota
)

type Chunk struct {
	code []byte
}

func NewChunk() *Chunk {
	return &Chunk{
		code: make([]byte, 0),
	}
}

func (c *Chunk) WriteByte(byte byte) error {
	c.code = append(c.code, byte)
	return nil
}
