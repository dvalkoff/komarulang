package tokenizer

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type Tokenizer struct {
	File string
}

func (t *Tokenizer) Scan(reader io.Reader) ([]Token, error) {
	scanner := bufio.NewScanner(reader)
	currentLine := 0
	tokens := make([]Token, 0)
	for ; scanner.Scan(); currentLine++ {
		line := scanner.Text()
		lineTokenizer := LineTokenizer{
			source:  []rune(line),
			current: 0,
			line:    currentLine,
			file:    t.File,
		}
		lineTokens, err := lineTokenizer.Scan()
		if err != nil {
			return nil, err
		}
		lineTokens = t.addSemicolon(lineTokens)
		tokens = append(tokens, lineTokens...)
	}
	if err := scanner.Err(); err != nil {
		return nil, TokenizerError{
			File:    t.File,
			Line:    currentLine,
			Message: "Tokenization failed",
			Cause:   err,
		}
	}
	tokens = append(tokens, GetEOF())
	return tokens, nil
}

func (t *Tokenizer) addSemicolon(lineTokens []Token) []Token {
	if len(lineTokens) == 0 {
		return lineTokens
	}
	lastTokenOnLine := lineTokens[len(lineTokens)-1]
	if lastTokenOnLine.TokenType.Match(Identifier, Print, Integer, Bool, RightBrace, RightParen, RightBrace, Break, Continue, Return) {
		lineTokens = append(lineTokens, Token{TokenType: Semicolon, Value: nil})
	}
	return lineTokens
}

type LineTokenizer struct {
	source  []rune
	current int
	line    int
	file    string
}

func (t *LineTokenizer) Scan() ([]Token, error) {
	tokens := make([]Token, 0)
	for !t.isEnd() {
		potentialToken := t.advance()
		if unicode.IsSpace(potentialToken) {
			continue
		}
		token, err := t.token(potentialToken)
		if err != nil {
			return nil, TokenizerError{
				File:    t.file,
				Line:    t.line,
				Message: fmt.Sprintf("Failed to recognize token %d:%d. token: %v", t.line, t.current, t.previous()),
				Cause:   err,
			}
		}
		if token.TokenType == EOL {
			return tokens, nil
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (t *LineTokenizer) token(val rune) (Token, error) {
	start := t.current-1
	end := t.current
	switch val {
	case '+':
		return Token{TokenType: Plus, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '-':
		return Token{TokenType: Minus, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '*':
		return Token{TokenType: Star, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '%':
		return Token{TokenType: Percent, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '(':
		return Token{TokenType: LeftParen, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case ')':
		return Token{TokenType: RightParen, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '{':
		return Token{TokenType: LeftBrace, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '}':
		return Token{TokenType: RightBrace, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case ';':
		return Token{TokenType: Semicolon, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '^':
		return Token{TokenType: Caret, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case ',':
		return Token{TokenType: Comma, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '&':
		if t.match('&') {
			end++
			return Token{TokenType: AmpersandAmpersand, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
		}
		return Token{TokenType: Ampersand, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '|':
		if t.match('|') {
			end++
			return Token{TokenType: VbarVbar, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
		}
		return Token{TokenType: Vbar, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '/':
		if t.match('/') {
			end++
			return Token{TokenType: EOL, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
		}
		return Token{TokenType: Slash, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '!':
		if t.match('=') {
			end++
			return Token{TokenType: BangEqual, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
		}
		return Token{TokenType: Bang, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '>':
		if t.match('=') {
			end++
			return Token{TokenType: GreaterEqual, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
		}
		return Token{TokenType: Greater, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '<':
		if t.match('=') {
			end++
			return Token{TokenType: LessEqual, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
		}
		return Token{TokenType: Less, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case '=':
		if t.match('=') {
			end++
			return Token{TokenType: EqualEqual, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
		}
		return Token{TokenType: Equal, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	default:
		if unicode.IsDigit(val) {
			return t.integer()
		}
		if unicode.IsLetter(val) {
			return t.keywordOrIdentifier()
		}
	}
	return Token{}, fmt.Errorf("Unrecongized token %v, position: %v", val, t.current)
}

func (t *LineTokenizer) integer() (Token, error) {
	start := t.current-1
	num := []rune{t.previous()}
	for !t.isEnd() && unicode.IsDigit(t.peek()) {
		num = append(num, t.advance())
	}
	end := t.current	
	intValue, err := strconv.Atoi(string(num))
	if err != nil {
		return Token{}, err
	}
	return Token{TokenType: Integer, Value: intValue, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
}

func (t *LineTokenizer) keywordOrIdentifier() (Token, error) {
	start := t.current-1
	wordRunes := []rune{t.previous()}
	for !t.isEnd() && (unicode.IsLetter(t.peek()) || unicode.IsDigit(t.peek())) {
		wordRunes = append(wordRunes, t.advance())
	}
	end := t.current
	word := string(wordRunes)
	switch word {
	case "var":
		return Token{TokenType: Var, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "true":
		return Token{TokenType: Bool, Value: true, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "false":
		return Token{TokenType: Bool, Value: false, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "print":
		return Token{TokenType: Print, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "if":
		return Token{TokenType: If, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "else":
		return Token{TokenType: Else, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "while":
		return Token{TokenType: While, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "for":
		return Token{TokenType: For, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "int":
		return Token{TokenType: Type, Value: word, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "bool":
		return Token{TokenType: Type, Value: word, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "fun":
		return Token{TokenType: Fun, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "return":
		return Token{TokenType: Return, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "break":
		return Token{TokenType: Break, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	case "continue":
		return Token{TokenType: Continue, Value: nil, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
	}
	return Token{TokenType: Identifier, Value: word, LineNumber: t.line, StartIndex: start, EndIndex: end}, nil
}

func (t *LineTokenizer) match(val rune) bool {
	if t.isEnd() {
		return false
	}
	current := t.peek()
	if current == val {
		t.advance()
		return true
	}
	return false
}

func (t *LineTokenizer) advance() rune {
	if !t.isEnd() {
		t.current++
	}
	return t.previous()
}

func (t *LineTokenizer) peek() rune {
	return t.source[t.current]
}

func (t *LineTokenizer) previous() rune {
	return t.source[t.current-1]
}

func (t *LineTokenizer) isEnd() bool {
	return len(t.source) <= t.current
}
