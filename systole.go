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
				if verbose {
					log.Printf("new min weight: %d", minWeight)
					log.Printf("xxx c: %s", sum.ColumnVector(0).SupportString())
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
	if verbose {
		log.Printf("Computing first homology")
	}
	U, B, _ := UBDecomposition(d1, d2, verbose)
	U, B = U.Dense(), B.Dense()
	return SystoleExhaustiveSearch(U, B, verbose)
}

func ComputeFirstCosystole(d1, d2 BinaryMatrix, verbose bool) (cosystole int) {
	if verbose {
		log.Printf("Computing first cohomology")
	}
	delta0 := d1.Transpose().Dense()
	delta1 := d2.Transpose().Dense()
	U, B, _ := UBDecomposition(delta1, delta0, verbose)
	U, B = U.Dense(), B.Dense()
	return SystoleExhaustiveSearch(U, B, verbose)
}
