package interp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var DefaultBuiltins = map[string]Value{
	"print":    VBuiltinFun(builtinPrint),
	"get_type": VBuiltinFun(builtinGetType),
	"bool":     VBuiltinFun(builtinBool),
	"number":   VBuiltinFun(builtinNumber),
	"ceil":     VBuiltinFun(builtinCeil),
	"floor":    VBuiltinFun(builtinFloor),
	"string":   VBuiltinFun(builtinString),
}

func builtinPrint(s *State, args []Value) (Value, error) {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString(arg.String())
	}
	fmt.Println(b.String())
	return nil, nil
}

func builtinGetType(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for get_type()")
	}
	return VString(args[0].Type().String()), nil
}

func builtinBool(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for bool()")
	}
	switch v := args[0].(type) {
	case VBool:
		return v, nil
	case VNumber:
		return VBool(v != 0), nil
	case VString:
		return VBool(v == "true"), nil
	case *VUserFun:
		return VBool(v != nil), nil
	case VBuiltinFun:
		return VBool(v != nil), nil
	}
	return nil, fmt.Errorf("unknown value type: %s", args[0].Type().String())
}

func builtinNumber(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for number()")
	}
	switch v := args[0].(type) {
	case VBool:
		if v {
			return VNumber(1), nil
		}
		return VNumber(0), nil
	case VNumber:
		return v, nil
	case VString:
		n, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return nil, err
		}
		return VNumber(n), nil
	case *VUserFun:
		return nil, fmt.Errorf("unable to convert function to number")
	case VBuiltinFun:
		return nil, fmt.Errorf("unable to convert function to number")
	}
	return nil, fmt.Errorf("unknown value type: %s", args[0].Type().String())
}

func builtinCeil(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for number()")
	}
	v, ok := args[0].(VNumber)
	if !ok {
		return nil, fmt.Errorf("argument for ceil() is expected number, but got %s", v.Type())
	}
	return VNumber(math.Ceil(float64(v))), nil
}

func builtinFloor(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for number()")
	}
	v, ok := args[0].(VNumber)
	if !ok {
		return nil, fmt.Errorf("argument for floor() is expected number, but got %s", v.Type())
	}
	return VNumber(math.Floor(float64(v))), nil
}

func builtinString(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for number()")
	}
	switch v := args[0].(type) {
	case VBool:
		return VString(fmt.Sprintf("%t", v)), nil
	case VNumber:
		return VString(fmt.Sprintf("%g", v)), nil
	case VString:
		return v, nil
	case *VUserFun:
		return VString(v.Type().String()), nil
	case VBuiltinFun:
		return VString(v.Type().String()), nil
	}
	return nil, fmt.Errorf("unknown value type: %s", args[0].Type().String())
}
