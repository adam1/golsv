package golsv

import (
	"log"
	"math"
	"time"
)

// xxx optimization 8/1 
//
//   - precompute matrix of row operations, store in sparse form, load and convert to dense.
//     - pass in Dense matrix to aligner
//
//   - use optimized Dense * Sparse multiplication
//
//   - (maybe) add special SubMultiplyRight to Dense, works same as
//     the optimized MultiplyRight but only uses part of each column
//     slice.
//
//     - then in the aligner use SubMultiplyRight to obtain the tail
//       of vector w, and only continue to compute the rest of w if
//       that tail is nonzero (meaning the vector is not in the image
//       the cumulative B to that point).
//   

type Aligner struct {
	B1smith *Sparse
	P *DenseBinaryMatrix
	B1colops []Operation
	Z1 *Sparse
	U BinaryMatrix // output matrix
	minWeight int
}

func NewAligner(B1smith *Sparse, P *DenseBinaryMatrix, B1colops []Operation,
	Z1 *Sparse) *Aligner {
	return &Aligner{
		B1smith: B1smith,
		P: P,
		B1colops: B1colops,
		Z1: Z1,
		U: NewSparseBinaryMatrix(B1smith.NumRows(), 0),
		minWeight: math.MaxInt,
	}
}

// Align computes a matrix U such that the columns of U, together with
// the columns of Bsmith, form a basis for Z.
func (a *Aligner) Align() BinaryMatrix {
	BIsSmith, Brank := a.B1smith.IsSmithNormalForm()
	if !BIsSmith {
		log.Fatal("matrix B1smith is not in Smith normal form")
	}
	log.Printf("B1smith = %v, Brank = %d", a.B1smith, Brank)

	// nb. unexpectedly B1colops keeps coming out empty. xxx why?  for
	// now assume it is true.
	if len(a.B1colops) > 0 {
		panic("B1colops not empty")
	}
	// Algorithm:
	//
	// let B = B1smith
	// let Z = Z1
	// let d = rank of B
	//
	// for each column vector v of Z:
	//    - compute w = P(v) (P is the row operations matrix)
	//    - let k = max index of support of w
	//    - if k > d, then
	//      - v is not in the span of B
	//      - append column w to B
	//      - reduce new column of B;
	//        this involves row swapping to create the pivot in
	//        position d and adding columns to clear column d
	//      - apply the same row operations to P
	//      - increment d
	//      - append column v to U
	//    - else
	//      - discard v since it is in the span of B
	//
	// in our main example, this should produce 19 column vectors in U,
	// since we already know that dim H_1 = 19.

 	log.Printf("beginning alignment")
	Z := a.Z1
	rows := Z.NumRows()

	statInterval := 1
	timeStart := time.Now()
	timeLastReport := timeStart
	doneLastReport := 0

	for j := 0; j < Z.NumColumns(); j++ {
		v := Z.Submatrix(0, rows, j, j+1)
		w := a.P.MultiplyRight(v)
		k := w.(*DenseBinaryMatrix).MaxColumnSupport(0)
		if k >= Brank {
			a.handleIndependentVector(v, w, Brank, k)
			Brank++
		}
		doneLastReport++
		if doneLastReport >= statInterval {
			now := time.Now()
			reportElapsed := now.Sub(timeLastReport)
			totalElapsed := now.Sub(timeStart)
			cRate := float64(doneLastReport) / reportElapsed.Seconds()
			tRate := float64(j) / totalElapsed.Seconds()
			log.Printf("align: processed %d/%d (%.2f%%) cols; crate=%1.0f trate=%1.0f found=%d",
				j, Z.NumColumns(), 100.0 * float64(j) / float64(Z.NumColumns()),
				cRate, tRate, a.U.NumColumns())
			timeLastReport = now
			doneLastReport = 0
			if reportElapsed.Seconds() < 1 {
				statInterval = statInterval * 2
			} else if reportElapsed.Seconds() > 10 {
				statInterval = 1 + statInterval / 2
			}
		}
	}
	log.Printf("done; min weight is %d", a.minWeight)
	return a.U
}

func (a *Aligner) handleIndependentVector(v BinaryMatrix, w BinaryMatrix, Brank int, k int) {
	a.U.(*Sparse).AppendColumn(v)
	log.Printf("found independent vector; found=%d total", a.U.NumColumns())
	// sneak peak at systole
	weight := v.(*Sparse).ColumnWeight(0)
	if weight < a.minWeight {
		a.minWeight = weight
		log.Printf("new min weight %d", weight)
	}
	a.B1smith.AppendColumn(w)
	a.makePivot(Brank, k)
	log.Printf("clearing column")
	a.clearColumn(Brank)
	log.Printf("done clearing column")
}

func (a *Aligner) makePivot(Brank int, k int) {
	// we're increasing the rank of B by 1, so we want a pivot in
	// position Brank.
	if k > Brank {
		a.B1smith.SwapRows(Brank, k)
		a.P.SwapRows(Brank, k)
	}
}

// xxx needs test!
func (a *Aligner) clearColumn(Brank int) {
	col := Brank
	row := Brank + 1
	ops := 0
	reportInterval := 1000
	for {
		row = a.B1smith.ScanDown(row, col)
		if row == -1 {
			break
		}
		a.B1smith.Set(row, col, 0)
		a.P.AddRow(col, row)
		row++
		ops++
		if ops % reportInterval == 0 {
			log.Printf("did %d row ops", ops)
		}
	}
	log.Printf("did %d row ops", ops)
}
