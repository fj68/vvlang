package interp

import "testing"

func TestTopLevelReturnValue(t *testing.T) {
	s := NewState()
	if err := s.Eval([]rune("return 1")); err != nil {
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
