package std

import (
	"fmt"
	"math"

	"github.com/fj68/vvlang/interp"
)

func Ceil(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for ceil()")
	}
	v, ok := args[0].(interp.VFloat)
	if !ok {
		return nil, fmt.Errorf("argument for ceil() is expected float, but got %s", args[0].Type())
	}
	return interp.VFloat(math.Ceil(float64(v))), nil
}

func Floor(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for floor()")
	}
	v, ok := args[0].(interp.VFloat)
	if !ok {
		return nil, fmt.Errorf("argument for floor() is expected float, but got %s", args[0].Type())
	}
	return interp.VFloat(math.Floor(float64(v))), nil
}
