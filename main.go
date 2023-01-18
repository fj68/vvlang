package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fj68/vvlang/interp"
)

func Print(s *interp.State, args []interp.Value) (interp.Value, error) {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString(arg.String())
	}
	fmt.Println(b.String())
	return nil, nil
}

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
	s.RegisterBuiltin("print", interp.VBuiltinFun(Print))
	value, err := s.Eval([]rune(string(text)))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(value)
}
