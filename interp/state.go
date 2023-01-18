package interp

import (
	"fmt"

	"github.com/fj68/new-lang/ast"
	"github.com/fj68/new-lang/parser"
)

type State struct {
	RetVal Value
}

func New() *State {
	return &State{}
}

func Eval(text []rune) error {
	p := New()
	return p.Eval(text)
}

func (p *State) Eval(text []rune) error {
	program, err := parser.Parse(text)
	if err != nil {
		return err
	}
	return p.evalProgram(program)
}

func (p *State) evalProgram(program []ast.Stmt) error {
	for _, stmt := range program {
		if err := p.evalStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (p *State) evalStmt(stmt ast.Stmt) error {
	switch v := stmt.(type) {
	case *ast.ExprStmt:
		return p.evalExprStmt(v)
	case *ast.ReturnStmt:
		return p.evalReturnStmt(v)
	default:
		return fmt.Errorf("unknown stmt: %s", v.Inspect())
	}
}

func (p *State) evalReturnStmt(stmt *ast.ReturnStmt) error {
	if stmt.Value == nil {
		return nil
	}
	value, err := p.evalExpr(stmt.Value)
	if err != nil {
		return err
	}
	p.RetVal = value
	return nil
}

func (p *State) evalExprStmt(stmt *ast.ExprStmt) error {
	_, err := p.evalExpr(stmt.Expr)
	if err != nil {
		return err
	}
	return nil
}

func (p *State) evalExpr(expr ast.Expr) (Value, error) {
	switch v := expr.(type) {
	case *ast.BoolLiteralExpr:
		return VBool(v.Value), nil
	case *ast.NumberLiteralExpr:
		return VNumber(v.Value), nil
	case *ast.StringLiteralExpr:
		return VString(v.Value), nil
	default:
		return nil, fmt.Errorf("unknown expr: %s", v.Inspect())
	}
}
