package std

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fj68/vvlang/interp"
)

var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomSeed(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for seed()")
	}
	v, ok := args[0].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("argument for seed() is expected int, but got %s", args[0].Type())
	}
	randomSource.Seed(int64(v))
	return interp.NoneValue, nil
}

func RandomFloat(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("too many arguments for float()")
	}
	return interp.VFloat(randomSource.Float64()), nil
}

func RandomInt(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments for int(min, max), but got %d", len(args))
	}
	min, ok1 := args[0].(interp.VInt)
	max, ok2 := args[1].(interp.VInt)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("arguments for int(min, max) are expected int, but got %s and %s", args[0].Type(), args[1].Type())
	}
	if max <= min {
		return nil, fmt.Errorf("max must be greater than min in int(min, max)")
	}
	n := int64(max - min)
	return interp.VInt(int64(min) + randomSource.Int63n(n)), nil
}
