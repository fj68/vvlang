package interp

import (
	"fmt"
	"strings"

	"github.com/fj68/vvlang/ast"
)

type ValueType int

const (
	VTBool ValueType = iota
	VTNumber
	VTString
	VTUserFun
	VTBuiltinFun
	VTList
)

func (ty ValueType) String() string {
	switch ty {
	case VTBool:
		return "bool"
	case VTNumber:
		return "number"
	case VTString:
		return "string"
	case VTUserFun:
		return "fun"
	case VTBuiltinFun:
		return "fun"
	case VTList:
		return "list"
	}
	return "unknown"
}

type Value interface {
	Type() ValueType
	String() string
	Equal(Value) (bool, error)
	LessThan(Value) (bool, error)
}

type VBool bool

func (v VBool) Type() ValueType {
	return VTBool
}

func (v VBool) String() string {
	return fmt.Sprintf("%t", bool(v))
}

func (v VBool) Equal(other Value) (bool, error) {
	x, ok := other.(VBool)
	if !ok {
		return false, fmt.Errorf("expected bool, but got %s", other.Type())
	}
	return bool(x) == bool(v), nil
}

func (v VBool) LessThan(other Value) (bool, error) {
	return false, fmt.Errorf("unable to compare bool")
}

type VNumber float64

func (v VNumber) Type() ValueType {
	return VTNumber
}

func (v VNumber) String() string {
	return fmt.Sprintf("%g", float64(v))
}

func (v VNumber) Equal(other Value) (bool, error) {
	x, ok := other.(VNumber)
	if !ok {
		return false, fmt.Errorf("expected number, but got %s", other.Type())
	}
	return float64(x) == float64(v), nil
}

func (v VNumber) LessThan(other Value) (bool, error) {
	x, ok := other.(VNumber)
	if !ok {
		return false, fmt.Errorf("expected number, but got %s", other.Type())
	}
	return float64(v) < float64(x), nil
}

type VString string

func (v VString) Type() ValueType {
	return VTString
}

func (v VString) String() string {
	return fmt.Sprintf("\"%s\"", string(v))
}

func (v VString) Equal(other Value) (bool, error) {
	x, ok := other.(VString)
	if !ok {
		return false, fmt.Errorf("expected string, but got %s", other.Type())
	}
	return string(x) == string(v), nil
}

func (v VString) LessThan(other Value) (bool, error) {
	x, ok := other.(VString)
	if !ok {
		return false, fmt.Errorf("expected string, but got %s", other.Type())
	}
	return string(v) < string(x), nil
}

type VUserFun struct {
	Args []string
	Body []ast.Stmt
}

func (v *VUserFun) Type() ValueType {
	return VTUserFun
}

func (v *VUserFun) String() string {
	return "fun"
}

func (v *VUserFun) Equal(other Value) (bool, error) {
	return Value(v) == other, nil
}

func (v *VUserFun) LessThan(other Value) (bool, error) {
	return false, fmt.Errorf("unable to compare functions")
}

type VBuiltinFun func(*State, []Value) (Value, error)

func (v VBuiltinFun) Type() ValueType {
	return VTBuiltinFun
}

func (v VBuiltinFun) String() string {
	return "fun"
}

func (v VBuiltinFun) Equal(other Value) (bool, error) {
	return Value(v) == other, nil
}

func (v VBuiltinFun) LessThan(other Value) (bool, error) {
	return false, fmt.Errorf("unable to compare functions")
}

type VList struct {
	Elements []Value
}

func (v *VList) Type() ValueType {
	return VTList
}

func (v *VList) String() string {
	var elements []string
	for _, elem := range v.Elements {
		elements = append(elements, elem.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (v *VList) Equal(other Value) (bool, error) {
	x, ok := other.(*VList)
	if !ok {
		return false, fmt.Errorf("expected list, but got %s", other.Type())
	}
	if len(v.Elements) != len(x.Elements) {
		return false, nil
	}
	for i, elem := range v.Elements {
		eq, err := elem.Equal(x.Elements[i])
		if err != nil {
			return false, err
		}
		if !eq {
			return false, nil
		}
	}
	return true, nil
}

func (v *VList) LessThan(other Value) (bool, error) {
	return false, fmt.Errorf("unable to compare lists")
}
