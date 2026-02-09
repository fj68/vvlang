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

func TestRecordSpread(t *testing.T) {
	s := NewState()
	text := "r1 = { a = 1, b = 2 }\nr2 = { ...r1, c = 3 }\nreturn r2"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	r, ok := v.(*VRecord)
	if !ok {
		t.Fatalf("expected *VRecord, got %T", v)
	}
	if len(r.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(r.Fields))
	}
	
	// Check that spread fields are present
	if val, ok := r.Fields["a"]; !ok {
		t.Fatalf("missing field 'a' from spread")
	} else if vn, ok := val.(VNumber); !ok || vn != VNumber(1) {
		t.Fatalf("expected field 'a' to be 1, got %v", val)
	}
	
	if val, ok := r.Fields["b"]; !ok {
		t.Fatalf("missing field 'b' from spread")
	} else if vn, ok := val.(VNumber); !ok || vn != VNumber(2) {
		t.Fatalf("expected field 'b' to be 2, got %v", val)
	}
	
	// Check new field
	if val, ok := r.Fields["c"]; !ok {
		t.Fatalf("missing field 'c'")
	} else if vn, ok := val.(VNumber); !ok || vn != VNumber(3) {
		t.Fatalf("expected field 'c' to be 3, got %v", val)
	}
}

func TestRecordSpreadOverride(t *testing.T) {
	s := NewState()
	text := "r1 = { a = 1, b = 2 }\nr2 = { ...r1, b = 20 }\nreturn r2.b"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	vn, ok := v.(VNumber)
	if !ok {
		t.Fatalf("expected VNumber, got %T", v)
	}
	if vn != VNumber(20) {
		t.Fatalf("expected 20 (overridden value), got %v", vn)
	}
}

func TestRecordMultipleSpreads(t *testing.T) {
	s := NewState()
	text := "r1 = { a = 1 }\nr2 = { b = 2 }\nr3 = { ...r1, ...r2, c = 3 }\nreturn r3"
	if err := s.Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
	v := s.RetVals.Pop()
	r, ok := v.(*VRecord)
	if !ok {
		t.Fatalf("expected *VRecord, got %T", v)
	}
	if len(r.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(r.Fields))
	}
	
	// Check all fields are present
	expectedFields := map[string]float64{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	for fieldName, expectedValue := range expectedFields {
		if val, ok := r.Fields[fieldName]; !ok {
			t.Fatalf("missing field '%s'", fieldName)
		} else if vn, ok := val.(VNumber); !ok || vn != VNumber(expectedValue) {
			t.Fatalf("expected field '%s' to be %v, got %v", fieldName, expectedValue, val)
		}
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