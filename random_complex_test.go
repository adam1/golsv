package golsv

import (
	"math/rand"
	"testing"
)


func TestRandomComplex(t *testing.T) {
	trials := 10
	minDimC_0 := 1
	maxDimC_0 := 10
	for i := 0; i < trials; i++ {
		dimC_0 := rand.Intn(maxDimC_0 - minDimC_0) + minDimC_0
		R := NewRandomComplexGenerator(dimC_0, false)
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
