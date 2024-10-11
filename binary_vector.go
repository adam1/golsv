package golsv

import (
	"bytes"
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"github.com/lukechampine/fastxor"
)

type BinaryVector struct {
	length int
	data []byte
}

func NewBinaryVector(n int) BinaryVector {
	return BinaryVector{
		n, 
		make([]byte, words(n)),
	}
}

func words(n int) int {
	return (n + 7) / 8
}

func NewBinaryVectorFromInts(ints []uint8) BinaryVector {
	rows := len(ints)
	v := NewBinaryVector(rows)
	for i, b := range ints {
		if b != 0 {
			v.Set(i, 1)
		}
	}
	return v
}

func NewBinaryVectorFromString(s string) BinaryVector {
	v := NewBinaryVector(len(s))
	for i, c := range s {
		if c == '1' {
			v.Set(i, 1)
		}
	}
	return v
}

func ZeroVector(n int) BinaryVector {
	return NewBinaryVector(n)
}

// xxx test
func (v BinaryVector) Add(u BinaryVector) BinaryVector {
	w := NewBinaryVector(v.Length())
	fastxor.Bytes(w.data, v.data, u.data)
	return w
}

// xxx test
func (v BinaryVector) AddInPlace(u BinaryVector) {
	fastxor.Bytes(v.data, v.data, u.data)
}

// xxx test
func (v BinaryVector) Equal(u BinaryVector) bool {
	return v.length == u.length && bytes.Equal(v.data, u.data)
}

// xxx test
func (v BinaryVector) Get(i int) uint8 {
	return (v.data[i/8] >> uint(i%8)) & 1
}

// xxx experimental
func (v BinaryVector) IntegerDotProduct(u BinaryVector) int {
	d := 0
	for i := 0; i < v.Length(); i++ {
		d += int(v.Get(i)) * int(u.Get(i))
	}
	return d
}

// xxx test
func (v BinaryVector) Length() int {
	return v.length
}

// xxx rename to DenseBinaryMatrix?
func (v BinaryVector) Matrix() *DenseBinaryMatrix {
	M := NewDenseBinaryMatrix(v.Length(), 1)
	copy(M.Data, v.data)
	return M
}

func (v BinaryVector) IsZero() bool {
	for _, d := range v.data {
		if d != 0 {
			return false
		}
	}
	return true
}

func (v BinaryVector) Project(n int, predicate func (i int) bool) BinaryVector {
	pv := NewBinaryVector(n)
	j := 0
	for i := 0; i < v.Length(); i++ {
		if predicate(i) {
			pv.Set(j, v.Get(i))
			j++
		}
	}
	return pv
}

// xxx test
func (v BinaryVector) RandomizeWithWeight(weight int) {
	n := v.Length()
	for i := 0; i < n; i++ {
		v.Set(i, 0)
	}
	for i := 0; i < weight; i++ {
		v.Set(mathrand.Intn(n), 1)
	}
}

// xxx test
func (v BinaryVector) Set(i int, b uint8) {
	if b == 0 {
		v.data[i/8] &= ^(1 << uint(i%8))
	} else {
		v.data[i/8] |= 1 << uint(i%8)
	}
}

// xxx test
func (v BinaryVector) SparseBinaryMatrix() *Sparse {
	rows := v.Length()
	M := NewSparseBinaryMatrix(rows, 1)
	for i := 0; i < rows; i++ {
		if v.Get(i) == 1 {
			M.Set(i, 0, 1)
		}
	}
	return M
}

func (v BinaryVector) String() string {
	var buf bytes.Buffer
	for i := 0; i < v.Length(); i++ {
		if v.Get(i) == 0 {
			buf.WriteString("0")
		} else {
			buf.WriteString("1")
		}
	}
	return buf.String()
}

func (v BinaryVector) SupportString() string {
	var buf bytes.Buffer
	first := true
	buf.WriteString("[+")
	for i := 0; i < v.Length(); i++ {
		if v.Get(i) == 1 {
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

func (v BinaryVector) Weight() int {
	weight := 0
	for i := 0; i < v.Length(); i++ {
		if v.Get(i) == 1 {
			weight++
		}
	}
	return weight
}

func randomBitVector(length int) BinaryVector {
	v := NewBinaryVector(length)
	_, err := rand.Read(v.data)
	if err != nil {
		panic(err)
	}
	return v
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

