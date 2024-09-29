package golsv

import (
	"reflect"
	"testing"
)

func TestZEdgeContains(t *testing.T) {
	tests := []struct {
		E ZEdge[ZVertexInt]
		V ZVertex[ZVertexInt]
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
		A ZEdge[ZVertexInt]
		B ZEdge[ZVertexInt]
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
		E ZEdge[ZVertexInt]
		V ZVertex[ZVertexInt]
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
		T ZTriangle[ZVertexInt]
		V ZVertex[ZVertexInt]
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
		A []ZTriangle[ZVertexInt]
		B []ZTriangle[ZVertexInt]
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

func TestZComplexEdgeToTriangleIncidenceMap(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
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

func TestZComplexHasNeighbor(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
		v,u ZVertex[ZVertexInt]
		Expected bool
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			ZVertexInt(0), ZVertexInt(1), true,
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			ZVertexInt(0), ZVertexInt(2), true,
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2, 3}}),
			ZVertexInt(0), ZVertexInt(1), true,
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2, 3}}),
			ZVertexInt(0), ZVertexInt(3), false,
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2, 3}}),
			ZVertexInt(0), ZVertexInt(4), false,
		},
	}
	for n, test := range tests {
		got := test.C.HasNeighbor(test.v, test.u)
		if got != test.Expected {
			t.Errorf("Test %d: C.HasNeighbor(%v, %v)=%v, expected %v", n, test.v, test.u, got, test.Expected)
		}
	}
}

func TestZComplexNeighbors(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
		v ZVertex[ZVertexInt]
		Expected []ZVertex[ZVertexInt]
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			ZVertexInt(0),
			[]ZVertex[ZVertexInt]{ZVertexInt(1), ZVertexInt(2)},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2, 3}}),
			ZVertexInt(0),
			[]ZVertex[ZVertexInt]{ZVertexInt(1), ZVertexInt(2)},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {1, 2, 3}}),
			ZVertexInt(1),
			[]ZVertex[ZVertexInt]{ZVertexInt(0), ZVertexInt(2), ZVertexInt(3)},
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
		C *ZComplex[ZVertexInt]
		V ZVertex[ZVertexInt]
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

func TestZComplexBFTriangleWalk(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
		ExpectedCount int
	}{
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2}}), 1},
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2}, {0,1,3}, {0,3,4}}), 3},
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2}, {0,1,3}, {0,3,4}, {2,5}, {5,6,7}}), 4},
		// note case: three independent triangles.  will only enumerate the first one.
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2}, {3,4,5}, {6,7,8}}), 1},
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
		C *ZComplex[ZVertexInt]
		ExpectedCount int
	}{
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2}}), 1},
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2}, {0,1,3}}), 2},
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2}, {0,1,3}, {0,3,4}}), 3},
	}
	for n, test := range tests {
		got := 0
		test.C.BFWalk3Cliques(ZVertexInt(0), func(c [3]ZVertex[ZVertexInt]) {
			got++
		})
		if got != test.ExpectedCount {
			t.Errorf("Test %d: got=%d, expected %d", n, got, test.ExpectedCount)
		}
	}
}

func TestZComplexNewFromBoundaryMaps(t *testing.T) {
	tests := []struct {
		d_1 BinaryMatrix
		d_2 BinaryMatrix
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
			NewZComplexFromMaximalSimplices([][]int{{0,1,2}}),
		},
	}
	for n, test := range tests {
		got := NewZComplexFromBoundaryMatrices(test.d_1, test.d_2)
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("Test %d: got=%v, expected=%v", n, got, test.Expected)
		}
	}
}

func TestZComplexNearerEdgeInNewIndex(t *testing.T) {
	tests := []struct {
		index map[ZVertex[ZVertexInt]]int
		a ZEdge[ZVertexInt]
		b ZEdge[ZVertexInt]
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
		index map[ZVertex[ZVertexInt]]int
		a ZTriangle[ZVertexInt]
		b ZTriangle[ZVertexInt]
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
		C *ZComplex[ZVertexInt]
		base int
		expectedVertexBasis []ZVertex[ZVertexInt]
		expectedEdgeBasis []ZEdge[ZVertexInt]
		expectedTriangleBasis []ZTriangle[ZVertexInt]
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			0,
			[]ZVertex[ZVertexInt]{ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)},
			[]ZEdge[ZVertexInt]{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(0)),
				NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2))},
			[]ZTriangle[ZVertexInt]{
				NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2))},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {2, 3}}),
			0,
			[]ZVertex[ZVertexInt]{ZVertexInt(0), ZVertexInt(1), ZVertexInt(2), ZVertexInt(3)},
			[]ZEdge[ZVertexInt]{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(0)),
				NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(3))},
			[]ZTriangle[ZVertexInt]{
				NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2))},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {2, 3, 4}}),
			0,
			[]ZVertex[ZVertexInt]{ZVertexInt(0), ZVertexInt(1), ZVertexInt(2), ZVertexInt(3), ZVertexInt(4)},
			[]ZEdge[ZVertexInt]{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(0)),
				NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(3)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(4)),
				NewZEdge[ZVertexInt](ZVertexInt(3), ZVertexInt(4))},
			[]ZTriangle[ZVertexInt]{
				NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
				NewZTriangle[ZVertexInt](ZVertexInt(2), ZVertexInt(3), ZVertexInt(4))},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {2, 3, 4}}),
			2,
			[]ZVertex[ZVertexInt]{ZVertexInt(2), ZVertexInt(0), ZVertexInt(1), ZVertexInt(3), ZVertexInt(4)},
			[]ZEdge[ZVertexInt]{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(2)),
				NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(3)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(4)),
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
				NewZEdge[ZVertexInt](ZVertexInt(3), ZVertexInt(4))},
			[]ZTriangle[ZVertexInt]{
				NewZTriangle[ZVertexInt](ZVertexInt(0), ZVertexInt(1), ZVertexInt(2)),
				NewZTriangle[ZVertexInt](ZVertexInt(2), ZVertexInt(3), ZVertexInt(4))},
		},
	}
	for n, test := range tests {
		test.C.SortBasesByDistance(test.base)
		gotVertexBasis := test.C.VertexBasis()
		if !reflect.DeepEqual(gotVertexBasis, test.expectedVertexBasis) {
			t.Errorf("Test %d: vertex basis: got=%v, expected=%v", n, gotVertexBasis, test.expectedVertexBasis)
		}
		gotEdgeBasis := test.C.EdgeBasis()
		if !reflect.DeepEqual(gotEdgeBasis, test.expectedEdgeBasis) {
			t.Errorf("Test %d: edge basis: got=%v, expected=%v", n, gotEdgeBasis, test.expectedEdgeBasis)
		}
		gotTriangleBasis := test.C.TriangleBasis()
		if !reflect.DeepEqual(gotTriangleBasis, test.expectedTriangleBasis) {
			t.Errorf("Test %d: triangle basis: got=%v, expected=%v", n, gotTriangleBasis, test.expectedTriangleBasis)
		}
	}
}

func TestZComplexVertexToEdgeIncidenceMap(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
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
