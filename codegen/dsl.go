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

type Add struct {
	Dst Register
	A   Register
	B   Register
}

func (a Add) String() string {
	return fmt.Sprintf("    add %v, %v, %v", a.Dst, a.A, a.B)
}

type Sub struct {
	Dst Register
	A   Register
	B   Register
}

func (s Sub) String() string {
	return fmt.Sprintf("    sub %v, %v, %v", s.Dst, s.A, s.B)
}

type Mul struct {
	Dst Register
	A   Register
	B   Register
}

func (m Mul) String() string {
	return fmt.Sprintf("    mul %v, %v, %v", m.Dst, m.A, m.B)
}

type Sdiv struct {
	Dst Register
	A   Register
	B   Register
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
