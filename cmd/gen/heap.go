// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "slices"

type Block struct {
	Address  uint64
	ElemSize int
	Objects  []Pointer
}

func Blk(addr uint64, esize int, objs ...Pointer) Block {
	return Block{addr, esize, objs}
}

type Object struct {
	Type   string
	Fields []Field
}

type Field struct {
	Offset  int
	Pointer Pointer
}

func Obj(typ string, ptrs ...Field) Object {
	return Object{Type: typ, Fields: ptrs}
}

func F(offset int, p Pointer) Field {
	return Field{offset, p}
}

const Nil Pointer = 0

const Free Pointer = 1

type Pointer int

const PointerSize = 8

type Heap struct {
	Objects []Object
	Blocks  []Block
}

func (h *Heap) BlockOf(p Pointer) *Block {
	b, _ := h.BlockIdx(p)
	return b
}

func (h *Heap) BlockIdx(p Pointer) (*Block, int) {
	for j := range h.Blocks {
		if i := slices.Index(h.Blocks[j].Objects, p); i >= 0 {
			return &h.Blocks[j], i
		}
	}
	return nil, -1
}

func (h *Heap) AddressOf(p Pointer) uint64 {
	b, i := h.BlockIdx(p)
	if b == nil {
		return 0
	}
	return b.Address + uint64(b.ElemSize*i)
}

type Root struct {
	Name    string
	Pointer Pointer
}

type Context struct {
	Root   int
	Block  *Block
	Object Pointer
	Field  int
}

var Empty = Context{-1, nil, Nil, -1}
