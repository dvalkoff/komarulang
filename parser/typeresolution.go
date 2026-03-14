package parser

import (
	"fmt"

	"github.com/dvalkoff/komarulang/env"
	token "github.com/dvalkoff/komarulang/tokenizer"
)


type TypeResolver struct {

}

type TokenEnv = *env.Environment[token.VarType]


func (t *TypeResolver) Resolve(stmts []Statement) error {
	env := env.NewEnvironment[token.VarType](nil)
	for _, stmt := range stmts {
		err := t.resolveStmt(env, stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TypeResolver) resolveStmt(env TokenEnv, stmt Statement) error {
	switch typed := stmt.(type) {
	case *Block:
		return t.resolveBlock(env, typed)
	case *VarDeclaration:
		return t.resolveVarDeclaration(env, typed)
	case *VarAssignment:
		return t.resolveVarAssignment(env, typed)
	case *ExprStatement:
		_, err := t.evaluateType(env, typed.Expr)
		return err
	case *PrintStatement:
		_, err := t.evaluateType(env, typed.Expr)
		return err
	case *IfStatement:
		return t.resolveIfStmt(env, typed)
	case *WhileStatement:
		return t.resolveWhileStmt(env, typed)
	case *ForStatement:
		return t.resolveForStmt(env, typed)
	}
	return fmt.Errorf("Unexpected stmt %v", stmt)
}

func (t *TypeResolver) resolveBlock(parent TokenEnv, block *Block) error {
	env := env.NewEnvironment(parent)
	for _, stmt := range block.Stmts {
		err := t.resolveStmt(env, stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TypeResolver) resolveVarDeclaration(env TokenEnv, varDecl *VarDeclaration) error {
	specifiedType := varDecl.VarType
	calculatedType, err := t.evaluateType(env, varDecl.Expr)
	if err != nil {
		return err
	}
	if specifiedType != token.NotSpecified && !t.compatible(specifiedType, calculatedType) {
		return TypeError{Expected: specifiedType, Got: calculatedType}
	}
	if specifiedType == token.NotSpecified {
		varDecl.VarType = calculatedType
	}
	env.New(varDecl.Identifier, varDecl.VarType)
	return nil
}

func (t *TypeResolver) resolveVarAssignment(env TokenEnv, assignment *VarAssignment) error {
	identifierType, ok := env.Get(assignment.Identifier)
	if !ok {
		return fmt.Errorf("Variable %v does not exist", assignment.Identifier)
	}
	calculatedType, err := t.evaluateType(env, assignment.Expr)
	if err != nil {
		return err
	}
	if !t.compatible(identifierType, calculatedType) {
		return TypeError{Expected: identifierType, Got: calculatedType}
	}
	return nil
}

func (t *TypeResolver) resolveIfStmt(env TokenEnv, stmt *IfStatement) error {
	if err := t.resolveCondition(env, stmt.Condition); err != nil {
		return err
	}
	if err := t.resolveStmt(env, stmt.Block); err != nil {
		return err
	}
	if stmt.ElseBlock != nil {
		return t.resolveStmt(env, stmt.ElseBlock)
	}
	return nil
}

func (t *TypeResolver) resolveWhileStmt(env TokenEnv, stmt *WhileStatement) error {
	if err := t.resolveCondition(env, stmt.Condition); err != nil {
		return err
	}
	return t.resolveStmt(env, stmt.Block)
}

func (t *TypeResolver) resolveCondition(env TokenEnv, condition Expression) error {
	condType, err := t.evaluateType(env, condition)
	if err != nil {
		return err
	}
	if !t.compatible(condType, token.BoolType) {
		return TypeError{Expected: token.BoolType, Got: condType}
	}
	return nil
}

func (t *TypeResolver) resolveForStmt(parent TokenEnv, stmt *ForStatement) error {
	env := env.NewEnvironment(parent)
	if err := t.resolveStmt(env, stmt.VarDecl); err != nil {
		return err
	}
	if err := t.resolveCondition(env, stmt.Condition); err != nil {
		return err
	}
	if err := t.resolveStmt(env, stmt.Increment); err != nil {
		return err
	}
	if err := t.resolveStmt(env, stmt.Block); err != nil {
		return err
	}
	return nil
}

func (t *TypeResolver) evaluateType(env TokenEnv, expression Expression) (token.VarType, error) {
	switch typed := expression.(type) {
	case *BinaryExpression:
		t1, err := t.evaluateType(env, typed.Left)
		if err != nil {
			return token.NotSpecified, err
		}
		t2, err := t.evaluateType(env, typed.Right)
		if err != nil {
			return token.NotSpecified, err
		}
		if !t.compatible(t1, t2) || !t.compatibleOperation(t1, typed.Operator) {
			return token.NotSpecified, TypeError{Expected: t1, Got: t2}
		}
		return typed.ExprType, nil
	case *UnaryExpression:
		t1, err := t.evaluateType(env, typed.Right)
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
		identifierType, ok := env.Get(typed.Value)
		if !ok {
			return token.NotSpecified, fmt.Errorf("Variable %v does not exist", typed.Value)
		}
		return identifierType, nil
	}
	return token.NotSpecified, fmt.Errorf("Unexpected expression %v", expression)
}

func (t *TypeResolver) compatible(t1, t2 token.VarType) bool {
	return t1 == t2
}

func (t *TypeResolver) compatibleOperation(t1 token.VarType, op token.TokenType) bool {
	return true // TODO: implement operation compatibility
}
