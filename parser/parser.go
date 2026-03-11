package parser

import (
	"slices"
	"fmt"

	"github.com/dvalkoff/komarulang/tokenizer/token"
)

type Expression any

type IntegerLiteral struct {
	Value int
}

type BinaryExpression struct {
	Left     Expression
	Operator token.TokenType
	Right    Expression
}

type Parser struct {
	tokens  []token.Token
	current int
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

func (p *Parser) Expression() (Expression, error) {
	return p.binary()
}

func (p *Parser) binary() (Expression, error) {
	expression, err := p.literal()
	if err != nil {
		return nil, err
	}

	for p.match(token.Plus, token.Minus) {
		operator := p.previous().TokenType
		right, err := p.literal()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (p *Parser) literal() (Expression, error) {
	if !p.match(token.Integer) {
		return nil, fmt.Errorf("Expected %v. got: %v", token.Integer, p.peek().TokenType)
	}
	intVal := p.previous().Value.(int)
	return IntegerLiteral{intVal}, nil
}

func (p *Parser) match(types ...token.TokenType) bool {
	if slices.ContainsFunc(types, p.checkType) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) checkType(tokenType token.TokenType) bool {
	return !p.isEOF() && p.peek().TokenType == tokenType
}

func (p *Parser) advance() token.Token {
	if !p.isEOF() {
		p.current++	
	}
	return p.previous()
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) isEOF() bool {
	return p.peek().TokenType == token.EOF
}