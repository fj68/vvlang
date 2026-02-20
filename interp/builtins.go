package interp

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

var DefaultBuiltins = map[string]Value{
	"not":    VBuiltinFun(builtinNot),
	"print":  VBuiltinFun(builtinPrint),
	"type":   VBuiltinFun(builtinType),
	"bool":   VBuiltinFun(builtinBool),
	"number": VBuiltinFun(builtinNumber),
	"ceil":   VBuiltinFun(builtinCeil),
	"floor":  VBuiltinFun(builtinFloor),
	"string": VBuiltinFun(builtinString),
	"len":    VBuiltinFun(builtinLen),
	"push":   VBuiltinFun(builtinPush),
	"help":   VBuiltinFun(builtinHelp),
}

func builtinNot(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for not()")
	}
	v, ok := args[0].(VBool)
	if !ok {
		return nil, fmt.Errorf("argument for not() is expected bool, but got %s", v.Type())
	}
	return VBool(!bool(v)), nil
}

func builtinPrint(s *State, args []Value) (Value, error) {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString(arg.String())
	}
	fmt.Println(b.String())
	return nil, nil
}

func builtinType(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for get_type()")
	}
	return VString(args[0].Type().String()), nil
}

func builtinBool(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for bool()")
	}
	switch v := args[0].(type) {
	case VBool:
		return v, nil
	case VNumber:
		return VBool(v != 0), nil
	case VString:
		return VBool(v == "true"), nil
	case *VUserFun:
		return VBool(v != nil), nil
	case VBuiltinFun:
		return VBool(v != nil), nil
	case *VList:
		return VBool(len(v.Elements) != 0), nil
	case *VRecord:
		return VBool(len(v.Fields) != 0), nil
	}
	return nil, fmt.Errorf("unknown value type: %s", args[0].Type().String())
}

func builtinNumber(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for number()")
	}
	switch v := args[0].(type) {
	case VBool:
		if v {
			return VNumber(1), nil
		}
		return VNumber(0), nil
	case VNumber:
		return v, nil
	case VString:
		n, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return nil, err
		}
		return VNumber(n), nil
	case *VUserFun:
		return nil, fmt.Errorf("unable to convert function to number")
	case VBuiltinFun:
		return nil, fmt.Errorf("unable to convert function to number")
	case *VList:
		return nil, fmt.Errorf("unable to convert list to number")
	case *VRecord:
		return nil, fmt.Errorf("unable to convert record to number")
	}
	return nil, fmt.Errorf("unknown value type: %s", args[0].Type().String())
}

func builtinCeil(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for ceil()")
	}
	v, ok := args[0].(VNumber)
	if !ok {
		return nil, fmt.Errorf("argument for ceil() is expected number, but got %s", v.Type())
	}
	return VNumber(math.Ceil(float64(v))), nil
}

func builtinFloor(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for floor()")
	}
	v, ok := args[0].(VNumber)
	if !ok {
		return nil, fmt.Errorf("argument for floor() is expected number, but got %s", v.Type())
	}
	return VNumber(math.Floor(float64(v))), nil
}

func builtinString(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for string()")
	}
	switch v := args[0].(type) {
	case VBool:
		return VString(fmt.Sprintf("%t", v)), nil
	case VNumber:
		return VString(fmt.Sprintf("%g", v)), nil
	case VString:
		return v, nil
	case *VUserFun:
		return VString(v.Type().String()), nil
	case VBuiltinFun:
		return VString(v.Type().String()), nil
	case *VList:
		return VString(v.String()), nil
	case *VRecord:
		return VString(v.String()), nil
	}
	return nil, fmt.Errorf("unknown value type: %s", args[0].Type().String())
}

func builtinLen(s *State, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("too many / less arguments for len()")
	}
	switch v := args[0].(type) {
	case VBool:
		return nil, fmt.Errorf("argument for len() is expected string or array, but got bool")
	case VNumber:
		return nil, fmt.Errorf("argument for len() is expected string or array, but got number")
	case VString:
		return VNumber(len([]rune(v))), nil
	case *VList:
		return VNumber(len(v.Elements)), nil
	case *VRecord:
		return VNumber(len(v.Fields)), nil
	case *VUserFun:
		return nil, fmt.Errorf("argument for len() is expected string or array, but got fun")
	case VBuiltinFun:
		return nil, fmt.Errorf("argument for len() is expected string or array, but got fun")
	}
	return nil, fmt.Errorf("unknown value type: %s", args[0].Type().String())
}

func builtinPush(s *State, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("too many / less arguments for push()")
	}
	list, ok := args[0].(*VList)
	if !ok {
		return nil, fmt.Errorf("argument for push() is expected list, but got %s", args[0].Type().String())
	}
	list.Elements = append(list.Elements, args[1])
	return nil, nil
}

// helpLang resolves the two-letter language code from $LANG (e.g. "ja_JP.UTF-8" -> "ja").
func helpLang() string {
	lang := os.Getenv("LANG")
	if lang == "" || lang == "C" || lang == "POSIX" {
		return "en"
	}
	code := strings.Split(lang, "_")[0]
	code = strings.Split(code, ".")[0]
	if code == "" {
		return "en"
	}
	return code
}

// pickDoc selects the appropriate language entry, falling back to "en" if not found.
func pickDoc(docs map[string]string) string {
	if docs == nil {
		return ""
	}
	lang := helpLang()
	if v, ok := docs[lang]; ok {
		return v
	}
	return docs["en"]
}

func builtinHelp(s *State, args []Value) (Value, error) {
	if len(args) == 0 {
		// Print all top-level env names
		names := make([]string, 0)
		for k := range s.Env.Values {
			names = append(names, k)
		}
		sort.Strings(names)
		fmt.Println(strings.Join(names, "\n"))
		return nil, nil
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("help() takes 0 or 1 argument")
	}

	switch v := args[0].(type) {
	case *VModule:
		// Print module docstring
		moduleDoc := pickDoc(v.Docstring)
		if moduleDoc != "" {
			fmt.Println(moduleDoc)
		}
		// Print one-liner per exported field
		keys := make([]string, 0, len(v.Fields))
		for k := range v.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fieldDoc := pickDoc(v.FieldDocstrings[k])
			// Show only first line of field docstring as one-liner
			oneLiner := ""
			if fieldDoc != "" {
				oneLiner = " - " + strings.SplitN(fieldDoc, "\n", 2)[0]
			}
			fmt.Printf("- %s%s\n", k, oneLiner)
		}
	default:
		fmt.Printf("type: %s\n", args[0].Type())
	}
	return nil, nil
}
