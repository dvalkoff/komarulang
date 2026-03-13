package parser

import (
	"errors"
	"slices"

	token "github.com/dvalkoff/komarulang/tokenizer"
)

type Expression any

type Statement interface{
	Statement()
}

type ForStatement struct {
	VarDecl Statement
	Condition Expression
	Increment Statement
	Block Statement
}

func (s ForStatement) Statement() {}

type WhileStatement struct {
	Condition Expression
	Block Statement
}

func (s WhileStatement) Statement() {}

type VarDeclaration struct {
	Identifier string
	Expr Expression
}

func (d VarDeclaration) Statement() {}

type VarAssignment struct {
	Identifier string
	Expr Expression
}

func (d VarAssignment) Statement() {}


type ExprStatement struct {
	Expr Expression
}

func (s ExprStatement) Statement() {}

type PrintStatement struct {
	Expr Expression
}

func (s PrintStatement) Statement() {}

type Block struct {
	Stmts []Statement
}

func (s Block) Statement() {}

type IfStatement struct {
	Condition Expression
	Block Statement
	ElseBlock Statement
}

func (i IfStatement) Statement() {}

type IntegerLiteral struct {
	Value int
}

type BooleanLiteral struct {
	Value bool
}

type IdentifierLiteral struct {
	Value string
}

type BinaryExpression struct {
	Left     Expression
	Operator token.TokenType
	Right    Expression
}

type UnaryExpression struct {
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


func (p *Parser) Parse() ([]Statement, error) {
	declarations := make([]Statement, 0)
	for !p.isEOF() {
		statement, err := p.declaration()
		if err != nil {
			return nil, err
		}
		declarations = append(declarations, statement)
	}
	return declarations, nil
}

func (p *Parser) declaration() (Statement, error) {
	var stmt Statement
	var err error
	switch {
	case p.match(token.Var):
		stmt, err = p.varDeclaration()
	case p.match(token.Identifier):
		stmt, err = p.assignment()
	case p.match(token.LeftBrace):
		stmt, err = p.block()
	case p.match(token.If):
		stmt, err = p.ifStatement()
	case p.match(token.While):
		stmt, err = p.whileStatement()
	case p.match(token.For):
		stmt, err = p.forStatement()
	default:
		stmt, err = p.statement()
	}
	if err != nil {
		return nil, err
	}
	err = p.consume(token.Semicolon)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) block() (Statement, error) {
	stmts := make([]Statement, 0)
	for !p.isEOF() && p.peek().TokenType != token.RightBrace {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}

	if p.isEOF() {
		return nil, ParserError{Expected: token.RightBrace, Got: token.EOF}
	}
	err := p.consume(token.RightBrace)
	if err != nil {
		return nil, err
	}
	return Block{Stmts: stmts}, nil
}

func (p *Parser) ifStatement() (Statement, error) {
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.match(token.LeftBrace) {
		return nil, ParserError{Expected: token.LeftBrace, Got: p.peek().TokenType}
	}
	block, err := p.block()
	if err != nil {
		return nil, err
	}
	var elseBlock Statement = nil 
	if p.match(token.Else) {
		if p.match(token.If) {
			elseBlock, err = p.ifStatement()
		} else if p.match(token.LeftBrace) {
			elseBlock, err = p.block()
		} else {
			err = errors.Join(
				ParserError{Expected: token.LeftBrace, Got: p.peek().TokenType},
				ParserError{Expected: token.If, Got: p.peek().TokenType},
			)
		}
		if err != nil {
			return nil, err
		}
	}
	return IfStatement{Condition: condition, Block: block, ElseBlock: elseBlock}, nil
}

func (p *Parser) forStatement() (Statement, error) {
	var varDecl Statement
	var err error
	switch {
	case p.match(token.Var):
		varDecl, err = p.varDeclaration()
	case p.match(token.Identifier):
		varDecl, err = p.assignment()
	default:
		return nil, errors.Join(
			ParserError{Expected: token.Var, Got: p.peek().TokenType},
			ParserError{Expected: token.Identifier, Got: p.peek().TokenType},
		)
	}
	if err != nil {
		return nil, err
	}
	if err = p.consume(token.Semicolon); err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if err = p.consume(token.Semicolon); err != nil {
		return nil, err
	}

	if !p.match(token.Identifier) {
		return nil, ParserError{Expected: token.Identifier, Got: p.peek().TokenType}
	}
	increment, err := p.assignment()
	if err != nil {
		return nil, err
	}
	if !p.match(token.LeftBrace) {
		return nil, ParserError{Expected: token.LeftBrace, Got: p.peek().TokenType}
	}
	block, err := p.block()
	if err != nil {
		return nil, err
	}
	return ForStatement{
		VarDecl: varDecl,
		Condition: condition,
		Increment: increment,
		Block: block,
	}, nil
}

func (p *Parser) whileStatement() (Statement, error) {
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.match(token.LeftBrace) {
		return nil, ParserError{Expected: token.LeftBrace, Got: p.peek().TokenType}
	}
	block, err := p.block()
	if err != nil {
		return nil, err
	}
	return WhileStatement{Condition: condition, Block: block}, nil
}

func (p *Parser) varDeclaration() (Statement, error) {
	if !p.match(token.Identifier) {
		return nil, ParserError{Expected: token.Identifier, Got: p.peek().TokenType}
	}
	identifier := p.previous().Value.(string)
	err := p.consume(token.Equal)
	if err != nil {
		return nil, err
	}
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return VarDeclaration{Identifier: identifier, Expr: expr}, nil
}

func (p *Parser) assignment() (Statement, error) {
	identifier := p.previous().Value.(string)
	err := p.consume(token.Equal)
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return VarAssignment{Identifier: identifier, Expr: expr}, nil
}

func (p *Parser) statement() (Statement, error) {
	if p.match(token.Print) {
		return p.printStatement()
	}
	expression, err := p.expression();
	if err != nil {
		return nil, err
	}
	return ExprStatement{Expr: expression}, nil
}

func (p *Parser) printStatement() (Statement, error) {
	if !p.match(token.LeftParen) {
		return nil, ParserError{Expected: token.LeftParen, Got: p.peek().TokenType}
	}
	expression, err := p.expression()
	if err != nil {
		return nil, err
	}
	if err := p.consume(token.RightParen); err != nil {
		return nil, err
	}
	return PrintStatement{Expr: expression}, nil
}

func (p *Parser) expression() (Expression, error) {
	return p.logicalOr()
}

func (p *Parser) logicalOr() (Expression, error) {
	expression, err := p.logicalAnd()
	if err != nil {
		return nil, err
	}

	for p.match(token.VbarVbar) {
		operator := p.previous().TokenType
		right, err := p.logicalAnd()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (p *Parser) logicalAnd() (Expression, error) {
	expression, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(token.AmpersandAmpersand) {
		operator := p.previous().TokenType
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (p *Parser) comparison() (Expression, error) {
	expression, err := p.bitwiseOR()
	if err != nil {
		return nil, err
	}

	for p.match(token.EqualEqual, token.BangEqual, token.Less, token.LessEqual, token.Greater, token.GreaterEqual) {
		operator := p.previous().TokenType
		right, err := p.bitwiseOR()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}

func (p *Parser) bitwiseOR() (Expression, error) {
	expression, err := p.bitwiseXOR()
	if err != nil {
		return nil, err
	}

	for p.match(token.Vbar) {
		operator := p.previous().TokenType
		right, err := p.bitwiseXOR()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}
func (p *Parser) bitwiseXOR() (Expression, error) {
	expression, err := p.bitwiseAND()
	if err != nil {
		return nil, err
	}

	for p.match(token.Caret) {
		operator := p.previous().TokenType
		right, err := p.bitwiseAND()
		if err != nil {
			return nil, err
		}
		expression = BinaryExpression{Left: expression, Operator: operator, Right: right}
	}
	return expression, nil
}
func (p *Parser) bitwiseAND() (Expression, error) {
	expression, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(token.Ampersand) {
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
		expression, err := p.expression()
		if err != nil {
			return nil, err
		}
		if err := p.consume(token.RightParen); err != nil {
			return nil, err
		}
		return expression, nil
	}
	if p.match(token.Identifier) {
		identifier := p.previous().Value.(string)
		return IdentifierLiteral{Value: identifier}, nil
	}
	return nil, ParserError{Expected: token.Integer, Got: p.peek().TokenType}
}

func (p *Parser) consume(tokenType token.TokenType) error {
	if p.checkType(tokenType) {
		p.advance()
		return nil
	}
	return ParserError{Expected: tokenType, Got: p.peek().TokenType}
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
