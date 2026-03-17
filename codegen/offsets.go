package codegen

import (
	"github.com/dvalkoff/komarulang/parser/types"
	"github.com/dvalkoff/komarulang/parser"
)

const (
	BytesInRegister = 8
)

type Offsets struct {
	StackSize int
	OffsetMap map[string]int
	Parent    *Offsets
}

func NewOffsets(parent *Offsets) *Offsets {
	return &Offsets{
		StackSize: 0,
		OffsetMap: map[string]int{},
		Parent:    parent,
	}
}

func (o *Offsets) Put(varDecl *parser.VarDeclaration) {
	o.OffsetMap[varDecl.Identifier] = o.StackSize
	o.StackSize += sizeOf(varDecl.VarType)
}

func (o *Offsets) PutFunArg(funArg *parser.FunctionArgument) {
	o.OffsetMap[funArg.Identifier] = o.StackSize
	o.StackSize += sizeOf(funArg.VarType)
}

func (o *Offsets) Get(identifier string) int {
	if o == nil {
		panic("No variable")
	}
	if val, ok := o.OffsetMap[identifier]; ok {
		return val
	}
	return o.StackSize + o.Parent.Get(identifier)
}

func (o *Offsets) AlignStackSize() {
	o.StackSize = (o.StackSize + 15) & ^15
}

func sizeOf(varType types.Type) int {
	switch varType {
	case types.IntType:
		return 8
	case types.BoolType:
		return 8
	}
	return 0
}


type RegisterAllocator struct {
	CurrentReg Register
	MaxReg Register
}

func NewRegisterAllocator(maxReg Register) *RegisterAllocator {
	return &RegisterAllocator{CurrentReg: Register(0), MaxReg: maxReg}
}

func (ra *RegisterAllocator) Alloc(varType types.Type) (Register, bool) {
	size := Register(sizeOf(varType) / BytesInRegister)
	if ra.CurrentReg + size > ra.MaxReg {
		return ra.CurrentReg, false
	}
	returnValue := ra.CurrentReg
	ra.CurrentReg += size
	return returnValue, true
}

func (ra *RegisterAllocator) Free(varType types.Type) {
	size := Register(sizeOf(varType) / BytesInRegister)
	if ra.CurrentReg > 0 {
		ra.CurrentReg = max(0, ra.CurrentReg - size)
	}
}