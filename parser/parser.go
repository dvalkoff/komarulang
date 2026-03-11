package parser

import (
	"slices"
	"fmt"

	token "github.com/dvalkoff/komarulang/tokenizer"
)

type Expression any

type IntegerLiteral struct {
	Value int
}

type BooleanLiteral struct {
	Value bool
}

type BinaryExpression struct {
	Left     Expression
	Operator token.TokenType
	Right    Expression
}

type UnaryExpression struct {
	Operator token.TokenType
	Right Expression
}

type Parser struct {
	tokens  []token.Token
	current int
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

func (p *Parser) Expression() (Expression, error) {
	return p.comparison()
}

func (p *Parser) comparison() (Expression, error) {
	expression, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(token.EqualEqual, token.BangEqual, token.Less, token.LessEqual, token.Greater, token.GreaterEqual) {
		operator := p.previous().TokenType
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (p *Parser) term() (Expression, error) {
	expression, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(token.Plus, token.Minus) {
		operator := p.previous().TokenType
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (p *Parser) factor() (Expression, error) {
	expression, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(token.Slash, token.Star, token.Percent) {
		operator := p.previous().TokenType
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (p *Parser) unary() (Expression, error) {
	if p.match(token.Minus, token.Bang) {
		operator := p.previous().TokenType
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return UnaryExpression{Operator: operator, Right: right}, nil
	}
	return p.primary()
}

func (p *Parser) primary() (Expression, error) {
	if p.match(token.Integer) {
		intVal := p.previous().Value.(int)
		return IntegerLiteral{intVal}, nil
	}
	if p.match(token.Bool) {
		boolVal := p.previous().Value.(bool)
		return BooleanLiteral{boolVal}, nil
	}
	if p.match(token.LeftParen) {
		expression, err := p.Expression()
		if err != nil {
			return nil, err
		}
		if err := p.consume(token.RightParen); err != nil {
			return nil, err
		}
		return expression, nil
	}
 	return nil, fmt.Errorf("Expected %v. got: %v", token.Integer, p.peek().TokenType)
}

func (p *Parser) consume(tokenType token.TokenType) error {
	if p.checkType(tokenType) {
		p.advance()
		return nil
	}
	return fmt.Errorf("Expected %v. got %v", tokenType, p.peek().TokenType)
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