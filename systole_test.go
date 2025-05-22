package golsv

import (
	"log"
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
		ctor                       func() (d1, d2 BinaryMatrix)
		wantSystole, wantCosystole int
	}{
		{emptyTriangle, 3, 1},
		{filledTriangle, 0, 0},
	}
	for i, test := range tests {
		verbose := false
		d1, d2 := test.ctor()
		U, B, _, _, _, _ := UBDecomposition(d1, d2, verbose)
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
		U, B, _, _, _, _ = UBDecomposition(delta1, delta0, verbose)
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
		U, B, _, _, _, _ := UBDecomposition(d1, d2, verbose)
		gotSystole := SystoleExhaustiveSearch(U, B, verbose)
		if gotSystole != i {
			t.Errorf("systole search for cyclic graph %d got=%d want=%d", i, gotSystole, i)
		}
		delta0 := d1.Transpose().Dense()
		delta1 := d2.Transpose().Dense()
		U, B, _, _, _, _ = UBDecomposition(delta1, delta0, verbose)
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

func TestSystoleTorus(t *testing.T) {
	verbose := false
	T := NewZComplexFromMaximalSimplices([][]int{
		{0, 3, 4}, {3, 6, 7}, {0, 6, 7},
		{0, 1, 4}, {3, 4, 7}, {0, 1, 7},
		{1, 4, 5}, {4, 7, 8}, {1, 7, 8},
		{1, 2, 5}, {4, 5, 8}, {1, 2, 8},
		{2, 5, 3}, {5, 6, 8}, {2, 6, 8},
		{0, 2, 3}, {3, 5, 6}, {0, 2, 6},
	})
	visualize := false
	if visualize {
		WriteComplexGraphvizFile(T, "torus.dot", verbose)
	}
	D1, D2 := T.D1(), T.D2()
	systole, dimZ1, dimB1, dimH1 := ComputeFirstSystole(D1, D2, verbose)
	wantSystole := 3
	if systole != wantSystole {
		t.Errorf("systole=%d want=%d", systole, wantSystole)
	}
	wantDimZ1 := 19
	if dimZ1 != wantDimZ1 {
		t.Errorf("dimZ1=%d want=%d", dimZ1, wantDimZ1)
	}
	wantDimB1 := 17
	if dimB1 != wantDimB1 {
		t.Errorf("dimB1=%d want=%d", dimB1, wantDimB1)
	}
	wantDimH1 := 2
	if dimH1 != wantDimH1 {
		t.Errorf("dimH1=%d want=%d", dimH1, wantDimH1)
	}
	cosystole := ComputeFirstCosystole(D1, D2, verbose)
	if verbose {
		log.Printf("systole=%d cosystole=%d", systole, cosystole)
	}
}

func TestSystoleKernelBasis(t *testing.T) {
	tests := []struct {
		M   BinaryMatrix
		Exp BinaryMatrix
	}{
		{
			NewDenseBinaryMatrixFromString(
				`1 0 1
			 0 1 1
			 0 0 0`),
			NewDenseBinaryMatrixFromString(
				`1
				 1
				 1`),
		},
		{
			NewDenseBinaryMatrix(3, 0),
			NewDenseBinaryMatrix(0, 0),
		},
	}
	verbose := false
	for i, test := range tests {
		K := kernelBasis(test.M, verbose)
		if !K.Equal(test.Exp) {
			t.Errorf("kernel basis [%d] got=%v:\n%swant=%v:\n%s", i, K, dumpMatrix(K), test.Exp, dumpMatrix(test.Exp))
		}
	}
}

func SheetWithTwoHoles() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{
		{0, 3, 4}, {0, 6, 7}, {0, 1, 8},
		{1, 2, 3}, {4, 5, 6},
	})
}

func TestSystolePlanarTwoHoles(t *testing.T) {
	verbose := false
	X := SheetWithTwoHoles()
	D1, D2 := X.D1(), X.D2()
	systole, dimZ1, dimB1, dimH1 := ComputeFirstSystole(D1, D2, verbose)
	wantSystole := 3
	if systole != wantSystole {
		t.Errorf("systole=%d want=%d", systole, wantSystole)
	}
	wantDimZ1 := 7
	if dimZ1 != wantDimZ1 {
		t.Errorf("dimZ1=%d want=%d", dimZ1, wantDimZ1)
	}
	wantDimB1 := 5
	if dimB1 != wantDimB1 {
		t.Errorf("dimB1=%d want=%d", dimB1, wantDimB1)
	}
	wantDimH1 := 2
	if dimH1 != wantDimH1 {
		t.Errorf("dimH1=%d want=%d", dimH1, wantDimH1)
	}
	cosystole := ComputeFirstCosystole(D1, D2, verbose)
	if verbose {
		log.Printf("systole=%d cosystole=%d", systole, cosystole)
	}
}

// xxx first idea: like above, compute some known systoles using the simplicial method
//
// xxx construct example where scanning from single vertex gives wrong result
func TestSimplicialSystoleSearchSheetWithTwoHoles(t *testing.T) {
	verbose := false
	X := SheetWithTwoHoles()
	S := NewSimplicialSystoleSearch(X, verbose)
	gotSystole := S.Systole()
	wantSystole := 3
	if gotSystole != wantSystole {
		t.Errorf("systole: got=%d want=%d", gotSystole, wantSystole)
	}
}
