package golsv

import (
	"testing"
)

func TestCalGCayleyExpanderInverse(t *testing.T) {
	gens := CartwrightStegerGenerators()
	maxDepth := 2
	verbose := false
	var modulus *F2Polynomial = nil
	quotient := false
	E := NewCalGCayleyExpander(gens, maxDepth, verbose, modulus, quotient, nil)
	E.Graph()

	pathLenCounts := make([]int, maxDepth+1)
	
	for g, _ := range E.attendance {
		gInv, gInvPath := E.elementInverse(g)
		p := NewElementCalGIdentity()
		p.Mul(g, gInv)
		if !p.IsIdentity() {
			t.Errorf("Inverse failed;\ng: %v\ngInv: %v\np: %v", g, gInv, p)
		}
		n := len(gInvPath)
		pathLenCounts[n]++
		if g.IsIdentity() {
			continue
		}
		end := gInvPath[len(gInvPath)-1]
		if !end.Contains(NewElementCalGIdentity()) {
			t.Errorf("Inverse path end does not contain identity: g: %v\nend: %v", g, end)
		}
	}
	if pathLenCounts[0] != 1 {
		t.Errorf("Inverse path length 0 count not 1: %v", pathLenCounts[0])
	}
	if pathLenCounts[1] != 14 {
		t.Errorf("Inverse path length 1 count not 14: %v", pathLenCounts[1])
	}
	if pathLenCounts[2] != 98 {
		t.Errorf("Inverse path length 2 count not 98: %v", pathLenCounts[2])
	}
}

func TestCalGCayleyExpanderInverseModf(t *testing.T) {
	gens := CartwrightStegerGenerators()
	maxDepth := 2
	verbose := false
	var modulus F2Polynomial = NewF2Polynomial("111")
	quotient := true
	E := NewCalGCayleyExpander(gens, maxDepth, verbose, &modulus, quotient, nil)
	E.Graph()

	pathLenCounts := make([]int, maxDepth+1)

	for g, _ := range E.attendance {
		gInv, gInvPath := E.elementInverse(g)
		p := NewElementCalGIdentity()
		p.Mul(g, gInv)
		if !p.IsIdentityModf(modulus) {
			t.Errorf("Inverse failed;\ng: %v\ngInv: %v\np: %v", g, gInv, p)
		}
		n := len(gInvPath)
		pathLenCounts[n]++
		if g.IsIdentity() {
			continue
		}
		end := gInvPath[len(gInvPath)-1]
		if !end.Contains(NewElementCalGIdentity()) {
			t.Errorf("Inverse path end does not contain identity: g: %v\nend: %v", g, end)
		}
	}
	if pathLenCounts[0] != 1 {
		t.Errorf("Inverse path length 0 count not 1: %v", pathLenCounts[0])
	}
	if pathLenCounts[1] != 14 {
		t.Errorf("Inverse path length 1 count not 14: %v", pathLenCounts[1])
	}
	if pathLenCounts[2] != 98 {
		t.Errorf("Inverse path length 2 count not 98: %v", pathLenCounts[2])
	}
}

func TestCalGCayleyExpanderComplex(t *testing.T) {
	gens := CartwrightStegerGenerators()
	maxDepth := 1
	verbose := false
	var modulus *F2Polynomial = nil
	quotient := false
	E := NewCalGCayleyExpander(gens, maxDepth, verbose, modulus, quotient, nil)
	graph := E.Graph()
	vertices := graph.VertexBasis()
	expectedVertices := 15
	if len(vertices) != expectedVertices {
		t.Errorf("Vertices: got=%d expected=%d", len(vertices), expectedVertices)
	}
	edges := graph.EdgeBasis()
	expectedEdges := 17
	if len(edges) != expectedEdges {
		t.Errorf("Edges: got=%d expected=%d", len(edges), expectedEdges)
	}
	C := E.Complex()
	triangles := C.TriangleBasis()
	expectedTriangles := 3
	if len(triangles) != expectedTriangles {
		t.Errorf("Triangles: got=%d expected=%d", len(triangles), expectedTriangles)
	}
}

func TestCalGCayleyExpanderEdgeOrder(t *testing.T) {
	gens := CartwrightStegerGenerators()
	maxDepth := 2
	verbose := false
	var modulus F2Polynomial = NewF2Polynomial("111")
	quotient := true
	var observer CalGObserver
	E := NewCalGCayleyExpander(gens, maxDepth, verbose, &modulus, quotient, observer)
	graph := E.Graph()
	edges := graph.EdgeBasis()

	for i, e := range edges {
		if i == len(edges)-1 {
			break
		}
		f := edges[i+1]
		if !E.edgeLessByVertexAttendance(e, f) {
			t.Errorf("Edge order failed: e=%v f=%v", e, f)
		}
	}
}

func TestNewEdgeElementCalGFromString(t *testing.T) {
	tests := []struct {
		s string
		want ZEdge[ElementCalG]
	}{
		{
			"[(1,0,0)(1,0,0)(1,0,0) (1,0,0)(1,0,01)(1,0,0)]",
			NewZEdge(
				ZVertex[ElementCalG](NewElementCalGFromString("(1,0,0)(1,0,0)(1,0,0)")),
				ZVertex[ElementCalG](NewElementCalGFromString("(1,0,0)(1,0,01)(1,0,0)"))),
		},
	}
	for i, test := range tests {
		e := NewEdgeElementCalGFromString(test.s)
		if !e.Equal(test.want) {
			t.Errorf("Test %d: got=%v expected=%v", i, e, test.want)
		}
	}
}
