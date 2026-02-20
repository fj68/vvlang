package interp

import (
	"testing"
)

func TestDefer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]float64
	}{
		{
			name: "basic defer",
			input: `
let x = 1
let f = fun() x = 2 end
begin
  defer f()
  x = 3
end
`,
			expected: map[string]float64{"x": 2},
		},
		{
			name: "multiple defers (LIFO)",
			input: `
let x = 1
let f1 = fun() x = x * 2 end
let f2 = fun() x = x + 1 end
begin
  defer f1()
  defer f2()
  x = 10
end
`,
			// x = 10
			// defer f2() -> x = 11
			// defer f1() -> x = 22
			expected: map[string]float64{"x": 22},
		},
		{
			name: "defer in if",
			input: `
let x = 1
let f = fun() x = 2 end
if true
  defer f()
  x = 3
end
`,
			expected: map[string]float64{"x": 2},
		},
		{
			name: "defer in false if",
			input: `
let x = 1
let f = fun() x = 2 end
if false
  defer f()
  x = 3
end
`,
			expected: map[string]float64{"x": 1},
		},
		{
			name: "defer in while",
			input: `
let x = 0
let i = 0
let f = fun() x = x + 1 end
while i < 3
  defer f()
  i = i + 1
end
`,
			// Each iteration has its own defer scope.
			// Iteration 1: defer f() -> x = 1
			// Iteration 2: defer f() -> x = 2
			// Iteration 3: defer f() -> x = 3
			expected: map[string]float64{"x": 3},
		},
		{
			name: "defer in function",
			input: `
let x = 1
let f_defer = fun() x = 2 end
let f = fun()
  defer f_defer()
  x = 3
end
f()
`,
			expected: map[string]float64{"x": 2},
		},
		{
			name: "nested blocks",
			input: `
let x = 0
let f1 = fun() x = x + 1 end
let f10 = fun() x = x + 10 end
begin
  defer f1()
  begin
    defer f10()
    x = 100
  end
end
`,
			expected: map[string]float64{"x": 111},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState("test.vv")
			err := s.Eval([]rune(tt.input))
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
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
		})
	}
}
