package golsv

import (
	"testing"
)

func TestBoundaryDecoderDecodeK3(t *testing.T) {
	graph := NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2}, {2, 0}})
	decoder := NewBoundaryDecoder(graph, false)

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