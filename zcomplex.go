package golsv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
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
	vertexBasis    []ZVertex[T]
	edgeBasis      []ZEdge[T]
	triangleBasis  []ZTriangle[T]
	d1             BinaryMatrix
	d2             BinaryMatrix
	edgeIndex      map[ZEdge[T]]int   // edge -> index in basis
	vertexIndex    map[ZVertex[T]]int // vertex -> index in basis
	adjacencyIndex map[int][]int      // vertex index -> list of adjacent vertex indices
	verbose        bool
}

func NewZComplex[T any](vertexBasis []ZVertex[T], edgeBasis []ZEdge[T], triangleBasis []ZTriangle[T], sortBases bool, verbose bool) *ZComplex[T] {
	if edgeBasis == nil {
		edgeBasis = make([]ZEdge[T], 0)
	}
	if triangleBasis == nil {
		triangleBasis = make([]ZTriangle[T], 0)
	}
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
		vertexBasis:   vertexBasis,
		edgeBasis:     edgeBasis,
		triangleBasis: triangleBasis,
		verbose:       verbose,
	}
	C.computeVertexIndex()
	C.computeEdgeIndex()
	if verbose {
		log.Printf("Complex: %v", &C)
	}
	//log.Printf("xxx\n%s", C.DumpBases())
	return &C
}

func (C *ZComplex[T]) BFWalk3Cliques(f func(c [3]ZVertex[T])) {
	// in order to avoid double-counting, we will create a temporary
	// map to store the 3-cliques we have already seen.  since we need
	// these keys to be unique, we use a ZTriangle[T] as the key, even
	// though this is not necessarily a triangle (2-simplex) in the
	// complex C.
	seen := make(map[ZTriangle[T]]struct{})
	for uId := range C.vertexBasis {
		//
		//       f     g     h
		//    u --- v --- w --- x
		//
		for _, vId := range C.Neighbors(uId) {
			for _, wId := range C.Neighbors(vId) {
				if wId == uId {
					continue
				}
				for _, xId := range C.Neighbors(wId) {
					if xId == vId {
						continue
					}
					if xId == uId {
						u := C.vertexBasis[uId]
						v := C.vertexBasis[vId]
						w := C.vertexBasis[wId]
						t := NewZTriangle(u, v, w)
						if _, ok := seen[t]; ok {
							continue
						}
						f([3]ZVertex[T]{u, v, w})
						seen[t] = struct{}{}
					}
				}
			}
		}
	}
}

func (C *ZComplex[T]) computeEdgeIndex() {
	if C.verbose {
		log.Printf("creating edge index; edges=%d", len(C.edgeBasis))
	}
	C.edgeIndex = make(map[ZEdge[T]]int)
	for i, edge := range C.edgeBasis {
		C.edgeIndex[edge] = i
	}
}

func (C *ZComplex[T]) computeVertexIndex() {
	if C.verbose {
		log.Printf("creating vertex index; vertices=%d", len(C.vertexBasis))
	}
	C.vertexIndex = make(map[ZVertex[T]]int)
	for i, vertex := range C.vertexBasis {
		C.vertexIndex[vertex] = i
	}
}

func (C *ZComplex[T]) computeD1() {
	if C.verbose {
		log.Printf("computing d1 boundary matrix")
	}
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
	if C.verbose {
		log.Printf("computing d2 boundary matrix")
	}
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

func (C *ZComplex[T]) AddEdge(u, v int) {
	if u < 0 || u >= len(C.vertexBasis) || v < 0 || v >= len(C.vertexBasis) {
		panic("vertex index out of range")
	}
	if u == v {
		panic("cannot add self-loop")
	}
	
	// Create edge from vertex indices
	edge := NewZEdge(C.vertexBasis[u], C.vertexBasis[v])
	
	// Check if edge already exists
	for _, existingEdge := range C.edgeBasis {
		if existingEdge.Equal(edge) {
			return // Edge already exists
		}
	}
	
	// Add edge to basis
	C.edgeBasis = append(C.edgeBasis, edge)
	
	// Update adjacency index in place if it exists
	if C.adjacencyIndex != nil {
		C.adjacencyIndex[u] = append(C.adjacencyIndex[u], v)
		C.adjacencyIndex[v] = append(C.adjacencyIndex[v], u)
	}
	
	// Invalidate other cached indices
	C.edgeIndex = nil // xxx improve this
	C.d1 = nil
	C.d2 = nil
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

func (C *ZComplex[T]) DepthFiltration(initialVertex ZVertex[T],
	handler func(depth int, subcomplex *ZComplex[T], verticesAtDepth []ZVertex[T])) {
	vertexIndicesToInclude := make(map[int]bool)
	curDepth := 0
	verticesAtDepth := make([]ZVertex[T], 0)
	C.BFS(initialVertex, func(v ZVertex[T], vdepth int) (stop bool) {
		if vdepth > curDepth {
			subcomplex := C.SubcomplexByVertices(vertexIndicesToInclude)
			handler(curDepth, subcomplex, verticesAtDepth)
			curDepth = vdepth
			verticesAtDepth = verticesAtDepth[:0]
		}
		n := C.vertexIndex[v]
		vertexIndicesToInclude[n] = true
		verticesAtDepth = append(verticesAtDepth, v)
		return false
	})
	subcomplex := C.SubcomplexByVertices(vertexIndicesToInclude)
	handler(curDepth, subcomplex, verticesAtDepth)
}

// DFS performs a depth-first search traversal starting from vertex v.
// For each visited vertex, it calls f with the vertex and its depth.
// If f returns true, the traversal stops early.
func (C *ZComplex[T]) DFS(v ZVertex[T], f func(u ZVertex[T], depth int) (stop bool)) {
	visited := make(map[ZVertex[T]]struct{})
	C.dfsUtil(v, 0, visited, f)
}

func (C *ZComplex[T]) dfsUtil(v ZVertex[T], depth int, visited map[ZVertex[T]]struct{}, f func(u ZVertex[T], depth int) (stop bool)) bool {
	visited[v] = struct{}{}
	if stop := f(v, depth); stop {
		return true
	}
	vertexIdx := C.vertexIndex[v]
	for _, neighborIdx := range C.Neighbors(vertexIdx) {
		neighbor := C.vertexBasis[neighborIdx]
		if _, ok := visited[neighbor]; !ok {
			if stop := C.dfsUtil(neighbor, depth+1, visited, f); stop {
				return true
			}
		}
	}
	return false
}

func (C *ZComplex[T]) DualComplex() *ZComplex[ZVertexInt] {
	// the dual complex, as we define it here, is the complex whose vertices are the
	// triangles of the original complex, and whose edges are the
	// triangles that share an edge in the original complex.
	// in addition, our dual complex has a triangle for each edge in the
	// original complex that is incident to three triangles in the original.
	// in other words, we take the dual graph of the original complex,
	// and where we have a triangle in the dual graph which has edges
	// all corresponding to the same edge in the original complex, we
	// fill that triangle in the dual complex. note that this is NOT
	// the same as taking the clique complex of the dual complex, since
	// in general, one could have a triangle in the dual graph corresponding
	// to three triangles in the original complex that are not all incident to the
	// a single edge but rather to two different edges.
	vertices := make([]ZVertex[ZVertexInt], len(C.triangleBasis))
	for i := range C.triangleBasis {
		vertices[i] = ZVertexInt(i)
	}
	edges := make([]ZEdge[ZVertexInt], 0)
	triangles := make([]ZTriangle[ZVertexInt], 0)
	edgeToTriangles := C.EdgeToTriangleIncidenceMap()
	for i := range C.EdgeBasis() {
		toAdd := make([]ZEdge[ZVertexInt], 0)
		T1 := edgeToTriangles[i]
		for _, t := range T1 {
			for _, u := range T1 {
				if t < u {
					toAdd = append(toAdd, NewZEdge(ZVertexInt(t), ZVertexInt(u)))
				}
			}
		}
		// if for the current edge in the original, we added three edges to
		// the dual, which must correspond to three triangles in the original
		// incident to the original edge, then also add a triangle to the dual.
		// to generalize this, for n \ge 3, we could add an n-simplex for
		// n triangles incident to the original edge.  that is not needed currently.
		edges = append(edges, toAdd...)
		if len(toAdd) == 3 {
			var t, u, v ZVertex[ZVertexInt]
			t = toAdd[0][0]
			u = toAdd[0][1]
			if toAdd[1][0] == t || toAdd[1][0] == u {
				v = toAdd[1][1]
			} else {
				v = toAdd[1][0]
			}
			triangle := NewZTriangle(t, u, v)
			triangles = append(triangles, triangle)
		}
	}
	sortBases := false
	verbose := C.verbose
	return NewZComplex(vertices, edges, triangles, sortBases, verbose)
}

func (C *ZComplex[T]) Degree(v int) int {
	return len(C.Neighbors(v))
}

func (C *ZComplex[T]) DeleteEdge(u, v int) {
	if u < 0 || u >= len(C.vertexBasis) || v < 0 || v >= len(C.vertexBasis) {
		panic("vertex index out of range")
	}
	if len(C.triangleBasis) > 0 {
		panic("DeleteEdge can only be used on graphs without triangles")
	}
	
	// Create edge from vertex indices
	targetEdge := NewZEdge(C.vertexBasis[u], C.vertexBasis[v])
	
	// Find and remove edge
	for i, edge := range C.edgeBasis {
		if edge.Equal(targetEdge) {
			// Remove edge at index i
			C.edgeBasis = append(C.edgeBasis[:i], C.edgeBasis[i+1:]...)
			
			// Update adjacency index in place if it exists
			if C.adjacencyIndex != nil {
				// Remove v from u's neighbors
				for j, neighbor := range C.adjacencyIndex[u] {
					if neighbor == v {
						C.adjacencyIndex[u] = append(C.adjacencyIndex[u][:j], C.adjacencyIndex[u][j+1:]...)
						break
					}
				}
				// Remove u from v's neighbors
				for j, neighbor := range C.adjacencyIndex[v] {
					if neighbor == u {
						C.adjacencyIndex[v] = append(C.adjacencyIndex[v][:j], C.adjacencyIndex[v][j+1:]...)
						break
					}
				}
			}
			
			// Invalidate other cached indices
			C.edgeIndex = nil // xxx improve this
			C.d1 = nil
			C.d2 = nil
			return
		}
	}
}

func (C *ZComplex[T]) EdgeBasis() []ZEdge[T] {
	return C.edgeBasis
}

func (C *ZComplex[T]) EdgesContainingVertex(v ZVertex[T]) []ZEdge[T] {
	var edges []ZEdge[T]
	nabes := C.Neighbors(C.vertexIndex[v])
	for _, uId := range nabes {
		u := C.vertexBasis[uId]
		e := NewZEdge(v, u)
		edges = append(edges, e)
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

func (C *ZComplex[T]) ensureAdjacencyIndex() {
	if C.adjacencyIndex != nil {
		return
	}
	C.adjacencyIndex = make(map[int][]int)
	for i := range C.vertexBasis {
		C.adjacencyIndex[i] = make([]int, 0)
	}
	for _, e := range C.edgeBasis {
		i, ok := C.vertexIndex[e[0]]
		if !ok {
			panic(fmt.Sprintf("vertex %v not in vertex index", e[0]))
		}
		j, ok := C.vertexIndex[e[1]]
		if !ok {
			panic(fmt.Sprintf("vertex %v not in vertex index", e[1]))
		}
		C.adjacencyIndex[i] = append(C.adjacencyIndex[i], j)
		C.adjacencyIndex[j] = append(C.adjacencyIndex[j], i)
	}
}

func (C *ZComplex[T]) Fill3Cliques() {
	if C.verbose {
		log.Printf("Filling 3-cliques")
	}
	C.BFWalk3Cliques(func(c [3]ZVertex[T]) {
		t := NewZTriangle(c[0], c[1], c[2])
		C.triangleBasis = append(C.triangleBasis, t)
	})
}

func (C *ZComplex[T]) MaximalSimplicesString() string {
	seenVertices := make(map[ZVertex[T]]struct{})
	seenEdges := make(map[ZEdge[T]]struct{})
	var buf bytes.Buffer
	buf.WriteString("[][]int{")
	n := 0
	for _, t := range C.triangleBasis {
		if n != 0 {
			buf.WriteString(", ")
		}
		n++
		buf.WriteString("{")
		for i, v := range t {
			seenVertices[v] = struct{}{}
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(v.String())
		}
		buf.WriteString("}")
		for _, e := range t.Edges() {
			seenEdges[e] = struct{}{}
		}
	}
	for _, e := range C.edgeBasis {
		if _, ok := seenEdges[e]; ok {
			continue
		}
		if n != 0 {
			buf.WriteString(", ")
		}
		n++
		buf.WriteString("{")
		for i, v := range e {
			seenVertices[v] = struct{}{}
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(v.String())
		}
		buf.WriteString("}")
	}
	for _, v := range C.vertexBasis {
		if _, ok := seenVertices[v]; ok {
			continue
		}
		if n != 0 {
			buf.WriteString(", ")
		}
		n++
		buf.WriteString(fmt.Sprintf("{%s}", v.String()))
	}
	buf.WriteString("}")
	return buf.String()
}

func parseSimplicesString(s string) (simplices [][]int, err error) {
	// this is slightly weird AI code, but it is good enough.
	// parse the string into a slice of slices of ints.
	// the string is of the form [][]int{{1, 2}, {3, 4, 5}, {6}}
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[][]int{") || !strings.HasSuffix(s, "}") {
		return nil, fmt.Errorf("invalid simplices string format: expected prefix '[][]int{' and suffix '}'")
	}
	// Remove the "[][]int" prefix.
	jsonStr := strings.TrimPrefix(s, "[][]int")
	// Replace Go-style slice braces with JSON-style array brackets.
	jsonStr = strings.ReplaceAll(jsonStr, "{", "[")
	jsonStr = strings.ReplaceAll(jsonStr, "}", "]")
	// Handle the case of an empty top-level list like "[][]int{}" which becomes "[]"
	// or a list with an empty inner list like "[][]int{{}}" which becomes "[[]]"
	if jsonStr == "" && s == "[][]int{}" { // Original string was exactly "[][]int{}"
		jsonStr = "[]"
	}
	err = json.Unmarshal([]byte(jsonStr), &simplices)
	if err != nil {
		// Provide context for debugging
		return nil, fmt.Errorf("failed to unmarshal JSON from transformed string '%s' (derived from '%s'): %w", jsonStr, s, err)
	}
	return
}

func (C *ZComplex[T]) IsRegular() bool {
	if C.NumVertices() == 0 {
		return true // vacuously true
	}
	
	firstDegree := C.Degree(0)
	for i := 1; i < C.NumVertices(); i++ {
		if C.Degree(i) != firstDegree {
			return false
		}
	}
	return true
}

func (C *ZComplex[T]) IsNeighbor(u, v int) bool {
	neighbors := C.Neighbors(u)
	for _, neighbor := range neighbors {
		if neighbor == v {
			return true
		}
	}
	return false
}

func (C *ZComplex[T]) Neighbors(v int) (nabes []int) {
	C.ensureAdjacencyIndex()
	return C.adjacencyIndex[v]
}

func (C *ZComplex[T]) NumEdges() int {
	return len(C.edgeBasis)
}

func (C *ZComplex[T]) NumTriangles() int {
	return len(C.triangleBasis)
}

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

// Returns a new complex with the bases sorted by distance from the given vertex.
func (C *ZComplex[T]) SortBasesByDistance(vertexIndex int) *ZComplex[T] {
	initialVertex := C.vertexBasis[vertexIndex]
	newVertices := make([]ZVertex[T], len(C.vertexBasis))
	newVertexIndex := make(map[ZVertex[T]]int)
	i := 0
	// Note that we handle each connected component of the complex.
	nextInitialVertex := initialVertex
	for i < len(C.vertexBasis) {
		C.BFS(nextInitialVertex, func(v ZVertex[T], depth int) (stop bool) {
			newVertices[i] = v
			newVertexIndex[v] = i
			i++
			return false
		})
		if i < len(C.vertexBasis) {
			// scan for a vertex we haven't seen yet
			for _, v := range C.vertexBasis {
				if _, ok := newVertexIndex[v]; !ok {
					nextInitialVertex = v
					break
				}
			}
		}
	}
	newEdges := make([]ZEdge[T], len(C.edgeBasis))
	copy(newEdges, C.edgeBasis)
	sort.Slice(newEdges, func(i, j int) bool {
		a := newEdges[i]
		b := newEdges[j]
		return nearerEdgeInNewIndex(newVertexIndex, a, b)
	})
	newTriangles := make([]ZTriangle[T], len(C.triangleBasis))
	copy(newTriangles, C.triangleBasis)
	sort.Slice(newTriangles, func(i, j int) bool {
		a := newTriangles[i]
		b := newTriangles[j]
		return nearerTriangleInNewIndex(newVertexIndex, a, b)
	})
	sortBases := false
	verbose := C.verbose
	return NewZComplex(newVertices, newEdges, newTriangles, sortBases, verbose)
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

func (C *ZComplex[T]) SubcomplexByDepth(depth int) *ZComplex[T] {
	vertexIndicesToInclude := make(map[int]bool)
	C.BFS(C.vertexBasis[0], func(v ZVertex[T], vdepth int) (stop bool) {
		if vdepth <= depth {
			vertexIndicesToInclude[C.vertexIndex[v]] = true
		} else if vdepth > depth {
			return true
		}
		return false
	})
	return C.SubcomplexByVertices(vertexIndicesToInclude)
}

// xxx test
func (C *ZComplex[T]) SubcomplexByEdges(edgeIndicesToInclude map[int]any) *ZComplex[T] {
	vertices := make(map[ZVertex[T]]bool)
	edgeBasis := make([]ZEdge[T], 0)
	for i, e := range C.edgeBasis {
		if _, ok := edgeIndicesToInclude[i]; ok {
			edgeBasis = append(edgeBasis, e)
			vertices[e[0]] = true
			vertices[e[1]] = true
		}
	}
	vertexBasis := make([]ZVertex[T], 0, len(vertices))
	for v := range vertices {
		vertexBasis = append(vertexBasis, v)
	}
	sortBases := false
	verbose := false
	return NewZComplex(vertexBasis, edgeBasis, nil, sortBases, verbose)
}

// xxx test
func (C *ZComplex[T]) SubcomplexByTriangles(triangleIndicesToInclude map[int]any) *ZComplex[T] {
	filtered := make([]ZTriangle[T], 0)
	for i, t := range C.triangleBasis {
		if _, ok := triangleIndicesToInclude[i]; ok {
			filtered = append(filtered, t)
		}
	}
	return NewZComplexFromTrianglesGeneric(filtered)
}

func (C *ZComplex[T]) SubcomplexByVertices(vertexIndicesToInclude map[int]bool) *ZComplex[T] {
	vertices := make([]ZVertex[T], 0)
	for i, v := range C.vertexBasis {
		if _, ok := vertexIndicesToInclude[i]; ok {
			vertices = append(vertices, v)
		}
	}
	edges := make([]ZEdge[T], 0)
	for _, e := range C.edgeBasis {
		if _, ok := vertexIndicesToInclude[C.vertexIndex[e[0]]]; ok {
			if _, ok := vertexIndicesToInclude[C.vertexIndex[e[1]]]; ok {
				edges = append(edges, e)
			}
		}
	}
	triangles := make([]ZTriangle[T], 0)
	for _, t := range C.triangleBasis {
		if _, ok := vertexIndicesToInclude[C.vertexIndex[t[0]]]; ok {
			if _, ok := vertexIndicesToInclude[C.vertexIndex[t[1]]]; ok {
				if _, ok := vertexIndicesToInclude[C.vertexIndex[t[2]]]; ok {
					triangles = append(triangles, t)
				}
			}
		}
	}
	sortBases := false
	verbose := false
	return NewZComplex(vertices, edges, triangles, sortBases, verbose)
}

func (C *ZComplex[T]) TriangleBasis() []ZTriangle[T] {
	return C.triangleBasis
}

func (C *ZComplex[T]) TriangularDepthFiltration(initialVertex ZVertex[T],
	handler func(depth int, subcomplex *ZComplex[T]) (stop bool)) {
	Y := C.SortBasesByDistance(C.vertexIndex[initialVertex])
	vertexIndicesToInclude := make(map[int]bool)
	stop := false
	for i, t := range Y.triangleBasis {
		vertexIndicesToInclude[Y.vertexIndex[t[0]]] = true
		vertexIndicesToInclude[Y.vertexIndex[t[1]]] = true
		vertexIndicesToInclude[Y.vertexIndex[t[2]]] = true
		subcomplex := Y.SubcomplexByVertices(vertexIndicesToInclude)
		stop = handler(i, subcomplex)
		if stop {
			break
		}
	}
	if !stop {
		// one last call that includes any vertices not already included
		if len(vertexIndicesToInclude) < len(Y.vertexBasis) {
			handler(len(Y.triangleBasis), Y)
		}
	}
}

func (C *ZComplex[T]) VertexBasis() []ZVertex[T] {
	return C.vertexBasis
}

func (C *ZComplex[T]) VertexIndex() map[ZVertex[T]]int {
	return C.vertexIndex
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

// Note: will only enumerate the vertices in the same connected component as v.
func (C *ZComplex[T]) BFS(v ZVertex[T], f func(u ZVertex[T], depth int) (stop bool)) {
	visited := make(map[ZVertex[T]]struct{})
	queue := NewZVertexQueue[T]()
	queue.Enqueue(ZVertexTask[T]{v, 0})
	stop := f(v, 0)
	if stop {
		return
	}
	visited[v] = struct{}{}
	for {
		task, ok := queue.Dequeue()
		if !ok {
			break
		}
		for _, e := range C.EdgesContainingVertex(task.v) {
			for _, w := range e {
				if _, ok := visited[w]; !ok {
					queue.Enqueue(ZVertexTask[T]{w, task.depth + 1})
					stop = f(w, task.depth+1)
					visited[w] = struct{}{}
					if stop {
						return
					}
				}
			}
		}
	}
}

type ZVertexTask[T any] struct {
	v     ZVertex[T]
	depth int
}

type ZVertexQueue[T any] struct {
	slice []ZVertexTask[T]
}

func NewZVertexQueue[T any]() *ZVertexQueue[T] {
	return &ZVertexQueue[T]{
		slice: make([]ZVertexTask[T], 0),
	}
}

func (Q *ZVertexQueue[T]) Enqueue(t ZVertexTask[T]) {
	Q.slice = append(Q.slice, t)
}

func (Q *ZVertexQueue[T]) Dequeue() (ZVertexTask[T], bool) {
	if len(Q.slice) == 0 {
		return ZVertexTask[T]{}, false
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
	C.BFS(v, func(u ZVertex[T], depth int) (stop bool) {
		for _, t := range C.TrianglesContainingVertex(u) {
			if _, ok := tVisited[t]; !ok {
				f(t)
				tVisited[t] = struct{}{}
			}
		}
		return false
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
	return NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2}, {2, 0}})
}

func NewZComplexFilledTriangle() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}})
}

func NewZComplexJoinedFilledTriangles() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2}})
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
	return NewZComplexFromMaximalSimplicesOptionalSort(S, true)
}

// If sortBases is false, preserve the order of simplices as given.
func NewZComplexFromMaximalSimplicesOptionalSort(S [][]int, sortBases bool) *ZComplex[ZVertexInt] {
	verticesSeen := make(map[ZVertex[ZVertexInt]]bool)
	vertexBasis := make([]ZVertex[ZVertexInt], 0)
	edgesSeen := make(map[ZEdge[ZVertexInt]]bool)
	edgeBasis := make([]ZEdge[ZVertexInt], 0)
	trianglesSeen := make(map[ZTriangle[ZVertexInt]]bool)
	triangleBasis := make([]ZTriangle[ZVertexInt], 0)

	for _, s := range S {
		switch len(s) {
		case 0:
			panic("empty simplex")
		case 1:
			v := ZVertexInt(s[0])
			if _, ok := verticesSeen[v]; !ok {
				verticesSeen[v] = true
				vertexBasis = append(vertexBasis, v)
			}
		case 2:
			v0 := ZVertexInt(s[0])
			v1 := ZVertexInt(s[1])
			if _, ok := verticesSeen[v0]; !ok {
				verticesSeen[v0] = true
				vertexBasis = append(vertexBasis, v0)
			}
			if _, ok := verticesSeen[v1]; !ok {
				verticesSeen[v1] = true
				vertexBasis = append(vertexBasis, v1)
			}
			e := NewZEdge[ZVertexInt](v0, v1)
			if _, ok := edgesSeen[e]; !ok {
				edgesSeen[e] = true
				edgeBasis = append(edgeBasis, e)
			}
		case 3:
			v0 := ZVertexInt(s[0])
			v1 := ZVertexInt(s[1])
			v2 := ZVertexInt(s[2])
			if _, ok := verticesSeen[v0]; !ok {
				verticesSeen[v0] = true
				vertexBasis = append(vertexBasis, v0)
			}
			if _, ok := verticesSeen[v1]; !ok {
				verticesSeen[v1] = true
				vertexBasis = append(vertexBasis, v1)
			}
			if _, ok := verticesSeen[v2]; !ok {
				verticesSeen[v2] = true
				vertexBasis = append(vertexBasis, v2)
			}
			e0 := NewZEdge[ZVertexInt](v0, v1)
			if _, ok := edgesSeen[e0]; !ok {
				edgesSeen[e0] = true
				edgeBasis = append(edgeBasis, e0)
			}
			e1 := NewZEdge[ZVertexInt](v0, v2)
			if _, ok := edgesSeen[e1]; !ok {
				edgesSeen[e1] = true
				edgeBasis = append(edgeBasis, e1)
			}
			e2 := NewZEdge[ZVertexInt](v1, v2)
			if _, ok := edgesSeen[e2]; !ok {
				edgesSeen[e2] = true
				edgeBasis = append(edgeBasis, e2)
			}
			t := NewZTriangle[ZVertexInt](v0, v1, v2)
			if _, ok := trianglesSeen[t]; !ok {
				trianglesSeen[t] = true
				triangleBasis = append(triangleBasis, t)
			}
		default:
			panic("simplex of dimension > 3")
		}
	}
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
