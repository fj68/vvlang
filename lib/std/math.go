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
	v, ok := args[0].(interp.VNumber)
	if !ok {
		return nil, fmt.Errorf("argument for ceil() is expected number, but got %s", args[0].Type())
	}
	return interp.VNumber(math.Ceil(float64(v))), nil
}

func Floor(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for floor()")
	}
	v, ok := args[0].(interp.VNumber)
	if !ok {
		return nil, fmt.Errorf("argument for floor() is expected number, but got %s", args[0].Type())
	}
	return interp.VNumber(math.Floor(float64(v))), nil
}
