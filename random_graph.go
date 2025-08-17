package golsv

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
)	

func RandomGraph(numVertices int, probEdge float64, verbose bool) (*ZComplex[ZVertexInt], error) {
	if probEdge < 0 || probEdge > 1 {
		panic("pEdge must be between 0 and 1")
	}
	if verbose {
		log.Printf("Generating random graph over %d vertices with edge probability %v", numVertices, probEdge)
	}
	entropyBytesPerBit := 2 // nb. increase if more precision is needed in probEdge
	entropy := make([]byte, numVertices * entropyBytesPerBit)
	bytes := make([]byte, 8)
	maxInt := uint64(1)<<(entropyBytesPerBit * 8) - 1
	cutoff := uint64(probEdge * float64(maxInt))
	d_1Sparse := NewSparseBinaryMatrix(numVertices, 0).Sparse()
	numEdges := 0
	for i := 0; i < numVertices; i++ {
		numCols := numVertices - i - 1
		entropy = entropy[:numCols * entropyBytesPerBit]
		_, err := rand.Read(entropy)
		if err != nil {
			panic(err)
		}
		for k := 0; k < len(entropy); k += entropyBytesPerBit {
			copy(bytes, entropy[k:k+entropyBytesPerBit])
			num := binary.LittleEndian.Uint64(bytes)
			if num <= cutoff {
				M := NewSparseBinaryMatrix(numVertices, 1)
				M.Set(i, 0, 1)
				M.Set(i+k/entropyBytesPerBit+1, 0, 1)
				d_1Sparse.AppendColumn(M)
				numEdges++
			}
		}
	}
	C := NewZComplexFromBoundaryMatrices(d_1Sparse, NewSparseBinaryMatrix(numEdges, 0))
	if verbose {
		log.Printf("Generated d_1: %v\n", d_1Sparse)
	}
	return C, nil
}


func RandomRegularGraphByBalancing(numVertices int, degree int, maxIterations int, verbose bool) (*ZComplex[ZVertexInt], error) {
	k := degree
	if k < 0 || k >= numVertices {
		return nil, fmt.Errorf("regularity degree %d must be between 0 and %d", k, numVertices-1)
	}
	if numVertices*k%2 != 0 {
		return nil, fmt.Errorf("cannot create %d-regular graph on %d vertices (n*k must be even)", k, numVertices)
	}
	if verbose {
		log.Printf("Generating regular clique complex by balancing over %d vertices with regularity degree %d", numVertices, k)
	}
	probEdge := float64(k) / float64(numVertices-1)
	if probEdge > 1.0 {
		probEdge = 1.0
	}
	C, err := RandomGraph(numVertices, probEdge, verbose)
	if err != nil {
		return nil, err
	}

	// Convert to adjacency representation for easier manipulation
	// xxx okay for now; should perhaps use ZComplex methods
	adjacency := make([]map[int]bool, numVertices)
	for i := 0; i < numVertices; i++ {
		adjacency[i] = make(map[int]bool)
	}

	d1 := C.D1()
	for e := 0; e < d1.NumColumns(); e++ {
		vertices := make([]int, 0, 2)
		for v := 0; v < numVertices; v++ {
			if d1.Get(v, e) == 1 {
				vertices = append(vertices, v)
			}
		}
		if len(vertices) == 2 {
			u, v := vertices[0], vertices[1]
			adjacency[u][v] = true
			adjacency[v][u] = true
		}
	}

	if verbose {
		degrees := vertexDegrees(adjacency)
		log.Printf("Initial degree variance: %.3f", degreeVariance(degrees))
	}

	// Iteratively balance the graph
	for iteration := 0; iteration < maxIterations; iteration++ {
		degrees := vertexDegrees(adjacency)
		
		// Find vertex with highest degree > k
		highVertex := -1
		highDegree := k
		for v := 0; v < numVertices; v++ {
			if degrees[v] > highDegree {
				highVertex = v
				highDegree = degrees[v]
			}
		}
		
		if highVertex == -1 {
			// No vertex has degree > k, check if we're done
			allCorrect := true
			for v := 0; v < numVertices; v++ {
				if degrees[v] != k {
					allCorrect = false
					break
				}
			}
			if allCorrect {
				if verbose {
					log.Printf("Converged to %d-regular graph after %d iterations", k, iteration)
				}
				break
			}
			
			// Need to add edges to vertices with degree < k
			lowVertex := -1
			for v := 0; v < numVertices; v++ {
				if degrees[v] < k {
					lowVertex = v
					break
				}
			}
			if lowVertex == -1 {
				break // This shouldn't happen
			}
			
			// Find another vertex with degree < k to connect to lowVertex
			otherLowVertex := -1
			for v := 0; v < numVertices; v++ {
				if v != lowVertex && degrees[v] < k && !adjacency[lowVertex][v] {
					otherLowVertex = v
					break
				}
			}
			if otherLowVertex != -1 {
				adjacency[lowVertex][otherLowVertex] = true
				adjacency[otherLowVertex][lowVertex] = true
			}
			continue
		}

		// Find vertex with degree < k
		lowVertex := -1
		for v := 0; v < numVertices; v++ {
			if degrees[v] < k {
				lowVertex = v
				break
			}
		}
		if lowVertex == -1 {
			// All vertices have degree >= k, but highVertex has degree > k
			// Need to remove an edge from highVertex
			for neighbor := range adjacency[highVertex] {
				if degrees[neighbor] > k {
					// Remove edge between two high-degree vertices
					delete(adjacency[highVertex], neighbor)
					delete(adjacency[neighbor], highVertex)
					break
				}
			}
			continue
		}

		// Find a neighbor of highVertex that also has high degree
		highNeighbor := -1
		for neighbor := range adjacency[highVertex] {
			if degrees[neighbor] > k {
				highNeighbor = neighbor
				break
			}
		}
		
		if highNeighbor == -1 {
			// No high-degree neighbor found, just remove any edge from highVertex
			for neighbor := range adjacency[highVertex] {
				highNeighbor = neighbor
				break
			}
		}
		
		if highNeighbor != -1 && !adjacency[highVertex][lowVertex] {
			// Move edge from (highVertex, highNeighbor) to (highVertex, lowVertex)
			delete(adjacency[highVertex], highNeighbor)
			delete(adjacency[highNeighbor], highVertex)
			adjacency[highVertex][lowVertex] = true
			adjacency[lowVertex][highVertex] = true
			
			if verbose && iteration%10 == 0 {
				newDegrees := vertexDegrees(adjacency)
				log.Printf("Iteration %d: variance=%.3f, moved edge (%d,%d) -> (%d,%d)", 
					iteration, degreeVariance(newDegrees), highVertex, highNeighbor, highVertex, lowVertex)
			}
		}
	}

	// Convert back to boundary matrix representation
	d1Sparse := NewSparseBinaryMatrix(numVertices, 0).Sparse()
	for u := 0; u < numVertices; u++ {
		for v := range adjacency[u] {
			if u < v { // Only add each edge once
				M := NewSparseBinaryMatrix(numVertices, 1)
				M.Set(u, 0, 1)
				M.Set(v, 0, 1)
				d1Sparse.AppendColumn(M)
			}
		}
	}

	numEdges := d1Sparse.NumColumns()
	result := NewZComplexFromBoundaryMatrices(d1Sparse, NewSparseBinaryMatrix(numEdges, 0))
	
	if verbose {
		finalDegrees := vertexDegrees(adjacency)
		log.Printf("Final degree variance: %.3f", degreeVariance(finalDegrees))
		log.Printf("Generated graph with %d edges", numEdges)
		log.Printf("Filling cliques")
	}
	
	result.Fill3Cliques()
	return result, nil
}

func vertexDegrees(adjacency []map[int]bool) []int {
	degrees := make([]int, len(adjacency))
	for v := 0; v < len(adjacency); v++ {
		degrees[v] = len(adjacency[v])
	}
	return degrees
}

// xxx deprecated
// func degreesDistribution(vertexDegree []int) map[int]int {
// 	distribution := make(map[int]int)
// 	for _, degree := range vertexDegree {
// 		distribution[degree]++
// 	}
// 	return distribution
// }

func degreeVariance(degrees []int) float64 {
	if len(degrees) == 0 {
		return 0
	}
	
	// Calculate mean
	sum := 0
	for _, degree := range degrees {
		sum += degree
	}
	mean := float64(sum) / float64(len(degrees))
	
	// Calculate variance
	sumSquaredDiffs := 0.0
	for _, degree := range degrees {
		diff := float64(degree) - mean
		sumSquaredDiffs += diff * diff
	}
	
	return sumSquaredDiffs / float64(len(degrees))
}
