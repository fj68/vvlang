package std

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
)

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

func Pop(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for pop()")
	}
	list, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("argument for pop() is expected list, but got %s", args[0].Type())
	}
	if len(list.Elements) == 0 {
		return nil, fmt.Errorf("pop() called on empty list")
	}
	last := list.Elements[len(list.Elements)-1]
	list.Elements = list.Elements[:len(list.Elements)-1]
	return last, nil
}

func Shift(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for shift()")
	}
	list, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("argument for shift() is expected list, but got %s", args[0].Type())
	}
	if len(list.Elements) == 0 {
		return nil, fmt.Errorf("shift() called on empty list")
	}
	first := list.Elements[0]
	list.Elements = list.Elements[1:]
	return first, nil
}

func Unshift(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("too many / less arguments for unshift()")
	}
	list, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("argument for unshift() is expected list, but got %s", args[0].Type())
	}
	list.Elements = append([]interp.Value{args[1]}, list.Elements...)
	return nil, nil
}

func Replace(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("too many / less arguments for replace()")
	}
	source, ok1 := args[0].(*interp.VList)
	oldSeq, ok2 := args[1].(*interp.VList)
	newSeq, ok3 := args[2].(*interp.VList)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("replace() expects 3 lists as arguments")
	}

	if len(oldSeq.Elements) == 0 {
		// If oldSeq is empty, we return a copy of the source (or we could insert between every element, but standard replace usually does copy or error)
		res := make([]interp.Value, len(source.Elements))
		copy(res, source.Elements)
		return &interp.VList{Elements: res}, nil
	}

	var result []interp.Value
	i := 0
	for i < len(source.Elements) {
		match := true
		if i+len(oldSeq.Elements) <= len(source.Elements) {
			for j := 0; j < len(oldSeq.Elements); j++ {
				eq, err := source.Elements[i+j].Equal(oldSeq.Elements[j])
				if err != nil || !eq {
					match = false
					break
				}
			}
		} else {
			match = false
		}

		if match {
			result = append(result, newSeq.Elements...)
			i += len(oldSeq.Elements)
		} else {
			result = append(result, source.Elements[i])
			i++
		}
	}

	return &interp.VList{Elements: result}, nil
}
