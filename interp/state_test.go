package interp

import "testing"

func TestState(t *testing.T) {
	text := "fun add(a, b) return a + b end var x = 1 return add(x, 0.5)"
	if err := Eval([]rune(text)); err != nil {
		t.Fatal(err)
	}
}
