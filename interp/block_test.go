package interp

import (
	"testing"
)

func TestBlockStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected string // if empty, no error expected
	}{
		{
			input: `
let x = 0
begin
  let y = 1
  x += 1
end
// Here x should be 1, y should be undefined
`,
			expected: "",
		},
	}

	for _, tt := range tests {
		s := NewState()
		err := s.Eval([]rune(tt.input))
		if err != nil {
			t.Errorf("Eval(%q) failed: %v", tt.input, err)
		}

		// Check x is 1
		val, err := s.Env.Get("x")
		if err != nil {
			t.Errorf("expected x to be defined, got error: %v", err)
		} else if float64(val.(VNumber)) != 1 {
			t.Errorf("expected x=1, got %v", val)
		}

		// Check y is not defined
		_, err = s.Env.Get("y")
		if err == nil {
			t.Errorf("expected y to be undefined, but it was found")
		}
	}
}
