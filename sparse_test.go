package golsv

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)


func TestOrderedIntSet_SetUnset(t *testing.T) {
	SET, UNSET := 1, 2
    tests := []struct {
		set  orderedIntSet
		op   int
		val  int
		want orderedIntSet
	}{
		{orderedIntSet{}, SET, 0, orderedIntSet{0}},
		{orderedIntSet{1}, SET, 0, orderedIntSet{0,1}},
		{orderedIntSet{1}, SET, 2, orderedIntSet{1,2}},
		{orderedIntSet{1}, SET, 1, orderedIntSet{1}},
		{orderedIntSet{1,77}, SET, 77, orderedIntSet{1,77}},
		{orderedIntSet{1,77}, SET, 0, orderedIntSet{0,1,77}},
		{orderedIntSet{1,77}, SET, 14, orderedIntSet{1,14,77}},
		{orderedIntSet{1,77}, SET, 77, orderedIntSet{1,77}},
		{orderedIntSet{1,77}, SET, 78, orderedIntSet{1,77,78}},
		{orderedIntSet{}, UNSET, 0, orderedIntSet{}},
		{orderedIntSet{0}, UNSET, 0, orderedIntSet{}},
		{orderedIntSet{0,1}, UNSET, 0, orderedIntSet{1}},
		{orderedIntSet{0,1}, UNSET, 1, orderedIntSet{0}},
		{orderedIntSet{0,1,77}, UNSET, -1, orderedIntSet{0,1,77}},
		{orderedIntSet{0,1,77}, UNSET, 0, orderedIntSet{1,77}},
		{orderedIntSet{0,1,77}, UNSET, 1, orderedIntSet{0,77}},
		{orderedIntSet{0,1,77}, UNSET, 77, orderedIntSet{0,1}},
		{orderedIntSet{0,1,77}, UNSET, 78, orderedIntSet{0,1,77}},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			switch tt.op {
			case SET:
				tt.set.Set(tt.val)
			case UNSET:
				tt.set.Unset(tt.val)
			}
			if !reflect.DeepEqual(tt.set, tt.want) {
				t.Errorf("got %v, want %v", tt.set, tt.want)
			}
		})
	}
}

func TestOrderedIntSet_Next(t *testing.T) {
	tests := []struct {
		set  orderedIntSet
		m int
		want int
	}{
		{orderedIntSet{}, 0, -1},
		{orderedIntSet{0}, 0, 0},
		{orderedIntSet{0}, 1, -1},
		{orderedIntSet{0,1}, 0, 0},
		{orderedIntSet{0,1}, 1, 1},
		{orderedIntSet{0,2}, 1, 2},
		{orderedIntSet{0,2,77}, 1, 2},
		{orderedIntSet{0,2,77}, 3, 77},
		{orderedIntSet{0,2,77}, 77, 77},
		{orderedIntSet{0,2,77}, 78, -1},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.set.Next(tt.m)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrderedIntSet_MergeDrop(t *testing.T) {
	tests := []struct {
		set  orderedIntSet
		other orderedIntSet
		want orderedIntSet
	}{
		{orderedIntSet{}, orderedIntSet{}, orderedIntSet{}},
		{orderedIntSet{}, orderedIntSet{1}, orderedIntSet{1}},
		{orderedIntSet{1}, orderedIntSet{}, orderedIntSet{1}},
		{orderedIntSet{1}, orderedIntSet{1}, orderedIntSet{}},
		{orderedIntSet{1}, orderedIntSet{1,2}, orderedIntSet{2}},
		{orderedIntSet{0,7}, orderedIntSet{0,7}, orderedIntSet{}},
		{orderedIntSet{1,7}, orderedIntSet{0}, orderedIntSet{0,1,7}},
		{orderedIntSet{1,7}, orderedIntSet{0,1}, orderedIntSet{0,7}},
		{orderedIntSet{1,7}, orderedIntSet{0,9}, orderedIntSet{0,1,7,9}},
	}
	for i, tt := range tests {
		name := fmt.Sprintf("%d", i)
		t.Run(name, func(t *testing.T) {
			tt.set.MergeDrop(&tt.other)
			if !tt.set.Equal(&tt.want) {
				t.Errorf("test %s: got %v, want %v", name, tt.set, tt.want)
			}
		})
	}
}

func TestOrderedIntSet_Toggle(t *testing.T) {
	tests := []struct {
		set  orderedIntSet
		index int
		want orderedIntSet
	}{
		{orderedIntSet{}, 0, orderedIntSet{0}},
		{orderedIntSet{0}, 0, orderedIntSet{}},
		{orderedIntSet{0}, 1, orderedIntSet{0,1}},
		{orderedIntSet{0,1}, 0, orderedIntSet{1}},
		{orderedIntSet{0,1}, 1, orderedIntSet{0}},
		{orderedIntSet{0,1}, 2, orderedIntSet{0,1,2}},
	}
	for i, tt := range tests {
		name := fmt.Sprintf("%d", i)
		t.Run(name, func(t *testing.T) {
			tt.set.Toggle(tt.index)
			if !tt.set.Equal(&tt.want) {
				t.Errorf("test %s: got %v, want %v", name, tt.set, tt.want)
			}
		})
	}
}

func Benchmark_SparseAddColumn(b *testing.B) {
	rows := 1000
	cols := 1000
	density := 0.01
	secureRandom := false
	s := NewRandomSparseBinaryMatrix(rows, cols, density, secureRandom)
	source := rand.Intn(cols / 2)
	for i := 0; i < b.N; i++ {
		target := source + 1 + rand.Intn((cols - 1)  / 2)
		s.AddColumn(source, target)
	}
}

func TestSparse_IsSmithNormalForm(t *testing.T) {
	tests := []struct {
		s    *Sparse
		wantSmith bool
		wantRank int
	}{
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,0,0},
			{0,1,0},
			{0,0,1},
		}), true, 3},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,1,0},
			{0,1,0},
			{0,0,1},
		}), false, 0},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,0,0},
			{0,1,0},
			{1,0,1},
		}), false, 0},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,0,0,0},
			{0,1,0,0},
			{0,0,1,0},
		}), true, 3},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,0,0},
			{0,1,0},
			{0,0,1},
			{0,0,0},
		}), true, 3},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,1,0,0},
			{0,1,0,0},
			{0,0,1,0},
		}), false, 0},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,1,0},
			{0,1,0},
			{0,0,1},
			{0,0,0},
			}), false, 0},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{0,0},
			{0,0},
			{0,0},
			{0,0},
			{1,0},
			{0,1},
			}), false, 0},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{0,0,0},
			{0,0,1},
			{0,0,0},
			{0,0,0},
			{1,0,0},
			{0,1,0},
			}), false, 0},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,0,0,0},
			{0,1,0,0},
			{0,0,1,0},
			}), true, 3},
		{
			NewSparseBinaryMatrixFromRowInts([][]uint8{
			{1,0,0,0},
			{0,1,0,1},
			{0,0,1,0},
			}), false, 0},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			gotSmith, gotRank := tt.s.IsSmithNormalForm()
			if gotSmith != tt.wantSmith {
				t.Errorf("gotSmith %v, wantSmith %v", gotSmith, tt.wantSmith)
			}
			if gotRank != tt.wantRank {
				t.Errorf("gotRank %v, wantRank %v", gotRank, tt.wantRank)
			}
		})
	}
}

func TestSparse_ReadWrite(t *testing.T) {
	trials := 100
	maxSize := 100
	density := 0.02
	file := "test.txt~"
	secureRandom := false
	for i := 0; i < trials; i++ {
		rows := 1 + rand.Intn(maxSize)
		cols := 1 + rand.Intn(maxSize)
		M := NewRandomSparseBinaryMatrix(rows, cols, density, secureRandom)
		// log.Printf("xxx writing matrix %s\n%s", M, dumpMatrix(M))
		M.WriteFile(file)
		N := ReadSparseBinaryMatrixFile(file)
		if !M.Equal(N) {
			t.Errorf("M != N")
		}
	}
}

func TestSparse_DenseSubmatrix(t *testing.T) {
	trials := 100
	maxSize := 100
	secureRandom := false
	for i := 0; i < trials; i++ {
		rows := 1 + rand.Intn(maxSize)
		cols := 1 + rand.Intn(maxSize)
		M := NewRandomSparseBinaryMatrix(rows, cols, 0.02, secureRandom)
		N := M.DenseSubmatrix(0, M.NumRows(), 0, M.NumColumns())
		if !M.Equal(N) {
			t.Errorf("M != N")
		}
	}
}

func TestSparseAppendDenseColumn(t *testing.T) {
	trials := 100
	maxSize := 100
	secureRandom := false
	for i := 0; i < trials; i++ {
		rows := 1 + rand.Intn(maxSize)
		cols := 1 + rand.Intn(maxSize)
		M := NewRandomSparseBinaryMatrix(rows, cols, 0.02, secureRandom)
		Mbefore := M.Copy()
		v := NewRandomDenseBinaryMatrix(rows, 1)
		M.AppendColumn(v)
		if M.NumColumns() != cols+1 {
			t.Errorf("M.NumColumns() != cols+1")
		}
		sub := M.DenseSubmatrix(0, M.NumRows(), 0, cols)
		if !sub.Equal(Mbefore) {
			t.Errorf("!sub.Equal(Mbefore)")
		}
		col := M.DenseSubmatrix(0, M.NumRows(), cols, cols+1)
		if !col.Equal(v) {
			t.Errorf("!col.Equal(v)")
		}
	}
}

func TestSparseAppendSparseColumn(t *testing.T) {
	trials := 100
	maxSize := 100
	secureRandom := false
	for i := 0; i < trials; i++ {
		rows := 1 + rand.Intn(maxSize)
		cols := 1 + rand.Intn(maxSize)
		M := NewRandomSparseBinaryMatrix(rows, cols, 0.02, secureRandom)
		Mbefore := M.Copy()
		v := NewRandomSparseBinaryMatrix(rows, 1, 0.5, secureRandom)
		M.AppendColumn(v)
		if M.NumColumns() != cols+1 {
			t.Errorf("M.NumColumns() != cols+1")
		}
		sub := M.Submatrix(0, M.NumRows(), 0, cols)
		if !sub.Equal(Mbefore) {
			t.Errorf("!sub.Equal(Mbefore)")
		}
		col := M.Submatrix(0, M.NumRows(), cols, cols+1)
		if !col.Equal(v) {
			t.Errorf("!col.Equal(v)")
		}
	}
}
