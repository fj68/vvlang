package encoding

import (
	"fmt"
	"unicode/utf8"

	"github.com/fj68/vvlang/interp"
)

func UTF8Encode(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("utf8.encode expects 1 argument")
	}
	c, ok := args[0].(interp.VChar)
	if !ok {
		return nil, fmt.Errorf("utf8.encode expects a character")
	}

	buf := make([]byte, 4)
	n := utf8.EncodeRune(buf, rune(c))
	bytes := buf[:n]

	elems := make([]interp.Value, len(bytes))
	for i, b := range bytes {
		elems[i] = interp.VInt(int64(b))
	}
	return &interp.VList{Elements: elems}, nil
}

func UTF8Decode(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("utf8.decode expects 1 argument")
	}
	list, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("utf8.decode expects a list of bytes")
	}

	bytes := make([]byte, len(list.Elements))
	for i, v := range list.Elements {
		b, ok := v.(interp.VInt)
		if !ok {
			return nil, fmt.Errorf("utf8.decode expects a list of integers")
		}
		bytes[i] = byte(b)
	}

	if len(bytes) == 0 {
		return interp.ErrorValue(interp.StringToValue("unexpected end of input")), nil
	}

	r, size := utf8.DecodeRune(bytes)
	if r == utf8.RuneError && size == 0 {
		return interp.ErrorValue(interp.StringToValue("unexpected end of input")), nil
	}
	if r == utf8.RuneError && size == 1 {
		return interp.ErrorValue(interp.StringToValue("invalid UTF-8 sequence")), nil
	}

	return interp.OkValue(&interp.VRecord{
		Fields: map[string]interp.Value{
			"char": interp.VChar(r),
			"size": interp.VInt(int64(size)),
		},
	}), nil
}
