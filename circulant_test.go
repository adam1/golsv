package golsv

import (
	"reflect"
	"testing"
)

func TestNormalizeCirculantSteps(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		steps    []int
		expected []int
	}{
		{
			name:     "empty steps",
			n:        5,
			steps:    []int{},
			expected: []int{},
		},
		{
			name:     "single positive step",
			n:        5,
			steps:    []int{1},
			expected: []int{1, 4}, // 1 and its inverse 5-1=4
		},
		{
			name:     "single negative step",
			n:        5,
			steps:    []int{-1},
			expected: []int{1, 4}, // -1 mod 5 = 4, and its inverse 5-4=1
		},
		{
			name:     "zero step (ignored)",
			n:        5,
			steps:    []int{0},
			expected: []int{},
		},
		{
			name:     "step equal to n (ignored as equivalent to 0)",
			n:        5,
			steps:    []int{5},
			expected: []int{},
		},
		{
			name:     "multiple steps",
			n:        6,
			steps:    []int{1, 2},
			expected: []int{1, 2, 4, 5}, // 1,5 and 2,4
		},
		{
			name:     "mixed positive and negative steps",
			n:        7,
			steps:    []int{1, -2, 3},
			expected: []int{1, 2, 3, 4, 5, 6}, // 1,6 + 5,2 + 3,4 = all non-zero
		},
		{
			name:     "duplicate steps (deduplicated)",
			n:        4,
			steps:    []int{1, 1, -1, 3},
			expected: []int{1, 3}, // 1,3 + 3,1 = just 1,3
		},
		{
			name:     "step larger than n",
			n:        3,
			steps:    []int{7},
			expected: []int{1, 2}, // 7 mod 3 = 1, so 1 and 3-1=2
		},
		{
			name:     "step much smaller than -n",
			n:        4,
			steps:    []int{-9},
			expected: []int{1, 3}, // -9 mod 4 = 3, so 3 and 4-3=1
		},
		{
			name:     "self-inverse step (n/2 when n is even)",
			n:        6,
			steps:    []int{3},
			expected: []int{3}, // 3 is its own inverse in Z/6Z
		},
		{
			name:     "complete graph generators",
			n:        4,
			steps:    []int{1, 2},
			expected: []int{1, 2, 3}, // 1,3 + 2,2 = 1,2,3
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := normalizeCirculantSteps(test.n, test.steps)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("normalizeCirculantSteps(%d, %v) = %v, expected %v",
					test.n, test.steps, result, test.expected)
			}
		})
	}
}

func TestNormalizeCirculantStepsEdgeCases(t *testing.T) {
	// Test edge cases
	result := normalizeCirculantSteps(0, []int{1, 2})
	if result != nil {
		t.Errorf("expected nil for n=0, got %v", result)
	}
	
	result = normalizeCirculantSteps(-1, []int{1, 2})
	if result != nil {
		t.Errorf("expected nil for n<0, got %v", result)
	}
	
	// Test that result is always sorted
	result = normalizeCirculantSteps(10, []int{7, 2, 5})
	for i := 1; i < len(result); i++ {
		if result[i-1] >= result[i] {
			t.Errorf("result not sorted: %v", result)
		}
	}
}

func TestCirculantGraph(t *testing.T) {
	tests := []struct {
		name                string
		n                   int
		steps               []int
		expectedVertices    []int // expected vertex order (as ints)
		expectedNumEdges    int
		expectedFirstEdge   [2]int // first edge in sorted order
		expectError         bool
	}{
		{
			name:             "invalid n=1",
			n:                1,
			steps:            []int{1},
			expectError:      true,
		},
		{
			name:             "invalid n=0",
			n:                0,
			steps:            []int{1},
			expectError:      true,
		},
		{
			name:             "trivial case: 2 vertices, step 1",
			n:                2,
			steps:            []int{1},
			expectedVertices: []int{0, 1}, // 0 at distance 0, 1 at distance 1
			expectedNumEdges: 1,
			expectedFirstEdge: [2]int{0, 1},
		},
		{
			name:             "triangle: 3 vertices, step 1",
			n:                3,
			steps:            []int{1},
			expectedVertices: []int{0, 1, 2}, // 0 at distance 0, 1,2 at distance 1
			expectedNumEdges: 3,
			expectedFirstEdge: [2]int{0, 1}, // edges from vertex 0 come first
		},
		{
			name:             "square: 4 vertices, step 1",
			n:                4,
			steps:            []int{1},
			expectedVertices: []int{0, 1, 3, 2}, // 0 at dist 0, 1,3 at dist 1, 2 at dist 2
			expectedNumEdges: 4,
			expectedFirstEdge: [2]int{0, 1}, // edges from vertex 0 come first
		},
		{
			name:             "pentagon: 5 vertices, step 1",
			n:                5,
			steps:            []int{1},
			expectedVertices: []int{0, 1, 4, 2, 3}, // BFS order from 0
			expectedNumEdges: 5,
			expectedFirstEdge: [2]int{0, 1},
		},
		{
			name:             "complete graph: 4 vertices, steps 1,2",
			n:                4,
			steps:            []int{1, 2},
			expectedVertices: []int{0, 1, 2, 3}, // all at distance ≤ 1 from 0
			expectedNumEdges: 6, // complete graph on 4 vertices
			expectedFirstEdge: [2]int{0, 1},
		},
		{
			name:             "hexagon with diagonals: 6 vertices, steps 1,3",
			n:                6,
			steps:            []int{1, 3},
			expectedVertices: []int{0, 1, 3, 5, 2, 4}, // BFS: 0->1,3,5->2,4
			expectedNumEdges: 9, // 6 edges from step 1, 3 edges from step 3
			expectedFirstEdge: [2]int{0, 1},
		},
		{
			name:             "empty steps",
			n:                5,
			steps:            []int{},
			expectedVertices: []int{0, 1, 2, 3, 4}, // no edges, distances all -1 except 0
			expectedNumEdges: 0,
		},
		{
			name:             "step 0 (self-loops avoided)",
			n:                3,
			steps:            []int{0, 1},
			expectedVertices: []int{0, 1, 2}, // step 0 creates no edges
			expectedNumEdges: 3, // only step 1 creates edges
			expectedFirstEdge: [2]int{0, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			complex, err := CirculantGraph(test.n, test.steps, false)
			
			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			// Check number of vertices
			vertices := complex.VertexBasis()
			if len(vertices) != test.n {
				t.Errorf("expected %d vertices, got %d", test.n, len(vertices))
			}
			
			// Check vertex ordering
			if test.expectedVertices != nil {
				actualVertices := make([]int, len(vertices))
				for i, v := range vertices {
					actualVertices[i] = int(v.(ZVertexInt))
				}
				if !reflect.DeepEqual(actualVertices, test.expectedVertices) {
					t.Errorf("vertex order: expected %v, got %v", test.expectedVertices, actualVertices)
				}
			}
			
			// Check number of edges
			edges := complex.EdgeBasis()
			if len(edges) != test.expectedNumEdges {
				t.Errorf("expected %d edges, got %d", test.expectedNumEdges, len(edges))
			}
			
			// Check first edge if specified
			if test.expectedNumEdges > 0 && test.expectedFirstEdge != [2]int{} {
				firstEdge := edges[0]
				actualFirstEdge := [2]int{
					int(firstEdge[0].(ZVertexInt)),
					int(firstEdge[1].(ZVertexInt)),
				}
				if actualFirstEdge != test.expectedFirstEdge {
					t.Errorf("first edge: expected %v, got %v", test.expectedFirstEdge, actualFirstEdge)
				}
			}
			
			// Verify it's a 1-dimensional complex (no triangles)
			triangles := complex.TriangleBasis()
			if len(triangles) != 0 {
				t.Errorf("expected no triangles, got %d", len(triangles))
			}
		})
	}
}

func TestCirculantGraphDistances(t *testing.T) {
	// Test that distances are computed correctly for a specific case
	complex, err := CirculantGraph(6, []int{1}, false) // hexagon
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	vertices := complex.VertexBasis()
	expectedOrder := []int{0, 1, 5, 2, 4, 3} // BFS order from 0
	actualOrder := make([]int, len(vertices))
	for i, v := range vertices {
		actualOrder[i] = int(v.(ZVertexInt))
	}
	
	if !reflect.DeepEqual(actualOrder, expectedOrder) {
		t.Errorf("hexagon vertex order: expected %v, got %v", expectedOrder, actualOrder)
	}
}

func TestCirculantGraphEdgeOrdering(t *testing.T) {
	// Test edge ordering for a case where we can predict the order
	complex, err := CirculantGraph(4, []int{1}, false) // square
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	edges := complex.EdgeBasis()
	if len(edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(edges))
	}
	
	// Expected order based on minimum distance to identity:
	// Distance 0: edges involving vertex 0: (0,1), (0,3)
	// Distance 1: edges not involving vertex 0: (1,2), (2,3)
	expectedEdges := [][2]int{
		{0, 1}, // min distance 0
		{0, 3}, // min distance 0
		{1, 2}, // min distance 1
		{2, 3}, // min distance 1
	}
	
	for i, expectedEdge := range expectedEdges {
		actualEdge := [2]int{
			int(edges[i][0].(ZVertexInt)),
			int(edges[i][1].(ZVertexInt)),
		}
		if actualEdge != expectedEdge {
			t.Errorf("edge %d: expected %v, got %v", i, expectedEdge, actualEdge)
		}
	}
}