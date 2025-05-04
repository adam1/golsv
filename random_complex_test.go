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

func TestRandom2dCliqueComplexComplete(t *testing.T) {
	// verify that with p=1, we get the clique complex of a complete graph
	numVertices := 10
	verbose := false
	R := NewRandomComplexGenerator(numVertices, verbose)
	d_1, d_2, err := R.RandomCliqueComplex(1.0)
	if err != nil {
		t.Error("wanted no error, got ", err)
	}
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

func TestRandom2dCliqueComplex(t *testing.T) {
	trials := 10
	minDimC_0 := 1
	maxDimC_0 := 20
	verbose := false
	for i := 0; i < trials; i++ {
		dimC_0 := rand.Intn(maxDimC_0 - minDimC_0) + minDimC_0
		R := NewRandomComplexGenerator(dimC_0, verbose)
		d_1, _, err := R.RandomCliqueComplex(0.3)
		if err != nil {
			t.Error("wanted no error, got ", err)
		}
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