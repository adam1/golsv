package golsv

import (
	"fmt"
	"log"
	"math"
	"time"
)

// We assume that the aligned basis for the p^th degree chain space C
// = C_p has already been computed, and so we have matrices U = U_p
// and B = B_p such that
//
//   C = Y \oplus U \oplus B
//
// where Z = Z_p = Y \oplus B, and we have identified a matrix with
// its column space.
//
// By definition, the p^th degree systole of the complex is the
// minimum weight of a vector in Z \setminus B.  Hence, any vector
// that is a non-trivial linear combination of the columns of U plus a
// linear combination of the columns of B is a systolic candidate.
// The two functions below implement a random search and an exhaustive
// search for the minimum weight of such a vector.

func SystoleRandomSearch(U, B BinaryMatrix, trials int, verbose bool) (minWeight int, minVector BinaryVector) {
	if U.NumColumns() == 0 {
		return 0, BinaryVector{}
	}
	reportInterval := 10
	timeStart := time.Now()
	timeLast := timeStart
	if verbose {
		log.Printf("computing minimum nonzero weight of columns of U")
	}
	minWeight = math.MaxInt
	for j := 0; j < U.NumColumns(); j++ {
		weight := U.ColumnWeight(j)
		if weight == 0 {
			panic(fmt.Sprintf("column %d of U weight is zero", j))
		}
		if weight < minWeight {
			minWeight = weight
			minVector = U.ColumnVector(j)
			if verbose {
				log.Printf("new min weight: %d", minWeight)
			}
		}
	}
	if verbose {
		log.Printf("sampling minimum nonzero weight for %d trials", trials)
	}

	var a *DenseBinaryMatrix
	for n := 0; n < trials; n++ {
		for {
			a = RandomLinearCombination(U).(*DenseBinaryMatrix)
			if !a.IsZero() {
				break
			}
		}
		if B.NumColumns() > 0 {
			b := RandomLinearCombination(B)
			a.AddMatrix(b)
		}
		weight := a.ColumnWeight(0)
		if weight == 0 {
			panic(fmt.Sprintf("random linear combination of B and U is zero"))
		}
		if weight < minWeight {
			minWeight = weight
			minVector = a.ColumnVector(0)
			if verbose {
				log.Printf("new min weight: %d", minWeight)
			}
		}
		if n > 0 && n%reportInterval == 0 {
			timeNow := time.Now()
			timeElapsed := timeNow.Sub(timeStart)
			timeInterval := timeNow.Sub(timeLast)
			timeLast = timeNow
			if verbose {
				log.Printf("trial %d/%d (%.2f%%) minwt=%d crate=%.2f trate=%.2f",
					n, trials, 100.0*float64(n)/float64(trials), minWeight,
					float64(reportInterval)/timeInterval.Seconds(),
					float64(n)/timeElapsed.Seconds())
			}
			if timeInterval.Seconds() < 10 {
				reportInterval = int(float64(reportInterval)*1.5)
			}
		}
	}
	return minWeight, minVector
}

func SystoleExhaustiveSearch(U, B BinaryMatrix, verbose bool) (minWeight int, minVector BinaryVector) {
	if U.NumColumns() == 0 {
		return 0, BinaryVector{}
	}
	minWeight = math.MaxInt
	EnumerateBinaryVectorSpace(U, func(a BinaryMatrix, indexU int) bool {
		if a.IsZero() {
			return true
		}
		EnumerateBinaryVectorSpace(B, func(b BinaryMatrix, indexB int) bool {
			sum := a.Copy().Dense()
			sum.AddMatrix(b)
			weight := sum.ColumnWeight(0)
			if weight < minWeight {
				minWeight = weight
				minVector = sum.ColumnVector(0)
				if verbose {
					log.Printf("exhaustive search; new min weight: %d", minWeight)
					//log.Printf("c: %s", sum.ColumnVector(0).SupportString())
				}
			}
			return true
		})
		return true
	})
	if minWeight == math.MaxInt {
		return 0, BinaryVector{}
	}
	return minWeight, minVector
}

// ComputeFirstSystole computes the degree one systole of the complex.
// All processing is done in memory, hence the function is only
// suitable for small complexes.  For larger complexes, use the
// procedure represented in worksets/Makefile, which uses the
// individual programs and intermediate files at each step.
// ComputeFirstSystole is useful for small complexes and for testing.
func ComputeFirstSystole(d1, d2 BinaryMatrix, verbose bool) (systole, dimZ1, dimB1, dimH1 int) {
	if verbose {
		log.Printf("Computing first homology")
	}
	var U, B BinaryMatrix
	U, B, _, dimZ1, dimB1, dimH1 = UBDecomposition(d1, d2, verbose)
	U, B = U.Dense(), B.Dense()
	systole, _ = SystoleExhaustiveSearch(U, B, verbose)
	return
}

func ComputeFirstCosystole(d1, d2 BinaryMatrix, verbose bool) (cosystole int) {
	if verbose {
		log.Printf("Computing first cohomology")
	}
	delta0 := d1.Transpose().Dense()
	delta1 := d2.Transpose().Dense()
	U, B, _, _, _, _ := UBDecomposition(delta1, delta0, verbose)
	U, B = U.Dense(), B.Dense()
	cosystole, _ = SystoleExhaustiveSearch(U, B, verbose)
	return
}

// The simplicial systole search algorithm is not guaranteed to find
// the global systole in all cases.  See thesis for details.  NB: The
// thesis describes "local" systoles where we restrict our attention
// to cycles incident to some vertex.  That is different in general to
// what is implemented here, which uses the ordinary exhaustive or
// randomized linear algebra search on subcomplexes for
// expediency. The two are thought to be equivalent in the case of
// Cayley complexes.
type SimplicialSystoleSearch[T any] struct {
	C                     *ZComplex[T]
	RandomTrials          int
	StartFiltration       int
	StopAtMinDegree       int
	StopNonzero           bool
	LogTriangleDepthsOnly bool
	Verbose               bool
}

func NewSimplicialSystoleSearch[T any](C *ZComplex[T], startFiltration int, randomTrials int, stopAtMinDegree int, stopNonzero bool, logTriangleDepthsOnly bool, verbose bool) *SimplicialSystoleSearch[T] {
	return &SimplicialSystoleSearch[T]{
		C:                     C,
		RandomTrials:          randomTrials,
		StartFiltration:       startFiltration,
		StopAtMinDegree:       stopAtMinDegree,
		StopNonzero:           stopNonzero,
		LogTriangleDepthsOnly: logTriangleDepthsOnly,
		Verbose:               verbose,
	}
}

func (S *SimplicialSystoleSearch[T]) Search() int {
	minWeight := 0
	for i, v := range S.C.VertexBasis() {
		w := S.SearchAtVertex(v)
		if w > 0 && (w < minWeight || minWeight == 0) {
			minWeight = w
		}
		if S.Verbose {
			log.Printf("finished simplicial search at vertex %d; min weight: %d", i, minWeight)
		}
	}
	return minWeight
}

func minVertexDegreeInEdgeVector[T any](C *ZComplex[T], edgeVector BinaryVector) int {
	vertexSet := make(map[int]bool)
	edgeBasis := C.EdgeBasis()
	vertexIndex := C.VertexIndex()

	for i := 0; i < edgeVector.Length(); i++ {
		if edgeVector.Get(i) == 1 {
			edge := edgeBasis[i]
			v0Index := vertexIndex[edge[0]]
			v1Index := vertexIndex[edge[1]]
			vertexSet[v0Index] = true
			vertexSet[v1Index] = true
		}
	}

	minDegree := math.MaxInt
	for vIndex := range vertexSet {
		degree := C.Degree(vIndex)
		if degree < minDegree {
			minDegree = degree
		}
	}

	if minDegree == math.MaxInt {
		return 0
	}
	return minDegree
}

// xxx potential optimization? reuse/extend UB from one filtration step to the next
func (S *SimplicialSystoleSearch[T]) SearchAtVertex(v ZVertex[T]) int {
	if S.Verbose {
		log.Printf("Complex: %s", S.C)
		log.Printf("Starting simplicial search at vertex v=%v step=%d", v, S.StartFiltration)
	}
	minWeight := 0
	prevTriangleDist := 0
	S.C.TriangularDepthFiltration(v, func(triangleIndex int, distanceMap map[int]int, subcomplex *ZComplex[T]) (stop bool) {
		if triangleIndex < S.StartFiltration {
			return false
		}
		if S.Verbose {
			//log.Printf("checking subcomplex of triangle depth filtration step %d", triangleIndex)
			//log.Printf("subcomplex: %s", subcomplex.MaximalSimplicesString())
		}
		if triangleIndex < len(subcomplex.TriangleBasis()) {
			t := subcomplex.TriangleBasis()[triangleIndex]
			vind := subcomplex.VertexIndex()
			v0 := vind[t[0]]
			v1 := vind[t[1]]
			v2 := vind[t[2]]
			d0 := distanceMap[v0]
			d1 := distanceMap[v1]
			d2 := distanceMap[v2]
			mindist := d0
			if d1 < mindist {
				mindist = d1
			}
			if d2 < mindist {
				mindist = d2
			}
			if mindist > prevTriangleDist {
				if S.Verbose {
					log.Printf("step=%d triangle distance increases to %d", triangleIndex, mindist)
				}
				prevTriangleDist = mindist
			}
			if S.LogTriangleDepthsOnly {
				return false
			}
		}
		ubVerbose := false
		U, B, _, dimZ1, dimB1, dimH1 := UBDecomposition(subcomplex.D1(), subcomplex.D2(), ubVerbose)
		if S.Verbose && triangleIndex < len(subcomplex.TriangleBasis()) {
			t := subcomplex.TriangleBasis()[triangleIndex]
			vind := subcomplex.VertexIndex()
			v0 := vind[t[0]]
			v1 := vind[t[1]]
			v2 := vind[t[2]]
			d0 := distanceMap[v0]
			d1 := distanceMap[v1]
			d2 := distanceMap[v2]
			log.Printf("step=%d vertices=[%d %d %d] distances=[%d %d %d] %s dimZ1=%d dimB1=%d dimH1=%d",
				triangleIndex, v0, v1, v2, d0, d1, d2, subcomplex, dimZ1, dimB1, dimH1)
		}
		U, B = U.Dense(), B.Dense()
		var localSystole int
		var localVector BinaryVector
		if S.RandomTrials > 0 {
			localSystole, localVector = SystoleRandomSearch(U, B, S.RandomTrials, S.Verbose)
		} else {
			localSystole, localVector = SystoleExhaustiveSearch(U, B, S.Verbose)
		}
		if localSystole > 0 && (localSystole < minWeight || minWeight == 0) {
			minWeight = localSystole
			if S.StopNonzero {
				if S.Verbose {
					log.Printf("stopping at triangle step %d", triangleIndex)
				}
				return true
			}
		}
		if !localVector.IsZero() {
			minDegree := minVertexDegreeInEdgeVector(subcomplex, localVector)
			if S.Verbose {
				log.Printf("step=%d systole=%d minDegree=%d", triangleIndex, localSystole, minDegree)
			}
			if S.StopAtMinDegree > 0 && minDegree >= S.StopAtMinDegree {
				if S.Verbose {
					log.Printf("stopping at triangle step %d (minDegree=%d >= %d)", triangleIndex, minDegree, S.StopAtMinDegree)
				}
				return true
			}
			if S.Verbose && triangleIndex < len(subcomplex.TriangleBasis()) {
				logIntersection(subcomplex, triangleIndex, localVector)
			}
		}
		return false
	})
	return minWeight
}

func EdgeVectorSupports[T any] (subcomplex *ZComplex[T], edgeVector BinaryVector) (vertexSupport map[int]bool, edgeSupport map[int]bool) {
	vertexSupport = make(map[int]bool)
	edgeSupport = make(map[int]bool)
	edgeBasis := subcomplex.EdgeBasis()
	vertexIndex := subcomplex.VertexIndex()
	for i := 0; i < edgeVector.Length(); i++ {
		if edgeVector.Get(i) == 1 {
			edgeSupport[i] = true
			edge := edgeBasis[i]
			v0 := vertexIndex[edge[0]]
			v1 := vertexIndex[edge[1]]
			vertexSupport[v0] = true
			vertexSupport[v1] = true
		}
	}
	return vertexSupport, edgeSupport
}

func TriangleSupports[T any](subcomplex *ZComplex[T], triangleIndex int) (vertexSupport map[int]bool, edgeSupport map[int]bool) {
	vertexSupport = make(map[int]bool)
	edgeSupport = make(map[int]bool)
	t := subcomplex.TriangleBasis()[triangleIndex]
	edges := t.Edges()
	edgeIndex := subcomplex.EdgeIndex()
	vertexIndex := subcomplex.VertexIndex()

	v0 := vertexIndex[t[0]]
	v1 := vertexIndex[t[1]]
	v2 := vertexIndex[t[2]]
	vertexSupport[v0] = true
	vertexSupport[v1] = true
	vertexSupport[v2] = true

	for _, edge := range edges {
		n := edgeIndex[edge]
		edgeSupport[n] = true
	}
	return vertexSupport, edgeSupport
}

func logIntersection[T any](subcomplex *ZComplex[T], triangleIndex int, edgeVector BinaryVector) {
	triangleVertexSupport, triangleEdgeSupport := TriangleSupports(subcomplex, triangleIndex)

	edgeVecVertexSupport, edgeVecEdgeSupport := EdgeVectorSupports(subcomplex, edgeVector)

	edgeIntersection := supportIntersection(edgeVecEdgeSupport, triangleEdgeSupport)
	edgeKeys := make([]int, 0, len(edgeIntersection))
	for k := range edgeIntersection {
		edgeKeys = append(edgeKeys, k)
	}
	log.Printf("local vector edge intersection with triangle t: %v", edgeKeys)

	vertexIntersection := supportIntersection(edgeVecVertexSupport, triangleVertexSupport)
	vertexKeys := make([]int, 0, len(vertexIntersection))
	for k := range vertexIntersection {
		vertexKeys = append(vertexKeys, k)
	}
	log.Printf("local vector vertex intersection with triangle t: %v", vertexKeys)
}

func supportIntersection(a, b map[int]bool) (result map[int]bool) {
	result = make(map[int]bool)
	for k := range a {
		if _, ok := b[k]; ok {
			result[k] = true
		}
	}
	return result
}
