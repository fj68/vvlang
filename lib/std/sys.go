package std

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fj68/vvlang/interp"
)

func Help(s *interp.State, args []interp.Value) (interp.Value, error) {
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
	case *interp.VModule:
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
func PhysEq(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("phys_eq() takes 2 arguments")
	}
	return interp.VBool(args[0] == args[1]), nil
}

func SetMaxRecursionDepth(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("set_max_recursion_depth() takes 1 argument")
	}
	depth, ok := args[0].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("set_max_recursion_depth() argument must be an integer, got %s", args[0].Type())
	}
	s.MaxRecursionDepth = int(depth)
	return interp.NoneValue, nil
}

func GetMaxRecursionDepth(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("get_max_recursion_depth() takes 0 arguments")
	}
	return interp.VInt(int64(s.MaxRecursionDepth)), nil
}

func Type(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("type() takes 1 argument")
	}
	return interp.StringToValue(args[0].Type().String()), nil
}

func Str(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("str() takes 1 argument")
	}
	return interp.StringToValue(args[0].Str()), nil
}
