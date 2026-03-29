package interp

import (
	"strings"
	"testing"
	"github.com/fj68/vvlang/mod"
)

type TestCase struct {
	Name         string
	Input        string
	Globals      map[string]Value
	Expected     Value
	ExpectedEnv  map[string]Value
	UndefinedEnv []string
	ExpectedErr  string
	EvalTest     bool
}

func RunTest(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		builder := NewStateBuilder(mod.DefaultConfig(), "test.vv")
		if tc.Globals != nil {
			builder.WithNatives(tc.Globals)
		}
		s := builder.Build()

		var err error
		if tc.EvalTest {
			err = s.EvalTest([]rune(tc.Input))
		} else {
			err = s.Eval([]rune(tc.Input))
		}

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
			t.Fatalf("Eval() error: %v", err)
		}

		// Check returned value
		if tc.Expected != nil {
			if s.RetVals.Len() == 0 {
				t.Fatalf("expected return value %v, but RetVals is empty", tc.Expected)
			}
			got := s.RetVals.Pop()
			eq, cmpErr := got.Equal(tc.Expected)
			if cmpErr != nil {
				t.Fatalf("Equal() failed: %v", cmpErr)
			}
			if !eq {
				t.Errorf("RetVals.Pop() = %v, want %v", got, tc.Expected)
			}
		}

		// Check local variables map[string]Value
		for varName, expectedVal := range tc.ExpectedEnv {
			val, err := s.ScopeManager.Resolve(varName)
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
		for _, varName := range tc.UndefinedEnv {
			val, err := s.ScopeManager.Resolve(varName)
			if err == nil {
				t.Errorf("variable %q should be undefined, but got %v", varName, val)
			}
		}
	})
}
