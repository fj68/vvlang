package interp

import (
	"testing"
)

func TestExtern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		globals map[string]Value
		wantErr bool
	}{
		{
			name:  "valid extern fun",
			input: "extern 'native' fun f() f()",
			globals: map[string]Value{
				"f": VBuiltinFun(func(s *State, args []Value) (Value, error) {
					return VNumber(42), nil
				}),
			},
			wantErr: false,
		},
		{
			name:  "valid extern let",
			input: "extern 'native' let v v + 1",
			globals: map[string]Value{
				"v": VNumber(10),
			},
			wantErr: false,
		},
		{
			name:    "extern missing name",
			input:   "extern 'native' fun g() g()",
			globals: map[string]Value{},
			wantErr: true,
		},
		{
			name:    "extern invalid placement",
			input:   "let x = 1 extern 'native' let v x + v",
			globals: map[string]Value{"v": VNumber(1)},
			wantErr: true,
		},
		{
			name:  "extern fun with args",
			input: "extern 'native' fun add(a, b) add(1, 2)",
			globals: map[string]Value{
				"add": VBuiltinFun(func(s *State, args []Value) (Value, error) {
					if len(args) != 2 {
						return nil, nil
					}
					return VNumber(float64(args[0].(VNumber) + args[1].(VNumber))), nil
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewState("test.vv")
			s.RegisterNatives(tt.globals)
			err := s.Eval([]rune(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("extern native exported", func(t *testing.T) {
		s := NewState("test.vv")
		s.RegisterNative("f", VBuiltinFun(func(s *State, args []Value) (Value, error) {
			return VNumber(42), nil
		}))
		err := s.Eval([]rune("pub extern 'native' fun f()"))
		if err != nil {
			t.Fatalf("Eval() error = %v", err)
		}
		val, err := s.Env.Get("f")
		if err != nil {
			t.Fatalf("Env.Get(\"f\") error = %v", err)
		}
		if _, ok := val.(VBuiltinFun); !ok {
			t.Errorf("expected VBuiltinFun, but got %T", val)
		}
	})
}
