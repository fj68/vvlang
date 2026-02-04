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
	s := interp.NewState()
	s.RegisterGlobals(interp.DefaultBuiltins)
	if err := s.Eval([]rune(string(text))); err != nil {
		fmt.Println(err)
		return
	}
}
