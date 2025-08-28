package golsv

import (
	"fmt"
	"log"
)

// xxx write overview

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
