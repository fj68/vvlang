package parser

import (
	"strings"
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []ast.Stmt
		expectedErr string
	}{
		{
			name:  "simple assignment",
			input: "let x = 1",
			expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "x",
					Body: &ast.NumberLiteralExpr{Value: 1},
				},
			},
		},
		{
			name:  "function declaration",
			input: "let add = fun(a, b) return a + b end",
			expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "add",
					Body: &ast.FunLiteralExpr{
						Args: []string{"a", "b"},
						Body: []ast.Stmt{
							&ast.ReturnStmt{
								Value: &ast.InfixExpr{
									Op:    "+",
									Left:  &ast.VarRefExpr{Name: "a"},
									Right: &ast.VarRefExpr{Name: "b"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "function shorthand",
			input: "fun add(a, b) return a + b end",
			expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "add",
					Body: &ast.FunLiteralExpr{
						Args: []string{"a", "b"},
						Body: []ast.Stmt{
							&ast.ReturnStmt{
								Value: &ast.InfixExpr{
									Op:    "+",
									Left:  &ast.VarRefExpr{Name: "a"},
									Right: &ast.VarRefExpr{Name: "b"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "while statement",
			input: "while 1 < 2 let x = 1 end",
			expected: []ast.Stmt{
				&ast.WhileStmt{
					Cond: &ast.InfixExpr{
						Op:    "<",
						Left:  &ast.NumberLiteralExpr{Value: 1},
						Right: &ast.NumberLiteralExpr{Value: 2},
					},
					Body: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.NumberLiteralExpr{Value: 1},
							},
						},
					},
				},
			},
		},
		{
			name:  "if statement",
			input: "if true let x = 1 end",
			expected: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.BoolLiteralExpr{Value: true},
					Then: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.NumberLiteralExpr{Value: 1},
							},
						},
					},
				},
			},
		},
		{
			name:  "if else statement",
			input: "if true let x = 1 else let x = 2 end",
			expected: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.BoolLiteralExpr{Value: true},
					Then: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.NumberLiteralExpr{Value: 1},
							},
						},
					},
					Else: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.NumberLiteralExpr{Value: 2},
							},
						},
					},
				},
			},
		},
		{
			name:  "function call",
			input: "let y = add(x, 0.5)",
			expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "y",
					Body: &ast.FunCallExpr{
						Fun: &ast.VarRefExpr{Name: "add"},
						Args: []ast.Expr{
							&ast.VarRefExpr{Name: "x"},
							&ast.NumberLiteralExpr{Value: 0.5},
						},
					},
				},
			},
		},
		{
			name:  "record literal",
			input: "let r = { a = 1, b = 2 }",
			expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "r",
					Body: &ast.RecordLiteralExpr{
						Fields: map[string]ast.Expr{
							"a": &ast.NumberLiteralExpr{Value: 1},
							"b": &ast.NumberLiteralExpr{Value: 2},
						},
					},
				},
			},
		},
		{
			name:  "record literal trailing comma",
			input: "let r = { name = 'value', }",
			expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "r",
					Body: &ast.RecordLiteralExpr{
						Fields: map[string]ast.Expr{
							"name": &ast.StringLiteralExpr{Value: "value"},
						},
					},
				},
			},
		},
		{
			name:        "return top level",
			input:       "return 1",
			expectedErr: "return statement is not allowed here (must be inside a function or block)",
		},
		{
			name:        "return no value",
			input:       "return",
			expectedErr: "return statement is not allowed here (must be inside a function or block)",
		},
		{
			name:  "complex nested",
			input: "fun add(a, b) while true if get_key() == 'enter' return a + b else let x = 0.8 return x end end end",
			expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "add",
					Body: &ast.FunLiteralExpr{
						Args: []string{"a", "b"},
						Body: []ast.Stmt{
							&ast.WhileStmt{
								Cond: &ast.BoolLiteralExpr{Value: true},
								Body: &ast.BlockStmt{
									Body: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.InfixExpr{
												Op: "==",
												Left: &ast.FunCallExpr{
													Fun: &ast.VarRefExpr{Name: "get_key"},
												},
												Right: &ast.StringLiteralExpr{Value: "enter"},
											},
											Then: &ast.BlockStmt{
												Body: []ast.Stmt{
													&ast.ReturnStmt{
														Value: &ast.InfixExpr{
															Op:    "+",
															Left:  &ast.VarRefExpr{Name: "a"},
															Right: &ast.VarRefExpr{Name: "b"},
														},
													},
												},
											},
											Else: &ast.BlockStmt{
												Body: []ast.Stmt{
													&ast.VarDeclStmt{
														Name: "x",
														Body: &ast.NumberLiteralExpr{Value: 0.8},
													},
													&ast.ReturnStmt{
														Value: &ast.VarRefExpr{Name: "x"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "valid extern fun",
			input: "extern 'native' fun f(a, b)",
			expected: []ast.Stmt{
				&ast.ExternStmt{
					Type:     "native",
					Name:     "f",
					Exported: false,
				},
			},
		},
		{
			name:  "valid extern let",
			input: "extern 'native' let v",
			expected: []ast.Stmt{
				&ast.ExternStmt{
					Type:     "native",
					Name:     "v",
					Exported: false,
				},
			},
		},
		{
			name:  "multiple valid extern statements",
			input: "extern 'native' fun f1()\nextern 'native' let v1\nextern 'native' fun f2()",
			expected: []ast.Stmt{
				&ast.ExternStmt{
					Type:     "native",
					Name:     "f1",
					Exported: false,
				},
				&ast.ExternStmt{
					Type:     "native",
					Name:     "v1",
					Exported: false,
				},
				&ast.ExternStmt{
					Type:     "native",
					Name:     "f2",
					Exported: false,
				},
			},
		},
		{
			name:        "invalid extern placement",
			input:       "let x = 1\nextern 'native' let v",
			expectedErr: "extern statement must be after import statements and before other statements",
		},
		{
			name:        "invalid extern syntax - missing literal",
			input:       "extern fun f()",
			expectedErr: "expected string literal, but got Fun",
		},
		{
			name:        "invalid extern syntax - missing fun/let",
			input:       "extern 'native' f()",
			expectedErr: "expected 'fun' or 'let' after extern literal, but got Ident",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := Parse([]rune(tt.input))
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
				t.Fatalf("Parse error: %v", err)
			}

			if len(program.Statements) != len(tt.expected) {
				t.Fatalf("expected %d statements, got %d.\nActual: %v\nExpected: %v", len(tt.expected), len(program.Statements), program.Statements, tt.expected)
			}

			for i, stmt := range program.Statements {
				if !stmt.Equals(tt.expected[i]) {
					t.Errorf("statement %d mismatch.\nExpected: %s\nActual:   %s", i, tt.expected[i].Inspect(), stmt.Inspect())
				}
			}
		})
	}
}
func TestParseImportStmtsOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*ast.ImportStmt
	}{
		{
			name:  "single import",
			input: "import m from './m.vv'\nlet x = 1",
			expected: []*ast.ImportStmt{
				{Alias: "m", Path: "./m.vv"},
			},
		},
		{
			name:  "multiple imports",
			input: "import a from './a.vv'\nimport b from \"./b.vv\"\nfun f() end",
			expected: []*ast.ImportStmt{
				{Alias: "a", Path: "./a.vv"},
				{Alias: "b", Path: "./b.vv"},
			},
		},
		{
			name:  "module docstring",
			input: "/// module doc\nimport m from './m.vv'\nlet x = 1",
			expected: []*ast.ImportStmt{
				{Alias: "m", Path: "./m.vv"},
			},
		},
		{
			name:     "no imports",
			input:    "let x = 1",
			expected: []*ast.ImportStmt{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New([]rune(tt.input))
			imports, err := p.ParseImportStmtsOnly()
			if err != nil {
				t.Fatalf("ParseImportStmtsOnly error: %v", err)
			}

			if len(imports) != len(tt.expected) {
				t.Fatalf("expected %d imports, got %d", len(tt.expected), len(imports))
			}

			for i, imp := range imports {
				if imp.Alias != tt.expected[i].Alias || imp.Path != tt.expected[i].Path {
					t.Errorf("import %d mismatch. Expected Alias:%s Path:%s, got Alias:%s Path:%s",
						i, tt.expected[i].Alias, tt.expected[i].Path, imp.Alias, imp.Path)
				}
			}
		})
	}
}

func TestLenParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected ast.Expr
	}{
		{
			input: "len(xs)",
			expected: &ast.BuiltinCallExpr{
				Op:    "len",
				Value: &ast.VarRefExpr{Name: "xs"},
			},
		},
		{
			input: "len([1, 2, 3])",
			expected: &ast.BuiltinCallExpr{
				Op: "len",
				Value: &ast.ListLiteralExpr{
					Elements: []ast.Expr{
						&ast.NumberLiteralExpr{Value: 1},
						&ast.NumberLiteralExpr{Value: 2},
						&ast.NumberLiteralExpr{Value: 3},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New([]rune("let x = " + tt.input))
			module, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			stmt := module.Statements[0].(*ast.VarDeclStmt)
			if !stmt.Body.Equals(tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected.Inspect(), stmt.Body.Inspect())
			}
		})
	}
}
