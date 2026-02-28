package lib

import (
	"embed"
	"math"

	"github.com/fj68/vvlang/interp"
	"github.com/fj68/vvlang/lib/std"
)

//go:embed std
var embedded embed.FS

var natives = map[string]map[string]interp.Value{
	"std/sys.vv": {
		"help":                    interp.VBuiltinFun(std.Help),
		"phys_eq":                 interp.VBuiltinFun(std.PhysEq),
		"set_max_recursion_depth": interp.VBuiltinFun(std.SetMaxRecursionDepth),
		"get_max_recursion_depth": interp.VBuiltinFun(std.GetMaxRecursionDepth),
		"type":                    interp.VBuiltinFun(std.Type),
		"str":                     interp.VBuiltinFun(std.Str),
	},
	"std/console.vv": {
		"print": interp.VBuiltinFun(std.Print),
	},
	"std/math.vv": {
		"pi":    interp.VFloat(math.Pi),
		"e":     interp.VFloat(math.E),
		"sin":   interp.VBuiltinFun(std.MathSin),
		"asin":  interp.VBuiltinFun(std.MathAsin),
		"cos":   interp.VBuiltinFun(std.MathCos),
		"acos":  interp.VBuiltinFun(std.MathAcos),
		"tan":   interp.VBuiltinFun(std.MathTan),
		"atan":  interp.VBuiltinFun(std.MathAtan),
		"atan2": interp.VBuiltinFun(std.MathAtan2),
		"pow":   interp.VBuiltinFun(std.MathPow),
		"sqrt":  interp.VBuiltinFun(std.MathSqrt),
	},
	"std/char.vv": {
		"to_upper": interp.VBuiltinFun(std.CharToUpper),
		"to_lower": interp.VBuiltinFun(std.CharToLower),
		"is_digit": interp.VBuiltinFun(std.CharIsDigit),
		"is_space": interp.VBuiltinFun(std.CharIsSpace),
		"to_bytes": interp.VBuiltinFun(std.CharToBytes),
	},
	"std/list.vv": {
		"push":    interp.VBuiltinFun(std.Push),
		"pop":     interp.VBuiltinFun(std.Pop),
		"shift":   interp.VBuiltinFun(std.Shift),
		"unshift": interp.VBuiltinFun(std.Unshift),
		"replace": interp.VBuiltinFun(std.Replace),
	},
	"std/float.vv": {
		"to_string": interp.VBuiltinFun(std.FloatToString),
		"floor":     interp.VBuiltinFun(std.Floor),
		"ceil":      interp.VBuiltinFun(std.Ceil),
		"from_int":  interp.VBuiltinFun(std.FloatFromInt),
		"round":     interp.VBuiltinFun(std.Round),
		"abs":       interp.VBuiltinFun(std.FloatAbs),
		"min":       interp.VBuiltinFun(std.FloatMin),
		"max":       interp.VBuiltinFun(std.FloatMax),
	},
	"std/int.vv": {
		"to_string": interp.VBuiltinFun(std.IntToString),
		"abs":       interp.VBuiltinFun(std.IntAbs),
		"min":       interp.VBuiltinFun(std.IntMin),
		"max":       interp.VBuiltinFun(std.IntMax),
	},
	"std/bool.vv": {
		"to_string": interp.VBuiltinFun(std.BoolToString),
	},
	"std/random.vv": {
		"seed":  interp.VBuiltinFun(std.RandomSeed),
		"float": interp.VBuiltinFun(std.RandomFloat),
		"int":   interp.VBuiltinFun(std.RandomInt),
	},
}

var Std = &Lib{
	Name:    "std",
	FS:      embedded,
	Natives: natives,
}
