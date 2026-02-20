package builtins

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
)

func StringLength(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for length()")
	}
	v, ok := args[0].(interp.VString)
	if !ok {
		return nil, fmt.Errorf("argument for length() is expected string, but got %s", args[0].Type())
	}
	return interp.VNumber(len([]rune(v))), nil
}
