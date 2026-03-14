package codegen

import (
	token "github.com/dvalkoff/komarulang/tokenizer"
	"github.com/dvalkoff/komarulang/parser"
)

type Offsets struct {
	StackSize int
	OffsetMap map[string]int
	Parent *Offsets
}

func NewOffsets(parent *Offsets) *Offsets {
	return &Offsets{
		StackSize: 0,
		OffsetMap: map[string]int{},
		Parent: parent,
	}
}

func (o *Offsets) Put(varDecl *parser.VarDeclaration) {
	o.OffsetMap[varDecl.Identifier] = o.StackSize
	o.StackSize += sizeOf(varDecl.VarType)
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

func sizeOf(varType token.VarType) int {
	switch varType {
	case token.IntType:
		return 8
	case token.BoolType:
		return 8
	}
	return 0
}
