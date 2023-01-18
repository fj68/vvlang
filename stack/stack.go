package stack

type Element[T any] struct {
	Value T
	Next  *Element[T]
}

type Stack[T any] struct {
	Top  *Element[T]
	size int
}

func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Len() int {
	return s.size
}

func (s *Stack[T]) Peek() T {
	return s.Top.Value
}

func (s *Stack[T]) Pop() T {
	if s.Top == nil {
		var zero T
		return zero
	}
	v := s.Top.Value
	s.Top = s.Top.Next
	s.size--
	return v
}

func (s *Stack[T]) Push(value T) {
	s.Top = &Element[T]{
		Value: value,
		Next:  s.Top,
	}
	s.size++
}
