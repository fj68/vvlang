package interp

import (
	"strings"
	"testing"
)

type interpTestCase struct {
	name         string
	input        string
	globals      map[string]Value
	expected     Value
	expectedEnv  map[string]Value
	undefinedEnv []string
	expectedErr  string
	evalTest     bool
}

func TestInterp(t *testing.T) {
	tests := []interpTestCase{
		// --------------------------------------------------------------------
		// State basic eval
		// --------------------------------------------------------------------
		{
			name:        "state basic eval",
			input:       "fun add(a, b) return a + b end x = 1 let result = add(x, 0.5)",
			expectedEnv: map[string]Value{"result": VNumber(1.5)},
		},

		// --------------------------------------------------------------------
		// While loops & breaks
		// --------------------------------------------------------------------
		{
			name:        "while loop increments",
			input:       "i = 0 while i < 3 i = i + 1 end let result = i",
			expectedEnv: map[string]Value{"result": VNumber(3)},
		},
		{
			name:        "while break",
			input:       "i = 0 while true if i == 2 break end i = i + 1 end let result = i",
			expectedEnv: map[string]Value{"result": VNumber(2)},
		},
		{
			name:        "while continue",
			input:       "i = 0 j = 0 while i < 5 i = i + 1 if i == 2 continue end j = j + 1 end let result = j",
			expectedEnv: map[string]Value{"result": VNumber(4)},
		},

		// --------------------------------------------------------------------
		// Scoping checks
		// --------------------------------------------------------------------
		{
			name: "if scoping",
			input: `
let x = 10
if true
  let x = 20
  let y = 30
end
`,
			expectedEnv:  map[string]Value{"x": VNumber(10)},
			undefinedEnv: []string{"y"},
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
			expectedEnv:  map[string]Value{"x": VNumber(10)},
			undefinedEnv: []string{"y"},
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
			expectedEnv: map[string]Value{"x": VNumber(1)},
		},

		// --------------------------------------------------------------------
		// Defer evaluation
		// --------------------------------------------------------------------
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
			expectedEnv: map[string]Value{"x": VNumber(2)},
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
			expectedEnv: map[string]Value{"x": VNumber(22)},
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
			expectedEnv: map[string]Value{"x": VNumber(2)},
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
			expectedEnv: map[string]Value{"x": VNumber(1)},
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
			expectedEnv: map[string]Value{"x": VNumber(3)},
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
			expectedEnv: map[string]Value{"x": VNumber(2)},
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
			expectedEnv: map[string]Value{"x": VNumber(111)},
		},

		// --------------------------------------------------------------------
		// Extern declarations
		// --------------------------------------------------------------------
		{
			name:  "valid extern fun",
			input: "extern 'native' fun f() f()",
			globals: map[string]Value{
				"f": VBuiltinFun(func(s *State, args []Value) (Value, error) {
					return VNumber(42), nil
				}),
			},
		},
		{
			name:  "valid extern let",
			input: "extern 'native' let v v + 1",
			globals: map[string]Value{
				"v": VNumber(10),
			},
		},
		{
			name:        "extern missing name",
			input:       "extern 'native' fun g() g()",
			globals:     map[string]Value{},
			expectedErr: "extern native: g not registered",
		},
		{
			name:        "extern invalid placement",
			input:       "let x = 1 extern 'native' let v x + v",
			globals:     map[string]Value{"v": VNumber(1)},
			expectedErr: "extern statement must be after import statements and before other statements",
		},
		{
			name:  "extern fun with args",
			input: "extern 'native' fun add(a, b) let result = add(1, 2)",
			globals: map[string]Value{
				"add": VBuiltinFun(func(s *State, args []Value) (Value, error) {
					if len(args) != 2 {
						return nil, nil
					}
					return VNumber(float64(args[0].(VNumber) + args[1].(VNumber))), nil
				}),
			},
			expectedEnv: map[string]Value{"result": VNumber(3)},
		},

		// --------------------------------------------------------------------
		// Records
		// --------------------------------------------------------------------
		{
			name:  "record literal eval",
			input: "let result = { name = 'value', key = 8 }",
			expectedEnv: map[string]Value{
				"result": &VRecord{
					Fields: map[string]Value{
						"name": VString("value"),
						"key":  VNumber(8),
					},
				},
			},
		},
		{
			name:  "record literal eval trailing comma",
			input: "let result = { name = 'value', key = 8, }",
			expectedEnv: map[string]Value{
				"result": &VRecord{
					Fields: map[string]Value{
						"name": VString("value"),
						"key":  VNumber(8),
					},
				},
			},
		},
		{
			name:        "record field access",
			input:       "r = { name = 'value', key = 8 }\nlet result = r.name",
			expectedEnv: map[string]Value{"result": VString("value")},
		},
		{
			name:        "record field access number",
			input:       "r = { name = 'value', key = 8 }\nlet result = r.key",
			expectedEnv: map[string]Value{"result": VNumber(8)},
		},
		{
			name: "nested records with chained field access",
			input: `
admins = { alice = { name = 'Alice', age = 30 } }
fun get_alice(r)
  return r.alice
end
alice_name = get_alice(admins).name
let result = alice_name
`,
			expectedEnv: map[string]Value{"result": VString("Alice")},
		},

		// --------------------------------------------------------------------
		// Destructuring
		// --------------------------------------------------------------------
		{
			name: "simple destructuring",
			input: `let { a, b } = { a = 1, b = 2 }
let result = { x = a, y = b }`,
			expectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"x": VNumber(1),
					"y": VNumber(2),
				},
			}},
		},
		{
			name: "destructuring with alias",
			input: `let { a as x, b as y } = { a = 1, b = 2 }
let result = { r1 = x, r2 = y }`,
			expectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"r1": VNumber(1),
					"r2": VNumber(2),
				},
			}},
		},
		{
			name: "mixed punning and alias",
			input: `let { a, b as y } = { a = 1, b = 2 }
let result = { r1 = a, r2 = y }`,
			expectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"r1": VNumber(1),
					"r2": VNumber(2),
				},
			}},
		},
		{
			name: "destructuring from function return",
			input: `fun some_func()
  return { value = 100, error = null }
end
let { value, error } = some_func()
let result = { v = value, e = error }`,
			expectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"v": VNumber(100),
					"e": VNull{},
				},
			}},
		},
		{
			name:        "missing field",
			input:       "let { a, c } = { a = 1, b = 2 }",
			expectedErr: "field 'c' not found in record",
		},
		{
			name:        "not a record",
			input:       "let { a } = 1",
			expectedErr: "expected record for field access, but got number",
		},

		// --------------------------------------------------------------------
		// Str & interpolation
		// --------------------------------------------------------------------
		{
			name:        "str(number)",
			input:       `let result = str(123)`,
			expectedEnv: map[string]Value{"result": VString("123")},
		},
		{
			name:        "str(bool)",
			input:       `let result = str(true)`,
			expectedEnv: map[string]Value{"result": VString("true")},
		},
		{
			name:        "str(list)",
			input:       `let result = str([1, 2])`,
			expectedEnv: map[string]Value{"result": VString("[1, 2]")},
		},
		{
			name:        "str(record)",
			input:       `let result = str({a=1})`,
			expectedEnv: map[string]Value{"result": VString("{ a = 1 }")},
		},
		{
			name:        "str(null)",
			input:       `let result = str(null)`,
			expectedEnv: map[string]Value{"result": VString("null")},
		},
		{
			name:        "str(var)",
			input:       "let x = 8\n let result = str(x)",
			expectedEnv: map[string]Value{"result": VString("8")},
		},
		{
			name:        "interpolation basic",
			input:       "let name = \"world\"\nlet result = \"hello, {name}!\"",
			expectedEnv: map[string]Value{"result": VString("hello, world!")},
		},
		{
			name:        "interpolation math",
			input:       "let a = 1\nlet b = 2\nlet result = \"{a} + {b} = {a + b}\"",
			expectedEnv: map[string]Value{"result": VString("1 + 2 = 3")},
		},
		{
			name:        "interpolation list",
			input:       "let l = [1, 2]\nlet result = \"list: {l}\"",
			expectedEnv: map[string]Value{"result": VString("list: [1, 2]")},
		},
		{
			name:        "interpolation record",
			input:       "let r = { a = 1, b = \"s\" }\nlet result = \"record: {r}\"",
			expectedEnv: map[string]Value{"result": VString("record: { a = 1, b = s }")},
		},
		{
			name:        "interpolation nested braces syntax",
			input:       `let result = "nested: {{1}} {2}"`,
			expectedEnv: map[string]Value{"result": VString("nested: {1} 2")},
		},

		// --------------------------------------------------------------------
		// Blocks & returns
		// --------------------------------------------------------------------
		{
			name: "block stmt scoping check",
			input: `let x = 0
begin
  let y = 1
  x += 1
end`,
			expectedEnv:  map[string]Value{"x": VNumber(1)},
			undefinedEnv: []string{"y"},
		},
		{
			name:        "top level return value",
			input:       "return 1",
			expectedErr: "return statement is not allowed here (must be inside a function or block)",
		},

		// --------------------------------------------------------------------
		// Lists
		// --------------------------------------------------------------------
		{
			name:  "list literal eval",
			input: "let result = [0, 1, 2]",
			expectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{VNumber(0), VNumber(1), VNumber(2)},
			}},
		},
		{
			name:  "list literal eval trailing comma",
			input: "let result = [0, 1, 2, ]",
			expectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{VNumber(0), VNumber(1), VNumber(2)},
			}},
		},
		{
			name:  "list literal empty",
			input: "let result = []",
			expectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{},
			}},
		},
		{
			name:  "list literal mixed",
			input: "let result = [42, 'hello', true]",
			expectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{VNumber(42), VString("hello"), VBool(true)},
			}},
		},

		// --------------------------------------------------------------------
		// len() syntax
		// --------------------------------------------------------------------
		{
			name:        "len of string",
			input:       "let s = 'hello'\nlet n = len(s)",
			expectedEnv: map[string]Value{"n": VNumber(5)},
		},
		{
			name:        "len of unicode string",
			input:       "let s = 'こんにちは'\nlet n = len(s)",
			expectedEnv: map[string]Value{"n": VNumber(5)},
		},
		{
			name:        "len of empty string",
			input:       "let n = len('')",
			expectedEnv: map[string]Value{"n": VNumber(0)},
		},
		{
			name:        "len of list",
			input:       "let xs = [1, 2, 3]\nlet n = len(xs)",
			expectedEnv: map[string]Value{"n": VNumber(3)},
		},
		{
			name:        "len of empty list",
			input:       "let n = len([])",
			expectedEnv: map[string]Value{"n": VNumber(0)},
		},
		{
			name:        "len invalid type",
			input:       "let n = len(123)",
			expectedErr: "argument for len() is expected string or list, but got number",
		},

		// --------------------------------------------------------------------
		// Test blocks
		// --------------------------------------------------------------------
		{
			name:        "test block is not run in normal eval",
			input:       "let x = 1\ntest 'not run'\n  x = 2\nend\nlet result = x",
			expectedEnv: map[string]Value{"result": VNumber(1)},
		},
		{
			name:        "test block is run in eval test mode",
			input:       "test 'run'\n  let x = 2\n  assert x == 2\nend\n",
			expectedEnv: map[string]Value{"x": VNumber(2)},
			evalTest:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState("test.vv")
			if tt.globals != nil {
				s.RegisterNatives(tt.globals)
			}

			var err error
			if tt.evalTest {
				err = s.EvalTest([]rune(tt.input))
			} else {
				err = s.Eval([]rune(tt.input))
			}

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("Eval() error: %v", err)
			}

			// Check returned value
			if tt.expected != nil {
				if s.RetVals.Len() == 0 {
					t.Fatalf("expected return value %v, but RetVals is empty", tt.expected)
				}
				got := s.RetVals.Pop()
				eq, cmpErr := got.Equal(tt.expected)
				if cmpErr != nil {
					t.Fatalf("Equal() failed: %v", cmpErr)
				}
				if !eq {
					t.Errorf("RetVals.Pop() = %v, want %v", got, tt.expected)
				}
			}

			// Check local variables map[string]Value
			for varName, expectedVal := range tt.expectedEnv {
				val, err := s.Env.Get(varName)
				if err != nil {
					t.Errorf("expected variable %q not found: %v", varName, err)
					continue
				}
				eq, cmpErr := val.Equal(expectedVal)
				if cmpErr != nil {
					t.Fatalf("Equal() failed matching env variable %q: %v", varName, cmpErr)
				}
				if !eq {
					t.Errorf("variable %q = %v, want %v", varName, val, expectedVal)
				}
			}

			// Check variables that should be out of scope or undefined
			for _, varName := range tt.undefinedEnv {
				val, err := s.Env.Get(varName)
				if err == nil {
					t.Errorf("variable %q should be undefined, but got %v", varName, val)
				}
			}
		})
	}
}

// Ensure the specific extern export test runs properly logic decoupled from loop
func TestExternNativeExported(t *testing.T) {
	s := NewState("test.vv")
	s.RegisterNative("f", VBuiltinFun(func(s *State, args []Value) (Value, error) {
		return VNumber(42), nil
	}))
	err := s.Eval([]rune("pub extern 'native' fun f()"))
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	val, err := s.Env.Get("f")
	if err != nil {
		t.Fatalf("Env.Get(\"f\") error = %v", err)
	}
	if _, ok := val.(VBuiltinFun); !ok {
		t.Errorf("expected VBuiltinFun, but got %T", val)
	}
}
