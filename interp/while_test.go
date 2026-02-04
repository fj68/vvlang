package interp

import "testing"

func TestWhileLoopIncrements(t *testing.T) {
	s := NewState()
	text := "i = 0 while i < 3 i = i + 1 end return i"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	n, ok := v.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", v)
	}
	if n != 3 {
		t.Fatalf("expected 3, got %v", n)
	}
}

func TestWhileBreak(t *testing.T) {
	s := NewState()
	text := "i = 0 while true if i == 2 break end i = i + 1 end return i"
	if err := s.Eval([]rune(text)); err != nil {
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

func TestWhileContinue(t *testing.T) {
	s := NewState()
	text := "i = 0 j = 0 while i < 5 i = i + 1 if i == 2 continue end j = j + 1 end return j"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	n, ok := v.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", v)
	}
	if n != 4 {
		t.Fatalf("expected 4, got %v", n)
	}
}
