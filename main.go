package main

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/interp"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Println("main [path]")
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
