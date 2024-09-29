package golsv

import (
	"reflect"
	"testing"
)

func TestBinaryMatrix_BinaryVectorToMatrix(t *testing.T) {
	tests := []struct {
		name string
		v    BinaryVector
		want *DenseBinaryMatrix
	}{
		{
			name: "Test 1",
			v:    BinaryVector{1, 0, 1},
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{0},
				{1},
			}),
		},
		{
			name: "Test 2",
			v:    BinaryVector{1, 0, 1, 0},
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
				{0, 0, 0},
				{1, 0, 1},
				{0, 1, 0},
				{1, 1, 1},
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
				{0, 0, 0},
				{1, 0, 1},
				{0, 1, 0},
				{1, 1, 1},
				// note redundancy, which is fine for now
				{1, 1, 1},
				{0, 1, 0},
				{1, 0, 1},
				{0, 0, 0},
			},
		},
		{
			name:       "Test Zero Generators",
			generators: NewDenseBinaryMatrix(3, 0),
			want: []BinaryVector{
				{0, 0, 0},
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

func TestBinaryVectorAdd(t *testing.T) {
	tests := []struct {
		a, b, want BinaryVector
	}{
		{a: BinaryVector{0, 0, 0, 0}, b: BinaryVector{0, 0, 0, 0}, want: BinaryVector{0, 0, 0, 0}},
		{a: BinaryVector{0, 0, 0, 0}, b: BinaryVector{0, 0, 0, 1}, want: BinaryVector{0, 0, 0, 1}},
		{a: BinaryVector{0, 0, 0, 1}, b: BinaryVector{0, 0, 0, 1}, want: BinaryVector{0, 0, 0, 0}},
		{a: BinaryVector{0, 0, 0, 1}, b: BinaryVector{1, 0, 0, 1}, want: BinaryVector{1, 0, 0, 0}},
	}
	for i, tt := range tests {
		got := tt.a.Add(tt.b)
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("%d: got=%v want=%v", i, got, tt.want)
		}
	}
}

func TestBinaryVectorWeight(t *testing.T) {
	tests := []struct {
		v    BinaryVector
		want int
	}{
		{v: BinaryVector{0, 0, 0, 0}, want: 0},
		{v: BinaryVector{0, 0, 0, 1}, want: 1},
		{v: BinaryVector{0, 0, 1, 1}, want: 2},
	}
	for i, tt := range tests {
		got := tt.v.Weight()
		if got != tt.want {
			t.Errorf("%d: got=%d want=%d", i, got, tt.want)
		}
	}
}

func TestBinaryVectorProject(t *testing.T) {
	tests := []struct {
		v    BinaryVector
		proj []int
		want BinaryVector
	}{
		{v: BinaryVector{0, 0, 0, 0}, proj: []int{0, 1, 2}, want: BinaryVector{0, 0, 0}},
		{v: BinaryVector{0, 0, 0, 1}, proj: []int{0, 1, 2}, want: BinaryVector{0, 0, 0}},
		{v: BinaryVector{0, 0, 1, 1}, proj: []int{0, 1, 2}, want: BinaryVector{0, 0, 1}},
		{v: BinaryVector{1, 0, 1, 1}, proj: []int{1, 2}, want: BinaryVector{0, 1}},
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
