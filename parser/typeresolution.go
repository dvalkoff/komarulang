package parser

import (
	"fmt"

	"github.com/dvalkoff/komarulang/env"
	token "github.com/dvalkoff/komarulang/tokenizer"
)

type TokenEnv = *env.Environment[token.VarType]

type LoopStmt interface {
	Statement
}

type SemanticAnalysisContext struct {
	VarEnv   *env.Environment[token.VarType]
	FunEnv   *env.Environment[*FunctionDecl]
	LabelEnv *env.Environment[LoopStmt]
}

func FromSemanticAnalysisContext(parent *SemanticAnalysisContext) *SemanticAnalysisContext {
	if parent == nil {
		return newSemanticAnalysisContext(nil, nil, nil)
	}
	return newSemanticAnalysisContext(parent.VarEnv, parent.FunEnv, parent.LabelEnv)
}

func newSemanticAnalysisContext(varEnv *env.Environment[token.VarType], funEnv *env.Environment[*FunctionDecl], lavelEnv *env.Environment[LoopStmt]) *SemanticAnalysisContext {
	return &SemanticAnalysisContext{
		VarEnv:   env.NewEnvironment(varEnv),
		FunEnv:   env.NewEnvironment(funEnv),
		LabelEnv: env.NewEnvironment(lavelEnv),
	}
}

type TypeResolver struct {}

func (t *TypeResolver) Resolve(stmts []Statement) error {
	semCtx := FromSemanticAnalysisContext(nil)
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
		return err
	case *PrintStatement:
		_, err := t.evaluateType(semCtx, typed.Expr)
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
	}
	return fmt.Errorf("Unexpected stmt %v", stmt)
}

func (t *TypeResolver) resolveFunctionDecl(parent *SemanticAnalysisContext, funcDecl *FunctionDecl) error {
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

func (t *TypeResolver) resolveVarDeclaration(semCtx *SemanticAnalysisContext, varDecl *VarDeclaration) error {
	specifiedType := varDecl.VarType
	calculatedType, err := t.evaluateType(semCtx, varDecl.Expr)
	if err != nil {
		return err
	}
	if semCtx.VarEnv.Exists(varDecl.Identifier) {
		return fmt.Errorf("Variable %v already exists", varDecl.Identifier)
	}
	if specifiedType != token.NotSpecified && !t.compatible(specifiedType, calculatedType) {
		return TypeError{Expected: specifiedType, Got: calculatedType}
	}
	if specifiedType == token.NotSpecified {
		varDecl.VarType = calculatedType
	}
	semCtx.VarEnv.New(varDecl.Identifier, varDecl.VarType)
	return nil
}

func (t *TypeResolver) resolveVarAssignment(semCtx *SemanticAnalysisContext, assignment *VarAssignment) error {
	identifierType, ok := semCtx.VarEnv.Get(assignment.Identifier)
	if !ok {
		return fmt.Errorf("Variable %v does not exist", assignment.Identifier)
	}
	calculatedType, err := t.evaluateType(semCtx, assignment.Expr)
	if err != nil {
		return err
	}
	if !t.compatible(identifierType, calculatedType) {
		return TypeError{Expected: identifierType, Got: calculatedType}
	}
	return nil
}

func (t *TypeResolver) resolveIfStmt(semCtx *SemanticAnalysisContext, stmt *IfStatement) error {
	if err := t.resolveCondition(semCtx, stmt.Condition); err != nil {
		return err
	}
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
	return t.resolveStmt(semCtx, stmt.Block)
}

func (t *TypeResolver) resolveCondition(semCtx *SemanticAnalysisContext, condition Expression) error {
	condType, err := t.evaluateType(semCtx, condition)
	if err != nil {
		return err
	}
	if !t.compatible(condType, token.BoolType) {
		return TypeError{Expected: token.BoolType, Got: condType}
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
	if err := t.resolveStmt(semCtx, stmt.Increment); err != nil {
		return err
	}
	if err := t.resolveStmt(semCtx, stmt.Block); err != nil {
		return err
	}
	return nil
}

func (t *TypeResolver) evaluateType(semCtx *SemanticAnalysisContext, expression Expression) (token.VarType, error) {
	switch typed := expression.(type) {
	case *BinaryExpression:
		t1, err := t.evaluateType(semCtx, typed.Left)
		if err != nil {
			return token.NotSpecified, err
		}
		t2, err := t.evaluateType(semCtx, typed.Right)
		if err != nil {
			return token.NotSpecified, err
		}
		if !t.compatible(t1, t2) || !t.compatibleOperation(t1, typed.Operator) {
			return token.NotSpecified, TypeError{Expected: t1, Got: t2}
		}
		return typed.ExprType, nil
	case *UnaryExpression:
		t1, err := t.evaluateType(semCtx, typed.Right)
		if err != nil {
			return token.NotSpecified, err
		}
		if !t.compatibleOperation(t1, typed.Operator) {
			return token.NotSpecified, NotCompatibleOperationError{Operation: typed.Operator, Type: t1}
		}
		return typed.ExprType, nil
	case *BooleanLiteral, *IntegerLiteral:
		return typed.Type(), nil
	case *IdentifierLiteral:
		identifierType, ok := semCtx.VarEnv.Get(typed.Value)
		if !ok {
			return token.NotSpecified, fmt.Errorf("Variable %v does not exist", typed.Value)
		}
		return identifierType, nil
	case *FunctionCall:
		return token.NotSpecified, nil // TODO
	}
	return token.NotSpecified, fmt.Errorf("Unexpected expression %v", expression)
}

func (t *TypeResolver) compatible(t1, t2 token.VarType) bool {
	return t1 == t2
}

func (t *TypeResolver) compatibleOperation(t1 token.VarType, op token.TokenType) bool {
	return true // TODO: implement operation compatibility
}
