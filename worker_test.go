package golsv

import (
	//	"log"
	"testing"
)

type testWork struct {
	index int
	record []int
}

func (w *testWork) Do() {
	w.record[w.index] = 1
}

func TestWorkGroup(t *testing.T) {
	n := 100
	G := NewWorkGroup(false)
	work := make([]Work, 0)
	record := make([]int, n)
	for i := 0; i < n; i++ {
		work = append(work, &testWork{i, record})
	}
	G.ProcessBatch(work)
	for i := 0; i < n; i++ {
		if record[i] != 1 {
			t.Errorf("work %d not done", i)
		}
	}
}
