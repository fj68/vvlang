package parser

import (
	"testing"

	"github.com/fj68/vvlang/ast"
)

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
