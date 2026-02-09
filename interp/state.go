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

func Eval(text []rune) error {
	s := NewState()
	return s.Eval(text)
}

func (s *State) RegisterGlobal(name string, value Value) {
	s.Env.Values[name] = value
}

func (s *State) RegisterGlobals(values map[string]Value) {
	for name, value := range values {
		s.RegisterGlobal(name, value)
	}
}

func (s *State) Eval(text []rune) error {
	program, err := parser.Parse(text)
	if err != nil {
		return err
	}
	return s.evalProgram(program)
}

func (s *State) pushEnv() {
	s.Env = NewEnv(s.Env)
}

func (s *State) popEnv() {
	s.Env = s.Env.outer
}

var ErrBreak = fmt.Errorf("break")
var ErrContinue = fmt.Errorf("continue")
var ErrReturn = fmt.Errorf("return")

func (s *State) evalProgram(program []ast.Stmt) error {
	for _, stmt := range program {
		if err := s.evalStmt(stmt); err != nil {
			if err == ErrReturn {
				// Top-level return: stop program execution but do not treat as an error
				return nil
			}
			return err
		}
	}
	return nil
}

func (s *State) evalBody(body []ast.Stmt) error {
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
	case *ast.BreakStmt:
		return ErrBreak
	case *ast.ContinueStmt:
		return ErrContinue
	case *ast.WhileStmt:
		return s.evalWhileStmt(v)
	default:
		return fmt.Errorf("unknown stmt: %s", v.Inspect())
	}
}

func (s *State) evalReturnStmt(stmt *ast.ReturnStmt) error {
	if stmt.Value == nil {
		// return without value still signals return control flow
		return ErrReturn
	}
	value, err := s.evalExpr(stmt.Value)
	if err != nil {
		return err
	}
	s.RetVals.Push(value)
	return ErrReturn
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
	case *ast.RecordLiteralExpr:
		return s.evalRecordLiteralExpr(v)
	case *ast.FieldAccessExpr:
		return s.evalFieldAccessExpr(v)
	case *ast.FunLiteralExpr:
		return s.evalFunLiteralExpr(v)
	case *ast.FunCallExpr:
		return s.evalFunCallExpr(v)
	case *ast.VarRefExpr:
		return s.evalVarRefExpr(v)
	case *ast.InfixExpr:
		return s.evalInfixExpr(v)
	case *ast.ListLiteralExpr:
		return s.evalListLiteralExpr(v)
	case *ast.IndexExpr:
		return s.evalIndexExpr(v)
	case *ast.SliceExpr:
		return s.evalSliceExpr(v)
	case *ast.SpreadExpr:
		return s.evalSpreadExpr(v)
	case *ast.PrefixExpr:
		return s.evalPrefixExpr(v)
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

func (s *State) evalRecordLiteralExpr(expr *ast.RecordLiteralExpr) (Value, error) {
	m := map[string]Value{}
	// Process elements in order
	for _, elem := range expr.Elements {
		switch e := elem.(type) {
		case *ast.RecordField:
			val, err := s.evalExpr(e.Value)
			if err != nil {
				return nil, err
			}
			m[e.Key] = val
		case *ast.RecordSpread:
			val, err := s.evalExpr(e.Expr)
			if err != nil {
				return nil, err
			}
			// The value should be a record
			rec, ok := val.(*VRecord)
			if !ok {
				return nil, fmt.Errorf("cannot spread non-record value of type %s", val.Type())
			}
			// Add all fields from the spread record
			for k, v := range rec.Fields {
				m[k] = v
			}
		}
	}
	return &VRecord{Fields: m}, nil
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

	if err := s.evalBody(f.Body); err != nil && err != ErrReturn {
		return nil, err
	}

	return s.RetVals.Pop(), nil
}

func (s *State) evalWhileStmt(stmt *ast.WhileStmt) error {
	for {
		v, err := s.evalExpr(stmt.Cond)
		if err != nil {
			return err
		}
		cond, ok := v.(VBool)
		if !ok {
			return fmt.Errorf("expected bool, but got %s", v.Type())
		}
		if !cond {
			break
		}
		for _, st := range stmt.Body {
			err := s.evalStmt(st)
			if err == ErrContinue {
				// continue to next iteration
				break
			}
			if err == ErrBreak {
				// exit the while loop
				return nil
			}
			if err == ErrReturn {
				// propagate return up
				return ErrReturn
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *State) callBuiltinFun(f VBuiltinFun, args []Value) (Value, error) {
	s.pushEnv()
	defer s.popEnv()

	return f(s, args)
}

func (s *State) evalVarRefExpr(expr *ast.VarRefExpr) (Value, error) {
	return s.Env.Get(expr.Name)
}

func (s *State) evalInfixExpr(expr *ast.InfixExpr) (Value, error) {
	left, err := s.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := s.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case "+":
		return s.evalAddExpr(left, right)
	case "==":
		return s.evalEqualExpr(left, right)
	case "<":
		return s.evalLessThanExpr(left, right)
	case "<=":
		return s.evalLessThanEqualExpr(left, right)
	case "and":
		return s.evalAndExpr(left, right)
	case "or":
		return s.evalOrExpr(left, right)
	default:
		return nil, fmt.Errorf("unknown operator: %s", expr.Op)
	}
}

func (s *State) evalAddExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VNumber)
	if !ok {
		return nil, fmt.Errorf("left side value of add expression is not a number")
	}
	rvalue, ok := right.(VNumber)
	if !ok {
		return nil, fmt.Errorf("right side value of add expression is not a number")
	}
	return VNumber(lvalue + rvalue), nil
}

func (s *State) evalEqualExpr(left Value, right Value) (Value, error) {
	v, err := left.Equal(right)
	if err != nil {
		return nil, err
	}
	return VBool(v), nil
}

func (s *State) evalLessThanExpr(left Value, right Value) (Value, error) {
	v, err := left.LessThan(right)
	if err != nil {
		return nil, err
	}
	return VBool(v), nil
}

func (s *State) evalLessThanEqualExpr(left Value, right Value) (Value, error) {
	v, err := s.evalEqualExpr(left, right)
	if err != nil {
		return nil, err
	}
	if bool(v.(VBool)) {
		return v, nil
	}
	return s.evalLessThanExpr(left, right)
}

func (s *State) evalAndExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VBool)
	if !ok {
		return nil, fmt.Errorf("left side of and expr is expected bool, but got %s", left.Type())
	}
	rvalue, ok := right.(VBool)
	if !ok {
		return nil, fmt.Errorf("right side of and expr is expected bool, but got %s", right.Type())
	}
	return VBool(bool(lvalue) && bool(rvalue)), nil
}

func (s *State) evalOrExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VBool)
	if !ok {
		return nil, fmt.Errorf("left side of or expr is expected bool, but got %s", left.Type())
	}
	rvalue, ok := right.(VBool)
	if !ok {
		return nil, fmt.Errorf("right side of or expr is expected bool, but got %s", right.Type())
	}
	return VBool(bool(lvalue) || bool(rvalue)), nil
}

func (s *State) evalPrefixExpr(expr *ast.PrefixExpr) (Value, error) {
	right, err := s.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case "-":
		num, ok := right.(VNumber)
		if !ok {
			return nil, fmt.Errorf("cannot negate %s", right.Type())
		}
		return VNumber(-float64(num)), nil
	default:
		return nil, fmt.Errorf("unknown prefix operator: %s", expr.Op)
	}
}

func (s *State) evalListLiteralExpr(expr *ast.ListLiteralExpr) (Value, error) {
	var elements []Value
	for _, elemExpr := range expr.Elements {
		// Check if this is a spread expression
		if spread, ok := elemExpr.(*ast.SpreadExpr); ok {
			val, err := s.evalExpr(spread.Expr)
			if err != nil {
				return nil, err
			}
			// The value should be a list
			list, ok := val.(*VList)
			if !ok {
				return nil, fmt.Errorf("cannot spread non-list value of type %s", val.Type())
			}
			// Add all elements from the list
			elements = append(elements, list.Elements...)
		} else {
			elem, err := s.evalExpr(elemExpr)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)
		}
	}
	return &VList{Elements: elements}, nil
}

func (s *State) evalIndexExpr(expr *ast.IndexExpr) (Value, error) {
	left, err := s.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	index, err := s.evalExpr(expr.Index)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case *VList:
		idx, ok := index.(VNumber)
		if !ok {
			return nil, fmt.Errorf("list index must be a number, got %s", index.Type())
		}
		// Convert to int, handle negative indices
		intIdx := int(float64(idx))
		if intIdx < 0 {
			intIdx = len(l.Elements) + intIdx
		}
		if intIdx < 0 || intIdx >= len(l.Elements) {
			return nil, fmt.Errorf("list index out of range: %d", intIdx)
		}
		return l.Elements[intIdx], nil
	case VString:
		idx, ok := index.(VNumber)
		if !ok {
			return nil, fmt.Errorf("string index must be a number, got %s", index.Type())
		}
		// Convert to int, handle negative indices
		intIdx := int(float64(idx))
		str := string(l)
		if intIdx < 0 {
			intIdx = len(str) + intIdx
		}
		if intIdx < 0 || intIdx >= len(str) {
			return nil, fmt.Errorf("string index out of range: %d", intIdx)
		}
		return VString(string(str[intIdx])), nil
	default:
		return nil, fmt.Errorf("cannot index %s", left.Type())
	}
}

func (s *State) evalSliceExpr(expr *ast.SliceExpr) (Value, error) {
	left, err := s.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case *VList:
		var start, end int
		start = 0
		end = len(l.Elements)

		if expr.Start != nil {
			startVal, err := s.evalExpr(expr.Start)
			if err != nil {
				return nil, err
			}
			startNum, ok := startVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice start must be a number, got %s", startVal.Type())
			}
			start = int(float64(startNum))
			if start < 0 {
				start = len(l.Elements) + start
			}
		}

		if expr.End != nil {
			endVal, err := s.evalExpr(expr.End)
			if err != nil {
				return nil, err
			}
			endNum, ok := endVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice end must be a number, got %s", endVal.Type())
			}
			end = int(float64(endNum))
			if end < 0 {
				end = len(l.Elements) + end
			}
		}

		// Clamp to valid range
		if start < 0 {
			start = 0
		}
		if end > len(l.Elements) {
			end = len(l.Elements)
		}
		if start > end {
			start = end
		}

		return &VList{Elements: l.Elements[start:end]}, nil

	case VString:
		str := string(l)
		var start, end int
		start = 0
		end = len(str)

		if expr.Start != nil {
			startVal, err := s.evalExpr(expr.Start)
			if err != nil {
				return nil, err
			}
			startNum, ok := startVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice start must be a number, got %s", startVal.Type())
			}
			start = int(float64(startNum))
			if start < 0 {
				start = len(str) + start
			}
		}

		if expr.End != nil {
			endVal, err := s.evalExpr(expr.End)
			if err != nil {
				return nil, err
			}
			endNum, ok := endVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice end must be a number, got %s", endVal.Type())
			}
			end = int(float64(endNum))
			if end < 0 {
				end = len(str) + end
			}
		}

		// Clamp to valid range
		if start < 0 {
			start = 0
		}
		if end > len(str) {
			end = len(str)
		}
		if start > end {
			start = end
		}

		return VString(str[start:end]), nil

	default:
		return nil, fmt.Errorf("cannot slice %s", left.Type())
	}
}

func (s *State) evalSpreadExpr(expr *ast.SpreadExpr) (Value, error) {
	// Spread expressions should only appear inside list/record literals,
	// so this shouldn't be called directly.
	return nil, fmt.Errorf("spread expression can only be used inside list or record literals")
}

func (s *State) evalFieldAccessExpr(expr *ast.FieldAccessExpr) (Value, error) {
	recordVal, err := s.evalExpr(expr.Record)
	if err != nil {
		return nil, err
	}
	rec, ok := recordVal.(*VRecord)
	if !ok {
		return nil, fmt.Errorf("cannot access field on non-record value of type %s", recordVal.Type())
	}
	fieldVal, ok := rec.Fields[expr.Field]
	if !ok {
		return nil, fmt.Errorf("record does not have field '%s'", expr.Field)
	}
	return fieldVal, nil
}