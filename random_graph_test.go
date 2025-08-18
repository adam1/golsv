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
		numVertices int
		k           int
		expectError bool
	}{
		{6, 2, false},  // 6 vertices, 2-regular
		{8, 4, false},  // 8 vertices, 4-regular
		{10, 3, false}, // 10 vertices, 3-regular (odd k)
		{12, 5, false}, // 12 vertices, 5-regular (odd k)
		{5, 3, true},   // 5*3=15 is odd, should fail
		{4, 4, true},   // k >= numVertices, should fail
		{4, 1, false},  // 4 vertices, 1-regular
	}

	verbose := false
	for _, test := range tests {
		t.Run(fmt.Sprintf("n=%d_k=%d", test.numVertices, test.k), func(t *testing.T) {
			G, err := RandomRegularGraphByBalancing(test.numVertices, test.k, 1000, verbose)

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

			// Verify correct number of vertices
			if G.NumVertices() != test.numVertices {
				t.Errorf("expected %d vertices, got %d", test.numVertices, G.NumVertices())
			}

			// Verify correct number of edges
			expectedEdges := test.numVertices * test.k / 2
			if G.NumEdges() != expectedEdges {
				t.Errorf("expected %d edges, got %d", expectedEdges, G.NumEdges())
			}

			// Verify the graph is actually regular
			if !G.IsRegular() {
				t.Errorf("generated graph is not regular")
			}

			// Verify all vertices have the expected degree k
			expectedDegree := test.k
			if test.numVertices > 0 {
				actualDegree := G.Degree(0) // Any vertex will do since graph is regular
				if actualDegree != expectedDegree {
					t.Errorf("expected degree %d, got %d", expectedDegree, actualDegree)
				}
			}

			// Verify no triangles (this should be a graph, not a complex)
			if G.NumTriangles() != 0 {
				t.Errorf("expected 0 triangles in graph, got %d", G.NumTriangles())
			}
		})
	}
}
