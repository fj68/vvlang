package std

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
)

func FloatToString(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for to_string()")
	}
	v, ok := args[0].(interp.VFloat)
	if !ok {
		return nil, fmt.Errorf("argument for to_string() is expected float, but got %s", args[0].Type())
	}
	return interp.StringToValue(fmt.Sprintf("%g", float64(v))), nil
}

func FloatToInt(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for to_int()")
	}
	v, ok := args[0].(interp.VFloat)
	if !ok {
		return nil, fmt.Errorf("argument for to_int() is expected float, but got %s", args[0].Type())
	}
	return interp.VInt(int64(v)), nil
}
