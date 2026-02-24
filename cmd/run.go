package cmd

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/interp"
	"github.com/fj68/vvlang/interp/builtins"
	"github.com/fj68/vvlang/lib"
)

func Run() {
	if len(os.Args) < 2 {
		fmt.Println("usage: vv [path]")
		return
	}
	if len(os.Args) > 3 {
		fmt.Println("usage: vv run [path]")
		return
	}

	var path string
	if os.Args[1] == "run" {
		path = os.Args[2]
	} else {
		path = os.Args[1]
	}

	if err := run(path); err != nil {
		fmt.Println(err)
	}
}

func run(path string) error {
	text, err := os.ReadFile(path)
	if err != nil {
		return err
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

	s.RegisterBuiltinModules(moduleBuiltins)

	if err := s.EnsureSystemLibrary("std", lib.Std); err != nil {
		return err
	}
	if err := s.Eval([]rune(string(text))); err != nil {
		return err
	}

	return nil
}
