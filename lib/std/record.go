package std

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
)

func Get(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("too many / less arguments for get()")
	}
	record, ok := args[0].(*interp.VRecord)
	if !ok {
		return nil, fmt.Errorf("argument for get() is expected record, but got %s", args[0].Type())
	}
	key, ok := args[1].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("argument for get() is expected string, but got %s", args[1].Type())
	}
	return record.Fields[key.Str()], nil
}

func Set(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("too many / less arguments for set()")
	}
	record, ok := args[0].(*interp.VRecord)
	if !ok {
		return nil, fmt.Errorf("argument for set() is expected record, but got %s", args[0].Type())
	}
	key, ok := args[1].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("argument for set() is expected string, but got %s", args[1].Type())
	}
	record.Fields[key.Str()] = args[2]
	return nil, nil
}

func Keys(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for keys()")
	}
	record, ok := args[0].(*interp.VRecord)
	if !ok {
		return nil, fmt.Errorf("argument for keys() is expected record, but got %s", args[0].Type())
	}
	keys := make([]interp.Value, 0, len(record.Fields))
	for k := range record.Fields {
		keys = append(keys, interp.StringToValue(k))
	}
	return &interp.VList{Elements: keys}, nil
}

func Values(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for values()")
	}
	record, ok := args[0].(*interp.VRecord)
	if !ok {
		return nil, fmt.Errorf("argument for values() is expected record, but got %s", args[0].Type())
	}
	values := make([]interp.Value, 0, len(record.Fields))
	for _, v := range record.Fields {
		values = append(values, v)
	}
	return &interp.VList{Elements: values}, nil
}

func ToList(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for to_list()")
	}
	record, ok := args[0].(*interp.VRecord)
	if !ok {
		return nil, fmt.Errorf("argument for to_list() is expected record, but got %s", args[0].Type())
	}
	list := make([]interp.Value, 0, len(record.Fields))
	for k, v := range record.Fields {
		list = append(list, &interp.VList{Elements: []interp.Value{interp.StringToValue(k), v}})
	}
	return &interp.VList{Elements: list}, nil
}
