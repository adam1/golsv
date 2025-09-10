package golsv

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// This file implements a construction from [LLR] (Lubotzky, Luria,
// Rosenthal 2019: Random Steiner systems and bounded degree
// coboundary expanders of every dimension).
//
// For 2D complexes, the construction works as follows:
//
// 1. Generate k independent (n,2)-Steiner systems S_1, ... S_k, using
//    the first (greedy/nibble) stage of Keevash's greedy algorithm.
//    Note in particular that we currently do not implement the second
//    (absorber/completion) stage at present.
//  
// 2. Form X_{n,k}^{(2)} = K_n^{1} ∪ S_1 ∪ ... ∪ S_k where K_n^{1} is
//    the complete 1-skeleton (complete graph) on n vertices.  Each
//    S_i consists of triangles such that every edge appears in
//    exactly one triangle.

type RandomCoboundaryExpanderGenerator struct {
	numVertices int  // n: number of vertices
	numSystems  int  // k: number of independent (n,2)-Steiner systems to union
	verbose     bool
}

func NewRandomCoboundaryExpanderGenerator(n, k int, verbose bool) *RandomCoboundaryExpanderGenerator {
	return &RandomCoboundaryExpanderGenerator{
		numVertices: n,
		numSystems:  k,
		verbose:     verbose,
	}
}

func (gen *RandomCoboundaryExpanderGenerator) Generate() (*ZComplex[ZVertexInt], error) {
	if gen.verbose {
		log.Printf("Generating 2D random coboundary expander: n=%d, k=%d\n", 
			gen.numVertices, gen.numSystems)
	}
	
	// Step 1: Generate k independent (n,2)-Steiner systems  
	allTriangles := make(map[ZTriangle[ZVertexInt]]bool)
	steinerGen := NewSteinerSystemGenerator(gen.numVertices, gen.verbose)
	
	for i := 0; i < gen.numSystems; i++ {
		if gen.verbose {
			log.Printf("Generating Steiner system %d/%d\n", i+1, gen.numSystems)
		}
		
		triangleMap, err := steinerGen.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate Steiner system %d: %v", i+1, err)
		}
		
		// Union triangles (map automatically deduplicates)
		for triangle := range triangleMap {
			allTriangles[triangle] = true
		}
		
		if gen.verbose {
			log.Printf("  Added %d triangles from system %d (total unique: %d)\n", 
				len(triangleMap), i+1, len(allTriangles))
		}
	}
	
	// Step 2: Create complete graph K_n^{(1)}
	completeGraph := CompleteGraph(gen.numVertices)
	
	// Step 3: Convert triangle map to slice
	triangleBasis := make([]ZTriangle[ZVertexInt], 0, len(allTriangles))
	for triangle := range allTriangles {
		triangleBasis = append(triangleBasis, triangle)
	}
	
	// Step 4: Build LLR complex X_{n,k}^{(2)} = K_n^{(1)} ∪ S_1 ∪ ... ∪ S_k
	sortBases := true
	complex := NewZComplex(
		completeGraph.VertexBasis(), 
		completeGraph.EdgeBasis(), 
		triangleBasis, 
		sortBases, 
		gen.verbose)
	
	if gen.verbose {
		log.Printf("Generated LLR complex with %d vertices, %d edges, %d triangles\n",
			len(complex.VertexBasis()), len(complex.EdgeBasis()), len(complex.TriangleBasis()))
	}
	
	return complex, nil
}

// SteinerSystemGenerator generates (n,2)-Steiner systems using the greedy
// stage of Keevash's construction
type SteinerSystemGenerator struct {
	numVertices int
	verbose     bool
}

func NewSteinerSystemGenerator(n int, verbose bool) *SteinerSystemGenerator {
	return &SteinerSystemGenerator{
		numVertices: n,
		verbose:     verbose,
	}
}

// Generate runs the greedy algorithm to produce a partial (n,2)-Steiner system.
// Returns a map of triangles such that every edge appears in at most one triangle.
func (g *SteinerSystemGenerator) Generate() (map[ZTriangle[ZVertexInt]]bool, error) {
	// Check that we have enough vertices to form triangles
	if g.numVertices < 3 {
		return nil, fmt.Errorf("cannot form Steiner Triple System with n=%d vertices (need n >= 3)", g.numVertices)
	}
	
	// Check divisibility condition: STS exists only when n ≡ 1 or 3 (mod 6)
	if g.numVertices%6 != 1 && g.numVertices%6 != 3 {
		return nil, fmt.Errorf("Steiner Triple System not possible for n=%d (n must be ≡ 1 or 3 mod 6)", g.numVertices)
	}
	
	// Check that n² doesn't overflow int (for edge encoding)
	if int64(g.numVertices)*int64(g.numVertices) > math.MaxInt {
		return nil, fmt.Errorf("n=%d too large (n² would overflow int)", g.numVertices)
	}
	
	if g.verbose {
		log.Printf("Generating (n,2)-Steiner system with n=%d vertices\n", g.numVertices)
	}
	
	// Edge encoding function: maps edge (i,j) where i < j to unique int
	edgeID := func(a, b ZVertexInt) int {
		i, j := int(a), int(b)
		if i > j {
			i, j = j, i
		}
		return i*g.numVertices + j
	}
	
	// Enumerate all possible triangles {i,j,k} with i < j < k
	allTriangles := make([]ZTriangle[ZVertexInt], 0)
	for i := 0; i < g.numVertices; i++ {
		for j := i + 1; j < g.numVertices; j++ {
			for k := j + 1; k < g.numVertices; k++ {
				triangle := NewZTriangle(ZVertexInt(i), ZVertexInt(j), ZVertexInt(k))
				allTriangles = append(allTriangles, triangle)
			}
		}
	}
	
	// Shuffle triangles for randomness
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allTriangles), func(i, j int) {
		allTriangles[i], allTriangles[j] = allTriangles[j], allTriangles[i]
	})
	
	// Greedily select legal triangles
	coveredEdges := make(map[int]bool)
	selectedTriangles := make(map[ZTriangle[ZVertexInt]]bool)
	
	for _, triangle := range allTriangles {
		// Get edge IDs for this triangle
		id1 := edgeID(triangle[0].(ZVertexInt), triangle[1].(ZVertexInt))
		id2 := edgeID(triangle[1].(ZVertexInt), triangle[2].(ZVertexInt))
		id3 := edgeID(triangle[2].(ZVertexInt), triangle[0].(ZVertexInt))
		
		// Check if triangle is legal (no edges already covered)
		if coveredEdges[id1] || coveredEdges[id2] || coveredEdges[id3] {
			continue
		}
		
		// Legal triangle - adopt it
		selectedTriangles[triangle] = true
		coveredEdges[id1] = true
		coveredEdges[id2] = true
		coveredEdges[id3] = true
	}
	
	if g.verbose {
		totalEdges := g.numVertices * (g.numVertices - 1) / 2
		coveredCount := len(coveredEdges)
		log.Printf("Generated %d triangles covering %d/%d edges\n", 
			len(selectedTriangles), coveredCount, totalEdges)
	}
	
	return selectedTriangles, nil
}
