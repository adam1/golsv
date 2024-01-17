package golsv

import (
	"encoding/gob"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// Reducer computes the Smith normal form of a matrix.
// We think of the input matrix M as being factorized as
//
//   M = P D Q
//
// where
//
//   P is an automorphism of the codomain of M,
//   D is a diagonal matrix, and
//   Q is an automorphism of the domain of M.
//
// The row and column operations performed are returned.  The row
// operations can be turned into an invertible matrix operating on the
// codomain of the original matrix.  The column operations can be
// turned into an invertible matrix operating on the domain of the
// original matrix.  These conversions are not automatically done here
// since they can be expensive. The matrix M may be modified in place,
// but is not guaranteed to be equal to D.
type Reducer interface {
	Reduce(M BinaryMatrix) (D BinaryMatrix, rowOps, colOps []Operation)
}

type DiagonalReducer struct {
	matrix  BinaryMatrix
	numWorkers int
	colOps  []Operation
	rowOps  []Operation
	colOpsMatrix BinaryMatrix
	coimageBasis BinaryMatrix
	kernelBasis BinaryMatrix
	reduced bool
	verbose bool
	statIntervalSteps int
	statColumnAdds int
	switchToDensePredicate func(remaining int, subdensity float64) bool
	// xxx old worker group
	workers []*worker
	workerWaitGroup sync.WaitGroup
	WriteIntermediateFile bool
	// xxx new worker group; merge these eventually
	colWorkGroup *WorkGroup
}

func NewDiagonalReducer(verbose bool) *DiagonalReducer {
	R := &DiagonalReducer{
		colOps: make([]Operation, 0),
		reduced: false,
		statIntervalSteps: 1000,
		switchToDensePredicate: defaultSwitchToDensePredicate,
		verbose: verbose,
	}
	gob.Register(&AddOp{})
	gob.Register(&SwapOp{})
	return R
}

func (R *DiagonalReducer) Reduce(M BinaryMatrix) (D BinaryMatrix, rowOps, colOps []Operation) {
	showSteps := false
	R.matrix = M
	rows := M.NumRows()
	cols := M.NumColumns()
	d := rows
	if cols < d {
		d = cols
	}
	R.setupWorkers()

	profiling := false
	profilingExitAt := 30000
	startTime := time.Now()
	lastStatTime := startTime

	if showSteps {
		log.Printf("initial matrix: %v\n%s", R.matrix, dumpMatrix(R.matrix))
	}
	for i := 0; i < d; i++ {
		subdensity := -1.0
		if i > 0 && i % R.statIntervalSteps == 0 {
			// if we are not dense, measure density and consider
			// switching to dense for the remainder
			_, ok := R.matrix.(*DenseBinaryMatrix)
			if !ok {
				subdensity = R.matrix.Density(i, i)
				doSwitch := R.switchToDensePredicate(d - i, subdensity)
				if doSwitch {
					R.reduceDenseSubmatrix(i)
					break
				}
			}
			if R.verbose  {
				// xxx get rid of early exit stuff
				if profiling && i > 0 && i % profilingExitAt == 0 {
					panic("exiting early for profiling")
				}
				now := time.Now()
				elapsed := now.Sub(lastStatTime)
				lastStatTime = now
				rate := float64(R.statIntervalSteps) / elapsed.Seconds()
				estimatedHoursRemaining := float64(d - i) / rate / 3600.0
				totalElapsed := now.Sub(startTime)
				totalRate := float64(i) / totalElapsed.Seconds()
				msg := fmt.Sprintf("reducing; i=%d coladd=%d trowop=%d tcolop=%d rate=%1.3f trate=%1.3f ehr=%1.2f",
					i, R.statColumnAdds, len(R.rowOps), len(R.colOps), rate, totalRate, estimatedHoursRemaining)
				if subdensity >=0 {
					msg += fmt.Sprintf(" subden=%1.8f", subdensity)
				}
				log.Println(msg)
				R.statColumnAdds = 0
			}
		}
		found := R.setPivotScanDown(i)
		if !found { // no more pivots
			break
		}
		R.clearRowParallel(i)
		R.clearColumnAfterRow(i)
		if showSteps {
			log.Printf("after %d: %v\n%s", i, R.matrix, dumpMatrix(R.matrix))
		}
	}
	R.reduced = true
	if R.verbose {
		log.Printf("done reducing; trowop=%d tcolop=%d", len(R.rowOps), len(R.colOps))
	}
	return R.matrix, R.rowOps, R.colOps
}

func (R *DiagonalReducer) clearColumnAfterRow(d int) {
	// we assume that the row has already been cleared, hence the
	// current row (d) has a 1 in column d and is zero elsewhere.  we
	// clear the column by adding the current row to rows below, and
	// since current row only affects column d, the affect of adding
	// the row is to simply set column d to zero.  this is faster than
	// actually adding the rows together.  of course, we must still
	// account for the operation as a row addition.
	k := d + 1
	for {
		k = R.matrix.ScanDown(k, d)
		if k == -1 {
			break
		}
		R.matrix.Set(k, d, 0)
		R.rowOps = append(R.rowOps, AddOp{d, k})
		k++
	}
}

func defaultSwitchToDensePredicate(remaining int, subdensity float64) bool {
	// heuristics chosen by basic benchmarking; not necessarily
	// optimal.
	return subdensity >= 0.003 && remaining >= 10000
}

// xxx deprecate
func (R *DiagonalReducer) computeKernelBasis() {
	cols := R.matrix.NumColumns()

	M := NewSparseBinaryMatrixIdentity(cols)
	ColumnOperationsMatrix(M, R.colOps, R.verbose)
	R.colOpsMatrix = M
	splitCol := firstZeroCol(R.matrix)
	if R.verbose {
		log.Printf("splitting column ops matrix at column %d", splitCol)
	}
	R.kernelBasis = R.colOpsMatrix.Submatrix(0, R.colOpsMatrix.NumRows(), splitCol, cols)
	R.coimageBasis = R.colOpsMatrix.Submatrix(0, R.colOpsMatrix.NumRows(), 0, splitCol)
	if R.verbose {
		log.Printf("done computing kernel basis")
	}
	return
}

func (R *DiagonalReducer) CoimageBasis() BinaryMatrix {
	if !R.reduced {
		panic("CoimageBasis called before Reduce")
	}
	return R.coimageBasis
}

func (R *DiagonalReducer) reduceDenseSubmatrix(index int) {
	if R.verbose {
		log.Printf("switching to dense at diagonal %d", index)
	}
	Rd := NewDiagonalReducer(R.verbose)
	if R.verbose {
		log.Printf("computing dense submatrix")
	}
	submatrix := R.matrix.DenseSubmatrix(index, R.matrix.NumRows(), index, R.matrix.NumColumns())
	if R.verbose {
		log.Printf("reducing dense submatrix=%v", submatrix)
	}
	Rd.Reduce(submatrix)
	subrank := firstZeroCol(submatrix)
	if R.verbose {
		log.Printf("finished reducing dense submatrix; submatrix=%v subrank=%d", submatrix, subrank)
	}
// 	if true { // xxx sanity check?
// 		if !binaryMatrixIsSmithNormalForm(submatrix) {
// 			panic("submatrix is not in smith normal form")
// 		}
// 	}
	R.matrix = NewSparseBinaryMatrixDiagonal(
		R.matrix.NumRows(), R.matrix.NumColumns(), index + subrank)
	for _, op := range Rd.colOps {
		op = op.Shift(index)
		R.colOps = append(R.colOps, op)
	}
	for _, op := range Rd.rowOps {
		op = op.Shift(index)
		R.rowOps = append(R.rowOps, op)
	}
}

func (R *DiagonalReducer) setPivotScanDown(i int) (found bool) {
	cols := R.matrix.NumColumns()
	// scan down, col by col, to find a 1, then do the necessary
	// permutations to move it to (i,i).
	r, c := -1, -1
    for j := i; j < cols; j++ {
		k := R.matrix.ScanDown(i, j)
		if k != -1 {
			r, c = k, j
			break
		}
	}
// 	if R.verbose {
// 		log.Printf("xxx found pivot at %d,%d", r, c)
// 	}
	if r == -1 {
		return false
	}
	if r != i {
		R.applyRowOp(SwapOp{i, r})
	}
	if c != i {
		R.applyColumnOp(SwapOp{i, c})
	}
	return true
}

type worker struct {
	index int
	group *sync.WaitGroup
	matrix BinaryMatrix
	todo chan any
	workRow int
	workColStart int
	workColEnd int
	resultOps []Operation
}

func New_worker(index int, group *sync.WaitGroup, matrix BinaryMatrix) *worker {
	w := &worker{
		index: index,
		group: group,
		matrix: matrix,
		todo: make(chan any),
		resultOps: make([]Operation, 0),
	}
	go w.run()
	return w
}

func (w *worker) run() {
	for {
		<- w.todo
		w.doWork()
	}
}

func (w *worker) doWork() {
	// log.Printf("xxx worker-%d: doWork", w.index)
	w.resultOps = w.resultOps[:0]
	for j := w.workColStart; j < w.workColEnd; j++ {
		if w.matrix.Get(w.workRow, j) == 1 {
			w.matrix.AddColumn(w.workRow, j)
			w.resultOps = append(w.resultOps, AddOp{w.workRow, j})
		}
	}
	w.group.Done()
}

func (w *worker) setWork(row, colStart, colEnd int) {
	// log.Printf("xxx worker-%d: row=%d colStart=%d colEnd=%d", w.index, row, colStart, colEnd)
	w.workRow = row
	w.workColStart = colStart
	w.workColEnd = colEnd
	w.group.Add(1)
	w.todo <- nil
}

func (w *worker) waitResult() []Operation {
	w.group.Wait()
	return w.resultOps
}

func (R *DiagonalReducer) setupWorkers() {
	R.numWorkers = runtime.NumCPU()
	R.workers = make([]*worker, R.numWorkers)
	if R.verbose {
		log.Printf("spawning %d workers", R.numWorkers)
	}
	for i := 0; i < R.numWorkers; i++ {
		R.workers[i] = New_worker(i, &R.workerWaitGroup, R.matrix)
	}
}

func (R *DiagonalReducer) applyColumnOp(op Operation) {
	R.matrix.ApplyColumnOperation(op)
	R.colOps = append(R.colOps, op)
}

func (R *DiagonalReducer) applyRowOp(op Operation) {
// 	if R.verbose {
// 		log.Printf("xxx applyRowOp: op=%v", op)
// 	}
	R.matrix.ApplyRowOperation(op)
	R.rowOps = append(R.rowOps, op)
// 	if R.verbose {
// 		log.Printf("xxx applyRowOp done")
// 	}
}

// xxx factoring out WorkGroup; here a column is a unit of work.  in
// the other case, if we restrict parallelization to a series of
// AddOps all with the same source column, a unit of work is also a
// column.  could be called ColumnWorkers.
func (R *DiagonalReducer) clearRowParallel(d int) {
	cols := R.matrix.NumColumns()
	colsRemaining := cols - d - 1
	colsPerWorker := colsRemaining / R.numWorkers
	if colsPerWorker == 0 {
		colsPerWorker = 1
	}
	// send
	started := 0
	for w := 0; w < R.numWorkers; w++ {
		start := d + 1 + (w * colsPerWorker)
		if start >= cols {
			break
		}
		started++
		end := start + colsPerWorker
		if w == R.numWorkers - 1 {
			end = cols
		}
		R.workers[w].setWork(d, start, end)
	}
	// sweep
	for w := 0; w < started; w++ {
		// xxx this is calling Wait too many times... might be harmless?
		// is there a guard that ensures that after this "waiting" that
		// the columns are actually done?
		colOps := R.workers[w].waitResult()
		R.colOps = append(R.colOps, colOps...)
		R.statColumnAdds += len(colOps)
	}
}

// xxx deprecated
// func (R *DiagonalReducer) writeColOpsFile() {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Printf("caught and resuming from panic: %v", r)
// 		}
// 	}()
// 	log.Printf("writing colOps.gob")
// 	WriteGobFile("colOps.gob", R.colOps)
// 	log.Printf("done writing colOps.gob")
// }

func firstZeroCol(M BinaryMatrix) int {
	cols := M.NumColumns()
	zeroCol := cols
	for j := 0; j < cols; j++ {
		if M.ColumnIsZero(j) {
			zeroCol = j
			// log.Printf("xxx found zero col %d", zeroCol)
			break
		}
	}
	return zeroCol
}
