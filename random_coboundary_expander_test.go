package golsv

import (
	"testing"
)

func TestSteinerSystemGenerator(t *testing.T) {
	tests := []struct {
		name        string
		numVertices int
		verbose     bool
	}{
		{"small system n=3", 3, false},   // 3 ≡ 3 (mod 6) ✓
		{"medium system n=7", 7, false},  // 7 ≡ 1 (mod 6) ✓  
		{"larger system n=9", 9, false},  // 9 ≡ 3 (mod 6) ✓
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gen := NewSteinerSystemGenerator(test.numVertices, test.verbose)
			
			triangleMap, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}
			
			// Basic sanity checks
			if triangleMap == nil {
				t.Fatal("Generate() returned nil triangles")
			}
			
			// Convert map to slice for iteration
			triangles := make([]ZTriangle[ZVertexInt], 0, len(triangleMap))
			for triangle := range triangleMap {
				triangles = append(triangles, triangle)
			}
			
			// Verify triangle properties
			for i, triangle := range triangles {
				// Check vertices are in valid range
				for j, v := range triangle {
					vi := int(v.(ZVertexInt))
					if vi < 0 || vi >= test.numVertices {
						t.Errorf("Triangle %d vertex %d = %d, out of range [0, %d)", i, j, vi, test.numVertices)
					}
				}
				
				// Check vertices are distinct
				v0, v1, v2 := int(triangle[0].(ZVertexInt)), int(triangle[1].(ZVertexInt)), int(triangle[2].(ZVertexInt))
				if v0 == v1 || v1 == v2 || v0 == v2 {
					t.Errorf("Triangle %d has duplicate vertices: %v", i, triangle)
				}
			}
			
			// Verify Steiner system property: each edge appears in at most one triangle
			edgeCount := make(map[ZEdge[ZVertexInt]]int)
			for _, triangle := range triangles {
				edges := triangle.Edges()
				for _, edge := range edges {
					edgeCount[edge]++
					if edgeCount[edge] > 1 {
						t.Errorf("Edge %v appears in multiple triangles, violates Steiner system property", edge)
					}
				}
			}
		})
	}
}

func TestSteinerSystemGeneratorEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		numVertices int
		expectError bool
	}{
		{"n=0", 0, true},  // n < 3, should error
		{"n=1", 1, true},  // n < 3, should error  
		{"n=2", 2, true},  // n < 3, should error
		{"n=4", 4, true},  // 4 ≡ 4 (mod 6) - not 1 or 3, should error
		{"n=5", 5, true},  // 5 ≡ 5 (mod 6) - not 1 or 3, should error
		{"n=6", 6, true},  // 6 ≡ 0 (mod 6) - not 1 or 3, should error
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gen := NewSteinerSystemGenerator(test.numVertices, false)
			
			_, err := gen.Generate()
			
			if test.expectError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSteinerSystemGeneratorManual(t *testing.T) {
	// Hardcoded n for manual testing - adjust as needed
	n := 13 // Change this to test different sizes: 3, 7, 9, 13, 15, 19, etc.
	
	gen := NewSteinerSystemGenerator(n, true) // verbose=true for debugging
	triangleMap, err := gen.Generate()
	
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	
	// Convert map to slice for display
	triangles := make([]ZTriangle[ZVertexInt], 0, len(triangleMap))
	for triangle := range triangleMap {
		triangles = append(triangles, triangle)
	}
	
	t.Logf("Generated %d triangles for n=%d:", len(triangles), n)
	for i, triangle := range triangles {
		v0, v1, v2 := int(triangle[0].(ZVertexInt)), int(triangle[1].(ZVertexInt)), int(triangle[2].(ZVertexInt))
		t.Logf("  %d: {%d, %d, %d}", i, v0, v1, v2)
	}
	
	// Calculate theoretical maximum
	totalEdges := n * (n - 1) / 2
	theoreticalMaxTriangles := totalEdges / 3 // Each triangle covers 3 edges
	
	t.Logf("Coverage analysis:")
	t.Logf("  Total edges in K_%d: %d", n, totalEdges)
	t.Logf("  Edges covered: %d", len(triangles)*3)
	t.Logf("  Coverage ratio: %.2f%%", float64(len(triangles)*3)/float64(totalEdges)*100)
	t.Logf("  Theoretical max triangles: %d", theoreticalMaxTriangles)
	t.Logf("  Efficiency: %.2f%%", float64(len(triangles))/float64(theoreticalMaxTriangles)*100)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func TestRandomCoboundaryExpanderGenerator(t *testing.T) {
	// Test the full LLR construction
	n := 7  // Start with n=7 (perfect STS exists)
	k := 2  // Use k=2 systems
	
	gen := NewRandomCoboundaryExpanderGenerator(n, k, true)
	complex, err := gen.Generate()
	
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	
	if complex == nil {
		t.Fatal("Generate() returned nil complex")
	}
	
	// Verify basic structure
	expectedVertices := n
	expectedEdges := n * (n - 1) / 2  // Complete graph
	
	if len(complex.VertexBasis()) != expectedVertices {
		t.Errorf("Expected %d vertices, got %d", expectedVertices, len(complex.VertexBasis()))
	}
	
	if len(complex.EdgeBasis()) != expectedEdges {
		t.Errorf("Expected %d edges, got %d", expectedEdges, len(complex.EdgeBasis()))
	}
	
	t.Logf("Generated LLR complex: %d vertices, %d edges, %d triangles",
		len(complex.VertexBasis()), len(complex.EdgeBasis()), len(complex.TriangleBasis()))
	
	// Verify all triangles are valid
	for i, triangle := range complex.TriangleBasis() {
		v0, v1, v2 := int(triangle[0].(ZVertexInt)), int(triangle[1].(ZVertexInt)), int(triangle[2].(ZVertexInt))
		
		if v0 < 0 || v0 >= n || v1 < 0 || v1 >= n || v2 < 0 || v2 >= n {
			t.Errorf("Triangle %d has vertex out of range: {%d, %d, %d}", i, v0, v1, v2)
		}
		
		if v0 == v1 || v1 == v2 || v0 == v2 {
			t.Errorf("Triangle %d has duplicate vertices: {%d, %d, %d}", i, v0, v1, v2)
		}
	}
}