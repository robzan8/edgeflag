package edgeflag

import (
	"image"
	"image/color"
)

type Mask struct {
	runs       []uint64
	Rect       image.Rectangle
	cols, rows int
}

func NewMask(r image.Rectangle) *Mask {
	cols := r.Dx()
	rows := (r.Dy() + 15) / 16 // ceil(r.Dy()/16)
	runs := make([]uint64, cols*rows)
	return &Mask{runs, r, cols, rows}
}

func (m *Mask) ColorModel() color.Model {
	return color.Alpha16Model
}

func (m *Mask) Bounds() image.Rectangle {
	return m.Rect
}

// if Draw with At().RGBA() is really slow, consider a function
// that converts a mask to an alpha image

func (m *Mask) At(x, y int) color.Color {
	p := image.Point{x, y}
	if !p.In(m.Rect) {
		return color.Alpha16{}
	}
	return pixToAlpha(pixAt(m, &p))
}

func pixAt(m *Mask, p *image.Point) uint64 {
	r := p.Sub(m.Rect.Min)
	run := m.runs[(r.Y/16)*m.Rect.Dx()+r.X]
	i := uint64(r.Y) % 16
	return run >> ((15 - i) * 4) & 0xf
}

// color correction here
func pixToAlpha(pix uint64) color.Alpha16 {
	var ones uint64 = 0
	for i := 0; i < 4; i++ {
		ones += pix & 1
		pix >>= 1
	}
	a := uint16(float64(ones)/4*0xffff + 0.5)
	return color.Alpha16{A: a}
}
