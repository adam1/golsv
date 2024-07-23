package golsv

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestBinaryMatrix_Transpose(t *testing.T) {
	tests := []struct {
		name   string
		matrix *DenseBinaryMatrix
		want   *DenseBinaryMatrix
	}{
		{
			name: "Test 1",
			matrix: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 1, 1},
				{1, 1, 0},
				{0, 0, 1},
			}),
			want: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 1, 0, 1},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.matrix.Transpose()
			if !got.Equal(tt.want) {
				t.Errorf("%s: got:\n%vwant:\n%v", tt.name, got, tt.want)
			}
		})
	}
}

func TestBinaryMatrix_RowOperationMatrix(t *testing.T) {
	tests := []struct {
		name       string
		operation  Operation
		wantMatrix *DenseBinaryMatrix
	}{
		{
			name:      "AddOp",
			operation: AddOp{0, 1},
			wantMatrix: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 0},
				{1, 1, 0},
				{0, 0, 1},
			}),
		},
		{
			name:      "SwapOp",
			operation: SwapOp{0, 1},
			wantMatrix: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0, 1, 0},
				{1, 0, 0},
				{0, 0, 1},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatrix := RowOperationMatrix(tt.operation, 3)
			if !reflect.DeepEqual(gotMatrix, tt.wantMatrix) {
				t.Errorf("%s: got:\n%vwant:\n%v", tt.name, gotMatrix, tt.wantMatrix)
			}
		})
	}
}

func TestBinaryMatrix_RowOperationsMatrix(t *testing.T) {
	tests := []struct {
		name        string
		operations  []Operation
		wantMatrix  *DenseBinaryMatrix
	}{
		{
			name: "Test 1",
			operations: []Operation{
				AddOp{0, 1},
				SwapOp{1, 2},
			},
			wantMatrix: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 0},
				{0, 0, 1},
				{1, 1, 0},
			}),
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatrix := RowOperationsMatrix(tt.operations, 3)
			if !reflect.DeepEqual(gotMatrix, tt.wantMatrix) {
				t.Errorf("%s: got:\n%vwant:\n%v", tt.name, gotMatrix, tt.wantMatrix)
			}
		})
	}
}

func TestBinaryMatrix_ColumnOperationMatrix(t *testing.T) {
	tests := []struct {
		name       string
		operation  Operation
		wantMatrix *DenseBinaryMatrix
	}{
		{
			name:      "AddOp",
			operation: AddOp{0, 1},
			wantMatrix: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 1, 0},
				{0, 1, 0},
				{0, 0, 1},
			}),
		},
		{
			name:      "SwapOp",
			operation: SwapOp{0, 1},
			wantMatrix: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{0, 1, 0},
				{1, 0, 0},
				{0, 0, 1},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatrix := ColumnOperationMatrix(tt.operation, 3)
			if !gotMatrix.Equal(tt.wantMatrix) {
				t.Errorf("%s: got:%v\n%s\nwant:%v\n%s", tt.name,
					gotMatrix, dumpMatrix(gotMatrix),
					tt.wantMatrix, dumpMatrix(tt.wantMatrix))
			}
		})
	}
}

func TestBinaryMatrix_ColumnOperationsMatrix(t *testing.T) {
	tests := []struct {
		name        string
		operations  []Operation
		wantMatrix  *DenseBinaryMatrix
	}{
		{
			name: "Test 1",
			operations: []Operation{
				AddOp{0, 1},
				SwapOp{1, 2},
			},
			wantMatrix: NewDenseBinaryMatrixFromRowInts([][]uint8{
				{1, 0, 1},
				{0, 0, 1},
				{0, 1, 0},
			}),
		},
		// Add more test cases here
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			M := ColumnOperationsMatrix(tt.operations, 3)
			if !M.Equal(tt.wantMatrix) {
				t.Errorf("%s: got:\n%vwant:\n%v", tt.name, M, tt.wantMatrix)
			}
		})
	}
}

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

// common tests for dense and sparse binary matrices

func newDense(rows, cols int) BinaryMatrix {
	return NewDenseBinaryMatrix(rows, cols)
}

func newSparse(rows, cols int) BinaryMatrix {
	return NewSparseBinaryMatrix(rows, cols)
}

func newDenseFromInts(ints [][]uint8) BinaryMatrix {
	return NewDenseBinaryMatrixFromRowInts(ints)
}

func newDenseRandom(rows, cols int) BinaryMatrix {
	return NewRandomDenseBinaryMatrix(rows, cols)
}

func newSparseFromInts(ints [][]uint8) BinaryMatrix {
	return NewSparseBinaryMatrixFromRowInts(ints)
}

func newDenseFromString(s string) BinaryMatrix {
	return NewDenseBinaryMatrixFromString(s)
}

func newSparseFromString(s string) BinaryMatrix {
	return NewSparseBinaryMatrixFromString(s)
}

func newSparseRandom(rows, cols int) BinaryMatrix {
	secureRandom := false
	return NewRandomSparseBinaryMatrix(rows, cols, 0.05, secureRandom)
}

type matrixType struct {
	constructor func(int, int) BinaryMatrix
	fromInts func([][]uint8) BinaryMatrix
	fromString func(string) BinaryMatrix
	random func(int, int) BinaryMatrix
}

var matrixTypes = []matrixType{
	{newDense, newDenseFromInts, newDenseFromString, newDenseRandom},
	{newSparse, newSparseFromInts, newSparseFromString, newSparseRandom},
}

func TestBinaryMatrix_GetSet(t *testing.T) {
	for _, mType := range matrixTypes {
		for _, d := range []int{1, 2, 3, 7, 9, 17, 24, 100} {
			M := mType.constructor(d, d)
			for i := 0; i < M.NumRows(); i++ {
				for j := 0; j < M.NumColumns(); j++ {
					b := uint8(rand.Intn(2))
					M.Set(i, j, b)
					c := M.Get(i, j)
					if c != b {
						t.Errorf("M.Get(%d, %d) = %d, want %d", i, j, c, b)
					}
				}
			}
		}
	}
}

func TestBinaryMatrix_NewFromString(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			name string
			str  string
			want BinaryMatrix
		}{
			{
				name: "Test 1",
				str:
				`1 0 1 0  
0 1 1 0
1 1 0 1
`,
				want: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
			},
			{
				name: "Test 2",
				str:
				`0 0 0
1 1 1
1 1 0`,
				want: mType.fromInts([][]uint8{
					{0, 0, 0},
					{1, 1, 1},
					{1, 1, 0},
				}),
			},
			{
				name: "Test 3",
				str:
				`0 0 1 0
0 0 0 0
0 0 0 0
0 0 0 1`,
				want: mType.fromInts([][]uint8{
					{0, 0, 1, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 0},
					{0, 0, 0, 1},
				}),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotMatrix := mType.fromString(tt.str)
				if !tt.want.Equal(gotMatrix) {
					t.Errorf("got\n%v\nwant\n%v", gotMatrix, tt.want)
				}
			})
		}
	}
}	

func TestBinaryMatrix_IsZero(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			name   string
			matrix BinaryMatrix
			want   bool
		}{
			{
				name: "Test 1",
				matrix: mType.fromInts([][]uint8{
					{0, 0, 0},
					{0, 0, 0},
				}),
				want: true,
			},
			{
				name: "Test 2",
				matrix: mType.fromInts([][]uint8{
					{0, 0, 0},
					{0, 1, 0},
				}),
				want: false,
			},
			{
				name: "Test 3",
				matrix: NewDenseBinaryMatrix(1, 362881),
				want: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isZero := tt.matrix.IsZero()
				if isZero != tt.want {
					t.Errorf("got %v, want %v", isZero, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_RandomCopy(t *testing.T) {
	for _, mType := range matrixTypes {
		trials := 100
		for i := 0; i < trials; i++ {
			matrix := mType.random(10, 10)
			copy := matrix.Copy()
			if !matrix.Equal(copy) {
				t.Errorf("copy = %v, want %v", copy, matrix)
			}
		}
	}
}

func TestBinaryMatrix_MultiplyRight(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			name   string
			matrix BinaryMatrix
			other  BinaryMatrix
			want   BinaryMatrix
		}{
			{
				name: "Test 1",
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				other: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				want: mType.fromInts([][]uint8{
					{0, 1, 1},
					{1, 0, 1},
					{1, 1, 1},
				}),
			},
			{
				name: "Test 2",
				matrix: mType.fromInts([][]uint8{
					{0, 1},
					{1, 0},
				}),
				other: mType.fromInts([][]uint8{
					{0, 1},
					{1, 0},
				}),
				want: mType.fromInts([][]uint8{
					{1, 0},
					{0, 1},
				}),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.matrix.MultiplyRight(tt.other)
				if !got.Equal(tt.want) {
					t.Errorf("%s: got:\n%vwant:\n%v", tt.name, got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_MultiplyLeft(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			name   string
			matrix BinaryMatrix
			other  BinaryMatrix
			want   BinaryMatrix
		}{
			{
				name: "Test 1",
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				other: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				want: mType.fromInts([][]uint8{
					{0, 1, 1},
					{1, 0, 1},
					{1, 1, 1},
				}),
			},
			// Add more test cases here
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.matrix.MultiplyLeft(tt.other)
				if !got.Equal(tt.want) {
					t.Errorf("%s: got:\n%vwant:\n%v", tt.name, got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_ColumnVector(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			name   string
			matrix BinaryMatrix
			column int
			want   BinaryVector
		}{
			{
				name: "Test 1",
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				column: 2,
				want: BinaryVector{1, 1, 0},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.matrix.ColumnVector(tt.column)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("%s: got:\n%v\nwant:\n%v", tt.name, got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_Columns(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			name   string
			matrix BinaryMatrix
			want   []BinaryVector
		}{
			{
				name: "Test 1",
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				want: []BinaryVector{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.matrix.Columns()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("%s: got:\n%v\nwant:\n%v", tt.name, got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_DenseSubmatrix(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			name             string
			matrix           BinaryMatrix
			rowStart, rowEnd int
			colStart, colEnd int
			want             BinaryMatrix
		}{
			{
				name: "Test 1",
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				rowStart: 0,
				rowEnd:   3,
				colStart: 0,
				colEnd:   4,
				want: NewDenseBinaryMatrixFromRowInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
			},
			{
				name: "Test 2",
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				rowStart: 1,
				rowEnd:   3,
				colStart: 2,
				colEnd:   4,
				want: NewDenseBinaryMatrixFromRowInts([][]uint8{
					{1, 0},
					{0, 1},
				}),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.matrix.DenseSubmatrix(tt.rowStart, tt.rowEnd, tt.colStart, tt.colEnd)
				if !got.Equal(tt.want) {
					t.Errorf("%s: got:%v\n%s\nwant:%v\n%s", tt.name, got, dumpMatrix(got), tt.want, dumpMatrix(tt.want))
				}
			})
		}
	}
}

func TestBinaryMatrix_Random(t *testing.T) {
	for _, mType := range matrixTypes {
		for rows := 1; rows < 10; rows++ {
			for cols := 1; cols < 10; cols++ {
				matrix := mType.random(rows, cols)
				// log.Printf("matrix = %v", matrix)
				gotRows := matrix.NumRows()
				if gotRows != rows {
					t.Errorf("rows = %d, want %d", gotRows, rows)
				}
				gotCols := matrix.NumColumns()
				if gotCols != cols {
					t.Errorf("cols = %d, want %d", gotCols, cols)
				}
				for i := 0; i < rows; i++ {
					for j := 0; j < cols; j++ {
						v := matrix.Get(i, j)
						if v != 0 && v != 1 {
							t.Errorf("value = %d, want 0 or 1", v)
						}
					}
				}
			}
		}
	}
}

func TestBinaryMatrix_MultiplyIdentity(t *testing.T) {
	for _, mType := range matrixTypes {
		trials := 100
		for i := 0; i < trials; i++ {
			matrix := mType.random(10, 10)
			identity := NewDenseBinaryMatrixIdentity(10)
			got := matrix.MultiplyRight(identity)
			if !got.Equal(matrix) {
				t.Errorf("got:\n%v\nwant:\n%v", got, matrix)
			}
		}
	}
}

func TestBinaryMatrix_AddColumn(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			matrix BinaryMatrix
			op     AddOp
			want   BinaryMatrix
		}{
			{
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				op: AddOp{0, 1},
				want: mType.fromInts([][]uint8{
					{1, 1, 1},
					{0, 1, 1},
					{1, 0, 0},
					{0, 0, 1},
				}),
			},
			{
				matrix: mType.fromInts([][]uint8{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				}),
				op: AddOp{0, 1},
				want: mType.fromInts([][]uint8{
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
				}),
			},
			{
				matrix: mType.fromInts([][]uint8{
					{0, 0, 1, 0, 0, 0, 0, 0, 0},
				}),
				op: AddOp{2, 3},
				want: mType.fromInts([][]uint8{
					{0, 0, 1, 1, 0, 0, 0, 0, 0},
				}),
			},
		}
		for i, tt := range tests {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				before := tt.matrix.Copy()
				tt.matrix.AddColumn(tt.op.Source, tt.op.Target)
				got := tt.matrix
				if !got.Equal(tt.want) {
					t.Errorf("before=%vop=%v\ngot=%v\nwant=%v", before, tt.op, got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_AddRow(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			matrix BinaryMatrix
			op     AddOp
			want   BinaryMatrix
		}{
			{
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				op: AddOp{0, 1},
				want: mType.fromInts([][]uint8{
					{1, 0, 1},
					{1, 1, 0},
					{1, 1, 0},
					{0, 0, 1},
				}),
			},
			{
				matrix: mType.fromInts([][]uint8{
					{0},
					{0},
					{0},
				}),
				op: AddOp{0, 1},
				want: mType.fromInts([][]uint8{
					{0},
					{0},
					{0},
				}),
			},
			{
				matrix: mType.fromInts([][]uint8{
					{0},
					{0},
					{1},
					{0},
				}),
				op: AddOp{2, 3},
				want: mType.fromInts([][]uint8{
					{0},
					{0},
					{1},
					{1},
				}),
			},
		}
		for i, tt := range tests {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				before := tt.matrix.Copy()
				tt.matrix.AddRow(tt.op.Source, tt.op.Target)
				got := tt.matrix
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("before=%vop=%v\ngot=%v\nwant=%v", before, tt.op, got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_ColumnWeight(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			m    BinaryMatrix
			col  int
			want int
		}{
			{
				m: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				col:  0,
				want: 2,
			},
			{
				m: mType.fromInts([][]uint8{
					{1, 0, 1, 0},
					{0, 1, 1, 0},
					{1, 1, 0, 1},
				}),
				col:  1,
				want: 2,
			},
			{
				m: mType.fromInts([][]uint8{
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
					t.Errorf("BinaryMatrix.ColumnWeight() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_ScanDown(t *testing.T) {
	for _, mType := range matrixTypes {
		tests := []struct {
			matrix   BinaryMatrix
			row, col int
			want     int
		}{
			{
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				row: 0, col: 0,
				want: 0,
			},
			{
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				row: 1, col: 0,
				want: 2,
			},
			{
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				row: 0, col: 1,
				want: 1,
			},
			{
				matrix: mType.fromInts([][]uint8{
					{1, 0, 1},
					{0, 1, 1},
					{1, 1, 0},
					{0, 0, 1},
				}),
				row: 3, col: 0,
				want: -1,
			},
			{
				matrix: mType.fromInts([][]uint8{
					{0},
					{0},
					{0},
					{0},
					{0},
					{0},
					{0},
					{0},
					{1},
				}),
				row: 0, col: 0,
				want: 8,
			},
		}
		for i, tt := range tests {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				got := tt.matrix.ScanDown(tt.row, tt.col)
				if got != tt.want {
					t.Errorf("got=%d, want=%d", got, tt.want)
				}
			})
		}
	}
}

func TestBinaryMatrix_ScanDownRandom(t *testing.T) {
	for _, mType := range matrixTypes {
		trials := 10
		maxSize := 10
		for i := 0; i < trials; i++ {
			rows := rand.Intn(maxSize) + 1
			cols := rand.Intn(maxSize) + 1
			M := mType.constructor(rows, cols)
			a := rand.Intn(rows)
			b := rand.Intn(cols)
			M.Set(a, b, 1)
			got := M.ScanDown(0, b)
			want := a
			if got != want {
				t.Errorf("M=%v a=%d b=%d got=%d want=%d\n%s", M, a, b, got, want,DumpMatrix(M))
			}
		}
	}
}

// xxx
func disabled_TestBinaryMatrix_WriteGobFile(t *testing.T) {
	N := 1000*1000*1000
	ops := make([]Operation, N)
	for i := 0; i < N; i++ {
		ops[i] = AddOp{Source: i, Target: i}
	}
	gob.Register(&AddOp{})
	//WriteGobFile("test.gob~", ops)
	WriteOperationsFile("test.gob~", ops)
}

func TestDenseTransposeSparseTranspose(t *testing.T) {
	trials := 10
	maxSize := 100
	for i := 0; i < trials; i++ {
		rows := rand.Intn(maxSize) + 1
		cols := rand.Intn(maxSize) + 1
		M := NewRandomDenseBinaryMatrix(rows, cols)
		T := M.Transpose()
		S := T.Sparse()
		U := S.Transpose()
		if !M.Equal(U) {
			t.Errorf("M=%v\n%s\nU=%v\n%s", M, dumpMatrix(M), U, dumpMatrix(U))
		}
	}
}

func TestGenericDense(t *testing.T) {
	n := 10
	M := NewSparseBinaryMatrixIdentity(n)
	N := genericDense(M)
	if !N.Equal(NewSparseBinaryMatrixIdentity(n)) {
		t.Errorf("N=%v", N)
	}
}

func TestGenericDenseSubmatrix(t *testing.T) {
	n := 10
	m := n - 1
	M := NewSparseBinaryMatrixIdentity(n)
	N := genericDenseSubmatrix(M, 0, m, 0, m)
	if !N.Equal(NewSparseBinaryMatrixIdentity(m)) {
		t.Errorf("N=%v", N)
	}
}

func TestRandomizeHotIndices(t *testing.T) {
	trials := 10
	maxSize := 20
	for i := 0; i < trials; i++ {
		rows := rand.Intn(maxSize) + 1
		weight := rand.Intn(rows + 1)
		got := randomizeHotIndices(rows, weight)
		// log.Printf("rows=%d weight=%d got=%v", rows, weight, got)
		if len(got) != weight {
			t.Errorf("rows=%d weight=%d got=%v", rows, weight, got)
		}
	}
}

func TestGenericRandomizeWithColumnWeight(t *testing.T) {
	trials := 10
	maxSize := 20
	for i := 0; i < trials; i++ {
		rows := rand.Intn(maxSize) + 1
		cols := rand.Intn(maxSize) + 1
		weight := rand.Intn(rows + 1)
		M := NewSparseBinaryMatrix(rows, cols)
		genericRandomizeWithColumnWeight(M, weight)
		for j := 0; j < cols; j++ {
			got := M.ColumnWeight(j)
			if got != weight {
				t.Errorf("col=%d weight=%d expected=%d \n%s", j, got, weight, dumpMatrix(M))
			}
		}
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
