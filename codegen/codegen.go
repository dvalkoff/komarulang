package codegen

import (
	"fmt"
	"strings"

	"github.com/dvalkoff/komarulang/parser"
	token "github.com/dvalkoff/komarulang/tokenizer"
)

type Program struct {
	Instructions []Instruction
}

func (p *Program) Emit(i ...Instruction) {
	p.Instructions = append(p.Instructions, i...)
}

func (p Program) String() string {
	var b strings.Builder

	b.WriteString(`.global _main
.align 2

_main:
`)

	for _, inst := range p.Instructions {
		b.WriteString(inst.String())
		b.WriteByte('\n')
	}

	b.WriteString(`
	mov x16, #1
	svc #0x80
`)

	return b.String()
}

type CodeGenerator struct {
	Prog *Program
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		Prog: &Program{},
	}
}

func (c *CodeGenerator) CompileStmt(stmt parser.Statement, reg Register) (Register, error) {
	switch e := stmt.(type) {
	case *parser.ExprStatement:
		return c.compileExpr(e.Expr, reg)
	}
	return 0, fmt.Errorf("Unexpected statement %v", stmt)
}

func (c *CodeGenerator) compileExpr(expr parser.Expression, reg Register) (Register, error) {
	switch e := expr.(type) {
	case *parser.IntegerLiteral:
		instructions := c.loadInt(reg, e.Value)
		c.Prog.Emit(instructions...)
		return reg, nil
	case *parser.BooleanLiteral:
		if e.Value {
			c.Prog.Emit(Mov{reg, TrueImm})
		} else {
			c.Prog.Emit(Mov{reg, FalseImm})
		}
		return reg, nil
	case *parser.UnaryExpression:
		left := reg
		right, err := c.compileExpr(e.Right, reg+1)
		if err != nil {
			return 0, err
		}
		switch e.Operator {
		case token.Minus:
			c.Prog.Emit(Neg{
				Dst: left,
				A:   right,
			})
		case token.Bang:
			c.Prog.Emit(Mov{
				Dst: left,
				Src: TrueImm,
			})
			c.Prog.Emit(BitwiseXor{
				BinaryOperation{
					Dst: left,
					A: left,
					B: right,
				},
			})
		return reg, nil
		}
	case *parser.BinaryExpression:
		left, err := c.compileExpr(e.Left, reg)
		if err != nil {
			return 0, err
		}
		right, err := c.compileExpr(e.Right, reg+1)
		if err != nil {
			return 0, err
		}
		switch e.Operator {
		case token.Plus:
			c.Prog.Emit(Add{
				BinaryOperation{
					Dst: left,
					A:   left,
					B:   right,
				},
			})
		case token.Minus:
			c.Prog.Emit(Sub{
				BinaryOperation{
					Dst: left,
					A:   left,
					B:   right,
				},
			})
		case token.Star:
			c.Prog.Emit(Mul{
				BinaryOperation{
					Dst: left,
					A:   left,
					B:   right,
				},
			})
		case token.Slash:
			c.Prog.Emit(Sdiv{
				BinaryOperation{
					Dst: left,
					A:   left,
					B:   right,
				},
			})
		case token.Percent:
			c.Prog.Emit(Sdiv{
				BinaryOperation{
					Dst: right + 1,
					A: left,
					B: right,
				},
			})
			c.Prog.Emit(MSub{
				Dst: left,
				A: right + 1,
				B: right,
				C: left,
			})
		case token.Vbar, token.VbarVbar:
			c.Prog.Emit(BitwiseOr{
				BinaryOperation{
					Dst: left,
					A: left,
					B: right,
				},
			})
		case token.Ampersand, token.AmpersandAmpersand:
			c.Prog.Emit(BitwiseAnd{
				BinaryOperation{
					Dst: left,
					A: left,
					B: right,
				},
			})
		case token.Caret:
			c.Prog.Emit(BitwiseXor{
				BinaryOperation{
					Dst: left,
					A: left,
					B: right,
				},
			})
		case token.EqualEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A: left,
				Value: CSET_EQ,
			})
		case token.BangEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A: left,
				Value: CSET_NE,
			})
		case token.GreaterEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A: left,
				Value: CSET_GE,
			})
		case token.Greater:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A: left,
				Value: CSET_GT,
			})
		case token.LessEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A: left,
				Value: CSET_LE,
			})
		case token.Less:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A: left,
				Value: CSET_LT,
			})
		}
		return reg, nil
	}

	return 0, fmt.Errorf("Unexpected expression %v", expr)
}

func (c *CodeGenerator) loadInt(reg Register, value int) []Instruction {
	if value >= 0 && value <= 65535 {
		return []Instruction{
			Mov{
				Dst: reg,
				Src: Imm(value),
			},
		}
	}
	instructions := []Instruction{}

	shift := 0
	chunk := (value >> shift) & 0xFFFF
	instructions = append(instructions,
		Movz{
			Dst: reg,
			Src: Imm(chunk),
			Lsl: Imm(shift),
		},
	)
	for shift = 16; shift < 64; shift += 16 {
		chunk := (value >> shift) & 0xFFFF
		if chunk == 0 {
			continue
		}
		instructions = append(instructions,
			Movk{
				Dst: reg,
				Src: Imm(chunk),
				Lsl: Imm(shift),
			},
		)
	}
	return instructions
}
