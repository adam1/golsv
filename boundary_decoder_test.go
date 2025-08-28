package golsv

import (
	//	"math/rand"
	"testing"
)

func TestSSFBoundaryDecoderDecodeK3(t *testing.T) {
	graph := NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2}, {2, 0}})
	decoder := NewSSFBoundaryDecoder(graph, false)

	errorVec := NewBinaryVector(graph.NumEdges())
	errorVec.Set(0, 1)

	syndrome := decoder.Syndrome(errorVec)

	err, decodedErrorVec := decoder.Decode(syndrome)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !decoder.SameCoset(errorVec, decodedErrorVec) {
		t.Errorf("SameCoset(errorVec, decodedErrorVec) returned false")
		t.Logf("Original error: %v", errorVec)
		t.Logf("Decoded error: %v", decodedErrorVec) 
		t.Logf("Syndrome: %v", syndrome)
	}
}

func TestSSFBoundaryDecoderDecodeK6(t *testing.T) {
	graph := NewZComplexFromMaximalSimplices([][]int{
		{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5},
		{1, 2}, {1, 3}, {1, 4}, {1, 5},
		{2, 3}, {2, 4}, {2, 5},
		{3, 4}, {3, 5},
		{4, 5},
	})
	decoder := NewSSFBoundaryDecoder(graph, false)

	errorVec := NewBinaryVector(graph.NumEdges())
// 	edge1 := rand.Intn(graph.NumEdges())
// 	edge2 := rand.Intn(graph.NumEdges())
	edge1 := 0
	edge2 := 9
	errorVec.Set(edge1, 1)
	errorVec.Set(edge2, 1)

	syndrome := decoder.Syndrome(errorVec)

	err, decodedErrorVec := decoder.Decode(syndrome)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if !decoder.SameCoset(errorVec, decodedErrorVec) {
		t.Errorf("SameCoset(errorVec, decodedErrorVec) returned false")
		t.Logf("Original error: %v", errorVec)
		t.Logf("Decoded error: %v", decodedErrorVec)
		t.Logf("Syndrome: %v", syndrome)
	}
}
