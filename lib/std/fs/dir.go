package fs

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/interp"
)

func DirRead(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("read expects 1 argument")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("read expects path to be a string")
	}

	path := pathVal.Str()
	entries, err := os.ReadDir(path)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}

	var names []interp.Value
	for _, entry := range entries {
		names = append(names, interp.StringToValue(entry.Name()))
	}

	return interp.OkValue(&interp.VList{Elements: names}), nil
}

func DirCreate(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("create expects 1 argument")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("create expects path to be a string")
	}

	path := pathVal.Str()
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(interp.NoneValue), nil
}

func DirRemove(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("remove expects 1 argument")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("remove expects path to be a string")
	}

	path := pathVal.Str()
	err := os.RemoveAll(path)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(interp.NoneValue), nil
}

func DirExists(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exists expects 1 argument")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("exists expects path to be a string")
	}

	path := pathVal.Str()
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return interp.OkValue(interp.VBool(false)), nil
		}
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	if !info.IsDir() {
		return interp.OkValue(interp.VBool(false)), nil
	}
	return interp.OkValue(interp.VBool(true)), nil
}
