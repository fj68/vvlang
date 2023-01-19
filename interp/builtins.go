package interp

import (
	"fmt"
	"strings"
)

var DefaultBuiltins = map[string]Value{
	"print":    VBuiltinFun(builtinPrint),
	"get_type": VBuiltinFun(builtinGetType),
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
		return nil, fmt.Errorf("too many / less arguments for typeof()")
	}
	return VString(args[0].Type().String()), nil
}
