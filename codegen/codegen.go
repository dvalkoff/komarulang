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

func (c *CodeGenerator) CompileExpr(expr parser.Expression, reg Register) (Register, error) {
	switch e := expr.(type) {
	case parser.IntegerLiteral:
		instructions := c.loadInt(reg, e.Value)
		c.Prog.Emit(instructions...)
		return reg, nil
	case parser.UnaryExpression:
		left := reg
		right, err := c.CompileExpr(e.Right, reg+1)
		if err != nil {
			return 0, err
		}
		switch e.Operator {
		case token.Minus:
			c.Prog.Emit(Neg{
				Dst: left,
				A:   right,
			})
			return reg, nil
		}
	case parser.BinaryExpression:
		left, err := c.CompileExpr(e.Left, reg)
		if err != nil {
			return 0, err
		}
		right, err := c.CompileExpr(e.Right, reg+1)
		if err != nil {
			return 0, err
		}
		switch e.Operator {
		case token.Plus:
			c.Prog.Emit(Add{
				Dst: left,
				A:   left,
				B:   right,
			})
		case token.Minus:
			c.Prog.Emit(Sub{
				Dst: left,
				A:   left,
				B:   right,
			})
		case token.Star:
			c.Prog.Emit(Mul{
				Dst: left,
				A:   left,
				B:   right,
			})
		case token.Slash:
			c.Prog.Emit(Sdiv{
				Dst: left,
				A:   left,
				B:   right,
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
