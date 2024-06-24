package golsv

import (
	"log"
	"math/rand"
	"reflect"
	"testing"
)

func TestCosystoleSearchTransform0(t *testing.T) {
	tests := []struct {
		EdgeStates [3]kEdgeState
		Expected [][3]bool
	}{
		{
			[3]kEdgeState{kEdgeOff, kEdgeOff, kEdgeOff},
			[][3]bool{
				{false, false, false},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeOff, kEdgeOff},
			[][3]bool{}, // pruned
		},
		{
			[3]kEdgeState{kEdgeOff, kEdgeOn, kEdgeOff},
			[][3]bool{}, // pruned
		},
		{
			[3]kEdgeState{kEdgeOff, kEdgeOff, kEdgeOn},
			[][3]bool{}, // pruned
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeOn, kEdgeOff},
			[][3]bool{
				{true, true, false},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeOff, kEdgeOn},
			[][3]bool{
				{true, false, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOff, kEdgeOn, kEdgeOn},
			[][3]bool{
				{false, true, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeOn, kEdgeOn},
			[][3]bool{}, // pruned
		},
	}
	for n, test := range tests {
		out := transform0Undecided(test.EdgeStates)
		if len(test.Expected) == 0 {
			if len(out) != 0 {
				t.Errorf("Test %d: transform0Undecided(%v) got=%v expected=%v", n, test.EdgeStates, out, test.Expected)
			}
		} else {
			if !reflect.DeepEqual(out, test.Expected) {
				t.Errorf("Test %d: transform0Undecided(%v) got=%v expected=%v", n, test.EdgeStates, out, test.Expected)
			}
		}
	}
}

func TestCosystoleSearchTransform1(t *testing.T) {
	tests := []struct {
		EdgeStates [3]kEdgeState
		Expected [][3]bool
	}{
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeOff, kEdgeOff},
			[][3]bool{
				{false, false, false},
			},
		},
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeOff, kEdgeOn},
			[][3]bool{
				{true, false, true},
			},
		},
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeOn, kEdgeOff},
			[][3]bool{
				{true, true, false},
			},
		},
		{
			[3]kEdgeState{kEdgeOff, kEdgeUndecided, kEdgeOn},
			[][3]bool{
				{false, true, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeUndecided, kEdgeOff},
			[][3]bool{
				{true, true, false},
			},
		},
		{
			[3]kEdgeState{kEdgeOff, kEdgeOn, kEdgeUndecided},
			[][3]bool{
				{false, true, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeOff, kEdgeUndecided},
			[][3]bool{
				{true, false, true},
			},
		},
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeOn, kEdgeOn},
			[][3]bool{
				{false, true, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeUndecided, kEdgeOn},
			[][3]bool{
				{true, false, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeOn, kEdgeUndecided},
			[][3]bool{
				{true, true, false},
			},
		},
	}
	for n, test := range tests {
		out := transform1Undecided(test.EdgeStates)
		if !reflect.DeepEqual(out, test.Expected) {
			t.Errorf("Test %d: transform1Undecided(%v) got=%v expected=%v", n, test.EdgeStates, out, test.Expected)
		}
	}
}


func TestCosystoleSearchTransform2(t *testing.T) {
	tests := []struct {
		EdgeStates [3]kEdgeState
		Expected [][3]bool
	}{
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeUndecided, kEdgeOff},
			[][3]bool{
				{false, false, false},
				{true, true, false},
			},
		},
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeOff, kEdgeUndecided},
			[][3]bool{
				{false, false, false},
				{true, false, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOff, kEdgeUndecided, kEdgeUndecided},
			[][3]bool{
				{false, false, false},
				{false, true, true},
			},
		},
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeUndecided, kEdgeOn},
			[][3]bool{
				{true, false, true},
				{false, true, true},
			},
		},
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeOn, kEdgeUndecided},
			[][3]bool{
				{true, true, false},
				{false, true, true},
			},
		},
		{
			[3]kEdgeState{kEdgeOn, kEdgeUndecided, kEdgeUndecided},
			[][3]bool{
				{true, true, false},
				{true, false, true},
			},
		},
	}
	for n, test := range tests {
		out := transform2Undecided(test.EdgeStates)
		if !reflect.DeepEqual(out, test.Expected) {
			t.Errorf("Test %d: transform2Undecided(%v) got=%v expected=%v", n, test.EdgeStates, out, test.Expected)
		}
	}
}

func TestCosystoleSearchTransform3(t *testing.T) {
	tests := []struct {
		EdgeStates [3]kEdgeState
		Expected [][3]bool
	}{
		{
			[3]kEdgeState{kEdgeUndecided, kEdgeUndecided, kEdgeUndecided},
			[][3]bool{
				{false, false, false},
				{false, true, true},
				{true, false, true},
				{true, true, false},
			},
		},
	}
	for n, test := range tests {
		out := transform3Undecided(test.EdgeStates)
		if !reflect.DeepEqual(out, test.Expected) {
			t.Errorf("Test %d: transform3Undecided(%v) got=%v expected=%v", n, test.EdgeStates, out, test.Expected)
		}
	}
}

type tcase struct {
	Complex *ZComplex[ZVertexInt]
	Leaf *StateNode
	Level int // level of leaf
	Expected [3]kEdgeState
}

func newTcase(simplices [][]int, branchData [][3]bool, level int, expected [3]kEdgeState) *tcase {
	complex := NewZComplexFromMaximalSimplices(simplices)
	_, leaf := branchFromBools(branchData, level)
	return &tcase{
		Complex: complex,
		Leaf: leaf,
		Level: level,
		Expected: expected,
	}
}

func branchFromBools(branchData [][3]bool, level int) (branch, leaf *StateNode) {
	root := newStateNode(nil)
	p := root
	leaf = root
	for _, data := range branchData {
		leaf = p.addChild(data)
		p = leaf
	}
	return root, leaf
}

func TestCosystoleSearchEdgeStateForTriangle(t *testing.T) {
	var tests []*tcase

	// three independent triangles, level 1
	tests = append(tests,
		newTcase([][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
			[][3]bool{{false, false, false}},
			1,
			[3]kEdgeState{kEdgeUndecided, kEdgeUndecided, kEdgeUndecided}))

	// three independent triangles, level 2
	tests = append(tests,
		newTcase([][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
			[][3]bool{{false, false, false}},
			2,
			[3]kEdgeState{kEdgeUndecided, kEdgeUndecided, kEdgeUndecided}))

	// triangles sharing one edge
	tests = append(tests,
		newTcase([][]int{{0, 1, 2}, {0, 1, 3}},
			[][3]bool{{false, false, false}},
			1,
			[3]kEdgeState{kEdgeOff, kEdgeUndecided, kEdgeUndecided}))

	// triangles sharing one edge
	tests = append(tests,
		newTcase([][]int{{0, 1, 2}, {0, 1, 3}},
			[][3]bool{{true, false, false}},
			1,
			[3]kEdgeState{kEdgeOn, kEdgeUndecided, kEdgeUndecided}))

	// three triangles
	tests = append(tests,
		newTcase([][]int{{0, 1, 2}, {0, 1, 3}, {1, 3, 4}},
			[][3]bool{{false, false, false}, {false, false, false}},
			2,
			[3]kEdgeState{kEdgeOff, kEdgeUndecided, kEdgeUndecided}))

	// three triangles
	tests = append(tests,
		newTcase([][]int{{0, 1, 2}, {0, 1, 3}, {1, 3, 4}},
			[][3]bool{{false, true, true}, {false, true, true}},
			2,
			[3]kEdgeState{kEdgeOn, kEdgeUndecided, kEdgeUndecided}))


	verbose := false
	for n, test := range tests {
		search := NewCosystoleSearch(test.Complex, verbose)
		search.prepare()
		out := edgeStateForTriangle(search.triangles, test.Leaf, test.Level)
		if !reflect.DeepEqual(out, test.Expected) {
			t.Errorf("Test %d: got=%v expected=%v", n, out, test.Expected)
		}
	}
}

func TestCosystoleSearchCocycles(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
		Expected []BinaryVector
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			[]BinaryVector{
				NewBinaryVectorFromInts([]uint8{0, 0, 0}),
				NewBinaryVectorFromInts([]uint8{0, 1, 1}),
				NewBinaryVectorFromInts([]uint8{1, 1, 0}),
				NewBinaryVectorFromInts([]uint8{1, 0, 1}),
			},
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2}, {0, 2}}),
			[]BinaryVector{
				NewBinaryVectorFromInts([]uint8{0, 0, 0}),
				NewBinaryVectorFromInts([]uint8{1, 0, 0}),
				NewBinaryVectorFromInts([]uint8{0, 1, 0}),
				NewBinaryVectorFromInts([]uint8{1, 1, 0}),
				NewBinaryVectorFromInts([]uint8{0, 0, 1}),
				NewBinaryVectorFromInts([]uint8{1, 0, 1}),
				NewBinaryVectorFromInts([]uint8{0, 1, 1}),
				NewBinaryVectorFromInts([]uint8{1, 1, 1}),
			},
		},
	}
	verbose := false
	for n, test := range tests {
		S := NewCosystoleSearch(test.C, verbose)
		cocycles := S.Cocycles()
		if !reflect.DeepEqual(cocycles, test.Expected) {
			t.Errorf("Test %d: got=%v expected=%v", n, cocycles, test.Expected)
		}
	}
}

func TestCosystoleSearchSmall(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
		Expected int
	}{
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}}),
			0,
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1}, {1, 2}, {0, 2}}),
			1,
		},
	}
	verbose := false
	for n, test := range tests {
		_, _, Z_1 := UBDecomposition(test.C.D1(), test.C.D2(), verbose)
		S := NewCosystoleSearch(test.C, verbose)
		cosystole := S.Cosystole(Z_1)
		if cosystole != test.Expected {
			t.Errorf("Test %d: got=%v expected=%v", n, cosystole, test.Expected)
		}
	}
}

func TestCosystoleSearchSimplicialVsGeneralSmall(t *testing.T) {
	tests := []struct {
		C *ZComplex[ZVertexInt]
	}{
		// these are specific cases that highlighted bugs in the past
		{
			NewZComplexFromBoundaryMatrices(
				NewDenseBinaryMatrixFromString(
`1 0 0 0 0 0 0 0 0 0
0 1 1 1 0 0 0 0 0 0
0 0 0 0 1 1 1 0 0 0
1 1 0 0 1 0 0 1 1 0
0 0 1 0 0 1 0 1 0 1
0 0 0 1 0 0 1 0 1 1`),
				NewDenseBinaryMatrixFromString(
`0
0
1
1
0
0
0
0
0
1`),
			),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 2, 4}, {0, 3, 4}, {2,3,4}}),

		},
		{
			NewZComplexFromBoundaryMatrices(
				NewDenseBinaryMatrixFromString(
`1 1 1 0 0 0 0 0 0 0
0 0 0 1 0 0 0 0 0 0
1 0 0 0 1 1 1 0 0 0
0 1 0 0 1 0 0 1 1 0
0 0 1 1 0 1 0 1 0 1
0 0 0 0 0 0 1 0 1 1`),
				NewDenseBinaryMatrixFromString(
`1 0 0 0 0 0
0 1 0 0 0 0
1 1 0 0 0 0
0 0 0 0 0 0
0 0 1 1 0 0
1 0 1 0 1 0
0 0 0 1 1 0
0 1 1 0 0 1
0 0 0 1 0 1
0 0 0 0 1 1`),
			),
		},
		{
			NewZComplexFromBoundaryMatrices(
				NewDenseBinaryMatrixFromString(
`1 1 0 0 0 0 0 0 0 0 0 0 0 0 0
0 0 1 1 1 1 0 0 0 0 0 0 0 0 0
0 0 1 0 0 0 1 1 1 1 0 0 0 0 0
0 0 0 1 0 0 1 0 0 0 1 1 0 0 0
1 0 0 0 0 0 0 1 0 0 0 0 1 1 0
0 1 0 0 1 0 0 0 1 0 1 0 1 0 1
0 0 0 0 0 1 0 0 0 1 0 1 0 1 1`),
				NewDenseBinaryMatrixFromString(
`0 0 0 0 0 0
0 0 0 0 0 0
1 1 0 0 0 0
1 0 0 0 0 0
0 0 1 0 0 0
0 1 1 0 0 0
1 0 0 1 0 0
0 0 0 0 0 0
0 0 0 0 1 0
0 1 0 1 1 0
0 0 0 0 0 1
0 0 0 1 0 1
0 0 0 0 0 0
0 0 0 0 0 0
0 0 1 0 1 1`),
			),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}, {0, 2, 3}, {1, 2, 3}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 2}, {0, 1, 3}, {1, 2, 4}, {0,2,5}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 1, 3}, {1, 2, 4}, {0,2,5}}),
		},
		{
			NewZComplexFromMaximalSimplices([][]int{{0, 2, 4}, {0, 3, 4}}),
		},
		{
			NewZComplexFromBoundaryMatrices(
				NewDenseBinaryMatrixFromString(
`1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
1 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
0 1 0 0 0 0 0 0 0 0 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0
0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 0 0 0 0 0 0 0 0 0
0 0 0 1 0 0 0 1 0 0 1 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 1 0 0 0 0 1 0 0 1 1 1 0 0 0
0 0 0 0 1 0 0 0 1 0 0 0 1 0 0 0 1 0 0 0 0 1 0 1 0 0 1 1 1
0 0 0 0 0 1 0 0 0 0 0 0 0 1 0 0 0 1 0 0 0 0 0 0 1 0 1 0 0
0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 1 0
0 0 0 0 0 0 0 0 0 1 0 0 0 0 1 0 0 0 0 1 0 0 1 0 0 1 0 0 1`),
				NewDenseBinaryMatrixFromString(
`1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
0 0 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
0 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
1 0 0 0 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0
0 0 0 0 0 0 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0
0 0 0 0 0 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0
0 1 0 0 0 0 0 0 0 1 1 0 0 0 0 0 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 0 0 0 0 0 0
0 0 1 0 0 0 0 0 0 1 0 0 0 1 0 0 0 0 0 0 0
0 0 0 1 0 0 0 0 0 0 0 1 0 1 0 0 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 1 0 1 0 0 0 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0
0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0
0 0 0 0 0 0 1 0 0 1 0 0 0 0 0 0 0 0 0 1 0
0 0 0 0 0 0 0 1 0 0 1 0 0 0 0 0 0 0 1 1 0
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1
0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0
0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 0 0 1 0 1
0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 0 0 0 0
0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0
0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 1 0 1 1`),
			),
		},
 	}
	verbose := false
	for n, test := range tests {
		d1 := test.C.D1()
		d2 := test.C.D2()
		if verbose {
			log.Printf("Test %d: %v", n, test.C)
			log.Printf("D1=%v\n%s", d1, dumpMatrix(d1))
			log.Printf("D2=%v\n%s", d2, dumpMatrix(d2))
			log.Printf("bases:\n%s", test.C.DumpBases())
		}
		_, _, Z_1 := UBDecomposition(d1, d2, verbose)
		if verbose {
			log.Printf("Computing cosystole by exhaustive search")
		}
		// sanity check that D_1 * Z_1 = 0
		if !test.C.D1().MultiplyRight(Z_1).IsZero() {
			t.Fatalf("Test %d: D_1 * Z_1 != 0", n)
		}
		cosystoleExhaustive := ComputeFirstCosystole(d1, d2, verbose)
		if verbose {
			log.Printf("cosystoleExhaustive=%d", cosystoleExhaustive)
			log.Printf("Computing cosystole by simplicial search")
		}
		S := NewCosystoleSearch(test.C, verbose)
		cosystoleSimplicial := S.Cosystole(Z_1)
		if verbose {
			log.Printf("cosystoleSimplicial=%d", cosystoleSimplicial)
		}
		if cosystoleSimplicial != cosystoleExhaustive {
			t.Fatalf("Test %d: cosystole simplicial=%v exhaustive=%v", n, cosystoleSimplicial, cosystoleExhaustive)
		}
	}
}

func TestCosystoleSearchSimplicialVsGeneralRandom(t *testing.T) {
	// test some larger randomly generated examples; compare with
	// exhaustive search
	trials := 10
	minVertices := 4
	maxVertices := 10
	vertices := minVertices + rand.Intn(maxVertices-minVertices+1)
	verbose := false
	reallyVerbose := false
	for n := 0; n < trials; n++ {
		gen := NewRandomComplexGenerator(vertices, verbose)
		d_1, d_2, err := gen.RandomSimplicialComplex()
		if err != nil {
			log.Fatalf("Error generating random complex: %v", err)
		}
		if verbose {
			log.Printf("Test %d: generated random complex: d_1=%v d_2=%v", n, d_1, d_2)
			if reallyVerbose {
				log.Printf("d_1=%v:\n%s", d_1, dumpMatrix(d_1))
				log.Printf("d_2=%v:\n%s", d_2, dumpMatrix(d_2))
			}
		}
		_, _, Z_1 := UBDecomposition(d_1, d_2, reallyVerbose)
		if verbose {
			log.Printf("Computing cosystole by exhaustive search")
		}
		cosystoleExhaustive := ComputeFirstCosystole(d_1, d_2, reallyVerbose)
		if verbose {
			log.Printf("cosystoleExhaustive=%d", cosystoleExhaustive)
			log.Printf("Computing cosystole by simplicial search")
		}
		complex := NewZComplexFromBoundaryMatrices(d_1, d_2)
		search := NewCosystoleSearch(complex, reallyVerbose)
		cosystoleSimplicial := search.Cosystole(Z_1)
		if verbose {
			log.Printf("cosystoleSimplicial=%d", cosystoleSimplicial)
		}
		if cosystoleSimplicial != cosystoleExhaustive {
			t.Fatalf("Test %d: cosystole simplicial=%v exhaustive=%v", n, cosystoleSimplicial, cosystoleExhaustive)
			log.Printf("d_1=%v:\n%s", d_1, dumpMatrix(d_1))
			log.Printf("d_2=%v:\n%s", d_2, dumpMatrix(d_2))
		}
	}
}

func TestCosystoleSearchApplyEdgeMask(t *testing.T) {
	tests := []struct {
		edges [3]ZEdge[ZVertexInt]
		states [3]kEdgeState
		mask map[ZEdge[ZVertexInt]]struct{}
		expected [3]kEdgeState
	}{
		{
			edges: [3]ZEdge[ZVertexInt]{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
				NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(0)),
			},
			states: [3]kEdgeState{
				kEdgeOn,
				kEdgeOff,
				kEdgeOn,
			},
			mask: map[ZEdge[ZVertexInt]]struct{}{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)): struct{}{},
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(0)): struct{}{},
			},
			expected: [3]kEdgeState{
				kEdgeOn,
				kEdgeOff,
				kEdgeOn,
			},
		},
		{
			edges: [3]ZEdge[ZVertexInt]{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)),
				NewZEdge[ZVertexInt](ZVertexInt(1), ZVertexInt(2)),
				NewZEdge[ZVertexInt](ZVertexInt(2), ZVertexInt(0)),
			},
			states: [3]kEdgeState{
				kEdgeOn,
				kEdgeOff,
				kEdgeUndecided,
			},
			mask: map[ZEdge[ZVertexInt]]struct{}{
				NewZEdge[ZVertexInt](ZVertexInt(0), ZVertexInt(1)): struct{}{},
			},
			expected: [3]kEdgeState{
				kEdgeOn,
				kEdgeOff,
				kEdgeOff,
			},
		},
	}
	for n, test := range tests {
		got := applyEdgeMask(test.edges, test.states, test.mask)
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("Test %d: got=%v expected=%v", n, got, test.expected)
		}
	}
}
