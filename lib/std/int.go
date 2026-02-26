package std

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
)

func IntToString(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for to_string()")
	}
	v, ok := args[0].(interp.VNumber)
	if !ok {
		return nil, fmt.Errorf("argument for to_string() is expected number, but got %s", args[0].Type())
	}
	return interp.StringToValue(fmt.Sprintf("%d", int64(v))), nil
}
