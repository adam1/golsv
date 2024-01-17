package golsv

import (
	"fmt"
	"testing"
)

func TestCayleyExpanderSmallGroups(t *testing.T) {
	tests := []struct {
		baseField string
		generators []MatGF
		expectedNumVertices int
	}{
		{"F16", C2Generators, 2},
		{"F16", C3Generators, 3},
		{"F16", S3Generators, 6},
		{"F4", C2Generators, 2},
		{"F4", C3Generators, 3},
		{"F4", S3Generators, 6},
	}
	for _, tt := range tests {
		lsv := NewLsvContext(tt.baseField)
		E := NewCayleyExpander(lsv, tt.generators, -1, false)
		E.Expand()
		got := E.NumVertices()
		if got != tt.expectedNumVertices {
			t.Errorf("expected %d vertices, got %d", tt.expectedNumVertices, got)
		}
	}
}

func TestCayleyExpanderFirstLsvGenerator(t *testing.T) {
	tests := []struct {
		baseField string
		expectedNumVertices int
	}{
		{"F16", 273},
		{"F4", 21},
	}
	for _, tt := range tests {
		lsv := NewLsvContext(tt.baseField)
		gens := lsv.Generators()[:1]
		E := NewCayleyExpander(lsv, gens, -1, false)
		E.Expand()
		got := E.NumVertices()
		if got != tt.expectedNumVertices {
			t.Errorf("expected %d vertices, got %d", tt.expectedNumVertices, got)
		}
	}
}

func TestCayleyExpanderDepth(t *testing.T) {
	tests := []struct {
		baseField string
		maxDepth int
		expectedNumVertices int
	}{
		{"F16", 1, 15},
		{"F4", 1, 15},
		{"F16", 2, 113},
		{"F4", 2, 113},
		
	}
	for _, tt := range tests {
		lsv := NewLsvContext(tt.baseField)
		E := NewCayleyExpander(lsv, lsv.Generators(), tt.maxDepth, false)
		E.Expand()
		got := E.NumVertices()
		if got != tt.expectedNumVertices {
			t.Errorf("expected %d vertices, got %d", tt.expectedNumVertices, got)
		}
	}
}

func TestCayleyExpanderComplex(t *testing.T) {
	tests := []struct {
		baseField string
		generators []MatGF
		wantVertexBasis string
		wantEdgeBasis string
		wantTriangleBasis string
	}{
		{
			"F16",
			C2Generators,
			"[[0 1 0 1 0 0 0 0 1] [1 0 0 0 1 0 0 0 1]]",
			"[[[0 1 0 1 0 0 0 0 1] [1 0 0 0 1 0 0 0 1]]]",
			"[]",
		},
		{
			"F16",
			C3Generators,
			"[[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]",
			"[[[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0]] [[0 0 1 1 0 0 0 1 0] [1 0 0 0 1 0 0 0 1]] [[0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]]",
			"[[[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]]",
		},
		{
			"F16",
			S3Generators,
			"[[0 0 1 0 1 0 1 0 0] [0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [0 1 0 1 0 0 0 0 1] [1 0 0 0 0 1 0 1 0] [1 0 0 0 1 0 0 0 1]]",
			"[[[0 0 1 0 1 0 1 0 0] [0 0 1 1 0 0 0 1 0]] [[0 0 1 0 1 0 1 0 0] [0 1 0 1 0 0 0 0 1]] [[0 0 1 0 1 0 1 0 0] [1 0 0 0 0 1 0 1 0]] [[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0]] [[0 0 1 1 0 0 0 1 0] [1 0 0 0 1 0 0 0 1]] [[0 1 0 0 0 1 1 0 0] [1 0 0 0 0 1 0 1 0]] [[0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]] [[0 1 0 1 0 0 0 0 1] [1 0 0 0 0 1 0 1 0]] [[0 1 0 1 0 0 0 0 1] [1 0 0 0 1 0 0 0 1]]]",
			"[[[0 0 1 0 1 0 1 0 0] [0 1 0 1 0 0 0 0 1] [1 0 0 0 0 1 0 1 0]] [[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]]",
		},
		{
			"F4",
			C2Generators,
			"[[0 1 0 1 0 0 0 0 1] [1 0 0 0 1 0 0 0 1]]",
			"[[[0 1 0 1 0 0 0 0 1] [1 0 0 0 1 0 0 0 1]]]",
			"[]",
		},
		{
			"F4",
			C3Generators,
			"[[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]",
			"[[[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0]] [[0 0 1 1 0 0 0 1 0] [1 0 0 0 1 0 0 0 1]] [[0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]]",
			"[[[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]]",
		},
		{
			"F4",
			S3Generators,
			"[[0 0 1 0 1 0 1 0 0] [0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [0 1 0 1 0 0 0 0 1] [1 0 0 0 0 1 0 1 0] [1 0 0 0 1 0 0 0 1]]",
			"[[[0 0 1 0 1 0 1 0 0] [0 0 1 1 0 0 0 1 0]] [[0 0 1 0 1 0 1 0 0] [0 1 0 1 0 0 0 0 1]] [[0 0 1 0 1 0 1 0 0] [1 0 0 0 0 1 0 1 0]] [[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0]] [[0 0 1 1 0 0 0 1 0] [1 0 0 0 1 0 0 0 1]] [[0 1 0 0 0 1 1 0 0] [1 0 0 0 0 1 0 1 0]] [[0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]] [[0 1 0 1 0 0 0 0 1] [1 0 0 0 0 1 0 1 0]] [[0 1 0 1 0 0 0 0 1] [1 0 0 0 1 0 0 0 1]]]",
			"[[[0 0 1 0 1 0 1 0 0] [0 1 0 1 0 0 0 0 1] [1 0 0 0 0 1 0 1 0]] [[0 0 1 1 0 0 0 1 0] [0 1 0 0 0 1 1 0 0] [1 0 0 0 1 0 0 0 1]]]",
		},
	}
	for _, tt := range tests {
		lsv := NewLsvContext(tt.baseField)
		E := NewCayleyExpander(lsv, tt.generators, -1, false)
		E.Expand()
		sortBases := true
		got := E.Complex(sortBases)
		gotVertexBasis := fmt.Sprintf("%v", got.VertexBasis())
		if gotVertexBasis != tt.wantVertexBasis {
			t.Errorf("expected vertex basis %s, got %s", tt.wantVertexBasis, gotVertexBasis)
		}
		gotEdgeBasis := fmt.Sprintf("%v", got.EdgeBasis())
		if gotEdgeBasis != tt.wantEdgeBasis {
			t.Errorf("expected edge basis %s, got %s", tt.wantEdgeBasis, gotEdgeBasis)
		}
		gotTriangleBasis := fmt.Sprintf("%v", got.TriangleBasis())
		if gotTriangleBasis != tt.wantTriangleBasis {
			t.Errorf("expected triangle basis %s, got %s", tt.wantTriangleBasis, gotTriangleBasis)
		}
	}
}
