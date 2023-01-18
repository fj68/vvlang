package ast

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Inspect() string
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
	for _, s := range stmt.Then {
		elseBody = append(elseBody, s.Inspect())
	}
	return fmt.Sprintf("WhileStmt{%s, %s, %s}", stmt.Cond.Inspect(), strings.Join(thenBody, ", "), strings.Join(elseBody, ", "))
}

type VarDeclStmt struct {
	Name string
	Body Expr
}

func (stmt *VarDeclStmt) Inspect() string {
	return fmt.Sprintf("VarDeclStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

type VarAssignStmt struct {
	VarRef *VarRefExpr
	Body   Expr
}

func (stmt *VarAssignStmt) Inspect() string {
	return fmt.Sprintf("VarAssignStmt{%s, %s}", stmt.VarRef.Inspect(), stmt.Body.Inspect())
}

type ExprStmt struct {
	Expr
}

func (stmt *ExprStmt) Inspect() string {
	return stmt.Expr.Inspect()
}
