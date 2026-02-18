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

func TestRecordLiteralEvalTrailingComma(t *testing.T) {
	s := NewState()
	text := "return { name = 'value', key = 8, }"
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
}
func TestRecordFieldAccess(t *testing.T) {
	s := NewState()
	text := "r = { name = 'value', key = 8 }\nreturn r.name"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	vs, ok := v.(VString)
	if !ok {
		t.Fatalf("expected VString, got %T", v)
	}
	if vs != VString("value") {
		t.Fatalf("expected 'value', got %s", vs)
	}
}

func TestRecordFieldAccessNumber(t *testing.T) {
	s := NewState()
	text := "r = { name = 'value', key = 8 }\nreturn r.key"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	vn, ok := v.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", v)
	}
	if vn != VNumber(8) {
		t.Fatalf("expected 8, got %v", vn)
	}
}

func TestNestedRecordsWithChainedFieldAccess(t *testing.T) {
	s := NewState()
	text := `admins = { alice = { name = 'Alice', age = 30 } }
fun get_alice(r)
  return r.alice
end
alice_name = get_alice(admins).name
return alice_name`
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	vs, ok := v.(VString)
	if !ok {
		t.Fatalf("expected VString, got %T", v)
	}
	if vs != VString("Alice") {
		t.Fatalf("expected 'Alice', got %s", vs)
	}
}