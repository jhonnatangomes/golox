package lox

import (
	"fmt"
	"os"
	"strconv"
)

type Compiler struct {
	chunk      *Chunk
	scanner    *Scanner
	source     string
	previous   Token
	current    Token
	hadError   bool
	panicMode  bool
	locals     []Local
	scopeDepth int
}

type Local struct {
	name  Token
	depth int
}

func NewCompiler(source string, chunk *Chunk) *Compiler {
	return &Compiler{
		chunk:      chunk,
		scanner:    NewScanner(source),
		source:     source,
		previous:   Token{},
		current:    Token{},
		hadError:   false,
		panicMode:  false,
		locals:     make([]Local, 0),
		scopeDepth: 0,
	}
}

func (compiler *Compiler) compile() bool {
	compiler.advance()
	for !compiler.match(TokenEOF) {
		compiler.declaration()
	}
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
func (compiler *Compiler) match(tokenType TokenType) bool {
	if !compiler.check(tokenType) {
		return false
	}
	compiler.advance()
	return true
}

func (compiler *Compiler) check(tokenType TokenType) bool {
	return compiler.current.tokenType == tokenType
}

func (compiler *Compiler) declaration() {
	if compiler.match(TokenVar) {
		compiler.varDeclaration()
	} else {
		compiler.statement()
	}

	if compiler.panicMode {
		compiler.synchronize()
	}
}

func (compiler *Compiler) synchronize() {
	compiler.panicMode = false

	for compiler.current.tokenType != TokenEOF {
		if compiler.previous.tokenType == TokenSemicolon {
			return
		}
		switch compiler.current.tokenType {
		case TokenClass, TokenFun, TokenVar, TokenFor, TokenIf, TokenWhile, TokenPrint, TokenReturn:
			return
		}
		compiler.advance()
	}
}

func (compiler *Compiler) varDeclaration() {
	global := compiler.parseVariable("Expect variable name.")
	if compiler.match(TokenEqual) {
		compiler.expression()
	} else {
		compiler.emitByte(byte(OpNil))
	}
	compiler.consume(TokenSemicolon, "Expect ';' after variable declaration.")

	compiler.defineVariable(global)
}

func (compiler *Compiler) parseVariable(errorMessage string) int {
	compiler.consume(TokenIdentifier, errorMessage)
	compiler.declareVariable()
	if compiler.scopeDepth > 0 {
		return 0
	}
	return compiler.identifierConstant(&compiler.previous)
}

func (compiler *Compiler) declareVariable() {
	if compiler.scopeDepth == 0 {
		return
	}
	name := compiler.previous
	for i := len(compiler.locals) - 1; i >= 0; i-- {
		local := compiler.locals[i]
		if local.depth != -1 && local.depth < compiler.scopeDepth {
			break
		}
		if name.lexeme == local.name.lexeme {
			compiler.error("Already a variable with this name in this scope.")
		}
	}
	compiler.addLocal(name)
}

func (compiler *Compiler) addLocal(name Token) {
	compiler.locals = append(compiler.locals, Local{name, -1})
}

func (compiler *Compiler) identifierConstant(token *Token) int {
	return compiler.makeConstant(StringValue(token.lexeme))
}

func (compiler *Compiler) defineVariable(global int) {
	if compiler.scopeDepth > 0 {
		compiler.markInitialized()
		return
	}
	compiler.emitBytes(byte(OpDefineGlobal), byte(global))
}

func (compiler *Compiler) markInitialized() {
	compiler.locals[len(compiler.locals)-1].depth = compiler.scopeDepth
}

func (compiler *Compiler) statement() {
	if compiler.match(TokenPrint) {
		compiler.printStatement()
	} else if compiler.match(TokenLeftBrace) {
		compiler.beginScope()
		compiler.block()
		compiler.endScope()
	} else {
		compiler.expressionStatement()
	}
}

func (compiler *Compiler) beginScope() {
	compiler.scopeDepth++
}

func (compiler *Compiler) endScope() {
	compiler.scopeDepth--
	for len(compiler.locals) > 0 && compiler.locals[len(compiler.locals)-1].depth > compiler.scopeDepth {
		compiler.emitByte(byte(OpPop))
		compiler.locals = compiler.locals[:len(compiler.locals)-1]
	}
}

func (compiler *Compiler) block() {
	for !compiler.check(TokenRightBrace) && !compiler.check(TokenEOF) {
		compiler.declaration()
	}
	compiler.consume(TokenRightBrace, "Expect '}' after block.")
}

func (compiler *Compiler) printStatement() {
	compiler.expression()
	compiler.consume(TokenSemicolon, "Expect ';' after value.")
	compiler.emitByte(byte(OpPrint))
}

func (compiler *Compiler) expressionStatement() {
	compiler.expression()
	compiler.consume(TokenSemicolon, "Expect ';' after expression.")
	compiler.emitByte(byte(OpPop))
}

func (compiler *Compiler) expression() {
	compiler.parsePrecedence(PrecedenceAssignment)
}

func (compiler *Compiler) number(_ bool) {
	value, _ := strconv.ParseFloat(compiler.previous.lexeme, 64)
	compiler.emitConstant(NumberValue(value))
}

func (compiler *Compiler) grouping(_ bool) {
	compiler.expression()
	compiler.consume(TokenRightParen, "Expect ')' after expression.")
}

func (compiler *Compiler) unary(_ bool) {
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
	canAssign := precedence <= PrecedenceAssignment
	prefixRule(canAssign)

	for precedence <= compiler.getRule(compiler.current.tokenType).precedence {
		compiler.advance()
		infixRule := compiler.getRule(compiler.previous.tokenType).infix
		infixRule(canAssign)
	}

	if canAssign && compiler.match(TokenEqual) {
		compiler.error("Invalid assignment target.")
	}
}

func (compiler *Compiler) binary(_ bool) {
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

func (compiler *Compiler) literal(_ bool) {
	switch compiler.previous.tokenType {
	case TokenFalse:
		compiler.emitByte(byte(OpFalse))
	case TokenTrue:
		compiler.emitByte(byte(OpTrue))
	case TokenNil:
		compiler.emitByte(byte(OpNil))
	}
}

func (compiler *Compiler) string(_ bool) {
	lexeme := compiler.previous.lexeme
	compiler.emitConstant(StringValue(lexeme[1 : len(lexeme)-1]))
}

func (compiler *Compiler) variable(canAssign bool) {
	compiler.namedVariable(compiler.previous, canAssign)
}

func (compiler *Compiler) namedVariable(token Token, canAssign bool) {
	var getOp, setOp byte
	arg := compiler.resolveLocal(token)
	if arg != -1 {
		getOp = byte(OpGetLocal)
		setOp = byte(OpSetLocal)
	} else {
		arg = compiler.identifierConstant(&token)
		getOp = byte(OpGetGlobal)
		setOp = byte(OpSetGlobal)
	}

	if canAssign && compiler.match(TokenEqual) {
		compiler.expression()
		compiler.emitBytes(setOp, byte(arg))
	} else {
		compiler.emitBytes(getOp, byte(arg))
	}
}

func (compiler *Compiler) resolveLocal(token Token) int {
	for i := len(compiler.locals) - 1; i >= 0; i-- {
		local := compiler.locals[i]
		if local.name.lexeme == token.lexeme {
			if local.depth == -1 {
				compiler.error("Cannot read local variable in its own initializer.")
			}
			return i
		}
	}
	return -1
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
		TokenIdentifier:   {compiler.variable, nil, PrecedenceNone},
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
	prefix     func(canAssign bool)
	infix      func(canAssign bool)
	precedence Precedence
}
