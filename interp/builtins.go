package interp

import (
	"fmt"
	"strings"
)

var DefaultBuiltins = map[string]Value{
	"print":    VBuiltinFun(builtinPrint),
	"get_type": VBuiltinFun(builtinGetType),
	"bool":     VBuiltinFun(builtinBool),
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
