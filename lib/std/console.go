package std

import (
	"fmt"
	"strings"

	"github.com/fj68/vvlang/interp"
)

func Print(s *interp.State, args []interp.Value) (interp.Value, error) {
	var strs []string
	for _, arg := range args {
		strs = append(strs, arg.Str())
	}
	fmt.Println(strings.Join(strs, " "))
	return interp.VNull{}, nil
}
