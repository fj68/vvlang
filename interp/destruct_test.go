package interp

import "testing"

func TestRecordDestructuring(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]Value
	}{
		{
			name:  "simple destructuring",
			input: "let { a, b } = { a = 1, b = 2 }\nreturn { x = a, y = b }",
			expected: map[string]Value{
				"x": VNumber(1),
				"y": VNumber(2),
			},
		},
		{
			name:  "destructuring with alias",
			input: "let { a as x, b as y } = { a = 1, b = 2 }\nreturn { r1 = x, r2 = y }",
			expected: map[string]Value{
				"r1": VNumber(1),
				"r2": VNumber(2),
			},
		},
		{
			name:  "mixed punning and alias",
			input: "let { a, b as y } = { a = 1, b = 2 }\nreturn { r1 = a, r2 = y }",
			expected: map[string]Value{
				"r1": VNumber(1),
				"r2": VNumber(2),
			},
		},
		{
			name: "destructuring from function return",
			input: `
fun some_func()
  return { value = 100, error = null }
end
let { value, error } = some_func()
return { v = value, e = error }
`,
			expected: map[string]Value{
				"v": VNumber(100),
				"e": VNull{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState("test.vv")
			if err := s.Eval([]rune(tt.input)); err != nil {
				t.Fatal(err)
			}
			v := s.RetVals.Pop()
			r, ok := v.(*VRecord)
			if !ok {
				t.Fatalf("expected *VRecord, got %T", v)
			}
			for k, expectedVal := range tt.expected {
				actualVal, ok := r.Fields[k]
				if !ok {
					t.Fatalf("missing field '%s'", k)
				}
				eq, err := actualVal.Equal(expectedVal)
				if err != nil {
					t.Fatal(err)
				}
				if !eq {
					t.Errorf("field '%s': expected %v, got %v", k, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestRecordDestructuringError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing field",
			input: "let { a, c } = { a = 1, b = 2 }",
		},
		{
			name:  "not a record",
			input: "let { a } = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState("test.vv")
			err := s.Eval([]rune(tt.input))
			if err == nil {
				t.Fatal("expected error but got nil")
			}
		})
	}
}
