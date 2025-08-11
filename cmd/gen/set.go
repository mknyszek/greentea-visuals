package main

import (
	"fmt"
	"strings"
)

type Set[T comparable] struct {
	m map[T]struct{}
}

func (s *Set[T]) Add(t T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}
	s.m[t] = struct{}{}
}

func (s *Set[T]) Has(t T) bool {
	if s.m == nil {
		return false
	}
	_, ok := s.m[t]
	return ok
}

func (s *Set[T]) Len() int {
	return len(s.m)
}

func (s *Set[T]) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	i := 0
	for e := range s.m {
		if i != 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "%v", e)
		i++
	}
	sb.WriteString("}")
	return sb.String()
}
