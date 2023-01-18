package interp

import "github.com/fj68/new-lang/ast"

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
}

type VBool bool

func (v VBool) Type() ValueType {
	return VTBool
}

type VNumber float64

func (v VNumber) Type() ValueType {
	return VTNumber
}

type VString string

func (v VString) Type() ValueType {
	return VTString
}

type VUserFun []ast.Stmt

func (v VUserFun) Type() ValueType {
	return VTUserFun
}

type VBuiltinFun func(*State) error

func (v VBuiltinFun) Type() ValueType {
	return VTBuiltinFun
}
