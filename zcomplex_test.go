package golsv

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestZEdgeContains(t *testing.T) {
	tests := []struct {
		E        ZEdge[ZVertexInt]
		V        ZVertex[ZVertexInt]
		Expected bool
	}{
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), ZVertexInt(0), true},
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), ZVertexInt(1), true},
		{NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(0)), ZVertexInt(1), true},
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), ZVertexInt(2), false},
	}
	for n, test := range tests {
		got := test.E.Contains(test.V)
		if got != test.Expected {
			t.Errorf("Test %d: E.Contains(V)=%v, expected %v", n, got, test.Expected)
		}
	}
}

func TestZEdgeEqual(t *testing.T) {
	tests := []struct {
		A        ZEdge[ZVertexInt]
		B        ZEdge[ZVertexInt]
		Expected bool
	}{
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), true},
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(0)), true},
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(2)), false},
	}
	for n, test := range tests {
		got := test.A.Equal(test.B)
		if got != test.Expected {
			t.Errorf("Test %d: A.Equal(B)=%v, expected %v", n, got, test.Expected)
		}
	}
}

func TestZEdgeOtherVertex(t *testing.T) {
	tests := []struct {
		E        ZEdge[ZVertexInt]
		V        ZVertex[ZVertexInt]
		Expected ZVertex[ZVertexInt]
	}{
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), ZVertexInt(0), ZVertexInt(1)},
		{NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)), ZVertexInt(1), ZVertexInt(0)},
	}
	for n, test := range tests {
		got := test.E.OtherVertex(test.V)
		if got != test.Expected {
			t.Errorf("Test %d: E.OtherVertex(V)=%v, expected %v", n, got, test.Expected)
		}
	}
}

func TestZTriangleMapKey(t *testing.T) {
	a := NewElementCalGFromString("(11,0,1)(101,1,1)(101,1,0)")
	b := NewElementCalGFromString("(111,0,1)(101,1,1)(101,1,0)")
	c := NewElementCalGFromString("(111,0,1)(111,1,1)(111,1,0)")

	T := NewZTriangle[ElementCalG](a, b, c)
	U := NewZTriangle[ElementCalG](b, c, a)
	if !T.Equal(U) {
		t.Error("T != U")
	}
	m := make(map[ZTriangle[ElementCalG]]int)
	m[T] = 1
	m[U] = 2
	if len(m) != 1 {
		t.Error("len(m) != 1")
	}
}

func TestZTriangleContainsVertex(t *testing.T) {
	T := ZTriangle[ZVertexInt]{ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)}
	tests := []struct {
		T        ZTriangle[ZVertexInt]
		V        ZVertex[ZVertexInt]
		Expected bool
	}{
		{T, ZVertexInt(0), true},
		{T, ZVertexInt(1), true},
		{T, ZVertexInt(2), true},
		{T, ZVertexInt(3), false},
	}
	for n, test := range tests {
		got := test.T.ContainsVertex(test.V)
		if got != test.Expected {
			t.Errorf("Test %d: T.ContainsVertex(V)=%v, expected %v", n, got, test.Expected)
		}
	}
}

func TestZTriangleSetEqual(t *testing.T) {
	U := NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2))
	V := NewZTriangle[ZVertexInt](ZVertexInt(3), ZVertexInt(1), ZVertexInt(0))
	tests := []struct {
		A        []ZTriangle[ZVertexInt]
		B        []ZTriangle[ZVertexInt]
		Expected bool
	}{
		{[]ZTriangle[ZVertexInt]{}, []ZTriangle[ZVertexInt]{}, true},
		{[]ZTriangle[ZVertexInt]{}, []ZTriangle[ZVertexInt]{U}, false},
		{[]ZTriangle[ZVertexInt]{U}, []ZTriangle[ZVertexInt]{U}, true},
		{[]ZTriangle[ZVertexInt]{U}, []ZTriangle[ZVertexInt]{V}, false},
		{[]ZTriangle[ZVertexInt]{U, V}, []ZTriangle[ZVertexInt]{V, U}, true},
	}
	for n, test := range tests {
		got := TriangleSetEqual(test.A, test.B)
		if got != test.Expected {
			t.Errorf("Test %d: TriangleSetEqual(A, B)=%v, expected %v", n, got, test.Expected)
		}
	}
}

func TestZComplexDepthFiltration(t *testing.T) {
	type data struct {
		Depth           int
		Subcomplex      *ZComplex[ZVertexInt]
		VerticesAtDepth []ZVertex[ZVertexInt]
	}
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		Expected []data
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			[]data{
				{0, NewZComplexFromMaximalSimplices([][]int{{0}}), []ZVertex[ZVertexInt]{ZVertexInt(0)}},
				{1, NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}), []ZVertex[ZVertexInt]{ZVertexInt(1), ZVertexInt(2)}},
			},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2, 3}}),
			[]data{
				{0, NewZComplexFromMaximalSimplices([][]int{{0}}), []ZVertex[ZVertexInt]{ZVertexInt(0)}},
				{1, NewZComplexFromMaximalSimplices([][]int{{0, 1}}), []ZVertex[ZVertexInt]{ZVertexInt(1)}},
				{2, NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2, 3}}), []ZVertex[ZVertexInt]{ZVertexInt(2), ZVertexInt(3)}},
			},
		},
	}
	for n, test := range tests {
		got := make([]data, 0)
		initialVertex := test.C.VertexBasis()[0]
		test.C.DepthFiltration(initialVertex, func(depth int, subcomplex *ZComplex[ZVertexInt], verticesAtDepth []ZVertex[ZVertexInt]) {
			got = append(got, data{
				Depth:           depth,
				Subcomplex:      subcomplex,
				VerticesAtDepth: append([]ZVertex[ZVertexInt](nil), verticesAtDepth...),
			})
		})
		if len(got) != len(test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
			continue
		}
		for i := range got {
			if !reflect.DeepEqual(got[i], test.Expected[i]) {
				t.Errorf("Test %d: got=%v, expected=%v", n, got[i], test.Expected[i])
			}
		}
	}
}

func TestZComplexEdgeToTriangleIncidenceMap(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		Expected map[int][]int
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			map[int][]int{
				0: {0},
				1: {0},
				2: {0},
			},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}}),
			map[int][]int{
				0: {0, 1},
				1: {0},
				2: {1},
				3: {0},
				4: {1},
			},
		},
	}
	for n, test := range tests {
		got := test.C.EdgeToTriangleIncidenceMap()
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexDFS(t *testing.T) {
	type vdata struct {
		vertex int
		depth  int
	}
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		start    int
		Expected []vdata
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0}}),
			0,
			[]vdata{{0, 0}},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1}}),
			0,
			[]vdata{{0, 0}, {1, 1}},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			0,
			[]vdata{{0, 0}, {1, 1}, {2, 2}},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 3}}),
			0,
			[]vdata{{0, 0}, {1, 1}, {2, 2}, {3, 2}},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 3}, {2, 4}}),
			0,
			[]vdata{{0, 0}, {1, 1}, {2, 2}, {4, 3}, {3, 2}},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 3}, {2, 4}, {2, 5}}),
			0,
			[]vdata{{0, 0}, {1, 1}, {2, 2}, {4, 3}, {5, 3}, {3, 2}},
		},
	}
	for n, test := range tests {
		got := make([]vdata, 0)
		test.C.DFS(ZVertexInt(test.start), func(vertex ZVertex[ZVertexInt], depth int) (stop bool) {
			v, ok := vertex.(ZVertexInt)
			if !ok {
				t.Errorf("test %d: vertex is not of type ZVertexInt", n)
				return true
			}
			got = append(got, vdata{int(v), depth})
			return false
		})
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexNeighbors(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		v        int
		Expected []int
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			0,
			[]int{1, 2},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2, 3}}),
			0,
			[]int{1, 2},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2, 3}}),
			1,
			[]int{0, 2, 3},
		},
	}
	for n, test := range tests {
		got := test.C.Neighbors(test.v)
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: C.Neighbors(v)=%v, expected %v", n, got, test.Expected)
		}
	}
}

func TestZComplexAddEdge(t *testing.T) {
	tests := []struct {
		name     string
		initial  [][]int // maximal simplices for initial complex
		u, v     int     // vertices to connect
		expected [][]int // expected neighbors after adding edge
	}{
		{
			name:    "add edge to empty graph",
			initial: [][]int{},
			u:       0,
			v:       1,
			expected: [][]int{
				{1},    // neighbors of vertex 0
				{0},    // neighbors of vertex 1
			},
		},
		{
			name:    "add edge to existing triangle",
			initial: [][]int{{0, 1, 2}, {3}}, // include vertex 3 in complex
			u:       0,
			v:       3,
			expected: [][]int{
				{1, 2, 3}, // neighbors of vertex 0
				{0, 2},    // neighbors of vertex 1
				{0, 1},    // neighbors of vertex 2
				{0},       // neighbors of vertex 3
			},
		},
		{
			name:    "add edge between existing vertices",
			initial: [][]int{{0, 1}, {2, 3}},
			u:       1,
			v:       2,
			expected: [][]int{
				{1},    // neighbors of vertex 0
				{0, 2}, // neighbors of vertex 1
				{1, 3}, // neighbors of vertex 2
				{2},    // neighbors of vertex 3
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var C *ZComplex[ZVertexInt]
			if len(test.initial) == 0 {
				// Create empty complex with just vertices
				vertices := []ZVertex[ZVertexInt]{ZVertexInt(0), ZVertexInt(1), ZVertexInt(2), ZVertexInt(3)}
				C = NewZComplex(vertices, nil, nil, false, false)
			} else {
				C = NewZComplexFromMaximalSimplices(test.initial)
			}

			initialEdgeCount := C.NumEdges()
			C.AddEdge(test.u, test.v)

			// Check edge was added
			if C.NumEdges() != initialEdgeCount+1 {
				t.Errorf("Expected %d edges, got %d", initialEdgeCount+1, C.NumEdges())
			}

			// Check neighbors are correct
			for vertex, expectedNeighbors := range test.expected {
				if vertex >= len(C.vertexBasis) {
					continue
				}
				neighbors := C.Neighbors(vertex)
				if !equalIntSlices(neighbors, expectedNeighbors) {
					t.Errorf("Vertex %d: expected neighbors %v, got %v", vertex, expectedNeighbors, neighbors)
				}
			}

			// Check that vertices are mutual neighbors
			if !C.IsNeighbor(test.u, test.v) {
				t.Errorf("Vertices %d and %d should be neighbors", test.u, test.v)
			}
			if !C.IsNeighbor(test.v, test.u) {
				t.Errorf("Vertices %d and %d should be neighbors", test.v, test.u)
			}
		})
	}
}

func TestZComplexAddEdgeDuplicateHandling(t *testing.T) {
	C := NewZComplexFromMaximalSimplices([][]int{{0, 1}})
	initialEdgeCount := C.NumEdges()

	// Adding the same edge should not increase edge count
	C.AddEdge(0, 1)
	if C.NumEdges() != initialEdgeCount {
		t.Errorf("Expected %d edges after adding duplicate, got %d", initialEdgeCount, C.NumEdges())
	}

	// Adding reverse direction should not increase edge count
	C.AddEdge(1, 0)
	if C.NumEdges() != initialEdgeCount {
		t.Errorf("Expected %d edges after adding reverse duplicate, got %d", initialEdgeCount, C.NumEdges())
	}
}

func TestZComplexAddEdgePanics(t *testing.T) {
	C := NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}})

	// Test self-loop panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for self-loop")
		}
	}()
	C.AddEdge(0, 0)
}

// Helper function to compare int slices ignoring order
func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[int]int)
	for _, x := range a {
		counts[x]++
	}
	for _, x := range b {
		counts[x]--
		if counts[x] < 0 {
			return false
		}
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func TestZComplexDegree(t *testing.T) {
	tests := []struct {
		name     string
		initial  [][]int // maximal simplices for initial complex
		vertex   int     // vertex to check degree
		expected int     // expected degree
	}{
		{
			name:     "isolated vertex",
			initial:  [][]int{{0}, {1, 2}}, // vertex 0 is isolated
			vertex:   0,
			expected: 0,
		},
		{
			name:     "vertex in edge",
			initial:  [][]int{{0, 1}},
			vertex:   0,
			expected: 1,
		},
		{
			name:     "vertex in triangle",
			initial:  [][]int{{0, 1, 2}},
			vertex:   0,
			expected: 2,
		},
		{
			name:     "vertex with multiple connections",
			initial:  [][]int{{0, 1}, {0, 2}, {0, 3}},
			vertex:   0,
			expected: 3,
		},
		{
			name:     "vertex in complex graph",
			initial:  [][]int{{0, 1}, {1, 2}, {2, 3}, {0, 3}}, // square
			vertex:   1,
			expected: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			C := NewZComplexFromMaximalSimplices(test.initial)
			degree := C.Degree(test.vertex)
			if degree != test.expected {
				t.Errorf("Expected degree %d for vertex %d, got %d", test.expected, test.vertex, degree)
			}
		})
	}
}

func TestZComplexIsNeighbor(t *testing.T) {
	tests := []struct {
		name     string
		initial  [][]int // maximal simplices for initial complex
		u, v     int     // vertices to check
		expected bool    // expected result
	}{
		{
			name:     "isolated vertices",
			initial:  [][]int{{0}, {1}}, // both isolated
			u:        0,
			v:        1,
			expected: false,
		},
		{
			name:     "connected by edge",
			initial:  [][]int{{0, 1}},
			u:        0,
			v:        1,
			expected: true,
		},
		{
			name:     "connected by edge (reverse)",
			initial:  [][]int{{0, 1}},
			u:        1,
			v:        0,
			expected: true,
		},
		{
			name:     "connected in triangle",
			initial:  [][]int{{0, 1, 2}},
			u:        0,
			v:        1,
			expected: true,
		},
		{
			name:     "not connected in triangle",
			initial:  [][]int{{0, 1, 2}, {3}},
			u:        0,
			v:        3,
			expected: false,
		},
		{
			name:     "connected through multiple paths",
			initial:  [][]int{{0, 1}, {0, 2}, {1, 2}}, // triangle
			u:        1,
			v:        2,
			expected: true,
		},
		{
			name:     "self-check",
			initial:  [][]int{{0, 1}},
			u:        0,
			v:        0,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			C := NewZComplexFromMaximalSimplices(test.initial)
			result := C.IsNeighbor(test.u, test.v)
			if result != test.expected {
				t.Errorf("Expected IsNeighbor(%d, %d) = %v, got %v", test.u, test.v, test.expected, result)
			}
		})
	}
}

func TestZComplexIsRegular(t *testing.T) {
	tests := []struct {
		name     string
		initial  [][]int // maximal simplices for initial complex
		expected bool    // expected regularity
	}{
		{
			name:     "empty graph",
			initial:  [][]int{},
			expected: true,
		},
		{
			name:     "single vertex",
			initial:  [][]int{{0}},
			expected: true,
		},
		{
			name:     "two isolated vertices",
			initial:  [][]int{{0}, {1}},
			expected: true, // both have degree 0
		},
		{
			name:     "single edge (2-regular)",
			initial:  [][]int{{0, 1}},
			expected: true, // both vertices have degree 1
		},
		{
			name:     "triangle (2-regular)",
			initial:  [][]int{{0, 1, 2}},
			expected: true, // all vertices have degree 2
		},
		{
			name:     "irregular graph",
			initial:  [][]int{{0, 1}, {1, 2}}, // vertex 1 has degree 2, others have degree 1
			expected: false,
		},
		{
			name:     "square (2-regular)",
			initial:  [][]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}},
			expected: true, // all vertices have degree 2
		},
		{
			name:     "star graph (irregular)",
			initial:  [][]int{{0, 1}, {0, 2}, {0, 3}}, // vertex 0 has degree 3, others have degree 1
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var C *ZComplex[ZVertexInt]
			if len(test.initial) == 0 {
				// Create empty complex
				vertices := []ZVertex[ZVertexInt]{}
				C = NewZComplex(vertices, nil, nil, false, false)
			} else {
				C = NewZComplexFromMaximalSimplices(test.initial)
			}
			
			result, _ := C.IsRegular()
			if result != test.expected {
				t.Errorf("Expected IsRegular() = %v, got %v", test.expected, result)
			}
		})
	}
}

func TestZComplexMoveEdge(t *testing.T) {
	tests := []struct {
		name        string
		initial     [][]int // maximal simplices for initial complex
		edgeToMove  [2]int  // vertices of edge to move
		newVertices [2]int  // new vertices for the edge
		expected    [][]int // expected neighbors after moving edge
	}{
		{
			name:        "move edge in simple graph",
			initial:     [][]int{{0, 1}, {2, 3}}, // two separate edges
			edgeToMove:  [2]int{0, 1},
			newVertices: [2]int{1, 2}, // move edge (0,1) to (1,2)
			expected: [][]int{
				{},      // neighbors of vertex 0
				{2},     // neighbors of vertex 1
				{1, 3},  // neighbors of vertex 2
				{2},     // neighbors of vertex 3
			},
		},
		{
			name:        "move edge in connected graph",
			initial:     [][]int{{0, 1}, {1, 2}, {2, 3}}, // path graph
			edgeToMove:  [2]int{1, 2},
			newVertices: [2]int{0, 3}, // move middle edge to connect endpoints
			expected: [][]int{
				{1, 3},  // neighbors of vertex 0
				{0},     // neighbors of vertex 1
				{3},     // neighbors of vertex 2
				{0, 2},  // neighbors of vertex 3
			},
		},
		{
			name:        "move edge to create different topology",
			initial:     [][]int{{0, 1}, {0, 2}, {1, 2}}, // triangle edges
			edgeToMove:  [2]int{0, 1},
			newVertices: [2]int{0, 3}, // move edge to new vertex
			expected: [][]int{
				{2, 3},  // neighbors of vertex 0
				{2},     // neighbors of vertex 1
				{0, 1},  // neighbors of vertex 2
				{0},     // neighbors of vertex 3
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Find the maximum vertex needed for this test
			maxVertex := 0
			for _, edge := range test.initial {
				for _, v := range edge {
					if v > maxVertex {
						maxVertex = v
					}
				}
			}
			for _, v := range test.edgeToMove {
				if v > maxVertex {
					maxVertex = v
				}
			}
			for _, v := range test.newVertices {
				if v > maxVertex {
					maxVertex = v
				}
			}
			
			// Create all vertices we need
			vertices := []ZVertex[ZVertexInt]{}
			for i := 0; i <= maxVertex; i++ {
				vertices = append(vertices, ZVertexInt(i))
			}
			
			// Create edges from test.initial
			edges := []ZEdge[ZVertexInt]{}
			for _, edge := range test.initial {
				if len(edge) == 2 {
					edges = append(edges, NewZEdge(ZVertexInt(edge[0]), ZVertexInt(edge[1])))
				}
			}
			
			// Create the complex with vertices and edges
			C := NewZComplex(vertices, edges, nil, false, false)

			initialEdgeCount := C.NumEdges()

			// Find the edge to move
			edgeIdx, ok := C.IndexOfEdge(test.edgeToMove[0], test.edgeToMove[1])
			if !ok {
				t.Fatalf("Edge (%d,%d) not found for moving", test.edgeToMove[0], test.edgeToMove[1])
			}

			// Verify original edge exists
			if !C.IsNeighbor(test.edgeToMove[0], test.edgeToMove[1]) {
				t.Fatalf("Original edge (%d,%d) should exist", test.edgeToMove[0], test.edgeToMove[1])
			}

			// Move the edge
			C.MoveEdge(edgeIdx, test.newVertices[0], test.newVertices[1])

			// Verify edge count remains the same
			if C.NumEdges() != initialEdgeCount {
				t.Errorf("Expected %d edges after moving, got %d", initialEdgeCount, C.NumEdges())
			}

			// Verify original edge no longer exists
			if C.IsNeighbor(test.edgeToMove[0], test.edgeToMove[1]) {
				t.Errorf("Original edge (%d,%d) should no longer exist", test.edgeToMove[0], test.edgeToMove[1])
			}

			// Verify new edge exists
			if !C.IsNeighbor(test.newVertices[0], test.newVertices[1]) {
				t.Errorf("New edge (%d,%d) should exist", test.newVertices[0], test.newVertices[1])
			}

			// Check all neighbors are as expected
			for vertex, expectedNeighbors := range test.expected {
				if vertex >= len(C.vertexBasis) {
					continue
				}
				neighbors := C.Neighbors(vertex)
				if !equalIntSlices(neighbors, expectedNeighbors) {
					t.Errorf("Vertex %d: expected neighbors %v, got %v", vertex, expectedNeighbors, neighbors)
				}
			}

			// Verify edge index is correctly updated
			newEdgeIdx, ok := C.IndexOfEdge(test.newVertices[0], test.newVertices[1])
			if !ok {
				t.Errorf("New edge (%d,%d) not found in edge index", test.newVertices[0], test.newVertices[1])
			}
			if newEdgeIdx != edgeIdx {
				t.Errorf("Expected moved edge to have same index %d, got %d", edgeIdx, newEdgeIdx)
			}

			// Verify old edge is not in index
			_, ok = C.IndexOfEdge(test.edgeToMove[0], test.edgeToMove[1])
			if ok {
				t.Errorf("Old edge (%d,%d) should not be in edge index after moving", test.edgeToMove[0], test.edgeToMove[1])
			}
		})
	}
}

func TestZComplexDeleteEdge(t *testing.T) {
	tests := []struct {
		name     string
		initial  [][]int // maximal simplices for initial complex
		u, v     int     // vertices to disconnect
		expected [][]int // expected neighbors after deleting edge
	}{
		{
			name:    "delete edge from graph",
			initial: [][]int{{0, 1}, {0, 2}}, // edges only, no triangles
			u:       0,
			v:       1,
			expected: [][]int{
				{2}, // neighbors of vertex 0
				{},  // neighbors of vertex 1
				{0}, // neighbors of vertex 2
			},
		},
		{
			name:    "delete edge from disconnected components",
			initial: [][]int{{0, 1}, {2, 3}},
			u:       0,
			v:       1,
			expected: [][]int{
				{},  // neighbors of vertex 0
				{},  // neighbors of vertex 1
				{3}, // neighbors of vertex 2
				{2}, // neighbors of vertex 3
			},
		},
		{
			name:    "delete edge from complex graph",
			initial: [][]int{{0, 1}, {1, 2}, {2, 3}, {0, 3}},
			u:       1,
			v:       2,
			expected: [][]int{
				{1, 3}, // neighbors of vertex 0
				{0},    // neighbors of vertex 1
				{3},    // neighbors of vertex 2
				{0, 2}, // neighbors of vertex 3
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			C := NewZComplexFromMaximalSimplices(test.initial)
			initialEdgeCount := C.NumEdges()

			// Verify edge exists before deletion
			if !C.IsNeighbor(test.u, test.v) {
				t.Fatalf("Edge (%d,%d) should exist before deletion", test.u, test.v)
			}

			// Find edge index
			edgeIdx, ok := C.IndexOfEdge(test.u, test.v)
			if !ok {
				t.Fatalf("Edge (%d,%d) not found in edge index", test.u, test.v)
			}

			C.DeleteEdge(edgeIdx)

			// Check edge was removed
			if C.NumEdges() != initialEdgeCount-1 {
				t.Errorf("Expected %d edges, got %d", initialEdgeCount-1, C.NumEdges())
			}

			// Check neighbors are correct
			for vertex, expectedNeighbors := range test.expected {
				if vertex >= len(C.vertexBasis) {
					continue
				}
				neighbors := C.Neighbors(vertex)
				if !equalIntSlices(neighbors, expectedNeighbors) {
					t.Errorf("Vertex %d: expected neighbors %v, got %v", vertex, expectedNeighbors, neighbors)
				}
			}

			// Check that vertices are no longer neighbors
			if C.IsNeighbor(test.u, test.v) {
				t.Errorf("Vertices %d and %d should not be neighbors after deletion", test.u, test.v)
			}
			if C.IsNeighbor(test.v, test.u) {
				t.Errorf("Vertices %d and %d should not be neighbors after deletion", test.v, test.u)
			}
		})
	}
}

func TestZComplexDeleteEdgeNonExistent(t *testing.T) {
	C := NewZComplexFromMaximalSimplices([][]int{{0, 1}, {2, 3}})
	initialEdgeCount := C.NumEdges()

	// Deleting non-existent edge should not change edge count
	edgeIdx, ok := C.IndexOfEdge(0, 2) // no edge between 0 and 2
	if ok {
		C.DeleteEdge(edgeIdx)
	}
	// Note: If edge doesn't exist, we don't call DeleteEdge at all
	if C.NumEdges() != initialEdgeCount {
		t.Errorf("Expected %d edges after deleting non-existent edge, got %d", initialEdgeCount, C.NumEdges())
	}

	// Neighbors should remain unchanged
	expectedNeighbors := map[int][]int{
		0: {1},
		1: {0},
		2: {3},
		3: {2},
	}
	for vertex, expected := range expectedNeighbors {
		neighbors := C.Neighbors(vertex)
		if !equalIntSlices(neighbors, expected) {
			t.Errorf("Vertex %d: expected neighbors %v, got %v", vertex, expected, neighbors)
		}
	}
}

func TestZComplexDeleteEdgeRoundTrip(t *testing.T) {
	// Test that adding and then deleting an edge returns to original state
	C := NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2}}) // edges only, no triangles
	
	// Record original state
	originalEdgeCount := C.NumEdges()
	originalNeighbors := make(map[int][]int)
	for i := 0; i < 4; i++ {
		if i < len(C.vertexBasis) {
			neighbors := C.Neighbors(i)
			originalNeighbors[i] = make([]int, len(neighbors))
			copy(originalNeighbors[i], neighbors)
		}
	}

	// Add vertex 3 to have somewhere to connect
	vertices := []ZVertex[ZVertexInt]{ZVertexInt(0), ZVertexInt(1), ZVertexInt(2), ZVertexInt(3)}
	edges := C.EdgeBasis()
	C = NewZComplex(vertices, edges, nil, false, false)

	// Add an edge then delete it
	C.AddEdge(0, 3)
	edgeIdx, ok := C.IndexOfEdge(0, 3)
	if !ok {
		t.Fatalf("Edge (0,3) not found after adding")
	}
	C.DeleteEdge(edgeIdx)

	// Check we're back to original state (for vertices that existed originally)
	if C.NumEdges() != originalEdgeCount {
		t.Errorf("Expected %d edges after round trip, got %d", originalEdgeCount, C.NumEdges())
	}

	for vertex, expectedNeighbors := range originalNeighbors {
		neighbors := C.Neighbors(vertex)
		if !equalIntSlices(neighbors, expectedNeighbors) {
			t.Errorf("Vertex %d: expected neighbors %v after round trip, got %v", vertex, expectedNeighbors, neighbors)
		}
	}
}

func TestZComplexDeleteEdgeTrianglePanic(t *testing.T) {
	// Test that DeleteEdge panics when used on complex with triangles
	C := NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}) // This creates a triangle

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when deleting edge from complex with triangles")
		}
	}()
	
	edgeIdx, ok := C.IndexOfEdge(0, 1)
	if ok {
		C.DeleteEdge(edgeIdx) // Should panic
	} else {
		t.Errorf("Expected edge (0,1) to exist in triangle complex")
	}
}

func TestZComplexTrianglesContainingVertex(t *testing.T) {
	T := ZTriangle[ZVertexInt]{ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)}
	U := ZTriangle[ZVertexInt]{ZVertexInt(3), ZVertexInt(1), ZVertexInt(0)}
	V := ZTriangle[ZVertexInt]{ZVertexInt(0), ZVertexInt(3), ZVertexInt(4)}
	C := NewZComplexFromTriangles([]ZTriangle[ZVertexInt]{T})
	D := NewZComplexFromTriangles([]ZTriangle[ZVertexInt]{T, U, V})

	tests := []struct {
		C        *ZComplex[ZVertexInt]
		V        ZVertex[ZVertexInt]
		Expected []ZTriangle[ZVertexInt]
	}{
		{C, ZVertexInt(0), []ZTriangle[ZVertexInt]{T}},
		{D, ZVertexInt(1), []ZTriangle[ZVertexInt]{T, U}},
		{D, ZVertexInt(0), []ZTriangle[ZVertexInt]{U, T, V}},
	}
	for n, test := range tests {
		got := test.C.TrianglesContainingVertex(test.V)
		if !TriangleSetEqual(got, test.Expected) {
			t.Errorf("Test %d: C.TrianglesContainingVertex(V)=%v, expected %v", n, got, test.Expected)
		}
	}
}

func TestZComplexBFS(t *testing.T) {
	tests := []struct {
		C             *ZComplex[ZVertexInt]
		ExpectedCount int
	}{
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}), 3},
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}, {0, 3, 4}}), 5},
		// note: the last simplex is not reachable from the first.
		{NewZComplexFromMaximalSimplices([][]int{{0, 1}, {0, 2}, {0, 3}, {0, 4}, {5, 6, 7}}), 5},
	}
	for n, test := range tests {
		got := 0
		test.C.BFS(ZVertexInt(0), func(v ZVertex[ZVertexInt], depth int) (stop bool) {
			got++
			return false
		})
		if got != test.ExpectedCount {
			t.Errorf("Test %d: got=%d, expected %d", n, got, test.ExpectedCount)
		}
	}
}

func TestZComplexBFTriangleWalk(t *testing.T) {
	tests := []struct {
		C             *ZComplex[ZVertexInt]
		ExpectedCount int
	}{
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}), 1},
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}, {0, 3, 4}}), 3},
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}, {0, 3, 4}, {2, 5}, {5, 6, 7}}), 4},
		// note case: three independent triangles.  will only enumerate the first one.
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}}), 1},
	}
	for n, test := range tests {
		got := 0
		test.C.BFTriangleWalk(ZVertexInt(0), func(T ZTriangle[ZVertexInt]) {
			got++
		})
		if got != test.ExpectedCount {
			t.Errorf("Test %d: got=%d, expected %d", n, got, test.ExpectedCount)
		}
	}
}

func TestZComplexBFWalk3Cliques(t *testing.T) {
	tests := []struct {
		C             *ZComplex[ZVertexInt]
		ExpectedCount int
	}{
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}), 1},
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}}), 2},
		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}, {0, 3, 4}}), 3},
	}
	for n, test := range tests {
		got := 0
		test.C.BFWalk3Cliques(func(c [3]ZVertex[ZVertexInt]) {
			got++
		})
		if got != test.ExpectedCount {
			t.Errorf("Test %d: got=%d, expected %d", n, got, test.ExpectedCount)
		}
	}
}

func TestZComplexDualComplex(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		Expected *ZComplex[ZVertexInt]
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			NewZComplexFromMaximalSimplices([][]int{{0}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}}),
			NewZComplexFromMaximalSimplices([][]int{{0, 1}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}, {0, 1, 4}}),
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
		},
	}
	for n, test := range tests {
		got := test.C.DualComplex()
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexMaximalSimplicesString(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		Expected string
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			"[][]int{{0, 1, 2}}",
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}}),
			"[][]int{{0, 1, 2}, {0, 1, 3}}",
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0}, {1, 2}, {0, 1, 3}}),
			"[][]int{{0, 1, 3}, {1, 2}}",
		},
	}
	for n, test := range tests {
		got := test.C.MaximalSimplicesString()
		if got != test.Expected {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexMaximalSimplicesStringRandom(t *testing.T) {
	trials := 10
	maxVertices := 10
	verbose := false
	for i := 0; i < trials; i++ {
		numVertices := 3 + rand.Intn(maxVertices)
		probEdge := 0.3
		generator := NewRandomComplexGenerator(numVertices, verbose)
		X, err := generator.RandomCliqueComplex(probEdge)
		if err != nil {
			t.Fatalf("Failed to generate random clique complex: %v", err)
		}
		s := X.MaximalSimplicesString()
		simplices, err := parseSimplicesString(s)
		if err != nil {
			t.Fatalf("Failed to parse simplices string: %v", err)
		}
		Y := NewZComplexFromMaximalSimplices(simplices)
		if X.String() != Y.String() {
			t.Errorf("Test %d: got=%v, expected=%v", i, X, Y)
		}
	}
}

func TestZComplexNewFromBoundaryMaps(t *testing.T) {
	tests := []struct {
		d_1      BinaryMatrix
		d_2      BinaryMatrix
		Expected *ZComplex[ZVertexInt]
	}{
		{
			NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 1, 0},
				{1, 0, 1},
				{0, 1, 1},
			}),
			NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{1},
				{1},
			}),
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
		},
	}
	for n, test := range tests {
		got := NewZComplexFromBoundaryMatrices(test.d_1, test.d_2)
		if got.DumpBases() != test.Expected.DumpBases() {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexNearerEdgeInNewIndex(t *testing.T) {
	tests := []struct {
		index    map[ZVertex[ZVertexInt]]int
		a        ZEdge[ZVertexInt]
		b        ZEdge[ZVertexInt]
		expected bool
	}{
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1},
			NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
			NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
			false,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1},
			NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
			NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(0)),
			false,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1, ZVertexInt(2): 2},
			NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
			NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
			true,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1, ZVertexInt(2): 2},
			NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
			NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
			false,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 10, ZVertexInt(1): 5, ZVertexInt(2): 7},
			NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
			NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
			false,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 10, ZVertexInt(1): 5, ZVertexInt(2): 7, ZVertexInt(3): 3},
			NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
			NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(3)),
			false,
		},
	}
	for n, test := range tests {
		got := nearerEdgeInNewIndex(test.index, test.a, test.b)
		if got != test.expected {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.expected)
		}
	}
}

func TestZComplexNearerTriangleInNewIndex(t *testing.T) {
	tests := []struct {
		index    map[ZVertex[ZVertexInt]]int
		a        ZTriangle[ZVertexInt]
		b        ZTriangle[ZVertexInt]
		expected bool
	}{
		// xxx cases wip
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1, ZVertexInt(2): 2},
			NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
			NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
			false,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1, ZVertexInt(2): 2, ZVertexInt(3): 3},
			NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
			NewZTriangle[ZVertexInt](ZVertexInt(1), ZVertexInt(2), ZVertexInt(3)),
			true,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1, ZVertexInt(2): 2, ZVertexInt(3): 3, ZVertexInt(4): 4},
			NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
			NewZTriangle[ZVertexInt](ZVertexInt(2), ZVertexInt(3), ZVertexInt(4)),
			true,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1, ZVertexInt(2): 2, ZVertexInt(3): 3, ZVertexInt(4): 4, ZVertexInt(5): 5},
			NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
			NewZTriangle[ZVertexInt](ZVertexInt(3), ZVertexInt(4), ZVertexInt(5)),
			true,
		},
		{
			map[ZVertex[ZVertexInt]]int{ZVertexInt(0): 0, ZVertexInt(1): 1, ZVertexInt(2): 2, ZVertexInt(3): 3, ZVertexInt(4): 4, ZVertexInt(5): 5},
			NewZTriangle[ZVertexInt](ZVertexInt(3), ZVertexInt(4), ZVertexInt(5)),
			NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
			false,
		},
	}
	for n, test := range tests {
		got := nearerTriangleInNewIndex(test.index, test.a, test.b)
		if got != test.expected {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.expected)
		}
	}
}

func TestZComplexSortBasesByDistance(t *testing.T) {
	tests := []struct {
		simplices         [][]int
		initialVertex     int
		expectedVertices  []int
		expectedEdges     [][2]int
		expectedTriangles [][3]int
	}{
		{
			[][]int{{0, 1, 2}},
			0,
			[]int{0, 1, 2},
			[][2]int{{0, 1}, {2, 0}, {1, 2}},
			[][3]int{{0, 1, 2}},
		},
		{
			[][]int{{0, 1, 2}, {2, 3}},
			0,
			[]int{0, 1, 2, 3},
			[][2]int{{0, 1}, {2, 0}, {1, 2}, {2, 3}},
			[][3]int{{0, 1, 2}},
		},
		{
			[][]int{{0, 1, 2}, {2, 3, 4}},
			0,
			[]int{0, 1, 2, 3, 4},
			[][2]int{{0, 1}, {0, 2}, {1, 2}, {2, 3}, {2, 4}, {3, 4}},
			[][3]int{{0, 1, 2}, {2, 3, 4}},
		},
		{
			[][]int{{0, 1, 2}, {2, 3, 4}},
			2,
			[]int{2, 0, 1, 3, 4},
			[][2]int{{0, 2}, {1, 2}, {2, 3}, {2, 4}, {0, 1}, {3, 4}},
			[][3]int{{0, 1, 2}, {2, 3, 4}},
		},
		// an example with two disconnected components
		{
			[][]int{{0, 1, 2}, {3, 4, 5}},
			0,
			[]int{0, 1, 2, 3, 4, 5},
			[][2]int{{0, 1}, {0, 2}, {1, 2}, {3, 4}, {3, 5}, {4, 5}},
			[][3]int{{0, 1, 2}, {3, 4, 5}},
		},
	}
	for n, test := range tests {
		// first, convert test data from bare ints
		X := NewZComplexFromMaximalSimplices(test.simplices)

		expectedVertexBasis := make([]ZVertex[ZVertexInt], len(test.expectedVertices))
		for i, s := range test.expectedVertices {
			expectedVertexBasis[i] = ZVertex[ZVertexInt](ZVertexInt(s))
		}
		expectedEdgeBasis := make([]ZEdge[ZVertexInt], len(test.expectedEdges))
		for i, s := range test.expectedEdges {
			expectedEdgeBasis[i] = NewZEdge(ZVertexInt(s[0]), ZVertexInt(s[1]))
		}
		expectedTriangleBasis := make([]ZTriangle[ZVertexInt], len(test.expectedTriangles))
		for i, s := range test.expectedTriangles {
			expectedTriangleBasis[i] = NewZTriangle(ZVertexInt(s[0]), ZVertexInt(s[1]), ZVertexInt(s[2]))
		}

		Y := X.SortBasesByDistance(test.initialVertex)
		gotVertexBasis := Y.VertexBasis()
		if !reflect.DeepEqual(gotVertexBasis, expectedVertexBasis) {
			t.Errorf("Test %d: vertex basis: got=%v, expected=%v", n, gotVertexBasis, expectedVertexBasis)
		}
		gotEdgeBasis := Y.EdgeBasis()
		if !reflect.DeepEqual(gotEdgeBasis, expectedEdgeBasis) {
			t.Errorf("Test %d: edge basis: got=%v, expected=%v", n, gotEdgeBasis, expectedEdgeBasis)
		}
		gotTriangleBasis := Y.TriangleBasis()
		if !reflect.DeepEqual(gotTriangleBasis, expectedTriangleBasis) {
			t.Errorf("Test %d: triangle basis: got=%v, expected=%v", n, gotTriangleBasis, expectedTriangleBasis)
		}
	}
}

func TestZComplexSortBasesByDistanceRandomComplexes(t *testing.T) {
	trials := 10
	minVertices := 5
	maxVertices := 10
	verbose := false

	for i := 0; i < trials; i++ {
		numVertices := rand.Intn(maxVertices-minVertices) + minVertices
		R := NewRandomComplexGenerator(numVertices, verbose)
		Xoriginal, err := R.RandomCliqueComplex(0.4)
		if err != nil {
			t.Errorf("Failed to generate random clique complex: %v", err)
			continue
		}
		//log.Printf("Xoriginal:\n%s", Xoriginal.DumpBases())
		initialVertex := rand.Intn(len(Xoriginal.VertexBasis()))

		var Xsorted *ZComplex[ZVertexInt]
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Trial %d: Panic during SortBasesByDistance: %v\noriginal complex:\n%v", i, r, Xoriginal.DumpBases())
				}
			}()
			Xsorted = Xoriginal.SortBasesByDistance(initialVertex)
		}()
		//log.Printf("XSorted:\n%s", Xsorted.DumpBases())

		// simple check that the bases have the same elements (ignoring order)
		resortedVertices := make([]ZVertex[ZVertexInt], len(Xsorted.VertexBasis()))
		copy(resortedVertices, Xsorted.VertexBasis())

		resortedEdges := make([]ZEdge[ZVertexInt], len(Xsorted.EdgeBasis()))
		copy(resortedEdges, Xsorted.EdgeBasis())

		resortedTriangles := make([]ZTriangle[ZVertexInt], len(Xsorted.TriangleBasis()))
		copy(resortedTriangles, Xsorted.TriangleBasis())

		Xresorted := NewZComplex(resortedVertices, resortedEdges, resortedTriangles, true, verbose)
		//log.Printf("XResorted:\n%s", Xresorted.DumpBases())

		if !reflect.DeepEqual(Xoriginal.VertexBasis(), Xresorted.VertexBasis()) {
			t.Errorf("Trial %d: Vertex bases are not equal (ignoring order)", i)
		}
		if !reflect.DeepEqual(Xoriginal.EdgeBasis(), Xresorted.EdgeBasis()) {
			t.Errorf("Trial %d: Edge bases are not equal (ignoring order)", i)
		}
		if !reflect.DeepEqual(Xoriginal.TriangleBasis(), Xresorted.TriangleBasis()) {
			t.Errorf("Trial %d: Triangle bases are not equal (ignoring order)", i)
		}
				
		// check that the first vertex is the initial vertex
		firstVertex, ok := Xsorted.VertexBasis()[0].(ZVertexInt)
		if !ok {
			t.Errorf("Trial %d: First vertex should be of type ZVertexInt, got %T", i, Xsorted.VertexBasis()[0])
			continue
		}
		if int(firstVertex) != initialVertex {
			t.Errorf("Trial %d: First vertex should be the initial vertex %d, got %v",
				i, initialVertex, firstVertex)
		}
	}
}

func TestZComplexSubcomplexByDepth(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		Depth    int
		Expected *ZComplex[ZVertexInt]
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			0,
			NewZComplexFromMaximalSimplices([][]int{{0}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			1,
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
		},
	}
	for n, test := range tests {
		got := test.C.SubcomplexByDepth(test.Depth)
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexSubcomplexByVertices(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		V        map[int]bool
		Expected *ZComplex[ZVertexInt]
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			map[int]bool{0: true, 1: true},
			NewZComplexFromMaximalSimplices([][]int{{0, 1}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}}),
			map[int]bool{0: true, 1: true},
			NewZComplexFromMaximalSimplices([][]int{{0, 1}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}}),
			map[int]bool{0: true, 1: true, 2: true},
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
		},
	}
	for n, test := range tests {
		got := test.C.SubcomplexByVertices(test.V)
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexTriangularDepthFiltration(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		Start    int
		Expected []*ZComplex[ZVertexInt]
	}{
		{
			NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}}, false),
			0,
			[]*ZComplex[ZVertexInt]{
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}}, false),
			},
		},
		{
			NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 3}}, false),
			0,
			[]*ZComplex[ZVertexInt]{
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}}, false),
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 3}}, false),
			},
		},
		{
			NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 2, 3}}, false),
			0,
			[]*ZComplex[ZVertexInt]{
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}}, false),
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 2, 3}}, false),
			},
		},
		{
			NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 2, 3}}, false),
			3,
			[]*ZComplex[ZVertexInt]{
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{3, 1, 2}}, false),
				// this is a bit tricky - we give all of the simplices here
				// to completely control the ordering.
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{3}, {1}, {2}, {0}, {1, 3}, {2, 3}, {0, 1}, {1, 2}, {0, 2}, {3, 1, 2}, {0, 1, 2}}, false),
			},
		},
		{
			NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 3}, {3, 4, 5}}, false),
			0,
			[]*ZComplex[ZVertexInt]{
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}}, false),
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 3}, {3, 4, 5}}, false),
			},
		},
		{
			NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 3}, {3, 4, 5}, {3, 6}}, false),
			0,
			[]*ZComplex[ZVertexInt]{
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}}, false),
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 3}, {3, 4, 5}}, false),
				NewZComplexFromMaximalSimplicesOptionalSort([][]int{{0, 1, 2}, {1, 3}, {3, 4}, {3, 5}, {3, 6}, {3, 4, 5}}, false),
			},
		},
	}
	for n, test := range tests {
		var got []*ZComplex[ZVertexInt]
		test.C.TriangularDepthFiltration(test.C.VertexBasis()[test.Start], func(step int, subcomplex *ZComplex[ZVertexInt]) (stop bool) {
			got = append(got, subcomplex)
			return false
		})
		if !reflect.DeepEqual(got, test.Expected) {
			var formatComplex = func(X *ZComplex[ZVertexInt]) string {
				s := X.String()
				s += "\n" + X.DumpBases()
				return s
			}
			var diff string
			for i, X := range test.Expected {
				if reflect.DeepEqual(X, got[i]) {
					diff += fmt.Sprintf("complex %d: equal\n\n", i)
					continue
				}
				diff += fmt.Sprintf("complex %d:\n\nExpected: %s", i, formatComplex(X))
				if len(got) >= i+1 {
					diff += fmt.Sprintf("\nGot: %s", formatComplex(got[i]))
				}
			}
			if len(got) > len(test.Expected) {
				diff += "Got more than expected:\n"
				for i := len(test.Expected); i < len(got); i++ {
					Y := got[i]
					diff += fmt.Sprintf("complex %d:\n\nGot: %s", i, formatComplex(Y))
				}
			}
			t.Errorf("Test %d: diff: %s", n, diff)
		}
	}
}

func TestZComplexVertexToEdgeIncidenceMap(t *testing.T) {
	tests := []struct {
		C        *ZComplex[ZVertexInt]
		Expected map[int][]int
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			map[int][]int{
				0: {0, 1},
				1: {0, 2},
				2: {1, 2},
			},
		},
	}
	for n, test := range tests {
		got := test.C.VertexToEdgeIncidenceMap()
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}
