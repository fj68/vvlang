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
	Body *BlockStmt
}

func (stmt *WhileStmt) Inspect() string {
	return fmt.Sprintf("WhileStmt{%s, %s}", stmt.Cond.Inspect(), stmt.Body)
}

type IfStmt struct {
	Cond Expr
	Then *BlockStmt
	Else *BlockStmt
}

func (stmt *IfStmt) Inspect() string {
	if stmt.Else == nil {
		return fmt.Sprintf("IfStmt{%s, %s}", stmt.Cond.Inspect(), stmt.Then.Inspect())
	}
	return fmt.Sprintf("IfStmt{%s, %s, %s}", stmt.Cond.Inspect(), stmt.Then.Inspect(), stmt.Else.Inspect())
}

type VarAssignStmt struct {
	Name string
	Body Expr
}

func (stmt *VarAssignStmt) Inspect() string {
	return fmt.Sprintf("VarAssignStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

type VarDeclStmt struct {
	Name     string
	Body     Expr
	Exported bool
}

func (stmt *VarDeclStmt) Inspect() string {
	return fmt.Sprintf("VarDeclStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

type AssignStmt struct {
	Name string
	Body Expr
}

func (stmt *AssignStmt) Inspect() string {
	return fmt.Sprintf("AssignStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

type BlockStmt struct {
	Body []Stmt
}

func (stmt *BlockStmt) Inspect() string {
	var body []string
	for _, s := range stmt.Body {
		body = append(body, s.Inspect())
	}
	return fmt.Sprintf("BlockStmt{[%s]}", strings.Join(body, ", "))
}

type ExprStmt struct {
	Expr
}

func (stmt *ExprStmt) Inspect() string {
	return stmt.Expr.Inspect()
}

type TestStmt struct {
	Name string
	Body []Stmt
}

func (stmt *TestStmt) Inspect() string {
	var body []string
	for _, s := range stmt.Body {
		body = append(body, s.Inspect())
	}
	return fmt.Sprintf("TestStmt{\"%s\", [%s]}", stmt.Name, strings.Join(body, ", "))
}

type AssertStmt struct {
	Cond Expr
}

func (stmt *AssertStmt) Inspect() string {
	return fmt.Sprintf("AssertStmt{%s}", stmt.Cond.Inspect())
}

type DeferStmt struct {
	Body Expr
}

func (stmt *DeferStmt) Inspect() string {
	return fmt.Sprintf("DeferStmt{%s}", stmt.Body.Inspect())
}

type ExternStmt struct {
	Type     string // e.g. "native"
	Name     string
	Exported bool
}

func (stmt *ExternStmt) Inspect() string {
	return fmt.Sprintf("ExternStmt{\"%s\", \"%s\"}", stmt.Type, stmt.Name)
}

type ImportStmt struct {
	Alias string
	Path  string
}

func (stmt *ImportStmt) Inspect() string {
	return fmt.Sprintf("ImportStmt{\"%s\", \"%s\"}", stmt.Alias, stmt.Path)
}

type Module struct {
	Statements []Stmt
	Exports    map[string]Stmt
}

func (m *Module) Inspect() string {
	var body []string
	for _, s := range m.Statements {
		body = append(body, s.Inspect())
	}
	return fmt.Sprintf("Module{[%s]}", strings.Join(body, ", "))
}
