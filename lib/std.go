package lib

import (
	"embed"

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
	},
	"std/console.vv": {
		"print": interp.VBuiltinFun(std.Print),
	},
	"std/math.vv": {
		"floor": interp.VBuiltinFun(std.Floor),
		"ceil":  interp.VBuiltinFun(std.Ceil),
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
		"to_int":    interp.VBuiltinFun(std.FloatToInt),
	},
	"std/int.vv": {
		"to_string": interp.VBuiltinFun(std.IntToString),
		"to_float":  interp.VBuiltinFun(std.IntToFloat),
	},
	"std/bool.vv": {
		"to_string": interp.VBuiltinFun(std.BoolToString),
	},
}

var Std = &Lib{
	Name:    "std",
	FS:      embedded,
	Natives: natives,
}
