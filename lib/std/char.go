package std

import (
	"unicode"

	"github.com/fj68/vvlang/interp"
)

func CharToUpper(s *interp.State, args []interp.Value) (interp.Value, error) {
	v := args[0].(interp.VChar)
	return interp.VChar(unicode.ToUpper(rune(v))), nil
}

func CharToLower(s *interp.State, args []interp.Value) (interp.Value, error) {
	v := args[0].(interp.VChar)
	return interp.VChar(unicode.ToLower(rune(v))), nil
}

func CharIsDigit(s *interp.State, args []interp.Value) (interp.Value, error) {
	v := args[0].(interp.VChar)
	return interp.VBool(unicode.IsDigit(rune(v))), nil
}

func CharIsSpace(s *interp.State, args []interp.Value) (interp.Value, error) {
	v := args[0].(interp.VChar)
	return interp.VBool(unicode.IsSpace(rune(v))), nil
}
