package parser

import (
	"fmt"

	"github.com/dvalkoff/komarulang/parser/types"
)

type ExpressionWrapper struct {
	Expr Expression
	TempVars []*VarDeclaration
}

func (ew *ExpressionWrapper) Expression() {
	
}

func (ew *ExpressionWrapper) Type() types.Type {
	return ew.Expr.Type()
}

type Flattener struct {
	counter int
	tmpVarName string
}

func NewFlattener() *Flattener {
	return &Flattener{
		counter: 0,
		tmpVarName: "tmpvar",
	}
}

func (f *Flattener) GetNextName() string {
	name := fmt.Sprintf("%v_%v", f.tmpVarName, f.counter)
	f.counter++
	return name
}

type FunctionCallsContext struct {
	OrderedFunctionCalls []*FunctionCall
	MapParent map[*FunctionCall]Expression
}

func (f *Flattener) Flatten(e Expression) Expression {
	funCalls := f.getFunCalls(e)
	if len(funCalls) == 0 {
		return e
	}
	funCallsMap := map[*FunctionCall]string{}
	varDeclarations := []*VarDeclaration{}
	for _, funCall := range funCalls {
		tempVarName := f.GetNextName()
		decl := &VarDeclaration{
			VarType: funCall.Type(),
			Identifier: tempVarName,
			Expr: funCall,
		}
		varDeclarations = append(varDeclarations, decl)
		funCallsMap[funCall] = tempVarName
	}

	e = f.replaceFunCalls(funCallsMap, e)
	return &ExpressionWrapper{
		Expr: e,
		TempVars: varDeclarations,
	}
}

func (f *Flattener) getFunCalls(e Expression) []*FunctionCall {
	switch typed := e.(type) {
	case *BinaryExpression:
		funCalls := make([]*FunctionCall, 0)
		funCalls = append(funCalls, f.getFunCalls(typed.Left)...)
		funCalls = append(funCalls, f.getFunCalls(typed.Right)...)
		return funCalls
	case *UnaryExpression:
		return f.getFunCalls(typed.Right)
	case *FunctionCall:
		funCalls := make([]*FunctionCall, 0)
		for _, arg := range typed.Arguments {
			argCalls := f.getFunCalls(arg)
			funCalls = append(funCalls, argCalls...)
		}
		funCalls = append(funCalls, typed)
		return funCalls
	default:
		return []*FunctionCall{}
	}
}

func (f *Flattener) replaceFunCalls(funCalls map[*FunctionCall]string, e Expression) Expression {
	switch typed := e.(type) {
	case *BinaryExpression:
		typed.Left = f.replaceFunCalls(funCalls, typed.Left)
		typed.Right = f.replaceFunCalls(funCalls, typed.Right)
		return typed
	case *UnaryExpression:
		typed.Right = f.replaceFunCalls(funCalls, typed.Right)
		return typed
	case *FunctionCall:
		for i := 0; i < len(typed.Arguments); i++ {
			typed.Arguments[i] = f.replaceFunCalls(funCalls, typed.Arguments[i])
		}
		return &IdentifierLiteral{
			VarType: typed.Type(),
			Value: funCalls[typed],
		}
	default:
		return typed
	}
}