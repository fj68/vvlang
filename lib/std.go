package lib

import (
	"embed"

	"github.com/fj68/vvlang/interp"
	"github.com/fj68/vvlang/interp/builtins"
)

//go:embed std
var Std embed.FS

var BuiltinModules = map[string]map[string]interp.Value{
	"std/sys.vv": {
		"help": interp.VBuiltinFun(builtins.Help),
	},
	"std/console.vv": {
		"print": interp.VBuiltinFun(builtins.Print),
	},
	"std/math.vv": {
		"floor": interp.VBuiltinFun(builtins.Floor),
		"ceil":  interp.VBuiltinFun(builtins.Ceil),
	},
	"std/string.vv": {
		"length": interp.VBuiltinFun(builtins.StringLength),
	},
	"std/list.vv": {
		"length": interp.VBuiltinFun(builtins.ListLength),
		"push":   interp.VBuiltinFun(builtins.Push),
	},
	"std/float.vv": {
		"to_string": interp.VBuiltinFun(builtins.FloatToString),
	},
	"std/int.vv": {
		"to_string": interp.VBuiltinFun(builtins.IntToString),
	},
	"std/bool.vv": {
		"to_string": interp.VBuiltinFun(builtins.BoolToString),
	},
}

