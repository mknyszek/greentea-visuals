package main

type MarkSweep struct {
	// Immutable.
	Roots []Root
	Heap  *Heap

	// Mutable.
	RootsVisited  int
	Stack         []Pointer
	Marked        Set[Pointer]
	FieldsVisited map[Pointer]int

	CurrRoot   int
	CurrBlock  *Block
	CurrObject Pointer
	CurrField  int
}

func (m *MarkSweep) heap() *Heap {
	return m.Heap
}

func (m *MarkSweep) roots() ([]Root, int) {
	return m.Roots, m.RootsVisited
}

func (m *MarkSweep) marked() *Set[Pointer] {
	return &m.Marked
}

func (m *MarkSweep) fieldsVisited() map[Pointer]int {
	return m.FieldsVisited
}

func (m *MarkSweep) curr() (int, *Block, Pointer, int) {
	return m.CurrRoot, m.CurrBlock, m.CurrObject, m.CurrField
}

func (m *MarkSweep) init() {
	m.CurrRoot = -1
	m.CurrObject = -1
	m.CurrBlock = nil
	m.CurrField = -1
}

func (m *MarkSweep) Next() bool {
	if m.FieldsVisited == nil {
		m.FieldsVisited = make(map[Pointer]int)
	}
	m.CurrRoot = -1
	m.CurrObject = -1
	m.CurrBlock = nil
	m.CurrField = -1
	if m.RootsVisited < len(m.Roots) {
		p := m.Roots[m.RootsVisited].Pointer
		if p != Nil && !m.Marked.Has(p) {
			m.Marked.Add(p)
			m.Stack = append(m.Stack, p)
		}
		m.CurrRoot = m.RootsVisited
		m.RootsVisited++
		return true
	}
	if n := len(m.Stack); n != 0 {
		p := m.Stack[n-1]
		obj := &m.Heap.Objects[p]
		if m.FieldsVisited[p] == len(obj.Fields) {
			m.Stack = m.Stack[:n-1]
		} else {
			for m.FieldsVisited[p] < len(obj.Fields) {
				f := obj.Fields[m.FieldsVisited[p]]
				if f.Pointer == Nil || m.Marked.Has(f.Pointer) {
					m.CurrField = m.FieldsVisited[p]
					m.FieldsVisited[p]++
					break
				}
				m.Marked.Add(f.Pointer)
				m.Stack = append(m.Stack, f.Pointer)
				m.CurrField = m.FieldsVisited[p]
				m.FieldsVisited[p]++
				break
			}
		}
		m.CurrObject = p
		return true
	}
	return false
}
