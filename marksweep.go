package main

import "iter"

type MarkSweep struct {
	// Immutable.
	roots []Root
	heap  *Heap

	// Mutable.
	rootsVisited  int
	stack         []Pointer
	marked        Set[Pointer]
	fieldsVisited map[Pointer]int
	ctx           Context
}

func NewMarkSweep(roots []Root, heap *Heap) *MarkSweep {
	return &MarkSweep{
		roots:         roots,
		heap:          heap,
		fieldsVisited: make(map[Pointer]int),
		ctx:           Empty,
	}
}

func (m *MarkSweep) reset() {
	*m = *NewMarkSweep(m.roots, m.heap)
}

func (m *MarkSweep) Heap() *Heap {
	return m.heap
}

func (m *MarkSweep) Roots() ([]Root, int) {
	return m.roots, m.rootsVisited
}

func (m *MarkSweep) Marked() *Set[Pointer] {
	return &m.marked
}

func (m *MarkSweep) FieldsVisited() map[Pointer]int {
	return m.fieldsVisited
}

func (m *MarkSweep) Context() Context {
	return m.ctx
}

func (m *MarkSweep) Evolve() iter.Seq[gcState] {
	return m.evolve
}

func (m *MarkSweep) evolve(yield func(gcState) bool) {
	defer m.reset()

	// First, the initial state.
	if !yield(m) {
		return
	}

	// Roots.
	for r := 0; r < len(m.roots); r++ {
		m.rootsVisited = r
		m.ctx.Root = r
		p := m.roots[r].Pointer

		// Yield selected root state.
		if !yield(m) {
			return
		}
		if p != Nil && !m.marked.Has(p) {
			m.marked.Add(p)
			m.stack = append(m.stack, p)
		}

		// Yield marked object state.
		if !yield(m) {
			return
		}
	}

	// Finished with roots.
	m.rootsVisited = len(m.roots)
	m.ctx.Root = -1

	// Heap.
	for len(m.stack) != 0 {
		// Take an object off the stack.
		p := m.stack[len(m.stack)-1]
		m.stack = m.stack[:len(m.stack)-1]

		// Yield new active block.
		if !yield(m) {
			return
		}

		// Iterate over the object's fields and mark new objects,
		// adding their blocks to the queue if necessary.
		m.ctx.Object = p
		m.ctx.Field = -1

		// Yield new active object.
		if !yield(m) {
			return
		}

		obj := &m.heap.Objects[p]
		for i, f := range obj.Fields {
			m.ctx.Field = i

			// Yield new active field.
			if !yield(m) {
				return
			}

			fp := f.Pointer
			if fp == Nil || m.marked.Has(fp) {
				m.fieldsVisited[p]++
				continue
			}
			m.marked.Add(fp)
			m.stack = append(m.stack, fp)
			m.fieldsVisited[p]++

			// Yield new object marked.
			if !yield(m) {
				return
			}
		}
	}

	// Deactivate everything.
	m.ctx.Block = nil
	m.ctx.Object = Nil
	m.ctx.Field = -1

	// Yield final state.
	if !yield(m) {
		return
	}
}
