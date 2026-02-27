package interp

import (
	"fmt"

	"github.com/fj68/vvlang/ast"
)

func (s *State) evalBuiltinCallExpr(expr *ast.BuiltinCallExpr) (Value, error) {
	switch expr.Op {
	case "not":
		return s.evalNotExpr(expr)
	case "type":
		return s.evalTypeExpr(expr)
	case "str":
		return s.evalStrExpr(expr)
	case "len":
		return s.evalLenExpr(expr)
	default:
		return nil, fmt.Errorf("unknown builtin op: %s", expr.Op)
	}
}

func (s *State) evalTypeExpr(expr *ast.BuiltinCallExpr) (Value, error) {
	v, err := s.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	return StringToValue(v.Type().String()), nil
}

func (s *State) evalNotExpr(expr *ast.BuiltinCallExpr) (Value, error) {
	v, err := s.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	b, ok := v.(VBool)
	if !ok {
		return nil, fmt.Errorf("argument for not() is expected bool, but got %s", v.Type())
	}
	return VBool(!bool(b)), nil
}

func (s *State) evalStrExpr(expr *ast.BuiltinCallExpr) (Value, error) {
	v, err := s.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	return StringToValue(v.Str()), nil
}

func (s *State) evalLenExpr(expr *ast.BuiltinCallExpr) (Value, error) {
	v, err := s.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	switch tv := v.(type) {
	case *VList:
		return VInt(len(tv.Elements)), nil
	default:
		return nil, fmt.Errorf("argument for len() is expected list, but got %s", v.Type())
	}
}

func (s *State) evalFunLiteralExpr(expr *ast.FunLiteralExpr) (Value, error) {
	return &VUserFun{s.Env, expr.Args, expr.Body}, nil
}

func (s *State) evalRecordLiteralExpr(expr *ast.RecordLiteralExpr) (Value, error) {
	m := map[string]Value{}
	for key, field := range expr.Fields {
		val, err := s.evalExpr(field)
		if err != nil {
			return nil, err
		}
		m[key] = val
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
	s.StackDepth++
	if s.StackDepth > s.MaxRecursionDepth {
		s.StackDepth--
		return nil, fmt.Errorf("call stack size overflow (max depth %d)", s.MaxRecursionDepth)
	}
	defer func() { s.StackDepth-- }()

	for {
		if len(f.Args) != len(args) {
			return nil, fmt.Errorf("not enough or too much arguments")
		}

		retVal, tc, err := func() (val Value, tc *VTailCall, err error) {
			oldEnv := s.Env
			s.Env = NewEnv(f.CapturedEnv)
			defer func() { s.Env = oldEnv }()

			s.pushDeferScope()
			defer func() {
				deferErr := s.popDeferScope()
				if err == nil {
					err = deferErr
				}
			}()

			for i, arg := range args {
				s.Env.Values[f.Args[i]] = arg
			}

			if err = s.evalBody(f.Body); err != nil && err != ErrReturn {
				return nil, nil, err
			}

			v := s.RetVals.Pop()
			if tailCall, ok := v.(*VTailCall); ok {
				return nil, tailCall, nil
			}
			return v, nil, nil
		}()

		if err != nil {
			return nil, err
		}

		if tc != nil {
			f = tc.Fun
			args = tc.Args
			continue
		}

		return retVal, nil
	}
}

func (s *State) callBuiltinFun(f VBuiltinFun, args []Value) (Value, error) {
	oldEnv := s.Env
	s.Env = NewEnv(s.Env)
	defer func() { s.Env = oldEnv }()

	return f(s, args)
}

func (s *State) evalVarRefExpr(expr *ast.VarRefExpr) (Value, error) {
	value, err := s.Env.Get(expr.Name)
	if err != nil {
		return nil, err
	}
	return value, nil
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
	case "-":
		return s.evalSubExpr(left, right)
	case "*":
		return s.evalMulExpr(left, right)
	case "/":
		return s.evalDivExpr(left, right)
	case "/.":
		return s.evalIDivExpr(left, right)
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
	case "%":
		return s.evalModExpr(left, right)
	default:
		return nil, fmt.Errorf("unknown operator: %s", expr.Op)
	}
}

func (s *State) evalAddExpr(left Value, right Value) (Value, error) {
	if l, ok := left.(VInt); ok {
		if r, ok := right.(VInt); ok {
			return VInt(l + r), nil
		}
		if r, ok := right.(VFloat); ok {
			return VFloat(float64(l) + float64(r)), nil
		}
		return nil, fmt.Errorf("right side value of add expression is not a number")
	}
	if l, ok := left.(VFloat); ok {
		if r, ok := right.(VFloat); ok {
			return VFloat(l + r), nil
		}
		if r, ok := right.(VInt); ok {
			return VFloat(float64(l) + float64(r)), nil
		}
		return nil, fmt.Errorf("right side value of add expression is not a number")
	}
	if l, ok := left.(*VList); ok {
		if r, ok := right.(*VList); ok {
			elems := make([]Value, 0, len(l.Elements)+len(r.Elements))
			elems = append(elems, l.Elements...)
			elems = append(elems, r.Elements...)
			return &VList{Elements: elems}, nil
		}
		return nil, fmt.Errorf("right side value of add expression is not a list: %T (%s)", right, right.String())
	}
	return nil, fmt.Errorf("unsupported types for '+': %T (%s) and %T (%s)", left, left.String(), right, right.String())
}

func (s *State) evalSubExpr(left Value, right Value) (Value, error) {
	if l, ok := left.(VInt); ok {
		if r, ok := right.(VInt); ok {
			return VInt(l - r), nil
		}
		if r, ok := right.(VFloat); ok {
			return VFloat(float64(l) - float64(r)), nil
		}
		return nil, fmt.Errorf("right side value of sub expression is not a number")
	}
	if l, ok := left.(VFloat); ok {
		if r, ok := right.(VFloat); ok {
			return VFloat(l - r), nil
		}
		if r, ok := right.(VInt); ok {
			return VFloat(float64(l) - float64(r)), nil
		}
		return nil, fmt.Errorf("right side value of sub expression is not a number")
	}
	return nil, fmt.Errorf("left side value of sub expression is not a number")
}

func (s *State) evalMulExpr(left Value, right Value) (Value, error) {
	if l, ok := left.(VInt); ok {
		if r, ok := right.(VInt); ok {
			return VInt(l * r), nil
		}
		if r, ok := right.(VFloat); ok {
			return VFloat(float64(l) * float64(r)), nil
		}
		return nil, fmt.Errorf("right side value of mul expression is not a number")
	}
	if l, ok := left.(VFloat); ok {
		if r, ok := right.(VFloat); ok {
			return VFloat(l * r), nil
		}
		if r, ok := right.(VInt); ok {
			return VFloat(float64(l) * float64(r)), nil
		}
		return nil, fmt.Errorf("right side value of mul expression is not a number")
	}
	return nil, fmt.Errorf("left side value of mul expression is not a number")
}

func (s *State) evalDivExpr(left Value, right Value) (Value, error) {
	var l, r float64
	if lv, ok := left.(VInt); ok {
		l = float64(lv)
	} else if lv, ok := left.(VFloat); ok {
		l = float64(lv)
	} else {
		return nil, fmt.Errorf("left side value of div expression is not a number")
	}

	if rv, ok := right.(VInt); ok {
		r = float64(rv)
	} else if rv, ok := right.(VFloat); ok {
		r = float64(rv)
	} else {
		return nil, fmt.Errorf("right side value of div expression is not a number")
	}

	if r == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return VFloat(l / r), nil
}

func (s *State) evalIDivExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VInt)
	if !ok {
		return nil, fmt.Errorf("left side value of floor div expression is not an int")
	}
	rvalue, ok := right.(VInt)
	if !ok {
		return nil, fmt.Errorf("right side value of floor div expression is not an int")
	}
	if rvalue == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return VInt(int64(lvalue) / int64(rvalue)), nil
}

func (s *State) evalModExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VInt)
	if !ok {
		return nil, fmt.Errorf("left side value of mod expression is not an int")
	}
	rvalue, ok := right.(VInt)
	if !ok {
		return nil, fmt.Errorf("right side value of mod expression is not an int")
	}
	if rvalue == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}
	return VInt(int64(lvalue) % int64(rvalue)), nil
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
		if num, ok := right.(VInt); ok {
			return VInt(-int64(num)), nil
		}
		if num, ok := right.(VFloat); ok {
			return VFloat(-float64(num)), nil
		}
		return nil, fmt.Errorf("cannot negate %s", right.Type())
	default:
		return nil, fmt.Errorf("unknown prefix operator: %s", expr.Op)
	}
}

func (s *State) evalListLiteralExpr(expr *ast.ListLiteralExpr) (Value, error) {
	var elements []Value
	for _, elemExpr := range expr.Elements {
		elem, err := s.evalExpr(elemExpr)
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
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
		idx, ok := index.(VInt)
		if !ok {
			return nil, fmt.Errorf("list index must be an int, got %s", index.Type())
		}
		intIdx := int(int64(idx))
		if intIdx < 0 {
			intIdx = len(l.Elements) + intIdx
		}
		if intIdx < 0 || intIdx >= len(l.Elements) {
			return nil, fmt.Errorf("list index out of range: %d", intIdx)
		}
		return l.Elements[intIdx], nil
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
			startNum, ok := startVal.(VInt)
			if !ok {
				return nil, fmt.Errorf("slice start must be an int, got %s", startVal.Type())
			}
			start = int(int64(startNum))
			if start < 0 {
				start = len(l.Elements) + start
			}
		}

		if expr.End != nil {
			endVal, err := s.evalExpr(expr.End)
			if err != nil {
				return nil, err
			}
			endNum, ok := endVal.(VInt)
			if !ok {
				return nil, fmt.Errorf("slice end must be an int, got %s", endVal.Type())
			}
			end = int(int64(endNum))
			if end < 0 {
				end = len(l.Elements) + end
			}
		}

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

	default:
		return nil, fmt.Errorf("cannot slice %s", left.Type())
	}
}

func (s *State) evalFieldAccessExpr(expr *ast.FieldAccessExpr) (Value, error) {
	recordVal, err := s.evalExpr(expr.Record)
	if err != nil {
		return nil, err
	}
	var record *VRecord
	switch tv := recordVal.(type) {
	case *VRecord:
		record = tv
	case *VModule:
		record = tv.VRecord
	default:
		return nil, fmt.Errorf("expected record for field access, but got %s", recordVal.Type())
	}
	fieldName := string(expr.Field)
	value, ok := record.Fields[fieldName]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found in record", fieldName)
	}
	return value, nil
}
