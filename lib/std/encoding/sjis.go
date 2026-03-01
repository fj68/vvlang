package encoding

import (
	"fmt"

	"github.com/fj68/vvlang/interp"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func SJISEncode(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sjis.encode expects 1 argument")
	}
	c, ok := args[0].(interp.VChar)
	if !ok {
		return nil, fmt.Errorf("sjis.encode expects a character")
	}

	str := string(rune(c))
	encoder := japanese.ShiftJIS.NewEncoder()
	sjisStr, _, err := transform.String(encoder, str)
	if err != nil {
		return nil, err
	}

	bytes := []byte(sjisStr)
	elems := make([]interp.Value, len(bytes))
	for i, b := range bytes {
		elems[i] = interp.VInt(int64(b))
	}
	return &interp.VList{Elements: elems}, nil
}

func SJISDecode(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sjis.decode expects 1 argument")
	}
	list, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("sjis.decode expects a list of bytes")
	}

	bytes := make([]byte, len(list.Elements))
	for i, v := range list.Elements {
		b, ok := v.(interp.VInt)
		if !ok {
			return nil, fmt.Errorf("sjis.decode expects a list of integers")
		}
		bytes[i] = byte(b)
	}

	// Determine if it's a 1-byte or 2-byte SJIS character.

	// This is not quite right for getting consumed bytes.

	// Determine if it's a 1-byte or 2-byte SJIS character.
	b0 := bytes[0]
	size := 1
	if (b0 >= 0x81 && b0 <= 0x9F) || (b0 >= 0xE0 && b0 <= 0xFC) {
		size = 2
	}

	if len(bytes) < size {
		return interp.ErrorValue(interp.StringToValue("truncated SJIS sequence")), nil
	}

	d := japanese.ShiftJIS.NewDecoder()
	utf8Str, _, err := transform.String(d, string(bytes[:size]))
	if err != nil || len(utf8Str) == 0 {
		return interp.ErrorValue(interp.StringToValue("invalid SJIS sequence")), nil
	}

	runes := []rune(utf8Str)
	return interp.OkValue(&interp.VRecord{
		Fields: map[string]interp.Value{
			"char": interp.VChar(runes[0]),
			"size": interp.VInt(int64(size)),
		},
	}), nil
}
