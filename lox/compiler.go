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
	compiler.emitConstant(NumberValue(value))
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
	case TokenBang:
		compiler.emitByte(byte(OpNot))
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
	case TokenBangEqual:
		compiler.emitByte(byte(OpNotEqual))
	case TokenEqualEqual:
		compiler.emitByte(byte(OpEqual))
	case TokenGreater:
		compiler.emitByte(byte(OpGreater))
	case TokenGreaterEqual:
		compiler.emitByte(byte(OpGreaterEqual))
	case TokenLess:
		compiler.emitByte(byte(OpLess))
	case TokenLessEqual:
		compiler.emitByte(byte(OpLessEqual))
	}
}

func (compiler *Compiler) literal() {
	switch compiler.previous.tokenType {
	case TokenFalse:
		compiler.emitByte(byte(OpFalse))
	case TokenTrue:
		compiler.emitByte(byte(OpTrue))
	case TokenNil:
		compiler.emitByte(byte(OpNil))
	}
}

func (compiler *Compiler) string() {
	lexeme := compiler.previous.lexeme
	compiler.emitConstant(StringValue(lexeme[1 : len(lexeme)-1]))
}

func (compiler *Compiler) getRule(tokenType TokenType) ParseRule {
	rules := map[TokenType]ParseRule{
		TokenLeftParen:    {compiler.grouping, nil, PrecedenceNone},
		TokenRightParen:   {nil, nil, PrecedenceNone},
		TokenLeftBrace:    {nil, nil, PrecedenceNone},
		TokenRightBrace:   {nil, nil, PrecedenceNone},
		TokenComma:        {nil, nil, PrecedenceNone},
		TokenDot:          {nil, nil, PrecedenceNone},
		TokenMinus:        {compiler.unary, compiler.binary, PrecedenceTerm},
		TokenPlus:         {nil, compiler.binary, PrecedenceTerm},
		TokenSemicolon:    {nil, nil, PrecedenceNone},
		TokenSlash:        {nil, compiler.binary, PrecedenceFactor},
		TokenStar:         {nil, compiler.binary, PrecedenceFactor},
		TokenBang:         {compiler.unary, nil, PrecedenceNone},
		TokenBangEqual:    {nil, compiler.binary, PrecedenceEquality},
		TokenEqual:        {nil, nil, PrecedenceNone},
		TokenEqualEqual:   {nil, compiler.binary, PrecedenceEquality},
		TokenGreater:      {nil, compiler.binary, PrecedenceComparison},
		TokenGreaterEqual: {nil, compiler.binary, PrecedenceComparison},
		TokenLess:         {nil, compiler.binary, PrecedenceComparison},
		TokenLessEqual:    {nil, compiler.binary, PrecedenceComparison},
		TokenIdentifier:   {nil, nil, PrecedenceNone},
		TokenString:       {compiler.string, nil, PrecedenceNone},
		TokenNumber:       {compiler.number, nil, PrecedenceNone},
		TokenAnd:          {nil, nil, PrecedenceNone},
		TokenClass:        {nil, nil, PrecedenceNone},
		TokenElse:         {nil, nil, PrecedenceNone},
		TokenFalse:        {compiler.literal, nil, PrecedenceNone},
		TokenFor:          {nil, nil, PrecedenceNone},
		TokenFun:          {nil, nil, PrecedenceNone},
		TokenIf:           {nil, nil, PrecedenceNone},
		TokenNil:          {compiler.literal, nil, PrecedenceNone},
		TokenOr:           {nil, nil, PrecedenceNone},
		TokenPrint:        {nil, nil, PrecedenceNone},
		TokenReturn:       {nil, nil, PrecedenceNone},
		TokenSuper:        {nil, nil, PrecedenceNone},
		TokenThis:         {nil, nil, PrecedenceNone},
		TokenTrue:         {compiler.literal, nil, PrecedenceNone},
		TokenVar:          {nil, nil, PrecedenceNone},
		TokenWhile:        {nil, nil, PrecedenceNone},
		TokenError:        {nil, nil, PrecedenceNone},
		TokenEOF:          {nil, nil, PrecedenceNone},
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
