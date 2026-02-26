package std

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
)

func ListLength(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for length()")
	}
	v, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("argument for length() is expected list, but got %s", args[0].Type())
	}
	return interp.VInt(len(v.Elements)), nil
}

func Push(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("too many / less arguments for push()")
	}
	list, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("argument for push() is expected list, but got %s", args[0].Type())
	}
	list.Elements = append(list.Elements, args[1])
	return nil, nil
}
