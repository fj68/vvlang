package std

import (
	"strings"

	"github.com/fj68/vvlang/interp"
)

func Print(s *interp.State, args []interp.Value) (interp.Value, error) {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString(arg.String())
	}
	println(b.String())
	return nil, nil
}
