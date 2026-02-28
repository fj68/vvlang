package std

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
)

func IntToString(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for to_string()")
	}
	v, ok := args[0].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("argument for to_string() is expected int, but got %s", args[0].Type())
	}
	return interp.StringToValue(fmt.Sprintf("%d", int64(v))), nil
}

func IntAbs(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("argument count mismatch: expected 1, got %d", len(args))
	}
	v, ok := args[0].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("argument for abs() is expected int, but got %s", args[0].Type())
	}
	if v < 0 {
		return -v, nil
	}
	return v, nil
}

func IntMin(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("argument count mismatch: expected 2, got %d", len(args))
	}
	a, ok1 := args[0].(interp.VInt)
	b, ok2 := args[1].(interp.VInt)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("arguments for min() are expected int")
	}
	if a < b {
		return a, nil
	}
	return b, nil
}

func IntMax(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("argument count mismatch: expected 2, got %d", len(args))
	}
	a, ok1 := args[0].(interp.VInt)
	b, ok2 := args[1].(interp.VInt)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("arguments for max() are expected int")
	}
	if a > b {
		return a, nil
	}
	return b, nil
}
