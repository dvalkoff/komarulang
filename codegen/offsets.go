package codegen

import (
	"github.com/dvalkoff/komarulang/parser/types"
	"github.com/dvalkoff/komarulang/parser"
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
