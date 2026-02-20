package main

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/interp"
	"github.com/fj68/vvlang/interp/builtins"
	"github.com/fj68/vvlang/mod"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: main [path]")
		return
	}
	path := os.Args[1]
	text, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	s := interp.NewState(path)

	var moduleBuiltins = map[string]map[string]interp.Value{
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

	s.NewState = func(sourcePath string) *interp.State {
		ns := interp.NewState(sourcePath)

		// Register module-specific built-ins
		for stdPath, funcs := range moduleBuiltins {
			if sourcePath == mod.GetPackagePath(stdPath) {
				ns.RegisterNatives(funcs)
				break
			}
		}

		ns.NewState = s.NewState
		return ns
	}
	if err := s.EnsureSystemLibrary("std", stdlib); err != nil {
		fmt.Println("EnsureSystemLibrary", err)
		return
	}
	if err := s.Eval([]rune(string(text))); err != nil {
		fmt.Println(err)
		return
	}
}
