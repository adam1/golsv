package golsv

import (
	"fmt"
	"log"
)

// Original BoundaryDecoder - commented out while developing SSF version
/*
type BoundaryDecoder[T any] struct {
	graph *ZComplex[T]
	partial_1 *DenseBinaryMatrix
	vertexToEdges map[int][]int
	verbose bool
}

func NewBoundaryDecoder[T any](graph *ZComplex[T], verbose bool) *BoundaryDecoder[T] {
	D := &BoundaryDecoder[T]{
		graph: graph,
		verbose: verbose,
	}
	if verbose {
		log.Printf("Computing partial_1")
	}
	D.partial_1 = D.graph.D1().Dense()
	if verbose {
		log.Printf("Computing vertex to edge incidence map")
	}
	D.vertexToEdges = D.graph.VertexToEdgeIncidenceMap()
	return D
}

func (D *BoundaryDecoder[T]) Decode(syndrome BinaryVector) (err error, errorVec BinaryVector) {
	// Let G = (V, E) be a graph. Typically, G may in fact be a 2-d
	// simplicial complex, but this algorithm does not use anything
	// above dimension 1 in the complex. Let f be the current
	// syndrome, whose weight we are decreasing at each step.
	if syndrome.Length() != D.graph.NumVertices() {
		panic("syndrome length does not match the number of vertices")
	}
	curError := NewBinaryVector(D.graph.NumEdges())
	if syndrome.IsZero() {
		return nil, curError
	}
	f := syndrome
	fWeight := f.Weight()
// 	originalFWeight := fWeight
	vertexBasis := D.graph.VertexBasis()

	log.Printf("Begin decoding syndrome; fWeight=%d", fWeight)
	round := 0
// 	startTime := time.Now()
// 	lastStatTime := startTime
// 	lastFWeight := fWeight
// 	statIntervalSteps := 100

	for fWeight > 0 {
		log.Printf("xxx round %d; fWeight=%d", round, fWeight)
		round++
		found := false
		for i, _ := range vertexBasis {
			// let S = Star(v) \subset E
			incidentEdges := D.vertexToEdges[i]
			nbrsf := 0
			halfDegree := float64(len(incidentEdges)) / 2
			vertex_i := vertexBasis[i]
			vertexIndex := D.graph.VertexIndex()
			edgeBasis := D.graph.EdgeBasis()
			
			for _, edgeIndex := range incidentEdges {
				edge := edgeBasis[edgeIndex]
				otherVertex := edge.OtherVertex(vertex_i)
				neighborIndex := vertexIndex[otherVertex]
				if f.Get(neighborIndex) == 1 {
					nbrsf++
				}
			}
			//log.Printf("xxx vertex %d, halfDegree=%f nbrsf=%d", i, halfDegree, nbrsf)
			if float64(nbrsf) > halfDegree {
				//log.Printf("xxx flipping")
				// set f = f + \partial_1(S)
				for _, edgeIndex := range incidentEdges {
					edge := edgeBasis[edgeIndex]
					otherVertex := edge.OtherVertex(vertex_i)
					neighborIndex := vertexIndex[otherVertex]
					f.Toggle(neighborIndex)
				}
				if len(incidentEdges) % 2 == 1 {
					f.Toggle(i)
				}
				fWeight = f.Weight()

				// set curError = curError + S
				for _, edgeIndex := range incidentEdges {
					curError.Toggle(edgeIndex)
				}
				found = true
			}
		}
		if !found {
			err := fmt.Errorf("failed to make progress at weight=%d", fWeight)
			log.Print(err)
			return err, NewBinaryVector(D.graph.NumEdges())
		}
	}
	return nil, curError
}

func (D *BoundaryDecoder[T]) Length() int {
	return D.graph.NumEdges()
}

// e is typically an error vector and c is a decoded error vector
func (D *BoundaryDecoder[T]) SameCoset(e, c BinaryVector) bool {
	u := e.Add(c)
	partial1u := D.partial_1.MultiplyRight(u.SparseBinaryMatrix())
	return partial1u.IsZero()
}

func (D *BoundaryDecoder[T]) Syndrome(error BinaryVector) BinaryVector {
	return D.partial_1.MultiplyRight(error.SparseBinaryMatrix()).ColumnVector(0)
}
*/

// SSFBoundaryDecoder implements the Small Set Flip (SSF) decoding algorithm
type SSFBoundaryDecoder[T any] struct {
	graph *ZComplex[T]
	partial_1 *DenseBinaryMatrix
	vertexToEdges map[int][]int
	verbose bool
}

func NewSSFBoundaryDecoder[T any](graph *ZComplex[T], verbose bool) *SSFBoundaryDecoder[T] {
	D := &SSFBoundaryDecoder[T]{
		graph: graph,
		verbose: verbose,
	}
	if verbose {
		log.Printf("Computing partial_1")
	}
	D.partial_1 = D.graph.D1().Dense()
	if verbose {
		log.Printf("Computing vertex to edge incidence map")
	}
	D.vertexToEdges = D.graph.VertexToEdgeIncidenceMap()
	return D
}

func (D *SSFBoundaryDecoder[T]) Decode(syndrome BinaryVector) (err error, errorVec BinaryVector) {
	// Let G = (V, E) be a graph. Typically, G may in fact be a 2-d
	// simplicial complex, but this algorithm does not use anything
	// above dimension 1 in the complex. Let f be the current
	// syndrome, whose weight we are decreasing at each step.
	if syndrome.Length() != D.graph.NumVertices() {
		panic("syndrome length does not match the number of vertices")
	}
	curError := NewBinaryVector(D.graph.NumEdges())
	if syndrome.IsZero() {
		return nil, curError
	}
	f := syndrome
	fWeight := f.Weight()
	vertexBasis := D.graph.VertexBasis()

	log.Printf("Begin SSF decoding syndrome; fWeight=%d", fWeight)
	round := 0

	for fWeight > 0 {
		//log.Printf("xxx round %d; fWeight=%d", round, fWeight)
		round++
		found := false
		for i, _ := range vertexBasis {
			if D.processVertex(i, &f, &curError) {
				found = true
				fWeight = f.Weight()
			}
		}
		if !found {
			err := fmt.Errorf("failed to make progress at weight=%d", fWeight)
			log.Print(err)
			return err, NewBinaryVector(D.graph.NumEdges())
		}
	}
	return nil, curError
}

// processVertex handles the decoding logic for a single vertex
// Returns true if any edges were flipped, false otherwise
func (D *SSFBoundaryDecoder[T]) processVertex(vertexIndex int, f *BinaryVector, curError *BinaryVector) bool {
	// let S = Star(v) \subset E.  we'll check every subset of
	// Star(v).  if it reduces the syndrome, we'll use it.
	incidentEdges := D.vertexToEdges[vertexIndex]
	incidentEdgesBits := NewBinaryVector(len(incidentEdges))
	sum := NewBinaryVector(f.Length())
	j := 0
	found := false
	fWeight := f.Weight()
	EnumerateBinaryVectors(len(incidentEdges), incidentEdgesBits, func() bool {
		j++
		if incidentEdgesBits.Weight() == 0 {
			return true
		}
		y := incidentEdgesBits
		// compute the boundary of y (i.e. the incident vertices
		// counting mod two), and add to f.
		sum.Clear()
		D.boundaryOfEdges(y, incidentEdges, sum)
		sum.Sum(sum, *f)
		sumWeight := sum.Weight()
		if sumWeight < fWeight {
			*f = sum
			D.toggleEdges(curError, y, incidentEdges)
			found = true
			return false
		}
		return true
	})
	return found
}

func (D *SSFBoundaryDecoder[T]) Length() int {
	return D.graph.NumEdges()
}

// e is typically an error vector and c is a decoded error vector
func (D *SSFBoundaryDecoder[T]) SameCoset(e, c BinaryVector) bool {
	u := e.Add(c)
	partial1u := D.partial_1.MultiplyRight(u.SparseBinaryMatrix())
	return partial1u.IsZero()
}

func (D *SSFBoundaryDecoder[T]) Syndrome(error BinaryVector) BinaryVector {
	return D.partial_1.MultiplyRight(error.SparseBinaryMatrix()).ColumnVector(0)
}

// boundaryOfEdges computes the boundary of a subset of edges
// For each edge in the subset, toggle the incident vertices in resultBuf
func (D *SSFBoundaryDecoder[T]) boundaryOfEdges(edgeBits BinaryVector, edgeSet []int, resultBuf BinaryVector) {
	edgeBasis := D.graph.EdgeBasis()
	vertexIndex := D.graph.VertexIndex()
	
	for i := 0; i < edgeBits.Length(); i++ {
		if edgeBits.Get(i) == 0 {
			continue
		}
		edgeIdx := edgeSet[i]
		edge := edgeBasis[edgeIdx]
		
		// Toggle both vertices of this edge
		v1Index := vertexIndex[edge[0]]
		v2Index := vertexIndex[edge[1]]
		resultBuf.Toggle(v1Index)
		resultBuf.Toggle(v2Index)
	}
}

// toggleEdges toggles edges in the current error vector
func (D *SSFBoundaryDecoder[T]) toggleEdges(curError *BinaryVector, edgeBits BinaryVector, edgeSet []int) {
	for i := 0; i < edgeBits.Length(); i++ {
		if edgeBits.Get(i) == 0 {
			continue
		}
		edgeIdx := edgeSet[i]
		curError.Toggle(edgeIdx)
	}
}
