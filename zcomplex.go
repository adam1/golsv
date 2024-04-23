package golsv

import (
	"bytes"
	"fmt"
	"log"
	"sort"
)
// this file models a 2-d simplicial complex.

// xxx the following is derived from complex.go; attempting to factor
// out the common part.

type ZVertex[T any] interface {
	Equal(T) bool
	Less(T) bool
	String() string
}

type ZEdge[T any] [2]ZVertex[T]

// NOTE: it's important to always use NewZEdge to create an ZEdge,
// because it ensures that the vertices are in the correct order.
// this canonical ordering allows edges to be used as map keys.
func NewZEdge[T any](a, b ZVertex[T]) ZEdge[T] {
	if a.Less(b.(T)) {
		return ZEdge[T]{a, b}
	} else {
		return ZEdge[T]{b, a}
	}
}

// xxx test
func (E ZEdge[T]) Contains(v ZVertex[T]) bool {
	return E[0] == v || E[1] == v
}

// xxx test
func (E ZEdge[T]) Equal(F ZEdge[T]) bool {
	return E[0].Equal(F[0].(T)) && E[1].Equal(F[1].(T))
}

func (E ZEdge[T]) Less(F ZEdge[T]) bool {
	if E[0].Less(F[0].(T)) {
		return true
	} else if F[0].Less(E[0].(T)) {
		return false
	} else {
		return E[1].Less(F[1].(T))
	}
}

func (E ZEdge[T]) String() string {
	return fmt.Sprintf("[%s %s]", E[0], E[1])
}

type ZTriangle[T any] [3]ZVertex[T]

func NewZTriangle[T any](a, b, c ZVertex[T]) ZTriangle[T] {
	if a.Less(b.(T)) {
		if b.Less(c.(T)) {
			return ZTriangle[T]{a, b, c}
		} else if a.Less(c.(T)) {
			return ZTriangle[T]{a, c, b}
		} else {
			return ZTriangle[T]{c, a, b}
		}
	} else {
		if a.Less(c.(T)) {
			return ZTriangle[T]{b, a, c}
		} else if b.Less(c.(T)) {
			return ZTriangle[T]{b, c, a}
		} else {
			return ZTriangle[T]{c, b, a}
		}
	}
}

func (T ZTriangle[VertexType]) ContainsVertex(v ZVertex[VertexType]) bool {
	return T[0] == v || T[1] == v || T[2] == v
}

func (T ZTriangle[VertexType]) Edges() [3]ZEdge[VertexType] {
	return [3]ZEdge[VertexType]{NewZEdge(T[0], T[1]), NewZEdge(T[1], T[2]), NewZEdge(T[2], T[0])}
}

func (T ZTriangle[VertexType]) Equal(U ZTriangle[VertexType]) bool {
	return T[0].Equal(U[0].(VertexType)) && T[1].Equal(U[1].(VertexType)) && T[2].Equal(U[2].(VertexType))
}

func (T ZTriangle[VertexType]) Less(U ZTriangle[VertexType]) bool {
	if T[0].Less(U[0].(VertexType)) {
		return true
	} else if U[0].Less(T[0].(VertexType)) {
		return false
	} else if T[1].Less(U[1].(VertexType)) {
		return true
	} else if U[1].Less(T[1].(VertexType)) {
		return false
	} else {
		return T[2].Less(U[2].(VertexType))
	}
}

func (T ZTriangle[VertexType]) String() string {
	return fmt.Sprintf("[%s %s %s]", T[0], T[1], T[2])
}

func TriangleSetEqual[VertexType any](S, U []ZTriangle[VertexType]) bool {
	if len(S) != len(U) {
		return false
	}
	sort.Slice(S, func(i, j int) bool {
		return S[i].Less(S[j])
	})
	sort.Slice(U, func(i, j int) bool {
		return U[i].Less(U[j])
	})
	for i := range S {
		if !S[i].Equal(U[i]) {
			return false
		}
	}
	return true
}

type ZComplex[T any] struct {
    vertexBasis   []ZVertex[T]
	edgeBasis     []ZEdge[T]
	triangleBasis []ZTriangle[T]
	d1            BinaryMatrix
	d2 		      BinaryMatrix
	EdgeIndex     map[ZEdge[T]]int   // edge -> index in basis
	VertexIndex   map[ZVertex[T]]int // vertex -> index in basis
	verbose       bool
}

func NewZComplex[T any](vertexBasis []ZVertex[T], edgeBasis []ZEdge[T], triangleBasis []ZTriangle[T], sortBases bool, verbose bool) *ZComplex[T] {
	if sortBases {
		sort.Slice(triangleBasis, func(i, j int) bool {
			return triangleBasis[i].Less(triangleBasis[j])
		})
		sort.Slice(edgeBasis, func(i, j int) bool {
			return edgeBasis[i].Less(edgeBasis[j])
		})
		sort.Slice(vertexBasis, func(i, j int) bool {
			return vertexBasis[i].Less(vertexBasis[j].(T))
		})
	}
	C := ZComplex[T]{
		vertexBasis: vertexBasis,
		edgeBasis: edgeBasis,
		triangleBasis: triangleBasis,
		verbose: verbose,
	}
	C.computeVertexIndex()
	C.computeEdgeIndex()
	if verbose {
		log.Printf("Complex: %v", &C)
	}
	// log.Printf("xxx\n%s", C.DumpBases())
	return &C
}

func (C *ZComplex[T]) computeEdgeIndex() {
	C.EdgeIndex = make(map[ZEdge[T]]int)
	for i, edge := range C.edgeBasis {
		C.EdgeIndex[edge] = i
	}
}

func (C *ZComplex[T]) computeVertexIndex() {
	C.VertexIndex = make(map[ZVertex[T]]int)
	for i, vertex := range C.vertexBasis {
		C.VertexIndex[vertex] = i
	}
}

func (C *ZComplex[T]) computeD1() {
	C.d1 = NewSparseBinaryMatrix(len(C.vertexBasis), len(C.edgeBasis))
	for j, edge := range C.edgeBasis {
		for _, v := range edge {
			row, ok := C.VertexIndex[v]
			if !ok {
				panic(fmt.Sprintf("vertex %v not in vertex index", v))
			}
			C.d1.Set(row, j, 1)
		}
	}
}

func (C *ZComplex[T]) computeD2() {
	C.d2 = NewSparseBinaryMatrix(len(C.edgeBasis), len(C.triangleBasis))
	for j, t := range C.triangleBasis {
		for _, e := range t.Edges() {
			row, ok := C.EdgeIndex[e]
			if !ok {
				panic(fmt.Sprintf("edge %v not in edge index", e))
			}
			C.d2.Set(row, j, 1)
		}
	}
}

func (C *ZComplex[T]) D1() BinaryMatrix {
	if C.d1 == nil {
		C.computeD1()
	}
	return C.d1
}

func (C *ZComplex[T]) D2() BinaryMatrix {
	if C.d2 == nil {
		C.computeD2()
	}
	return C.d2
}

func (C *ZComplex[T]) EdgeBasis() []ZEdge[T] {
	return C.edgeBasis
}

func (C *ZComplex[T]) EdgesContainingVertex(v ZVertex[T]) []ZEdge[T] {
	var edges []ZEdge[T]
	for _, t := range C.edgeBasis {
		if t.Contains(v) {
			edges = append(edges, t)
		}
	}
	return edges
}

func (C *ZComplex[T]) TrianglesContainingVertex(v ZVertex[T]) []ZTriangle[T] {
	var triangles []ZTriangle[T]
	for _, t := range C.triangleBasis {
		if t.ContainsVertex(v) {
			triangles = append(triangles, t)
		}
	}
	return triangles
}

// xxx test
func (C *ZComplex[T]) PathToEdgeVector(path ZPath[T]) BinaryVector {
	v := NewBinaryVector(len(C.edgeBasis))
	for _, e := range path {
		if i, ok := C.EdgeIndex[e]; ok {
			v[i] = 1
		} else {
			panic(fmt.Sprintf("edge %v not in edge index", e))
		}
	}
	return v
}

func (C *ZComplex[T]) String() string {
	return fmt.Sprintf("complex with %v triangles, %v edges, %v vertices",
		len(C.triangleBasis), len(C.edgeBasis), len(C.vertexBasis))
}

func (C *ZComplex[T]) TriangleBasis() []ZTriangle[T] {
	return C.triangleBasis
}

func (C *ZComplex[T]) VertexBasis() []ZVertex[T] {
	return C.vertexBasis
}

func (C *ZComplex[T]) BFS(v ZVertex[T], f func(u ZVertex[T])) {
	visited := make(map[ZVertex[T]]struct{})
	queue := NewZVertexQueue[T]()
	queue.Enqueue(v)
	f(v)
	visited[v] = struct{}{}
	for {
		u, ok := queue.Dequeue()
		if !ok {
			break
		}
		for _, e := range C.EdgesContainingVertex(u) {
			for _, w := range e {
				if _, ok := visited[w]; !ok {
					queue.Enqueue(w)
					f(w)
					visited[w] = struct{}{}
				}
			}
		}
	}
}

type ZVertexQueue[T any] struct {
	slice []ZVertex[T]
}

func NewZVertexQueue[T any]() *ZVertexQueue[T] {
	return &ZVertexQueue[T]{
		slice: make([]ZVertex[T], 0),
	}
}

func (Q *ZVertexQueue[T]) Enqueue(t ZVertex[T]) {
	Q.slice = append(Q.slice, t)
}

func (Q *ZVertexQueue[T]) Dequeue() (ZVertex[T], bool) {
	if len(Q.slice) == 0 {
		return nil, false
	}
	t := Q.slice[0]
	Q.slice = Q.slice[1:]
	return t, true
}

func (Q *ZVertexQueue[T]) Len() int {
	return len(Q.slice)
}

func (C *ZComplex[T]) BFTriangleWalk(v ZVertex[T], f func(t ZTriangle[T])) {
	tVisited := make(map[ZTriangle[T]]struct{})
	C.BFS(v, func(u ZVertex[T]) {
		for _, t := range C.TrianglesContainingVertex(u) {
			if _, ok := tVisited[t]; !ok {
				f(t)
				tVisited[t] = struct{}{}
			}
		}
	})
}

type ZTriangleQueue[T any] struct {
	slice []ZTriangle[T]
}

func NewZTriangleQueue[T any]() *ZTriangleQueue[T] {
	return &ZTriangleQueue[T]{
		slice: make([]ZTriangle[T], 0),
	}
}

func (Q *ZTriangleQueue[T]) Enqueue(t ZTriangle[T]) {
	Q.slice = append(Q.slice, t)
}

func (Q *ZTriangleQueue[T]) Dequeue() (ZTriangle[T], bool) {
	if len(Q.slice) == 0 {
		return ZTriangle[T]{}, false
	}
	t := Q.slice[0]
	Q.slice = Q.slice[1:]
	return t, true
}

func (Q *ZTriangleQueue[T]) Len() int {
	return len(Q.slice)
}

// returns a minimum weight edge cochain that is a cocycle but not a
// coboundary, i.e. a cochain whose weight is equal to S^1(C).  if B^1
// = Z^1, or if there are no edges, return an empty cochain.
//
// xxx experimental WIP!  incomplete.
func (C *ZComplex[T]) CosystolicVector() ZPath[T] {
	// first, prepare a sorted list of triangles (xxx is this
	// necessary?  could just use the triangleBasis slice.  is this
	// list topologically sorted?)
	var triangles []ZTriangle[T]
	C.BFTriangleWalk(C.vertexBasis[0], func(t ZTriangle[T]) {
		triangles = append(triangles, t)
	})
	// we want to enumerate Z^1 \setminus B^1, i.e. edge sets that
	// correspond to cocycles but not coboundaries, i.e. cochains that
	// vanish on B_1 but not on Z_1.  we will do this by choosing an
	// edge set from each triangle, such that the edge set vanishes on
	// the boundary of the triangle, which happens precisely when the
	// edge set contains an even number of edges in the triangle,
	// i.e. zero or two edges.  since there are in general multiple
	// choices for each triangle, we build a tree of choices.  there
	// is one level in the tree for each triangle, plus the root.  the
	// levels in the tree correspond to the index in the triangles
	// slice.  each vertex in the tree corresponds to a choice of edge
	// set for the triangle.  an edge in the tree from a vertex down
	// to a child vertex corresponds to a choice of edge set for the
	// child triangle that is compatible with the choices made for the
	// ancestor triangles in the tree (which in general may have
	// overlapping edges).  in this way, we are effectively coloring
	// every edge in the complex with one of three colors:
	//
	//   -1 = OFF: the edge IS NOT in the edge set
	//    0 = OPEN: it is undecided whether the edge is in the edge set
	//    1 = ON: the edge IS in the edge set
	//
	// as we build the tree, the leaf vertices correspond to possible
	// edge sets decided thus far.  when we reach the end of the
	// triangles list, all possible edge sets have been decided, and
	// each leaf vertex corresponds to a possible edge set.  we can
	// then test each edge set to see if it is a coboundary, i.e. to
	// see whether it does not vanish on Z_1. taking the minimum
	// weight edge set of those that are not coboundaries gives us a
	// cosystolic vector.
	//
	// this procedure is likely exponential in the number of edges in
	// the complex, although it depends on how the possibilities
	// branch.

	// special case: no triangles.  in this case, every cochain
	// vacuously vanishes on the boundary, hence is a cocycle.  to be
	// a coboundary of minimum weight then, we just need to choose a
	// single edge on a cycle, if there is such an edge.
	if len(triangles) == 0 {
		// xxx in order to test whether an edge is a coboundary
		// (below), we'll need to know generators for the cycles.
		// pass that in. we'll need to use it here and below.
		
		return ZPath[T]{}
	}
	root := newNode(nil)
	leaves := []*node{root}
	newLeaves := make([]*node, 0)
	
	for i, t := range triangles {
		fmt.Printf("%d: t=%v\n", i, t)
		// for this triangle, determine the branching at each current
		// leaf vertex.  the current leaf vertices represent the edge
		// set choices for the previous triangle.
		newLeaves = newLeaves[:0]
		for j, p := range leaves {
			fmt.Printf("  %d: p=%v\n", j, p)
			// we need to scan up the tree to find triangle nodes that
			// have settings for the edges in the current triangle t.
			c := C.scanUp(triangles, p, i - 1, t)

			switch c {
			// =================================================
			// num open is 0
			case [3]kEdgeColor{kEdgeOff, kEdgeOff, kEdgeOff}:
				fallthrough
			case [3]kEdgeColor{kEdgeOff, kEdgeOn, kEdgeOn}:
				fallthrough
			case [3]kEdgeColor{kEdgeOn, kEdgeOff, kEdgeOn}:
				fallthrough
			case [3]kEdgeColor{kEdgeOn, kEdgeOn, kEdgeOff}:
				// no-op; triangle is resolved
			// impossible; check for sanity
			case [3]kEdgeColor{kEdgeOff, kEdgeOff, kEdgeOn}:
				fallthrough
			case [3]kEdgeColor{kEdgeOff, kEdgeOn, kEdgeOff}:
				fallthrough
			case [3]kEdgeColor{kEdgeOn, kEdgeOff, kEdgeOff}:
				fallthrough
			case [3]kEdgeColor{kEdgeOn, kEdgeOn, kEdgeOn}:
				panic("contradiction")
			// =================================================
			// num open is 1
			case [3]kEdgeColor{kEdgeOff, kEdgeOff, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOff, kEdgeOpen, kEdgeOff}:
				q := p.addChild()
				q.setEdgeColor(1, kEdgeOff)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOpen, kEdgeOff, kEdgeOff}:
				q := p.addChild()
				q.setEdgeColor(0, kEdgeOff)
				newLeaves = append(newLeaves, q)
			// edge at index 2 open
			case [3]kEdgeColor{kEdgeOn, kEdgeOff, kEdgeOpen}:
				fallthrough
			case [3]kEdgeColor{kEdgeOff, kEdgeOn, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(2, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOn, kEdgeOn, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
			// edge at index 1 open
			case [3]kEdgeColor{kEdgeOn, kEdgeOpen, kEdgeOff}:
				fallthrough
			case [3]kEdgeColor{kEdgeOff, kEdgeOpen, kEdgeOn}:
				q := p.addChild()
				q.setEdgeColor(1, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOn, kEdgeOpen, kEdgeOn}:
				q := p.addChild()
				q.setEdgeColor(1, kEdgeOff)
				newLeaves = append(newLeaves, q)
			// edge at index 0 open
			case [3]kEdgeColor{kEdgeOpen, kEdgeOn, kEdgeOff}:
				fallthrough
			case [3]kEdgeColor{kEdgeOpen, kEdgeOff, kEdgeOn}:
				q := p.addChild()
				q.setEdgeColor(1, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOpen, kEdgeOn, kEdgeOn}:
				q := p.addChild()
				q.setEdgeColor(1, kEdgeOff)
				newLeaves = append(newLeaves, q)
			// num open is 2
			case [3]kEdgeColor{kEdgeOff, kEdgeOpen, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(1, kEdgeOff)
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(1, kEdgeOn)
				q.setEdgeColor(2, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOn, kEdgeOpen, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(1, kEdgeOn)
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(1, kEdgeOff)
				q.setEdgeColor(2, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOpen, kEdgeOff, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(0, kEdgeOff)
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(0, kEdgeOn)
				q.setEdgeColor(2, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOpen, kEdgeOn, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(0, kEdgeOn)
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(0, kEdgeOff)
				q.setEdgeColor(2, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOpen, kEdgeOpen, kEdgeOff}:
				q := p.addChild()
				q.setEdgeColor(0, kEdgeOff)
				q.setEdgeColor(1, kEdgeOff)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(0, kEdgeOn)
				q.setEdgeColor(1, kEdgeOn)
				newLeaves = append(newLeaves, q)
			case [3]kEdgeColor{kEdgeOpen, kEdgeOpen, kEdgeOn}:
				q := p.addChild()
				q.setEdgeColor(0, kEdgeOn)
				q.setEdgeColor(1, kEdgeOff)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(0, kEdgeOn)
				q.setEdgeColor(1, kEdgeOff)
				newLeaves = append(newLeaves, q)
			// num open is 3
			case [3]kEdgeColor{kEdgeOpen, kEdgeOpen, kEdgeOpen}:
				q := p.addChild()
				q.setEdgeColor(0, kEdgeOff)
				q.setEdgeColor(1, kEdgeOff)
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(0, kEdgeOff)
				q.setEdgeColor(1, kEdgeOn)
				q.setEdgeColor(2, kEdgeOn)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(0, kEdgeOn)
				q.setEdgeColor(1, kEdgeOff)
				q.setEdgeColor(2, kEdgeOn)
				newLeaves = append(newLeaves, q)
				q = p.addChild()
				q.setEdgeColor(0, kEdgeOn)
				q.setEdgeColor(1, kEdgeOn)
				q.setEdgeColor(2, kEdgeOff)
				newLeaves = append(newLeaves, q)
			}
		}
		leaves = newLeaves
	}
	log.Printf("leaves: %d\n", len(leaves))
	for i, p := range leaves {
		log.Printf("leaf %d: %v\n", i, p.edgeColors)
		// xxx convert to binary vector representing the bra.
		// xxx evaluate the bra on the basis vectors of Z_1.
	}
	// xxx wip
	return nil
}

func (C *ZComplex[T]) scanUp(triangles []ZTriangle[T], p *node, level int, t ZTriangle[T]) (colors [3]kEdgeColor) {
	tEdges := t.Edges()
	for {
		if p == nil || level == -1 {
			break
		}
		// let u be the triangle corresponding to the current node.
		// if u shares an edge with t, copy any colorings (kEdgeOn or
		// kEdgeOff, but not kEdgeOpen) to the result.
		uEdges := triangles[level].Edges()
		match := false
		for i, e := range tEdges {
			for j, f := range uEdges {
				if e.Equal(f) {
					if p.edgeColors[j] != kEdgeOpen {
						colors[i] = p.edgeColors[j]
					}
					match = true
					break
				}
			}
			if match {
				break
			}
		}
		if colors[0] != kEdgeOpen && colors[1] != kEdgeOpen && colors[2] != kEdgeOpen {
			break
		}
		p = p.parent
		level--
	}
	return
}

type kEdgeColor int8

const (
	kEdgeOff kEdgeColor = -1
	kEdgeOpen kEdgeColor = 0
	kEdgeOn kEdgeColor = 1
)

type node struct {
	edgeColors [3]kEdgeColor
	parent *node
	children []*node
}

func newNode(parent *node) *node {
	return &node{
		parent: parent,
		children: make([]*node, 0),
	}
}

func (p *node) addChild() *node {
	q := newNode(p)
	p.children = append(p.children, q)
	return q
}

func (p *node) setEdgeColor(index int, c kEdgeColor) {
	p.edgeColors[index] = c
}

type ZPath[T any] []ZEdge[T]

// xxx test
func (p ZPath[T]) String() string {
	var buf bytes.Buffer
	for i, edge := range p {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("[")
		buf.WriteString(edge.String())
		buf.WriteString("]")
	}
	return buf.String()
}

// some basic examples
type ZVertexInt int

func (v ZVertexInt) Equal(w ZVertexInt) bool {
	return v == w
}

func (v ZVertexInt) Less(w ZVertexInt) bool {
	return v < w
}

func (v ZVertexInt) String() string {
	return fmt.Sprintf("%d", v)
}

func NewZComplexEmptyTriangle() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{{0,1}, {1,2}, {2,0}})
}

func NewZComplexFilledTriangle() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{{0,1,2}})
}

func NewZComplexFromTriangles(S []ZTriangle[ZVertexInt]) *ZComplex[ZVertexInt] {
	vertices := make(map[ZVertex[ZVertexInt]]bool)
	edges := make(map[ZEdge[ZVertexInt]]bool)
	for _, t := range S {
		vertices[t[0]] = true
		vertices[t[1]] = true
		vertices[t[2]] = true
		edges[ZEdge[ZVertexInt]{t[0], t[1]}] = true
		edges[ZEdge[ZVertexInt]{t[1], t[2]}] = true
		edges[ZEdge[ZVertexInt]{t[2], t[0]}] = true
	}
	vertexBasis := make([]ZVertex[ZVertexInt], 0, len(vertices))
	for v := range vertices {
		vertexBasis = append(vertexBasis, v)
	}
	edgeBasis := make([]ZEdge[ZVertexInt], 0, len(edges))
	for e := range edges {
		edgeBasis = append(edgeBasis, e)
	}
	sortBases := true
	verbose := false
	return NewZComplex(vertexBasis, edgeBasis, S, sortBases, verbose)
}

func NewZComplexFromMaximalSimplices(S [][]int) *ZComplex[ZVertexInt] {
	vertices := make(map[ZVertex[ZVertexInt]]bool)
	edges := make(map[ZEdge[ZVertexInt]]bool)
	triangles := make(map[ZTriangle[ZVertexInt]]bool)

	for _, s := range S {
		switch len(s) {
		case 0:
			panic("empty simplex")
		case 1:
			vertices[ZVertexInt(s[0])] = true
		case 2:
			vertices[ZVertexInt(s[0])] = true
			vertices[ZVertexInt(s[1])] = true
			edges[NewZEdge[ZVertexInt](ZVertexInt(s[0]), ZVertexInt(s[1]))] = true
		case 3:
			vertices[ZVertexInt(s[0])] = true
			vertices[ZVertexInt(s[1])] = true
			vertices[ZVertexInt(s[2])] = true
			edges[NewZEdge[ZVertexInt](ZVertexInt(s[0]), ZVertexInt(s[1]))] = true
			edges[NewZEdge[ZVertexInt](ZVertexInt(s[1]), ZVertexInt(s[2]))] = true
			edges[NewZEdge[ZVertexInt](ZVertexInt(s[2]), ZVertexInt(s[0]))] = true
			triangles[NewZTriangle[ZVertexInt](ZVertexInt(s[0]), ZVertexInt(s[1]), ZVertexInt(s[2]))] = true
		default:
			panic("simplex of dimension > 3")
		}
	}
	vertexBasis := make([]ZVertex[ZVertexInt], 0, len(vertices))
	for v := range vertices {
		vertexBasis = append(vertexBasis, v)
	}
	edgeBasis := make([]ZEdge[ZVertexInt], 0, len(edges))
	for e := range edges {
		edgeBasis = append(edgeBasis, e)
	}
	triangleBasis := make([]ZTriangle[ZVertexInt], 0, len(triangles))
	for t := range triangles {
		triangleBasis = append(triangleBasis, t)
	}
	sortBases := true
	verbose := false
	return NewZComplex(vertexBasis, edgeBasis, triangleBasis, sortBases, verbose)
}

	
