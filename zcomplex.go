package golsv

import (
	"bytes"
	"fmt"
	"log"
	"sort"
)

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

