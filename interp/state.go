package interp

import (
	"fmt"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/parser"
	"github.com/fj68/vvlang/stack"
)

type State struct {
	Env     *Env
	RetVals stack.Stack[Value]
}

func NewState() *State {
	return &State{
		Env: NewEnv(nil),
	}
}

func Eval(text []rune) (Value, error) {
	s := NewState()
	return s.Eval(text)
}

func (s *State) RegisterBuiltin(name string, value Value) {
	s.Env.Values[name] = value
}

func (s *State) Eval(text []rune) (Value, error) {
	program, err := parser.Parse(text)
	if err != nil {
		return nil, err
	}
	err = s.evalProgram(program)
	if err != nil {
		return nil, err
	}

	return s.RetVals.Pop(), nil
}

func (s *State) pushEnv() {
	s.Env = NewEnv(s.Env)
}

func (s *State) popEnv() {
	s.Env = s.Env.outer
}

func (s *State) evalProgram(program []ast.Stmt) error {
	for _, stmt := range program {
		if err := s.evalStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) evalBody(body []ast.Stmt) error {
	// TODO: prepare s.evalBodyStmt
	for _, stmt := range body {
		if err := s.evalStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) evalStmt(stmt ast.Stmt) error {
	switch v := stmt.(type) {
	case *ast.ExprStmt:
		return s.evalExprStmt(v)
	case *ast.ReturnStmt:
		return s.evalReturnStmt(v)
	case *ast.VarDeclStmt:
		return s.evalVarDeclStmt(v)
	case *ast.IfStmt:
		return s.evalIfStmt(v)
	default:
		return fmt.Errorf("unknown stmt: %s", v.Inspect())
	}
}

func (s *State) evalReturnStmt(stmt *ast.ReturnStmt) error {
	if stmt.Value == nil {
		return nil
	}
	value, err := s.evalExpr(stmt.Value)
	if err != nil {
		return err
	}
	s.RetVals.Push(value)
	return nil
}

func (s *State) evalIfStmt(stmt *ast.IfStmt) error {
	v, err := s.evalExpr(stmt.Cond)
	if err != nil {
		return err
	}
	cond, ok := v.(VBool)
	if !ok {
		return fmt.Errorf("expected bool, but got %s", v.Type())
	}
	if cond {
		return s.evalBody(stmt.Then)
	}
	if stmt.Else == nil {
		return nil
	}
	return s.evalBody(stmt.Else)
}

func (s *State) evalExprStmt(stmt *ast.ExprStmt) error {
	_, err := s.evalExpr(stmt.Expr)
	if err != nil {
		return err
	}
	return nil
}

func (s *State) evalVarDeclStmt(stmt *ast.VarDeclStmt) error {
	v, err := s.evalExpr(stmt.Body)
	if err != nil {
		return err
	}
	s.Env.Set(stmt.Name, v)
	return nil
}

func (s *State) evalExpr(expr ast.Expr) (Value, error) {
	switch v := expr.(type) {
	case *ast.BoolLiteralExpr:
		return VBool(v.Value), nil
	case *ast.NumberLiteralExpr:
		return VNumber(v.Value), nil
	case *ast.StringLiteralExpr:
		return VString(v.Value), nil
	case *ast.FunLiteralExpr:
		return s.evalFunLiteralExpr(v)
	case *ast.FunCallExpr:
		return s.evalFunCallExpr(v)
	case *ast.VarRefExpr:
		return s.evalVarRefExpr(v)
	case *ast.InfixExpr:
		return s.evalInfixExpr(v)
	default:
		return nil, fmt.Errorf("unknown expr: %s", v.Inspect())
	}
}

func (s *State) evalFunLiteralExpr(expr *ast.FunLiteralExpr) (Value, error) {
	f := &VUserFun{expr.Args, expr.Body}
	if expr.Name != "" {
		s.Env.Set(expr.Name, f)
	}
	return f, nil
}

func (s *State) evalFunCallExpr(expr *ast.FunCallExpr) (Value, error) {
	f, err := s.evalExpr(expr.Fun)
	if err != nil {
		return nil, err
	}
	if f, ok := f.(*VUserFun); ok {
		args, err := s.evalArgs(expr.Args)
		if err != nil {
			return nil, err
		}
		return s.callUserFun(f, args)
	}
	if f, ok := f.(VBuiltinFun); ok {
		args, err := s.evalArgs(expr.Args)
		if err != nil {
			return nil, err
		}
		return s.callBuiltinFun(f, args)
	}
	return nil, fmt.Errorf("unable to call %s", f.Type())
}

func (s *State) evalArgs(exprs []ast.Expr) ([]Value, error) {
	var args []Value
	for _, expr := range exprs {
		value, err := s.evalExpr(expr)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}
	return args, nil
}

func (s *State) callUserFun(f *VUserFun, args []Value) (Value, error) {
	if len(f.Args) != len(args) {
		return nil, fmt.Errorf("not enough or too much arguments")
	}

	s.pushEnv()
	defer s.popEnv()

	for i, arg := range args {
		s.Env.Values[f.Args[i]] = arg
	}

	if err := s.evalBody(f.Body); err != nil {
		return nil, err
	}

	return s.RetVals.Pop(), nil
}

func (s *State) callBuiltinFun(f VBuiltinFun, args []Value) (Value, error) {
	return f(s, args)
}

func (s *State) evalVarRefExpr(expr *ast.VarRefExpr) (Value, error) {
	return s.Env.Get(expr.Name)
}

func (s *State) evalInfixExpr(expr *ast.InfixExpr) (Value, error) {
	switch expr.Op {
	case "+":
		return s.evalAddExpr(expr)
	default:
		return nil, fmt.Errorf("unknown operator: %s", expr.Op)
	}
}

func (s *State) evalAddExpr(expr *ast.InfixExpr) (Value, error) {
	left, err := s.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	lvalue, ok := left.(VNumber)
	if !ok {
		return nil, fmt.Errorf("left side value of add expression is not a number")
	}
	right, err := s.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	rvalue, ok := right.(VNumber)
	if !ok {
		return nil, fmt.Errorf("right side value of add expression is not a number")
	}
	return VNumber(lvalue + rvalue), nil
}
