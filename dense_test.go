package golsv

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestDenseBinaryMatrix_FromRows(t *testing.T) {
	tests := []struct {
		name string
		rows []BinaryVector
		want *DenseBinaryMatrix
	}{
		{
			name: "Test 1",
			rows: []BinaryVector{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			},
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
		},
		{
			name: "Test 2",
			rows: []BinaryVector{
				{0, 0, 0},
				{1, 1, 1},
				{1, 1, 0},
			},
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0, 0, 0},
				{1, 1, 1},
				{1, 1, 0},
			}),
		},
		{
			name: "Test 3",
			rows: []BinaryVector{
				{0, 0, 1, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 1},
			},
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0, 0, 1, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 1},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatrix := NewDenseBinaryMatrixFromRowVectors(tt.rows)
			if !reflect.DeepEqual(gotMatrix, tt.want) {
				t.Errorf("got\n%v\nwant\n%v", gotMatrix, tt.want)
			}
		})
	}
}

func TestDenseBinaryMatrix_DenseSubmatrix(t *testing.T) {
	tests := []struct {
		name string
		m    *DenseBinaryMatrix
		rowStart int
		rowEnd int
		colStart int
		colEnd int
		want *DenseBinaryMatrix
	}{
		{
			name: "Test 1",
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			rowStart: 0,
			rowEnd: 2,
			colStart: 0,
			colEnd: 2,
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0},
				{0, 1},
			}),
		},
		{
			name: "Test 2",
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			rowStart: 0,
			rowEnd: 3,
			colStart: 1,
			colEnd: 3,
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0, 1},
				{1, 1},
				{1, 0},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatrix := tt.m.DenseSubmatrix(tt.rowStart, tt.rowEnd, tt.colStart, tt.colEnd)
			if !gotMatrix.Equal(tt.want) {
				t.Errorf("got: %v\n%s\nwant: %v\n%s", gotMatrix, dumpMatrix(gotMatrix), tt.want, dumpMatrix(tt.want))
			}
		})
	}
}

func TestDenseBinaryMatrix_SparseExamples(t *testing.T) {
	tests := []struct {
		name string
		data [][]uint8
	}{
		{
			name: "Test 1",
			data: [][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			},
		},
		{
			name: "Test 2",
			data: [][]uint8{
				{0, 0, 0},
				{1, 1, 1},
				{1, 1, 0},
			},
		},
		{
			name: "Test 3",
			data: [][]uint8{
				{0, 1},
			},
		},
		{
			name: "Test 4",
			data: [][]uint8{
				{1},
				{0},
			},
		},
		{
			name: "Test 5",
			data: [][]uint8{
				{0},
				{0},
				{0},
				{0},
				{0},
				{0},
				{0},
				{0},
				{1},
				{1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dense := NewDenseBinaryMatrixFromRowInts(tt.data)
			want := NewSparseBinaryMatrixFromRowInts(tt.data)
			got := dense.Sparse()
			//if !reflect.DeepEqual(got, want) {
			if !want.Equal(got) {
				t.Errorf("got\n%v\nwant\n%v", dumpMatrix(got), dumpMatrix(want))
			}
		})
	}
}

func TestDenseBinaryMatrix_SparseRand(t *testing.T) {
	trials := 1000
	maxSize := 100
	for i := 0; i < trials; i++ {
		rows := rand.Intn(maxSize) + 1
		cols := rand.Intn(maxSize) + 1
		m := NewRandomDenseBinaryMatrix(rows, cols)
		// log.Printf("xxx m = %v\n%v", m, dumpMatrix(m))
		got := m.Sparse()
		if !m.Equal(got) {
			// log.Printf("xxx fail")
			t.Errorf("got\n%v\n%v\nwant\n%v\n%v", got, dumpMatrix(got), m, dumpMatrix(m))
		}
	}
}

func TestDenseBinaryMatrix_SparseParallel(t *testing.T) {
	trials := 100
	maxSize := 500
	for i := 0; i < trials; i++ {
		rows := rand.Intn(maxSize) + 1
		cols := rand.Intn(maxSize) + 1
		m := NewRandomDenseBinaryMatrix(rows, cols)
		// log.Printf("xxx m = %v\n%v", m, dumpMatrix(m))
		got := m.SparseParallel()
		if !m.Equal(got) {
			// log.Printf("xxx fail")
			t.Errorf("got\n%v\n%v\nwant\n%v\n%v", got, dumpMatrix(got), m, dumpMatrix(m))
		}
	}
}

func TestDenseMultiplyRightSparseVector(t *testing.T) {
	tests := []struct {
		name string
		m    *DenseBinaryMatrix
		v    *Sparse
		want *DenseBinaryMatrix
	}{
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			v: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{0},
				{1},
				{0},
			}),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0},
				{1},
				{1},
			}),
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			v: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{0},
				{1},
				{0},
				{1},
			}),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0},
				{1},
				{0},
			}),
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			v: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{1},
				{1},
				{1},
			}),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0},
				{0},
				{1},
			}),
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0, 0},
				{0, 1, 1, 0, 1},
				{1, 1, 0, 1, 1},
			}),
			v: NewSparseBinaryMatrixFromRowInts([][]uint8{
				{1},
				{1},
				{1},
				{0},
				{0},
			}),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0},
				{0},
				{0},
			}),
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.m.MultiplyRight(tt.v)
			if !got.Equal(tt.want) {
				t.Errorf("got\n%v\n%s\nwant\n%v\n%s",
					got, dumpMatrix(got), tt.want, dumpMatrix(tt.want))
			}
		})
	}
}

func TestDenseMultiplyRightSparseVsDense(t *testing.T) {
	trials := 10
	maxSize := 100
	for i := 0; i < trials; i++ {
		numRows := rand.Intn(maxSize) + 1
		numCols := rand.Intn(maxSize) + 1
		M := NewRandomDenseBinaryMatrix(numRows, numCols)
		B := NewRandomDenseBinaryMatrix(numCols, numCols)
		P := M.MultiplyRight(B)
		Q := M.MultiplyRight(B.Sparse())
		if !P.Equal(Q) {
			t.Errorf("P != Q\n%s\n%s\n%s\n%s",
				P, dumpMatrix(P), Q, dumpMatrix(Q))
		}
	}
}

func TestDenseBinaryMatrixMaxColumnSupport(t *testing.T) {
	tests := []struct {
		m    *DenseBinaryMatrix
		col  int
		want int
	}{
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			col:  0,
			want: 2,
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			col:  1,
			want: 2,
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			col:  2,
			want: 1,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if got := tt.m.MaxColumnSupport(tt.col); got != tt.want {
				t.Errorf("DenseBinaryMatrix.MaxColumnSupport() = %v, want %v", got, tt.want)
			}
		})
	}
		
}

func TestDenseBinaryMatrixColumnWeight(t *testing.T) {
	tests := []struct {
		m    *DenseBinaryMatrix
		col  int
		want int
	}{
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			col:  0,
			want: 2,
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			col:  1,
			want: 2,
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			col:  3,
			want: 1,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if got := tt.m.ColumnWeight(tt.col); got != tt.want {
				t.Errorf("DenseBinaryMatrix.ColumnWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDenseBinaryMatrixAddMatrix(t *testing.T) {
	tests := []struct {
		m    *DenseBinaryMatrix
		n    *DenseBinaryMatrix
		want *DenseBinaryMatrix
	}{
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
			}),
			n: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 0, 0},
				{0, 1, 0, 1},
			}),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0, 0, 1, 0},
				{0, 0, 1, 1},
			}),
		},
		{
			m: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
			n: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 0, 0},
				{0, 1, 0, 1},
				{0, 0, 1, 1},
			}),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0, 0, 1, 0},
				{0, 0, 1, 1},
				{1, 1, 1, 0},
			}),
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tt.m.AddMatrix(tt.n)
			got := tt.m
			if !got.Equal(tt.want) {
				t.Errorf("DenseBinaryMatrix.AddMatrix() = %v, want %v", got, tt.want)
			}
		})
	}
			
}

const bNumRows = 1000*100
const bNumCols = 1000*100
const bDensity = 0.005
const bCol = 10
const bRow = 1000
const bNumScans = 2000

func BenchmarkDenseBinaryMatrixScanDownOld(b *testing.B) {
	numRows := bNumRows
	numCols := bNumCols
	density := bDensity
	M := NewRandomDenseBinaryMatrixWithDensity(numRows, numCols, density)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < bNumScans; j++ {
			M.ScanDownOld(bRow + j, bCol + j)
		}
	}
}

func BenchmarkDenseBinaryMatrixScanDownNew(b *testing.B) {
	numRows := bNumRows
	numCols := bNumCols
	density := bDensity
	M := NewRandomDenseBinaryMatrixWithDensity(numRows, numCols, density)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < bNumScans; j++ {
			M.ScanDown(bRow + j, bCol + j)
		}
	}
}
