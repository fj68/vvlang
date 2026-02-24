package interp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleImport(t *testing.T) {
	tmpDir := t.TempDir()

	// math.vv
	mathPath := filepath.Join(tmpDir, "math.vv")
	mathCode := `
pub let pi = 3.14
pub fun square(x) return x * x end
let secret = 42
`
	if err := os.WriteFile(mathPath, []byte(mathCode), 0644); err != nil {
		t.Fatal(err)
	}

	// main.vv
	mainPath := filepath.Join(tmpDir, "main.vv")
	mainCode := `
import math from './math.vv'
let result = math.square(math.pi)
`
	if err := os.WriteFile(mainPath, []byte(mainCode), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewState(mainPath)
	if err := s.Eval([]rune(mainCode)); err != nil {
		t.Fatal(err)
	}

	res, err := s.Env.Get("result")
	if err != nil {
		t.Fatalf("expected result: %v", err)
	}
	num, ok := res.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", res)
	}
	expected := 3.14 * 3.14
	if float64(num) != expected {
		t.Fatalf("expected %v, got %v", expected, num)
	}
}

func TestModuleVisibility(t *testing.T) {
	tmpDir := t.TempDir()

	// math.vv
	mathPath := filepath.Join(tmpDir, "math.vv")
	mathCode := `
pub let pi = 3.14
let secret = 42
`
	if err := os.WriteFile(mathPath, []byte(mathCode), 0644); err != nil {
		t.Fatal(err)
	}

	// main.vv
	mainPath := filepath.Join(tmpDir, "main.vv")
	mainCode := `
import math from './math.vv'
let result = math.secret
`
	if err := os.WriteFile(mainPath, []byte(mainCode), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewState(mainPath)
	err := s.Eval([]rune(mainCode))
	if err == nil {
		t.Fatal("expected error accessing non-exported symbol, but got nil")
	}
}

func TestNestedImport(t *testing.T) {
	tmpDir := t.TempDir()

	// constants.vv
	constPath := filepath.Join(tmpDir, "constants.vv")
	constCode := `pub let pi = 3.14`
	if err := os.WriteFile(constPath, []byte(constCode), 0644); err != nil {
		t.Fatal(err)
	}

	// math.vv
	mathPath := filepath.Join(tmpDir, "math.vv")
	mathCode := `
import c from './constants.vv'
pub fun circle_area(r) return c.pi * r * r end
`
	if err := os.WriteFile(mathPath, []byte(mathCode), 0644); err != nil {
		t.Fatal(err)
	}

	// main.vv
	mainPath := filepath.Join(tmpDir, "main.vv")
	mainCode := `
import math from './math.vv'
let result = math.circle_area(2)
`
	if err := os.WriteFile(mainPath, []byte(mainCode), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewState(mainPath)
	if err := s.Eval([]rune(mainCode)); err != nil {
		t.Fatal(err)
	}

	res, err := s.Env.Get("result")
	if err != nil {
		t.Fatalf("expected result: %v", err)
	}
	num, ok := res.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", res)
	}
	expected := 3.14 * 2 * 2
	if float64(num) != expected {
		t.Fatalf("expected %v, got %v", expected, num)
	}
}

func TestCyclicImport(t *testing.T) {
	tmpDir := t.TempDir()

	// a.vv
	aPath := filepath.Join(tmpDir, "a.vv")
	aCode := `import b from './b.vv'
pub let val = 1`
	if err := os.WriteFile(aPath, []byte(aCode), 0644); err != nil {
		t.Fatal(err)
	}

	// b.vv
	bPath := filepath.Join(tmpDir, "b.vv")
	bCode := `import a from './a.vv'
pub let val = 2`
	if err := os.WriteFile(bPath, []byte(bCode), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewState(aPath)
	err := s.Eval([]rune(aCode))
	if err == nil {
		t.Fatal("expected error for cyclic dependency, but got nil")
	}
	if !contains(err.Error(), "cyclic dependency") {
		t.Fatalf("expected 'cyclic dependency' error, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || find(s, substr))
}

func find(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
