package main

import "iter"

type Queue[T comparable] struct {
	head, tail *qItem[T]
}

func (q *Queue[T]) Empty() bool {
	return q.head == nil
}

func (q *Queue[T]) Has(value T) bool {
	for i := range q.All() {
		if value == i {
			return true
		}
	}
	return false
}

func (q *Queue[T]) Push(value T) {
	i := &qItem[T]{value: value}
	if q.tail == nil {
		q.head, q.tail = i, i
		return
	}
	q.tail.next = i
	q.tail = i
}

func (q *Queue[T]) Peek() (T, bool) {
	if q.Empty() {
		return *new(T), false
	}
	return q.head.value, true
}

func (q *Queue[T]) Pop() (T, bool) {
	if q.Empty() {
		return *new(T), false
	}
	i := q.head
	q.head = i.next
	if q.head == nil {
		q.tail = nil
	}
	i.next = nil
	return i.value, true
}

func (q *Queue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		i := q.head
		for i != nil {
			if !yield(i.value) {
				break
			}
			i = i.next
		}
	}
}

type qItem[T any] struct {
	next  *qItem[T]
	value T
}
