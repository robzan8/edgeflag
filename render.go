package edgeflag

import (
	"image"
	"math"
)

type line struct {
	P, Q fixPoint
}

type Path []Point

func Render(dst Mask, r image.Rectangle, src Path, sp Point) {
	r = r.Intersect(dst.Bounds())
	fr := fixRectOf(r)
	l := line{Q: fixPointOf(src[0])}
	for i := range src {
		if math.IsNaN(src[i].X) { // a NaN Point marks the next point to be a goto
			i++
			if i < len(src) {
				l.Q = fixPointOf(src[i])
			}
		} else {
			l.P = l.Q
			l.Q = fixPointOf(src[i])
			// clipping

			scanLine(&dst, &fr, &l)
		}
	}
	fill(&dst, &r)
}

// fixed point .8 bit signed numbers stored in int
type fix int

const one = 1 << 16

type fixPoint struct {
	x, y fix
}

func fixPointOf(p Point) (fp fixPoint) {
	fp.x = fix(p.X * one)
	fp.y = fix(p.Y * one)
	return
}

type fixRect struct {
	min, max fixPoint
}

func fixRectOf(r image.Rectangle) (fr fixRect) {
	fr.min.x = fix(r.Min.X) << 16
	fr.min.y = fix(r.Min.Y) << 16
	fr.max.x = fix(r.Max.X) << 16
	fr.max.y = fix(r.Max.Y) << 16
	return
}

var sampleX [8]fix = [8]fix{(0.0/8 + 1.0/16) * one, (6.0/8 + 1.0/16) * one, (3.0/8 + 1.0/16) * one, // fix this 1 - sample
	(5.0/8 + 1.0/16) * one, (1.0/8 + 1.0/16) * one, (7.0/8 + 1.0/16) * one, (2.0/8 + 1.0/16) * one, (4.0/8 + 1.0/16) * one}

func scanLine(dst *Mask, fr *fixRect, l *line) {
	p := fixPointOf(l.P)
	r := fixPointOf(l.R)

	var minY, maxY fix
	if p.y < r.y {
		minY, maxY = p.y, r.y
	} else {
		minY, maxY = r.y, p.y
	}
	if minY < fr.min.y {
		minY = fr.min.y
	}
	if maxY > fr.max.y {
		maxY = fr.max.y
	}

	var m float64 = (l.R.X - l.P.X) / (l.R.Y - l.P.Y)
	var firstScan fix = (minY+0xfff)&^0xfff | 0x1000
	for y := firstScan; y < maxY; y += 0x2000 {
		var x fix = p.x + fix(float64(y-p.y)*m)
		sampleI := uint((y >> 13) & 7)
		pixelX := int((x + sampleX[sampleI]) >> 16)
		if pixelX < dst.stride {
			if pixelX < 0 {
				pixelX = 0
			}
			i := int(y>>19)*dst.stride + pixelX
			flag := samples64(1<<63) >> (uint(y>>13)&(7<<3) + sampleI)
			dst.data[i] ^= flag
		}
	}
}

// bitwise tricks so that fills in bounds even though using samples64
func fill(dst *Mask, r *image.Rectangle) {
	for y := r.Min.Y; y < dst.stride8y; y++ { // terrible temporary hack
		var filling samples64 = 0
		for x := r.Min.X; x < r.Max.X; x++ {
			i := y*dst.stride + x
			edge := dst.data[i]
			dst.data[i] ^= filling
			filling ^= edge
		}
	}
}
