package main

import (
	"fmt"
	"image"
	"image/color"
	"iter"
	"log"
	"math"
	"os"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func main() {
	roots := []Root{
		{"var a *T", 2},
		{"var b *T", 6},
	}
	heap := &Heap{
		Objects: []Object{
			Nil:  Obj("nil"),    // Nil.
			Free: Obj("<free>"), // Free block sentinel.
			2:    Obj("T", F(0, 4)),
			3:    Obj("T", F(0, Nil)),
			4:    Obj("[4]*T", F(0, Nil), F(8, Nil), F(16, 7), F(24, Nil)),
			5:    Obj("[4]*T", F(0, Nil), F(8, 9), F(16, 8), F(24, Nil)),
			6:    Obj("T", F(0, 5)),
			7:    Obj("T", F(0, Nil)),
			8:    Obj("T", F(0, Nil)),
			9:    Obj("T", F(0, Nil)),
			10:   Obj("[4]*T", F(0, Nil), F(8, Nil), F(16, Nil), F(24, 3)),
			11:   Obj("[4]*T", F(0, Nil), F(8, Nil), F(16, 13), F(24, 12)),
			12:   Obj("T", F(0, Nil)),
			13:   Obj("T", F(0, 5)),
		},
		Blocks: []Block{
			Blk(0xa000, 16, 2, 7, Free, Free, 9, 8, 12),
			Blk(0xb000, 32, Free, 4, 5, Free),
			Blk(0xc000, 16, Free, Free, 6, 3, 13, Free, Free),
			Blk(0xd000, 32, Free, 10, Free, 11),
		},
	}

	ms := NewMarkSweep(roots, heap)
	i := 0
	for s := range ms.Evolve() {
		fname := fmt.Sprintf("./img/marksweep-%03d.png", i)
		fmt.Println("generating", fname)
		must(Draw(s).SavePNG(fname))
		i++
	}

	gt := NewGreenTea(roots, heap)
	i = 0
	for s := range gt.Evolve() {
		fname := fmt.Sprintf("./img/greentea-%03d.png", i)
		fmt.Println("generating", fname)
		must(Draw(s).SavePNG(fname))
		i++
	}
}

type gcState interface {
	Evolve() iter.Seq[gcState]
	Heap() *Heap
	Roots() ([]Root, int)
	Marked() *Set[Pointer]
	FieldsVisited() map[Pointer]int
	Context() Context
	DrawExtra(*gg.Context, image.Rectangle)
}

func Draw(s gcState) *gg.Context {
	c := gg.NewContext(1920, 1080)

	// Clear.
	c.SetRGB(1, 1, 1)
	c.DrawRectangle(0, 0, 1920, 1080)
	c.Fill()

	info := "type T struct{\n" +
		"\u2800   children *[4]*T\n" +
		"\u2800   value    int\n" +
		"}"

	area := drawObjGraph(c, info, s)
	s.DrawExtra(c, area)
	return c
}

func (m *MarkSweep) DrawExtra(c *gg.Context, area image.Rectangle) {
	const padding = 32
	c.SetColor(color.Black)
	must(setFontFace(c, "./RobotoMono-Regular.ttf", 36))

	c.DrawStringAnchored("Stack", float64(area.Min.X)+padding, float64(area.Max.Y)-padding, 0, 0)
	ly := float64(area.Max.Y) - padding - 48
	c.SetLineWidth(5.0)
	c.DrawLine(float64(area.Min.X)+padding, ly, float64(area.Max.X)-padding, ly)
	c.Stroke()

	must(setFontFace(c, "./RobotoMono-Regular.ttf", 32))
	ly -= 16
	for _, p := range m.stack {
		c.DrawStringAnchored(fmt.Sprintf("0x%x", m.heap.AddressOf(p)), float64(area.Min.X)+padding, ly, 0, 0)
		ly -= 40
	}
}

func (g *GreenTea) DrawExtra(c *gg.Context, area image.Rectangle) {
	const padding = 32
	c.SetColor(color.Black)
	must(setFontFace(c, "./RobotoMono-Regular.ttf", 36))

	c.DrawStringAnchored("Queue", float64(area.Min.X)+padding, float64(area.Max.Y)-padding, 0, 0)
	ly := float64(area.Max.Y) - padding - 48
	c.SetLineWidth(5.0)
	c.DrawLine(float64(area.Min.X)+padding, ly, float64(area.Max.X)-padding, ly)
	c.Stroke()

	must(setFontFace(c, "./RobotoMono-Regular.ttf", 32))
	ly -= 16
	for b := range g.queue.All() {
		c.DrawStringAnchored(fmt.Sprintf("0x%x", b.Address), float64(area.Min.X)+padding, ly, 0, 0)
		ly -= 40
	}
}

func drawObjGraph(c *gg.Context, info string, s gcState) image.Rectangle {
	faded := color.Gray{Y: 153}
	highlight := color.RGBA{R: 0xff, G: 0, B: 0, A: 255}

	roots, rootsVisited := s.Roots()
	h := s.Heap()
	marked := s.Marked()
	fieldsVisited := s.FieldsVisited()
	ctx := s.Context()

	split := c.Width() / 4
	infoArea := image.Rect(0, 0, split, 192)
	rootsArea := image.Rect(0, 192, split, 192+c.Height()/4)
	heapArea := image.Rect(split, 0, c.Width(), c.Height())
	restArea := image.Rect(0, rootsArea.Max.Y, split, c.Height())

	c.SetColor(color.Black)
	must(setFontFace(c, "./RobotoMono-Regular.ttf", 26))

	c.SetLineCapButt()
	c.SetLineJoin(gg.LineJoinRound)
	c.SetLineWidth(4.0)

	c.DrawRectangle(float64(infoArea.Min.X+16), float64(infoArea.Min.Y+16), float64(infoArea.Dx()-32), float64(infoArea.Dy()-32))
	c.Stroke()

	c.DrawStringWrapped(info, float64(infoArea.Min.X+32), float64(infoArea.Min.Y+32), 0, 0, float64(infoArea.Dx()-64), 1.25, gg.AlignLeft)

	must(setFontFace(c, "./RobotoMono-Regular.ttf", 36))

	var rootAnchors []image.Point
	for i := range roots {
		const padding = 16
		const baseHeight = 36

		r := &roots[i]
		if ctx.Root >= 0 && i == ctx.Root {
			c.SetColor(highlight)
		} else if i < rootsVisited {
			c.SetColor(color.Black)
		} else {
			c.SetColor(faded)
		}

		inc := rootsArea.Dy() / (len(roots) + 1)
		anchor := image.Pt(rootsArea.Min.X+rootsArea.Dx()*3/4, rootsArea.Min.Y+inc*(i+1))
		c.DrawStringAnchored(r.Name, float64(rootsArea.Min.X)+padding*2, float64(anchor.Y), 0, 0.5)
		c.DrawRectangle(float64(rootsArea.Min.X)+padding, float64(anchor.Y-baseHeight/2-padding), float64(anchor.X-rootsArea.Min.X), baseHeight+2*padding)
		c.Stroke()

		anchor.X += padding
		rootAnchors = append(rootAnchors, anchor)
	}

	c.SetColor(color.Black)
	must(setFontFace(c, "./RobotoMono-Regular.ttf", 28))

	const ptrWordSize = 64

	const blockColumns = 1
	const blockHeight = 128

	blockWidth := float64(heapArea.Dx()/blockColumns) * 0.8
	blockRows := (len(h.Blocks) + blockColumns - 1) / blockColumns
	blockColInc := float64(heapArea.Dx() / blockColumns)
	blockRowInc := float64(heapArea.Dy() / (blockRows + 1))

	// Draw boxes.
	objBoxes := make(map[Pointer]image.Rectangle)
	for i := range h.Blocks {
		b := &h.Blocks[i]
		col := i % blockColumns
		row := (i / blockColumns) + 1
		cx, cy := float64(heapArea.Min.X)+blockColInc/2+float64(col)*blockColInc, float64(heapArea.Min.Y)+float64(row)*blockRowInc

		must(setFontFace(c, "./RobotoMono-Regular.ttf", 28))

		bx := cx - blockWidth/2
		by := cy - blockHeight/2

		if ctx.Block == b {
			c.SetColor(highlight)
			c.SetDash()
			c.SetLineWidth(3.0)
		} else {
			c.SetColor(color.Black)
			c.SetLineWidth(2.0)
		}
		c.SetDash(4.0)
		c.DrawRoundedRectangle(bx, by, blockWidth, blockHeight, 12.0)
		c.Stroke()
		c.DrawStringAnchored(fmt.Sprintf("0x%x", b.Address), bx, by-12, 0, 0)

		must(setFontFace(c, "./RobotoMono-Regular.ttf", 24))

		const objPadding = 16
		baseObjX := bx + objPadding
		for j, p := range b.Objects {
			obj := &h.Objects[p]

			ox := baseObjX
			oy := by + blockHeight - objPadding - ptrWordSize
			width := b.ElemSize / PointerSize * ptrWordSize
			baseObjX += float64(width + objPadding)

			// Draw object pointer fields.
			objBoxes[p] = image.Rect(int(ox), int(oy), int(ox)+width, int(oy+ptrWordSize))
			for k, f := range obj.Fields {
				fi := f.Offset / PointerSize

				if marked.Has(p) {
					c.SetColor(color.Black)
				} else {
					c.SetColor(faded)
				}

				c.SetDash()
				c.SetLineWidth(2.0)
				c.DrawRectangle(ox+float64(fi*ptrWordSize), oy, ptrWordSize, ptrWordSize)
				c.Stroke()

				if marked.Has(p) {
					if ctx.Object == p && ctx.Field >= 0 && ctx.Field == k {
						c.SetColor(highlight)
					} else {
						c.SetColor(color.Black)
					}
				} else {
					c.SetColor(faded)
				}

				cx := ox + float64(fi*ptrWordSize) + ptrWordSize/2
				cy := oy + ptrWordSize/2
				c.DrawCircle(cx, cy, ptrWordSize/6)
				c.Fill()
			}

			// Draw object boundary.
			if ctx.Object == p {
				c.SetColor(highlight)
			} else if marked.Has(p) {
				c.SetColor(color.Black)
			} else {
				c.SetColor(faded)
			}
			if obj.Type == "<free>" {
				c.SetDash(2.0)
			} else {
				c.SetDash()
				must(setFontFace(c, "./RobotoMono-Regular.ttf", 24))
				c.DrawStringAnchored(obj.Type, ox, oy-8, 0, 0)
				must(setFontFace(c, "./RobotoMono-Regular.ttf", 18))
				c.DrawStringAnchored(fmt.Sprintf("+0x%x", uint64(j*b.ElemSize)), ox+float64(width), oy-8, 1, 0)
			}

			c.SetLineWidth(4.0)
			c.DrawRectangle(ox, oy, float64(width), ptrWordSize)
			c.Stroke()
		}
	}

	// Draw arrows.
	c.SetColor(color.Black)
	for i := range roots {
		r := &roots[i]
		dstR, ok := objBoxes[r.Pointer]
		if !ok {
			continue
		}
		if ctx.Root >= 0 && i == ctx.Root {
			c.SetColor(highlight)
		} else if i < rootsVisited {
			c.SetColor(color.Black)
		} else {
			c.SetColor(faded)
		}
		src := rootAnchors[i]
		dst := minDistPtOnRect(src, dstR, ptrWordSize/3)

		drawArrow(c, float64(src.X), float64(src.Y), float64(dst.X), float64(dst.Y), 3.0)
	}
	for i := range h.Objects {
		p := Pointer(i)
		obj := &h.Objects[p]
		src := objBoxes[p]

		for i, f := range obj.Fields {
			fi := f.Offset / PointerSize
			dstR, ok := objBoxes[f.Pointer]
			if !ok {
				continue
			}

			if ctx.Object == p && ctx.Field >= 0 && ctx.Field == i {
				c.SetColor(highlight)
			} else if i < fieldsVisited[p] {
				c.SetColor(color.Black)
			} else {
				c.SetColor(faded)
			}

			src := image.Pt(src.Min.X+fi*ptrWordSize+ptrWordSize/2, src.Min.Y+ptrWordSize/2)
			dst := minDistPtOnRect(src, dstR, ptrWordSize/3)

			drawArrow(c, float64(src.X), float64(src.Y), float64(dst.X), float64(dst.Y), 3.0)
		}
	}
	return restArea
}

func drawArrow(c *gg.Context, srcX, srcY, dstX, dstY, width float64) {
	c.SetLineWidth(width)

	dist2 := (dstX-srcX)*(dstX-srcX) + (dstY-srcY)*(dstY-srcY)
	c.MoveTo(srcX, srcY)
	c.LineTo(dstX, dstY)
	c.Stroke()

	const alBase = 7
	const th = math.Pi / 8
	al := alBase * width
	vx := (srcX - dstX) / math.Sqrt(dist2) * al
	vy := (srcY - dstY) / math.Sqrt(dist2) * al
	vx1 := vx*math.Cos(th) - vy*math.Sin(th)
	vy1 := vx*math.Sin(th) + vy*math.Cos(th)
	vx2 := vx*math.Cos(-th) - vy*math.Sin(-th)
	vy2 := vx*math.Sin(-th) + vy*math.Cos(-th)

	ah1X := vx1 + float64(dstX)
	ah1Y := vy1 + float64(dstY)
	ah2X := vx2 + float64(dstX)
	ah2Y := vy2 + float64(dstY)
	c.MoveTo(float64(dstX), float64(dstY))
	c.LineTo(ah1X, ah1Y)
	c.LineTo(ah2X, ah2Y)
	c.LineTo(float64(dstX), float64(dstY))
	c.Fill()
}

func minDistPtOnRect(src image.Point, rect image.Rectangle, div int) image.Point {
	minDist2 := -1
	var dstX, dstY int
	for d := range rectAnchors(rect, div) {
		dx := d.X - src.X
		dy := d.Y - src.Y
		if dist2 := dx*dx + dy*dy; minDist2 < 0 || dist2 < minDist2 {
			minDist2 = dist2
			dstX = d.X
			dstY = d.Y
		}
	}
	return image.Pt(dstX, dstY)
}

func rectAnchors(rect image.Rectangle, div int) iter.Seq[image.Point] {
	return func(yield func(image.Point) bool) {
		for x := rect.Min.X + div/2; x <= rect.Max.X-div/2; x += div {
			if !yield(image.Pt(x, rect.Min.Y)) {
				return
			}
		}
		for y := rect.Min.Y + div/2; y <= rect.Max.Y-div/2; y += div {
			if !yield(image.Pt(rect.Min.X, y)) {
				return
			}
			if !yield(image.Pt(rect.Max.X, y)) {
				return
			}
		}
		for x := rect.Min.X + div/2; x <= rect.Max.X-div/2; x += div {
			if !yield(image.Pt(x, rect.Max.Y)) {
				return
			}
		}
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type fontFaceKey struct {
	path string
	size float64
}

var fontCache = make(map[string]*truetype.Font)
var faceCache = make(map[fontFaceKey]font.Face)

func setFontFace(c *gg.Context, path string, size float64) error {
	if f, ok := faceCache[fontFaceKey{path, size}]; ok {
		c.SetFontFace(f)
		return nil
	}
	if ft, ok := fontCache[path]; ok {
		f := truetype.NewFace(ft, &truetype.Options{Size: size})
		faceCache[fontFaceKey{path, size}] = f
		c.SetFontFace(f)
		return nil
	}
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ft, err := truetype.Parse(fontBytes)
	if err != nil {
		return err
	}
	fontCache[path] = ft
	f := truetype.NewFace(ft, &truetype.Options{Size: size})
	faceCache[fontFaceKey{path, size}] = f
	c.SetFontFace(f)
	return nil
}
