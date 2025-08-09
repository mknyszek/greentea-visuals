package main

import "iter"

type GreenTea struct {
	// Immutable.
	roots []Root
	heap  *Heap

	// Mutable.
	rootsVisited  int
	queue         Queue[*Block]
	marked        Set[Pointer]
	scanned       Set[Pointer]
	blockVisited  int
	fieldsVisited map[Pointer]int
	ctx           Context
}

func NewGreenTea(roots []Root, heap *Heap) *GreenTea {
	return &GreenTea{
		roots:         roots,
		heap:          heap,
		fieldsVisited: make(map[Pointer]int),
		ctx:           Empty,
	}
}

func (g *GreenTea) reset() {
	*g = *NewGreenTea(g.roots, g.heap)
}

func (g *GreenTea) Heap() *Heap {
	return g.heap
}

func (g *GreenTea) Roots() ([]Root, int) {
	return g.roots, g.rootsVisited
}

func (g *GreenTea) Marked() *Set[Pointer] {
	return &g.marked
}

func (g *GreenTea) FieldsVisited() map[Pointer]int {
	return g.fieldsVisited
}

func (g *GreenTea) Context() Context {
	return g.ctx
}

func (g *GreenTea) Evolve() iter.Seq[gcState] {
	return g.evolve
}

func (g *GreenTea) evolve(yield func(gcState) bool) {
	defer g.reset()

	// First, the initial state.
	if !yield(g) {
		return
	}

	// Roots.
	for r := 0; r < len(g.roots); r++ {
		g.rootsVisited = r
		g.ctx.Root = r
		p := g.roots[r].Pointer

		// Yield selected root state.
		if !yield(g) {
			return
		}
		if p != Nil && !g.marked.Has(p) {
			g.marked.Add(p)
			b := g.heap.BlockOf(p)
			if !g.queue.Has(b) {
				g.queue.Push(b)
			}
		}

		// Yield marked object state.
		if !yield(g) {
			return
		}
	}

	// Finished with roots.
	g.rootsVisited = len(g.roots)
	g.ctx.Root = -1

	// Heap.
	for !g.queue.Empty() {
		// Take a block off the queue.
		b, _ := g.queue.Pop()
		g.ctx.Block = b

		// Yield new active block.
		if !yield(g) {
			return
		}

		// Iterate over marked-and-not-scanned objects.
		for _, p := range b.Objects {
			if !g.marked.Has(p) || g.scanned.Has(p) {
				continue
			}

			// Iterate over the object's fields and mark new objects,
			// adding their blocks to the queue if necessary.
			g.ctx.Object = p
			g.ctx.Field = -1

			// Yield new active object.
			if !yield(g) {
				return
			}

			obj := &g.heap.Objects[p]
			for i, f := range obj.Fields {
				g.ctx.Field = i

				// Yield new active field.
				if !yield(g) {
					return
				}

				fp := f.Pointer
				if fp == Nil || g.marked.Has(fp) {
					g.fieldsVisited[p]++
					continue
				}
				g.marked.Add(fp)
				b := g.heap.BlockOf(fp)
				if !g.queue.Has(b) {
					g.queue.Push(b)
				}
				g.fieldsVisited[p]++

				// Yield new object marked.
				if !yield(g) {
					return
				}
			}
			g.scanned.Add(p)
		}
	}

	// Deactivate everything.
	g.ctx.Block = nil
	g.ctx.Object = Nil
	g.ctx.Field = -1

	// Yield final state.
	if !yield(g) {
		return
	}
}
