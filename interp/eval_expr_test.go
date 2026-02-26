package interp

import (
	"testing"
)

func TestEvalExpr(t *testing.T) {
	tests := []TestCase{
		{
			Name:  "record literal eval",
			Input: "let result = { name = 'value', key = 8 }",
			ExpectedEnv: map[string]Value{
				"result": &VRecord{
					Fields: map[string]Value{
						"name": StringToValue("value"),
						"key":  VNumber(8),
					},
				},
			},
		},
		{
			Name:  "record literal eval trailing comma",
			Input: "let result = { name = 'value', key = 8, }",
			ExpectedEnv: map[string]Value{
				"result": &VRecord{
					Fields: map[string]Value{
						"name": StringToValue("value"),
						"key":  VNumber(8),
					},
				},
			},
		},
		{
			Name:        "record field access",
			Input:       "r = { name = 'value', key = 8 }\nlet result = r.name",
			ExpectedEnv: map[string]Value{"result": StringToValue("value")},
		},
		{
			Name:        "record field access number",
			Input:       "r = { name = 'value', key = 8 }\nlet result = r.key",
			ExpectedEnv: map[string]Value{"result": VNumber(8)},
		},
		{
			Name: "nested records with chained field access",
			Input: `
admins = { alice = { name = 'Alice', age = 30 } }
fun get_alice(r)
  return r.alice
end
alice_name = get_alice(admins).name
let result = alice_name
`,
			ExpectedEnv: map[string]Value{"result": StringToValue("Alice")},
		},
		{
			Name: "simple destructuring",
			Input: `let { a, b } = { a = 1, b = 2 }
let result = { x = a, y = b }`,
			ExpectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"x": VNumber(1),
					"y": VNumber(2),
				},
			}},
		},
		{
			Name: "destructuring with alias",
			Input: `let { a as x, b as y } = { a = 1, b = 2 }
let result = { r1 = x, r2 = y }`,
			ExpectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"r1": VNumber(1),
					"r2": VNumber(2),
				},
			}},
		},
		{
			Name: "mixed punning and alias",
			Input: `let { a, b as y } = { a = 1, b = 2 }
let result = { r1 = a, r2 = y }`,
			ExpectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"r1": VNumber(1),
					"r2": VNumber(2),
				},
			}},
		},
		{
			Name: "destructuring from function return",
			Input: `fun some_func()
  return { value = 100, error = null }
end
let { value, error } = some_func()
let result = { v = value, e = error }`,
			ExpectedEnv: map[string]Value{"result": &VRecord{
				Fields: map[string]Value{
					"v": VNumber(100),
					"e": VNull{},
				},
			}},
		},
		{
			Name:        "missing field",
			Input:       "let { a, c } = { a = 1, b = 2 }",
			ExpectedErr: "field 'c' not found in record",
		},
		{
			Name:        "not a record",
			Input:       "let { a } = 1",
			ExpectedErr: "expected record for field access, but got number",
		},
		{
			Name:        "str(number)",
			Input:       `let result = str(123)`,
			ExpectedEnv: map[string]Value{"result": StringToValue("123")},
		},
		{
			Name:        "str(bool)",
			Input:       `let result = str(true)`,
			ExpectedEnv: map[string]Value{"result": StringToValue("true")},
		},
		{
			Name:        "str(list)",
			Input:       `let result = str([1, 2])`,
			ExpectedEnv: map[string]Value{"result": StringToValue("[1, 2]")},
		},
		{
			Name:        "str(record)",
			Input:       `let result = str({a=1})`,
			ExpectedEnv: map[string]Value{"result": StringToValue("{ a = 1 }")},
		},
		{
			Name:        "str(null)",
			Input:       `let result = str(null)`,
			ExpectedEnv: map[string]Value{"result": StringToValue("null")},
		},
		{
			Name:        "str(var)",
			Input:       "let x = 8\n let result = str(x)",
			ExpectedEnv: map[string]Value{"result": StringToValue("8")},
		},
		{
			Name:        "interpolation basic",
			Input:       "let name = \"world\"\nlet result = \"hello, {name}!\"",
			ExpectedEnv: map[string]Value{"result": StringToValue("hello, world!")},
		},
		{
			Name:        "interpolation math",
			Input:       "let a = 1\nlet b = 2\nlet result = \"{a} + {b} = {a + b}\"",
			ExpectedEnv: map[string]Value{"result": StringToValue("1 + 2 = 3")},
		},
		{
			Name:        "interpolation list",
			Input:       "let l = [1, 2]\nlet result = \"list: {l}\"",
			ExpectedEnv: map[string]Value{"result": StringToValue("list: [1, 2]")},
		},
		{
			Name:        "interpolation record",
			Input:       "let r = { a = 1, b = \"s\" }\nlet result = \"record: {r}\"",
			ExpectedEnv: map[string]Value{"result": StringToValue("record: { a = 1, b = s }")},
		},
		{
			Name:        "interpolation nested braces syntax",
			Input:       `let result = "nested: {{1}} {2}"`,
			ExpectedEnv: map[string]Value{"result": StringToValue("nested: {1} 2")},
		},
		{
			Name:  "list literal eval",
			Input: "let result = [0, 1, 2]",
			ExpectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{VNumber(0), VNumber(1), VNumber(2)},
			}},
		},
		{
			Name:  "list literal eval trailing comma",
			Input: "let result = [0, 1, 2, ]",
			ExpectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{VNumber(0), VNumber(1), VNumber(2)},
			}},
		},
		{
			Name:  "list literal empty",
			Input: "let result = []",
			ExpectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{},
			}},
		},
		{
			Name:  "list literal mixed",
			Input: "let result = [42, 'hello', true]",
			ExpectedEnv: map[string]Value{"result": &VList{
				Elements: []Value{VNumber(42), StringToValue("hello"), VBool(true)},
			}},
		},
		{
			Name:        "len of string",
			Input:       "let s = 'hello'\nlet n = len(s)",
			ExpectedEnv: map[string]Value{"n": VNumber(5)},
		},
		{
			Name:        "len of unicode string",
			Input:       "let s = 'こんにちは'\nlet n = len(s)",
			ExpectedEnv: map[string]Value{"n": VNumber(5)},
		},
		{
			Name:        "len of empty string",
			Input:       "let n = len('')",
			ExpectedEnv: map[string]Value{"n": VNumber(0)},
		},
		{
			Name:        "len of list",
			Input:       "let xs = [1, 2, 3]\nlet n = len(xs)",
			ExpectedEnv: map[string]Value{"n": VNumber(3)},
		},
		{
			Name:        "len of empty list",
			Input:       "let n = len([])",
			ExpectedEnv: map[string]Value{"n": VNumber(0)},
		},
		{
			Name:        "len invalid type",
			Input:       "let n = len(123)",
			ExpectedErr: "argument for len() is expected list, but got number",
		},
	}

	for _, tc := range tests {
		RunTest(t, tc)
	}
}
