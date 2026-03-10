package interpreter

type Environment struct {
	Values map[string]Value
	Outer  *Environment
}

func NewEnvironment(outer *Environment) *Environment {
	return &Environment{
		Values: make(map[string]Value),
		Outer:  outer,
	}
}

func (e *Environment) Get(name string) (Value, bool) {
	if v, ok := e.Values[name]; ok {
		return v, true
	}
	if e.Outer != nil {
		return e.Outer.Get(name)
	}
	return nil, false
}

func (e *Environment) Set(name string, val Value) {
	e.Values[name] = val
}

func (e *Environment) Update(name string, val Value) {
	if _, ok := e.Values[name]; ok {
		e.Values[name] = val
		return
	}
	if e.Outer != nil {
		e.Outer.Update(name, val)
		return
	}
	// Fallback: define in current scope (or panic?)
	e.Values[name] = val
}
