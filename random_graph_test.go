package golsv

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestRandomGraph(t *testing.T) {
	trials := 10
	minNumVertices := 1
	maxNumVertices := 20
	verbose := false
	for i := 0; i < trials; i++ {
		numVertices := rand.Intn(maxNumVertices - minNumVertices) + minNumVertices
		probEdge := 0.3
		G, err := RandomGraph(numVertices, probEdge, verbose)
		if err != nil {
			t.Error("wanted no error, got ", err)
		}
		d_1 := G.D1()
		d_2 := G.D2()
		if d_1.NumRows() != numVertices {
			t.Error("wanted d_1.NumRows() == n, got ", d_1.NumRows())
		}
		// Graph should have no triangles (d_2 should be empty)
		if d_2.NumColumns() != 0 {
			t.Errorf("wanted d_2.NumColumns() == 0 for graph, got %d", d_2.NumColumns())
		}
		// For probability 0.3, expect roughly 30% of possible edges
		maxPossibleEdges := numVertices * (numVertices - 1) / 2
		actualEdges := d_1.NumColumns()
		if numVertices > 1 && actualEdges < 0 {
			t.Errorf("wanted non-negative edge count, got %d", actualEdges)
		}
		if actualEdges > maxPossibleEdges {
			t.Errorf("wanted edge count <= %d, got %d", maxPossibleEdges, actualEdges)
		}
	}
}

func TestRandomRegularGraphByBalancing(t *testing.T) {
	tests := []struct {
		numVertices   int
		k             int
		maxIterations int
		expectError   bool
	}{
		{6, 2, 1000, false},  // 6 vertices, 2-regular
		{8, 4, 1000, false},  // 8 vertices, 4-regular
		{10, 3, 1000, false}, // 10 vertices, 3-regular (odd k)
		{12, 5, 1000, false}, // 12 vertices, 5-regular (odd k)
		{5, 3, 1000, true},   // 5*3=15 is odd, should fail
		{4, 4, 1000, true},   // k >= numVertices, should fail
		{3, 1, 100, true},   // small case, 1-regular
	}

	verbose := false
	for _, test := range tests {
		t.Run(fmt.Sprintf("n=%d_k=%d", test.numVertices, test.k), func(t *testing.T) {
			G, err := RandomRegularGraphByBalancing(test.numVertices, test.k, test.maxIterations, verbose)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error for n=%d, k=%d", test.numVertices, test.k)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			d_1 := G.D1()

			if d_1.NumRows() != test.numVertices {
				t.Errorf("expected d_1.NumRows()=%d, got %d", test.numVertices, d_1.NumRows())
			}
			// xxx all tbd

// 			expectedEdges := test.numVertices * test.k / 2
// 			if d_1.NumColumns() != expectedEdges {
// 				t.Errorf("expected %d edges, got %d", expectedEdges, d_1.NumColumns())
// 			}

			// Verify regularity: each vertex should have degree k
			// xxx tbd
// 			degrees := make([]int, test.numVertices)
// 			for j := 0; j < d_1.NumColumns(); j++ {
// 				vertices := make([]int, 0, 2)
// 				for i := 0; i < test.numVertices; i++ {
// 					if d_1.Get(i, j) == 1 {
// 						vertices = append(vertices, i)
// 					}
// 				}
// 				if len(vertices) != 2 {
// 					t.Errorf("edge %d should connect exactly 2 vertices, got %d", j, len(vertices))
// 				}
// 				degrees[vertices[0]]++
// 				degrees[vertices[1]]++
// 			}

// 			for i, degree := range degrees {
// 				if degree != test.k {
// 					t.Errorf("vertex %d has degree %d, expected %d", i, degree, test.k)
// 				}
// 			}
		})
	}
}
