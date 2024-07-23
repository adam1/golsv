package golsv

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"strings"
	"github.com/lukechampine/fastxor"
)

const columnFirst = true // xxx get rid of this; assume it everywhere
const wordSize = 8

type DenseBinaryMatrix struct {
 	Rows, Cols, Stride int
	Data []byte
	buf []byte
}

func NewDenseBinaryMatrix(rows, cols int) *DenseBinaryMatrix {
	// align each row/column to a word boundary
	stride, words := stride(rows, cols)
	return &DenseBinaryMatrix{
		Rows: rows,
		Cols: cols,
		Stride: stride,
		Data: make([]byte, words),
		buf: make([]byte, stride),
	}
}

func stride(rows, cols int) (stride, words int) {
	// align each row/column to a word boundary
	if columnFirst {
		stride = rows / wordSize
		if rows % wordSize != 0 {
			stride += 1
		}
		words = stride * cols
	} else {
		stride = cols / wordSize
		if cols % wordSize != 0 {
			stride += 1
		}
		words = rows * stride
	}
	return
}

func NewDenseBinaryMatrixFromRowInts(ints [][]uint8) *DenseBinaryMatrix {
	rows := len(ints)
	cols := len(ints[0])
	M := NewDenseBinaryMatrix(rows, cols)
	genericSetFromRowInts(M, ints)
	return M
}

func NewDenseBinaryMatrixFromRowVectors(rowData []BinaryVector) *DenseBinaryMatrix {
	rows := len(rowData)
	cols := len(rowData[0])
	M := NewDenseBinaryMatrix(rows, cols)
	genericSetFromRowVectors(M, rowData)
	return M
}

func NewDenseBinaryMatrixFromString(s string) *DenseBinaryMatrix {
	// e.g.
	// 0 1 1 
	// 1 0 0 
	// 0 1 1 
	rowStrings := strings.Split(strings.TrimSpace(s), "\n")
	rows := len(rowStrings)
	cols := len(strings.Split(strings.TrimSpace(rowStrings[0]), " "))
	M := NewDenseBinaryMatrix(rows, cols)
	genericSetFromString(M, s)
	return M
}


func NewRandomDenseBinaryMatrix(rows, cols int) *DenseBinaryMatrix {
	return NewRandomDenseBinaryMatrixWithDensity(rows, cols, 0.5)
}

func NewRandomDenseBinaryMatrixWithDensity(rows, cols int, density float64) *DenseBinaryMatrix {
	stride, words := stride(rows, cols)
	M := &DenseBinaryMatrix{
		Rows: rows,
		Cols: cols,
		Stride: stride,
		Data: make([]byte, words),
		buf: make([]byte, stride),
	}
	noiseLen := rows * cols * 8
	noise := make([]byte, noiseLen)
	_, err := rand.Read(noise)
	if err != nil {
		panic(err)
	}
	var i, j int
	threshold := uint64(density * math.MaxUint64)
	for k := 0; k < noiseLen; k += 8 {
		bytes := noise[k:k+8]
		num := binary.LittleEndian.Uint64(bytes)
		if num <= threshold {
			M.Set(i, j, 1)
		}
		i++
		if i == rows {
			i = 0
			j++
		}
	}
	return M
}

func NewDenseBinaryMatrixIdentity(n int) BinaryMatrix {
	M := NewDenseBinaryMatrix(n, n)
	genericSetIdentityDiagonals(M)
	return M
}

func (M *DenseBinaryMatrix) Add(N BinaryMatrix) {
	// xxx optimize: if N is dense, can use fastxor.Bytes
	genericAdd(M, N)
}

func (M *DenseBinaryMatrix) AddColumn(source, target int) {
	if true && columnFirst {
		sourceWord, sourceBit := M.index(0, source)
		targetWord, targetBit := M.index(0, target)
		if sourceBit != 0 || targetBit != 0 {
			panic("alignment failure")
		}
		fastxor.Bytes(
			M.Data[targetWord:targetWord + M.Stride],
			M.Data[targetWord:targetWord + M.Stride],
			M.Data[sourceWord:sourceWord + M.Stride])
	} else {
		genericAddColumn(M, source, target)
	}
}

// xxx test
func (M *DenseBinaryMatrix) AddMatrix(B BinaryMatrix) {
	if M.Rows != B.NumRows() || M.Cols != B.NumColumns() {
		panic("dimension mismatch")
	}
	Bdense, ok := B.(*DenseBinaryMatrix)
	if ok {
		for j := 0; j < M.Cols; j++ {
			fastxor.Bytes(
				M.Data[j * M.Stride:(j + 1) * M.Stride],
				M.Data[j * M.Stride:(j + 1) * M.Stride],
				Bdense.Data[j * M.Stride:(j + 1) * M.Stride])
		}
	} else {
		panic("not implemented")
	}
}

func (M *DenseBinaryMatrix) AddRow(source, target int) {
	genericAddRow(M, source, target)
}

func (M *DenseBinaryMatrix) ApplyColumnOperation(op Operation) {
	genericApplyColumnOperation(M, op)
}

func (M *DenseBinaryMatrix) ApplyRowOperation(op Operation) {
	genericApplyRowOperation(M, op)
}

func (M *DenseBinaryMatrix) AsColumnVector() BinaryVector {
	return genericAsColumnVector(M)
}

func (M *DenseBinaryMatrix) AsRowVector() BinaryVector {
	return genericAsRowVector(M)
}

func (M *DenseBinaryMatrix) ColumnIsZero(index int) bool {
	return genericColumnIsZero(M, index)
}

func (M *DenseBinaryMatrix) Columns() []BinaryVector {
	return genericColumns(M)
}

func (M *DenseBinaryMatrix) ColumnWeight(col int) int {
	w := 0
	for k := 0; k < M.Stride; k++ {
		c := M.Data[col * M.Stride + k]
		for b := 0; b < wordSize; b++ {
			if (c >> b) & 1 == 1 {
				w++
			}
		}
	}
	return w
}

func (M *DenseBinaryMatrix) ColumnVector(index int) BinaryVector {
// 	xxx if BinaryVector were packed into bytes, this would be simple
// 	if columnFirst {
// 	}
	return genericColumnVector(M, index)
}

func (M *DenseBinaryMatrix) Copy() BinaryMatrix {
	N := NewDenseBinaryMatrix(M.Rows, M.Cols)
	genericCopy(M, N)
	return N
}

func (M *DenseBinaryMatrix) Dense() *DenseBinaryMatrix {
	return M
}

func (M *DenseBinaryMatrix) DenseSubmatrix(rowStart, rowEnd, colStart, colEnd int) *DenseBinaryMatrix {
	if columnFirst {
		if rowStart == 0 && rowEnd == M.Rows {
			// this particular case is optimized
			return &DenseBinaryMatrix{
				Rows: M.Rows,
				Cols: colEnd - colStart,
				Stride: M.Stride,
				Data: M.Data[colStart * M.Stride:colEnd * M.Stride],
			}
		}
	}
	return genericDenseSubmatrix(M, rowStart, rowEnd, colStart, colEnd)
}

func (M *DenseBinaryMatrix) Density(row, col int) float64 {
	// xxx optimize
	return genericDensity(M, row, col)
}

func (M *DenseBinaryMatrix) Equal(other BinaryMatrix) bool {
	return genericEqual(M, other)
}

func (M *DenseBinaryMatrix) index(i, j int) (word, bit int) {
	if columnFirst {
		word = j * M.Stride + i / wordSize
		bit = i % wordSize
	} else {
		word = i * M.Stride + j / wordSize
		bit = j % wordSize
	}
	return
}

func (M *DenseBinaryMatrix) Get(i, j int) uint8 {
	word, bit := M.index(i, j)
	return uint8((M.Data[word] >> bit) & 1)
}

func (M *DenseBinaryMatrix) GetRows() []BinaryVector {
	return genericGetRows(M)
}

func (M *DenseBinaryMatrix) IsZero() bool {
	return genericIsZero(M)
}

func (M *DenseBinaryMatrix) MaxColumnSupport(col int) int {
	for k := M.Stride - 1; k >= 0; k-- {
		c := M.Data[col * M.Stride + k]
		if c != 0 {
			for j := wordSize - 1; j >= 0; j-- {
				if (c >> j) & 1 != 0 {
					return k * wordSize + j
				}
			}
		}
	}
	return -1
}

func (M *DenseBinaryMatrix) MultiplyLeft(B BinaryMatrix) BinaryMatrix {
	return B.MultiplyRight(M)
}

func (M *DenseBinaryMatrix) MultiplyRight(B BinaryMatrix) (product BinaryMatrix) {
	Bsparse, ok := B.(*Sparse)
	if columnFirst && ok {
		// optimized case: assume M is column first. each column of
		// the product is a linear combination of the columns of M
		// selected by the support of corresponding column of B. this
		// sum can be computed cumulatively with fastxor. and if B is
		// a Sparse matrix, we have immediate access to the support of
		// each column.
		P := NewDenseBinaryMatrix(M.NumRows(), B.NumColumns())
		for j := 0; j < P.Cols; j++ {
			Pword, _ := P.index(0, j)
			support := Bsparse.Support(j)
			for _, k := range support {
				Mword, _ := M.index(0, k)
				fastxor.Bytes(P.Data[Pword:Pword + P.Stride],
					P.Data[Pword:Pword + P.Stride],
					M.Data[Mword:Mword + M.Stride])
			}
		}
		product = P
	} else {
		product = NewDenseBinaryMatrix(M.NumRows(), B.NumColumns())
		genericMultiply(M, B, product)
	}
	return product
}

func (M *DenseBinaryMatrix) NumColumns() int {
	return M.Cols
}

func (M *DenseBinaryMatrix) NumRows() int {
	return M.Rows
}

func (M *DenseBinaryMatrix) Overwrite(row int, col int, N BinaryMatrix) {
	// xxx optimize
	genericOverwrite(M, row, col, N)
}

func (M *DenseBinaryMatrix) Project(rowPredicate func(int) bool) BinaryMatrix {
	return genericProject(M, rowPredicate)
}

func (M *DenseBinaryMatrix) RowIsZero(index int) bool {
	return genericRowIsZero(M, index)
}

func (M *DenseBinaryMatrix) RowVector(index int) BinaryVector {
	return genericRowVector(M, index)
}

// xxx optimization idea: first scan down in 32 byte chunks, doing
// fastxor against all zeros. when we find a nonzero chunk, switch to
// existing byte-by-byte scanning.
func (M *DenseBinaryMatrix) ScanDownOld(row, col int) int {
	if !columnFirst {
		return genericScanDown(M, row, col)
	}
// 	log.Printf("xxx M=%v\n%s", M, DumpMatrix(M))
	rows := M.NumRows()
	res := -1
// 	log.Printf("xxx row=%d col=%d", row, col)
	word, bit := M.index(row, col)
	for i := row; res < 0 && i < rows; {
		for i < rows && bit < wordSize {
// 			log.Printf("xxx i=%d word=%d bit=%d", i, word, bit)
			if (M.Data[word] >> bit) & 1 != 0 {
				res = i
// 				log.Printf("xxx res=%d", res)
				break
			}
			bit++
			i++
		}
		word++
		bit = 0
	}
	return res
}

// xxx benchmarking new approach - precomputed masks instead of arithmetic
func (M *DenseBinaryMatrix) ScanDown(row, col int) int {
	if !columnFirst {
		return genericScanDown(M, row, col)
	}
	// log.Printf("xxx M=%v\n%s", M, DumpMatrix(M))
	rows := M.NumRows()
	res := -1
	// go byte-by-byte using literal masks
	word, margin := M.index(row, col)
	// log.Printf("xxx row=%d col=%d word=%d margin=%d", row, col, word, margin)
	i := row
	for i < rows && res < 0 {
		byte := M.Data[word]
// 		log.Printf("xxx i=%d byte=%08b", i, byte)
		// on the first byte we inspect, we may need to ignore some
		// leading bits, which we call the margin.  zero them out.
		if margin != 0 {
			switch margin {
			case 1: byte &= 0b11111110
			case 2: byte &= 0b11111100
			case 3: byte &= 0b11111000
			case 4: byte &= 0b11110000
			case 5: byte &= 0b11100000
			case 6: byte &= 0b11000000
			case 7: byte &= 0b10000000
			}
			i -= margin
			margin = 0
		}
// 		log.Printf("xxx i=%d byte=%08b (margin cleared) ", i, byte)
		// xxx possibly skip this byte if it's all zeros
		if byte != 0 {
			if byte & 0b00000001 != 0 {
				res = i
				break
			} else if byte & 0b00000010 != 0 {
				res = i + 1
				break
			} else if byte & 0b00000100 != 0 {
				res = i + 2
				break
			} else if byte & 0b00001000 != 0 {
				res = i + 3
				break
			} else if byte & 0b00010000 != 0 {
				res = i + 4
				break
			} else if byte & 0b00100000 != 0 {
				res = i + 5
				break
			} else if byte & 0b01000000 != 0 {
				res = i + 6
				break
			} else if byte & 0b10000000 != 0 {
				res = i + 7
				break
			}
		}
		i += 8
		word++
	}
	return res
}

func (M *DenseBinaryMatrix) ScanRight(row, col int) int {
	return genericScanRight(M, row, col)
}

func (M *DenseBinaryMatrix) Set(i, j int, b uint8) {
	word, bit := M.index(i, j)
	if b == 0 {
		M.Data[word] &= ^(1 << bit)
	} else {
		M.Data[word] |= 1 << bit
	}
}

// xxx old
// func (M *DenseBinaryMatrix) Sparse() BinaryMatrix {
// 	S := NewSparseBinaryMatrix(M.Rows, M.Cols)
// 	if !columnFirst {
// 		S.Overwrite(0, 0, M)
// 		return S
// 	}
// 	// log.Printf("xxx stride=%d", M.Stride)
// 	for j := 0; j < M.Cols; j++ {
// 		M.copyColumnZeroedDirect(S, j)
// 	}
// 	return S
// }

func (M *DenseBinaryMatrix) Sparse() *Sparse {
	return M.SparseParallel().(*Sparse)
}

func (M *DenseBinaryMatrix) SparseParallel() BinaryMatrix {
	S := NewSparseBinaryMatrix(M.Rows, M.Cols)
	if !columnFirst {
		S.Overwrite(0, 0, M)
		return S
	}
	verbose := false
	colWorkGroup := NewWorkGroup(verbose)
	batch := make([]Work, 0)
	for j := 0; j < M.Cols; j++ {
		batch = append(batch, &copyColumnZeroedWork{M, S, j})
	}
	colWorkGroup.ProcessBatch(batch)
	return S
}

type copyColumnZeroedWork struct {
	M *DenseBinaryMatrix
	dest *Sparse
	col int
}

func (w *copyColumnZeroedWork) Do() {
	w.M.copyColumnZeroedDirect(w.dest, w.col)
}

func (M *DenseBinaryMatrix) copyColumnZeroedDirect(dest *Sparse, j int) {
	row := 0
	ints := make(orderedIntSet, 0)
	for k := 0; k < M.Stride; k++ {
		word := M.Data[j * M.Stride + k]
		if word == 0 {
			row += wordSize
			continue
		}
		for i := 0; i < wordSize; i++ {
			if word & (1 << i) != 0 {
				ints = append(ints, row)
			}
			row++
			if row >= M.Rows {
				break
			}
		}
		if row >= M.Rows {
			break
		}
	}
	dest.SetColumnData(j, ints)
}

func (M *DenseBinaryMatrix) String() string {
	return genericString(M)
}

func (M *DenseBinaryMatrix) Submatrix(rowStart, rowEnd, colStart, colEnd int) BinaryMatrix {
	return M.DenseSubmatrix(rowStart, rowEnd, colStart, colEnd)
}

func (M *DenseBinaryMatrix) SwapColumns(i, j int) {
	iWord, iBit := M.index(0, i)
	if iBit != 0 {
		panic("misaligned")
	}
	jWord, jBit := M.index(0, j)
	if jBit != 0 {
		panic("misaligned")
	}
	copy(M.buf, M.Data[iWord:iWord+M.Stride])
	copy(M.Data[iWord:iWord+M.Stride], M.Data[jWord:jWord+M.Stride])
	copy(M.Data[jWord:jWord+M.Stride], M.buf)
}

func (M *DenseBinaryMatrix) SwapRows(i, j int) {
	genericSwapRows(M, i, j)
}

func (M *DenseBinaryMatrix) Transpose() BinaryMatrix {
	// xxx optimized version; faster to produce sparse matrix
	return genericTranspose(M)
}
