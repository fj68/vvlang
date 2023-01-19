package ast

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Inspect() string
}

type BreakStmt struct{}

func (stmt *BreakStmt) Inspect() string {
	return "BreakStmt"
}

type ContinueStmt struct{}

func (stmt *ContinueStmt) Inspect() string {
	return "ContinueStmt"
}

type ReturnStmt struct {
	Value Expr
}

func (stmt *ReturnStmt) Inspect() string {
	if stmt.Value == nil {
		return "ReturnStmt{}"
	}
	return fmt.Sprintf("ReturnStmt{%s}", stmt.Value.Inspect())
}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
}

func (stmt *WhileStmt) Inspect() string {
	var body []string
	for _, s := range stmt.Body {
		body = append(body, s.Inspect())
	}
	return fmt.Sprintf("WhileStmt{%s, %s}", stmt.Cond.Inspect(), strings.Join(body, ", "))
}

type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (stmt *IfStmt) Inspect() string {
	var thenBody []string
	for _, s := range stmt.Then {
		thenBody = append(thenBody, s.Inspect())
	}
	var elseBody []string
	for _, s := range stmt.Else {
		elseBody = append(elseBody, s.Inspect())
	}
	return fmt.Sprintf("IfStmt{%s, %s, %s}", stmt.Cond.Inspect(), strings.Join(thenBody, ", "), strings.Join(elseBody, ", "))
}

type VarDeclStmt struct {
	Name string
	Body Expr
}

func (stmt *VarDeclStmt) Inspect() string {
	return fmt.Sprintf("VarDeclStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

type ExprStmt struct {
	Expr
}

func (stmt *ExprStmt) Inspect() string {
	return stmt.Expr.Inspect()
}
