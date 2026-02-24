package interp

import (
	"strings"
	"testing"
)

func TestInterpolation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`let name = "world"; return "hello, {name}!"`, `"hello, world!"`},
		{`let a = 1; let b = 2; return "{a} + {b} = {a + b}"`, `"1 + 2 = 3"`},
		{`let l = [1, 2]; return "list: {l}"`, `"list: [1, 2]"`},
		{`let r = { a = 1, b = "s" }; return "record: {r}"`, `"record: { a = 1, b = s }"`},
		{`let f = fun() end; return "fun: {f}"`, `"fun: <function>"`},
		{`return "nested: {{1}} {2}"`, `"nested: {1} 2"`},
	}

	for _, tt := range tests {
		s := NewState("test.vv")
		// replace ; with \n for test convenience
		input := strings.ReplaceAll(tt.input, ";", "\n")
		err := s.Eval([]rune(input))
		if err != nil {
			t.Fatalf("eval failed for input %q: %v", tt.input, err)
		}
		if s.RetVals.Len() == 0 {
			t.Errorf("input %q: expected a return value on RetVals, but got none", tt.input)
			continue
		}
		got := s.RetVals.Pop().String()
		if got != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, got)
		}
	}
}
