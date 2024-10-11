package golsv

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type Vertex = MatGF

var NewVertexFromString = NewMatGFFromString

func (V *Vertex) Translate(lsv *LsvContext, g *MatGF) *Vertex {
	// Note: group is acting on the *left* since the cayley graph
	// words are formed by multiplying on the *right*.
	U := g.Multiply(lsv, V)
	U.MakeCanonical(lsv)
	return U
}

type Edge [2]Vertex

// NOTE: it's important to always use NewEdge to create an Edge,
// because it ensures that the vertices are in the correct order.
// this canonical ordering allows edges to be used as map keys.
func NewEdge(a, b Vertex) Edge {
	if a.Less(&b) {
		return Edge{a, b}
	} else {
		return Edge{b, a}
	}
}

func NewEdgeFromString(s string) Edge {
	entries := strings.Split(strings.Trim(s, " []"), "] [")
	if len(entries) != 2 {
		panic(fmt.Sprintf("invalid edge string: %s", s))
	}
	a := NewMatGFFromString(entries[0])
	b := NewMatGFFromString(entries[1])
	return NewEdge(a, b)
}

func (E Edge) Contains(lsv *LsvContext, v Vertex) bool {
	return E[0].Equal(lsv, &v) || E[1].Equal(lsv, &v)
}

func (E Edge) ContainsOrigin(lsv *LsvContext) bool {
	return E.Contains(lsv, *MatGfIdentity)
}

func (E Edge) Equal(lsv *LsvContext, F Edge) bool {
	return E[0].Equal(lsv, &F[0]) && E[1].Equal(lsv, &F[1])
}

func (E Edge) Intersects(lsv *LsvContext, F Edge) bool {
	return E[0].Equal(lsv, &F[0]) || E[0].Equal(lsv, &F[1]) ||
		E[1].Equal(lsv, &F[0]) || E[1].Equal(lsv, &F[1])
}

func (E Edge) Translate(lsv *LsvContext, g *MatGF) Edge {
	return NewEdge(*E[0].Translate(lsv, g), *E[1].Translate(lsv, g))
}

func (E Edge) Less(F Edge) bool {
	if E[0].Less(&F[0]) {
		return true
	} else if F[0].Less(&E[0]) {
		return false
	} else {
		return E[1].Less(&F[1])
	}
}

// xxx perhaps these places that require an LsvContext should
// be methods on LsvContext?
func (E Edge) OrbitContainsEdge(lsv *LsvContext, f Edge) (g *MatGF, ok bool) {
	a, b := f[0], f[1]
	c, d := E[0], E[1]
	// Note: group acts on left.
	//
	// if g*E = f, then either
	//     1) g*[c,d] = [a,b]
	//       so g*c = a and g*d = b, hence g = a c^-1
	// 
	//  OR 2) g*[c,d] = [b,a]
	//       so g*c = b and g*d = a, hence g = b c^-1
	//
	cInv := c.Inverse(lsv)
	g = a.Multiply(lsv, cInv)
	gd := g.Multiply(lsv, &d)
	if gd.Equal(lsv, &b) {
		return g, true
	}		
	g = b.Multiply(lsv, cInv)
	gd = g.Multiply(lsv, &d)
	if gd.Equal(lsv, &a) {
		return g, true
	}
	return nil, false
}

func (E Edge) String() string {
	return fmt.Sprintf("[%s %s]", E[0], E[1])
}

// We don't record edge orientation; the ordering of the vertices in
// the Edge []Vertex type is lexical, not geometric.  In this
// Generator method, we return the generator that takes E[0] to E[1].
// If used in a context where it known that the orientation should be
// reversed, the caller should use the inverse of the returned
// generator.
func (E Edge) Generator(lsv *LsvContext) *MatGF {
	return E[0].Inverse(lsv).Multiply(lsv, &E[1]).MakeCanonical(lsv)
}


// xxx this might be more appropriately called EdgeSet, unless we
// enforce that the edges are sorted such that adjacent edges share a
// vertex.
type Path []Edge

// NewPathFromEdges sorts the edges so that adjacent edges share a
// vertex.
func NewPathFromEdges(lsv *LsvContext, edges []Edge) Path {
	n := len(edges)
	path := make([]Edge, n)
	marks := make([]bool, n)
	path[0] = edges[0]
	marks[0] = true
	for i := 1; i < n; i++ {
		foundNextVertex := false
		for j := 0; j < n; j++ {
			if marks[j] {
				continue
			}
			if path[i-1].Intersects(lsv, edges[j]) {
				path[i] = edges[j]
				marks[j] = true
				foundNextVertex = true
				break
			}
		}
		if !foundNextVertex {
			panic("discontinuous edge set")
		}
	}
	return path
}

func (p Path) Chords(lsv *LsvContext, gens []MatGF) []Edge {
	// log.Printf("xxx Chords: len(gens)= %v", len(gens))
	chords := make([]Edge, 0)
 	_, vertices := p.Word(lsv)
	// if the path is a cycle, i.e. the first and last vertices are
	// the same, then we don't want to consider the last vertex as the
	// end of a chord, since it would be redundant with counting that
	// from chord as beginning at the first vertex.  and if the path
	// is a cycle, we also don't want to consider the next-to-last
	// vertex as the end of a chord either, since there is already an
	// edge connecting the next-to-last vertex to the first vertex.
	n := len(vertices)
	isCycle := vertices[0].Equal(lsv, &vertices[n-1])
	if isCycle {
		n--
	}
	for i := 0; i < n - 2; i++ {
		v := vertices[i]
		end := n
		if isCycle && i == 0 && end > 0 {
			end--
		}
		for j := i+2; j < end; j++ {
			// log.Printf("xxx i=%d j=%d", i, j)
			u := vertices[j]
			for _, g := range gens {
				// log.Printf("xxx g=%s", g)
				if v.Translate(lsv, &g).Equal(lsv, &u) {
					chords = append(chords, NewEdge(v, u))
				}
			}
		}
	}
	return chords
}

// xxx wip
// func (p Path) EnumerateChords(lsv *LsvContext, gens []MatGF, minLen int, maxLen int, yield chan Path) []Edge {
// 	// xxx
//  	_, vertices := p.Word(lsv)
// 	n := len(vertices)
// 	for i := 0; i < n - 2; i++ {
// 		v := vertices[i]

// 		// xxx use custom todo-queue enumerator here, pruning at
// 		// words that intersect p at less than the current length. and
// 		// when incrementing length, does not have to redo enumeration
// 		// of subwords.
// 		for _, g := range gens {
			
// 		}
// 	}
// 	return nil
// }

// xxx test
func (p Path) Copy() Path {
	q := make(Path, len(p))
	copy(q, p)
	return q
}

func (p Path) Equal(lsv *LsvContext, q Path) bool {
	if len(p) != len(q) {
		return false
	}
	for i := range p {
		if !p[i].Equal(lsv, q[i]) {
			return false
		}
	}
	return true
}

func (p Path) SplitByChord(lsv *LsvContext, chord Edge) (Path, Path) {
	// log.Printf("xxx SplitByChord: chord=%s", chord)
	return nil, nil
}

func (p Path) Start(lsv *LsvContext) Vertex {
	if len(p) == 0 {
		panic("empty path")
	}
	if len(p) == 1 {
		return p[0][0]
	}
	// the start is the vertex that is not the intersection of the
	// first two edges.
	if p[0][0].Equal(lsv, &p[1][0]) || p[0][0].Equal(lsv, &p[1][1]) {
		return p[0][1]
	} else {
		return p[0][0]
	}
}

func (p Path) String() string {
	s := make([]string, len(p))
	for i := range p {
		s[i] = p[i].String()
	}
	return fmt.Sprintf("[%s]", strings.Join(s, "\n"))
}

// func (p Path) Translate(g *MatGF) Path {
// 	q := make(Path, len(p))
// 	for i := range p {
// 		q[i] = p[i].Translate(g)
// 	}
// 	return q
// }

func (p Path) Triangle(lsv *LsvContext) Triangle {
	if len(p) != 3 {
		panic("path does not have three edges")
	}
	// get the three distinct vertices
	if !p[1][0].Equal(lsv, &p[0][0]) && !p[1][0].Equal(lsv, &p[0][1]) {
		return NewTriangle(p[0][0], p[0][1], p[1][0])
	} else if !p[1][1].Equal(lsv, &p[0][0]) && !p[1][1].Equal(lsv, &p[0][1]) {
		return NewTriangle(p[0][0], p[0][1], p[1][1])
	}
	panic(fmt.Sprintf("path does not have three distinct vertices:\n%v", p))
}

// Here we account for the orientation of each edge.  In particular,
// for each edge E, E.Generator() gives us the generator g such that
// E[1] = E[0]*g, and E[0] and E[1] are sorted lexically by the
// ordering on MatGF.  In the context of the current Path, for each of
// these we may need the inverse of g instead.
func (p Path) Word(lsv *LsvContext) (word []MatGF, vertices []MatGF) {
	cur := p.Start(lsv)
	word = make([]MatGF, len(p))
	vertices = make([]MatGF, len(p)+1)
	vertices[0] = cur
	for i, edge := range p {
		g := edge.Generator(lsv)
		if cur.Equal(lsv, &edge[0]) {
			word[i] = *g
			cur = edge[1]
		} else {
			word[i] = *g.Inverse(lsv)
			cur = edge[0]
		}
		vertices[i+1] = cur
	}
	return
}

type Triangle [3]Vertex

func NewTriangle(a, b, c Vertex) Triangle {
	if a.Less(&b) {
		if b.Less(&c) {
			return Triangle{a, b, c}
		} else if a.Less(&c) {
			return Triangle{a, c, b}
		} else {
			return Triangle{c, a, b}
		}
	} else {
		if a.Less(&c) {
			return Triangle{b, a, c}
		} else if b.Less(&c) {
			return Triangle{b, c, a}
		} else {
			return Triangle{c, b, a}
		}
	}
}

func NewTriangleFromString(s string) Triangle {
	entries := strings.Split(strings.Trim(s, " []"), "] [")
	if len(entries) != 3 {
		panic(fmt.Sprintf("invalid triangle string: %s", s))
	}
	a := NewMatGFFromString(entries[0])
	b := NewMatGFFromString(entries[1])
	c := NewMatGFFromString(entries[2])
	return NewTriangle(a, b, c)
}

func (T Triangle) Edges() [3]Edge {
	return [3]Edge{NewEdge(T[0], T[1]), NewEdge(T[1], T[2]), NewEdge(T[2], T[0])}
}

// not currently used
// func (T Triangle) Path() Path {
// 	edges := T.Edges()
// 	return Path{edges[0], edges[1], edges[2]}
// }

func (T Triangle) OrbitContainsEdge(lsv *LsvContext, e Edge) (g *MatGF, found bool) {
	for _, f := range T.Edges() {
		if g, ok := f.OrbitContainsEdge(lsv, e); ok {
			// log.Printf("found edge %v in orbit of %v via %v", e, f, g)
			return g, true
		}
	}		
	return nil, false
}

func (T Triangle) OrbitContainsTriangle(lsv *LsvContext, t Triangle) (g *MatGF, found bool) {
	for _, e := range T.Edges() {
		g, ok := T.OrbitContainsEdge(lsv, e)
		if ok {
			U := T.Translate(lsv, g)
			if U.Equal(lsv, &t) {
				return g, true
			}
		}
	}
	return nil, false
}

func (T Triangle) Translate(lsv *LsvContext, g *MatGF) Triangle {
	return NewTriangle(*T[0].Translate(lsv, g), *T[1].Translate(lsv, g), *T[2].Translate(lsv, g))
}

// xxx test
func (T Triangle) Less(U Triangle) bool {
	if T[0].Less(&U[0]) {
		return true
	} else if U[0].Less(&T[0]) {
		return false
	} else if T[1].Less(&U[1]) {
		return true
	} else if U[1].Less(&T[1]) {
		return false
	} else {
		return T[2].Less(&U[2])
	}
}

// xxx test
func (T Triangle) Equal(lsv *LsvContext, u *Triangle) bool {
	return T[0].Equal(lsv, &(*u)[0]) && T[1].Equal(lsv, &(*u)[1]) && T[2].Equal(lsv, &(*u)[2])
}

// xxx test
func (T Triangle) SharedEdges(lsv *LsvContext, U Triangle) []Edge {
	set := make(map[Edge]any)
	TEdges := T.Edges()
	UEdges := U.Edges()
	for _, e := range TEdges {
		for _, f := range UEdges {
			if e.Equal(lsv, f) {
				set[e] = nil
			}
		}
	}
	edges := make([]Edge, 0, len(set))
	for e := range set {
		edges = append(edges, e)
	}
	return edges
}

func (T Triangle) String() string {
	return fmt.Sprintf("[%s %s %s]", T[0], T[1], T[2])
}


type Complex struct {
    vertexBasis   []Vertex
	edgeBasis     []Edge
	triangleBasis []Triangle
	d1            BinaryMatrix
	d2 		      BinaryMatrix
	d2Transpose   BinaryMatrix
	EdgeIndex     map[Edge]int   // edge -> index in basis
	VertexIndex   map[Vertex]int // vertex -> index in basis
	verbose       bool
}

func NewComplex(vertexBasis []Vertex, edgeBasis []Edge, triangleBasis []Triangle, sortBases bool, verbose bool) *Complex {
	if sortBases {
		sort.Slice(triangleBasis, func(i, j int) bool {
			return triangleBasis[i].Less(triangleBasis[j])
		})
		sort.Slice(edgeBasis, func(i, j int) bool {
			return edgeBasis[i].Less(edgeBasis[j])
		})
		sort.Slice(vertexBasis, func(i, j int) bool {
			return vertexBasis[i].Less(&vertexBasis[j])
		})
	}
	C := Complex{
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

// xxx test
func ComplexFromTriangles(triangles []Triangle, sortBases bool, verbose bool) *Complex {
	triangleSet := make(map[Triangle]any)
	edgeSet := make(map[Edge]any)
	vertexSet := make(map[Vertex]any)
	for _, T := range triangles {
		triangleSet[T] = nil
	}
	triangleBasis := make([]Triangle, len(triangleSet))
	i := 0
	for T := range triangleSet {
		triangleBasis[i] = T
		i++
		for _, v := range T {
			vertexSet[v] = nil
		}
		for _, e := range T.Edges() {
			edgeSet[e] = nil
		}
	}
	edgeBasis := make([]Edge, len(edgeSet))
	i = 0
	for e := range edgeSet {
		edgeBasis[i] = e
		i++
	}
	vertexBasis := make([]Vertex, len(vertexSet))
	i = 0
	for v := range vertexSet {
		vertexBasis[i] = v
		i++
	}
	return NewComplex(vertexBasis, edgeBasis, triangleBasis, sortBases, verbose)
}

func (C *Complex) D1() BinaryMatrix {
	if C.d1 == nil {
		C.computeD1()
	}
	return C.d1
}

func (C *Complex) D2() BinaryMatrix {
	if C.d2 == nil {
		C.computeD2()
	}
	return C.d2
}

// xxx possibly deprecate
func (C *Complex) D2Transpose() BinaryMatrix {
	if C.d2Transpose == nil {
		C.computeD2Transpose()
	}
	return C.d2Transpose
}

func (C *Complex) EdgeBasis() []Edge {
	return C.edgeBasis
}

func (C *Complex) String() string {
	return fmt.Sprintf("complex with %v triangles, %v edges, %v vertices",
		len(C.triangleBasis), len(C.edgeBasis), len(C.vertexBasis))
}

func (C *Complex) DumpBases() (s string) {
	s += "vertexBasis:\n"
	for i, v := range C.vertexBasis {
		s += fmt.Sprintf("  %d: %v\n", i, v)
	}
	s += "edgeBasis:\n"
	for i, e := range C.edgeBasis {
		u, ok := C.VertexIndex[e[0]]
		if !ok {
			panic("vertex in edge not found in vertex index")
		}
		v, ok := C.VertexIndex[e[1]]
		if !ok {
			panic("vertex in edge not found in vertex index")
		}
		s += fmt.Sprintf("  %d: [%d %d]\n", i, u, v)
	}
	s += "triangleBasis:\n"
	for i, t := range C.triangleBasis {
		s += fmt.Sprintf("  %d: [%d %d %d]\n", i, C.VertexIndex[t[0]], C.VertexIndex[t[1]], C.VertexIndex[t[2]])
	}		
	return s
}

func (C *Complex) NumVertices() int {
	return len(C.vertexBasis)
}

func (C *Complex) NumEdges() int {
	return len(C.edgeBasis)
}

func (C *Complex) NumTriangles() int {
	return len(C.triangleBasis)
}

func (C *Complex) computeEdgeIndex() {
	C.EdgeIndex = make(map[Edge]int)
	for i, edge := range C.edgeBasis {
		C.EdgeIndex[edge] = i
	}
}

func (C *Complex) computeVertexIndex() {
	C.VertexIndex = make(map[Vertex]int)
	for i, vertex := range C.vertexBasis {
		C.VertexIndex[vertex] = i
	}
}

func (C *Complex) computeD1() {
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

func (C *Complex) computeD2() {
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

func (C *Complex) computeD2Transpose() {
	C.d2Transpose = NewSparseBinaryMatrix(len(C.triangleBasis), len(C.edgeBasis))
	for j, t := range C.triangleBasis {
		for _, e := range t.Edges() {
			row, ok := C.EdgeIndex[e]
			if !ok {
				panic(fmt.Sprintf("edge %v not in edge index", e))
			}
			C.d2Transpose.Set(j, row, 1)
		}
	}
}

func (C *Complex) PathToEdgeVector(path Path) BinaryVector {
	v := NewBinaryVector(len(C.edgeBasis))
	for _, e := range path {
		if i, ok := C.EdgeIndex[e]; ok {
			v.Set(i, 1)
		} else {
			panic(fmt.Sprintf("edge %v not in edge index", e))
		}
	}
	return v
}

func (C *Complex) EdgeVectorToPath(lsv *LsvContext, v BinaryMatrix) Path {
	if v.NumColumns() != 1 {
		panic("edge vector must be a column vector")
	}
	edges := make([]Edge, 0)
	for i := 0; i < v.NumRows(); i++ {
		if v.Get(i, 0) == 1 {
			edges = append(edges, C.edgeBasis[i])
		}
	}
	return NewPathFromEdges(lsv, edges)
}

func (C *Complex) TriangleBasis() []Triangle {
	return C.triangleBasis
}

func (C *Complex) VertexBasis() []Vertex {
	return C.vertexBasis
}

func WriteStringFile[S fmt.Stringer] (items []S, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()
	for i := 0; i < len(items); i++ {
		_, err = f.WriteString(items[i].String() + "\n")
		if err != nil {
			panic(err)
		}
	}
}

func ReadStringFile[T any] (filename string, fromString func(string) T) []T {
	objects := make([]T, 0)
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		object := fromString(strings.TrimSpace(line))
		objects = append(objects, object)
	}
	return objects
}

func WriteVertexFile(vertices []Vertex, filename string) {
	WriteStringFile(vertices, filename)
}

func ReadVertexFile(filename string) (vertices []Vertex) {
	return ReadStringFile(filename, NewVertexFromString)
}

func WriteEdgeFile(edges []Edge, filename string) {
	WriteStringFile(edges, filename)
}

func ReadEdgeFile(filename string) (edges []Edge) {
	return ReadStringFile(filename, NewEdgeFromString)
}

func WriteTriangleFile(triangles []Triangle, filename string) {
	WriteStringFile(triangles, filename)
}

func ReadTriangleFile(filename string) (triangles []Triangle) {
	return ReadStringFile(filename, NewTriangleFromString)
}

