package std

import (
	"fmt"
	"math"

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

func FloatFromInt(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for from_int()")
	}
	v, ok := args[0].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("argument for from_int() is expected int, but got %s", args[0].Type())
	}
	return interp.VFloat(float64(v)), nil
}

func Round(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("argument count mismatch: expected 1, got %d", len(args))
	}
	v, ok := args[0].(interp.VFloat)
	if !ok {
		return nil, fmt.Errorf("argument for round() is expected float, but got %s", args[0].Type())
	}
	return interp.VFloat(math.Round(float64(v))), nil
}

func FloatAbs(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("argument count mismatch: expected 1, got %d", len(args))
	}
	v, ok := args[0].(interp.VFloat)
	if !ok {
		return nil, fmt.Errorf("argument for abs() is expected float, but got %s", args[0].Type())
	}
	return interp.VFloat(math.Abs(float64(v))), nil
}

func FloatMin(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("argument count mismatch: expected 2, got %d", len(args))
	}
	a, ok1 := args[0].(interp.VFloat)
	b, ok2 := args[1].(interp.VFloat)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("arguments for min() are expected float")
	}
	return interp.VFloat(math.Min(float64(a), float64(b))), nil
}

func FloatMax(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("argument count mismatch: expected 2, got %d", len(args))
	}
	a, ok1 := args[0].(interp.VFloat)
	b, ok2 := args[1].(interp.VFloat)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("arguments for max() are expected float")
	}
	return interp.VFloat(math.Max(float64(a), float64(b))), nil
}
