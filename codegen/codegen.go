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

	for _, inst := range p.Instructions {
		b.WriteString(inst.String())
		b.WriteByte('\n')
	}

	return b.String()
}

type CodeGenerator struct {
	Prog                 *Program
	EntrypointIdentifier string
}

func NewCodeGenerator(entrypoint string) *CodeGenerator {
	return &CodeGenerator{
		Prog:                 &Program{},
		EntrypointIdentifier: entrypoint,
	}
}

func (c *CodeGenerator) Compile(stmts []parser.Statement) error {
	c.Prog.Emit(
		Global{
			NewSubroutineDecl(c.EntrypointIdentifier),
		},
	)
	c.Prog.Emit(Align{2})

	c.Prog.Emit(PrintAsmSubroutine{})
	var entrypoint *parser.FunctionDecl
	for _, stmt := range stmts {
		if decl, ok := stmt.(*parser.FunctionDecl); ok {
			if decl.Name == c.EntrypointIdentifier {
				entrypoint = decl
			} else {
				_, err := c.compileStmt(nil, decl, 0)
				if err != nil {
					return err
				}
			}
		}
	}
	_, err := c.compileStmt(nil, entrypoint, 0)
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
	case *parser.BreakStatement:
		return c.compileBreak(env, typed, reg)
	case *parser.ContinueStatement:
		return c.compileContinue(env, typed, reg)
	case *parser.FunctionDecl:
		return c.compileFunction(env, typed, reg)
	case *parser.ReturnStatement:
		return c.compileReturnStmt(env, typed, reg)
	}
	return 0, fmt.Errorf("Unexpected statement %v", stmt)
}

func (c *CodeGenerator) compileFunction(parent *Offsets, funcDecl *parser.FunctionDecl, reg Register) (Register, error) {
	subroutineDecl := NewSubroutineDecl(funcDecl.Name)
	c.Prog.Emit(subroutineDecl)

	offsets := NewOffsets(parent)
	offsets.StackSize += 8 // x30
	for _, arg := range funcDecl.Arguments {
		offsets.PutFunArg(arg)
	}
	block := funcDecl.Body.(*parser.Block)
	for _, stmt := range block.Stmts {
		if decl, ok := stmt.(*parser.VarDeclaration); ok {
			offsets.Put(decl)
		}
	}
	offsets.AlignStackSize()

	c.Prog.Emit(StackAllocator{
		Value: Imm(offsets.StackSize),
	})
	c.Prog.Emit(Str{
		A:      30,
		Offset: Imm(0),
	})
	argumentRegister := Register(0)
	for _, arg := range funcDecl.Arguments {
		err := c.compileFunArgument(offsets, arg, argumentRegister)
		if err != nil {
			return reg, err
		}
		argumentRegister += 1
	}
	reg = Register(0)
	for _, stmt := range block.Stmts {
		_, err := c.compileStmt(offsets, stmt, reg)
		if err != nil {
			return reg, err
		}
	}

	epilogueLabel := AsmLabel{funcDecl.EpilogueLabel.String()}
	c.Prog.Emit(epilogueLabel)
	c.Prog.Emit(Ldr{
		A:      30,
		Offset: Imm(0),
	})
	c.Prog.Emit(StackDeallocator{
		Value: Imm(offsets.StackSize),
	})
	if funcDecl.Name == c.EntrypointIdentifier {
		return c.compileExit(offsets, reg)
	}
	c.Prog.Emit(AsmReturn{})
	return reg, nil
}

func (c *CodeGenerator) compileFunArgument(env *Offsets, funArg *parser.FunctionArgument, reg Register) error {
	c.Prog.Emit(Str{
		A: reg,
		Offset: Imm(env.Get(funArg.Identifier)),
	})
	return nil
}

func (c *CodeGenerator) compileExit(env *Offsets, reg Register) (Register, error) {
	c.Prog.Emit(Mov{
		Dst: 0,
		Src: Imm(0),
	})
	c.Prog.Emit(Mov{
		Dst: 16,
		Src: Imm(1),
	})
	c.Prog.Emit(Svc{
		Value: "#0x80",
	})
	return reg, nil
}

func (c *CodeGenerator) compileReturnStmt(env *Offsets, returnStmt *parser.ReturnStatement, reg Register) (Register, error) {
	reg, err := c.compileExpr(env, returnStmt.Expression, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(Mov{
		0,
		reg,
	})
	epilogueLabel := AsmLabel{returnStmt.EpilogueLabel.String()}
	c.Prog.Emit(Bjump{
		Label: epilogueLabel,
	})
	return reg, nil
}

func (c *CodeGenerator) compilePrint(env *Offsets, printStmt *parser.PrintStatement, reg Register) (Register, error) {
	reg, err := c.compileExpr(env, printStmt.Expr, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(CallPrintSubroutine{})
	return reg, nil
}

func (c *CodeGenerator) compileBreak(env *Offsets, breakStmt *parser.BreakStatement, reg Register) (Register, error) {
	c.Prog.Emit(Bjump{
		Label: AsmLabel{breakStmt.GotoLabel.String()},
	})
	return reg, nil
}

func (c *CodeGenerator) compileContinue(env *Offsets, continueStmt *parser.ContinueStatement, reg Register) (Register, error) {
	c.Prog.Emit(Bjump{
		Label: AsmLabel{continueStmt.GotoLabel.String()},
	})
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

	forLoopLabel := AsmLabel{forStatement.LabelStart.String()}
	forLoopEndLabel := AsmLabel{forStatement.LabelEnd.String()}
	incrementLabel := AsmLabel{forStatement.LabelIncrement.String()}
	c.Prog.Emit(forLoopLabel)
	reg, err = c.compileExpr(env, forStatement.Condition, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(Cbz{
		A:     reg,
		Label: forLoopEndLabel,
	})
	reg, err = c.compileStmt(env, forStatement.Block, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(incrementLabel)
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
	whileLoopLabel := AsmLabel{whileStatement.LabelStart.String()}
	whileLoopEndLabel := AsmLabel{whileStatement.LabelEnd.String()}
	c.Prog.Emit(whileLoopLabel)
	reg, err := c.compileExpr(env, whileStatement.Condition, reg)
	if err != nil {
		return reg, err
	}
	c.Prog.Emit(Cbz{
		A:     reg,
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
	endIfLabel := AsmLabel{parser.NewLabel(parser.EndIfType).String()}
	elseLabel := AsmLabel{parser.NewLabel(parser.ElseType).String()}
	reg, err := c.compileExpr(env, ifStatement.Condition, reg)
	c.Prog.Emit(Cbz{
		A:     reg,
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
		A:      reg,
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
		A:      reg,
		Offset: Imm(offsets.Get(varAssignment.Identifier)),
	})
	return reg, nil
}

func (c *CodeGenerator) compileExpr(offsets *Offsets, expr parser.Expression, reg Register) (Register, error) {
	switch e := expr.(type) {
	case *parser.FunctionCall:
		reg := Register(0)
		for _, arg := range e.Arguments {
			c.compileExpr(offsets, arg, reg)
			reg += 1
		}
		c.Prog.Emit(CallSubroutine{
			NewSubroutineDecl(e.Name),
		})
		return Register(0), nil // TODO: func1() && func2() - result is overwritten in registers? 
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
			A:      reg,
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
					A:   left,
					B:   right,
				},
			})
		}
		return reg, nil
	case *parser.BinaryExpression:
		left, err := c.compileExpr(offsets, e.Left, reg)
		if err != nil {
			return 0, err
		}
		right, err := c.compileExpr(offsets, e.Right, reg+1)
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
					A:   left,
					B:   right,
				},
			})
			c.Prog.Emit(MSub{
				Dst: left,
				A:   right + 1,
				B:   right,
				C:   left,
			})
		case token.Vbar, token.VbarVbar:
			c.Prog.Emit(BitwiseOr{
				BinaryOperation{
					Dst: left,
					A:   left,
					B:   right,
				},
			})
		case token.Ampersand, token.AmpersandAmpersand:
			c.Prog.Emit(BitwiseAnd{
				BinaryOperation{
					Dst: left,
					A:   left,
					B:   right,
				},
			})
		case token.Caret:
			c.Prog.Emit(BitwiseXor{
				BinaryOperation{
					Dst: left,
					A:   left,
					B:   right,
				},
			})
		case token.EqualEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     left,
				Value: CSET_EQ,
			})
		case token.BangEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     left,
				Value: CSET_NE,
			})
		case token.GreaterEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     left,
				Value: CSET_GE,
			})
		case token.Greater:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     left,
				Value: CSET_GT,
			})
		case token.LessEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     left,
				Value: CSET_LE,
			})
		case token.Less:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     left,
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
