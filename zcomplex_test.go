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

func TestZComplexDepthGradedSubcomplexes(t *testing.T) {
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
		test.C.DepthGradedSubcomplexes(initialVertex, func(depth int, subcomplex *ZComplex[ZVertexInt], verticesAtDepth []ZVertex[ZVertexInt]) {
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
		d1, d2, err := generator.RandomCliqueComplex(probEdge)
		if err != nil {
			t.Fatalf("Failed to generate random clique complex: %v", err)
		}
		X := NewZComplexFromBoundaryMatrices(d1, d2)
		s := X.MaximalSimplicesString()
		simplices, err := parseSimplicesString(s)
		if err != nil {
			t.Fatalf("Failed to parse simplices string: %v", err)
		}
		Y := NewZComplexFromMaximalSimplices(simplices)
		if !reflect.DeepEqual(X, Y) {
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
		d_1, d_2, err := R.RandomCliqueComplex(0.4)
		if err != nil {
			t.Errorf("Failed to generate random clique complex: %v", err)
			continue
		}
		Xoriginal := NewZComplexFromBoundaryMatrices(d_1, d_2)
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
