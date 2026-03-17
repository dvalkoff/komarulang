package parser

import (
	"errors"
	"fmt"
	"slices"

	"github.com/dvalkoff/komarulang/parser/types"
	token "github.com/dvalkoff/komarulang/tokenizer"
)

type Expression interface {
	Type() types.Type
}

type Statement interface {
	Statement()
}

type FunctionDecl struct {
	Name       string
	Arguments  []*FunctionArgument
	ReturnType types.Type
	Body       Statement
	ReturnStmtsCount int
	EpilogueLabel Label
}

func (d *FunctionDecl) Statement() {}

type FunctionArgument struct {
	VarType       types.Type
	Identifier    string
}

type FunctionCall struct {
	ReturnType types.Type
	Name      string
	Arguments []Expression
}

func (d *FunctionCall) Type() types.Type {
	return d.ReturnType
}

func (c *FunctionCall) Statement() {}

type ReturnStatement struct {
	ReturnType    types.Type
	Expression    Expression
	EpilogueLabel Label
}

func (c *ReturnStatement) Statement() {}

func (e *ReturnStatement) Type() types.Type {
	return e.ReturnType
}

type BreakStatement struct {
	GotoLabel Label
}

func (c *BreakStatement) Statement() {}

type ContinueStatement struct {
	GotoLabel Label
}

func (c *ContinueStatement) Statement() {}

const (
	UndefinedLabelType LabelType = ""
	EndIfType          LabelType = "end_if"
	ElseType           LabelType = "else"
	WhileLoop          LabelType = "while_loop"
	WhileLoopEnd       LabelType = "while_loop_end"
	ForLoop            LabelType = "for_loop"
	ForLoopEnd         LabelType = "for_loop_end"
	LoopStart          LabelType = "loop_start"
	LoopLabel          LabelType = "loop_label"
	LoopEnd            LabelType = "loop_end"
	IncrementLabel     LabelType = "increment_for_loop"
	FunctionEpilogue   LabelType = "fun_epilogue"
)

type LabelType string

var labelCounter = 0

type Label struct {
	LabelType LabelType
	Index     int
}

func NewLabel(labelType LabelType) Label {
	index := labelCounter
	labelCounter++
	return Label{
		LabelType: labelType,
		Index:     index,
	}
}

func (l Label) String() string {
	return fmt.Sprintf("%v_%v", l.LabelType, l.Index)
}

type ForStatement struct {
	LabelStart     Label
	LabelIncrement Label
	LabelEnd       Label
	VarDecl        Statement
	Condition      Expression
	Increment      Statement
	Block          Statement
}

func (s *ForStatement) Statement() {}

type WhileStatement struct {
	LabelStart Label
	LabelEnd   Label
	Condition  Expression
	Block      Statement
}

func (s *WhileStatement) Statement() {}

type VarDeclaration struct {
	VarType    types.Type
	Identifier string
	Expr       Expression
}

func (d *VarDeclaration) Statement() {}

type VarAssignment struct {
	Identifier string
	Expr       Expression
}

func (d *VarAssignment) Statement() {}

type ExprStatement struct {
	Expr Expression
}

func (s *ExprStatement) Statement() {}

type PrintStatement struct {
	Expr Expression
}

func (s *PrintStatement) Statement() {}

type Block struct {
	Stmts []Statement
}

func (s *Block) Statement() {}

type IfStatement struct {
	Condition Expression
	Block     Statement
	ElseBlock Statement
}

func (i *IfStatement) Statement() {}

type IntegerLiteral struct {
	Value int
}

func (e *IntegerLiteral) Type() types.Type {
	return types.IntType
}

type BooleanLiteral struct {
	Value bool
}

func (e *BooleanLiteral) Type() types.Type {
	return types.BoolType
}

type IdentifierLiteral struct {
	VarType types.Type
	Value string
}

func (e *IdentifierLiteral) Type() types.Type {
	return e.VarType
}

type VoidLiteral struct {
}

func (e *VoidLiteral) Type() types.Type {
	return types.VoidType
}

type BinaryExpression struct {
	ExprType types.Type
	Left     Expression
	Operator token.TokenType
	Right    Expression
}

func (e *BinaryExpression) Type() types.Type {
	return e.ExprType
}

type UnaryExpression struct {
	ExprType types.Type
	Operator token.TokenType
	Right    Expression
}

func (e *UnaryExpression) Type() types.Type {
	return e.ExprType
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
		if p.match(token.LeftParen) {
			p.rewind()
			p.rewind()
			stmt, err = p.statement()
		} else {
			stmt, err = p.assignment()
		}
	case p.match(token.LeftBrace):
		stmt, err = p.block()
	case p.match(token.If):
		stmt, err = p.ifStatement()
	case p.match(token.While):
		stmt, err = p.whileStatement()
	case p.match(token.For):
		stmt, err = p.forStatement()
	case p.match(token.Fun):
		stmt, err = p.funcDeclaration()
	case p.match(token.Return):
		stmt, err = p.returnStatement()
	case p.match(token.Break):
		stmt = &BreakStatement{}
	case p.match(token.Continue):
		stmt = &ContinueStatement{}
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

func (p *Parser) returnStatement() (*ReturnStatement, error) {
	if p.peek().TokenType != token.Semicolon {
		expression, err := p.expression()
		if err != nil {
			return nil, err
		}
		return &ReturnStatement{ReturnType: types.NotSpecified, Expression: expression}, nil
	}
	return &ReturnStatement{ReturnType: types.VoidType, Expression: &VoidLiteral{}}, nil
}

func (p *Parser) funcDeclaration() (*FunctionDecl, error) {
	if !p.match(token.Identifier) {
		return nil, ParserError{Expected: token.Identifier, Got: p.peek()}
	}
	functionName := p.previous().Value.(string)
	if err := p.consume(token.LeftParen); err != nil {
		return nil, err
	}
	funArguments := make([]*FunctionArgument, 0)
	if !p.isEOF() && p.peek().TokenType != token.RightParen {
		if !p.match(token.Identifier) {
			return nil, ParserError{Expected: token.Identifier, Got: p.peek()}
		}
		paramName := p.previous().Value.(string)
		if !p.match(token.Type) {
			return nil, ParserError{Expected: token.Identifier, Got: p.peek()}
		}
		paramType, err := types.FromToken(p.previous())
		if err != nil {
			return nil, err
		}
		funArguments = append(funArguments, &FunctionArgument{
			VarType:    paramType,
			Identifier: paramName,
		})
	}
	for !p.isEOF() && p.peek().TokenType != token.RightParen {
		if err := p.consume(token.Comma); err != nil {
			return nil, ParserError{Expected: token.Comma, Got: p.peek()}
		}
		if !p.match(token.Identifier) {
			return nil, ParserError{Expected: token.Identifier, Got: p.peek()}
		}
		paramName := p.previous().Value.(string)
		if !p.match(token.Type) {
			return nil, ParserError{Expected: token.Identifier, Got: p.peek()}
		}
		paramType, err := types.FromToken(p.previous())
		if err != nil {
			return nil, err
		}
		funArguments = append(funArguments, &FunctionArgument{
			VarType:    paramType,
			Identifier: paramName,
		})
	}
	if p.isEOF() {
		return nil, ParserError{Expected: token.RightParen, Got: p.peek()}
	}
	p.consume(token.RightParen)

	returnType := types.VoidType
	if p.match(token.Type) {
		var err error
		returnType, err = types.FromToken(p.previous())
		if err != nil {
			return nil, err
		}
	}

	if !p.match(token.LeftBrace) {
		return nil, ParserError{Expected: token.LeftBrace, Got: p.peek()}
	}
	funBlock, err := p.block()
	if err != nil {
		return nil, err
	}
	return &FunctionDecl{
		Name:       functionName,
		Arguments:  funArguments,
		ReturnType: returnType,
		Body:       funBlock,
	}, nil
}

func (p *Parser) block() (*Block, error) {
	stmts := make([]Statement, 0)
	for !p.isEOF() && p.peek().TokenType != token.RightBrace {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}

	if p.isEOF() {
		return nil, ParserError{Expected: token.RightBrace, Got: p.peek()}
	}
	err := p.consume(token.RightBrace)
	if err != nil {
		return nil, err
	}
	return &Block{Stmts: stmts}, nil
}

func (p *Parser) ifStatement() (*IfStatement, error) {
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.match(token.LeftBrace) {
		return nil, ParserError{Expected: token.LeftBrace, Got: p.peek()}
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
				ParserError{Expected: token.LeftBrace, Got: p.peek()},
				ParserError{Expected: token.If, Got: p.peek()},
			)
		}
		if err != nil {
			return nil, err
		}
	}
	return &IfStatement{Condition: condition, Block: block, ElseBlock: elseBlock}, nil
}

func (p *Parser) forStatement() (*ForStatement, error) {
	var varDecl Statement
	var err error
	switch {
	case p.match(token.Var):
		varDecl, err = p.varDeclaration()
	case p.match(token.Identifier):
		varDecl, err = p.assignment()
	default:
		return nil, errors.Join(
			ParserError{Expected: token.Var, Got: p.peek()},
			ParserError{Expected: token.Identifier, Got: p.peek()},
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
		return nil, ParserError{Expected: token.Identifier, Got: p.peek()}
	}
	increment, err := p.assignment()
	if err != nil {
		return nil, err
	}
	if !p.match(token.LeftBrace) {
		return nil, ParserError{Expected: token.LeftBrace, Got: p.peek()}
	}
	block, err := p.block()
	if err != nil {
		return nil, err
	}
	return &ForStatement{
		VarDecl:   varDecl,
		Condition: condition,
		Increment: increment,
		Block:     block,
	}, nil
}

func (p *Parser) whileStatement() (*WhileStatement, error) {
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if !p.match(token.LeftBrace) {
		return nil, ParserError{Expected: token.LeftBrace, Got: p.peek()}
	}
	block, err := p.block()
	if err != nil {
		return nil, err
	}
	return &WhileStatement{Condition: condition, Block: block}, nil
}

func (p *Parser) varDeclaration() (*VarDeclaration, error) {
	if !p.match(token.Identifier) {
		return nil, ParserError{Expected: token.Identifier, Got: p.peek()}
	}
	identifier := p.previous().Value.(string)

	varType := types.NotSpecified
	if p.match(token.Type) {
		var err error
		varType, err = types.FromToken(p.previous())
		if err != nil {
			return nil, err
		}
	}

	err := p.consume(token.Equal)
	if err != nil {
		return nil, err
	}
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &VarDeclaration{VarType: varType, Identifier: identifier, Expr: expr}, nil
}

func (p *Parser) assignment() (*VarAssignment, error) {
	identifier := p.previous().Value.(string)
	err := p.consume(token.Equal)
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &VarAssignment{Identifier: identifier, Expr: expr}, nil
}

func (p *Parser) statement() (Statement, error) {
	if p.match(token.Print) {
		return p.printStatement()
	}
	expression, err := p.expression()
	if err != nil {
		return nil, err
	}
	return &ExprStatement{Expr: expression}, nil
}

func (p *Parser) printStatement() (*PrintStatement, error) {
	if !p.match(token.LeftParen) {
		return nil, ParserError{Expected: token.LeftParen, Got: p.peek()}
	}
	expression, err := p.expression()
	if err != nil {
		return nil, err
	}
	if err := p.consume(token.RightParen); err != nil {
		return nil, err
	}
	return &PrintStatement{Expr: expression}, nil
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
		expression = &BinaryExpression{ExprType: types.BoolType, Left: expression, Operator: operator, Right: right}
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
		expression = &BinaryExpression{ExprType: types.BoolType, Left: expression, Operator: operator, Right: right}
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
		expression = &BinaryExpression{ExprType: types.BoolType, Left: expression, Operator: operator, Right: right}
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
		expression = &BinaryExpression{ExprType: types.IntType, Left: expression, Operator: operator, Right: right}
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
		expression = &BinaryExpression{ExprType: types.IntType, Left: expression, Operator: operator, Right: right}
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
		expression = &BinaryExpression{ExprType: types.IntType, Left: expression, Operator: operator, Right: right}
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
		expression = &BinaryExpression{ExprType: types.IntType, Left: expression, Operator: operator, Right: right}
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
		expression = &BinaryExpression{ExprType: types.IntType, Left: expression, Operator: operator, Right: right}
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
		varType := types.NotSpecified
		switch operator {
		case token.Minus:
			varType = types.IntType
		case token.Bang:
			varType = types.BoolType
		}
		return &UnaryExpression{ExprType: varType, Operator: operator, Right: right}, nil
	}
	return p.primary()
}

func (p *Parser) primary() (Expression, error) {
	if p.match(token.Integer) {
		intVal := p.previous().Value.(int)
		return &IntegerLiteral{intVal}, nil
	}
	if p.match(token.Bool) {
		boolVal := p.previous().Value.(bool)
		return &BooleanLiteral{boolVal}, nil
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
		if p.match(token.LeftParen) {
			return p.evaluateFunCall(identifier)
		}
		return &IdentifierLiteral{Value: identifier}, nil
	}
	return nil, ParserError{Expected: token.Integer, Got: p.peek()}
}

func (p *Parser) evaluateFunCall(funName string) (*FunctionCall, error) {
	parameters := make([]Expression, 0)
	if !p.isEOF() && p.peek().TokenType != token.RightParen {
		param, err := p.expression()
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, param)
	}

	for !p.isEOF() && p.peek().TokenType != token.RightParen {
		if err := p.consume(token.Comma); err != nil {
			return nil, err
		}
		param, err := p.expression()
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, param)
	}
	if p.isEOF() {
		return nil, ParserError{Expected: token.RightParen, Got: p.peek()}
	}
	p.consume(token.RightParen)
	return &FunctionCall{
		Name:      funName,
		Arguments: parameters,
	}, nil
}

func (p *Parser) consume(tokenType token.TokenType) error {
	if p.checkType(tokenType) {
		p.advance()
		return nil
	}
	return ParserError{Expected: tokenType, Got: p.peek()}
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

func (p *Parser) rewind() token.Token {
	if p.current == 0 {
		return p.tokens[0]
	}
	p.current--
	return p.peek()
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
