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
		"pi": interp.VFloat(math.Pi),
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
	},
	"std/int.vv": {
		"to_string": interp.VBuiltinFun(std.IntToString),
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
