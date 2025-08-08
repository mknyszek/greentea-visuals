package main

import (
	"iter"
)

type GreenTea struct {
	// Immutable.
	Roots []Root
	Heap  *Heap

	// Mutable.
	RootsVisited  int
	Queue         Queue[*Block]
	Marked        Set[Pointer]
	Scanned       Set[Pointer]
	BlockVisited  int
	FieldsVisited map[Pointer]int

	CurrRoot   int
	CurrBlock  *Block
	CurrObject Pointer
	CurrField  int
}

func (g *GreenTea) heap() *Heap {
	return g.Heap
}

func (g *GreenTea) roots() ([]Root, int) {
	return g.Roots, g.RootsVisited
}

func (g *GreenTea) marked() *Set[Pointer] {
	return &g.Marked
}

func (g *GreenTea) fieldsVisited() map[Pointer]int {
	return g.FieldsVisited
}

func (g *GreenTea) curr() (int, *Block, Pointer, int) {
	return g.CurrRoot, g.CurrBlock, g.CurrObject, g.CurrField
}

func (g *GreenTea) init() {
	g.CurrRoot = -1
	g.CurrObject = -1
	g.CurrBlock = nil
	g.CurrField = -1
}

func (g *GreenTea) Next() bool {
	if g.FieldsVisited == nil {
		g.FieldsVisited = make(map[Pointer]int)
	}
	g.CurrRoot = -1
	g.CurrObject = -1
	g.CurrBlock = nil
	g.CurrField = -1
	if g.RootsVisited < len(g.Roots) {
		p := g.Roots[g.RootsVisited].Pointer
		if p != Nil && !g.Marked.Has(p) {
			g.Marked.Add(p)
			b := g.Heap.BlockOf(p)
			if !g.Queue.Has(b) {
				g.Queue.Push(b)
			}
		}
		g.CurrRoot = g.RootsVisited
		g.RootsVisited++
		return true
	}
	for {
		b, ok := g.Queue.Peek()
		if !ok {
			break
		}
		var p Pointer
		for g.BlockVisited < len(b.Objs) {
			p = b.Objs[g.BlockVisited]
			if g.Marked.Has(p) && !g.Scanned.Has(p) {
				break
			}
			g.BlockVisited++
		}
		if g.BlockVisited < len(b.Objs) {
			obj := &g.Heap.Objects[p]
			if g.FieldsVisited[p] == len(obj.Fields) {
				g.Scanned.Add(p)
			} else {
				for g.FieldsVisited[p] < len(obj.Fields) {
					f := obj.Fields[g.FieldsVisited[p]]
					fp := f.Pointer
					if fp == Nil || g.Marked.Has(fp) {
						g.CurrField = g.FieldsVisited[p]
						g.FieldsVisited[p]++
						break
					}
					g.Marked.Add(fp)
					b := g.Heap.BlockOf(fp)
					if !g.Queue.Has(b) {
						g.Queue.Push(b)
					}
					g.CurrField = g.FieldsVisited[p]
					g.FieldsVisited[p]++
					break
				}
				g.CurrObject = p
				g.CurrBlock = b
				return true
			}
		} else {
			g.Queue.Pop()
			g.BlockVisited = 0
		}
	}
	return false
}

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
