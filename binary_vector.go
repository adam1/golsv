package golsv

import (
	"bytes"
	"fmt"
	"math/rand"
)

// xxx densify?
type BinaryVector []uint8

func NewBinaryVector(n int) BinaryVector {
	return make([]uint8, n)
}

func NewBinaryVectorFromInts(ints []uint8) BinaryVector {
	cols := len(ints)
	V := NewBinaryVector(cols)
	for j := 0; j < cols; j++ {
		V[j] = ints[j]
	}
	return V
}

func ZeroVector(n int) BinaryVector {
	return make([]uint8, n)
}

// xxx test
func (V BinaryVector) Add(U BinaryVector) BinaryVector {
	W := NewBinaryVector(len(V))
	for i := 0; i < len(V); i++ {
		W[i] = V[i] ^ U[i]
	}
	return W
}

// xxx test
func (V BinaryVector) Equal(U BinaryVector) bool {
	if len(V) != len(U) {
		return false
	}
	for i := 0; i < len(V); i++ {
		if V[i] != U[i] {
			return false
		}
	}
	return true
}

// xxx experimental
func (V BinaryVector) IntegerDotProduct(U BinaryVector) int {
	d := 0
	for i := 0; i < len(V); i++ {
		d += int(V[i]) * int(U[i])
	}
	return d
}

// xxx test
func (V BinaryVector) Length() int {
	return len(V)
}

// xxx rename to DenseBinaryMatrix
func (V BinaryVector) Matrix() *DenseBinaryMatrix {
	rows := len(V)
	M := NewDenseBinaryMatrix(rows, 1)
	for i := 0; i < rows; i++ {
		M.Set(i, 0, V[i])
	}
	return M
}

func (V BinaryVector) IsZero() bool {
	for _, elem := range V {
		if elem != 0 {
			return false
		}
	}
	return true
}

func (V BinaryVector) Project(n int, predicate func (i int) bool) BinaryVector {
	pV := NewBinaryVector(n)
	j := 0
	for i, b := range V {
		if predicate(i) {
			pV[j] = b
			j++
		}
	}
	return pV
}

// xxx test
func (V BinaryVector) RandomizeWithWeight(weight int) {
	n := len(V)
	for i := 0; i < n; i++ {
		V[i] = 0
	}
	for i := 0; i < weight; i++ {
		V[rand.Intn(n)] = 1
	}
}

// xxx test
func (V BinaryVector) SparseBinaryMatrix() *Sparse {
	rows := len(V)
	M := NewSparseBinaryMatrix(rows, 1)
	for i := 0; i < rows; i++ {
		M.Set(i, 0, V[i])
	}
	return M
}

func (V BinaryVector) String() string {
	var buf bytes.Buffer
	for _, b := range V {
		if b == 0 {
			buf.WriteString("0")
		} else {
			buf.WriteString("1")
		}
	}
	return buf.String()
}

func (V BinaryVector) SupportString() string {
	var buf bytes.Buffer
	first := true
	buf.WriteString("[+")
	for i, b := range V {
		if b == 1 {
			if !first {
				buf.WriteString(",")
			}
			buf.WriteString(fmt.Sprintf("%d", i))
			first = false
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func (V BinaryVector) Weight() int {
	weight := 0
	for _, b := range V {
		if b == 1 {
			weight++
		}
	}
	return weight
}

func randomBitVector(length int) BinaryVector {
	vector := make([]uint8, length)
	for i := 0; i < length; i++ {
		vector[i] = uint8(rand.Intn(2))
	}
	return vector
}

// xxx move to BinaryVector
func dotProduct(v1, v2 []int) int {
    product := 0
    for i := 0; i < len(v1); i++ {
        product += v1[i] * v2[i]
    }
    return product % 2
}

func AllBinaryVectors(n int) []BinaryVector {
	m := 1 << uint(n)
	vectors := make([]BinaryVector, m)
	for i := 0; i < m; i++ {
		support := make([]int, 0)
		for j := 0; j < n; j++ {
			coefficient := uint8((i >> j) & 1)
			if coefficient == 1 {
				support = append(support, j)
			}
		}
		coefficients := NewSparseBinaryMatrix(n, 1)
		coefficients.SetColumnData(0, support)
		vectors[i] = coefficients.AsColumnVector()
	}
	return vectors
}

// this is partially a compatibility shim for older code
// that still uses the BinaryVector type.
func EnumerateBinaryVectorSpaceList(generators BinaryMatrix) []BinaryVector {
	results := make([]BinaryVector, 0)
	EnumerateBinaryVectorSpace(generators, func(v BinaryMatrix, index int) (ok bool) {
		results = append(results, v.AsColumnVector())
		return true
	})
	return results
}

// note: will generate duplicate vectors if the generators are not
// linearly independent.
func EnumerateBinaryVectorSpace(generators BinaryMatrix, F func(v BinaryMatrix, index int) (ok bool) ) {
	n := generators.NumColumns()
	if n == 0 {
		F(NewDenseBinaryMatrix(generators.NumRows(), 1), 0)
		return
	}
	// iterate over integers 0 to 2^n - 1 and use the binary
	// representation of each integer to construct a linear
	// combination of the generators.  this is pretty fast if we
	// utilize optimized dense * sparse multiplication.
	m := pow(2, n)
	for i := 0; i < m; i++ {
		support := make([]int, 0)
		for j := 0; j < n; j++ {
			coefficient := uint8((i >> j) & 1)
			if coefficient == 1 {
				support = append(support, j)
			}
		}
		coefficients := NewSparseBinaryMatrix(n, 1)
		coefficients.SetColumnData(0, support)
		vector := LinearCombination(generators, coefficients)
		if !F(vector, i) {
			break
		}
	}
}

func SampleBinaryVectorSpaceList(generators BinaryMatrix, numSamples int, secureRandom bool) []BinaryVector {
	results := make([]BinaryVector, 0, numSamples)
	numGens := generators.NumColumns()
	if numGens == 0 {
		return results
	}		
	for i := 0; i < numSamples; i++ {
		coefficients := NewRandomSparseBinaryMatrix(numGens, 1, 0.5, secureRandom)
		results = append(results, LinearCombination(generators, coefficients).AsColumnVector())
	}
	return results
}

// xxx test
func RandomLinearCombination(generators BinaryMatrix) BinaryMatrix {
	numGens := generators.NumColumns()
	if numGens == 0 {
		return nil
	}		
	secureRandom := true
	coefficients := NewRandomSparseBinaryMatrix(numGens, 1, 0.5, secureRandom)
	return LinearCombination(generators, coefficients)
}

func LinearCombination(generators BinaryMatrix, coefficients *Sparse) BinaryMatrix {
	return generators.MultiplyRight(coefficients)
}

func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

