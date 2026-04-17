package codegen

import (
	"fmt"

	"github.com/dvalkoff/komarulang/strcase"
)

type Instruction interface {
	String() string
}

type Register int

type SpilledValue struct {
	NewOffsets *Offsets
	ValueOffset int
}

func (r Register) String() string {
	return fmt.Sprintf("x%d", int(r))
}

type Operand interface {
	String() string
}

const (
	TrueImm  = Imm(1)
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

type BitwiseLeftShift struct {
	BinaryOperation
}

func (s BitwiseLeftShift) String() string {
	return fmt.Sprintf("    lsl %v, %v, %v", s.Dst, s.A, s.B)
}

type BitwiseRightShift struct {
	BinaryOperation
}

func (s BitwiseRightShift) String() string {
	return fmt.Sprintf("    lsr %v, %v, %v", s.Dst, s.A, s.B)
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
	A     Register
	Value CSetValue
}

func (s CSet) String() string {
	return fmt.Sprintf("    cset %v, %v", s.A, s.Value)
}

type StackAllocator struct {
	Value Imm
}

func (a StackAllocator) String() string {
	if a.Value > 0 {
		return fmt.Sprintf("    sub sp, sp, %v", a.Value)
	}
	return ""
}

type StackDeallocator struct {
	Value Imm
}

func (a StackDeallocator) String() string {
	if a.Value > 0 {
		return fmt.Sprintf("    add sp, sp, %v", a.Value)
	}
	return ""
}

type Str struct {
	A      Register
	Offset Imm
}

func (s Str) String() string {
	return fmt.Sprintf("    str %v, [sp, %v]", s.A, s.Offset)
}

type Ldr struct {
	A      Register
	Offset Imm
}

func (s Ldr) String() string {
	return fmt.Sprintf("    ldr %v, [sp, %v]", s.A, s.Offset)
}

type LdrDirect struct {
	A      Register
	Address Register
}

func (s LdrDirect) String() string {
	return fmt.Sprintf("    ldr %v, [%v]", s.A, s.Address)
}

type StrDirect struct {
	A      Register
	Address Register
}

func (s StrDirect) String() string {
	return fmt.Sprintf("    str %v, [%v]", s.A, s.Address)
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
	A     Register
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

type CallPrintSubroutine struct{}

func (psb CallPrintSubroutine) String() string {
	return "    BL _print_int"
}

type CallSubroutine struct{
	Identifer SubroutineDecl
}

func (cs CallSubroutine) String() string {
	return fmt.Sprintf("    BL %v", cs.Identifer.Value())
}

type SubroutineDecl struct {
	Name string
}

func NewSubroutineDecl(name string) SubroutineDecl {
	return SubroutineDecl{"_" + strcase.ToSnake(name)}
}

func (sd SubroutineDecl) String() string {
	return fmt.Sprintf("%v:", sd.Value())
}

func (sd SubroutineDecl) Value() string {
	return sd.Name
}

type AsmReturn struct{}

func (ar AsmReturn) String() string {
	return "    ret"
}

type Global struct {
	Identifier SubroutineDecl
}

func (g Global) String() string {
	return fmt.Sprintf(".global %v", g.Identifier.Value())
}

type Align struct {
	Value int
}

func (a Align) String() string {
	return fmt.Sprintf(".align %v", a.Value)
}

type Svc struct {
	Value string
}

func (s Svc) String() string {
	return fmt.Sprintf("    svc %v", s.Value)
}

type PrintAsmSubroutine struct{}

func (a PrintAsmSubroutine) String() string {
	return `
// _print_int: prints integer in x0 to stdout
// Buffer layout (48 bytes, sp+0 to sp+47):
//   sp+0:  saved x30
//   sp+1 to sp+46: digit buffer (fills backwards from sp+46)
//   sp+47: newline character
// Clobbers: x0-x6, x10, x16
// Preserves: x30 (saved/restored)
_print_int:

    sub sp, sp, #48 // allocating 32 bytes for int64
    mov x5, #47 // sp pointer to current char
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
    mov x4, #47
    sub x2, x4, x5 // calculating and setting len

    svc #0x80
    ldr x30, [sp, #0]   // restore x30
    add sp, sp, #48 // deallocating stack
    ret
`
}

type Extern struct {
	Identifier SubroutineDecl
}

func (g Extern) String() string {
	return fmt.Sprintf(".extern %v", g.Identifier.Value())
}

type DirectAddress struct {
	Dst Register
	Offset Imm
}

func (da DirectAddress) String() string {
	return fmt.Sprintf("    add %v, sp, %v", da.Dst, da.Offset)
}