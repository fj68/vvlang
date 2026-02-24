package interp

import (
	"testing"
)

func TestStrSyntax(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`return str(123)`, `"123"`},
		{`return str(true)`, `"true"`},
		{`return str([1, 2])`, `"[1, 2]"`},
		{`return str({a=1})`, `"{ a = 1 }"`},
		{`return str(null)`, `"null"`},
		{"let x = 8\n return str(x)", `"8"`},
	}

	for _, tt := range tests {
		s := NewState("test.vv")
		err := s.Eval([]rune(tt.input))
		if err != nil {
			t.Fatalf("eval failed for input %q: %v", tt.input, err)
		}
		got := s.RetVals.Pop().String()
		if got != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, got)
		}
	}
}
