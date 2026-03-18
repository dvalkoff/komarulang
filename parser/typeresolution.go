package parser

import (
	"fmt"

	"github.com/dvalkoff/komarulang/env"
	"github.com/dvalkoff/komarulang/parser/types"
	token "github.com/dvalkoff/komarulang/tokenizer"
)

type TokenEnv = *env.Environment[types.Type]

type LoopStmt interface {
	Statement
}

type SemanticAnalysisContext struct {
	VarEnv           *env.Environment[types.Type]
	FunEnv           *env.Environment[*FunctionDecl]
	LabelEnv         *env.Environment[LoopStmt]
	CurrentFunction  *FunctionDecl
}

func FromSemanticAnalysisContext(parent *SemanticAnalysisContext) *SemanticAnalysisContext {
	if parent == nil {
		return newSemanticAnalysisContext(nil, nil, nil, nil)
	}
	return newSemanticAnalysisContext(parent.VarEnv, parent.FunEnv, parent.LabelEnv, parent.CurrentFunction)
}

func newSemanticAnalysisContext(varEnv *env.Environment[types.Type], funEnv *env.Environment[*FunctionDecl], lavelEnv *env.Environment[LoopStmt], curFunc *FunctionDecl) *SemanticAnalysisContext {
	return &SemanticAnalysisContext{
		VarEnv:   env.NewEnvironment(varEnv),
		FunEnv:   env.NewEnvironment(funEnv),
		LabelEnv: env.NewEnvironment(lavelEnv),
		CurrentFunction: curFunc,
	}
}

type TypeResolver struct{
	flattener *Flattener
}

func NewTypeResolver() *TypeResolver {
	return &TypeResolver{
		flattener: NewFlattener(),
	}
}

func (t *TypeResolver) Resolve(stmts []Statement) error {
	semCtx := FromSemanticAnalysisContext(nil)
	semCtx.FunEnv.New("malloc",&FunctionDecl{
		Name: "malloc",
		Arguments: []*FunctionArgument{{VarType: types.IntType}},
		ReturnType: types.IntPointer,
	})
	semCtx.FunEnv.New("free",&FunctionDecl{
		Name: "free",
		Arguments: []*FunctionArgument{{VarType: types.IntPointer}},
		ReturnType: types.VoidType,
	})
	for _, stmt := range stmts {
		if funcDecl, ok := stmt.(*FunctionDecl); ok {
			if semCtx.FunEnv.Exists(funcDecl.Name) {
				return fmt.Errorf("Function %v already exists", funcDecl.Name)
			}
			semCtx.FunEnv.New(funcDecl.Name, funcDecl)
		}
	}
	for _, stmt := range stmts {
		err := t.resolveStmt(semCtx, stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TypeResolver) resolveStmt(semCtx *SemanticAnalysisContext, stmt Statement) error {
	switch typed := stmt.(type) {
	case *Block:
		return t.resolveBlock(semCtx, typed)
	case *VarDeclaration:
		return t.resolveVarDeclaration(semCtx, typed)
	case *VarAssignment:
		return t.resolveVarAssignment(semCtx, typed)
	case *ExprStatement:
		_, err := t.evaluateType(semCtx, typed.Expr)
		typed.Expr = t.flattener.Flatten(typed.Expr)
		return err
	case *PrintStatement:
		_, err := t.evaluateType(semCtx, typed.Expr)
		typed.Expr = t.flattener.Flatten(typed.Expr)
		return err
	case *IfStatement:
		return t.resolveIfStmt(semCtx, typed)
	case *WhileStatement:
		return t.resolveWhileStmt(semCtx, typed)
	case *ForStatement:
		return t.resolveForStmt(semCtx, typed)
	case *BreakStatement:
		return t.resolveBreakStmt(semCtx, typed)
	case *ContinueStatement:
		return t.resolveContinueStmt(semCtx, typed)
	case *FunctionDecl:
		return t.resolveFunctionDecl(semCtx, typed)
	case *ReturnStatement:
		return t.resolveReturnStatement(semCtx, typed)
	}
	return fmt.Errorf("Unexpected stmt %v", stmt)
}

func (t *TypeResolver) resolveFunctionDecl(semCtx *SemanticAnalysisContext, funcDecl *FunctionDecl) error {
	if val, ok := semCtx.FunEnv.Get(funcDecl.Name); ok && val != funcDecl {
		return fmt.Errorf("Function %v already exists", funcDecl.Name)
	}
	funcDecl.EpilogueLabel = NewLabel(FunctionEpilogue)
	semCtx.FunEnv.New(funcDecl.Name, funcDecl)
	childCtx := newSemanticAnalysisContext(semCtx.VarEnv, semCtx.FunEnv, nil, funcDecl)
	for _, arg := range funcDecl.Arguments {
		if childCtx.VarEnv.Exists(arg.Identifier) {
			return fmt.Errorf("Variable %v already exists", funcDecl.Name)
		}
		childCtx.VarEnv.New(arg.Identifier, arg.VarType)
	}

	err := t.resolveStmt(childCtx, funcDecl.Body)
	if err != nil {
		return err
	}
	if funcDecl.ReturnStmtsCount == 0 && funcDecl.ReturnType != types.VoidType {
		return fmt.Errorf("Function %v has no return statements", funcDecl.Name)
	}
	return nil
}

func (t *TypeResolver) resolveBlock(parent *SemanticAnalysisContext, block *Block) error {
	semCtx := FromSemanticAnalysisContext(parent)
	for _, stmt := range block.Stmts {
		err := t.resolveStmt(semCtx, stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TypeResolver) resolveReturnStatement(semCtx *SemanticAnalysisContext, returnStmt *ReturnStatement) error {
	if semCtx.CurrentFunction == nil {
		return fmt.Errorf("Declaring return statement outside of a function body is not allowed")
	}
	currentFun := semCtx.CurrentFunction
	exprType, err := t.evaluateType(semCtx, returnStmt.Expression)
	returnStmt.Expression = t.flattener.Flatten(returnStmt.Expression)
	if err != nil {
		return err
	}
	returnStmt.ReturnType = exprType
	if !t.compatible(exprType, currentFun.ReturnType) {
		return TypeError{Expected: currentFun.ReturnType, Got: returnStmt.ReturnType}
	}
	returnStmt.EpilogueLabel = currentFun.EpilogueLabel
	currentFun.ReturnStmtsCount++
	return nil
}

func (t *TypeResolver) resolveVarDeclaration(semCtx *SemanticAnalysisContext, varDecl *VarDeclaration) error {
	specifiedType := varDecl.VarType
	calculatedType, err := t.evaluateType(semCtx, varDecl.Expr)
	varDecl.Expr = t.flattener.Flatten(varDecl.Expr)
	if err != nil {
		return err
	}
	if semCtx.VarEnv.Exists(varDecl.Identifier) {
		return fmt.Errorf("Variable %v already exists", varDecl.Identifier)
	}
	if specifiedType != types.NotSpecified && !t.compatible(specifiedType, calculatedType) {
		return TypeError{Expected: specifiedType, Got: calculatedType}
	}
	if specifiedType == types.NotSpecified {
		varDecl.VarType = calculatedType
	}
	semCtx.VarEnv.New(varDecl.Identifier, varDecl.VarType)
	return nil
}

func (t *TypeResolver) resolveVarAssignment(semCtx *SemanticAnalysisContext, assignment *VarAssignment) error {
	leftExprType, err := t.evaluateType(semCtx, assignment.LeftExpr)
	assignment.LeftExpr = t.flattener.Flatten(assignment.LeftExpr)
	if err != nil {
		return err
	}
	calculatedType, err := t.evaluateType(semCtx, assignment.Expr)
	assignment.Expr = t.flattener.Flatten(assignment.Expr)
	if err != nil {
		return err
	}
	if !t.compatible(leftExprType, calculatedType) {
		return TypeError{Expected: leftExprType, Got: calculatedType}
	}
	return nil
}

func (t *TypeResolver) resolveIfStmt(semCtx *SemanticAnalysisContext, stmt *IfStatement) error {
	if err := t.resolveCondition(semCtx, stmt.Condition); err != nil {
		return err
	}
	stmt.Condition = t.flattener.Flatten(stmt.Condition)
	if err := t.resolveStmt(semCtx, stmt.Block); err != nil {
		return err
	}
	if stmt.ElseBlock != nil {
		return t.resolveStmt(semCtx, stmt.ElseBlock)
	}
	return nil
}

func (t *TypeResolver) resolveBreakStmt(semCtx *SemanticAnalysisContext, stmt *BreakStatement) error {
	if loop, ok := semCtx.LabelEnv.Get(string(LoopLabel)); ok {
		switch typedLoop := loop.(type) {
		case *WhileStatement:
			stmt.GotoLabel = typedLoop.LabelEnd
		case *ForStatement:
			stmt.GotoLabel = typedLoop.LabelEnd
		default:
			return fmt.Errorf("Unknown loop type %v, %t", typedLoop, typedLoop)
		}
		return nil
	} else {
		return fmt.Errorf("Break is not in a loop")
	}
}

func (t *TypeResolver) resolveContinueStmt(semCtx *SemanticAnalysisContext, stmt *ContinueStatement) error {
	if loop, ok := semCtx.LabelEnv.Get(string(LoopLabel)); ok {
		switch typedLoop := loop.(type) {
		case *WhileStatement:
			stmt.GotoLabel = typedLoop.LabelStart
		case *ForStatement:
			stmt.GotoLabel = typedLoop.LabelIncrement
		default:
			return fmt.Errorf("Unknown loop type %v, %t", typedLoop, typedLoop)
		}
		return nil
	} else {
		return fmt.Errorf("Continue is not in a loop")
	}
}

func (t *TypeResolver) resolveWhileStmt(parent *SemanticAnalysisContext, stmt *WhileStatement) error {
	semCtx := FromSemanticAnalysisContext(parent)
	stmt.LabelStart = NewLabel(LoopStart)
	stmt.LabelEnd = NewLabel(LoopEnd)
	semCtx.LabelEnv.New(string(LoopLabel), stmt)
	if err := t.resolveCondition(semCtx, stmt.Condition); err != nil {
		return err
	}
	stmt.Condition = t.flattener.Flatten(stmt.Condition)
	return t.resolveStmt(semCtx, stmt.Block)
}

func (t *TypeResolver) resolveCondition(semCtx *SemanticAnalysisContext, condition Expression) error {
	condType, err := t.evaluateType(semCtx, condition)
	if err != nil {
		return err
	}
	if !t.compatible(condType, types.BoolType) {
		return TypeError{Expected: types.BoolType, Got: condType}
	}
	return nil
}

func (t *TypeResolver) resolveForStmt(parent *SemanticAnalysisContext, stmt *ForStatement) error {
	semCtx := FromSemanticAnalysisContext(parent)
	stmt.LabelStart = NewLabel(LoopStart)
	stmt.LabelEnd = NewLabel(LoopEnd)
	stmt.LabelIncrement = NewLabel(IncrementLabel)
	semCtx.LabelEnv.New(string(LoopLabel), stmt)
	if err := t.resolveStmt(semCtx, stmt.VarDecl); err != nil {
		return err
	}
	if err := t.resolveCondition(semCtx, stmt.Condition); err != nil {
		return err
	}
	stmt.Condition = t.flattener.Flatten(stmt.Condition)
	if err := t.resolveStmt(semCtx, stmt.Increment); err != nil {
		return err
	}
	if err := t.resolveStmt(semCtx, stmt.Block); err != nil {
		return err
	}
	return nil
}

func (t *TypeResolver) evaluateType(semCtx *SemanticAnalysisContext, expression Expression) (types.Type, error) {
	switch typed := expression.(type) {
	case *BinaryExpression:
		t1, err := t.evaluateType(semCtx, typed.Left)
		if err != nil {
			return types.NotSpecified, err
		}
		t2, err := t.evaluateType(semCtx, typed.Right)
		if err != nil {
			return types.NotSpecified, err
		}
		if !t.compatible(t1, t2) || !t.compatibleOperation(t1, typed.Operator) {
			return types.NotSpecified, TypeError{Expected: t1, Got: t2}
		}
		return typed.ExprType, nil
	case *UnaryExpression:
		t1, err := t.evaluateType(semCtx, typed.Right)
		if err != nil {
			return types.NotSpecified, err
		}
		if !t.compatibleOperation(t1, typed.Operator) {
			return types.NotSpecified, NotCompatibleOperationError{Operation: typed.Operator, Type: t1}
		}
		return typed.ExprType, nil
	case *BooleanLiteral, *IntegerLiteral, *VoidLiteral:
		return typed.Type(), nil
	case *IdentifierLiteral:
		identifierType, ok := semCtx.VarEnv.Get(typed.Value)
		if !ok {
			return types.NotSpecified, fmt.Errorf("Variable %v does not exist", typed.Value)
		}
		typed.VarType = identifierType
		return identifierType, nil
	case *FunctionCall:
		return t.evaluateFunCall(semCtx, typed)
	}
	return types.NotSpecified, fmt.Errorf("Unexpected expression %v", expression)
}

func (t *TypeResolver) evaluateFunCall(semCtx *SemanticAnalysisContext, funCall *FunctionCall) (types.Type, error) {
	funDecl, ok := semCtx.FunEnv.Get(funCall.Name)
	if !ok {
		return types.NotSpecified, fmt.Errorf("Function %v does not exist", funCall.Name)
	}
	if len(funDecl.Arguments) != len(funCall.Arguments) {
		return types.NotSpecified, fmt.Errorf("Expected %v arguments. Got: %v", len(funDecl.Arguments), len(funCall.Arguments))
	}

	for i := 0; i < len(funDecl.Arguments); i++ {
		funArg := funDecl.Arguments[i]
		callArg := funCall.Arguments[i]
		callType, err := t.evaluateType(semCtx, callArg)
		if err != nil {
			return types.NotSpecified, err
		}
		if !t.compatible(funArg.VarType, callType) {
			return types.NotSpecified, TypeError{Expected: funArg.VarType, Got: callType}
		}
	}
	funCall.ReturnType = funDecl.ReturnType
	return funDecl.ReturnType, nil
}

func (t *TypeResolver) compatible(t1, t2 types.Type) bool {
	if t1 == types.IntType && t2 == types.IntPointer || t2 == types.IntType && t1 == types.IntPointer {
		return true
	}
	return t1 == t2
}

func (t *TypeResolver) compatibleOperation(t1 types.Type, op token.TokenType) bool {
	return true // TODO: implement operation compatibility
}
