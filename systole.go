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

func SystoleRandomSearch(U, B BinaryMatrix, trials int, verbose bool) (minWeight int) {
	if U.NumColumns() == 0 {
		return 0
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
			if verbose {
				log.Printf("new min weight: %d", minWeight)
			}
		}
		if n > 0 && n % reportInterval == 0 {
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
				reportInterval *= 2
			}
		}
	}
	return minWeight
}

func SystoleExhaustiveSearch(U, B BinaryMatrix, verbose bool) (minWeight int) {
	if U.NumColumns() == 0 {
		return 0
	}
	minWeight = math.MaxInt
	EnumerateBinaryVectorSpace(U, func(a BinaryMatrix) bool {
		if a.IsZero() {
			return true
		}
		EnumerateBinaryVectorSpace(B, func(b BinaryMatrix) bool {
			sum := a.Copy().Dense()
			sum.AddMatrix(b)
			weight := sum.ColumnWeight(0)
			if weight < minWeight {
				minWeight = weight
				if verbose {
					log.Printf("new min weight: %d", minWeight)
				}
			}
			return true
		})
		return true
	})
	return minWeight
}

// ComputeFirstSystole computes the degree one systole of the complex.
// All processing is done in memory, hence the function is only
// suitable for small complexes.  For larger complexes, use the
// procedure represented in worksets/Makefile, which uses the
// individual programs and intermediate files at each step.
// ComputeFirstSystole is useful for small complexes and for testing.
func ComputeFirstSystole(d1, d2 BinaryMatrix, verbose bool) int {
	log.Printf("Computing first homology")
	U, B := UBDecomposition(d1, d2, verbose)
	U, B = U.Dense(), B.Dense()
	return SystoleExhaustiveSearch(U, B, verbose)
}

func ComputeFirstCosystole(d1, d2 BinaryMatrix, verbose bool) int {
	log.Printf("Computing first cohomology")
	delta0 := d1.Transpose().Dense()
	delta1 := d2.Transpose().Dense()
	U, B := UBDecomposition(delta1, delta0, verbose)
	U, B = U.Dense(), B.Dense()
	return SystoleExhaustiveSearch(U, B, verbose)
}

func UBDecomposition(d1, d2 BinaryMatrix, verbose bool) (U, B BinaryMatrix) {
	// here we adapt the recipes from worskets/Makefile to be run
	dimC0 := d1.NumRows()
	dimC1 := d1.NumColumns()
	dimC2 := d2.NumColumns()

	_, _, d1colops, d1rank := smithNormalForm(d1)
	dimZ1 := d1.NumColumns() - d1rank

	_, _, d2colops, d2rank := smithNormalForm(d2)
	dimB1 := d2rank
	dimH1 := dimZ1 - dimB1
	if verbose {
		log.Printf(
`
C_2 ------------> C_1 ------------> C_0
dim(C_2)=%-8d dim(C_1)=%-8d dim(C_0)=%-8d
                  dim(Z_1)=%-8d
                  dim(B_1)=%-8d
                  dim(H_1)=%-8d
`, dimC2, dimC1, dimC0, dimZ1, dimB1, dimH1)
	}


	Z1 := automorphism(d1colops, dimC1, d1rank, dimC1, verbose)

	d2coimage := automorphism(d2colops, dimC2, 0, d2rank, verbose)
	B1 := d2.MultiplyRight(d2coimage)

	B1smith, B1rowops, B1colops, _ := smithNormalForm(B1)

	PT := automorphism(B1rowops, dimC1, 0, dimC1, verbose)
	P := PT.Transpose()

	U = align(B1smith.Sparse(), P.Dense(), B1colops, Z1.Sparse(), verbose)
	return U, B1
}

func smithNormalForm(M BinaryMatrix) (smith BinaryMatrix, rowops, colops []Operation, rank int) {
	verbose := false
	R := NewDiagonalReducer(verbose)
	D, rowOps, colOps := R.Reduce(M.Copy())
	isSmith, rank := D.Sparse().IsSmithNormalForm()
	if !isSmith {
		panic("smith normal form not reached")
	}
	return D, rowOps, colOps, rank
}

func automorphism(ops []Operation, dim, cropStart, cropEnd int, verbose bool) BinaryMatrix {
  	M := NewDenseBinaryMatrixIdentity(dim)
	reader := NewOperationSliceReader(ops)
	streamer := NewOpsFileMatrixStreamer(reader, M, verbose)
 	streamer.Stream()
	if cropStart > 0 || cropEnd < dim {
		if verbose {
			log.Printf("cropping to columns %d-%d", cropStart, cropEnd)
		}
		M = M.Submatrix(0, M.NumRows(), cropStart, cropEnd)
	}
	return M
}

func align(B1smith *Sparse, P *DenseBinaryMatrix, B1colops []Operation, Z1 *Sparse, verbose bool) *DenseBinaryMatrix {
	aligner := NewAligner(B1smith, P, B1colops, Z1, verbose)
	return aligner.Align().Dense()
}
