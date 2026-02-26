package interp

import (
	"testing"
)

func TestEvalStmt(t *testing.T) {
	tests := []TestCase{
		{
			Name:        "state basic eval",
			Input:       "fun add(a, b) return a + b end x = 1 let result = add(x, 0.5)",
			ExpectedEnv: map[string]Value{"result": VNumber(1.5)},
		},
		{
			Name:        "while loop increments",
			Input:       "i = 0 while i < 3 i = i + 1 end let result = i",
			ExpectedEnv: map[string]Value{"result": VNumber(3)},
		},
		{
			Name:        "while break",
			Input:       "i = 0 while true if i == 2 break end i = i + 1 end let result = i",
			ExpectedEnv: map[string]Value{"result": VNumber(2)},
		},
		{
			Name:        "while continue",
			Input:       "i = 0 j = 0 while i < 5 i = i + 1 if i == 2 continue end j = j + 1 end let result = j",
			ExpectedEnv: map[string]Value{"result": VNumber(4)},
		},
		{
			Name: "if scoping",
			Input: `
let x = 10
if true
  let x = 20
  let y = 30
end
`,
			ExpectedEnv:  map[string]Value{"x": VNumber(10)},
			UndefinedEnv: []string{"y"},
		},
		{
			Name: "while scoping",
			Input: `
let x = 10
let i = 0
while i < 1
  let x = 20
  let y = 30
  i = i + 1
end
`,
			ExpectedEnv:  map[string]Value{"x": VNumber(10)},
			UndefinedEnv: []string{"y"},
		},
		{
			Name: "nested scoping",
			Input: `
let x = 1
if true
  let x = 2
  if true
    let x = 3
  end
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(1)},
		},
		{
			Name: "basic defer",
			Input: `
let x = 1
let f = fun() x = 2 end
begin
  defer f()
  x = 3
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(2)},
		},
		{
			Name: "multiple defers (LIFO)",
			Input: `
let x = 1
let f1 = fun() x = x * 2 end
let f2 = fun() x = x + 1 end
begin
  defer f1()
  defer f2()
  x = 10
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(22)},
		},
		{
			Name: "defer in if",
			Input: `
let x = 1
let f = fun() x = 2 end
if true
  defer f()
  x = 3
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(2)},
		},
		{
			Name: "defer in false if",
			Input: `
let x = 1
let f = fun() x = 2 end
if false
  defer f()
  x = 3
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(1)},
		},
		{
			Name: "defer in while",
			Input: `
let x = 0
let i = 0
let f = fun() x = x + 1 end
while i < 3
  defer f()
  i = i + 1
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(3)},
		},
		{
			Name: "defer in function",
			Input: `
let x = 1
let f_defer = fun() x = 2 end
let f = fun()
  defer f_defer()
  x = 3
end
f()
`,
			ExpectedEnv: map[string]Value{"x": VNumber(2)},
		},
		{
			Name: "nested blocks",
			Input: `
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
			ExpectedEnv: map[string]Value{"x": VNumber(111)},
		},
		{
			Name:  "valid extern fun",
			Input: "extern 'native' fun f() f()",
			Globals: map[string]Value{
				"f": VBuiltinFun(func(s *State, args []Value) (Value, error) {
					return VNumber(42), nil
				}),
			},
		},
		{
			Name:  "valid extern let",
			Input: "extern 'native' let v v + 1",
			Globals: map[string]Value{
				"v": VNumber(10),
			},
		},
		{
			Name:        "extern missing name",
			Input:       "extern 'native' fun g() g()",
			Globals:     map[string]Value{},
			ExpectedErr: "extern native: g not registered",
		},
		{
			Name:        "extern invalid placement",
			Input:       "let x = 1 extern 'native' let v x + v",
			Globals:     map[string]Value{"v": VNumber(1)},
			ExpectedErr: "extern statement must be after import statements and before other statements",
		},
		{
			Name:  "extern fun with args",
			Input: "extern 'native' fun add(a, b) let result = add(1, 2)",
			Globals: map[string]Value{
				"add": VBuiltinFun(func(s *State, args []Value) (Value, error) {
					if len(args) != 2 {
						return nil, nil
					}
					return VNumber(float64(args[0].(VNumber) + args[1].(VNumber))), nil
				}),
			},
			ExpectedEnv: map[string]Value{"result": VNumber(3)},
		},
		{
			Name: "block stmt scoping check",
			Input: `let x = 0
begin
  let y = 1
  x += 1
end`,
			ExpectedEnv:  map[string]Value{"x": VNumber(1)},
			UndefinedEnv: []string{"y"},
		},
		{
			Name:        "top level return value",
			Input:       "return 1",
			ExpectedErr: "return statement is not allowed here (must be inside a function or block)",
		},
		{
			Name:        "test block is not run in normal eval",
			Input:       "let x = 1\ntest 'not run'\n  x = 2\nend\nlet result = x",
			ExpectedEnv: map[string]Value{"result": VNumber(1)},
		},
		{
			Name:        "test block is run in eval test mode",
			Input:       "test 'run'\n  let x = 2\n  assert x == 2\nend\n",
			ExpectedEnv: map[string]Value{"x": VNumber(2)},
			EvalTest:    true,
		},
		{
			Name: "top-level let is evaluated in test mode",
			Input: `
let x = 10
test 'use let'
  assert x == 10
  let y = x + 5
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(10), "y": VNumber(15)},
			EvalTest:    true,
		},
		{
			Name: "top-level fun is evaluated in test mode",
			Input: `
fun add(a, b) return a + b end
test 'use fun'
  let result = add(2, 3)
  assert result == 5
end
`,
			ExpectedEnv: map[string]Value{"result": VNumber(5)},
			EvalTest:    true,
		},
		{
			Name: "top-level extern is evaluated in test mode",
			Input: `
extern 'native' fun my_native()
test 'use extern'
  let result = my_native()
  assert result == 100
end
`,
			Globals: map[string]Value{
				"my_native": VBuiltinFun(func(s *State, args []Value) (Value, error) {
					return VNumber(100), nil
				}),
			},
			ExpectedEnv: map[string]Value{"result": VNumber(100)},
			EvalTest:    true,
		},
		{
			Name: "top-level assignments and expressions are ignored in test mode",
			Input: `
let x = 1
x = 2
test 'check x'
  assert x == 1
end
`,
			ExpectedEnv: map[string]Value{"x": VNumber(1)},
			EvalTest:    true,
		},
	}

	for _, tc := range tests {
		RunTest(t, tc)
	}
}
