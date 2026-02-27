package interp

import (
	"fmt"
	"sort"
	"strings"
	"unsafe"

	"github.com/fj68/vvlang/ast"
)

type ValueType int

const (
	VTBool ValueType = iota
	VTInt
	VTFloat
	VTChar
	VTUserFun
	VTBuiltinFun
	VTList
	VTRecord
	VTTailCall
)

type ptrPair struct {
	a, b unsafe.Pointer
}

func (ty ValueType) String() string {
	switch ty {
	case VTBool:
		return "bool"
	case VTInt:
		return "int"
	case VTFloat:
		return "float"
	case VTChar:
		return "char"
	case VTUserFun:
		return "fun"
	case VTBuiltinFun:
		return "fun"
	case VTList:
		return "list"
	case VTRecord:
		return "record"
	case VTTailCall:
		return "tailcall"
	}
	return "unknown"
}

type Value interface {
	Type() ValueType
	String() string
	Str() string
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

func (v VBool) Str() string {
	return v.String()
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

type VInt int64

func (v VInt) Type() ValueType {
	return VTInt
}

func (v VInt) String() string {
	return fmt.Sprintf("%d", int64(v))
}

func (v VInt) Str() string {
	return v.String()
}

func (v VInt) Equal(other Value) (bool, error) {
	x, ok := other.(VInt)
	if !ok {
		return false, fmt.Errorf("expected int, but got %s", other.Type())
	}
	return int64(x) == int64(v), nil
}

func (v VInt) LessThan(other Value) (bool, error) {
	x, ok := other.(VInt)
	if !ok {
		return false, fmt.Errorf("expected int, but got %s", other.Type())
	}
	return int64(v) < int64(x), nil
}

type VFloat float64

func (v VFloat) Type() ValueType {
	return VTFloat
}

func (v VFloat) String() string {
	return fmt.Sprintf("%g", float64(v))
}

func (v VFloat) Str() string {
	return v.String()
}

func (v VFloat) Equal(other Value) (bool, error) {
	x, ok := other.(VFloat)
	if !ok {
		return false, fmt.Errorf("expected float, but got %s", other.Type())
	}
	return float64(x) == float64(v), nil
}

func (v VFloat) LessThan(other Value) (bool, error) {
	x, ok := other.(VFloat)
	if !ok {
		return false, fmt.Errorf("expected float, but got %s", other.Type())
	}
	return float64(v) < float64(x), nil
}

type VChar rune

func (v VChar) Type() ValueType {
	return VTChar
}

func (v VChar) String() string {
	return fmt.Sprintf("'%c'", rune(v))
}

func (v VChar) Str() string {
	return string(rune(v))
}

func (v VChar) Equal(other Value) (bool, error) {
	x, ok := other.(VChar)
	if !ok {
		return false, fmt.Errorf("expected char, but got %s", other.Type())
	}
	return rune(x) == rune(v), nil
}

func (v VChar) LessThan(other Value) (bool, error) {
	x, ok := other.(VChar)
	if !ok {
		return false, fmt.Errorf("expected char, but got %s", other.Type())
	}
	return rune(v) < rune(x), nil
}

type VUserFun struct {
	// CapturedEnv is the environment at the time the function was defined.
	// This is essential for:
	// 1. Closures: allowing functions to access variables from their outer scopes.
	// 2. Modules: ensuring functions in imported modules can access other symbols
	//    within that same module's top-level scope, even when called from elsewhere.
	CapturedEnv *Env
	Args        []string
	Body        []ast.Stmt
}

func (v *VUserFun) Type() ValueType {
	return VTUserFun
}

func (v *VUserFun) String() string {
	return "<function>"
}

func (v *VUserFun) Str() string {
	return v.String()
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
	return "<function>"
}

func (v VBuiltinFun) Str() string {
	return v.String()
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
	if len(v.Elements) > 0 {
		allChar := true
		for _, e := range v.Elements {
			if _, ok := e.(VChar); !ok {
				allChar = false
				break
			}
		}
		if allChar {
			var b strings.Builder
			b.WriteRune('"')
			for _, e := range v.Elements {
				// We don't use e.Str() here because we want to escape special chars if needed?
				// For now let's just use raw.
				b.WriteString(e.(VChar).Str())
			}
			b.WriteRune('"')
			return b.String()
		}
	} else {
		// Empty list could be an empty string, but let's stick to [] for now
		// unless we have a reason to prefer "".
		// Actually, if we want it to be compatible with string expectations:
		// return "[]"
	}

	var elements []string
	for _, elem := range v.Elements {
		elements = append(elements, elem.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (v *VList) Str() string {
	if len(v.Elements) == 0 {
		return ""
	}
	allChar := true
	for _, e := range v.Elements {
		if _, ok := e.(VChar); !ok {
			allChar = false
			break
		}
	}
	if allChar {
		var b strings.Builder
		for _, e := range v.Elements {
			b.WriteString(e.(VChar).Str())
		}
		return b.String()
	}
	return v.String()
}

func (v *VList) Equal(other Value) (bool, error) {
	return v.equal(other, make(map[ptrPair]bool))
}

func (v *VList) equal(other Value, seen map[ptrPair]bool) (bool, error) {
	x, ok := other.(*VList)
	if !ok {
		return false, nil
	}
	if v == x {
		return true, nil
	}
	if len(v.Elements) != len(x.Elements) {
		return false, nil
	}

	pair := ptrPair{unsafe.Pointer(v), unsafe.Pointer(x)}
	if seen[pair] {
		return true, nil
	}
	seen[pair] = true

	for i, elem := range v.Elements {
		var eq bool
		var err error
		if el, ok := elem.(interface {
			equal(Value, map[ptrPair]bool) (bool, error)
		}); ok {
			eq, err = el.equal(x.Elements[i], seen)
		} else {
			eq, err = elem.Equal(x.Elements[i])
		}

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

type VRecord struct {
	Fields map[string]Value
}

func (v *VRecord) Type() ValueType {
	return VTRecord
}

func (v *VRecord) String() string {
	var keys []string
	for k := range v.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s = %s", k, v.Fields[k].Str()))
	}
	return fmt.Sprintf("{ %s }", strings.Join(parts, ", "))
}

func (v *VRecord) Str() string {
	return v.String()
}

func (v *VRecord) Equal(other Value) (bool, error) {
	return v.equal(other, make(map[ptrPair]bool))
}

func (v *VRecord) equal(other Value, seen map[ptrPair]bool) (bool, error) {
	var o *VRecord
	switch tv := other.(type) {
	case *VRecord:
		o = tv
	case *VModule:
		o = tv.VRecord
	default:
		return false, nil
	}
	if v == o {
		return true, nil
	}
	if len(o.Fields) != len(v.Fields) {
		return false, nil
	}

	pair := ptrPair{unsafe.Pointer(v), unsafe.Pointer(o)}
	if seen[pair] {
		return true, nil
	}
	seen[pair] = true

	for k, val := range v.Fields {
		ov, ok := o.Fields[k]
		if !ok {
			return false, nil
		}

		var eq bool
		var err error
		if el, ok := val.(interface {
			equal(Value, map[ptrPair]bool) (bool, error)
		}); ok {
			eq, err = el.equal(ov, seen)
		} else {
			eq, err = val.Equal(ov)
		}

		if err != nil {
			return false, err
		}
		if !eq {
			return false, nil
		}
	}
	return true, nil
}

func (v *VRecord) LessThan(other Value) (bool, error) {
	return false, fmt.Errorf("unable to compare records")
}

// VModule wraps a VRecord with docstring metadata. It represents an imported module.
type VModule struct {
	*VRecord
	Docstring       map[string]string
	FieldDocstrings map[string]map[string]string
}

func (v *VModule) Type() ValueType {
	return VTRecord // still reported as record for type() calls
}

func (v *VModule) Equal(other Value) (bool, error) {
	return Value(v) == other, nil
}

func (v *VModule) LessThan(other Value) (bool, error) {
	return false, fmt.Errorf("unable to compare modules")
}

func (v *VModule) Str() string {
	return v.VRecord.Str()
}

type VTailCall struct {
	Fun  *VUserFun
	Args []Value
}

func (v *VTailCall) Type() ValueType {
	return VTTailCall
}

func (v *VTailCall) String() string {
	return "<tailcall>"
}

func (v *VTailCall) Str() string {
	return v.String()
}

func (v *VTailCall) Equal(other Value) (bool, error) {
	return false, fmt.Errorf("cannot compare tail calls")
}

func (v *VTailCall) LessThan(other Value) (bool, error) {
	return false, fmt.Errorf("unable to compare tail calls")
}

func StringToValue(s string) Value {
	runes := []rune(s)
	elems := make([]Value, len(runes))
	for i, r := range runes {
		elems[i] = VChar(r)
	}
	return &VList{Elements: elems}
}

// NoneValue is the standard representation of `null` internally, `{ type = "none" }`.
var NoneValue Value = &VRecord{
	Fields: map[string]Value{
		"type": StringToValue("none"),
	},
}
