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
	// xxx throw into map
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
				{1, 0, 1},
				{1, 1, 0},
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
			t.Errorf("Test %d: got=%v, expected %v", n, got, test.Expected)
		}
	}
}
