package golsv

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestBinaryVectorAdd(t *testing.T) {
	tests := []struct {
		a, b, want BinaryVector
	}{
		{a: NewBinaryVectorFromString("0000"), b: NewBinaryVectorFromString("0000"), want: NewBinaryVectorFromString("0000")},
		{a: NewBinaryVectorFromString("0000"), b: NewBinaryVectorFromString("0001"), want: NewBinaryVectorFromString("0001")},
		{a: NewBinaryVectorFromString("0001"), b: NewBinaryVectorFromString("0001"), want: NewBinaryVectorFromString("0000")},
		{a: NewBinaryVectorFromString("0001"), b: NewBinaryVectorFromString("1001"), want: NewBinaryVectorFromString("1000")},
	}
	for i, tt := range tests {
		got := tt.a.Add(tt.b)
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("%d: got=%v want=%v", i, got, tt.want)
		}
	}
}

func TestBinaryVectorMatrix(t *testing.T) {
	tests := []struct {
		name string
		v    BinaryVector
		want *DenseBinaryMatrix
	}{
		{
			name: "Test 1",
			v:    NewBinaryVectorFromString("101"),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{0},
				{1},
			}),
		},
		{
			name: "Test 2",
			v:    NewBinaryVectorFromString("1010"),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{0},
				{1},
				{0},
			}),
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.Matrix()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got:\n%vwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestBinaryVectorProject(t *testing.T) {
	tests := []struct {
		v    BinaryVector
		proj []int
		want BinaryVector
	}{
		{v: NewBinaryVectorFromString("0000"), proj: []int{0, 1, 2}, want: NewBinaryVectorFromString("000")},
		{v: NewBinaryVectorFromString("0001"), proj: []int{0, 1, 2}, want: NewBinaryVectorFromString("000")},
		{v: NewBinaryVectorFromString("0011"), proj: []int{0, 1, 2}, want: NewBinaryVectorFromString("001")},
		{v: NewBinaryVectorFromString("1011"), proj: []int{1, 2}, want: NewBinaryVectorFromString("01")},
	}
	for i, tt := range tests {
		hot := make(map[int]bool)
		for _, j := range tt.proj {
			hot[j] = true
		}
		predicate := func(i int) bool {
			_, ok := hot[i]
			return ok
		}
		got := tt.v.Project(len(tt.proj), predicate)
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("%d: got=%v want=%v", i, got, tt.want)
		}
	}
}

func TestBinaryVectorRandomizeWithWeight(t *testing.T) {
	trials := 10
	for i := 0; i < trials; i++ {
		n := 100
		w := rand.Intn(n)
		v := NewBinaryVector(n)
		v.RandomizeWithWeight(w)
		if v.Weight() != w {
			t.Errorf("got=%d want=%d", v.Weight(), w)
		}
	}
}

func TestBinaryVectorWeight(t *testing.T) {
	tests := []struct {
		v    BinaryVector
		want int
	}{
		{v: NewBinaryVectorFromString("0000"), want: 0},
		{v: NewBinaryVectorFromString("0001"), want: 1},
		{v: NewBinaryVectorFromString("0011"), want: 2},
	}
	for i, tt := range tests {
		got := tt.v.Weight()
		if got != tt.want {
			t.Errorf("%d: got=%d want=%d", i, got, tt.want)
		}
	}
}

func TestBinaryMatrix_LinearCombination(t *testing.T) {
	tests := []struct {
		name string
		generators   BinaryMatrix
		coefficients *Sparse
		want         BinaryMatrix
	}{
		{
			name: "Test 1",
			generators: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 1},
				{1, 0, 1},
			}),
			coefficients: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{0},
				{1},
			}),
			want: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{0},
				{1},
				{0},
			}),
		},
		{
			name: "Test 2",
			generators: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 1},
				{1, 0, 1},
			}),
			coefficients: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{1},
				{1},
			}),
			want: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{0},
				{0},
				{0},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LinearCombination(tt.generators, tt.coefficients)
			if !got.Equal(tt.want) {
				t.Errorf("got:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestBinaryMatrix_EnumerateBinaryVectorSpace(t *testing.T) {
	tests := []struct {
		name       string
		generators BinaryMatrix
		want       []BinaryVector
	}{
		{
			name:       "Test 1",
			generators: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0},
				{0, 1},
				{1, 0},
			}),
			want: []BinaryVector{
				NewBinaryVectorFromString("000"),
				NewBinaryVectorFromString("101"),
				NewBinaryVectorFromString("010"),
				NewBinaryVectorFromString("111"),
			},
		},
		{
			name:       "Test 2",
			generators: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 1},
				{1, 0, 1},
			}),
			want: []BinaryVector{
				NewBinaryVectorFromString("000"),
				NewBinaryVectorFromString("101"),
				NewBinaryVectorFromString("010"),
				NewBinaryVectorFromString("111"),
				// note redundancy, which is fine for now
				NewBinaryVectorFromString("111"),
				NewBinaryVectorFromString("010"),
				NewBinaryVectorFromString("101"),
				NewBinaryVectorFromString("000"),
			},
		},
		{
			name:       "Test Zero Generators",
			generators: NewDenseBinaryMatrix(3, 0),
			want: []BinaryVector{
				NewBinaryVectorFromString("000"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnumerateBinaryVectorSpaceList(tt.generators)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
	
}

func TestAllBinaryVectors(t *testing.T) {
	for n := 0; n < 5; n++ {
		res := AllBinaryVectors(n)
		if len(res) != 1<<uint(n) {
			t.Errorf("n=%d got=%d", n, len(res))
		}
		seen := make(map[string]bool)
		for _, v := range res {
			s := v.String()
			if seen[s] {
				t.Errorf("n=%d duplicate %s", n, s)
			}
			seen[s] = true
		}
	}
}

func TestEnumerateBinaryVectors(t *testing.T) {
	n := 3
	res := make([]BinaryVector, 0)
	buf := NewBinaryVector(n)
	EnumerateBinaryVectors(n, buf, func() (continue_ bool) {
		res = append(res, buf.Clone())
		return true
	})
	expected := []BinaryVector{
		NewBinaryVectorFromString("000"),
		NewBinaryVectorFromString("100"),
		NewBinaryVectorFromString("010"),
		NewBinaryVectorFromString("110"),
		NewBinaryVectorFromString("001"),
		NewBinaryVectorFromString("101"),
		NewBinaryVectorFromString("011"),
		NewBinaryVectorFromString("111"),
	}
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("got=%v want=%v", res, expected)
	}
}

func TestCountBits64(t *testing.T) {
	tests := []struct {
		input []byte
		want  int
	}{
		{input: []byte{}, want: 0},
		{input: []byte{0x00}, want: 0},
		{input: []byte{0x01}, want: 1},
		{input: []byte{0x03}, want: 2},
		{input: []byte{0x07}, want: 3},
		{input: []byte{0x0F}, want: 4},
		{input: []byte{0x1F}, want: 5},
		{input: []byte{0x3F}, want: 6},
		{input: []byte{0x7F}, want: 7},
		{input: []byte{0xFF}, want: 8},
		{input: []byte{0x00, 0x01}, want: 1},
		{input: []byte{0xFF, 0xFF}, want: 16},
		{input: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, want: 64},
		{input: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}, want: 65},
		{input: []byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}, want: 8},
		{input: []byte{0x55, 0xAA}, want: 8},
		{input: []byte{0x49, 0x92}, want: 6},
		{input: []byte{0xB6, 0x6D}, want: 10},
		{input: []byte{0x0F, 0xF0, 0x33, 0xCC, 0x99, 0x66, 0x5A, 0xA5, 0x01, 0x80}, want: 34},
	}
	for i, tt := range tests {
		got := countBits64(tt.input)
		if got != tt.want {
			t.Errorf("%d: countBits64(%x) = %d, want %d", i, tt.input, got, tt.want)
		}
	}
}

