package fs

import (
	"fmt"
	"io"
	"os"

	"github.com/fj68/vvlang/interp"
)

func FileOpen(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("open expects 2 arguments")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("open expects path to be a string (list of chars)")
	}
	isAppendVal, ok := args[1].(interp.VBool)
	if !ok {
		return nil, fmt.Errorf("open expects is_append to be a boolean")
	}

	path := pathVal.Str()
	flag := os.O_RDWR
	if isAppendVal {
		flag |= os.O_APPEND
	}
	f, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return nil, err
	}
	return &interp.VUserData{Value: f}, nil
}

func FileOpenRead(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("open_read expects 1 argument")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("open_read expects path to be a string")
	}

	path := pathVal.Str()
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(&interp.VUserData{Value: f}), nil
}

func FileOpenWrite(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("open_write expects 2 arguments")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("open_write expects path to be a string")
	}
	isAppendVal, ok := args[1].(interp.VBool)
	if !ok {
		return nil, fmt.Errorf("open_write expects is_append to be a boolean")
	}

	path := pathVal.Str()
	flag := os.O_WRONLY | os.O_TRUNC
	if isAppendVal {
		flag = os.O_WRONLY | os.O_APPEND
	}
	f, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return nil, err
	}
	return &interp.VUserData{Value: f}, nil
}

func FileCreate(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("create expects 1 argument")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("create expects path to be a string")
	}

	path := pathVal.Str()
	f, err := os.Create(path)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(&interp.VUserData{Value: f}), nil
}

func FileExists(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exists expects 1 argument")
	}
	pathVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("exists expects path to be a string")
	}

	path := pathVal.Str()
	_, err := os.Stat(path)
	if err == nil {
		return interp.OkValue(interp.VBool(true)), nil
	}
	if os.IsNotExist(err) {
		return interp.OkValue(interp.VBool(false)), nil
	}
	return interp.ErrorValue(interp.StringToValue(err.Error())), nil
}

func FileClose(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("close expects 1 argument")
	}
	fdVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("close expects fd to be a userdata")
	}
	f, ok := fdVal.Value.(*os.File)
	if !ok {
		return nil, fmt.Errorf("invalid file descriptor")
	}
	err := f.Close()
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(interp.NoneValue), nil
}

func FileRead(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("read expects 2 arguments")
	}
	fdVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("read expects fd to be a userdata")
	}
	f, ok := fdVal.Value.(*os.File)
	if !ok {
		return nil, fmt.Errorf("invalid file descriptor")
	}
	lenVal, ok := args[1].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("read expects length to be an int")
	}

	buf := make([]byte, lenVal)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}

	// Convert read bytes (up to n) to list of ints
	ints := make([]interp.Value, n)
	for i := 0; i < n; i++ {
		ints[i] = interp.VInt(int64(buf[i]))
	}

	return interp.OkValue(&interp.VList{Elements: ints}), nil
}

func FileWrite(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("write expects 2 arguments")
	}
	fdVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("write expects fd to be a userdata")
	}
	f, ok := fdVal.Value.(*os.File)
	if !ok {
		return nil, fmt.Errorf("invalid file descriptor")
	}
	contentVal, ok := args[1].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("write expects content to be a string (list of chars)")
	}

	// extract bytes from string or list of ints
	buf := make([]byte, len(contentVal.Elements))
	for i, v := range contentVal.Elements {
		if c, ok := v.(interp.VChar); ok {
			buf[i] = byte(c)
		} else if n, ok := v.(interp.VInt); ok {
			buf[i] = byte(n)
		} else {
			return nil, fmt.Errorf("write content must be list of chars or ints, got %s at index %d", v.Type(), i)
		}
	}

	n, err := f.Write(buf)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}

	return interp.OkValue(interp.VInt(n)), nil
}
