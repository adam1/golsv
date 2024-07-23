package golsv

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"strconv"
	"time"

	"golang.org/x/exp/mmap"
)

type Operation interface{
	Shift(n int) Operation
	String() string
}

type SwapOp struct {
	I, J int
}

func (op SwapOp) Shift(n int) Operation {
	return SwapOp{op.I + n, op.J + n}
}

func (op SwapOp) String() string {
	return fmt.Sprintf("S %d %d", op.I, op.J)
}

type AddOp struct {
	Source, Target int
}

func (op AddOp) Shift(n int) Operation {
	return AddOp{op.Source + n, op.Target + n}
}

func (op AddOp) String() string {
	return fmt.Sprintf("A %d %d", op.Source, op.Target)
}

func OperationFromString(s string) Operation {
	tokenizer := bufio.NewScanner(strings.NewReader(s))
	tokenizer.Split(bufio.ScanWords)
	opType := ""
	var a, b int
	var err error
	for i := 0; i < 3; i++ {
		// xxx optimize: should be able to do this with fewer allocations
		tokenizer.Scan()
		token := tokenizer.Text()
		if err := tokenizer.Err(); err != nil {
			panic(err)
		}
		switch i {
		case 0:
			if token != "S" && token != "A" {
				panic("unknown operation")
			}
			opType = token
		case 1:
			a, err = strconv.Atoi(token)
			if err != nil {
				panic(err)
			}
		case 2:
			b, err = strconv.Atoi(token)
			if err != nil {
				panic(err)
			}
		}
	}
	switch opType {
	case "S":
		return SwapOp{a, b}
	case "A":
		return AddOp{a, b}
	}
	panic("unknown operation")
}

func RowOperationMatrix(op Operation, n int) BinaryMatrix {
	matrix := NewDenseBinaryMatrixIdentity(n)
	switch op := op.(type) {
	case SwapOp:
		matrix.SwapRows(op.I, op.J)
	case AddOp:
		matrix.AddRow(op.Source, op.Target)
	default:
		panic("unknown row operation")
	}
	return matrix
}

func RowOperationsMatrix(ops []Operation, n int) BinaryMatrix {
	M := NewDenseBinaryMatrixIdentity(n)
	for _, op := range ops {
		M.ApplyRowOperation(op)
	}
	return M
}

func ColumnOperationMatrix(op Operation, n int) BinaryMatrix {
	matrix := NewDenseBinaryMatrixIdentity(n)
	switch op := op.(type) {
	case SwapOp:
		matrix.SwapColumns(op.I, op.J)
	case AddOp:
		matrix.AddColumn(op.Source, op.Target)
	default:
		panic("unknown column operation")
	}
	return matrix
}

func ColumnOperationsMatrix(ops []Operation, n int) BinaryMatrix {
	M := NewDenseBinaryMatrixIdentity(n)
	for _, op := range ops {
		M.ApplyColumnOperation(op)
	}
	return M
}

// new streamer

type OpsFileMatrixStreamer struct {
	reader OperationReader
	matrix BinaryMatrix
	verbose bool
}

func NewOpsFileMatrixStreamer(reader OperationReader, matrix BinaryMatrix, verbose bool) *OpsFileMatrixStreamer {
	return &OpsFileMatrixStreamer{
		reader: reader,
		matrix: matrix,
		verbose: verbose,
	}
}

func (S *OpsFileMatrixStreamer) Stream() {
	if S.verbose {
		log.Printf("streaming ops from %v", S.reader)
	}
	defer S.reader.Close()
	batcher := NewOpsBatchReader(S.reader, S.matrix)
	colWorkGroup := NewWorkGroup(S.verbose)
	statInterval := 1000
	statOps := 0
	statLastTime := time.Now()
	startTime := statLastTime
	i := 0
	for {
		batch, err := batcher.Read()
		colWorkGroup.ProcessBatch(batch)
		statOps += len(batch)
		if i % statInterval == 0 && i > 0 && S.verbose {
			now := time.Now()
			elapsed := now.Sub(statLastTime)
			rate := float64(statOps) / elapsed.Seconds()
			totalRate := float64(i) / now.Sub(startTime).Seconds()
			log.Printf("completed %d ops; crate=%1.1f trate=%1.1f", i, rate, totalRate)
			statLastTime = now
			statOps = 0
			if elapsed.Seconds() < 1 {
				// scale back the stat interval to prevent logging
				// from becoming too verbose and possibly a
				// bottleneck.
				statInterval *= 10
			}
		}
		i += len(batch)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
}

type OperationReader interface {
	Read() (Operation, error)
	Close() error
}

type OperationSliceReader struct {
	ops []Operation
}

func NewOperationSliceReader(ops []Operation) *OperationSliceReader {
	return &OperationSliceReader{ops}
}

func (R *OperationSliceReader) Read() (Operation, error) {
	if len(R.ops) == 0 {
		return nil, io.EOF
	}
	op := R.ops[0]
	R.ops = R.ops[1:]
	return op, nil
}

func (R *OperationSliceReader) Close() error {
	return nil
}

func (R *OperationSliceReader) String() string {
	return "OperationSliceReader"
}

type OperationFileReader struct {
	filename string
	file *os.File
	reader *bufio.Reader
	lineCount int
	statInterval int
	startTime time.Time
	lastStatTime time.Time
}

func OpenOperationFile(filename string) *OperationFileReader {
	R := &OperationFileReader{
		filename: filename,
		statInterval: 10*1000,
		startTime: time.Now(),
	}
	// log.Printf("opening %s", filename)
	R.lastStatTime = R.startTime
	var err error
	R.file, err = os.Open(filename)
	if err != nil {
		panic(err)
	}
	R.reader = bufio.NewReader(R.file)
	return R
}

func (R *OperationFileReader) Read() (Operation, error) {
	line, err := R.reader.ReadString('\n')
	eof := false
	if err != nil {
		if err == io.EOF {
			eof = true
		} else {
			panic(err)
		}
	}
	line = strings.TrimSpace(line)
	if line == "" {
		if eof {
			return nil, io.EOF
		}
		panic("empty line")
	}
	op := OperationFromString(line)
	R.lineCount++
	if R.lineCount % R.statInterval == 0 {
		now := time.Now()
		elapsed := now.Sub(R.lastStatTime)
		rate := float64(R.statInterval) / elapsed.Seconds()
		log.Printf("read %d lines; rate=%1.1f", R.lineCount, rate)
		if elapsed.Seconds() < 1 {
			R.statInterval *= 10
		} else if elapsed.Seconds() > 10 {
			R.statInterval = 1 + R.statInterval / 10
		}
		R.lastStatTime = now
	}
	if eof {
		return op, io.EOF
	}
	return op, err
}

func (R *OperationFileReader) Close() error {
	if R.file == nil {
		return nil
	}
	err := R.file.Close()
	R.file = nil
	return err
}

// basic version for testing
type OperationFileSimpleReverseReader struct {
	filename string
	file *os.File
	reader *bufio.Reader
	ops []Operation
	index int
}

func OpenOperationFileSimpleReverse(filename string) *OperationFileSimpleReverseReader {
	// log.Printf("opening with simple reverse reader %s", filename)
	R := &OperationFileSimpleReverseReader{
		filename: filename,
		ops: make([]Operation, 0),
	}
	var err error
	R.file, err = os.Open(filename)
	if err != nil {
		panic(err)
	}
	R.reader = bufio.NewReader(R.file)
	// for this simple version, read all the lines int a slice
	for {
		line, err := R.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		line = strings.TrimSpace(line)
		op := OperationFromString(line)
		R.ops = append(R.ops, op)
	}
	R.index = len(R.ops) - 1
	return R
}

func (R *OperationFileSimpleReverseReader) Read() (Operation, error) {
	if R.index < 0 {
		return nil, io.EOF
	}
	op := R.ops[R.index]
	R.index--
	return op, nil
}

func (R *OperationFileSimpleReverseReader) Close() error {
	if R.file == nil {
		return nil
	}
	err := R.file.Close()
	R.file = nil
	return err
}

// -----------------------------------------------------------------------
// fancy reverse reader

type OperationFileReverseReader struct {
	filename string
	mmap *mmap.ReaderAt
	length int
	index int
}

func OpenOperationFileReverse(filename string) *OperationFileReverseReader {
	R := &OperationFileReverseReader{
		filename: filename,
	}
	var err error
	R.mmap, err = mmap.Open(filename)
	if err != nil {
		panic(err)
	}
	R.length = R.mmap.Len()
	// log.Printf("mmap'd file=%s length=%d", R.filename, R.length)
	R.index = R.length - 1
	return R
}

func (R *OperationFileReverseReader) Read() (Operation, error) {
	var op Operation
	for {
		line, err := R.readLine()
		if line != "" {
			op = OperationFromString(line)
		}
		if err != nil {
			if err == io.EOF {
				return op, io.EOF
			}
			panic(err)
		}
		if op != nil {
			break
		}
	}
	return op, nil
}

func (R *OperationFileReverseReader) readLine() (string, error) {
	buf := make([]byte, 0)
	for {
		c := R.mmap.At(R.index)
		R.index--
		if c == '\n' {
			break
		}
		buf = append([]byte{c}, buf...) // prepend
		if R.index < 0 {
			return string(buf), io.EOF
		}
	}
	return string(buf), nil
}

func (R *OperationFileReverseReader) Close() error {
	if R.mmap != nil {
		err := R.mmap.Close()
		if err != nil {
			panic(err)
		}
		R.mmap = nil
	}
	return nil
}


// -----------------------------------------------------------------------

type WorkReader interface {
	Read() ([]Work, error)
}

type OpsBatchReader struct {
	opsReader OperationReader
	matrix BinaryMatrix
	underlyingEof bool
	remainder Operation
}

func NewOpsBatchReader(opsReader OperationReader, matrix BinaryMatrix) *OpsBatchReader {
	return &OpsBatchReader{
		opsReader: opsReader,
		matrix: matrix,
	}
}

func (B *OpsBatchReader) Read() ([]Work, error) {
	batch := make([]Work, 0)
	if B.remainder != nil {
		batch = append(batch, &OpsWork{B.remainder, B.matrix})
		B.remainder = nil
	}
	if B.underlyingEof {
		return batch, io.EOF
	}
	batchReady := false
	for !batchReady {
		op, err := B.opsReader.Read()
		if op != nil {
			if len(batch) == 0 {
				batch = append(batch, &OpsWork{op, B.matrix})
			} else {
				if B.safelyConcurrent(batch[0].(*OpsWork).op, op) {
					batch = append(batch, &OpsWork{op, B.matrix})
				} else {
					batchReady = true
					B.remainder = op
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				B.underlyingEof = true
				break
			}
			panic(err)
		}
	}
	return batch, nil
}

func (B *OpsBatchReader) safelyConcurrent(u, v Operation) bool {
	if add1, ok := u.(*AddOp); ok {
		if add2, ok := v.(*AddOp); ok {
			return add1.Source == add2.Source && add1.Target != add2.Target
		}
	}
	return false
}

// OpsBatcher groups operations into batches that can be safely
// parallelized.
type OpsBatcher struct {
	matrix BinaryMatrix
	ops []Operation
	p int
}

func NewOpsBatcher(matrix BinaryMatrix, ops []Operation) *OpsBatcher {
	return &OpsBatcher{matrix, ops, 0}
}

func (B *OpsBatcher) Next() []Work {
	if B.p >= len(B.ops) {
		return nil
	}
	batch := make([]Work, 0)
	base := B.ops[B.p]
	batch = append(batch, &OpsWork{base, B.matrix})
	B.p++
	for q := B.p; q < len(B.ops); q++ {
		if B.safelyConcurrent(base, B.ops[q]) {
			batch = append(batch, &OpsWork{B.ops[q], B.matrix})
		} else {
			break
		}
		B.p = q
	}
	return batch
}

func (B *OpsBatcher) safelyConcurrent(u, v Operation) bool {
	if add1, ok := u.(*AddOp); ok {
		if add2, ok := v.(*AddOp); ok {
			return add1.Source == add2.Source && add1.Target != add2.Target
		}
	}
	return false
}

type OpsWork struct {
	op Operation
	M BinaryMatrix
}

func (W *OpsWork) Do() {
	W.M.ApplyColumnOperation(W.op)
}

// WriteOperationsFile writes a text file in the following format:
//
//   A 3 4
//   S 1 2
//
// "A x y" means AddOperation with source x and target y.
// "S x y" means SwapOperation with indices x and y.
func WriteOperationsFile(filename string, ops []Operation) {
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
	for _, op := range ops {
		_, err := f.WriteString(op.String() + "\n")
		if err != nil {
			panic(err)
		}
	}
}

func ReadOperationsFile(filename string) []Operation {
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
	ops := make([]Operation, 0, 100)
	reader := bufio.NewReader(f)
	lines := 0
	statInterval := 10*1000
	startTime := time.Now()
	lastStatTime := startTime
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		ops = append(ops, OperationFromString(line))
		lines++
		if lines % statInterval == 0 {
			now := time.Now()
			elapsed := now.Sub(lastStatTime)
			rate := float64(statInterval) / elapsed.Seconds()
			log.Printf("read %d lines; rate=%1.1f", lines, rate)
			if elapsed.Seconds() < 1 {
				statInterval *= 10
			} else if elapsed.Seconds() > 10 {
				statInterval = 1 + statInterval / 10
			}
			lastStatTime = now
		}
	}
	return ops
}

type OperationWriter interface {
	Write(op Operation) error
	Close() error
	Count() int
}

type OperationSliceWriter struct {
	ops []Operation
}

func NewOperationSliceWriter() *OperationSliceWriter {
	return &OperationSliceWriter{make([]Operation, 0)}
}

func (W *OperationSliceWriter) Write(op Operation) error {
	W.ops = append(W.ops, op)
	return nil
}

func (W *OperationSliceWriter) Close() error {
	return nil
}

func (W *OperationSliceWriter) Count() int {
	return len(W.ops)
}

func (W *OperationSliceWriter) Slice() []Operation {
	return W.ops
}


type OperationFileWriter struct {
	filename string
	file *os.File
	writer *bufio.Writer
	count int
}

func OpenOperationFileWriter(filename string) *OperationFileWriter {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	return &OperationFileWriter{filename, file, bufio.NewWriter(file), 0}
}

func (W *OperationFileWriter) Write(op Operation) error {
	_, err := W.writer.WriteString(op.String() + "\n")
	if err != nil {
		return err
	}
	W.count++
	return nil
}

func (W *OperationFileWriter) Close() error {
	err := W.writer.Flush()
	if err != nil {
		return err
	}
	err = W.file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (W *OperationFileWriter) Count() int {
	return W.count
}


type OperationNilWriter struct {
	count int
}

func NewOperationNilWriter() *OperationNilWriter {
	return &OperationNilWriter{0}
}

func (W *OperationNilWriter) Write(op Operation) error {
	W.count++
	return nil
}

func (W *OperationNilWriter) Close() error {
	return nil
}

func (W *OperationNilWriter) Count() int {
	return W.count
}
