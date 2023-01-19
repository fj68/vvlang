package interp

import (
	"fmt"
	"strings"
)

type Env struct {
	Values map[string]Value
	outer  *Env
}

func NewEnv(outer *Env) *Env {
	return &Env{
		Values: map[string]Value{},
		outer:  outer,
	}
}

func (env *Env) Get(name string) (Value, error) {
	if v, ok := env.Values[name]; ok {
		return v, nil
	}

	if env.outer == nil {
		return nil, fmt.Errorf("variable named '%s' is not found", name)
	}

	return env.outer.Get(name)
}

func (env *Env) Set(name string, value Value) {
	for e := env.outer; e != nil; e = e.outer {
		if _, ok := e.Values[name]; ok {
			e.Values[name] = value
			return
		}
	}
	env.Values[name] = value
}

func (env *Env) String() string {
	var b strings.Builder
	for name, value := range env.Values {
		b.WriteString(fmt.Sprintf("%s = %s\n", name, value))
	}
	return b.String()
}
