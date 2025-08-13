package golsv

import (
	"fmt"
	"log"
	"sort"
)

// normalizeCirculantSteps normalizes step generators for a circulant graph.
// It ensures all steps are positive representatives in Z/nZ and includes
// both each step and its additive inverse for symmetric graph generation.
func normalizeCirculantSteps(n int, steps []int) []int {
	if n <= 0 {
		return nil
	}
	
	stepSet := make(map[int]bool)
	for _, step := range steps {
		if step == 0 {
			continue // skip zero steps (self-loops)
		}
		
		// Normalize to positive representative
		normalizedStep := step % n
		if normalizedStep < 0 {
			normalizedStep += n
		}
		
		// Skip if this would be a self-loop (e.g., step n in Z/nZ)
		if normalizedStep == 0 {
			continue
		}
		
		// Add both the step and its additive inverse
		stepSet[normalizedStep] = true
		stepSet[n - normalizedStep] = true
	}
	
	// Convert back to sorted slice for consistent ordering
	normalizedSteps := make([]int, 0, len(stepSet))
	for step := range stepSet {
		normalizedSteps = append(normalizedSteps, step)
	}
	
	sort.Ints(normalizedSteps)
	return normalizedSteps
}

// CirculantGraph creates a circulant graph as a 1-dimensional simplicial complex.
// Generates the graph as a Cayley graph using BFS from the identity element (vertex 0).
// Vertices and edges are automatically ordered by distance from the identity.
func CirculantGraph(n int, steps []int, verbose bool) (*ZComplex[ZVertexInt], error) {
	if n < 2 {
		return nil, fmt.Errorf("number of vertices %d must be at least 2", n)
	}
	
	normalizedSteps := normalizeCirculantSteps(n, steps)
	
	if verbose {
		log.Printf("Generating circulant Cayley graph with %d vertices", n)
		log.Printf("Original steps: %v", steps)
		log.Printf("Normalized steps: %v", normalizedSteps)
	}
	
	// Generate Cayley graph using BFS from identity (vertex 0)
	visited := make(map[int]bool)
	vertexBasis := make([]ZVertex[ZVertexInt], 0, n)
	edgeSet := make(map[[2]int]bool)
	edgeBasis := make([]ZEdge[ZVertexInt], 0)
	
	// BFS queue: each element is (vertex, depth)
	queue := []struct{vertex, depth int}{{0, 0}}
	visited[0] = true
	vertexBasis = append(vertexBasis, ZVertexInt(0))
	
	if verbose {
		log.Printf("Starting BFS from identity vertex 0")
	}
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if verbose {
			log.Printf("Processing vertex %d at depth %d", current.vertex, current.depth)
		}
		
		// Apply each normalized generator (step) to current vertex
		for _, step := range normalizedSteps {
			neighbor := (current.vertex + step) % n
			
			// Create edge (always store with smaller vertex first for consistency)
			var edge [2]int
			if current.vertex < neighbor {
				edge = [2]int{current.vertex, neighbor}
			} else {
				edge = [2]int{neighbor, current.vertex}
			}
			
			// Add edge if not already seen
			if !edgeSet[edge] {
				edgeSet[edge] = true
				edgeBasis = append(edgeBasis, NewZEdge(ZVertexInt(edge[0]), ZVertexInt(edge[1])))
				
				if verbose {
					log.Printf("  Added edge [%d %d]", edge[0], edge[1])
				}
			}
			
			// Add neighbor to queue if not visited
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, struct{vertex, depth int}{neighbor, current.depth + 1})
				vertexBasis = append(vertexBasis, ZVertexInt(neighbor))
				
				if verbose {
					log.Printf("  Added vertex %d at depth %d", neighbor, current.depth + 1)
				}
			}
		}
	}
	
	// Add any unreachable vertices at the end (shouldn't happen for connected circulant graphs)
	for i := 0; i < n; i++ {
		if !visited[i] {
			vertexBasis = append(vertexBasis, ZVertexInt(i))
			if verbose {
				log.Printf("Added unreachable vertex %d", i)
			}
		}
	}
	
	if verbose {
		log.Printf("Generated Cayley graph with %d vertices and %d edges", len(vertexBasis), len(edgeBasis))
	}
	
	// Create 1-dimensional simplicial complex (no triangles)
	sortBases := false // Already ordered by BFS
	return NewZComplex(vertexBasis, edgeBasis, nil, sortBases, verbose), nil
}

// CirculantComplex creates a circulant clique complex by generating a circulant graph
// and then filling in all 3-cliques to create triangles.
func CirculantComplex(n int, steps []int, verbose bool) (*ZComplex[ZVertexInt], error) {
	complex, err := CirculantGraph(n, steps, verbose)
	if err != nil {
		return nil, err
	}
	if verbose {
		log.Printf("Filling 3-cliques in circulant graph")
	}
	complex.Fill3Cliques()
	if verbose {
		log.Printf("Created circulant clique complex with %d vertices, %d edges, %d triangles",
			len(complex.VertexBasis()), len(complex.EdgeBasis()), len(complex.TriangleBasis()))
	}
	return complex, nil
}
