package codegen

import "fmt"

type Instruction interface {
	String() string
}

type Register int

func (r Register) String() string {
	return fmt.Sprintf("x%d", int(r))
}

type Operand interface {
	String() string
}

const (
	TrueImm = Imm(1)
	FalseImm = Imm(0)
)

type Imm int

func (i Imm) String() string {
	return fmt.Sprintf("#%d", int(i))
}

type Mov struct {
	Dst Register
	Src Operand
}

func (m Mov) String() string {
	return fmt.Sprintf("    mov %v, %v", m.Dst, m.Src)
}

type Movz struct {
	Dst Register
	Src Operand
	Lsl Operand
}

func (m Movz) String() string {
	return fmt.Sprintf("    movz %v, %v, lsl %v", m.Dst, m.Src, m.Lsl)
}

type Movk struct {
	Dst Register
	Src Operand
	Lsl Operand
}

func (m Movk) String() string {
	return fmt.Sprintf("    movk %v, %v, lsl %v", m.Dst, m.Src, m.Lsl)
}

type BinaryOperation struct {
	Dst Register
	A   Register
	B   Register	
}

type Add struct {
	BinaryOperation
}

func (a Add) String() string {
	return fmt.Sprintf("    add %v, %v, %v", a.Dst, a.A, a.B)
}

type Sub struct {
	BinaryOperation
}

func (s Sub) String() string {
	return fmt.Sprintf("    sub %v, %v, %v", s.Dst, s.A, s.B)
}

type Mul struct {
	BinaryOperation
}

func (m Mul) String() string {
	return fmt.Sprintf("    mul %v, %v, %v", m.Dst, m.A, m.B)
}

type Sdiv struct {
	BinaryOperation
}

func (s Sdiv) String() string {
	return fmt.Sprintf("    sdiv %v, %v, %v", s.Dst, s.A, s.B)
}

type Neg struct {
	Dst Register
	A   Register
}

func (n Neg) String() string {
	return fmt.Sprintf("    neg %v, %v", n.Dst, n.A)
}

type Udiv struct {
	BinaryOperation
}

func (s Udiv) String() string {
	return fmt.Sprintf("    udiv %v, %v, %v", s.Dst, s.A, s.B)
}

type MSub struct {
	Dst Register
	A   Register
	B   Register
	C   Register
}

func (s MSub) String() string {
	return fmt.Sprintf("    msub %v, %v, %v, %v", s.Dst, s.A, s.B, s.C)
}

type BitwiseAnd struct {
	BinaryOperation
}

func (s BitwiseAnd) String() string {
	return fmt.Sprintf("    and %v, %v, %v", s.Dst, s.A, s.B)
}

type BitwiseOr struct {
	BinaryOperation
}

func (s BitwiseOr) String() string {
	return fmt.Sprintf("    orr %v, %v, %v", s.Dst, s.A, s.B)
}

type BitwiseXor struct {
	BinaryOperation
}

func (s BitwiseXor) String() string {
	return fmt.Sprintf("    eor %v, %v, %v", s.Dst, s.A, s.B)
}

type Cmd struct {
	A Register
	B Register
}

func (s Cmd) String() string {
	return fmt.Sprintf("    cmp %v, %v", s.A, s.B)
}

const (
	CSET_EQ CSetValue = "EQ"
	CSET_NE CSetValue = "NE"

	CSET_GT CSetValue = "GT"
	CSET_GE CSetValue = "GE"
	
	CSET_LT CSetValue = "LT"
	CSET_LE CSetValue = "LE"
)

type CSetValue string

type CSet struct {
	A Register
	Value CSetValue
}

func (s CSet) String() string {
	return fmt.Sprintf("    cset %v, %v", s.A, s.Value)
}

type StackAllocator struct {
	Value Imm
}

func (a StackAllocator) String() string {
	return fmt.Sprintf("    sub sp, sp, %v", a.Value)
}

type StackDeallocator struct {
	Value Imm
}

func (a StackDeallocator) String() string {
	return fmt.Sprintf("    add sp, sp, %v", a.Value)
}

type Str struct {
	A Register
	Offset Imm
}

func (s Str) String() string {
	return fmt.Sprintf("    str %v, [sp, %v]", s.A, s.Offset)
}

type Ldr struct {
	A Register
	Offset Imm
}

func (s Ldr) String() string {
	return fmt.Sprintf("    ldr %v, [sp, %v]", s.A, s.Offset)
}

type AsmLabel struct {
	Name string
}

func (s AsmLabel) Value() string {
	return s.Name
}

func (s AsmLabel) String() string {
	return fmt.Sprintf("    %v:", s.Value())
}

type Cbz struct {
	A Register
	Label AsmLabel
}

func (s Cbz) String() string {
	return fmt.Sprintf("    cbz %v, %v", s.A, s.Label.Value())
}

type Bjump struct {
	Label AsmLabel
}

func (b Bjump) String() string {
	return fmt.Sprintf("    B %v", b.Label.Value())
}

type PrintSubroutine struct {}

func (psb PrintSubroutine) String() string {
	return "    BL _print_int"
}