package golsv

import (
	"log"
	"math/rand"
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

func torus() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{
		{0, 3, 4}, {3, 6, 7}, {0, 6, 7},
		{0, 1, 4}, {3, 4, 7}, {0, 1, 7},
		{1, 4, 5}, {4, 7, 8}, {1, 7, 8},
		{1, 2, 5}, {4, 5, 8}, {1, 2, 8},
		{2, 5, 3}, {5, 6, 8}, {2, 6, 8},
		{0, 2, 3}, {3, 5, 6}, {0, 2, 6},
	})
}

func TestSystoleEtAlParticular(t *testing.T) {
	tests := []struct {
		X *ZComplex[ZVertexInt]
		ExpectedSystole   int
		ExpectedDimZ1     int
		ExpectedDimB1     int
		ExpectedDimH1     int
		ExpectedCosystole int
	}{
		{torus(), 3, 19, 17, 2, 6},
		// ribbon
		{NewZComplexFromMaximalSimplices([][]int{{0,1,2},{1,2,3},{2,3,4},{3,4,5},{0,4,5},{0,1,5}}), 3, 7, 6, 1, 3},
		// thick ribbon
		{NewZComplexFromMaximalSimplices([][]int{{0,1,3},{1,3,4},{1,2,4},{2,4,5},{0,2,5},{0,3,5},
                                 			    {3,6,7},{3,4,7},{4,7,8},{4,5,8},{5,6,8},{3,5,6}}), 3, 13, 12, 1, 5},
	}
	for i, test := range tests {
		verbose := false
		D1, D2 := test.X.D1(), test.X.D2()
		systole, dimZ1, dimB1, dimH1 := ComputeFirstSystole(D1, D2, verbose)
		if systole != test.ExpectedSystole {
			t.Errorf("test %d: systole: got=%d expected=%d", i, systole, test.ExpectedSystole)
		}
		if dimZ1 != test.ExpectedDimZ1 {
			t.Errorf("test %d: dimZ1: got=%d expected=%d", i, dimZ1, test.ExpectedDimZ1)
		}
		if dimB1 != test.ExpectedDimB1 {
			t.Errorf("test %d: dimB1: got=%d expected=%d", i, dimB1, test.ExpectedDimB1)
		}
		if dimH1 != test.ExpectedDimH1 {
			t.Errorf("test %d: dimH1: got=%d expected=%d", i, dimH1, test.ExpectedDimH1)
		}
		cosystole := ComputeFirstCosystole(D1, D2, verbose)
		if cosystole != test.ExpectedCosystole {
			t.Errorf("test %d: cosystole: got=%d expected=%d", i, cosystole, test.ExpectedCosystole)
		}
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

func sheetWithTwoHoles() *ZComplex[ZVertexInt] {
	return NewZComplexFromMaximalSimplices([][]int{
		{0, 3, 4}, {0, 6, 7}, {0, 1, 8},
		{1, 2, 3}, {4, 5, 6},
	})
}

func TestSystolePlanarTwoHoles(t *testing.T) {
	verbose := false
	X := sheetWithTwoHoles()
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

func TestSimplicialSystoleVsExhaustiveSearchSpecificExamples(t *testing.T) {
	tests := []struct {
		X               *ZComplex[ZVertexInt]
		ExpectedSystoleSimplicialSearch int
		ExpectedSystoleExhaustiveSearch int
	}{
// 		{NewZComplexFromMaximalSimplices([][]int{{0, 1}, {0, 2}, {1, 2}}), 3, 3},
// 		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}), 0, 0},
// 		{sheetWithTwoHoles(), 3, 3},
// 		{torus(), 3, 3},
// 		{NewZComplexFromMaximalSimplices([][]int{{0}, {1}}), 0, 0},       // disconnected
// 		{NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {3}}), 0, 0}, // disconnected

		// An example where the simplicial systole algorithm
		// finds a cycle of length equal to the global systole plus
		// one.  The global systole is 4.
//   		{NewZComplexFromMaximalSimplices([][]int{{0, 3, 7}, {0, 6, 9}, {2, 6, 9}, {3, 5, 7}, {1, 2}, {2, 5}, {3, 4}, {4, 6}, {8, 9}}), 5, 4},

// 		// An example where the (nonlocal) simplicial systole
// 		// algorithm gives a result that is not within one of a lower
// 		// bound of the global systole.  This can happen
// 		// if the global systole is zero.
// 		{NewZComplexFromMaximalSimplices([][]int{{0, 2, 3}, {0, 2, 4}, {0, 2, 5}, {0, 2, 9}, {0, 2, 11}, {0, 3, 4}, {0, 3, 11}, {0, 4, 5}, {0, 5, 6}, {0, 5, 9}, {0, 6, 11}, {1, 3, 7}, {1, 3, 10}, {1, 5, 9}, {1, 5, 10}, {2, 3, 4}, {2, 3, 10}, {2, 3, 11}, {2, 4, 5}, {2, 4, 10}, {2, 5, 9}, {2, 5, 10}, {2, 10, 11}, {3, 4, 7}, {3, 4, 10}, {3, 10, 11}, {4, 5, 10}, {8, 10}}), 4, 0},
	}
	for i, test := range tests {
		stopNonzero := true
		verbose := false
		S := NewSimplicialSystoleSearch(test.X, stopNonzero, verbose)
		gotSystoleSimplicialSearch := S.Search()
		if gotSystoleSimplicialSearch != test.ExpectedSystoleSimplicialSearch {
			t.Errorf("test %d: gotSystoleSimplicialSearch=%d expected=%d", i, gotSystoleSimplicialSearch, test.ExpectedSystoleSimplicialSearch)
		}
		gotSystoleExhaustiveSearch, _, _, _ := ComputeFirstSystole(test.X.D1(), test.X.D2(), verbose)
		if gotSystoleExhaustiveSearch != test.ExpectedSystoleExhaustiveSearch {
			t.Errorf("test %d: gotSystoleExhaustiveSearch=%d expected=%d", i, gotSystoleExhaustiveSearch, test.ExpectedSystoleExhaustiveSearch)
		}
	}
}

func cyclicGraphComplex(n int) *ZComplex[ZVertexInt] {
	edges := make([][]int, n)
	for i := 0; i < n; i++ {
		edges[i] = []int{i, (i + 1) % n}
	}
	return NewZComplexFromMaximalSimplices(edges)
}

func TestSimplicialSystoleSearchCyclicGraphs(t *testing.T) {
	stopNonzero := true
	verbose := false
	maxLength := 10
	for i := 3; i < maxLength; i++ {
		X := cyclicGraphComplex(i)
		S := NewSimplicialSystoleSearch(X, stopNonzero, verbose)
		gotSystole := S.Search()
		if gotSystole != i {
			t.Errorf("test %d: got=%d expected=%d", i, gotSystole, i)
		}
	}
}

func TestSimplicialSystoleSearchAtVertexVsGlobal(t *testing.T) {
	stopNonzero := true
	verbose := false
	// in this example, starting at vertex 0 finds a systole of 4,
	// whereas starting at 2 finds a systole of 3.
	X := NewZComplexFromMaximalSimplices([][]int{{0, 1, 4}, {1, 2, 5}, {2, 3, 6}, {0, 3, 7}, {2, 6, 8}, {2, 5, 9}, {5, 8, 9}})
	{
		S := NewSimplicialSystoleSearch(X, stopNonzero, verbose)
		got := S.SearchAtVertex(ZVertexInt(0))
		expected := 4
		if got != expected {
			t.Errorf("test: got=%d expected=%d", got, expected)
		}
	}
	{
		S := NewSimplicialSystoleSearch(X, stopNonzero, verbose)
		got := S.SearchAtVertex(ZVertexInt(2))
		expected := 3
		if got != expected {
			t.Errorf("test: got=%d expected=%d", got, expected)
		}
	}
	{
		S := NewSimplicialSystoleSearch(X, stopNonzero, verbose)
		got := S.Search()
		expected := 3
		if got != expected {
			t.Errorf("test: got=%d expected=%d", got, expected)
		}
	}
}

// TestSimplicialSystoleSearchRandomCliqueComplex creates random
// clique complexes, computes the systole using both exhaustive search
// and simplicial search methods, then verifies that both methods
// produce compatible results.  The specific guarantee (see the thesis
// for proofs) is that if the global systole is nonzero, then the
// global systole is no less than the simplicial systole minus one.
func TestSimplicialSystoleSearchRandomCliqueComplex(t *testing.T) {
	trials := 10
	maxVertices := 10
	stopNonzero := true
	verbose := false

	for trial := 0; trial < trials; trial++ {
		numVertices := 3 + rand.Intn(maxVertices)
		probEdge := 0.3
		generator := NewRandomComplexGenerator(numVertices, verbose)
		X, err := generator.RandomCliqueComplex(probEdge)
		if err != nil {
			t.Fatalf("Failed to generate random clique complex: %v", err)
		}
		//log.Printf("random clique complex: %s", X)

		exhaustiveSystole, _, _, _ := ComputeFirstSystole(X.D1(), X.D2(), verbose)

		S := NewSimplicialSystoleSearch(X, stopNonzero, verbose)
		simplicialSystole := S.Search()

		if exhaustiveSystole == 0 {
			// we make no claim here
		} else {
			if exhaustiveSystole < simplicialSystole - 1 {
				log.Printf("complex: %s", X.MaximalSimplicesString())
				t.Fatalf("Trial %d: Mismatch between systole search methods - exhaustive=%d, simplicial=%d",
					trial, exhaustiveSystole, simplicialSystole)
			}
		}
	}
}
