package interp

import "testing"

func TestRecordLiteralEval(t *testing.T) {
	s := NewState()
	text := "return { name = 'value', key = 8 }"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	r, ok := v.(*VRecord)
	if !ok {
		t.Fatalf("expected *VRecord, got %T", v)
	}
	if len(r.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(r.Fields))
	}
	nameVal, ok := r.Fields["name"]
	if !ok {
		t.Fatalf("missing field 'name'")
	}
	ns, ok := nameVal.(VString)
	if !ok {
		t.Fatalf("expected VString for name, got %T", nameVal)
	}
	if ns != VString("value") {
		t.Fatalf("expected 'value', got %s", ns)
	}
	keyVal, ok := r.Fields["key"]
	if !ok {
		t.Fatalf("missing field 'key'")
	}
	kv, ok := keyVal.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber for key, got %T", keyVal)
	}
	if kv != VNumber(8) {
		t.Fatalf("expected 8, got %v", kv)
	}
}
