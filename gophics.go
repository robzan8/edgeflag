package edgeflag

import (
	"math"
)

type PtType int

const (
	Move PtType = iota
	Line
)

type PathPt struct {
	Point
	Type PtType
}

type Path struct {
	Pts []Point
}

func NewPath() *Path {
	return &Path{make([]Point, 0, 8)}
}

func (p *Path) Move(to Point) {
	p.Pts = append(p.Pts, Point{X: math.NaN()}, to)
}

func (p *Path) Line(to Point) {
	p.Pts = append(p.Pts, to)
}

func (p *Path) Quad(c, to Point) {
	// quadratic bezier segmentation
}

func (p *Path) Cubic(c0, c1, to Point) {
	// cubic bezier segmentation
}

func (p *Path) Arc( /* whatever */ ) {
	// arc segmentation
}
