package golsv

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// this file specifically deals with binary matrices, i.e. matrices
// with entries in the field F_2.

type BinaryMatrix interface {
	Add(N BinaryMatrix)
	AddColumn(source, target int)
	AddRow(source, target int)
	ApplyColumnOperation(op Operation)
	ApplyRowOperation(op Operation)
	AsColumnVector() BinaryVector
	AsRowVector() BinaryVector
	ColumnIsZero(index int) bool
	Columns() []BinaryVector
	ColumnVector(index int) BinaryVector
	ColumnWeight(index int) int
	Copy() BinaryMatrix
	Dense() *DenseBinaryMatrix
	DenseSubmatrix(rowStart, rowEnd, colStart, colEnd int) *DenseBinaryMatrix
	Density(row, col int) float64
	// xxx wip; for now hacked into SparseBinaryMatrix directly
	// DropColumn(index int) BinaryMatrix
	Equal(other BinaryMatrix) bool
	Get(i, j int) uint8
	IsZero() bool
	MultiplyLeft(B BinaryMatrix) BinaryMatrix
	MultiplyRight(B BinaryMatrix) BinaryMatrix
	NumColumns() int
	NumRows() int
	Overwrite(row int, col int, M BinaryMatrix)
	Project(func(int) bool) BinaryMatrix
	RowIsZero(index int) bool
	GetRows() []BinaryVector
	RowVector(index int) BinaryVector
	ScanDown(row, col int) int
	ScanRight(row, col int) int
	Set(i, j int, value uint8)
	Sparse() *Sparse
	// xxx
// 	SplitColumns(offset int) (BinaryMatrix, BinaryMatrix)
// 	SplitRows(offset int) (BinaryMatrix, BinaryMatrix)
	Submatrix(rowStart, rowEnd, colStart, colEnd int) BinaryMatrix
	SwapColumns(i, j int)
	SwapRows(i, j int)
	Transpose() BinaryMatrix
}

// xxx test
func genericAdd(M, N BinaryMatrix) {
	if M.NumRows() != N.NumRows() || M.NumColumns() != N.NumColumns() {
		panic("mismatched dimensions")
	}
	for i := 0; i < M.NumRows(); i++ {
		for j := 0; j < M.NumColumns(); j++ {
			M.Set(i, j, M.Get(i, j) ^ N.Get(i, j))
		}
	}
}

// xxx test
func genericAddColumn(M BinaryMatrix, source, target int) {
	for i := 0; i < M.NumRows(); i++ {
		b := M.Get(i, source)
		if b != 0 {
			M.Set(i, target, M.Get(i, target) ^ b)
		}
	}
}

// xxx test
func genericAddRow(M BinaryMatrix, source, target int) {
	for j := 0; j < M.NumColumns(); j++ {
		b := M.Get(source, j)
		if b != 0 {
			M.Set(target, j, M.Get(target, j) ^ b)
		}
	}
}

// xxx test
func genericApplyColumnOperation(M BinaryMatrix, op Operation) {
	switch op := op.(type) {
	case SwapOp:
		M.SwapColumns(op.I, op.J)
	case AddOp:
		M.AddColumn(op.Source, op.Target)
	default:
		panic("unknown row operation")
	}
}

// xxx test
func genericApplyRowOperation(M BinaryMatrix, op Operation) {
	switch op := op.(type) {
	case SwapOp:
		M.SwapRows(op.I, op.J)
	case AddOp:
		M.AddRow(op.Source, op.Target)
	default:
		panic("unknown row operation")
	}
}

// xxx test
func genericAsColumnVector(M BinaryMatrix) BinaryVector {
	if M.NumColumns() != 1 {
		panic("number of columns is not one")
	}
	return M.ColumnVector(0)
}

// xxx test
func genericAsRowVector(M BinaryMatrix) BinaryVector {
	if M.NumRows() != 1 {
		panic("number of rows is not one")
	}
	return M.RowVector(0)
}

// xxx test
func genericColumnIsZero(M BinaryMatrix, index int) bool {
	for i := 0; i < M.NumRows(); i++ {
		if M.Get(i, index) != 0 {
			return false
		}
	}
	return true
}

// xxx test
func genericColumns(M BinaryMatrix) []BinaryVector {
	cols := M.NumColumns()
	res := make([]BinaryVector, cols)
	for j := 0; j < cols; j++ {
		res[j] = M.ColumnVector(j)
	}
	return res
}

// xxx test
func genericColumnVector(M BinaryMatrix, index int) BinaryVector {
	rows := M.NumRows()
	vector := NewBinaryVector(rows)
	for i := 0; i < rows; i++ {
		if M.Get(i, index) != 0 {
			vector.Set(i, 1)
		}
	}
	return vector
}

// xxx test
func genericCopy(source, target BinaryMatrix) {
	if (source.NumRows() != target.NumRows()) ||
		(source.NumColumns() != target.NumColumns()) {
		panic("mismatched dimensions")
	}
	for i := 0; i < source.NumRows(); i++ {
		for j := 0; j < source.NumColumns(); j++ {
			target.Set(i, j, source.Get(i, j))
		}
	}
}

func genericDense(M BinaryMatrix) *DenseBinaryMatrix {
	return genericDenseSubmatrix(M, 0, M.NumRows(), 0, M.NumColumns())
}

func genericDenseSubmatrix(M BinaryMatrix, rowStart, rowEnd, colStart, colEnd int) *DenseBinaryMatrix {
	rows := rowEnd - rowStart
	cols := colEnd - colStart
	N := NewDenseBinaryMatrix(rows, cols)
	for i := rowStart; i < rowEnd; i++ {
		for j := colStart; j < colEnd; j++ {
			N.Set(i - rowStart, j - colStart, M.Get(i, j))
		}
	}
	return N
}

// xxx test
func genericDensity(M BinaryMatrix, row, col int) float64 {
	rows := M.NumRows()
	cols := M.NumColumns()
	count := 0
	for i := row; i < rows; i++ {
		for j := col; j < cols; j++ {
			if M.Get(i, j) != 0 {
				count++
			}
		}
	}
	return float64(count) / float64((rows - row) * (cols - col))
}

// xxx test
func genericEqual(A, B BinaryMatrix) bool {
	if (A.NumRows() != B.NumRows()) ||
		(A.NumColumns() != B.NumColumns()) {
		return false
	}
	for i := 0; i < A.NumRows(); i++ {
		for j := 0; j < A.NumColumns(); j++ {
			if A.Get(i, j) != B.Get(i, j) {
				return false
			}
		}
	}
	return true
}

// xxx deprecate
// func genericFirstNonPivotCol(M BinaryMatrix) int {
// 	rows := M.NumRows()
// 	cols := M.NumColumns()
// 	firstNonPivotCol := -1
// 	for j := 0; j < cols; j++ {
// 		for i := 0; i < rows; i++ {
// 			if M.Get(i, j) == 1 {
// 				break
// 			}
// 			if i == rows - 1 {
// 				firstNonPivotCol = j
// 				break
// 			}
// 		}
// 		if firstNonPivotCol != -1 {
// 			break
// 		}
// 	}
// 	return firstNonPivotCol
// }

// xxx test
func genericIsZero(M BinaryMatrix) bool {
	for i := 0; i < M.NumRows(); i++ {
		if !M.RowIsZero(i) {
			return false
		}
	}
	return true
}

// A * B = C
func genericMultiply(A, B, C BinaryMatrix) {
	rows := A.NumRows()
	cols := B.NumColumns()
	inner := B.NumRows()

	if (A.NumColumns() != inner) ||
		(C.NumRows() != rows) ||
		(C.NumColumns() != cols) {
		panic(fmt.Sprintf("matrix shapes not compatible: (%dx%d)*(%dx%d)=(%dx%d)",
			A.NumRows(), A.NumColumns(), B.NumRows(), B.NumColumns(),
			C.NumRows(), C.NumColumns()))
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			for k := 0; k < inner; k++ {
				C.Set(i, j, C.Get(i, j) ^ A.Get(i, k) * B.Get(k,j))
			}
		}
	}
}

func genericOverwrite(M BinaryMatrix, row int, col int, N BinaryMatrix) {
	rows := N.NumRows()
	cols := N.NumColumns()
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			M.Set(i + row, j + col, N.Get(i, j))
		}
	}
}

// xxx test
func genericProject(M BinaryMatrix, rowPredicate func(int) bool) BinaryMatrix {
	var numRows int
	for i := 0; i < M.NumRows(); i++ {
		if rowPredicate(i) {
			numRows++
		}
	}
	N := NewSparseBinaryMatrix(numRows, M.NumColumns())
	row := 0
	for i := 0; i < M.NumRows(); i++ {
		if rowPredicate(i) {
			for j := 0; j < M.NumColumns(); j++ {
				N.Set(row, j, M.Get(i, j))
			}
			row++
		}
	}
	return N
}

func genericSparse(M BinaryMatrix) *Sparse {
	rows := M.NumRows()
	cols := M.NumColumns()
	N := NewSparseBinaryMatrix(rows, cols)

	statInterval := 1000
	startTime := time.Now()
	lastStatTime := startTime
	lastStat := 0

	for j := 0; j < cols; j++ {
		support := make([]int, 0, 128) // guessing that 128 is a good initial capacity; saves a few allocations
		for i := 0; i < rows; i++ {
			if M.Get(i, j) == 1 {
				support = append(support, i)
			}
		}
		N.SetColumnData(j, support)
		if j - lastStat >= statInterval {
			now := time.Now()
			elapsed := now.Sub(lastStatTime)
			interval := j - lastStat
			crate := float64(interval) / elapsed.Seconds()
			telapsed := now.Sub(startTime)
			trate := float64(j) / telapsed.Seconds()
			log.Printf("sparse: %d/%d (%.2f%%) columns processed; crate=%d trate=%d\n",
				j, cols, float64(j) / float64(cols) * 100.0, int(crate), int(trate))
			if elapsed.Seconds() < 1 {
				statInterval *= 2
			} else if elapsed.Seconds() > 10 {
				statInterval = 1 + statInterval / 2
			}
			lastStatTime = now
			lastStat = j
		}
	}
	return N
}

func genericTranspose(M BinaryMatrix) BinaryMatrix {
	rows := M.NumRows()
	cols := M.NumColumns()
	N := NewSparseBinaryMatrix(cols, rows)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if M.Get(i, j) == 1 {
				N.Set(j, i, 1)
			}
		}
	}
	return N
}
	
// xxx test
// func genericSplitColumns(M BinaryMatrix, offset int) (left, right BinaryMatrix) {
// 	cols := M.NumColumns()
// 	if offset <= 0 {
// 		return EmptyBinaryMatrix, M
// 	}
// 	if offset >= cols {
// 		return M, EmptyBinaryMatrix
// 	}
// 	// xxx
// 	left 
	

// 	left = make([]BinaryVector, offset)
// 	for i := 0; i < offset; i++ {
// 		left[i] = M.ColumnVector(i)
// 	}
// 	right = make([]BinaryVector, cols - offset)
// 	for i := offset; i < cols; i++ {
// 		right[i - offset] = M.ColumnVector(i)
// 	}
// 	return left, right
// }

// // xxx test
// func genericSplitRows(M BinaryMatrix, offset int) (left, right BinaryMatrix) {
// 	rows := M.NumRows()
// 	if offset >= rows {
// 		return M.GetRows(), []BinaryVector{}
// 	}
// 	if offset <= 0 {
// 		return []BinaryVector{}, M.GetRows()
// 	}
// 	left = make([]BinaryVector, offset)
// 	for i := 0; i < offset; i++ {
// 		left[i] = M.RowVector(i)
// 	}
// 	right = make([]BinaryVector, rows - offset)
// 	for i := offset; i < rows; i++ {
// 		right[i - offset] = M.RowVector(i)
// 	}
// 	return left, right
// }

// xxx test
func genericSubmatrix(M BinaryMatrix, rowStart, rowEnd, colStart, colEnd int) BinaryMatrix {
	rows := rowEnd - rowStart
	cols := colEnd - colStart
	var N BinaryMatrix
	if _, ok := M.(*DenseBinaryMatrix); ok {
		N = NewDenseBinaryMatrix(rows, cols)
	} else {
		N = NewSparseBinaryMatrix(rows, cols)
	}
	for i := rowStart; i < rowEnd; i++ {
		for j := colStart; j < colEnd; j++ {
			if M.Get(i, j) == 1 {
				N.Set(i - rowStart, j - colStart, 1)
			}
		}
	}
	return N
}

func genericRandomize(M BinaryMatrix) {
	rows := M.NumRows()
	cols := M.NumColumns()
	for i := 0; i < rows; i++ {
		vector := randomBitVector(cols)
		for j := 0; j < cols; j++ {
			M.Set(i, j, vector.Get(j))
		}
	}
}

func genericRandomizeWithDensity(M BinaryMatrix, density float64) {
	// xxx would be more robust to use better random number
	// generation; generate one large bit string and then
	// use that to fill in the matrix
	rows := M.NumRows()
	cols := M.NumColumns()
	for i := 0; i < rows; i++ {
		vector := randomBitVector(cols)
		for j := 0; j < cols; j++ {
			if rand.Float64() < density {
				M.Set(i, j, vector.Get(j))
			}
		}
	}
}

func genericRandomizeWithColumnWeight(M BinaryMatrix, weight int) {
	rows := M.NumRows()
	cols := M.NumColumns()
	for j := 0; j < cols; j++ {
		hotIndices := randomizeHotIndices(rows, weight)
		for i := 0; i < rows; i++ {
			M.Set(i, j, 0)
		}
		for _, i := range hotIndices {
			M.Set(i, j, 1)
		}
	}
}

func randomizeHotIndices(rows, weight int) []int {
	if weight > rows {
		panic("weight > rows")
	}
	all := make([]int, rows)
	for i := 0; i < rows; i++ {
		all[i] = i
	}
	rand.Shuffle(rows, func(i, j int) {
		all[i], all[j] = all[j], all[i]
	})
	return all[:weight]
}

// xxx test
func genericRowIsZero(M BinaryMatrix, index int) bool {
	for j := 0; j < M.NumColumns(); j++ {
		if M.Get(index, j) != 0 {
			return false
		}
	}
	return true
}

// xxx test
func genericGetRows(M BinaryMatrix) []BinaryVector {
	rows := M.NumRows()
	res := make([]BinaryVector, rows)
	for i := 0; i < rows; i++ {
		res[i] = M.RowVector(i)
	}
	return res
}

// xxx test
func genericRowVector(M BinaryMatrix, index int) BinaryVector {
	cols := M.NumColumns()
	vector := NewBinaryVector(cols)
	for j := 0; j < cols; j++ {
		if M.Get(index, j) != 0 {
			vector.Set(j, 1)
		}
	}
	return vector
}

// xxx test
func genericScanDown(M BinaryMatrix, row, col int) int {
	rows := M.NumRows()
	for i := row; i < rows; i++ {
		if M.Get(i, col) != 0 {
			return i
		}
	}
	return -1
}

// xxx test
func genericScanRight(M BinaryMatrix, row, col int) int {
	cols := M.NumColumns()
	for j := col; j < cols; j++ {
		if M.Get(row, j) != 0 {
			return j
		}
	}
	return -1
}

func genericSetFromRowInts(M BinaryMatrix, ints [][]uint8) {
	rows := M.NumRows()
	cols := M.NumColumns()
	if (len(ints) != rows) || (len(ints[0]) != cols) {
		panic(fmt.Sprintf("matrix shape not compatible: (%dx%d) != (%dx%d)",
			rows, cols, len(ints), len(ints[0])))
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			M.Set(i, j, ints[i][j])
		}
	}	
}

func genericSetFromString(M BinaryMatrix, s string) {
	// e.g.
	// 0 1 1 
	// 1 0 0 
	// 0 1 1 
	rows := M.NumRows()
	cols := M.NumColumns()

	rowStrings := strings.Split(strings.TrimSpace(s), "\n")
	tRows := len(rowStrings)
	tCols := len(strings.Split(strings.TrimSpace(rowStrings[0]), " "))

	if (tRows != rows) || (tCols != cols) {
		panic(fmt.Sprintf("matrix shape not compatible: (%dx%d) != (%dx%d)",
			rows, cols, tRows, tCols))
	}
	for i := 0; i < rows; i++ {
		row := strings.TrimSpace(rowStrings[i])
		colStrings := strings.Split(row, " ")
		for j := 0; j < cols; j++ {
			var b uint8
			if colStrings[j] == "0" {
			} else if colStrings[j] == "1" {
				b = 1
			} else {
				panic("invalid matrix element")
			}
			M.Set(i, j, b)
		}
	}
}

func genericSetFromRowVectors(M BinaryMatrix, rowData []BinaryVector) {
	rows := M.NumRows()
	cols := M.NumColumns()
	if (len(rowData) != rows) || (rowData[0].Length() != cols) {
		panic(fmt.Sprintf("matrix shape not compatible: (%dx%d) != (%dx%d)",
			rows, cols, len(rowData), rowData[0].Length()))
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if rowData[i].Get(j) != 0 {
				M.Set(i, j, 1)
			}
		}
	}	
}

func genericSetIdentityDiagonals(M BinaryMatrix) {
	rows := M.NumRows()
	cols := M.NumColumns()
	d := rows
	if cols < rows {
		d = cols
	}
	for i := 0; i < d; i++ {
		M.Set(i, i, 1)
	}
}

func genericString(M BinaryMatrix) string {
	str := fmt.Sprintf("%d x %d matrix", M.NumRows(), M.NumColumns())
	dump := false
	if dump {
		str += "\n" + dumpMatrix(M)
	}
	return str
}

func dumpMatrix(M BinaryMatrix) string {
	str := ""
	for i := 0; i < M.NumRows(); i++ {
		for j := 0; j < M.NumColumns(); j++ {
			if M.Get(i, j) == 1 {
				str += "1 "
			} else {
				str += "0 "
			}
		}
		str += "\n"
	}
	return str
}

func DumpMatrix(M BinaryMatrix) string {
	return dumpMatrix(M)
}

// xxx test
func genericSwapColumns(M BinaryMatrix, i, j int) {
	for k := 0; k < M.NumRows(); k++ {
		a := M.Get(k, i)
		b := M.Get(k, j)
		if a != b {
			M.Set(k, i, b)
			M.Set(k, j, a)
		}
	}
}

// xxx test
func genericSwapRows(M BinaryMatrix, i, j int) {
	for k := 0; k < M.NumColumns(); k++ {
		a := M.Get(i, k)
		b := M.Get(j, k)
		if a != b {
			M.Set(i, k, b)
			M.Set(j, k, a)
		}
	}
}


func vectorsToMatrix(vectors [][]int) [][]int {
    rows := len(vectors)
    cols := len(vectors[0])

    matrix := make([][]int, rows)
    for i := 0; i < rows; i++ {
        matrix[i] = make([]int, cols)
        copy(matrix[i], vectors[i])
    }
    return matrix
}

func countPivotRows(m BinaryMatrix) int {
	count := 0
    rows := m.NumRows()
	for i := 0; i < rows; i++ {
		if m.RowIsZero(i) {
			break
		} else {
			count++
		}
	}
	return 0
}

