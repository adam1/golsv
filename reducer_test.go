package golsv

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"testing"
)

func doKernelBasisTest(t *testing.T, A BinaryMatrix, numSamples int) (kernelMatrix BinaryMatrix) {
	//  A = input matrix
	//  B = reduced matrix
	showSteps := false
	if showSteps {
		fmt.Printf("A:\n%v\n", A)
	}
	B := A.Copy()
	var reducer Reducer
	verbose := showSteps
	reducer = NewDiagonalReducer(verbose)
	colOpWriter := NewOperationSliceWriter()
	D := reducer.Reduce(B, NewOperationSliceWriter(), colOpWriter)
	DSparse := D.Sparse()
	isSmith, rank := DSparse.IsSmithNormalForm()
	if !isSmith {
		t.Errorf("Matrix is not in Smith normal form")
	}
	// xxx temp hack until the colOps -> kernelBasis API is fixed
	kernelMatrix = reducer.(*DiagonalReducer).computeKernelBasis(colOpWriter.Slice())
	if showSteps {
		fmt.Printf("kernelMatrix: %v", kernelMatrix)
	}
	dreducer := reducer.(*DiagonalReducer)
	coimageMatrix := dreducer.CoimageBasis()
	if rank != coimageMatrix.NumColumns() {
		t.Errorf("Rank of reduced matrix does not match length of coimage basis")
	}
	// the matrix should produce zero for vectors in the kernel
	zeroImageVector := ZeroVector(A.NumRows())
	var inputs []BinaryVector
	secureRandom := false
	if numSamples > 0 {
		inputs = SampleBinaryVectorSpaceList(kernelMatrix, numSamples, secureRandom)
	} else {
		inputs = EnumerateBinaryVectorSpaceList(kernelMatrix)
	}
	for _, v := range inputs {
		result := A.MultiplyRight(v.Matrix()).AsColumnVector()
		if !reflect.DeepEqual(result, zeroImageVector) {
			t.Errorf("Error checking Av = 0; A=\n%vv=%v\nresult=\n%v\nwant zero", A, v, result)
		}
	}
	// the matrix should produces nonzero for vectors in the coimage
	if numSamples > 0 {
		inputs = SampleBinaryVectorSpaceList(coimageMatrix, numSamples, secureRandom)
	} else {
		inputs = EnumerateBinaryVectorSpaceList(coimageMatrix)
	}
	for _, v := range inputs {
		if v.IsZero() {
			continue
		}
		result := A.MultiplyRight(v.Matrix()).AsColumnVector()
		if reflect.DeepEqual(result, zeroImageVector) {
			t.Errorf("Error checking Av != 0; A=\n%vv=%v\nresult=\n%v", A, v, result)
		}
	}
	return kernelMatrix
}

func TestReducer_KernelBasis(t *testing.T) {
	tests := []struct {
		name      string
		A         BinaryMatrix
		dimKernel int
	}{
		{
			name: "Test 1",
			A: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 0},
			}),
			dimKernel: 1,
		},
		{
			name: "Test 2",
			A: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 0},
				{1, 1, 1},
			}),
			dimKernel: 1,
		},
		{
			name: "Test 3",
			A: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 0},
				{1, 1, 1},
				{1, 1, 1},
			}),
			dimKernel: 1,
		},
		{
			name: "Test 4",
			A: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 0},
				{1, 1, 1},
				{1, 1, 1},
				{1, 1, 1},
			}),
			dimKernel: 1,
		},
		{
			name: "Test 5",
			A: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{0, 1, 1, 0},
				{0, 1, 1, 0},
				{0, 1, 1, 1},
				{0, 0, 0, 0},
			}),
			dimKernel: 2,
		},
		{
			name: "Test 6",
			A: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1},
			}),
			dimKernel: 0,
		},
		{
			name: "Test 7",
			A: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{0},
			}),
			dimKernel: 1,
		},
		{
			name: "Test 8",
			A: NewSparseBinaryMatrixFromString(
				`1 1 1 1 0 0 1 
0 0 0 0 0 1 1
1 0 1 0 0 0 0
1 0 1 1 1 0 0
0 0 0 1 0 1 0
0 1 0 1 0 1 0
0 1 0 1 0 1 0`),
			dimKernel: 2,
		},
		{
			name: "Test 9",
			A: NewSparseBinaryMatrixFromString(
				`1 0 0 0 0 0 1 0 1 0 0 0 1 0 0 0 0
 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0
 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 1
 0 0 1 0 1 0 0 0 0 0 0 0 0 0 0 0 1
 1 0 0 0 0 0 1 0 1 0 0 0 1 0 1 0 1
 1 0 1 0 1 0 0 0 0 0 1 0 1 0 1 0 1
 0 0 1 0 1 0 1 0 1 0 1 0 0 0 1 0 0
 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0
 0 0 1 0 0 0 0 0 0 0 1 0 0 0 1 0 0
 1 0 0 0 1 0 0 0 0 0 1 0 1 0 0 0 1`),
			dimKernel: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kernelMatrix := doKernelBasisTest(t, tt.A, 0)
			gotDimKernel := kernelMatrix.NumColumns()
			if gotDimKernel != tt.dimKernel {
				t.Errorf("dim(kernel) = %d, want %d", gotDimKernel, tt.dimKernel)
			}
		})
	}
}

func TestReducer_RandomDenseMatrixKernelBasis(t *testing.T) {
	trials := 100
	rowsMax := 10
	colsMax := 10
	numSamples := 1000
	for n := 0; n < trials; n++ {
		rows := rand.Intn(rowsMax) + 1
		cols := rand.Intn(colsMax) + 1
		t.Run(fmt.Sprintf("trial %d: %d x %d", n, rows, cols), func(t *testing.T) {
			A := NewRandomDenseBinaryMatrix(rows, cols)
			doKernelBasisTest(t, A, numSamples)
		})
	}
}

func TestReducer_RandomSparseMatrixKernelBasis(t *testing.T) {
	trials := 10
	rowsMax := 100
	colsMax := 100
	numSamples := 1000
	secureRandom := false
	for n := 0; n < trials; n++ {
		rows := rand.Intn(rowsMax) + 1
		cols := rand.Intn(colsMax) + 1
		t.Run(fmt.Sprintf("trial %d: %d x %d", n, rows, cols), func(t *testing.T) {
			A := NewRandomSparseBinaryMatrix(rows, cols, 0.01, secureRandom)
			doKernelBasisTest(t, A, numSamples)
		})
	}
}

const benchDim = 9000

func BenchmarkReducer_Dense(b *testing.B) {
	benchmarkReducer(b, NewRandomDenseBinaryMatrixWithDensity(benchDim, benchDim, 0.001))
}

func BenchmarkReducer_Sparse(b *testing.B) {
	secureRandom := false
	benchmarkReducer(b, NewRandomSparseBinaryMatrix(benchDim, benchDim, 0.001, secureRandom))
}

func benchmarkReducer(b *testing.B, A BinaryMatrix) {
	for n := 0; n < b.N; n++ {
		R := NewDiagonalReducer(false)
		R.Reduce(A, NewOperationSliceWriter(), NewOperationSliceWriter())
	}
}

func TestReducer_SwitchToDense(t *testing.T) {
	// do a reduction with and without switching to dense, verify
	// identical results.
	verbose := false
	trials := 10
	maxSize := 100
	secureRandom := false
	for n := 0; n < trials; n++ {
		rows := rand.Intn(maxSize) + 1
		cols := rand.Intn(maxSize) + 1
		t.Run(fmt.Sprintf("trial %d: %d x %d", n, rows, cols), func(t *testing.T) {
			A := NewRandomSparseBinaryMatrix(rows, cols, 0.2, secureRandom)
			// log.Printf("xxx A: %s\n%s", A, dumpMatrix(A))

			// log.Printf("xxx ==================== Reducer 1:\n")
			R1 := NewDiagonalReducer(verbose)
			B1 := A.Copy()
			D1 := R1.Reduce(B1, NewOperationSliceWriter(), NewOperationSliceWriter())
			// log.Printf("xxx D1: %s\n%s", D1, dumpMatrix(D1))

			// log.Printf("xxx ==================== Reducer 2:\n")
			R2 := NewDiagonalReducer(verbose)
			R2.switchToDensePredicate = func(remaining int, subdensity float64) bool {
				return remaining <= cols/2
			}
			R2.statIntervalSteps = 1
			B2 := A.Copy()
			D2 := R2.Reduce(B2, NewOperationSliceWriter(), NewOperationSliceWriter())
			// log.Printf("xxx D2: %s\n%s", D2, dumpMatrix(D2))

			if !D1.Equal(D2) {
				t.Errorf("D1 != D2")
			}
		})
	}
}

func TestReducerCheckOps(t *testing.T) {
	trials := 10
	minSize := 10
	maxSize := 40
	verbose := false
	for n := 0; n < trials; n++ {
		var m int
		if minSize < maxSize {
			m = minSize + rand.Intn(maxSize-minSize) + 1
		} else {
			m = minSize
		}
		M := NewRandomDenseBinaryMatrix(m, m)
		R := NewDiagonalReducer(verbose)
		rowOpWriter, colOpWriter := NewOperationSliceWriter(), NewOperationSliceWriter()
		D := R.Reduce(M.Copy(), rowOpWriter, colOpWriter)
		// apply rowOps and colOps to M, verify that it equals D
		for _, op := range rowOpWriter.Slice() {
			M.ApplyRowOperation(op)
		}
		for _, op := range colOpWriter.Slice() {
			M.ApplyColumnOperation(op)
		}
		if !M.Equal(D) {
			t.Errorf("M != D")
		}
	}
}

func TestReducerCheckOpsMatrices(t *testing.T) {
	trials := 10
	minSize := 10
	maxSize := 10
	verbose := false
	for n := 0; n < trials; n++ {
		var m int
		if minSize < maxSize {
			m = minSize + rand.Intn(maxSize-minSize) + 1
		} else {
			m = minSize
		}
		M := NewRandomDenseBinaryMatrix(m, m)
		R := NewDiagonalReducer(verbose)
		rowOpWriter, colOpWriter := NewOperationSliceWriter(), NewOperationSliceWriter()
		D := R.Reduce(M.Copy(), rowOpWriter, colOpWriter)
		// convert rowOps and colOps to matrices, apply to M, verify
		// that it equals D
		rowOpsMat := RowOperationsMatrix(rowOpWriter.Slice(), M.NumRows())
		colOpsMat := ColumnOperationsMatrix(colOpWriter.Slice(), M.NumColumns())
		P := rowOpsMat.MultiplyRight(M).MultiplyRight(colOpsMat)
		if !P.Equal(D) {
			t.Errorf("P != D")
		}
	}
}

func TestReducerCheckOpsMatricesSpecific(t *testing.T) {
	tests := []struct {
		M BinaryMatrix
	}{
		{NewDenseBinaryMatrixFromString(
			`0 0 0 1 0 1 0 1 0 0 0 0 1 0 1 0 1 0 0
1 0 0 1 0 0 0 0 1 0 0 0 0 1 0 1 1 1 0
0 0 0 0 0 0 1 0 0 1 1 1 0 1 0 1 1 0 0
0 1 1 0 1 0 0 0 0 0 1 1 0 0 1 1 1 1 0
0 0 0 1 1 1 0 0 1 1 1 1 1 1 1 1 1 1 0
1 1 0 1 0 1 1 1 1 1 0 0 1 0 0 0 1 0 0
0 0 0 0 1 0 0 0 0 0 0 1 1 0 1 1 1 0 0
0 1 0 1 1 0 0 0 0 1 0 1 1 0 0 1 1 0 1
0 1 0 0 0 0 0 1 1 1 0 0 1 1 0 1 1 1 1
0 0 0 0 1 1 0 0 0 0 1 0 0 1 0 1 1 0 1
0 1 0 0 0 0 1 1 0 1 1 0 0 1 0 0 0 1 1
1 1 1 1 1 1 0 1 0 0 1 0 0 1 0 1 0 0 0
1 0 0 1 0 0 0 1 0 1 0 1 0 0 0 1 0 1 0
0 0 1 0 1 0 1 0 0 1 0 1 1 1 1 0 0 0 0
1 0 0 0 1 0 0 0 1 1 1 1 0 0 0 0 1 0 0
1 1 0 0 0 1 0 1 0 1 1 0 1 1 1 0 1 1 0
1 0 1 1 0 1 1 0 1 1 1 1 1 0 1 1 1 1 1
1 0 1 0 0 1 1 1 0 1 1 1 0 1 0 0 1 1 1
0 0 1 1 0 0 0 0 1 0 1 0 1 0 0 0 0 1 0`),
		},
	}
	for _, test := range tests {
		R := NewDiagonalReducer(false)
		rowOpWriter, colOpWriter := NewOperationSliceWriter(), NewOperationSliceWriter()
		D := R.Reduce(test.M.Copy(), rowOpWriter, colOpWriter)
		// convert rowOps and colOps to matrices, apply to M, verify
		// that it equals D
		rowOpsMat := RowOperationsMatrix(rowOpWriter.Slice(), test.M.NumRows())
		colOpsMat := ColumnOperationsMatrix(colOpWriter.Slice(), test.M.NumColumns())
		P := rowOpsMat.MultiplyRight(test.M).MultiplyRight(colOpsMat)
		if !P.Equal(D) {
			t.Errorf("P != D")
		}
	}
}

func TestReducerInvertRandom(t *testing.T) {
	trials := 10
	minSize := 10
	maxSize := 10
	notInv := 0
	good := 0
	for n := 0; n < trials; n++ {
		var m int
		if minSize < maxSize {
			m = minSize + rand.Intn(maxSize-minSize) + 1
		} else {
			m = minSize
		}
		M := NewRandomDenseBinaryMatrix(m, m)
		MInv, ok := attemptInvert(t, M)
		if !ok {
			notInv++
			continue
		}
		P := M.MultiplyRight(MInv)
		I := NewDenseBinaryMatrixIdentity(m)
		if P.Equal(I) {
			good++
		} else {
			t.Errorf("M * MInv != I")
		}
	}
}

func attemptInvert(t *testing.T, M BinaryMatrix) (MInv BinaryMatrix, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			if r.(string) == "not invertible" {
				ok = false
			} else {
				t.Errorf("panic: %v", r)
			}
		}
	}()
	reducer := NewDiagonalReducer(false)
	MInv = reducer.Invert(M.Copy())
	return MInv, true
}

func TestImageBasis(t *testing.T) {
	tests := []struct {
		M    BinaryMatrix
		want BinaryMatrix
	}{
		{
			NewDenseBinaryMatrixFromString(
				`1 0 0
			     0 1 0
			     0 0 1`),
			NewDenseBinaryMatrixFromString(
				`1 0 0
				 0 1 0
				 0 0 1`),
		},
		{
			NewDenseBinaryMatrixFromString(
				`1 0 1
			     0 1 1
			     0 0 0`),
			NewDenseBinaryMatrixFromString(
				`1 0
				 0 1
				 0 0`),
		},
		{
			NewDenseBinaryMatrixFromString(
				`0 0 1
			     0 1 1
			     0 0 0`),
			NewDenseBinaryMatrixFromString(
				`0 1
				 1 0
				 0 0`),
		},
	}
	for _, test := range tests {
		verbose := false
		got := ImageBasis(test.M, verbose)
		if !got.Equal(test.want) {
			t.Errorf("got %v:\n%s\nwant %v:\n%s", got, dumpMatrix(got), test.want, dumpMatrix(test.want))
		}
	}
}

func TestShor9qubit(t *testing.T) {
	d_1 := NewDenseBinaryMatrixFromString(
		`1 1 0 0 0 0 0 0 0
		 0 1 1 0 0 0 0 0 0
		 0 0 0 1 1 0 0 0 0
		 0 0 0 0 1 1 0 0 0
		 0 0 0 0 0 0 1 1 0
		 0 0 0 0 0 0 0 1 1`)
	d_2 := NewDenseBinaryMatrixFromString(
		`1 0
		 1 0
		 1 0
	 	 1 1
		 1 1
	 	 1 1
  		 0 1
		 0 1
		 0 1`)

	P := d_1.MultiplyRight(d_2)
	if !P.IsZero() {
		t.Errorf("d_1 * d_2 != 0")
	}
	verbose := false
	U_1, B_1, Z_1, _, _, _ := UBDecomposition(d_1, d_2, verbose)
	dimH_1 := U_1.NumColumns()
	if dimH_1 != 1 {
		t.Errorf("dim(H_1) = %d, want 1", dimH_1)
	}
	if verbose {
		log.Printf("Z_1: %s\n%s", Z_1, DumpMatrix(Z_1))
		log.Printf("B_1: %s\n%s", B_1, DumpMatrix(B_1))
	}
	S_1 := SystoleExhaustiveSearch(U_1, B_1, verbose)
	if S_1 != 3 {
		t.Errorf("S_1 = %d, want 3", S_1)
	}
	Su1, _, _, _ := ComputeFirstSystole(d_1, d_2, verbose)
	if Su1 != 3 {
		t.Errorf("Su1 = %d, want 3", Su1)
	}

	// 	// xxx experimental; considering the effect of an automorphism
	// 	f := NewDenseBinaryMatrixFromString(
	// 		`1 0 0 0 0 0 0 0 0
	// 		 1 1 0 0 0 0 0 0 0
	// 		 0 0 1 0 0 0 0 0 0
	// 		 0 0 0 1 0 0 0 0 0
	// 		 0 0 0 0 1 0 0 0 0
	// 		 0 0 0 0 0 1 0 0 0
	// 		 0 0 0 0 0 0 1 0 0
	// 		 0 0 0 0 0 0 0 1 0
	// 		 0 0 0 0 0 0 0 0 1`)
	// 	fZ_1 := f.MultiplyRight(Z_1)
	// 	log.Printf("fZ_1: %s\n%s", fZ_1, DumpMatrix(fZ_1))

	// 	fB_1 := f.MultiplyRight(B_1)
	// 	log.Printf("fB_1: %s\n%s", fB_1, DumpMatrix(fB_1))

	// 	d_1Uy := f.MultiplyRight(d_1).MultiplyRight(f) // note: f = f^-1
	// 	d_2Uy := f.MultiplyRight(d_2).MultiplyRight(f)

	// 	log.Printf("d_1Uy: %s\n%s", d_1Uy, DumpMatrix(d_1Uy))
	// 	log.Printf("d_2Uy: %s\n%s", d_2Uy, DumpMatrix(d_2Uy))

}
