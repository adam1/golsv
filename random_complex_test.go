package golsv

import (
	"fmt"
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

func TestRandom2dCliqueComplexComplete(t *testing.T) {
	for numVertices := 1; numVertices < 20; numVertices++ {
		// verify that with p=1, we get the clique complex of a complete graph
		verbose := false
		R := NewRandomComplexGenerator(numVertices, verbose)
		X, err := R.RandomCliqueComplex(1.0)
		if err != nil {
			t.Error("wanted no error, got ", err)
		}
		d_1 := X.D1()
		d_2 := X.D2()
		if d_1.NumRows() != numVertices {
			t.Error("wanted d_1.NumRows() == n, got ", d_1.NumRows())
		}
		// Check that the number of edges equals numVertices choose 2
		expectedEdges := numVertices * (numVertices - 1) / 2
		if d_1.NumColumns() != expectedEdges {
			t.Errorf("wanted %d edges (complete graph), got %d", expectedEdges, d_1.NumColumns())
		}

		// Check that the number of triangles equals numVertices choose 3
		expectedTriangles := numVertices * (numVertices - 1) * (numVertices - 2) / 6
		if d_2.NumColumns() != expectedTriangles {
			t.Errorf("wanted %d triangles (complete 2-complex), got %d", expectedTriangles, d_2.NumColumns())
		}
	}
}

func TestRandom2dCliqueComplex(t *testing.T) {
	trials := 10
	minDimC_0 := 1
	maxDimC_0 := 20
	verbose := false
	for i := 0; i < trials; i++ {
		dimC_0 := rand.Intn(maxDimC_0 - minDimC_0) + minDimC_0
		R := NewRandomComplexGenerator(dimC_0, verbose)
		X, err := R.RandomCliqueComplex(0.3)
		if err != nil {
			t.Error("wanted no error, got ", err)
		}
		d_1 := X.D1()
		if d_1.NumRows() != dimC_0 {
			t.Error("wanted d_1.NumRows() == n, got ", d_1.NumRows())
		}
		// xxx not sure what to check here
		// if d_1.NumColumns() < 1 {
		// 	t.Error("wanted d_1.NumColumns() > 0, got ", d_1.NumColumns())
		// }
		// if d_2.NumRows() < 1 {
		// 	t.Error("wanted d_2.NumRows() > 0, got ", d_2.NumRows())
		// }
		// if d_2.NumColumns() < 1 {
		// 	t.Error("wanted d_2.NumColumns() > 0, got ", d_2.NumColumns())
		// }
	}
}

// func TestRandomRegularCliqueComplex(t *testing.T) {
// 	tests := []struct {
// 		numVertices int
// 		k           int
// 		expectError bool
// 	}{
// 		{6, 2, false},  // 6 vertices, 2-regular (12 total degree, 6 edges)
// 		{8, 3, false},  // 8 vertices, 3-regular (24 total degree, 12 edges)
// 		{6, 3, false},  // 6 vertices, 3-regular (18 total degree, 9 edges)
// 		{5, 3, true},   // 5*3=15 is odd, should fail
// 		{4, 4, true},   // k >= numVertices, should fail
// 	}

// 	verbose := false
// 	for _, test := range tests {
// 		t.Run(fmt.Sprintf("n=%d_k=%d", test.numVertices, test.k), func(t *testing.T) {
// 			R := NewRandomComplexGenerator(test.numVertices, verbose)
// 			d_1, d_2, err := R.RandomRegularCliqueComplex(test.k)

// 			if test.expectError {
// 				if err == nil {
// 					t.Errorf("expected error for n=%d, k=%d", test.numVertices, test.k)
// 				}
// 				return
// 			}

// 			if err != nil {
// 				t.Errorf("unexpected error: %v", err)
// 				return
// 			}

// 			if d_1.NumRows() != test.numVertices {
// 				t.Errorf("expected d_1.NumRows()=%d, got %d", test.numVertices, d_1.NumRows())
// 			}

// 			expectedEdges := test.numVertices * test.k / 2
// 			if d_1.NumColumns() != expectedEdges {
// 				t.Errorf("expected %d edges, got %d", expectedEdges, d_1.NumColumns())
// 			}

// 			// Verify regularity: each vertex should have degree k
// 			degrees := make([]int, test.numVertices)
// 			for j := 0; j < d_1.NumColumns(); j++ {
// 				vertices := make([]int, 0, 2)
// 				for i := 0; i < test.numVertices; i++ {
// 					if d_1.Get(i, j) == 1 {
// 						vertices = append(vertices, i)
// 					}
// 				}
// 				if len(vertices) != 2 {
// 					t.Errorf("edge %d should connect exactly 2 vertices, got %d", j, len(vertices))
// 				}
// 				degrees[vertices[0]]++
// 				degrees[vertices[1]]++
// 			}

// 			for i, degree := range degrees {
// 				if degree != test.k {
// 					t.Errorf("vertex %d has degree %d, expected %d", i, degree, test.k)
// 				}
// 			}

// 			// Basic checks on d_2 (triangles from clique filling)
// 			if d_2.NumRows() != d_1.NumColumns() {
// 				t.Errorf("expected d_2.NumRows()=%d, got %d", d_1.NumColumns(), d_2.NumRows())
// 			}

// 			// Each triangle should have weight 3
// 			for j := 0; j < d_2.NumColumns(); j++ {
// 				weight := d_2.ColumnWeight(j)
// 				if weight != 3 {
// 					t.Errorf("triangle %d has weight %d, expected 3", j, weight)
// 				}
// 			}
// 		})
// 	}
// }

func TestRandomCirculantStepsWithTriangles(t *testing.T) {
	tests := []struct {
		n           int
		k           int
		expectError bool
	}{
		{6, 2, false},   // 6 vertices, 2-regular
		{8, 4, false},   // 8 vertices, 4-regular
		{10, 6, false},  // 10 vertices, 6-regular
		{12, 8, false},  // 12 vertices, 8-regular
		{6, 3, true},    // k must be even
		{4, 4, true},    // k >= n
		{1, 0, true},    // n must be at least 2
		{5, 1, true},    // k must be even
	}

	trials := 10 // Reduced to 10 for regular testing after debugging
	for _, test := range tests {
		t.Run(fmt.Sprintf("n=%d_k=%d", test.n, test.k), func(t *testing.T) {
			if test.expectError {
				_, err := randomCirculantStepsWithTriangles(test.n, test.k)
				if err == nil {
					t.Errorf("expected error for n=%d, k=%d", test.n, test.k)
				}
				return
			}

			// Run multiple trials to verify consistency
			for trial := 0; trial < trials; trial++ {
				steps, err := randomCirculantStepsWithTriangles(test.n, test.k)
				if err != nil {
					t.Errorf("trial %d: unexpected error: %v", trial, err)
					continue
				}

				// Verify step count
				expectedSteps := test.k / 2
				if len(steps) != expectedSteps {
					t.Errorf("trial %d: expected %d steps, got %d", trial, expectedSteps, len(steps))
					continue
				}

				// Normalize steps to get full step set (including inverses)
				normalized := normalizeCirculantSteps(test.n, steps)

				// Verify normalized step count equals k
				if len(normalized) != test.k {
					t.Errorf("trial %d: expected %d normalized steps, got %d. Original steps: %v, Normalized: %v", 
						trial, test.k, len(normalized), steps, normalized)
					continue
				}

				// Verify all steps are valid (> 0 and < n)
				for _, step := range normalized {
					if step <= 0 || step >= test.n {
						t.Errorf("trial %d: invalid step %d (must be in range [1, %d))", trial, step, test.n)
					}
				}

				// Verify no duplicate steps in normalized set
				stepSet := make(map[int]bool)
				for _, step := range normalized {
					if stepSet[step] {
						t.Errorf("trial %d: duplicate step %d in normalized set %v", trial, step, normalized)
					}
					stepSet[step] = true
				}

				// Verify all original steps are in valid range
				for _, step := range steps {
					if step <= 0 || step >= test.n {
						t.Errorf("trial %d: invalid original step %d", trial, step)
					}
					if step == test.n-1 {
						t.Errorf("trial %d: step %d equals n-1, should be excluded", trial, step)
					}
				}

				// Verify step 1 is always included
				found1 := false
				for _, step := range steps {
					if step == 1 {
						found1 = true
						break
					}
				}
				if !found1 {
					t.Errorf("trial %d: step 1 should always be included, got steps %v", trial, steps)
				}
			}
		})
	}
}


func TestRandomCirculantComplex(t *testing.T) {
	tests := []struct {
		n                      int
		k                      int
		expectError            bool
		expectedTrianglesRange [2]int
	}{
		{6, 2, false, [2]int{0,0}},   // 6 vertices, 2-regular circulant
		{8, 4, false, [2]int{0,10}},  // 8 vertices, 4-regular circulant
		{10, 6, false, [2]int{1,100}}, // 10 vertices, 6-regular circulant
		{6, 3, true, [2]int{0,0}},   // k must be even
		{4, 4, true, [2]int{0,0}},   // k >= n
		{1, 0, true, [2]int{0,0}},   // n must be at least 2
		{5, 1, true, [2]int{0,0}},   // k must be even
	}

	verbose := false
	for _, test := range tests {
		t.Run(fmt.Sprintf("n=%d_k=%d", test.n, test.k), func(t *testing.T) {
			R := NewRandomComplexGenerator(test.n, verbose)
			complex, err := R.RandomCirculantComplex(test.n, test.k)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error for n=%d, k=%d", test.n, test.k)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			d_1 := complex.D1()
			d_2 := complex.D2()

			if d_1.NumRows() != test.n {
				t.Errorf("expected d_1.NumRows()=%d, got %d", test.n, d_1.NumRows())
			}

			expectedEdges := test.n * test.k / 2
			if d_1.NumColumns() != expectedEdges {
				t.Errorf("expected %d edges, got %d", expectedEdges, d_1.NumColumns())
			}

			// Verify regularity: each vertex should have degree k
			degrees := make([]int, test.n)
			for j := 0; j < d_1.NumColumns(); j++ {
				vertices := make([]int, 0, 2)
				for i := 0; i < test.n; i++ {
					if d_1.Get(i, j) == 1 {
						vertices = append(vertices, i)
					}
				}
				if len(vertices) != 2 {
					t.Errorf("edge %d should connect exactly 2 vertices, got %d", j, len(vertices))
				}
				degrees[vertices[0]]++
				degrees[vertices[1]]++
			}

			for i, degree := range degrees {
				if degree != test.k {
					t.Errorf("vertex %d has degree %d, expected %d", i, degree, test.k)
				}
			}

			// Basic checks on d_2 (triangles from clique filling)
			if d_2.NumRows() != d_1.NumColumns() {
				t.Errorf("expected d_2.NumRows()=%d, got %d", d_1.NumColumns(), d_2.NumRows())
			}

			// Check expected number of triangles
			numTriangles := d_2.NumColumns()
			if numTriangles < test.expectedTrianglesRange[0] ||
				numTriangles > test.expectedTrianglesRange[1] {
				t.Errorf("expected triangles in range %v, got %d", test.expectedTrianglesRange, numTriangles)
			}

			// Each triangle should have weight 3
			for j := 0; j < d_2.NumColumns(); j++ {
				weight := d_2.ColumnWeight(j)
				if weight != 3 {
					t.Errorf("triangle %d has weight %d, expected 3", j, weight)
				}
			}
		})
	}
}


func TestRandomRegularCliqueComplexByBalancing(t *testing.T) {
	tests := []struct {
		numVertices int
		degree      int
		expectError bool
	}{
		{6, 2, false},  // 6 vertices, 2-regular
		{8, 3, false},  // 8 vertices, 3-regular
		{10, 4, false}, // 10 vertices, 4-regular
		{5, 3, true},   // 5*3=15 is odd, should fail
		{4, 4, true},   // degree >= numVertices, should fail
	}
	verbose := false
	for _, test := range tests {
		t.Run(fmt.Sprintf("n=%d_k=%d", test.numVertices, test.degree), func(t *testing.T) {
			R := NewRandomComplexGenerator(test.numVertices, verbose)
			C, err := R.RandomRegularCliqueComplexByBalancing(test.degree, 1000)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error for n=%d, degree=%d", test.numVertices, test.degree)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if C.NumVertices() != test.numVertices {
				t.Errorf("expected %d vertices, got %d", test.numVertices, C.NumVertices())
			}
			isRegular, actualDegree := C.IsRegular()
			if !isRegular {
				t.Errorf("generated graph is not regular")
			}
			if test.numVertices > 0 && actualDegree != test.degree {
				t.Errorf("expected degree %d, got %d", test.degree, actualDegree)
			}
			expectedEdges := test.numVertices * test.degree / 2
			if C.NumEdges() != expectedEdges {
				t.Errorf("expected %d edges, got %d", expectedEdges, C.NumEdges())
			}
			// xxx there may or may not be triangles
// 			if C.NumTriangles() == 0 && test.numVertices > 2 {
// 				t.Errorf("expected triangles in clique complex, got 0")
			// 			}
			C.D1()
			C.D2()
		})
	}
}
