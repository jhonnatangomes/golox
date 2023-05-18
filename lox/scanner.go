package lox

type Scanner struct {
	source  string
	start   int
	current int
	line    int
}

type TokenType int

type Token struct {
	tokenType TokenType
	lexeme    string
	line      int
}

const (
	TokenLeftParen TokenType = iota
	TokenRightParen
	TokenLeftBrace
	TokenRightBrace
	TokenComma
	TokenDot
	TokenMinus
	TokenPlus
	TokenSemicolon
	TokenSlash
	TokenStar

	TokenBang
	TokenBangEqual
	TokenEqual
	TokenEqualEqual
	TokenGreater
	TokenGreaterEqual
	TokenLess
	TokenLessEqual

	TokenIdentifier
	TokenString
	TokenNumber

	TokenAnd
	TokenClass
	TokenElse
	TokenFalse
	TokenFor
	TokenFun
	TokenIf
	TokenNil
	TokenOr
	TokenPrint
	TokenReturn
	TokenSuper
	TokenThis
	TokenTrue
	TokenVar
	TokenWhile

	TokenError
	TokenEOF
)

func NewScanner(source string) *Scanner {
	return &Scanner{
		source:  source,
		start:   0,
		current: 0,
		line:    1,
	}
}

func (scanner *Scanner) scanToken() Token {
	scanner.skipWhitespace()
	scanner.start = scanner.current
	if scanner.isAtEnd() {
		return scanner.makeToken(TokenEOF)
	}
	c := scanner.advance()
	if isAlpha(c) {
		return scanner.identifier()
	}
	if isDigit(c) {
		return scanner.number()
	}
	switch c {
	case '(':
		return scanner.makeToken(TokenLeftParen)
	case ')':
		return scanner.makeToken(TokenRightParen)
	case '{':
		return scanner.makeToken(TokenLeftBrace)
	case '}':
		return scanner.makeToken(TokenRightBrace)
	case ';':
		return scanner.makeToken(TokenSemicolon)
	case ',':
		return scanner.makeToken(TokenComma)
	case '.':
		return scanner.makeToken(TokenDot)
	case '-':
		return scanner.makeToken(TokenMinus)
	case '+':
		return scanner.makeToken(TokenPlus)
	case '/':
		return scanner.makeToken(TokenSlash)
	case '*':
		return scanner.makeToken(TokenStar)
	case '!':
		{
			if scanner.match('=') {
				return scanner.makeToken(TokenBangEqual)
			} else {
				return scanner.makeToken(TokenBang)
			}
		}
	case '=':
		{
			if scanner.match('=') {
				return scanner.makeToken(TokenEqualEqual)
			} else {
				return scanner.makeToken(TokenEqual)
			}
		}
	case '<':
		{
			if scanner.match('=') {
				return scanner.makeToken(TokenLessEqual)
			} else {
				return scanner.makeToken(TokenLess)
			}
		}
	case '>':
		{
			if scanner.match('=') {
				return scanner.makeToken(TokenGreaterEqual)
			} else {
				return scanner.makeToken(TokenGreater)
			}
		}
	case '"':
		return scanner.string()
	}
	return scanner.errorToken("Unexpected character.")
}

func (scanner *Scanner) skipWhitespace() {
	for {
		c := scanner.peek()
		switch c {
		case ' ', '\r', '\t':
			scanner.advance()
		case '\n':
			scanner.line += 1
			scanner.advance()
		case '/':
			if scanner.peekNext() == '/' {
				for scanner.peek() != '\n' && !scanner.isAtEnd() {
					scanner.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

func (scanner *Scanner) peekNext() byte {
	if scanner.isAtEnd() {
		return '\000'
	}
	return scanner.source[scanner.current+1]
}

func (scanner *Scanner) peek() byte {
	if scanner.isAtEnd() {
		return '\000'
	}
	return scanner.source[scanner.current]
}

func (scanner *Scanner) advance() byte {
	scanner.current += 1
	return scanner.source[scanner.current-1]
}

func (scanner *Scanner) match(expected byte) bool {
	if scanner.isAtEnd() {
		return false
	}
	if scanner.source[scanner.current] != expected {
		return false
	}
	scanner.current += 1
	return true
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.current >= len(scanner.source)
}

func (scanner *Scanner) makeToken(tokenType TokenType) Token {
	return Token{
		tokenType: tokenType,
		lexeme:    scanner.source[scanner.start:scanner.current],
		line:      scanner.line,
	}
}

func (scanner *Scanner) errorToken(message string) Token {
	return Token{
		tokenType: TokenError,
		lexeme:    message,
		line:      scanner.line,
	}
}

func (scanner *Scanner) string() Token {
	for scanner.peek() != '"' && !scanner.isAtEnd() {
		if scanner.peek() == '\n' {
			scanner.line += 1
		}
		scanner.advance()
	}
	if scanner.isAtEnd() {
		return scanner.errorToken("Unterminated string.")
	}
	scanner.advance()
	return scanner.makeToken(TokenString)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (scanner *Scanner) number() Token {
	for isDigit(scanner.peek()) {
		scanner.advance()
	}
	if scanner.peek() == '.' && isDigit(scanner.peekNext()) {
		scanner.advance()
		for isDigit(scanner.peek()) {
			scanner.advance()
		}
	}
	return scanner.makeToken(TokenNumber)
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func (scanner *Scanner) identifier() Token {
	for isAlpha(scanner.peek()) || isDigit(scanner.peek()) {
		scanner.advance()
	}
	identifier := scanner.source[scanner.start:scanner.current]
	switch identifier {
	case "and":
		return scanner.makeToken(TokenAnd)
	case "class":
		return scanner.makeToken(TokenClass)
	case "else":
		return scanner.makeToken(TokenElse)
	case "false":
		return scanner.makeToken(TokenFalse)
	case "for":
		return scanner.makeToken(TokenFor)
	case "fun":
		return scanner.makeToken(TokenFun)
	case "if":
		return scanner.makeToken(TokenIf)
	case "nil":
		return scanner.makeToken(TokenNil)
	case "or":
		return scanner.makeToken(TokenOr)
	case "print":
		return scanner.makeToken(TokenPrint)
	case "return":
		return scanner.makeToken(TokenReturn)
	case "super":
		return scanner.makeToken(TokenSuper)
	case "this":
		return scanner.makeToken(TokenThis)
	case "true":
		return scanner.makeToken(TokenTrue)
	case "var":
		return scanner.makeToken(TokenVar)
	case "while":
		return scanner.makeToken(TokenWhile)
	default:
		return scanner.makeToken(TokenIdentifier)
	}
}
