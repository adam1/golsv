package golsv

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	mathrand "math/rand"
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

	// We can always construct a regular graph by balancing.
	unders, overs := splitVerticesByDegree(G, regularity)
	if verbose {
		log.Printf("Initiating degree balancing")
	}
	mathrand.Shuffle(len(unders), func(i, j int) { unders[i], unders[j] = unders[j], unders[i] })
	mathrand.Shuffle(len(overs), func(i, j int) { overs[i], overs[j] = overs[j], overs[i] })

	for len(overs) > 0 {
		o := overs[0] // shift from overs
		overs = overs[1:]
		
		overage := G.Degree(o) - regularity
		for overage > 0 {
			oNbrs := G.Neighbors(o)
			moved := false
			
			for _, x := range oNbrs {
				for j, u := range unders {
					if u != x && !G.IsNeighbor(x, u) {
						// move (o, x) to (u, x)
						eidx, ok := G.IndexOfEdge(o, x)
						if !ok {
							panic("Edge not found")
						}
						G.MoveEdge(eidx, u, x)
						overage--
						moved = true
						
						// if u is no longer underweight, remove from unders
						if G.Degree(u) >= regularity {
							unders = append(unders[:j], unders[j+1:]...)
						}
						break
					}
				}
				if moved {
					break
				}
			}
			if !moved {
				// shouldn't happen
				return nil, fmt.Errorf("failed to find an edge to move")
			}
		}
	}
	return G, nil
}

func deleteRandomEdge(G *ZComplex[ZVertexInt]) bool {
	edges := G.EdgeBasis()
	if len(edges) == 0 {
		return false
	}
	edgeIdx := mathrand.Intn(len(edges))
	G.DeleteEdge(edgeIdx)
	return true
}

func addRandomEdge(G *ZComplex[ZVertexInt]) bool {
	// Find all possible edges to add
	possibleEdges := make([][2]int, 0)
	for u := 0; u < G.NumVertices(); u++ {
		for v := u + 1; v < G.NumVertices(); v++ {
			if !G.IsNeighbor(u, v) {
				possibleEdges = append(possibleEdges, [2]int{u, v})
			}
		}
	}
	if len(possibleEdges) == 0 {
		return false
	}
	// Pick a random edge to add
	edgeIdx := mathrand.Intn(len(possibleEdges))
	edge := possibleEdges[edgeIdx]
	G.AddEdge(edge[0], edge[1])
	return true
}

func correctNumEdges(G *ZComplex[ZVertexInt], regularity int, verbose bool) {
	delta := G.NumEdges() - numEdgesRegular(G.NumVertices(), regularity)
	if delta > 0 {
		if verbose {
			log.Printf("Pruning %d edges", delta)
		}
		for delta > 0 {
			if !deleteRandomEdge(G) {
				panic("could not find edge to remove")
			}
			delta--
		}
	} else if delta < 0 {
		if verbose {
			log.Printf("Adding %d edges", -delta)
		}
		for delta < 0 {
			if !addRandomEdge(G) {
				panic("could not find vertices to connect")
			}
			delta++
		}
	}
}

func vertexDegrees(C *ZComplex[ZVertexInt]) []int {
	degrees := make([]int, C.NumVertices())
	for v := 0; v < C.NumVertices(); v++ {
		degrees[v] = len(C.Neighbors(v))
	}
	return degrees
}

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

func numEdgesRegular(numVertices, regularity int) int {
	return numVertices * regularity / 2
}

func splitVerticesByDegree(G *ZComplex[ZVertexInt], regularity int) ([]int, []int) {
	var unders, overs []int
	for v := 0; v < G.NumVertices(); v++ {
		degree := len(G.Neighbors(v))
		if degree < regularity {
			unders = append(unders, v)
		} else if degree > regularity {
			overs = append(overs, v)
		}
	}
	return unders, overs
}
