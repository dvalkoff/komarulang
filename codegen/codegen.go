package codegen

import (
	"fmt"
	"strings"

	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/parser/types"
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
				err := c.compileStmt(nil, decl)
				if err != nil {
					return err
				}
			}
		}
	}
	err := c.compileStmt(nil, entrypoint)
	return err
}

func (c *CodeGenerator) compileStmt(env *Offsets, stmt parser.Statement) error {
	switch typed := stmt.(type) {
	case *parser.ExprStatement:
		_, err := c.compileExpr(env, typed.Expr, NewRegisterAllocatorDefault())
		return err
	case *parser.Block:
		return c.compileBlock(env, typed)
	case *parser.VarDeclaration:
		return c.compileVarDeclaration(env, typed)
	case *parser.VarAssignment:
		return c.compileVarAssignment(env, typed)
	case *parser.IfStatement:
		return c.compileIf(env, typed)
	case *parser.WhileStatement:
		return c.compileWhile(env, typed)
	case *parser.ForStatement:
		return c.compileFor(env, typed)
	case *parser.PrintStatement:
		return c.compilePrint(env, typed)
	case *parser.BreakStatement:
		return c.compileBreak(env, typed)
	case *parser.ContinueStatement:
		return c.compileContinue(env, typed)
	case *parser.FunctionDecl:
		return c.compileFunction(env, typed)
	case *parser.ReturnStatement:
		return c.compileReturnStmt(env, typed)
	}
	return fmt.Errorf("Unexpected statement %v", stmt)
}

func (c *CodeGenerator) compileFunction(parent *Offsets, funcDecl *parser.FunctionDecl) error {
	subroutineDecl := NewSubroutineDecl(funcDecl.Name)
	c.Prog.Emit(subroutineDecl)

	offsets := NewOffsets(parent)
	offsets.StackSize += BytesInRegister // x30
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
			return err
		}
		argumentRegister += 1
	}
	for _, stmt := range block.Stmts {
		err := c.compileStmt(offsets, stmt)
		if err != nil {
			return err
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
		return c.compileExit(offsets)
	}
	c.Prog.Emit(AsmReturn{})
	return nil
}

func (c *CodeGenerator) compileFunArgument(env *Offsets, funArg *parser.FunctionArgument, reg Register) error {
	c.Prog.Emit(Str{
		A: reg,
		Offset: Imm(env.Get(funArg.Identifier)),
	})
	return nil
}

func (c *CodeGenerator) compileExit(env *Offsets) error {
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
	return nil
}

func (c *CodeGenerator) compileReturnStmt(env *Offsets, returnStmt *parser.ReturnStatement) error {
	vh, err := c.compileExpr(env, returnStmt.Expression, NewRegisterAllocatorDefault())
	if err != nil {
		return err
	}
	reg := c.LoadIntoRegister(vh, 0)
	if reg != 0 {
		c.Prog.Emit(Mov{
			Register(0),
			reg,
		})
	}

	epilogueLabel := AsmLabel{returnStmt.EpilogueLabel.String()}
	c.Prog.Emit(Bjump{
		Label: epilogueLabel,
	})
	return nil
}

func (c *CodeGenerator) compilePrint(env *Offsets, printStmt *parser.PrintStatement) error {
	vh, err := c.compileExpr(env, printStmt.Expr, NewRegisterAllocatorDefault())
	reg := c.LoadIntoRegister(vh, 0)
	if reg != 0 {
		c.Prog.Emit(Mov{
			Register(0),
			reg,
		})
	}
	if err != nil {
		return err
	}
	c.Prog.Emit(CallPrintSubroutine{})
	return nil
}

func (c *CodeGenerator) compileBreak(env *Offsets, breakStmt *parser.BreakStatement) error {
	c.Prog.Emit(Bjump{
		Label: AsmLabel{breakStmt.GotoLabel.String()},
	})
	return nil
}

func (c *CodeGenerator) compileContinue(env *Offsets, continueStmt *parser.ContinueStatement) error {
	c.Prog.Emit(Bjump{
		Label: AsmLabel{continueStmt.GotoLabel.String()},
	})
	return nil
}

func (c *CodeGenerator) compileFor(parent *Offsets, forStatement *parser.ForStatement) error {
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

	err := c.compileStmt(env, forStatement.VarDecl)
	if err != nil {
		return err
	}

	forLoopLabel := AsmLabel{forStatement.LabelStart.String()}
	forLoopEndLabel := AsmLabel{forStatement.LabelEnd.String()}
	incrementLabel := AsmLabel{forStatement.LabelIncrement.String()}
	c.Prog.Emit(forLoopLabel)
	conditionVh, err := c.compileExpr(env, forStatement.Condition, NewRegisterAllocatorDefault())
	if err != nil {
		return err
	}
	conditionReg := c.LoadIntoRegister(conditionVh, 0)
	c.Prog.Emit(Cbz{
		A:     conditionReg,
		Label: forLoopEndLabel,
	})
	err = c.compileStmt(env, forStatement.Block)
	if err != nil {
		return err
	}
	c.Prog.Emit(incrementLabel)
	err = c.compileStmt(env, forStatement.Increment)
	c.Prog.Emit(Bjump{
		Label: forLoopLabel,
	})
	c.Prog.Emit(forLoopEndLabel)

	if allocationRequired {
		c.Prog.Emit(StackDeallocator{
			Value: Imm(env.StackSize),
		})
	}
	return nil
}

func (c *CodeGenerator) compileWhile(env *Offsets, whileStatement *parser.WhileStatement) error {
	whileLoopLabel := AsmLabel{whileStatement.LabelStart.String()}
	whileLoopEndLabel := AsmLabel{whileStatement.LabelEnd.String()}
	c.Prog.Emit(whileLoopLabel)
	conditionVh, err := c.compileExpr(env, whileStatement.Condition, NewRegisterAllocatorDefault())
	if err != nil {
		return err
	}
	conditionReg := c.LoadIntoRegister(conditionVh, 0)
	c.Prog.Emit(Cbz{
		A:     conditionReg,
		Label: whileLoopEndLabel,
	})
	err = c.compileStmt(env, whileStatement.Block)
	if err != nil {
		return err
	}
	c.Prog.Emit(Bjump{
		Label: whileLoopLabel,
	})
	c.Prog.Emit(whileLoopEndLabel)
	return nil
}

func (c *CodeGenerator) compileIf(env *Offsets, ifStatement *parser.IfStatement) error {
	endIfLabel := AsmLabel{parser.NewLabel(parser.EndIfType).String()}
	elseLabel := AsmLabel{parser.NewLabel(parser.ElseType).String()}
	conditionVh, err := c.compileExpr(env, ifStatement.Condition, NewRegisterAllocatorDefault())
	if err != nil {
		return err
	}
	conditionReg := c.LoadIntoRegister(conditionVh, 0)
	c.Prog.Emit(Cbz{
		A:     conditionReg,
		Label: elseLabel,
	})
	err = c.compileStmt(env, ifStatement.Block)
	if err != nil {
		return err
	}
	c.Prog.Emit(Bjump{
		Label: endIfLabel,
	})
	c.Prog.Emit(elseLabel)
	if ifStatement.ElseBlock != nil {
		if err := c.compileStmt(env, ifStatement.ElseBlock); err != nil {
			return err
		}
	}
	c.Prog.Emit(endIfLabel)

	return nil
}

func (c *CodeGenerator) compileBlock(parent *Offsets, block *parser.Block) error {
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
		err := c.compileStmt(offsets, stmt)
		if err != nil {
			return err
		}
	}

	if offsets.StackSize > 0 {
		c.Prog.Emit(StackDeallocator{
			Value: Imm(offsets.StackSize),
		})
	}

	return nil
}

func (c *CodeGenerator) compileVarDeclaration(offsets *Offsets, varDecl *parser.VarDeclaration) error {
	vh, err := c.compileExpr(offsets, varDecl.Expr, NewRegisterAllocatorDefault())
	if err != nil {
		return err
	}
	reg := c.LoadIntoRegister(vh, 0)
	c.Prog.Emit(Str{
		A:      reg,
		Offset: Imm(offsets.Get(varDecl.Identifier)),
	})
	return nil
}

func (c *CodeGenerator) compileVarAssignment(offsets *Offsets, varAssignment *parser.VarAssignment) error {
	vh, err := c.compileExpr(offsets, varAssignment.Expr, NewRegisterAllocatorDefault())
	if err != nil {
		return err
	}
	reg := c.LoadIntoRegister(vh, 0)
	c.Prog.Emit(Str{
		A:      reg,
		Offset: Imm(offsets.Get(varAssignment.Identifier)),
	})
	return nil
}

type DebugInstruction struct {
	Value string
}

func (di DebugInstruction) String() string {
	return di.Value
}

func (c *CodeGenerator) compileExpr(offsets *Offsets, expr parser.Expression, regAllocator *RegisterAllocator) (*ValueHolder, error) {
	if int(regAllocator.MaxReg + 1) * BytesInRegister < sizeOf(expr.Type()) {
		return nil, fmt.Errorf("Can not compile expression. It won't fit into registers")
	}
	switch e := expr.(type) {
	case *parser.FunctionCall:
		return c.compileWithSpilling(offsets, expr, regAllocator)
	case *parser.IntegerLiteral:
		dst, succeeded := regAllocator.Alloc(expr.Type())
		if !succeeded {
			return c.compileWithSpilling(offsets, expr, regAllocator)
		}
		instructions := c.loadInt(dst, e.Value)
		c.Prog.Emit(instructions...)
		return &ValueHolder{Reg: dst}, nil
	case *parser.BooleanLiteral:
		dst, succeeded := regAllocator.Alloc(expr.Type())
		if !succeeded {
			return c.compileWithSpilling(offsets, expr, regAllocator)
		}
		if e.Value {
			c.Prog.Emit(Mov{dst, TrueImm})
		} else {
			c.Prog.Emit(Mov{dst, FalseImm})
		}
		return &ValueHolder{Reg: dst}, nil
	case *parser.VoidLiteral:
		return &ValueHolder{Reg: 0}, nil
	case *parser.IdentifierLiteral:
		dst, succeeded := regAllocator.Alloc(expr.Type())
		if !succeeded {
			return c.compileWithSpilling(offsets, expr, regAllocator)
		}
		c.Prog.Emit(Ldr{
			A:      dst,
			Offset: Imm(offsets.Get(e.Value)),
		})
		return &ValueHolder{Reg: dst}, nil
	case *parser.UnaryExpression:
		dst, succeeded := regAllocator.Alloc(expr.Type())
		if !succeeded {
			return c.compileWithSpilling(offsets, expr, regAllocator)
		}
		left := dst
		rightVh, err := c.compileExpr(offsets, e.Right, regAllocator)
		if err != nil {
			return nil, err
		}
		right := c.AllocateAndLoadIntoRegister(rightVh, regAllocator)
		regAllocator.Free(e.Right.Type())

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
					Dst: dst,
					A:   left,
					B:   right,
				},
			})
		}
		return &ValueHolder{Reg: dst}, nil
	case *parser.BinaryExpression:
		leftVh, err := c.compileExpr(offsets, e.Left, regAllocator)
		if err != nil {
			return nil, err
		}
		left := c.AllocateAndLoadIntoRegister(leftVh, regAllocator)
		rightVh, err := c.compileExpr(offsets, e.Right, regAllocator)
		if err != nil {
			return nil, err
		}
		right := c.AllocateAndLoadIntoRegister(rightVh, regAllocator)
		regAllocator.Free(e.Right.Type())
		
		dst := left
		switch e.Operator {
		case token.Plus:
			c.Prog.Emit(Add{
				BinaryOperation{
					Dst: dst,
					A:   left,
					B:   right,
				},
			})
		case token.Minus:
			c.Prog.Emit(Sub{
				BinaryOperation{
					Dst: dst,
					A:   left,
					B:   right,
				},
			})
		case token.Star:
			c.Prog.Emit(Mul{
				BinaryOperation{
					Dst: dst,
					A:   left,
					B:   right,
				},
			})
		case token.Slash:
			c.Prog.Emit(Sdiv{
				BinaryOperation{
					Dst: dst,
					A:   left,
					B:   right,
				},
			})
		case token.Percent: // TODO: operation uses right+1 - potential registers overflow
			c.Prog.Emit(Sdiv{
				BinaryOperation{
					Dst: right + 1, 
					A:   left,
					B:   right,
				},
			})
			c.Prog.Emit(MSub{
				Dst: dst,
				A:   right + 1,
				B:   right,
				C:   left,
			})
		case token.Vbar, token.VbarVbar:
			c.Prog.Emit(BitwiseOr{
				BinaryOperation{
					Dst: dst,
					A:   left,
					B:   right,
				},
			})
		case token.Ampersand, token.AmpersandAmpersand:
			c.Prog.Emit(BitwiseAnd{
				BinaryOperation{
					Dst: dst,
					A:   left,
					B:   right,
				},
			})
		case token.Caret:
			c.Prog.Emit(BitwiseXor{
				BinaryOperation{
					Dst: dst,
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
				A:     dst,
				Value: CSET_EQ,
			})
		case token.BangEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     dst,
				Value: CSET_NE,
			})
		case token.GreaterEqual:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     dst,
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
				A:     dst,
				Value: CSET_LE,
			})
		case token.Less:
			c.Prog.Emit(Cmd{
				A: left,
				B: right,
			})
			c.Prog.Emit(CSet{
				A:     dst,
				Value: CSET_LT,
			})
		}
		return &ValueHolder{Reg: dst}, nil
	}

	return nil, fmt.Errorf("Unexpected expression %v, %t", expr, expr)
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

func (c *CodeGenerator) compileFunctionCall(offsets *Offsets, expr *parser.FunctionCall, regAllocator *RegisterAllocator) (*ValueHolder, error) {
	if regAllocator.CurrentReg != 0 {
		return nil, fmt.Errorf("Function call cannot be compiled because of registers overwrite")
	}
	for _, arg := range expr.Arguments {
		_, err := c.compileExpr(offsets, arg, regAllocator)
		if err != nil {
			return nil, err
		}
	}
	c.Prog.Emit(CallSubroutine{
		NewSubroutineDecl(expr.Name),
	})
	reg, ok := regAllocator.Alloc(expr.Type())
	if !ok {
		return nil, fmt.Errorf("Cannot allocate registers for function return value")
	}
	return &ValueHolder{Reg: reg}, nil
}

func (c *CodeGenerator) compileWithSpilling(parent *Offsets, expr parser.Expression, regAllocator *RegisterAllocator) (*ValueHolder, error) {
	returnOffsets := NewOffsets(parent)
	returnOffsets.StackSize += sizeOf(expr.Type()) // reserving a space for expression spilled value
	returnOffsets.AlignStackSize()
	if returnOffsets.StackSize > 0 {
		c.Prog.Emit(StackAllocator{
			Value: Imm(returnOffsets.StackSize),
		})
	}

	tempOffsets := NewOffsets(returnOffsets)
	tempOffsets.StackSize += int(regAllocator.CurrentReg) * BytesInRegister // adding a space for temporarily spilled values
	tempOffsets.AlignStackSize()
	if tempOffsets.StackSize > 0 {
		c.Prog.Emit(StackAllocator{
			Value: Imm(tempOffsets.StackSize),
		})
	}

	for reg := range regAllocator.CurrentReg { // temporarilty loading registers into stack
		c.Prog.Emit(Str{
			A: reg,
			Offset: Imm(int(reg * BytesInRegister)),
		})
	}
	
	var returnValueHolder *ValueHolder
	var err error
	switch typedExpr := expr.(type) {
	case *parser.FunctionCall:
		returnValueHolder, err = c.compileFunctionCall(tempOffsets, typedExpr, NewRegisterAllocator(regAllocator.MaxReg))
	default:
		returnValueHolder, err = c.compileExpr(tempOffsets, typedExpr, NewRegisterAllocator(regAllocator.MaxReg))
	}
	if err != nil {
		return nil, err
	}
	returnValueRegister := c.LoadIntoRegister(returnValueHolder, 0)
	c.Prog.Emit(Str{
		A: Register(returnValueRegister),
		Offset: Imm(tempOffsets.StackSize),
	})
	for reg := range regAllocator.CurrentReg { // loading temporarliy spilled values back to registers
		c.Prog.Emit(Ldr{
			A: reg,
			Offset: Imm(int(reg * BytesInRegister)),
		})
	}
	if tempOffsets.StackSize > 0 {
		c.Prog.Emit(StackDeallocator{
			Value: Imm(tempOffsets.StackSize), // deallocating stack, leaving only expression result
		})
	}

	valueHolder := &ValueHolder{
		Spilled: &SpilledValue{
			NewOffsets: returnOffsets,
			ValueOffset: 0,
		},
	}
	return valueHolder, nil
}

func (c *CodeGenerator) LoadIntoRegister(valueHolder *ValueHolder, dst Register) Register {
	var reg Register
	if valueHolder.IsRegister() {
		reg = valueHolder.Reg
	} else {
		reg = dst
		c.Prog.Emit(Ldr{
			A: dst,
			Offset: Imm(valueHolder.Spilled.ValueOffset),
		})
		c.Prog.Emit(StackDeallocator{
			Value: Imm(valueHolder.Spilled.NewOffsets.StackSize),
		})
	}
	return reg
}

func (c *CodeGenerator) AllocateAndLoadIntoRegister(valueHolder *ValueHolder, registerAllocator *RegisterAllocator) Register {
	var reg Register
	if valueHolder.IsRegister() {
		reg = valueHolder.Reg
	} else {
		var ok bool
		reg, ok = registerAllocator.Alloc(types.IntType) // for now only 8 bytes size
		if !ok {
			panic("failed to allocate register")
		}
		c.Prog.Emit(Ldr{
			A: reg,
			Offset: Imm(valueHolder.Spilled.ValueOffset),
		})
		c.Prog.Emit(StackDeallocator{
			Value: Imm(valueHolder.Spilled.NewOffsets.StackSize),
		})
	}
	return reg
}
