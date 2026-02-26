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
		"help":    interp.VBuiltinFun(std.Help),
		"phys_eq": interp.VBuiltinFun(std.PhysEq),
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
		"to_bytes": interp.VBuiltinFun(std.CharToBytes),
	},
	"std/list.vv": {
		"length": interp.VBuiltinFun(std.ListLength),
		"push":   interp.VBuiltinFun(std.Push),
	},
	"std/float.vv": {
		"to_string": interp.VBuiltinFun(std.FloatToString),
	},
	"std/int.vv": {
		"to_string": interp.VBuiltinFun(std.IntToString),
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
