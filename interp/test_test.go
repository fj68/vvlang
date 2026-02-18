package interp

import "testing"

func TestTestBlockIsNotRun(t *testing.T) {
	s := NewState()
	text := "let x = 1 test 'test block is not run in normal evaluation' x = 2 end return x"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	n, ok := v.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", v)
	}
	if n != 1 {
		t.Fatalf("expected 1, got %v", n)
	}
}

func TestTestBlockIsRun(t *testing.T) {
	s := NewState()
	text := "let x = 1 test 'test block is run in test-mode evaluation' let x = 2 assert x == 2 return x end return x"
	if err := s.EvalTest([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	n, ok := v.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", v)
	}
	if n != 2 {
		t.Fatalf("expected 2, got %v", n)
	}
}
