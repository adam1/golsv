package golsv

import (
	"log"
	"time"
)

// Following [EKZ]
//
// We have a simplicial complex $X = (V, E, T)$ with a set of vertices
// $V$, edges $E$, and triangles $T$, expressed as an $\F_2$ chain complex
//
//            \partial_2          \partial_1
//    \F_2^T -----------> \F_2^E -----------> \F_2^V
//
// and its dual chain complex
//
//            \delta_1             \delta_0
//    \F_2^T <----------- \F_2^E <----------- \F_2^V.
//
// Input: $\delta_1(\bra{e}) \in \F_2^E$
//        i.e. a 2-cochain that is a 1-coboundary
//        aka a set of triangles that are the coboundary of a set of edges
//        aka a Z-error syndrome
//
// Output: $\bra{c} \in B^1(X) = \im \delta_0$ such that $\bra{e+c}$
//         is equivalent to $\bra{e}$ mod $B^1(X)$, assuming that
//         $\bra{e}$ has small enough weight.  If $\bra{e}$ has too
//         high a weight, the algorithm may not return the correct
//         output.


type CoboundaryDecoder[T any] struct {
	complex *ZComplex[T]
	delta_0, delta_1 *DenseBinaryMatrix // xxx delta_0 could be deprecated
	verbose bool
	vertexToEdges map[int][]int
	edgeToTriangles map[int][]int
}

func NewCoboundaryDecoder[T any](complex *ZComplex[T], verbose bool) *CoboundaryDecoder[T] {
	D := &CoboundaryDecoder[T]{
		complex: complex,
		verbose: verbose,
	}
	if verbose {
		log.Printf("Computing delta_0")
	}
	D.delta_0 = D.complex.D1().Transpose().Dense()
	if verbose {
		log.Printf("Computing delta_1")
	}
	D.delta_1 = D.complex.D2().Transpose().Dense()
	if verbose {
		log.Printf("Computing vertex to edge incidence map")
	}
	D.vertexToEdges = D.complex.VertexToEdgeIncidenceMap()
	if verbose {
		log.Printf("Computing edge to triangle incidence map")
	}
	D.edgeToTriangles = D.complex.EdgeToTriangleIncidenceMap()
	return D
}

// xxx wip TEST

// xxx optimization plan: 
//
// - focus on the inner loop, enumeratePowerSet and its callback
//
// - we are currently using map keys as set elements, and maps as sets.
//   replace this with bit vectors. in particular, the current `f` can 
//   be represented by a bit vector of length k where k is initial number of
//   triangles in syndrome (i.e. the initial number of 1s in the syndrome).
//   note that this number only decreases as the algorithm proceeds. however,
//   it is NOT guaranteed that the triangles at each reduction step are a 
//   subset of the previous set of triangles (in principle.  i do not know if it
//   ever occurs that this is violated.) at each vertex step, we enumerate the 
//   power set of the incident edges, and by taking the coboundary (incident triangles) of
//   each of these edges, we have a maximum number of triangles to care about at each vertex step, 
//   specifically, k plus the number of triangles at this vertex not already counted in the original syndrome.
//
// - since our fastest bit vector implementation is the DenseBinaryMatrix,
//   we should use that, i.e. represent a bit vector as single column dense matrix.
//
// 

func (D *CoboundaryDecoder[T]) Decode(syndrome BinaryVector) (error BinaryVector) {
	if syndrome.Length() != D.complex.NumTriangles() {
		panic("syndrome length does not match number of triangles")
	}
	c := make(map[int]any) // keys are edge indices
	if syndrome.IsZero() {
		return NewBinaryVector(D.complex.NumEdges())
	}
	f := syndromeToSetMap(syndrome)
	fweight := len(f)
	sumIsZero := false
	log.Printf("Decoding syndrome; initial weight=%d", fweight)
	round := 0
	startTime := time.Now()
	lastStatTime := startTime
	statIntervalSteps := 1000
	for !sumIsZero {
		round++
		for i := range D.complex.VertexBasis() {
// 			if D.verbose {
// 				log.Printf("round=%d fweight=%d checking vertex %d/%d", round, fweight, i, D.complex.NumVertices()-1)
// 			}
			incidentEdges, ok := D.vertexToEdges[i]
			if !ok {
				panic("vertex not found in vertexToEdges map")
			}
			// for each 1-cochain y entirely inside the
			// edge-neighborhood $\partial_0(v)$
// 			if D.verbose {
// 				log.Printf("vertex %d has %d incident edges", i, len(incidentEdges))
// 			}
			foundY := false
			enumeratePowerSet(incidentEdges, func(y []int) bool {
				// compute the coboundary of y, i.e. the incident
				// triangles counting mod two

				delta_1y := coboundaryOfEdges(D.edgeToTriangles, y)
				// 				if D.verbose {
				// 					log.Printf("coboundary of y has weight %d", len(delta_1y))
				// 				}
				sum := symmetricDifference(f, delta_1y)
				sumWeight := len(sum)
				if sumWeight == 0 {
					sumIsZero = true
				}
				if sumWeight < fweight {
					foundY = true
					f = sum
					fweight = sumWeight
					c = symmetricDifferenceMS(c, y)
					log.Printf("found reducing cochain; yweight=%d newfweight=%d", len(y), fweight)
					return false
				}
				return true
			})
			if foundY {
				break
			}
			if i > 0 && i % statIntervalSteps == 0 {
				if D.verbose {
					now := time.Now()
					elapsed := now.Sub(lastStatTime)
					lastStatTime = now
					rate := float64(statIntervalSteps) / elapsed.Seconds()
					totalElapsed := now.Sub(startTime)
					totalRate := float64(i) / totalElapsed.Seconds()
					log.Printf("round=%d i=%d/%d rate=%1.3f telapsedh=%1.1f trate=%1.3f",
						round, i, len(D.complex.VertexBasis())-1, rate, totalElapsed.Hours(), totalRate)
				}
			}
		}
	}
	return setMapToVector(c, D.complex.NumEdges())
}

func setMapToVector(m map[int]any, length int) BinaryVector {
	v := NewBinaryVector(length)
	for k := range m {
		v[k] = 1
	}
	return v
}

// xxx test
func symmetricDifferenceMS(a map[int]any, b []int) map[int]any {
	res := make(map[int]any)
	for k := range a {
		res[k] = struct{}{}
	}
	for _, k := range b {
		if _, ok := a[k]; ok {
			delete(res, k)
		} else {
			res[k] = struct{}{}
		}
	}
	return res
}

// xxx test
func symmetricDifference(a, b map[int]any) map[int]any {
	res := make(map[int]any)
	for k := range a {
		if _, ok := b[k]; !ok {
			res[k] = struct{}{}
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			res[k] = struct{}{}
		}
	}
	return res
}

// returns a map of triangle indices; xxx move to ZComplex? xxx test
func syndromeToSetMap(syndrome BinaryVector) map[int]any {
	res := make(map[int]any)
	for i, v := range syndrome {
		if v == 1 {
			res[i] = struct{}{}
		}
	}
	return res
}

// returns a map of triangle indices; xxx move to ZComplex?
func coboundaryOfEdges(incidenceMap map[int][]int, edges []int) map[int]any {
	res := make(map[int]any)
	for _, e := range edges {
		incidentTriangles, ok := incidenceMap[e]
		if !ok {
			panic("edge not found in edgeToTriangles map")
		}
		for _, t := range incidentTriangles {
			if _, ok := res[t]; ok {
				delete(res, t)
			} else {
				res[t] = struct{}{}
			}
		}
	}
	return res
}

// courtesy of Claude
func enumeratePowerSet(keys []int, callback func([]int) bool ) {
	// sort.Ints(keys)
	var enumerate func(index int, current []int) bool
	enumerate = func(index int, current []int) bool {
		if index == len(keys) {
			return callback(current)
		}
		if !enumerate(index+1, current) {
			return false
		}
		newSubset := make([]int, len(current), len(current)+1)
		copy(newSubset, current)
		newSubset = append(newSubset, keys[index])
		return enumerate(index+1, newSubset)
	}
	enumerate(0, []int{})
}

func (D *CoboundaryDecoder[T]) Length() int {
	return D.complex.NumEdges()
}

func (D *CoboundaryDecoder[T]) Syndrome(error BinaryVector) BinaryVector {
	return D.delta_1.MultiplyRight(error.SparseBinaryMatrix()).ColumnVector(0)
}

type Decoder interface {
	Decode(syndrome BinaryVector) (error BinaryVector)
	Length() int
	Syndrome(error BinaryVector) BinaryVector
}

type DecoderSampler struct {
	decoder Decoder
	minErrorWeight int
	maxErrorWeight int
	samplesPerWeight int
	verbose bool
}

func NewDecoderSampler(decoder Decoder, minErrorWeight int, maxErrorWeight int, samplesPerWeight int, verbose bool) *DecoderSampler {
	return &DecoderSampler{
		decoder: decoder,
		minErrorWeight: minErrorWeight,
		maxErrorWeight: maxErrorWeight,
		samplesPerWeight: samplesPerWeight,
		verbose: verbose,
	}
}

func (S *DecoderSampler) Run() {
	n := S.decoder.Length()
	for weight := S.minErrorWeight; weight <= S.maxErrorWeight; weight++ {
		log.Printf("Sampling error weight=%d samples=%d", weight, S.samplesPerWeight)
		successCount := 0
		for i := 0; i < S.samplesPerWeight; i++ {
			error := NewBinaryVector(n)
			error.RandomizeWithWeight(weight)
			syndrome := S.decoder.Syndrome(error)
			before := time.Now()
			res := S.decoder.Decode(syndrome)
			now := time.Now()
			elapsed := now.Sub(before)
			// xxx this is checking for exact error decoding as opposed to coset ...
			if res.Equal(error) {
				successCount++
			}
			if S.verbose {
				log.Printf("decode success %d/%d dur=%d", successCount, i+1, int(elapsed.Seconds()))
			}
		}
	}
}




