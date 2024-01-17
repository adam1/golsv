package golsv

import (
	"fmt"
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
	reducer.Reduce(B)
	// xxx temp hack until the colOps -> kernelBasis API is fixed
	reducer.(*DiagonalReducer).computeKernelBasis()
	kernelMatrix = reducer.(*DiagonalReducer).kernelBasis
	// xxx get colOps, convert to kernelMatrix
	if showSteps {
		fmt.Printf("kernelMatrix: %v", kernelMatrix)
	}
	dreducer := reducer.(*DiagonalReducer) 
	coimageMatrix := dreducer.CoimageBasis()

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
		name string
		A    BinaryMatrix
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
		R.Reduce(A)
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
			D1, _, _ := R1.Reduce(B1)
			// log.Printf("xxx D1: %s\n%s", D1, dumpMatrix(D1))

			// log.Printf("xxx ==================== Reducer 2:\n")
			R2 := NewDiagonalReducer(verbose)
			R2.switchToDensePredicate = func(remaining int, subdensity float64) bool {
				return remaining <= cols / 2
			}
			R2.statIntervalSteps = 1
			B2 := A.Copy()
			D2, _, _ := R2.Reduce(B2)
			// log.Printf("xxx D2: %s\n%s", D2, dumpMatrix(D2))

			if !D1.Equal(D2) {
				t.Errorf("D1 != D2")
			}
		})
	}
}
