package ast

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Inspect() string
	Equals(other Stmt) bool
}

type BreakStmt struct{}

func (stmt *BreakStmt) Inspect() string {
	return "BreakStmt"
}

func (stmt *BreakStmt) Equals(other Stmt) bool {
	_, ok := other.(*BreakStmt)
	return ok
}

type ContinueStmt struct{}

func (stmt *ContinueStmt) Inspect() string {
	return "ContinueStmt"
}

func (stmt *ContinueStmt) Equals(other Stmt) bool {
	_, ok := other.(*ContinueStmt)
	return ok
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

func (stmt *ReturnStmt) Equals(other Stmt) bool {
	o, ok := other.(*ReturnStmt)
	if !ok {
		return false
	}
	if stmt.Value == nil || o.Value == nil {
		return stmt.Value == o.Value
	}
	return stmt.Value.Equals(o.Value)
}

type WhileStmt struct {
	Cond Expr
	Body *BlockStmt
}

func (stmt *WhileStmt) Inspect() string {
	return fmt.Sprintf("WhileStmt{%s, %s}", stmt.Cond.Inspect(), stmt.Body.Inspect())
}

func (stmt *WhileStmt) Equals(other Stmt) bool {
	o, ok := other.(*WhileStmt)
	if !ok {
		return false
	}
	return stmt.Cond.Equals(o.Cond) && stmt.Body.Equals(o.Body)
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

func (stmt *IfStmt) Equals(other Stmt) bool {
	o, ok := other.(*IfStmt)
	if !ok {
		return false
	}
	if !stmt.Cond.Equals(o.Cond) || !stmt.Then.Equals(o.Then) {
		return false
	}
	if stmt.Else == nil || o.Else == nil {
		return stmt.Else == o.Else
	}
	return stmt.Else.Equals(o.Else)
}

type VarAssignStmt struct {
	Name string
	Body Expr
}

func (stmt *VarAssignStmt) Inspect() string {
	return fmt.Sprintf("VarAssignStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

func (stmt *VarAssignStmt) Equals(other Stmt) bool {
	o, ok := other.(*VarAssignStmt)
	if !ok {
		return false
	}
	return stmt.Name == o.Name && stmt.Body.Equals(o.Body)
}

type VarDeclStmt struct {
	Name      string
	Body      Expr
	Exported  bool
	Docstring map[string]string
}

func (stmt *VarDeclStmt) Inspect() string {
	return fmt.Sprintf("VarDeclStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

func (stmt *VarDeclStmt) Equals(other Stmt) bool {
	o, ok := other.(*VarDeclStmt)
	if !ok {
		return false
	}
	return stmt.Name == o.Name && stmt.Body.Equals(o.Body) && stmt.Exported == o.Exported
}

type RecFunDeclStmt struct {
	Funs []*VarDeclStmt
}

func (stmt *RecFunDeclStmt) Inspect() string {
	var funs []string
	for _, f := range stmt.Funs {
		funs = append(funs, f.Inspect())
	}
	return fmt.Sprintf("RecFunDeclStmt{[%s]}", strings.Join(funs, ", "))
}

func (stmt *RecFunDeclStmt) Equals(other Stmt) bool {
	o, ok := other.(*RecFunDeclStmt)
	if !ok {
		return false
	}
	if len(stmt.Funs) != len(o.Funs) {
		return false
	}
	for i := range stmt.Funs {
		if !stmt.Funs[i].Equals(o.Funs[i]) {
			return false
		}
	}
	return true
}

type AssignStmt struct {
	Name string
	Body Expr
}

func (stmt *AssignStmt) Inspect() string {
	return fmt.Sprintf("AssignStmt{\"%s\", %s}", stmt.Name, stmt.Body.Inspect())
}

func (stmt *AssignStmt) Equals(other Stmt) bool {
	o, ok := other.(*AssignStmt)
	if !ok {
		return false
	}
	return stmt.Name == o.Name && stmt.Body.Equals(o.Body)
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

func (stmt *BlockStmt) Equals(other Stmt) bool {
	o, ok := other.(*BlockStmt)
	if !ok {
		return false
	}
	if len(stmt.Body) != len(o.Body) {
		return false
	}
	for i := range stmt.Body {
		if !stmt.Body[i].Equals(o.Body[i]) {
			return false
		}
	}
	return true
}

type ExprStmt struct {
	Expr
}

func (stmt *ExprStmt) Inspect() string {
	return stmt.Expr.Inspect()
}

func (stmt *ExprStmt) Equals(other Stmt) bool {
	o, ok := other.(*ExprStmt)
	if !ok {
		return false
	}
	return stmt.Expr.Equals(o.Expr)
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

func (stmt *TestStmt) Equals(other Stmt) bool {
	o, ok := other.(*TestStmt)
	if !ok {
		return false
	}
	if stmt.Name != o.Name || len(stmt.Body) != len(o.Body) {
		return false
	}
	for i := range stmt.Body {
		if !stmt.Body[i].Equals(o.Body[i]) {
			return false
		}
	}
	return true
}

type AssertStmt struct {
	Cond Expr
}

func (stmt *AssertStmt) Inspect() string {
	return fmt.Sprintf("AssertStmt{%s}", stmt.Cond.Inspect())
}

func (stmt *AssertStmt) Equals(other Stmt) bool {
	o, ok := other.(*AssertStmt)
	if !ok {
		return false
	}
	return stmt.Cond.Equals(o.Cond)
}

type DeferStmt struct {
	Body Expr
}

func (stmt *DeferStmt) Inspect() string {
	return fmt.Sprintf("DeferStmt{%s}", stmt.Body.Inspect())
}

func (stmt *DeferStmt) Equals(other Stmt) bool {
	o, ok := other.(*DeferStmt)
	if !ok {
		return false
	}
	return stmt.Body.Equals(o.Body)
}

type ExternStmt struct {
	Type     string // e.g. "native"
	Name     string
	Exported bool
}

func (stmt *ExternStmt) Inspect() string {
	return fmt.Sprintf("ExternStmt{\"%s\", \"%s\"}", stmt.Type, stmt.Name)
}

func (stmt *ExternStmt) Equals(other Stmt) bool {
	o, ok := other.(*ExternStmt)
	if !ok {
		return false
	}
	return stmt.Type == o.Type && stmt.Name == o.Name && stmt.Exported == o.Exported
}

type ImportStmt struct {
	Alias string
	Path  string
}

func (stmt *ImportStmt) Inspect() string {
	return fmt.Sprintf("ImportStmt{\"%s\", \"%s\"}", stmt.Alias, stmt.Path)
}

func (stmt *ImportStmt) Equals(other Stmt) bool {
	o, ok := other.(*ImportStmt)
	if !ok {
		return false
	}
	return stmt.Alias == o.Alias && stmt.Path == o.Path
}

type Module struct {
	Statements []Stmt
	Exports    map[string]Stmt
	Docstring  map[string]string
}

func (m *Module) Inspect() string {
	var body []string
	for _, s := range m.Statements {
		body = append(body, s.Inspect())
	}
	return fmt.Sprintf("Module{[%s]}", strings.Join(body, ", "))
}

func (m *Module) Equals(other *Module) bool {
	if m == nil || other == nil {
		return m == other
	}
	if len(m.Statements) != len(other.Statements) {
		return false
	}
	for i := range m.Statements {
		if !m.Statements[i].Equals(other.Statements[i]) {
			return false
		}
	}
	return true
}
