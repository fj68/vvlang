package interp

import (
	"fmt"

	"github.com/fj68/new-lang/ast"
)

type ValueType int

const (
	VTBool ValueType = iota
	VTNumber
	VTString
	VTUserFun
	VTBuiltinFun
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
	}
	return "unknown"
}

type Value interface {
	Type() ValueType
	String() string
}

type VBool bool

func (v VBool) Type() ValueType {
	return VTBool
}

func (v VBool) String() string {
	return fmt.Sprintf("%t", bool(v))
}

type VNumber float64

func (v VNumber) Type() ValueType {
	return VTNumber
}

func (v VNumber) String() string {
	return fmt.Sprintf("%g", float64(v))
}

type VString string

func (v VString) Type() ValueType {
	return VTString
}

func (v VString) String() string {
	return fmt.Sprintf("\"%s\"", string(v))
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

type VBuiltinFun func(*State, []Value) (Value, error)

func (v VBuiltinFun) Type() ValueType {
	return VTBuiltinFun
}

func (v VBuiltinFun) String() string {
	return "fun"
}
