package interp

import (
	"testing"
)

func TestScoping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]float64
		err      bool
	}{
		{
			name: "if scoping",
			input: `
let x = 10
if true
  let x = 20
  let y = 30
end
`,
			expected: map[string]float64{"x": 10},
		},
		{
			name: "while scoping",
			input: `
let x = 10
let i = 0
while i < 1
  let x = 20
  let y = 30
  i = i + 1
end
`,
			expected: map[string]float64{"x": 10},
		},
		{
			name: "nested scoping",
			input: `
let x = 1
if true
  let x = 2
  if true
    let x = 3
  end
end
`,
			expected: map[string]float64{"x": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState("test.vv")
			err := s.Eval([]rune(tt.input))
			if (err != nil) != tt.err {
				t.Fatalf("Eval() error = %v, wantErr %v", err, tt.err)
			}

			for varName, expectedVal := range tt.expected {
				val, err := s.Env.Get(varName)
				if err != nil {
					t.Errorf("variable %s not found: %v", varName, err)
					continue
				}
				num, ok := val.(VNumber)
				if !ok {
					t.Errorf("variable %s is not a number, got %T", varName, val)
					continue
				}
				if float64(num) != expectedVal {
					t.Errorf("variable %s = %v, want %v", varName, num, expectedVal)
				}
			}

			// Verify local variables are NOT present
			if tt.name == "if scoping" || tt.name == "while scoping" {
				_, err := s.Env.Get("y")
				if err == nil {
					t.Errorf("variable y should be undefined out of scope")
				}
			}
		})
	}
}
