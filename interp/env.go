package interp

import "fmt"

type Env struct {
	Values map[string]Value
	outer  *Env
}

func (env *Env) Get(name string) (Value, error) {
	if v, ok := env.Values[name]; ok {
		return v, nil
	}

	if env.outer == nil {
		return nil, fmt.Errorf("no variable named '%s' is not found", name)
	}

	return env.outer.Get(name)
}

func (env *Env) Set(name string, value Value) {
	for e := env; e.outer != nil; e = e.outer {
		if _, ok := e.Values[name]; ok {
			e.Values[name] = value
			return
		}
	}
	env.Values[name] = value
}
