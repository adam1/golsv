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


func RandomRegularGraphByBalancing(numVertices int, regularity int, maxIterations int, verbose bool) (*ZComplex[ZVertexInt], error) {
	if regularity < 0 || regularity >= numVertices {
		return nil, fmt.Errorf("regularity degree %d must be between 0 and %d", regularity, numVertices-1)
	}
	if numVertices * regularity % 2 != 0 {
		return nil, fmt.Errorf("cannot create %d-regular graph on %d vertices (n * k must be even)", regularity, numVertices)
	}
	if verbose {
		log.Printf("Generating regular graph by balancing over %d vertices with regularity degree %d", numVertices, regularity)
	}
	probEdge := float64(regularity) / float64(numVertices-1)
	if probEdge > 1.0 {
		probEdge = 1.0
	}
	G, err := RandomGraph(numVertices, probEdge, verbose)
	if err != nil {
		return nil, err
	}
	if verbose {
		degrees := vertexDegrees(G)
		log.Printf("Initial degree variance: %.3f", degreeVariance(degrees))
	}
	correctNumEdges(G, regularity, verbose)
	
	// seems we should always be able to construct a regular graph by
	// balancing.

	// xxx split the vertices into two slices: underweight and overweight
	unders, overs := splitVerticesByDegree(G)

	// xxx shuffle unders and overs

	for len(overs) > 0 {
		var o int // xxx shift from overs
		overage := G.Degree(o) - regularity
		for overage > 0 {
			oNbrs := G.Neighbors(o)
			for _, x := range oNbrs {
				for j, u := range unders {
					// xxx if x is not a neighbor of u, found an edge
					// to move.  move (o, x) to (o, u).
				}
			}
		}
	}
}

func correctNumEdges(G *ZComplex[ZVertexInt], regularity int, verbose bool) {
	delta := C.NumEdges() - numEdgesRegular(C.NumVertices(), regularity)
	if delta > 0 {
		if verbose {
			log.Printf("Pruning %d edges", delta)
		}
		panic("xxx Pruning edges not implemented")
		// xxx look first for edges between overweight vertices

		// xxx if none available, look for edges between half-overweight vertices
		
	} else if delta < 0 {
		if verbose {
			log.Printf("Adding %d edges", delta)
		}
		panic("xxx Adding edges not implemented")
		// xxx look for two underweight vertices

		// xxx if none available, panic
	}
}


// func vertexDegrees(adjacency []map[int]bool) []int {
// 	degrees := make([]int, len(adjacency))
// 	for v := 0; v < len(adjacency); v++ {
// 		degrees[v] = len(adjacency[v])
// 	}
// 	return degrees
// }

func vertexDegrees(C *ZComplex[ZVertexInt]) []int {
	degrees := make([]int, C.NumVertices())
	for v := 0; v < C.NumVertices(); v++ {
		degrees[v] = len(C.Neighbors(v)
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
