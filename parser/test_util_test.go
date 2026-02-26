package parser

import (
	"strings"
	"testing"

	"github.com/fj68/vvlang/ast"
)

type TestCase struct {
	Name        string
	Input       string
	Expected    []ast.Stmt
	ExpectedErr string
}

func RunTest(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		program, err := Parse([]rune(tc.Input))
		if tc.ExpectedErr != "" {
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.ExpectedErr)
			}
			if !strings.Contains(err.Error(), tc.ExpectedErr) {
				t.Fatalf("expected error containing %q, got %q", tc.ExpectedErr, err.Error())
			}
			return
		}
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}

		if len(program.Statements) != len(tc.Expected) {
			t.Fatalf("expected %d statements, got %d.\nActual: %v\nExpected: %v", len(tc.Expected), len(program.Statements), program.Statements, tc.Expected)
		}

		for i, stmt := range program.Statements {
			if !stmt.Equals(tc.Expected[i]) {
				t.Errorf("statement %d mismatch.\nExpected: %s\nActual:   %s", i, tc.Expected[i].Inspect(), stmt.Inspect())
			}
		}
	})
}
