package lox

import (
	"fmt"
	"os"
	"strconv"
)

type Compiler struct {
	chunk     *Chunk
	scanner   *Scanner
	source    string
	previous  Token
	current   Token
	hadError  bool
	panicMode bool
}

func NewCompiler(source string, chunk *Chunk) *Compiler {
	return &Compiler{
		chunk:     chunk,
		scanner:   NewScanner(source),
		source:    source,
		previous:  Token{},
		current:   Token{},
		hadError:  false,
		panicMode: false,
	}
}

func (compiler *Compiler) compile() bool {
	compiler.advance()
	compiler.expression()
	compiler.consume(TokenEOF, "Expect end of expression.")
	compiler.end()
	return !compiler.hadError
}

func (compiler *Compiler) emitByte(byte byte) {
	compiler.currentChunk().Write(byte, compiler.previous.line)
}

func (compiler *Compiler) currentChunk() *Chunk {
	return compiler.chunk
}

func (compiler *Compiler) end() {
	compiler.emitReturn()

	//debug
	// if !compiler.hadError {
	// 	compiler.currentChunk().Disassemble("code")
	// }
}

func (compiler *Compiler) emitReturn() {
	compiler.emitByte(byte(OpReturn))
}

func (compiler *Compiler) emitBytes(byte1, byte2 byte) {
	compiler.emitByte(byte1)
	compiler.emitByte(byte2)
}

func (compiler *Compiler) emitConstant(value Value) {
	compiler.emitBytes(byte(OpConstant), byte(compiler.makeConstant(value)))
}

func (compiler *Compiler) makeConstant(value Value) int {
	return compiler.currentChunk().AddConstant(value)
}

func (compiler *Compiler) advance() {
	compiler.previous = compiler.current

	for {
		compiler.current = compiler.scanner.scanToken()
		if compiler.current.tokenType != TokenError {
			break
		}
		compiler.errorAtCurrent(compiler.current.lexeme)
	}
}

func (compiler *Compiler) errorAtCurrent(message string) {
	compiler.errorAt(&compiler.current, message)
}

func (compiler *Compiler) error(message string) {
	compiler.errorAt(&compiler.previous, message)
}

func (compiler *Compiler) errorAt(token *Token, message string) {
	if compiler.panicMode {
		return
	}
	compiler.panicMode = true
	fmt.Fprintf(os.Stderr, "[line %d] Error", token.line)

	if token.tokenType == TokenEOF {
		fmt.Fprintf(os.Stderr, " at end")
	} else if token.tokenType == TokenError {
		// Nothing.
	} else {
		fmt.Fprintf(os.Stderr, " at '%s'", token.lexeme)
	}

	fmt.Fprintf(os.Stderr, ": %s\n", message)
	compiler.hadError = true
}

func (compiler *Compiler) consume(tokenType TokenType, message string) {
	if compiler.current.tokenType == tokenType {
		compiler.advance()
		return
	}
	compiler.errorAtCurrent(message)
}

func (compiler *Compiler) expression() {
	compiler.parsePrecedence(PrecedenceAssignment)
}

func (compiler *Compiler) number() {
	value, _ := strconv.ParseFloat(compiler.previous.lexeme, 64)
	compiler.emitConstant(Value(value))
}

func (compiler *Compiler) grouping() {
	compiler.expression()
	compiler.consume(TokenRightParen, "Expect ')' after expression.")
}

func (compiler *Compiler) unary() {
	operatorType := compiler.previous.tokenType
	compiler.parsePrecedence(PrecedenceUnary)

	switch operatorType {
	case TokenMinus:
		compiler.emitByte(byte(OpNegate))
	}
}

func (compiler *Compiler) parsePrecedence(precedence Precedence) {
	compiler.advance()
	prefixRule := compiler.getRule(compiler.previous.tokenType).prefix
	if prefixRule == nil {
		compiler.error("Expect expression.")
		return
	}
	prefixRule()

	for precedence <= compiler.getRule(compiler.current.tokenType).precedence {
		compiler.advance()
		infixRule := compiler.getRule(compiler.previous.tokenType).infix
		infixRule()
	}
}

func (compiler *Compiler) binary() {
	operatorType := compiler.previous.tokenType
	rule := compiler.getRule(operatorType)
	compiler.parsePrecedence(rule.precedence + 1)

	switch operatorType {
	case TokenPlus:
		compiler.emitByte(byte(OpAdd))
	case TokenMinus:
		compiler.emitByte(byte(OpSubtract))
	case TokenStar:
		compiler.emitByte(byte(OpMultiply))
	case TokenSlash:
		compiler.emitByte(byte(OpDivide))
	}
}

func (compiler *Compiler) getRule(tokenType TokenType) ParseRule {
	rules := map[TokenType]ParseRule{
		TokenLeftParen:    ParseRule{compiler.grouping, nil, PrecedenceNone},
		TokenRightParen:   ParseRule{nil, nil, PrecedenceNone},
		TokenLeftBrace:    ParseRule{nil, nil, PrecedenceNone},
		TokenRightBrace:   ParseRule{nil, nil, PrecedenceNone},
		TokenComma:        ParseRule{nil, nil, PrecedenceNone},
		TokenDot:          ParseRule{nil, nil, PrecedenceNone},
		TokenMinus:        ParseRule{compiler.unary, compiler.binary, PrecedenceTerm},
		TokenPlus:         ParseRule{nil, compiler.binary, PrecedenceTerm},
		TokenSemicolon:    ParseRule{nil, nil, PrecedenceNone},
		TokenSlash:        ParseRule{nil, compiler.binary, PrecedenceFactor},
		TokenStar:         ParseRule{nil, compiler.binary, PrecedenceFactor},
		TokenBang:         ParseRule{nil, nil, PrecedenceNone},
		TokenBangEqual:    ParseRule{nil, nil, PrecedenceNone},
		TokenEqual:        ParseRule{nil, nil, PrecedenceNone},
		TokenEqualEqual:   ParseRule{nil, nil, PrecedenceNone},
		TokenGreater:      ParseRule{nil, nil, PrecedenceNone},
		TokenGreaterEqual: ParseRule{nil, nil, PrecedenceNone},
		TokenLess:         ParseRule{nil, nil, PrecedenceNone},
		TokenLessEqual:    ParseRule{nil, nil, PrecedenceNone},
		TokenIdentifier:   ParseRule{nil, nil, PrecedenceNone},
		TokenString:       ParseRule{nil, nil, PrecedenceNone},
		TokenNumber:       ParseRule{compiler.number, nil, PrecedenceNone},
		TokenAnd:          ParseRule{nil, nil, PrecedenceNone},
		TokenClass:        ParseRule{nil, nil, PrecedenceNone},
		TokenElse:         ParseRule{nil, nil, PrecedenceNone},
		TokenFalse:        ParseRule{nil, nil, PrecedenceNone},
		TokenFor:          ParseRule{nil, nil, PrecedenceNone},
		TokenFun:          ParseRule{nil, nil, PrecedenceNone},
		TokenIf:           ParseRule{nil, nil, PrecedenceNone},
		TokenNil:          ParseRule{nil, nil, PrecedenceNone},
		TokenOr:           ParseRule{nil, nil, PrecedenceNone},
		TokenPrint:        ParseRule{nil, nil, PrecedenceNone},
		TokenReturn:       ParseRule{nil, nil, PrecedenceNone},
		TokenSuper:        ParseRule{nil, nil, PrecedenceNone},
		TokenThis:         ParseRule{nil, nil, PrecedenceNone},
		TokenTrue:         ParseRule{nil, nil, PrecedenceNone},
		TokenVar:          ParseRule{nil, nil, PrecedenceNone},
		TokenWhile:        ParseRule{nil, nil, PrecedenceNone},
		TokenError:        ParseRule{nil, nil, PrecedenceNone},
		TokenEOF:          ParseRule{nil, nil, PrecedenceNone},
	}
	return rules[tokenType]
}

type Precedence int

const (
	PrecedenceNone Precedence = iota
	PrecedenceAssignment
	PrecedenceOr
	PrecedenceAnd
	PrecedenceEquality
	PrecedenceComparison
	PrecedenceTerm
	PrecedenceFactor
	PrecedenceUnary
	PrecedenceCall
	PrecedencePrimary
)

type ParseRule struct {
	prefix     func()
	infix      func()
	precedence Precedence
}
