package interp

import (
	"testing"
	"github.com/fj68/vvlang/mod"
)

func TestExternNativeExported(t *testing.T) {
	s := NewState(mod.DefaultConfig(), "test.vv")
	s.RegisterNative("f", VBuiltinFun(func(s *State, args []Value) (Value, error) {
		return VInt(42), nil
	}))
	err := s.Eval([]rune("pub extern 'native' fun f()"))
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	val, err := s.ScopeManager.Resolve("f")
	if err != nil {
		t.Fatalf("Env.Get(\"f\") error = %v", err)
	}
	if _, ok := val.(VBuiltinFun); !ok {
		t.Errorf("expected VBuiltinFun, but got %T", val)
	}
}
