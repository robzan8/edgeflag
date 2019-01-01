package edgeflag

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"
)

type evaluator struct {
	a, b, c Point
}

func evaluatorOf(c *Curve) (e evaluator) {
	e.a.X = c.P.X - 2.0*c.Q.X + c.R.X
	e.a.Y = c.P.Y - 2.0*c.Q.Y + c.R.Y
	e.b.X = 2.0 * (c.Q.X - c.P.X)
	e.b.Y = 2.0 * (c.Q.Y - c.P.Y)
	e.c = c.P
	return
}

func (e *evaluator) evalAt(t float64) (p Point) {
	p.X = t*(t*e.a.X+e.b.X) + e.c.X
	p.Y = t*(t*e.a.Y+e.b.Y) + e.c.Y
	return
}

func mid(p, q Point) Point {
	return Point{(p.X + q.X) / 2, (p.Y + q.Y) / 2}
}

func mid34(p, q Point) Point {
	return Point{(p.X + 3.0*q.X) / 4, (p.Y + 3.0*q.Y) / 4}
}

// resolution indipendent. If scene scales, tessellation doesn't change
func cubicToQuads(path Path, a, b, c, d Point) Path {
	ab := mid(a, b)
	bc := mid(b, c)
	cd := mid(c, d)
	anch1 := mid34(a, ab)
	anch4 := mid34(d, cd)
	e := mid(ab, bc)
	f := mid(bc, cd)
	ctrl3 := mid(e, f)
	anch2 := mid34(ctrl3, e)
	anch3 := mid34(ctrl3, f)
	ctrl2 := mid(anch1, anch2)
	ctrl4 := mid(anch3, anch4)
	return append(path, Curve{a, anch1, ctrl2}, Curve{ctrl2, anch2, ctrl3},
		Curve{ctrl3, anch3, ctrl4}, Curve{ctrl4, anch4, d})
}

var imgWidth int = 1280
var imgHeight int = 720
var bounds image.Rectangle = image.Rectangle{Max: image.Point{imgWidth, imgHeight}}

var path Path
var benchMsk Mask = MakeMask(bounds)

// initialize path from gophers.json
func init() {
	gophers, err := os.Open("img/gophers.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer gophers.Close()

	decoder := json.NewDecoder(gophers)
	var points []Point
	err = decoder.Decode(&points)
	if err != nil {
		fmt.Println(err)
		return
	}

	path = make([]Curve, 0, 10)
	for i := 3; i < len(points); i += 4 {
		a := points[i-3]
		b := points[i-2]
		c := points[i-1]
		d := points[i]
		a.Y = -a.Y + float64(imgHeight)
		b.Y = -b.Y + float64(imgHeight)
		c.Y = -c.Y + float64(imgHeight)
		d.Y = -d.Y + float64(imgHeight)
		path = cubicToQuads(path, a, b, c, d)
	}
}

func TestRender(t *testing.T) {
	testMsk := MakeMask(bounds)
	Render(testMsk, bounds, path, ZP) //, Rectangle{Max: Point{float64(imgWidth), float64(imgHeight)}})

	img := image.NewRGBA(bounds)

	gray := &image.Uniform{color.RGBA{60, 60, 60, 255}}
	draw.DrawMask(img, bounds, gray, image.ZP, testMsk, image.ZP, draw.Src)

	pngFile, err := os.Create("img/gophers.png")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pngFile.Close()

	err = png.Encode(pngFile, img)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// TestPerf is for benchmarking and profiling the renderer.
// A path is rendered (many times) on a 720p image (render_perf_test.png).
// There are a few big curves/segments and many small ones,
// as it is expected in typical real applications.
// The benchmark always runs for 30 seconds.
func BenchmarkRender(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Render(benchMsk, bounds, path, ZP) //, Rectangle{Max: Point{float64(imgWidth), float64(imgHeight)}})
	}
}
