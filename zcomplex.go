package golsv

import (
	"bytes"
	"fmt"
	"log"
	"sort"
)
// this file models a 2-d simplicial complex.  it was originally
// derived from complex.go in an attempt to factor out the common part
// in a type-generic way.

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

func (E ZEdge[T]) Contains(v ZVertex[T]) bool {
	return E[0] == v || E[1] == v
}

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

func (E ZEdge[T]) OtherVertex(v ZVertex[T]) ZVertex[T] {
	if v.Equal(E[0].(T)) {
		return E[1]
	}
	return E[0]
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
	edgeIndex     map[ZEdge[T]]int   // edge -> index in basis
	vertexIndex   map[ZVertex[T]]int // vertex -> index in basis
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
	C.edgeIndex = make(map[ZEdge[T]]int)
	for i, edge := range C.edgeBasis {
		C.edgeIndex[edge] = i
	}
}

func (C *ZComplex[T]) computeVertexIndex() {
	C.vertexIndex = make(map[ZVertex[T]]int)
	for i, vertex := range C.vertexBasis {
		C.vertexIndex[vertex] = i
	}
}

func (C *ZComplex[T]) computeD1() {
	C.d1 = NewSparseBinaryMatrix(len(C.vertexBasis), len(C.edgeBasis))
	for j, edge := range C.edgeBasis {
		for _, v := range edge {
			row, ok := C.vertexIndex[v]
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
			row, ok := C.edgeIndex[e]
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

func (C *ZComplex[T]) EdgeIndex() map[ZEdge[T]]int {
	return C.edgeIndex
}

// maps edge index to list of triangle indices
func (C *ZComplex[T]) EdgeToTriangleIncidenceMap() map[int][]int {
	d := C.D2()
	m := make(map[int][]int)
	for j := 0; j < d.NumColumns(); j++ {
		if d.ColumnWeight(j) != 3 {
			panic("expected 3 ones in column")
		}
		a := d.ScanDown(0, j)
		b := d.ScanDown(a+1, j)
		c := d.ScanDown(b+1, j)
		m[a] = append(m[a], j)
		m[b] = append(m[b], j)
		m[c] = append(m[c], j)
	}
	return m
}

func (C *ZComplex[T]) HasNeighbor(v, u ZVertex[T]) bool {
	for _, x := range C.Neighbors(v) {
		if x.Equal(u.(T)) {
			return true
		}
	}
	return false
}

func (C *ZComplex[T]) Neighbors(v ZVertex[T]) (nabes []ZVertex[T]) {
	for _, e := range C.EdgesContainingVertex(v) {
		nabes = append(nabes, e.OtherVertex(v))
	}
	return
}

// xxx test
func (C *ZComplex[T]) NumEdges() int {
	return len(C.edgeBasis)
}

// xxx test
func (C *ZComplex[T]) NumTriangles() int {
	return len(C.triangleBasis)
}

// xxx test
func (C *ZComplex[T]) NumVertices() int {
	return len(C.vertexBasis)
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
		if i, ok := C.edgeIndex[e]; ok {
			v.Set(i, 1)
		} else {
			panic(fmt.Sprintf("edge %v not in edge index", e))
		}
	}
	return v
}

func (C *ZComplex[T]) SortBasesByDistance(vertexIndex int) {
	base := C.vertexBasis[vertexIndex]
	newVertices := make([]ZVertex[T], len(C.vertexBasis))
	newVertexIndex := make(map[ZVertex[T]]int)
	i := 0
	C.BFS(base, func(v ZVertex[T]) {
		newVertices[i] = v
		newVertexIndex[v] = i
		i++
	})
	C.vertexBasis = newVertices
	C.vertexIndex = newVertexIndex
	sort.Slice(C.edgeBasis, func(i, j int) bool {
		a := C.edgeBasis[i]
		b := C.edgeBasis[j]
		return nearerEdgeInNewIndex(newVertexIndex, a, b)
	}) 
	C.computeEdgeIndex()
	sort.Slice(C.triangleBasis, func(i, j int) bool {
		a := C.triangleBasis[i]
		b := C.triangleBasis[j]
		return nearerTriangleInNewIndex(newVertexIndex, a, b)
	}) 
	C.d1 = nil
	C.d2 = nil
}

func nearerEdgeInNewIndex[T any](index map[ZVertex[T]]int, a ZEdge[T], b ZEdge[T]) bool {
	var m, n, o, p int
	var ok bool
	m, ok = index[a[0]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", a[0]))
	}
	n, ok = index[a[1]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", a[1]))
	}
	if n < m {
		m = n
	}
	o, ok = index[b[0]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", b[0]))
	}
	p, ok = index[b[1]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", b[1]))
	}
	if p < o {
		o = p
	}
	return m < o
}

func nearerTriangleInNewIndex[T any](index map[ZVertex[T]]int, a ZTriangle[T], b ZTriangle[T]) bool {
	var m, n, o, p int
	var ok bool
	m, ok = index[a[0]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", a[0]))
	}
	n, ok = index[a[1]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", a[1]))
	}
	if n < m {
		m = n
	}
	n, ok = index[a[2]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", a[2]))
	}
	if n < m {
		m = n
	}
	o, ok = index[b[0]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", b[0]))
	}
	p, ok = index[b[1]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", b[1]))
	}
	if p < o {
		o = p
	}
	p, ok = index[b[2]]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in new vertex index", b[2]))
	}
	if p < o {
		o = p
	}
	return m < o
}

func (C *ZComplex[T]) String() string {
	return fmt.Sprintf("complex with %v triangles, %v edges, %v vertices",
		len(C.triangleBasis), len(C.edgeBasis), len(C.vertexBasis))
}

func (C *ZComplex[T]) DumpBases() (s string) {
	s += "vertexBasis:\n"
	for i, v := range C.vertexBasis {
		s += fmt.Sprintf("  %d: %v\n", i, v)
	}
	s += "edgeBasis:\n"
	for i, e := range C.edgeBasis {
		s += fmt.Sprintf("  %d: %s\n", i, e)
	}
	s += "triangleBasis:\n"
	for i, t := range C.triangleBasis {
		s += fmt.Sprintf("  %d: %s\n", i, t)
	}		
	return s
}

// xxx test
func (C *ZComplex[T]) SubcomplexByTriangles(n int) *ZComplex[T] {
	truncated := C.TriangleBasis()[:n]
	return NewZComplexFromTrianglesGeneric(truncated)
}

func (C *ZComplex[T]) TriangleBasis() []ZTriangle[T] {
	return C.triangleBasis
}

func (C *ZComplex[T]) VertexBasis() []ZVertex[T] {
	return C.vertexBasis
}

// xxx test
func (C *ZComplex[T]) VertexToSparse(v ZVertex[T]) *Sparse {
	row, ok := C.vertexIndex[v]
	if !ok {
		panic(fmt.Sprintf("vertex %v not in vertex index", v))
	}
	M := NewSparseBinaryMatrix(len(C.vertexBasis), 1)
	M.Set(row, 0, 1)
	return M
}

// the resulting map has keys that are vertex indices and values that
// are lists of edge indices.
func (C *ZComplex[T]) VertexToEdgeIncidenceMap() map[int][]int {
	d := C.D1()
	m := make(map[int][]int)
	for j := 0; j < d.NumColumns(); j++ {
		if d.ColumnWeight(j) != 2 {
			panic("expected 2 ones in column")
		}
		a := d.ScanDown(0, j)
		b := d.ScanDown(a+1, j)
		m[a] = append(m[a], j)
		m[b] = append(m[b], j)
	}
	return m
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

// Note: this will only enumerate triangles that are connected to the
// start vertex.
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

// Note: this will only enumerate 3-cliques that are connected to the
// start vertex.  It will generate each 3-clique exactly once.
func (C *ZComplex[T]) BFWalk3Cliques(v ZVertex[T], f func(c [3]ZVertex[T])) {
	// in order to avoid double-counting, we will create a temporary
	// map to store the 3-cliques we have already seen.  since we need
	// these keys to be unique, we use a ZTriangle[T] as the key, even
	// though this is not necessarily a triangle (2-simplex) in the
	// complex C.
	seen := make(map[ZTriangle[T]]struct{})
	C.BFS(v, func(u ZVertex[T]) {
		// for each neighbor x
		//     for each other neighbor y
		//         check whether u-x-y is a 3-clique
		nabes := C.Neighbors(u)
		for i, x := range nabes {
			for j, y := range nabes {
				if i == j {
					continue
				}
				// we have u-x and u-y.  check for x-y.
				if C.HasNeighbor(x, y) {
					t := NewZTriangle[T](u, x, y)
					if _, ok := seen[t]; ok {
						continue
					}
					f([3]ZVertex[T]{u, x, y})
					seen[t] = struct{}{}
				}
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

func NewZComplexJoinedFilledTriangles() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{{0,1,2}, {1,2,}})
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

// xxx test
func NewZComplexFromTrianglesGeneric[T any](S []ZTriangle[T]) *ZComplex[T] {
	vertices := make(map[ZVertex[T]]bool)
	edges := make(map[ZEdge[T]]bool)
	for _, t := range S {
		vertices[t[0]] = true
		vertices[t[1]] = true
		vertices[t[2]] = true
		edges[NewZEdge(t[0], t[1])] = true
		edges[NewZEdge(t[1], t[2])] = true
		edges[NewZEdge(t[2], t[0])] = true
	}
	vertexBasis := make([]ZVertex[T], 0, len(vertices))
	for v := range vertices {
		vertexBasis = append(vertexBasis, v)
	}
	edgeBasis := make([]ZEdge[T], 0, len(edges))
	for e := range edges {
		edgeBasis = append(edgeBasis, e)
	}
	sortBases := false
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

func NewZComplexFromBoundaryMatrices(d_1, d_2 BinaryMatrix) *ZComplex[ZVertexInt] {
	numVertices := d_1.NumRows()
	vertexBasis := make([]ZVertex[ZVertexInt], 0, numVertices)
	for i := 0; i < numVertices; i++ {
		vertexBasis = append(vertexBasis, ZVertexInt(i))
	}
	numEdges := d_1.NumColumns()
	edgeBasis := make([]ZEdge[ZVertexInt], 0, numEdges)
	edgeIndex := make(map[ZEdge[ZVertexInt]]int)
	for j := 0; j < numEdges; j++ {
		v0 := d_1.ScanDown(0, j)
		if v0 < 0 {
			panic("expected vertex")
		}
		v1 := d_1.ScanDown(v0+1, j)
		if v1 < 0 {
			panic("expected vertex")
		}
		if d_1.ScanDown(v1+1, j) >= 0 {
			panic("too many ones in column")
		}
		edge := NewZEdge[ZVertexInt](ZVertexInt(v0), ZVertexInt(v1))
		edgeBasis = append(edgeBasis, edge)
		edgeIndex[edge] = j
	}
	numTriangles := d_2.NumColumns()
	triangleBasis := make([]ZTriangle[ZVertexInt], 0, numTriangles)
	for j := 0; j < numTriangles; j++ {
		e0 := d_2.ScanDown(0, j)
		if e0 < 0 {
			panic("expected edge")
		}
		e1 := d_2.ScanDown(e0+1, j)
		if e1 < 0 {
			panic("expected edge")
		}
		e2 := d_2.ScanDown(e1+1, j)
		if e2 < 0 {
			panic("expected edge")
		}
		if d_2.ScanDown(e2+1, j) >= 0 {
			panic("too many ones in column")
		}
		v0 := edgeBasis[e0][0]
		v1 := edgeBasis[e0][1]
		var v2 ZVertex[ZVertexInt]
		found := false
		k := 0
		if edgeBasis[e1][k] != v0 && edgeBasis[e1][k] != v1 {
			v2 = edgeBasis[e1][k]
			found = true
		}
		if !found {
			k = 1
			if edgeBasis[e1][k] != v0 && edgeBasis[e1][k] != v1 {
				v2 = edgeBasis[e1][k]
				found = true
			}
		}
		if !found {
			panic("expected three vertices in triangle")
		}
		triangleBasis = append(triangleBasis, NewZTriangle[ZVertexInt](v0, v1, v2))
	}
	sortBases := false
	verbose := false
	return NewZComplex(vertexBasis, edgeBasis, triangleBasis, sortBases, verbose)
}
	
