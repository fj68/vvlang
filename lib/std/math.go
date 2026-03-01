package std

import (
	"fmt"
	"math"

	"github.com/fj68/vvlang/interp"
)

func mathFunc1(name string, f func(float64) float64) func(*interp.State, []interp.Value) (interp.Value, error) {
	return func(s *interp.State, args []interp.Value) (interp.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("argument count mismatch for %s(): expected 1, got %d", name, len(args))
		}
		v, ok := args[0].(interp.VFloat)
		if !ok {
			return nil, fmt.Errorf("argument for %s() is expected float, but got %s", name, args[0].Type())
		}
		return interp.VFloat(f(float64(v))), nil
	}
}

func mathFunc2(name string, f func(float64, float64) float64) func(*interp.State, []interp.Value) (interp.Value, error) {
	return func(s *interp.State, args []interp.Value) (interp.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("argument count mismatch for %s(): expected 2, got %d", name, len(args))
		}
		a, ok1 := args[0].(interp.VFloat)
		b, ok2 := args[1].(interp.VFloat)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("arguments for %s() are expected float", name)
		}
		return interp.VFloat(f(float64(a), float64(b))), nil
	}
}

var MathSin = mathFunc1("sin", math.Sin)
var MathAsin = mathFunc1("asin", math.Asin)
var MathCos = mathFunc1("cos", math.Cos)
var MathAcos = mathFunc1("acos", math.Acos)
var MathTan = mathFunc1("tan", math.Tan)
var MathAtan = mathFunc1("atan", math.Atan)
var MathAtan2 = mathFunc2("atan2", math.Atan2)
var MathPow = mathFunc2("pow", math.Pow)
var MathSqrt = mathFunc1("sqrt", math.Sqrt)
