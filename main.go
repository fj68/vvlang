package main

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/interp"
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
	s.RegisterGlobals(interp.DefaultBuiltins)
	s.NewState = func(sourcePath string) *interp.State {
		ns := interp.NewState(sourcePath)
		ns.RegisterGlobals(interp.DefaultBuiltins)
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
