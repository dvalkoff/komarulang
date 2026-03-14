package env


type Environment[T any] struct {
	data map[string]T
	parent *Environment[T]
}

func NewEnvironment[T any](parent *Environment[T]) *Environment[T] {
	return &Environment[T]{
		data: map[string]T{},
		parent: parent,
	}
}

func (e *Environment[T]) Exists(key string) bool {
	if e == nil {
		return false
	}
	if _, ok := e.data[key]; ok {
		return true
	}
	return e.parent.Exists(key)
}

func (e *Environment[T]) Get(key string) (T, bool) {
	if e == nil {
		var zero T
		return zero, false
	}
	if val, ok := e.data[key]; ok {
		return val, true
	}
	return e.parent.Get(key)
}

func (e *Environment[T]) New(key string, val T) {
	e.data[key] = val
}

func (e *Environment[T]) Set(key string, val T) {
	if _, ok := e.data[key]; ok {
		e.data[key] = val
		return
	}
	e.parent.Set(key, val)
}
