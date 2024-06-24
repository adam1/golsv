package golsv

import (
	"math/rand"
	"testing"
)


func TestRandomComplex(t *testing.T) {
	trials := 10
	minDimC_0 := 1
	maxDimC_0 := 10
	verbose := false
	for i := 0; i < trials; i++ {
		dimC_0 := rand.Intn(maxDimC_0 - minDimC_0) + minDimC_0
		R := NewRandomComplexGenerator(dimC_0, verbose)
		d_1, d_2, err := R.RandomComplex()
		if err != nil {
			t.Error("wanted no error, got ", err)
		}
		if d_1.NumRows() != dimC_0 {
			t.Error("wanted d_1.NumRows() == n, got ", d_1.NumRows())
		}
		if d_1.NumColumns() < 1 {
			t.Error("wanted d_1.NumColumns() > 0, got ", d_1.NumColumns())
		}
		if d_2.NumRows() < 1 {
			t.Error("wanted d_2.NumRows() > 0, got ", d_2.NumRows())
		}
		if d_2.NumColumns() < 1 {
			t.Error("wanted d_2.NumColumns() > 0, got ", d_2.NumColumns())
		}
	}
}

func TestRandomSimplicialComplex(t *testing.T) {
	trials := 10
	minDimC_0 := 1
	maxDimC_0 := 20
	verbose := false
	for i := 0; i < trials; i++ {
		dimC_0 := rand.Intn(maxDimC_0 - minDimC_0) + minDimC_0
		R := NewRandomComplexGenerator(dimC_0, verbose)
		d_1, d_2, err := R.RandomSimplicialComplex()
		if err != nil {
			t.Error("wanted no error, got ", err)
		}
		if d_1.NumRows() != dimC_0 {
			t.Error("wanted d_1.NumRows() == n, got ", d_1.NumRows())
		}
		// check column weights
		for j := 0; j < d_1.NumColumns(); j++ {
			weight := d_1.ColumnWeight(j)
			if weight != 2 {
				t.Errorf("wanted d_1.ColumnWeight(%d)=2 got=%d", j, weight)
			}
		}
		// check for duplicate columns
		seen := make(map[string]bool)
		cols := d_1.Columns()
		for _, col := range cols {
			key := col.String()
			if _, ok := seen[key]; ok {
				t.Errorf("duplicate column %s", key)
			}
			seen[key] = true
		}
		// check column weights
		for j := 0; j < d_2.NumColumns(); j++ {
			weight := d_2.ColumnWeight(j)
			if weight != 3 {
				t.Errorf("wanted d_2.ColumnWeight(%d)=3 got=%d", j, weight)
			}
		}
		// check for duplicate columns
		seen = make(map[string]bool)
		cols = d_2.Columns()
		for _, col := range cols {
			key := col.String()
			if _, ok := seen[key]; ok {
				t.Errorf("duplicate column %s", key)
			}
			seen[key] = true
		}
	}
}
