package codegen

import (
	"fmt"
	"strings"

	"github.com/dvalkoff/komarulang/parser"
	token "github.com/dvalkoff/komarulang/tokenizer"
)

type IdentifierOffsets = map[string]int

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

`)

	b.WriteString(`
// x0 comes as a argument
_print_int:

    sub sp, sp, #32 // allocating 32 bytes for int64
    mov x5, #19 // sp pointer to current char
    str x30, [sp, #0]    // save x30

    mov x4, #10       // newline ASCII
    add x6, sp, x5
    strb w4, [x6]
    sub x5, x5, #1

    // at start of _print_int
    mov x10, #0
    cmp x0, #0
    B.GE skip_negative    // if positive, skip
    neg x0, x0            // make positive
    mov x10, #1  // store '-' character flag
    skip_negative:

	cmp x0, #0
	B.NE main_algo    // if not zero, go to loop
	// x0 IS zero — handle special case
	mov x4, #48                // '0' ASCII
	add x6, sp, x5
	strb w4, [x6]              // store '0'
	sub x5, x5, #1
	B while_num_non_zero_end

	main_algo:


    while_num_non_zero:
    cmp x0, #0
    B.EQ while_num_non_zero_end // if zero, goto end of loop
    mov x1, #10 // setting a divider into a register
    sdiv x2, x0, x1 // 2-step modulo
    msub x4, x2, x1, x0
    add x4, x4, #48   // shifting for ASCII. '0' = 48 in ASCII
    add x6, sp, x5     // x6 = sp + x5 (actual address)
    strb w4, [x6]       // store at that address
    sub x5, x5, #1 // decreasing a pointer
    sdiv x0, x0, x1 // dividing by 10
    B while_num_non_zero // returning to a loop starting point

    while_num_non_zero_end:

    cmp x10, #0
    B.EQ skip_sign
    mov x4, #45       // newline ASCII
    add x6, sp, x5
    strb w4, [x6]
    sub x5, x5, #1

    skip_sign:

    mov x16, #4      // write
    mov x0, #1       // stdout
    add x1, sp, x5      // x1 = address of string
    add x1, x1, #1      // x1 = address of string
    mov x4, #19
    sub x2, x4, x5 // calculating and setting len

    svc #0x80
    ldr x30, [sp, #0]   // restore x30
    add sp, sp, #32 // deallocating stack
    ret
`)

	b.WriteString(`
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

func (c *CodeGenerator) Compile(stmts []parser.Statement) error {
	block := &parser.Block{Stmts: stmts}
	_, err := c.compileStmt(nil, block, 0)
	return err
}

func (c *CodeGenerator) compileStmt(env *Offsets, stmt parser.Statement, reg Register) (Register, error) {
	switch typed := stmt.(type) {
	case *parser.ExprStatement:
		return c.compileExpr(env, typed.Expr, reg)
	case *parser.Block:
		return c.compileBlock(env, typed, reg)
	case *parser.VarDeclaration:
		return c.compileVarDeclaration(env, typed, reg)
	case *parser.VarAssignment:
		return c.compileVarAssignment(env, typed, reg)
	case *parser.IfStatement:
		return c.compileIf(env, typed, reg)
	case *parser.WhileStatement:
		return c.compileWhile(env, typed, reg)
	case *parser.ForStatement:
		return c.compileFor(env, typed, reg)
	case *parser.PrintStatement:
		return c.compilePrint(env, typed, reg)
	}
	return 0, fmt.Errorf("Unexpected statement %v", stmt)
}

func (c *CodeGenerator) compilePrint(env *Offsets, printStmt *parser.PrintStatement, reg Register) (Register, error) {
	reg, err := c.compileExpr(env, printStmt.Expr, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(PrintSubroutine{})
	return reg, nil
}

func (c *CodeGenerator) compileFor(parent *Offsets, forStatement *parser.ForStatement, reg Register) (Register, error) {
	env := parent
	allocationRequired := false
	if varDecl, ok := forStatement.VarDecl.(*parser.VarDeclaration); ok {
		env = NewOffsets(parent)
		env.Put(varDecl)
		env.AlignStackSize()
		c.Prog.Emit(StackAllocator{
			Value: Imm(env.StackSize),
		})
		allocationRequired = true
	}

	reg, err := c.compileStmt(env, forStatement.VarDecl, reg)
	if err != nil {
		return reg, err
	}

	forLoopLabel := NewLabel(ForLoop)
	forLoopEndLabel := NewLabel(ForLoopEnd)
	c.Prog.Emit(forLoopLabel)
	reg, err = c.compileExpr(env, forStatement.Condition, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(Cbz{
		A: reg,
		Label: forLoopEndLabel,
	})
	reg, err = c.compileStmt(env, forStatement.Block, reg)
	if err != nil {
		return reg, err
	}
	reg, err = c.compileStmt(env, forStatement.Increment, reg)
	c.Prog.Emit(Bjump{
		Label: forLoopLabel,
	})
	c.Prog.Emit(forLoopEndLabel)


	if allocationRequired {
		c.Prog.Emit(StackDeallocator{
			Value: Imm(env.StackSize),
		})
	}
	return reg, nil
}

func (c *CodeGenerator) compileWhile(env *Offsets, whileStatement *parser.WhileStatement, reg Register) (Register, error) {
	whileLoopLabel := NewLabel(WhileLoop)
	whileLoopEndLabel := NewLabel(WhileLoopEnd)
	c.Prog.Emit(whileLoopLabel)
	reg, err := c.compileExpr(env, whileStatement.Condition, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(Cbz{
		A: reg,
		Label: whileLoopEndLabel,
	})
	reg, err = c.compileStmt(env, whileStatement.Block, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(Bjump{
		Label: whileLoopLabel,
	})
	c.Prog.Emit(whileLoopEndLabel)
	return reg, nil
}

func (c *CodeGenerator) compileIf(env *Offsets, ifStatement *parser.IfStatement, reg Register) (Register, error) {
	endIfLabel := NewLabel(EndIfType)
	elseLabel := NewLabel(ElseType)
	reg, err := c.compileExpr(env, ifStatement.Condition, reg)
	c.Prog.Emit(Cbz{
		A: reg,
		Label: elseLabel,
	})
	if err != nil {
		return reg, err
	}
	reg, err = c.compileStmt(env, ifStatement.Block, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(Bjump{
		Label: endIfLabel,
	})
	c.Prog.Emit(elseLabel)
	if ifStatement.ElseBlock != nil {
		reg, err = c.compileStmt(env, ifStatement.ElseBlock, reg)
	}
	c.Prog.Emit(endIfLabel)

	return reg, err
}

func (c *CodeGenerator) compileBlock(parent *Offsets, block *parser.Block, reg Register) (Register, error) {
	offsets := NewOffsets(parent)
	for _, stmt := range block.Stmts {
		if decl, ok := stmt.(*parser.VarDeclaration); ok {
			offsets.Put(decl)
		}
	}
	offsets.AlignStackSize()

	if offsets.StackSize > 0 {
		c.Prog.Emit(StackAllocator{
			Value: Imm(offsets.StackSize),
		})
	}

	for _, stmt := range block.Stmts {
		_, err := c.compileStmt(offsets, stmt, reg)
		if err != nil {
			return reg, err
		}
	}

	if offsets.StackSize > 0 {
		c.Prog.Emit(StackDeallocator{
			Value: Imm(offsets.StackSize),
		})
	}

	return reg, nil
}

func (c *CodeGenerator) compileVarDeclaration(offsets *Offsets, varDecl *parser.VarDeclaration, reg Register) (Register, error) {
	reg, err := c.compileExpr(offsets, varDecl.Expr, reg)
	if err != nil {
		return reg, err
	}

	c.Prog.Emit(Str{
		A: reg,
		Offset: Imm(offsets.Get(varDecl.Identifier)),
	})
	return reg, nil
}

func (c *CodeGenerator) compileVarAssignment(offsets *Offsets, varAssignment *parser.VarAssignment, reg Register) (Register, error) {
	reg, err := c.compileExpr(offsets, varAssignment.Expr, reg)
	if err != nil {
		return reg, err
	}

	c.Prog.Emit(Str{
		A: reg,
		Offset: Imm(offsets.Get(varAssignment.Identifier)),
	})
	return reg, nil
}

func (c *CodeGenerator) compileExpr(offsets *Offsets, expr parser.Expression, reg Register) (Register, error) {
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
	case *parser.IdentifierLiteral:
		c.Prog.Emit(Ldr{
			A: reg,
			Offset: Imm(offsets.Get(e.Value)),
		})
		return reg, nil
	case *parser.UnaryExpression:
		left := reg
		right, err := c.compileExpr(offsets, e.Right, reg+1)
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
		}
		return reg, nil
	case *parser.BinaryExpression:
		left, err := c.compileExpr(offsets, e.Left, reg)
		if err != nil {
			return 0, err
		}
		right, err := c.compileExpr(offsets,e.Right, reg+1)
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

	return 0, fmt.Errorf("Unexpected expression %v, %t", expr, expr)
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
