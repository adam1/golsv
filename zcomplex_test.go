package golsv

import (
	"testing"
)

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


// xxx wip
// func TestZComplexCosystolicVector(t *testing.T) {
// 	// xxx some basic examples first
// 	tests := []struct {
// 		C *ZComplex[ZVertexInt]
// 		Expected int
// 	}{
// 		//		{NewZComplexFilledTriangle(), 0},
// 		{NewZComplexEmptyTriangle(), 1},
// 	}
// 	for n, test := range tests {
// 		v := test.C.CosystolicVector()
// 		if len(v) != test.Expected {
// 			t.Errorf("Test %d: len(v) got=%d, expected %d", n, len(v), test.Expected)
// 		}
// 	}
// }
