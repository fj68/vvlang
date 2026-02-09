package interp

import "testing"

func TestListLiteralEval(t *testing.T) {
	s := NewState()
	text := "return [0, 1, 2]"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	list, ok := v.(*VList)
	if !ok {
		t.Fatalf("expected *VList, got %T", v)
	}
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}
	for i := 0; i < 3; i++ {
		elem, ok := list.Elements[i].(VNumber)
		if !ok {
			t.Fatalf("expected VNumber for element %d, got %T", i, list.Elements[i])
		}
		if elem != VNumber(float64(i)) {
			t.Fatalf("expected %d, got %v", i, elem)
		}
	}
}

func TestListLiteralEvalTrailingComma(t *testing.T) {
	s := NewState()
	text := "return [0, 1, 2, ]"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	list, ok := v.(*VList)
	if !ok {
		t.Fatalf("expected *VList, got %T", v)
	}
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}
}

func TestListLiteralEmpty(t *testing.T) {
	s := NewState()
	text := "return []"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	list, ok := v.(*VList)
	if !ok {
		t.Fatalf("expected *VList, got %T", v)
	}
	if len(list.Elements) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(list.Elements))
	}
}

func TestListLiteralMixed(t *testing.T) {
	s := NewState()
	text := "return [42, 'hello', true]"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	list, ok := v.(*VList)
	if !ok {
		t.Fatalf("expected *VList, got %T", v)
	}
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}

	// Check first element is a number
	num, ok := list.Elements[0].(VNumber)
	if !ok {
		t.Fatalf("expected VNumber for element 0, got %T", list.Elements[0])
	}
	if num != VNumber(42) {
		t.Fatalf("expected 42, got %v", num)
	}

	// Check second element is a string
	str, ok := list.Elements[1].(VString)
	if !ok {
		t.Fatalf("expected VString for element 1, got %T", list.Elements[1])
	}
	if str != VString("hello") {
		t.Fatalf("expected 'hello', got %s", str)
	}

	// Check third element is a bool
	b, ok := list.Elements[2].(VBool)
	if !ok {
		t.Fatalf("expected VBool for element 2, got %T", list.Elements[2])
	}
	if !bool(b) {
		t.Fatalf("expected true, got %v", b)
	}
}
