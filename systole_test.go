package golsv

import (
	"testing"
)

func emptyTriangle() (d1, d2 BinaryMatrix) {
	return NewDenseBinaryMatrixFromString(
		`1 0 1
		 1 1 0
		 0 1 1`),
		NewDenseBinaryMatrix(3, 0)
}

func filledTriangle() (d1, d2 BinaryMatrix) {
	return NewDenseBinaryMatrixFromString(
		`1 0 1
		 1 1 0
		 0 1 1`),
		NewDenseBinaryMatrixFromString(
		`1
		 1
		 1`)
}

func TestSystoleSearchSmallExamples(t *testing.T) {
	tests := []struct {
		ctor func() (d1, d2 BinaryMatrix)
		wantSystole, wantCosystole int
	}{
		{emptyTriangle, 3, 1},
		{filledTriangle, 0, 0},
	}
	for i, test := range tests {
		verbose := false
		d1, d2 := test.ctor()
		U, B := UBDecomposition(d1, d2, verbose)
		gotSystoleRandom := SystoleRandomSearch(U, B, 100, verbose)
		if gotSystoleRandom != test.wantSystole {
			t.Errorf("systole random search [%d] got=%d want=%d", i, gotSystoleRandom, test.wantSystole)
		}
		gotSystoleExhaustive := SystoleExhaustiveSearch(U, B, verbose)
		if gotSystoleExhaustive != test.wantSystole {
			t.Errorf("systole exhaustive search [%d] got=%d want=%d", i, gotSystoleExhaustive, test.wantSystole)
		}
		delta0 := d1.Transpose().Dense()
		delta1 := d2.Transpose().Dense()
		U, B = UBDecomposition(delta1, delta0, verbose)
		gotCosystoleRandom := SystoleRandomSearch(U, B, 100, verbose)
		if gotCosystoleRandom != test.wantCosystole {
			t.Errorf("cosystole random search [%d] got=%d want=%d", i, gotCosystoleRandom, test.wantCosystole)
		}
		gotCosystoleExhaustive := SystoleExhaustiveSearch(U, B, verbose)
		if gotCosystoleExhaustive != test.wantCosystole {
			t.Errorf("cosystole exhaustive search [%d] got=%d want=%d", i, gotCosystoleExhaustive, test.wantCosystole)
		}
	}
}

func TestSystoleCyclicGraphs(t *testing.T) {
	verbose := false
	maxLength := 10
	for i := 3; i < maxLength; i++ {
		d1, d2 := cyclicGraph(i)
		U, B := UBDecomposition(d1, d2, verbose)
		gotSystole := SystoleExhaustiveSearch(U, B, verbose)
		if gotSystole != i {
			t.Errorf("systole search for cyclic graph %d got=%d want=%d", i, gotSystole, i)
		}
		delta0 := d1.Transpose().Dense()
		delta1 := d2.Transpose().Dense()
		U, B = UBDecomposition(delta1, delta0, verbose)
		gotCosystole := SystoleExhaustiveSearch(U, B, verbose)
		if gotCosystole != 1 {
			t.Errorf("cosystole search for cyclic graph %d got=%d want=1", i, gotCosystole)
		}
	}
}

func cyclicGraph(n int) (d1, d2 BinaryMatrix) {
	d1 = NewDenseBinaryMatrix(n, n)
	d2 = NewDenseBinaryMatrix(n, 0)
	for i := 0; i < n; i++ {
		d1.Set(i, i, 1)
		d1.Set((i+1)%n, i, 1)
	}
	return
}
