package golsv

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)	

type RandomComplexGenerator struct {
	dimC_0 int
	dimC_1 int
	dimC_2 int
	dimZ_1 int
	dimH_1 int
	dimB_1 int
	verbose bool
}

func NewRandomComplexGenerator(dimC_0 int, verbose bool) *RandomComplexGenerator {
	return &RandomComplexGenerator{
		dimC_0,
		-1,
		-1,
		-1,
		-1,
		-1,
		verbose,
	}
}

func (R *RandomComplexGenerator) RandomComplex() (d_1, d_2 BinaryMatrix, err error) {
	R.dimC_1, err = randomizeDimC_1(R.dimC_0)
	if err != nil {
		return nil, nil, err
	}
	R.dimC_2, err = randomizeDimC_2(R.dimC_0, R.dimC_1)
	if err != nil {
		return nil, nil, err
	}
	density := 0.1

	var kernelMatrix BinaryMatrix
	d_1, kernelMatrix = randomize_d_1(R.dimC_0, R.dimC_1, density)
	if R.verbose {
		log.Printf("generated d_1: %v\n", d_1)
	}

	R.dimZ_1 = kernelMatrix.NumColumns()
	R.dimB_1, err = randomizeDimB_1(R.dimZ_1)
	if err != nil {
		return nil, nil, err
	}
	R.dimH_1 = R.dimZ_1 - R.dimB_1

// 	log.Printf("xxx dimC_0=%d dimC_1=%d dimC_2=%d dimZ_1=%d dimB_1=%d dimH_1=%d",
// 		R.dimC_0, R.dimC_1, R.dimC_2, R.dimZ_1, R.dimB_1, R.dimH_1)

	// shuffle the columns of kernelMatrix and truncate to dimB_1
	// columns
	if R.verbose {
		log.Printf("shuffling columns of kernel matrix")
	}
	kernelMatrix.(*Sparse).ShuffleColumns()
	kernelMatrix = kernelMatrix.DenseSubmatrix(0, kernelMatrix.NumRows(), 0, R.dimB_1)

	if R.verbose {
		log.Printf("generating d_2")
	}
	d_2Sparse := NewSparseBinaryMatrix(R.dimC_1, 0)
	for i := 0; i < R.dimC_2; i++ {
		col := RandomLinearCombination(kernelMatrix)
		d_2Sparse.AppendColumn(col)
	}
	if R.verbose {
		log.Printf("generated d_2: %v\n", d_2Sparse)
	}
	return d_1, d_2Sparse, nil
}

func randomize_d_1(dimC_0 int, dimC_1 int, density float64) (d_1 BinaryMatrix, kernelMatrix BinaryMatrix) {
	for true {
		d_1 := NewRandomDenseBinaryMatrixWithDensity(dimC_0, dimC_1, density)
		// to generate d_2:
		//
		//   compute a basis for Z_1
		//
		//   choose dim B_1 randomly in range 0..dim Z_1,
		//
		//   for each column of d_2
		//
		//      choose a random linear combination of the columns of Z_1
		//
		verbose := false
		reducer := NewDiagonalReducer(verbose)
		// xxx the copy here prevents the original d_1 from being modified
		// which causes a bug for some reason...
		D, _, _ := reducer.Reduce(d_1.Copy())

		Dsparse := D.Sparse()
		smithNormal, _ := Dsparse.IsSmithNormalForm()
		if !smithNormal {
			panic(fmt.Errorf("D is not in Smith normal form"))
		}
		reducer.computeKernelBasis()
		kernelMatrix := reducer.kernelBasis
		dimZ_1 := kernelMatrix.NumColumns()
		if dimZ_1 > 0 {
			return d_1, kernelMatrix
		}
		if verbose {
			log.Printf("dimZ_1=0, trying again")
		}
	}
	return nil, nil
}

func randomizeDimC_1(dimC_0 int) (int, error) {
	min := dimC_0 / 2 // minimum number of edges in a connected graph
	if min < 1 {
		min = 1
	}
	max := dimC_0 * (dimC_0 - 1) / 2 // number of edges in a complete graph
	if max < 1 {
		max = 1
	}
	if max <= min {
		max = min + 1
	}
	d, err := rand.Int(rand.Reader, big.NewInt(int64(max - min)))
	if err != nil {
		return -1, err
	}
	return min + int(d.Int64()), nil
}

func randomizeDimC_2(dimC_0, dimC_1 int) (int, error) {
	min := dimC_1 / 3 // minimum number of triangles such that each edge could be in at least 1 triangle
	if min < 1 {
		min = 1
	}
	max := dimC_0 * (dimC_0 - 1) * (dimC_0 - 2) / 3 // number of triangles for a clique complex
	if max <= min {
		max = min + 1
	}
	d, err := rand.Int(rand.Reader, big.NewInt(int64(max - min)))
	if err != nil {
		return -1, err
	}
	return min + int(d.Int64()), nil
}

func randomizeDimB_1(dimZ_1 int) (int, error) {
	min := 0
	max := dimZ_1
	if max <= min {
		max = min + 1
	}
	d, err := rand.Int(rand.Reader, big.NewInt(int64(max - min)))
	if err != nil {
		return -1, err
	}
	if d.Int64() == 0 {
		return 1, nil
	}
	return min + int(d.Int64()), nil
}
