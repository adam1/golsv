package golsv

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Sparse stores a sparse mod 2 matrix in column-oriented
// list-of-lists format.  (aka "column support" format)
type Sparse struct {
	Rows, Cols int
	ColData []orderedIntSet
	verbose bool
}

func NewSparseBinaryMatrix(rows, cols int) *Sparse {
	colData := make([]orderedIntSet, cols)
	for j := 0; j < cols; j++ {
		colData[j] = make(orderedIntSet, 0)
	}
	return &Sparse{
		Rows: rows,
		Cols: cols,
		ColData: colData,
	}
}

func NewSparseBinaryMatrixFromRowInts(ints [][]uint8) *Sparse {
	rows := len(ints)
	cols := len(ints[0])
	M := NewSparseBinaryMatrix(rows, cols)
	genericSetFromRowInts(M, ints)
	return M
}

func NewSparseBinaryMatrixFromString(s string) *Sparse {
	// e.g.
	// 0 1 1 
	// 1 0 0 
	// 0 1 1 
	rowStrings := strings.Split(strings.TrimSpace(s), "\n")
	rows := len(rowStrings)
	cols := len(strings.Split(strings.TrimSpace(rowStrings[0]), " "))
	M := NewSparseBinaryMatrix(rows, cols)
	genericSetFromString(M, s)
	return M
}

func NewRandomSparseBinaryMatrix(rows, cols int, density float64, secureRandom bool) *Sparse {
	if secureRandom {
		M := NewRandomDenseBinaryMatrixWithDensity(rows, cols, density)
		return M.Sparse()
	}
	M := NewSparseBinaryMatrix(rows, cols)
	genericRandomizeWithDensity(M, density)
	return M
}

func NewSparseBinaryMatrixIdentity(size int) *Sparse {
	M := NewSparseBinaryMatrix(size, size)
	genericSetIdentityDiagonals(M)
	return M
}

// xxx test
func NewSparseBinaryMatrixDiagonal(rows, cols, rank int) *Sparse {
	M := NewSparseBinaryMatrix(rows, cols)
	for i := 0; i < rank; i++ {
		M.Set(i, i, 1)
	}
	return M
}

func (S *Sparse) AddColumn(source, target int) {
	S.ColData[target].MergeDrop(&S.ColData[source])
}

func (S *Sparse) AddRow(source, target int) {
	genericAddRow(S, source, target)
	// xxx this seems klunky;  not being utilized currently
// 	for j := 0; j < S.NumColumns(); j++ {
// 		if S.Get(source, j) == 1 {
// 			added := false
// 			for _, k := range S.ColData[j] {
// 				if k > target {
// 					break
// 				} else if k == target {
// 					S.ColData[j].Unset(k)
// 					added = true
// 				}
// 			}
// 			if !added {
// 				S.ColData[j].Set(target)
// 			}
// 		}
// 	}
}

func (S *Sparse) ApplyColumnOperation(op Operation) {
	genericApplyColumnOperation(S, op)
}

func (S *Sparse) ApplyRowOperation(op Operation) {
	genericApplyRowOperation(S, op)
}

func (S *Sparse) AppendColumn(v BinaryMatrix) {
	if v.NumColumns() != 1 {
		panic("AppendColumn: v must be a column vector")
	}
	if v.NumRows() != S.NumRows() {
		panic("AppendColumn: v must have the same number of rows as S")
	}
	if vsparse, ok := v.(*Sparse); ok {
		// optimized case
		S.ColData = append(S.ColData, vsparse.ColData[0])
		S.Cols++
	} else if vd, ok := v.(*DenseBinaryMatrix); ok {
		// optimized case
		vs := vd.Sparse()
		S.ColData = append(S.ColData, vs.ColData[0])
		S.Cols++
	} else {
		panic("AppendColumn: unimplemented")
	}
}

func (S *Sparse) AsColumnVector() BinaryVector {
	return genericAsColumnVector(S)
}

func (S *Sparse) AsRowVector() BinaryVector {
	return genericAsRowVector(S)
}

func (S *Sparse) ColumnIsZero(index int) bool {
	return genericColumnIsZero(S, index)
}

func (S *Sparse) Columns() []BinaryVector {
	return genericColumns(S)
}

func (S *Sparse) ColumnVector(index int) BinaryVector {
	return genericColumnVector(S, index)
}

// xxx test; add to interface
func (S *Sparse) ColumnWeight(index int) int {
	return len(S.ColData[index])
}

func (S *Sparse) Copy() BinaryMatrix {
	N := NewSparseBinaryMatrix(S.Rows, S.Cols)
	genericCopy(S, N)
	return N
}

func (S *Sparse) DenseSubmatrix(rowStart, rowEnd, colStart, colEnd int) *DenseBinaryMatrix {
	var M *DenseBinaryMatrix
	M = NewDenseBinaryMatrix(rowEnd - rowStart, colEnd - colStart)
	verbose := false
	colWorkGroup := NewWorkGroup(verbose)
	batch := make([]Work, 0)
	for j := colStart; j < colEnd; j++ {
		batch = append(batch, &copyToDenseInitedColumnWork{S, M, j, j - colStart, rowStart, rowEnd})
	}
	colWorkGroup.ProcessBatch(batch)
	return M
}

type copyToDenseInitedColumnWork struct {
	src *Sparse
	dest *DenseBinaryMatrix
	col, colDest int
	rowStart, rowEnd int
}

func (w *copyToDenseInitedColumnWork) Do() {
	for _, i := range w.src.ColData[w.col] {
		if i >= w.rowStart && i < w.rowEnd {
			w.dest.Set(i - w.rowStart, w.colDest, 1)
		}
	}
}

func (S *Sparse) Density(row, col int) float64 {
	count := 0
	for j := col; j < S.Cols; j++ {
		for _, m := range S.ColData[j] {
			if m >= row && m > 0 {
				count++
			}
		}
	}
	return float64(count) / float64((S.Rows - row) * (S.Cols - col))
}

// xxx
// func (S *Sparse) SplitColumns(offset int) (left, right []BinaryVector) {
// 	// xxx optimize
// 	return genericSplitColumns(S, offset)
// }

// func (S *Sparse) SplitRows(offset int) (left, right []BinaryVector) {
// 	return genericSplitRows(S, offset)
// }

// xxx
// func (S *Sparse) TrimColumns(n int) {
// 	S.ColData = S.ColData[n:]
// }

func (S *Sparse) Equal(other BinaryMatrix) bool {
	return genericEqual(S, other)
}

func (S *Sparse) Get(i, j int) uint8 {
	for _, k := range S.ColData[j] {
		if k > i {
			break
		} else if k == i {
			return 1
		}
	}
	return 0
}

func (S *Sparse) IsSmithNormalForm() (is bool, rank int) {
	d := S.Rows
	if d > S.Cols {
		d = S.Cols
	}
	diagState := -1
	for n := 0; n < d; n++ {
		x := S.Get(n, n)
		switch diagState {
		case -1:
			if x == 0 {
				diagState = 0
			} else if x == 1 {
				diagState = 1
				rank++
			}
		case 0:
			if x == 0 {
			} else if x == 1 {
				return false, 0
			}
		case 1:
			if x == 0 {
				diagState = 0
			} else if x == 1 {
				rank++
			}
		}
		switch x {
		case 0:
			if S.ColumnWeight(n) != 0 {
				return false, 0
			}
		case 1:
			if S.ColumnWeight(n) != 1 {
				return false, 0
			}
		}
	}
	for j := d; j < S.Cols; j++ {
		if S.ColumnWeight(j) != 0 {
			return false, 0
		}
	}
	return true, rank
}

func (S *Sparse) IsZero() bool {
	return genericIsZero(S)
}

func (S *Sparse) MultiplyLeft(B BinaryMatrix) BinaryMatrix {
	return B.MultiplyRight(S)
}

func (S *Sparse) MultiplyRight(B BinaryMatrix) BinaryMatrix {
	// xxx optimize?
	product := NewSparseBinaryMatrix(S.NumRows(), B.NumColumns())
	genericMultiply(S, B, product)
	return product
}

func (S *Sparse) NumColumns() int {
	return S.Cols
}

func (S *Sparse) NumRows() int {
	return S.Rows
}

func (S *Sparse) Overwrite(row int, col int, N BinaryMatrix) {
	// xxx optimize
	genericOverwrite(S, row, col, N)
}

func (S *Sparse) RowIsZero(index int) bool {
	return genericRowIsZero(S, index)
}

func (S *Sparse) GetRows() []BinaryVector {
	return genericGetRows(S)
}

func (S *Sparse) RowVector(index int) BinaryVector {
	return genericRowVector(S, index)
}

func (S *Sparse) ScanDown(row, col int) int {
	return S.ColData[col].Next(row)
}

func (S *Sparse) ScanRight(row, col int) int {
	// xxx optimize?
	return genericScanRight(S, row, col)
}

func (S *Sparse) Set(i, j int, b uint8) {
	if b == 0 {
		S.ColData[j].Unset(i)
	} else {
		S.ColData[j].Set(i)
	}
}

func (S *Sparse) SetColumnData(col int, data orderedIntSet) {
	S.ColData[col] = data
}

func (S *Sparse) SetVerbose(verbose bool) {
	S.verbose = verbose
}

func (S *Sparse) ShuffleColumns() {
	rand.Shuffle(S.Cols, func(i, j int) {
		S.ColData[i], S.ColData[j] = S.ColData[j], S.ColData[i]
	})
}

func (S *Sparse) Sparse() *Sparse {
	return S
}

func (S *Sparse) String() string {
	return genericString(S)
}

func (S *Sparse) Submatrix(rowStart, rowEnd, colStart, colEnd int) BinaryMatrix {
	if rowStart == 0 && rowEnd == S.Rows {
		// this particular case is optimized
		return &Sparse{
			Rows: S.Rows,
			Cols: colEnd - colStart,
			ColData: S.ColData[colStart:colEnd],
		}
	}
	return genericSubmatrix(S, rowStart, rowEnd, colStart, colEnd)
}

func (S *Sparse) Support(col int) []int {
	return S.ColData[col]
}

func (S *Sparse) SwapColumns(i, j int) {
	S.ColData[i], S.ColData[j] = S.ColData[j], S.ColData[i]
}

func (S *Sparse) SwapRows(i, j int) {
	// xxx could be optimized:
	// * we never need to allocate
	// * if the values i and j are unequal, we can memmove the data around in the ColdData
	genericSwapRows(S, i, j)
}

func (S *Sparse) Transpose() BinaryMatrix {
	var M *DenseBinaryMatrix
	M = NewDenseBinaryMatrix(S.NumColumns(), S.NumRows())

	statInterval := 1
	timeStart := time.Now()
	timeLastReport := timeStart
	jLastReport := 0
	
	for j := 0; j < S.NumColumns(); j++ {
		for _, i := range S.ColData[j] {
			M.Set(j, i, 1)
		}
		doneLastReport := j - jLastReport
		if S.verbose && doneLastReport >= statInterval {
			now := time.Now()
			intervalElapsed := now.Sub(timeLastReport)
			totalElapsed := now.Sub(timeStart)
			crate := float64(doneLastReport) / intervalElapsed.Seconds()
			trate := float64(j) / totalElapsed.Seconds()
			log.Printf("transpose: processed %d/%d (%.2f%%) crate=%1.0f trate=%1.0f",
				j, S.NumColumns(), 100*float64(j)/float64(S.NumColumns()),
				crate, trate)
			timeLastReport = now
			jLastReport = j
			if intervalElapsed.Seconds() > 10 {
				statInterval = 1 + statInterval / 2
			} else if intervalElapsed.Seconds() < 1 {
				statInterval *= 2
			}
		}
	}
	return M
}

// WriteFile writes the matrix to a text file whose format closely
// follows the Sparse structure:
//
//   - one line per column
//   - each line is a space-separated list of integers
//     representing the row indices of the 1s in that column.
//   - first line is a header with the number of rows and columns
//
func (S *Sparse) WriteFile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()
	header := fmt.Sprintf("%d %d\n", S.Rows, S.Cols)
	_, err = f.WriteString(header)
	if err != nil {
		panic(err)
	}
	var b strings.Builder
	for j := 0; j < S.Cols; j++ {
		b.Reset()
		ints := S.ColData[j]
		for _, m := range ints {
			fmt.Fprintf(&b, "%d ", m)
		}
		b.WriteString("\n")
		_, err = f.WriteString(b.String())
		if err != nil {
			panic(err)
		}
	}
}

func ReadSparseBinaryMatrixFile(filename string) BinaryMatrix {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()
	rows, cols := -1, -1
	r := bufio.NewReader(f)
	lineNum := 0
	var colData []orderedIntSet
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		fields := strings.Fields(line)
		ints := make([]int, len(fields))
		for i, s := range fields {
			m, err := strconv.Atoi(s)
			if err != nil {
				panic(err)
			}
			ints[i] = m
		}
		if lineNum > 0 {
			if rows == -1 || cols == -1 {
				panic("header missing")
			}
			colData = append(colData, ints)
		} else {
			if len(ints) < 2 {
				panic("malformed header")
			}
			rows = ints[0]
			cols = ints[1]
		}
		lineNum++
	}
	if cols != len(colData) {
		panic("mismatched number of columns")
	}
	return &Sparse{
		Rows:    rows,
		Cols:    cols,
		ColData: colData,
	}
}


type orderedIntSet []int

// xxx later: upgrade to binary search

func (O *orderedIntSet) Max() int {
	if len(*O) == 0 {
		return -1
	}
	return (*O)[len(*O)-1]
}

func (O *orderedIntSet) Set(m int) {
    for p := 0; p < len(*O); p++ {
		if (*O)[p] > m {
			*O = append((*O)[:p], append([]int{m}, (*O)[p:]...)...)
			return
		}
		if (*O)[p] == m {
			return
		}
	}
	*O = append(*O, m)
}

func (O *orderedIntSet) Toggle(m int) {
	for p := 0; p < len(*O); p++ {
		if (*O)[p] > m {
			*O = append((*O)[:p], append([]int{m}, (*O)[p:]...)...)
			return
		}
		if (*O)[p] == m {
			*O = append((*O)[:p], (*O)[p+1:]...)
			return
		}
	}
	*O = append(*O, m)
}

func (O *orderedIntSet) Unset(m int) {
	for p := 0; p < len(*O); p++ {
		if (*O)[p] == m {
			*O = append((*O)[:p], (*O)[p+1:]...)
		}
	}
}

// Next returns the next element in the set after the given integer,
// or -1 if there is no such element.
func (O *orderedIntSet) Next(m int) int {
	for p := 0; p < len(*O); p++ {
		if (*O)[p] >= m {
			return (*O)[p]
		}
	}
	return -1
}

// MergeDrop merges orderedIntSet N into O, dropping entries that
// appear both in O and N. aka XOR.
func (O *orderedIntSet) MergeDrop(N *orderedIntSet) {
	if O == N {
		*O = []int{}
		return
	}
	q := 0
	for _, n := range *N {
		added := false
		for p := q; p < len(*O); p++ {
			if (*O)[p] > n {
				*O = append(*O, -1)
				copy((*O)[p+1:], (*O)[p:])
				(*O)[p] = n
				added = true
				q = p
				break
			} else if (*O)[p] == n {
				*O = append((*O)[:p], (*O)[p+1:]...)
				added = true
				q = p
				break
			}
		}
		if !added {
			*O = append(*O, n)
		}
	}
}

func (O *orderedIntSet) Equal(N *orderedIntSet) bool {
	return reflect.DeepEqual(O, N)
}
