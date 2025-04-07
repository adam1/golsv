package golsv

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	delta_0, delta_1, Z_1T *DenseBinaryMatrix // xxx delta_0 could be deprecated?
	verbose bool
	vertexToEdges map[int][]int
	edgeToTriangles map[int][]int
}

func NewCoboundaryDecoder[T any](complex *ZComplex[T], Z_1 BinaryMatrix, verbose bool) *CoboundaryDecoder[T] {
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
		log.Printf("Computing Z_1^T")
	}
	D.Z_1T = Z_1.Transpose().Dense()
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

func (D *CoboundaryDecoder[T]) Decode(syndrome BinaryVector) (err error, errorVec BinaryVector) {
	// Let the complex X = (V, E, T) with vertex set V, edge set E,
	// triangle set T.  Let f be the current syndrome, whose weight we
	// are decreasing at each step.
	if syndrome.Length() != D.complex.NumTriangles() {
		panic("syndrome length does not match number of triangles")
	}
	if syndrome.IsZero() {
		return nil, NewBinaryVector(D.complex.NumEdges())
	}
	curError := NewBinaryVector(D.complex.NumEdges())
	f := syndrome
	fWeight := f.Weight()
	originalFWeight := fWeight
	log.Printf("Begin decoding syndrome; fWeight=%d", fWeight)
	round := 0
	startTime := time.Now()
	lastStatTime := startTime
	lastFWeight := fWeight
	statIntervalSteps := 100
	// NOTE: triLocalToGlobal is persisted and modified across
	// iterations of the outer loop here.  This is part of an
	// optimization where we continually reduce the triangle space
	// (and hence the length of the syndrome vector). We start with
	// the full space and at each step remove more triangles (by
	// projecting) that are not pertinent.
	var triLocalToGlobal map[int]int
	for fWeight > 0 {
		round++
		vertexIndices := D.chooseVertices(f, triLocalToGlobal)
		log.Printf("Checking %d vertices", len(vertexIndices))
		prevWeight := fWeight
		for i, k := range vertexIndices {
			f, fWeight, triLocalToGlobal = D.processVertex(f, fWeight, i, k, triLocalToGlobal, curError)
			if i > 0 && i % statIntervalSteps == 0 && D.verbose {
				now := time.Now()
				elapsed := now.Sub(lastStatTime)
				lastStatTime = now
				weightReductionRate := float64(lastFWeight - fWeight) / elapsed.Seconds()
				avgWeightReductionRate := float64(originalFWeight - fWeight) / (now.Sub(startTime).Seconds())
				etah := float64(fWeight) / avgWeightReductionRate / 3600.0
				log.Printf("round=%d fweight=%d wredrate=%1.1f avgwredrate=%1.1f etah=%1.1f",
					round, fWeight, weightReductionRate, avgWeightReductionRate, etah)
				lastFWeight = fWeight
			}
			if fWeight == 0 {
				break
			}
		}
		if fWeight == prevWeight {
			// we completed a scan of vertices without making
			// progress.  the algorithm has failed.
			err := fmt.Errorf("failed to make progress at weight=%d", fWeight)
			log.Print(err)
			return err, NewBinaryVector(D.complex.NumEdges())
		}
	}
	return nil, curError
}

// f is the current (projected) syndrome.
func (D *CoboundaryDecoder[T]) processVertex(f BinaryVector, fWeight int, iterationIndex int, basisIndex int, triLocalToGlobal map[int]int, curError BinaryVector) (newF BinaryVector, newFWeight int, newTriLocalToGlobal map[int]int) {
	incidentEdges, ok := D.vertexToEdges[basisIndex]
	if !ok {
		panic("vertex not found in vertexToEdges map")
	}
	incidentEdgesBits := NewBinaryVector(len(incidentEdges))
	incidentTriangles := D.incidentTriangles(incidentEdges)
	// Let $S = support(f) \subset T$, and let $T_v \subset T$ be the
	// triangles incident to $v$.  The triangle set $S \cup T_v$
	// represents the set of triangles that we might be interested in
	// for this step and this vertex in the procedure.  We project
	// from the full triangle space $\F_2^T$ down to $\F_2^{S \cup
	// T_v}$, in order to minimize the size of bitstrings we are
	// dealing with. Calling the full triangle space global and the
	// smaller reduced triangle space local, the triLocalToGlobal and
	// triGlobalToLocal maps translate indices of these two arrays.
	var triGlobalToLocal map[int]int
	f, triLocalToGlobal, triGlobalToLocal = D.localProjection(f, triLocalToGlobal, incidentTriangles)
	// for each 1-cochain y entirely inside the edge-neighborhood
	// $\partial_0(v)$
	sum := NewBinaryVector(f.Length())
	j := 0
	EnumerateBinaryVectors(len(incidentEdges), incidentEdgesBits, func() bool {
		j++
		if incidentEdgesBits.Weight() == 0 {
			return true
		}
		y := incidentEdgesBits
		// compute the coboundary of y, i.e. the incident triangles
		// counting mod two.  and add to f.
		sum.Clear()
		D.coboundaryOfEdges(y, incidentEdges, triGlobalToLocal, sum)
		sum.Sum(sum, f)
		sumWeight := sum.Weight()
		if sumWeight < fWeight {
			f = sum
			fWeight = sumWeight
			D.toggleEdges(curError, y, incidentEdges)
			if false && D.verbose {
				log.Printf("Found reducing cochain; vbidx=%d vitidx=%d edcchidx=%d ywt=%d errwt=%d fwt=%d",
					basisIndex, iterationIndex, j, y.Weight(), curError.Weight(), fWeight)
			}
			return false
		}
		return true
	})
	return f, fWeight, triLocalToGlobal
}

// f is the current syndrome in the current (reduced/projected)
// triangle space.  chooseVertices returns the list of vertices that
// are incident to any triangle in the support of f.
func (D *CoboundaryDecoder[T]) chooseVertices(f BinaryVector, triLocalToGlobal map[int]int) []int {
	res := make([]int, 0)
	triangleBasis := D.complex.TriangleBasis()
	vertexIndex := D.complex.VertexIndex()
	set := NewBinaryVector(D.complex.NumVertices())
	for i := 0; i < f.Length(); i++ {
		if f.Get(i) == 1 {
			k := D.unprojectTriangleIndex(i, triLocalToGlobal)
			// lookup the incident vertices to this triangle
			triangle := triangleBasis[k]
			for _, v := range triangle {
				m, ok := vertexIndex[v]
				if !ok {
					panic("vertex not found in vertexIndex")
				}
				if set.Get(m) == 0 {
					set.Set(m, 1)
					res = append(res, m)
				}
			}
		}
	}
	return res
}

// xxx test
func (D *CoboundaryDecoder[T]) incidentTriangles(edgeIndices []int) []int {
	set := make(map[int]struct{})
	for _, e := range edgeIndices {
		incidentTriangles, ok := D.edgeToTriangles[e]
		if !ok {
			panic("edge not found in edgeToTriangles map")
		}
		for _, t := range incidentTriangles {
			set[t] = struct{}{}
		}
	}
	return intSetToSlice(set)
}

func intSetToSlice(set map[int]struct{}) []int {
	res := make([]int, 0, len(set))
	for k := range set {
		res = append(res, k)
	}
	return res
}

// xxx test
func (D *CoboundaryDecoder[T]) localProjection(f BinaryVector, triLocalToGlobal map[int]int, incidentTriangles []int) (v BinaryVector, triLocalToGlobalNew map[int]int, triGlobalToLocal map[int]int) {
	// log.Printf("xxx localProjection l(f)=%d w(f)=%d", f.Length(), f.Weight())
	// starting from all triangles T, take the union of:
	// - triangles in the support of f (use triLocalToGlobal to translate indices)
	// - incident triangles (these are indices in T)
	u := NewBinaryVector(D.complex.NumTriangles())
	for i := 0; i < f.Length(); i++ {
		if f.Get(i) == 1 {
			k := D.unprojectTriangleIndex(i, triLocalToGlobal)
			u.Set(k, 1)
		}
	}
	// log.Printf("xxx l(u)=%d w(u)=%d", u.Length(), u.Weight())
	incidentTrianglesMap := make(map[int]struct{})
	delta := 0
	for _, t := range incidentTriangles {
		incidentTrianglesMap[t] = struct{}{}
		if u.Get(t) == 0 {
			delta++
		}
	}
// 	log.Printf("xxx w(u)=%d", u.Weight())
	// project and produce new local-to-global and global-to-local
	// maps
	n := u.Weight() + delta
	triLocalToGlobalNew = make(map[int]int)
	triGlobalToLocal = make(map[int]int)
	localIndex := 0
	v = u.Project(n, func(i int) bool {
		_, isIncident := incidentTrianglesMap[i]
		if isIncident || u.Get(i) == 1 {
			triLocalToGlobalNew[localIndex] = i
			triGlobalToLocal[i] = localIndex
			localIndex++
			return true
		}
		return false
	})
	return v, triLocalToGlobalNew, triGlobalToLocal
}

func (D *CoboundaryDecoder[T]) unprojectTriangleIndex(i int, triLocalToGlobal map[int]int) int {
	// special case: if triLocalToGlobal is empty, then we are at the
	// first step and haven't projected down yet, so the indexing is
	// unchanged.
	var k int
	if len(triLocalToGlobal) == 0 {
		k = i
	} else {
		var ok bool
		k, ok = triLocalToGlobal[i]
		if !ok {
			panic("triangle not found in triLocalToGlobal")
		}
	}
	return k
}

// xxx test
func (D *CoboundaryDecoder[T]) coboundaryOfEdges(edgeBits BinaryVector, edgeSet []int, triGlobalToLocal map[int]int, resultBuf BinaryVector) {
	// log.Printf("xxx coboundaryOfEdges w=%d", edgeBits.Weight())
	for i := 0; i < edgeBits.Length(); i++ {
		if edgeBits.Get(i) == 0 {
			continue
		}
		edgeIndex := edgeSet[i]
		incidentTriangles, ok := D.edgeToTriangles[edgeIndex]
		if !ok {
			panic("edge not found in edgeToTriangles map")
		}
		for _, t := range incidentTriangles {
			tBitIndex, ok := triGlobalToLocal[t]
			if !ok {
				panic(fmt.Sprintf("triangle %d not found in triGlobalToLocal", t))
			}
			resultBuf.Toggle(tBitIndex)
		}
	}
}

// xxx test
func (D *CoboundaryDecoder[T]) toggleEdges(c BinaryVector, y BinaryVector, edges []int) {
	for i := 0; i < y.Length(); i++ {
		if y.Get(i) == 1 {
			edgeIndex := edges[i]
			c.Toggle(edgeIndex)
		}
	}
}

func (D *CoboundaryDecoder[T]) Length() int {
	return D.complex.NumEdges()
}

func (D *CoboundaryDecoder[T]) SameCoset(e, c BinaryVector) bool {
	u := e.Add(c)
	// Check whether diff u is in the image of $\delta_0$.
	// 
	// Abusing notation, let $Z_1$ also stand for the matrix with
	// columns that generate $Z_1$ the subspace of $\F_2^E$, $E$ edge
	// set.
	//
	// $\im \delta_0 = (\ker \partial_1)^\perp$, so
	//
	// $u^T Z_1 = 0 \implies u \in (\ker \partial_1)^\perp = \im \delta_0$.
	//
	// But $u^T Z_1 = 0 \iff (u^T Z_1)^T = 0$
	//
	// and $(u^T Z_1)^T = Z_1^T u$.  The latter is a fast
	// multiplication if $Z_1$ is dense and $u$ is sparse.
	Z_1Tu := D.Z_1T.MultiplyRight(u.SparseBinaryMatrix())
	return Z_1Tu.IsZero()
}

func (D *CoboundaryDecoder[T]) Syndrome(error BinaryVector) BinaryVector {
	return D.delta_1.MultiplyRight(error.SparseBinaryMatrix()).ColumnVector(0)
}

type Decoder interface {
	Decode(syndrome BinaryVector) (err error, errorVec BinaryVector)
	Length() int
	SameCoset(e, c BinaryVector) bool
	Syndrome(error BinaryVector) BinaryVector
}

type DecoderSampler struct {
	decoder Decoder
	minErrorWeight int
	maxErrorWeight int
	samplesPerWeight int
	results DecoderSamplerResults
	resultsFilename string
	verbose bool
}

type DecoderSamplerResults struct {
	ErrorWeight int
	SuccessCount int
	FailCount int
	EqualCount int
	SameCosetCount int
}

func NewDecoderSampler(decoder Decoder, minErrorWeight int, maxErrorWeight int, samplesPerWeight int, resultsFilename string, verbose bool) *DecoderSampler {
	return &DecoderSampler{
		decoder: decoder,
		minErrorWeight: minErrorWeight,
		maxErrorWeight: maxErrorWeight,
		samplesPerWeight: samplesPerWeight,
		resultsFilename: resultsFilename,
		verbose: verbose,
	}
}

func (S *DecoderSampler) Run() {
	n := S.decoder.Length()
	for weight := S.minErrorWeight; weight <= S.maxErrorWeight; weight++ {
		log.Printf("Sampling error weight=%d samples=%d", weight, S.samplesPerWeight)
		S.results.ErrorWeight = weight
		S.results.SuccessCount = 0
		S.results.EqualCount = 0
		S.results.SameCosetCount = 0
		S.results.FailCount = 0
		for i := 0; i < S.samplesPerWeight; i++ {
			errorVec := NewBinaryVector(n)
			errorVec.RandomizeWithWeight(weight)
			syndrome := S.decoder.Syndrome(errorVec)
			before := time.Now()
			err, decodedErrorVec := S.decoder.Decode(syndrome)
			status := "failure"
			now := time.Now()
			elapsed := now.Sub(before)
			if err != nil {
				S.results.FailCount++
			} else if decodedErrorVec.Equal(errorVec) {
				status = "success"
				S.results.SuccessCount++
				S.results.EqualCount++
			} else if S.decoder.SameCoset(errorVec, decodedErrorVec) {
				status = "success"
				S.results.SuccessCount++
				S.results.SameCosetCount++
			} else {
				S.results.FailCount++
			}
			if S.verbose {
				log.Printf("Decode %s: errweight=%d equal=%d sameCoset=%d fail=%d success=%d/%d/%d successrate=%1.1f%% dur=%d",
					status, weight, S.results.EqualCount, S.results.SameCosetCount, S.results.FailCount, S.results.SuccessCount, i+1, S.samplesPerWeight,
					float64(S.results.SuccessCount*100)/float64(i+1), int(elapsed.Seconds()))
			}
			S.writeResultsFile()
		}
	}
}

func (S *DecoderSampler) writeResultsFile() {
	if S.resultsFilename == "" {
		return
	}
	jsonData, err := json.MarshalIndent(S.results, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(S.resultsFilename, jsonData, 0644)
	if err != nil {
		panic(err)
	}
	log.Printf("Wrote file %s", S.resultsFilename)
}

