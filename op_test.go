package golsv

import (
	"io"
	"log"
	"math/rand"
	"reflect"
	"testing"
)

func TestReadWriteOperationsFile(t *testing.T) {
	N := 100
	dim := 100
	ops := randomOperations(N, dim)
	filename := "test.ops~"
	WriteOperationsFile(filename, ops)
	ops2 := ReadOperationsFile(filename)
	if !reflect.DeepEqual(ops, ops2) {
		t.Errorf("ops != ops2")
	}
}

func randomOperations(n, dim int) []Operation {
	ops := make([]Operation, n)
	lastAddSource := -1
	for i := 0; i < n; i++ {
		if rand.Intn(2) == 0 {
			var source int
			if rand.Intn(2) == 0 && lastAddSource >= 0 {
				source = lastAddSource
			} else {
				source = rand.Intn(dim)
				lastAddSource = source
			}			
			ops[i] = AddOp{Source: source, Target: rand.Intn(dim)}
		} else {
			ops[i] = SwapOp{I: rand.Intn(dim), J: rand.Intn(dim)}
		}
	}
	return ops
}

// xxx deprecated
// func TestColumnOperationsMatrixParallel(t *testing.T) {
// 	dim := 10
// 	numOps := 100
// 	ops := randomOperations(numOps, dim)
// 	verbose := false
// 	A := NewSparseBinaryMatrixIdentity(dim)
// 	B := A.Copy()

// 	ColumnOperationsMatrix(A, ops, verbose)
// 	ColumnOperationsMatrixParallel(B, ops, verbose)

// 	if !A.Equal(B) {
// 		t.Errorf("A != B")
// 	}
// }

func Benchmark_OperationFromString(b *testing.B) {
	dim := 1000*1000
	ops := randomOperations(b.N, dim)
	strs := make([]string, b.N)
	for i, op := range ops {
		strs[i] = op.String()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OperationFromString(strs[i])
	}
}

func TestReadOpsToColumnMatrix(t *testing.T) {
	trials := 1
	dim := 1000
	maxOps := 1000
	filename := "test.ops~"
	for i := 0; i < trials; i++ {
		numOps := 1 + rand.Intn(maxOps)
		ops := randomOperations(numOps, dim)
		WriteOperationsFile(filename, ops)
		ops = ReadOperationsFile(filename)
		M := ColumnOperationsMatrix(ops, dim)
		// log.Printf("xxx M: %s\n%s", M, dumpMatrix(M))

		N := NewDenseBinaryMatrixIdentity(dim)
		verbose := false
		reader := OpenOperationFile(filename)
		streamer := NewOpsFileMatrixStreamer(reader, N, verbose)
		streamer.Stream()
		// log.Printf("xxx N: %s\n%s", N, dumpMatrix(N))

		if !M.Equal(N) {
			t.Errorf("M != N")
		}
	}
}

func TestOperationFileForwardReverseReaders(t *testing.T) {
	N := 100
	dim := 100
	ops := randomOperations(N, dim)
	filename := "test.ops~"
	WriteOperationsFile(filename, ops)
	reverseReader := OpenOperationFileSimpleReverse(filename)
	ops2 := reverseOps(readOperationsFile(reverseReader))
	if !reflect.DeepEqual(ops, ops2) {
		t.Errorf("ops != ops2")
	}
}

func readOperationsFile(reader OperationReader) []Operation {
	defer reader.Close()
	ops := make([]Operation, 0)
	for {
		op, err := reader.Read()
		if op != nil {
			ops = append(ops, op)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
	}
	return ops
}

func reverseOps(in []Operation) []Operation {
	out := make([]Operation, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = in[len(in)-1-i]
	}
	return out
}

func TestOperationFileFancyReverseReader(t *testing.T) {
	trials := 10
	N := 100
	dim := 100
	for i := 0; i < trials; i++ {
		ops := randomOperations(N, dim)
		filename := "test.ops~"
		WriteOperationsFile(filename, ops)
		reverseReader := OpenOperationFileReverse(filename)
		ops2 := reverseOps(readOperationsFile(reverseReader))
		if !reflect.DeepEqual(ops, ops2) {
			log.Printf("xxx ops: %v", ops)
			log.Printf("xxx ops2: %v", ops2)
			t.Errorf("ops != ops2")
		}
	}
}

