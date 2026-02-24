package cmd

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/interp"
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

	s.RegisterBuiltinModules(lib.BuiltinModules)

	if err := s.EnsureSystemLibrary(lib.Name, lib.Std); err != nil {
		return err
	}
	if err := s.Eval([]rune(string(text))); err != nil {
		return err
	}

	return nil
}
