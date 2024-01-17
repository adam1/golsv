package golsv

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

const radiusDelta = 10.0

type CalGVisualizer struct {
	outputFile string
	pvertices []pvertex // this order is referenced by the "faces" in the OFF format
	elementMap map[ElementCalG]int // map from element to index in pvertices
	edges []ZEdge[ElementCalG]
	triangles []ZTriangle[ElementCalG]
	faces [][3]int // indices into pvertices
}

type pvertex struct {
	u ElementCalG
	depth int
	pos point
}

type point struct {
	x, y, z float64
}

func NewCalGVisualizer(outputFile string) *CalGVisualizer {
	return &CalGVisualizer{
		outputFile: outputFile,
		pvertices: make([]pvertex, 0),
		elementMap: make(map[ElementCalG]int),
		faces: make([][3]int, 0),
	}
}

func (V *CalGVisualizer) BeginVertices() {
}

func (V *CalGVisualizer) computeFaces() {
	skip := 0
	for _, t := range V.triangles {
		a, ok := V.elementMap[t[0].(ElementCalG)]
		if !ok {
			skip++
			continue
		}
		b, ok := V.elementMap[t[1].(ElementCalG)]
		if !ok {
			skip++
			continue
		}
		c, ok := V.elementMap[t[2].(ElementCalG)]
		if !ok {
			skip++
			continue
		}
		V.faces = append(V.faces, [3]int{a, b, c})
	}
	if skip > 0 {
		log.Printf("warning: skipped %d triangles", skip)
	}
}

func (V *CalGVisualizer) Edges(edges []ZEdge[ElementCalG]) {
	V.edges = edges
}

func (V *CalGVisualizer) End() {
	log.Printf("computing vertex positions")
	V.positionVertices()
	log.Printf("computing faces")
	V.computeFaces()
	log.Printf("writing output file %s", V.outputFile)
	V.writeOutput()
	log.Printf("done")
}

func (V *CalGVisualizer) EndVertices() {
}

// first stab at layout: arrange on a circle in arrival order
func (V *CalGVisualizer) positionRadialFlat() {
	curDepth := 0
	start := 0
	for i := range V.pvertices {
		p := V.pvertices[i]
		if p.depth > curDepth {
			//V.positionRadialFlatDepth(start, i, curDepth)
			V.positionRadialFlatClusteredDepth(start, i, curDepth)
			curDepth = p.depth
			start = i
		}
	}
	V.positionRadialFlatDepth(start, len(V.pvertices), curDepth)
}

// xxx wip; dont think this is working
func (V *CalGVisualizer) positionRadialFlatClusteredDepth(start, end, depth int) {
	log.Printf("xxx depth=%d start=%d end=%d", depth, start, end)

// 	// xxx for each position, attempt to choose a vertex that is connected
// 	// to the previous vertex.
// 	n := end - start
// 	r := radiusDelta * float64(depth)
// 	thetaDelta := 2.0 * math.Pi / float64(n)
// 	var prev int
// 	for i := start; i < end; i++ {
// 		theta := float64(i) * thetaDelta
// 		x := r * math.Cos(theta)
// 		y := r * math.Sin(theta)
// 		z := 0.0

// 		for j := i; j < n; j++ {
// 			p := &V.pvertices[i]
// 			if i == start {
// 				// first vertex in this depth
// 				p.pos.x = x
// 				p.pos.y = y
// 				p.pos.z = z
// 				prev = j
// 				break
// 			}
// 			if p.pos.x > 0 {
// 				// already positioned
// 				continue
// 			}
// 			for _, e := range V.edges {
// 				if e.Contains(prev) && e.Contains(j) {
// 					p.pos.x = x
// 					p.pos.y = y
// 					p.pos.z = z
// 					prev = j
// 					break
// 				}
// 			}
// 			// xxx if we still haven't chosen, choose the next one
			
// 		}
// 		prev = p
// 	}
}

func (V *CalGVisualizer) positionRadialFlatDepth(start, end, depth int) {
	n := end - start
	r := radiusDelta * float64(depth)
	thetaDelta := 2.0 * math.Pi / float64(n)
	for i := start; i < end; i++ {
		p := &V.pvertices[i]
		theta := float64(i) * thetaDelta
		p.pos.x = r * math.Cos(theta)
		p.pos.y = r * math.Sin(theta)
		p.pos.z = 0.0
	}
}

func (V *CalGVisualizer) positionVertices() {
	V.positionRadialFlat()
}

func (V *CalGVisualizer) Triangles(triangles []ZTriangle[ElementCalG]) {
	log.Printf("xxx visualizer received %d triangles", len(triangles))
	V.triangles = triangles
}

func (V *CalGVisualizer) Vertex(u ElementCalG, id int, depth int) {
	V.pvertices = append(V.pvertices, pvertex{u, depth, point{-1, -1, -1}})
	V.elementMap[u] = len(V.pvertices)
}

func (V *CalGVisualizer) writeOutput() {
	// write OFF (Object File Format) file
	// https://en.wikipedia.org/wiki/OFF_(file_format)

	// header
	// OFF
	// nVertices nFaces nEdges
	// x0 y0 z0
	// x1 y1 z1
	// ...
	// xN yN zN
	// nVertices v1 v2 v3
	// ...
	// nVertices v1 v2 v3

	f, err := os.Create(V.outputFile)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()
	_, err = f.WriteString("OFF\n")
	if err != nil {
		panic(err)
	}

	// nVertices nFaces nEdges
	nVertices := len(V.pvertices)
	nFaces := len(V.faces)
	nEdges := 0 // this can be zero to let the viewer compute it
	_, err = fmt.Fprintf(f, "%d %d %d\n", nVertices, nFaces, nEdges)
	if err != nil {
		panic(err)
	}

	var b strings.Builder
	for _, p := range V.pvertices {
		fmt.Fprintf(&b, "%f %f %f\n", p.pos.x, p.pos.y, p.pos.z)
	}
	_, err = f.WriteString(b.String())
	if err != nil {
		panic(err)
	}

	b.Reset()
	for _, face := range V.faces {
		fmt.Fprintf(&b, "3 %d %d %d\n", face[0], face[1], face[2])
	}
	_, err = f.WriteString(b.String())
	if err != nil {
		panic(err)
	}
}
