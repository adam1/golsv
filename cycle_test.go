package golsv

import (
	"testing"
)

func TestEnumerateWordsN(t *testing.T) {
	for length := 1; length <= 3; length++ {
		count := 0
		handler :=  func(word []MatGF, path Path, product MatGF) (Continue bool) {
			count++
			// fmt.Println(words)
 			if len(word) != len(path) {
				t.Errorf("Expected word and path to be the same length, got %d and %d", len(word), len(path))
			}
			return true
		}			
		EnumerateWordsN(lsv, lsv.Generators(), length, handler)
		expected := 14 * pow(13, length-1)
		if count != expected {
			t.Errorf("Expected %d words, got %d", expected, count)
		}
	}
}

func TestPow(t *testing.T) {
	n := 14 * 14 * 14
	m := pow(14, 3)
	if m != n {
		t.Errorf("Expected %d, got %d", n, m)
	}
}

func TestFindCyclesLength2(t *testing.T) {
	n := 0
	handleCycle := func(cycle []MatGF, path Path, product MatGF) (Continue bool) {
		n++
		if len(cycle) != 2 {
			t.Errorf("Expected cycle length 2, got %d", len(cycle))
		}
		if len(path) != 2 {
			t.Errorf("Expected path length 2, got %d", len(path))
		}
		if !path[0][0].Equal(lsv, MatGfIdentity) {
			t.Errorf("Expected path[0][0] to be identity, got %v", path[0][0])
		}
		if !path[0][1].Equal(lsv, &cycle[0]) {
			t.Errorf("Expected path[0][1] to be %v, got %v", cycle[0], path[0][1])
		}
		if !path[1][0].Equal(lsv, MatGfIdentity) {
			t.Errorf("Expected path[1][0] to be identity, got %v", path[1][0])
		}
		if !path[1][1].Equal(lsv, &cycle[0]) {
			t.Errorf("Expected path[1][1] to be %v, got %v", cycle[0], path[1][1])
		}
		return true
	}
	FindCycles(lsv, lsv.Generators(), 2, handleCycle)
	want := 0
	if n != want {
		t.Errorf("Expected %d cycles, got %d", want, n)
	}
}

func TestLsvTrianglesAtOrigin(t *testing.T) {
	triangles := LsvTrianglesAtOrigin(lsv)
	if len(triangles) != 21 {
		t.Errorf("Expected 21 paths, got %d", len(triangles))
	}
	for _, triangle := range triangles {
		if len(triangle) != 3 {
			t.Errorf("Expected triangle length 3, got %d", len(triangle))
			continue
		}
		if !triangle[0].Equal(lsv, MatGfIdentity) {
			t.Errorf("Expected triangle[0] to be identity, got %v", triangle[0])
		}
		checkTriangle(t, triangle)
	}
}

func TestFindLsvCycles(t *testing.T) {
	// currently nothing to check here; just make sure it doesn't crash
	FindLsvCycles(lsv, 1, 2)
}

func checkTriangle(t *testing.T, triangle Triangle) {
	// start with first vertex and multiply by the generator implied
	// by each edge
    a := triangle[0]
	b := triangle[1]
	c := triangle[2]
	g := a.Inverse(lsv).Multiply(lsv, &b)
	h := b.Inverse(lsv).Multiply(lsv, &c)
	i := c.Inverse(lsv).Multiply(lsv, &a)
	product := a.Multiply(lsv, g).Multiply(lsv, h).Multiply(lsv, i)
	if !product.Equal(lsv, MatGfIdentity) {
		t.Errorf("Expected triangle %v to close, got %v", triangle, product)
	}
}

func TestCyclicSubgroupPath(t *testing.T) {
	path := CyclicSubgroupPath(lsv, &lsv.Generators()[0])
	if len(path) != 273 {
		t.Errorf("Expected path length 273, got %d", len(path))
	}
}

func TestCyclesLength4(t *testing.T) {
	n := 0
	handleCycle := func(cycle []MatGF, path Path, product MatGF) (Continue bool) {
		n++
		if len(cycle) != 4 {
			t.Errorf("Expected cycle length 4, got %d", len(cycle))
		}
		if len(path) != 4 {
			t.Errorf("Expected path length 4, got %d", len(path))
		}
		if !path[0][0].Equal(lsv, MatGfIdentity) {
			t.Errorf("Expected path[0][0] to be identity, got %v", path[0][0])
		}
		if !cycle[0].Multiply(lsv, &cycle[1]).Multiply(lsv, &cycle[2]).Multiply(lsv, &cycle[3]).Equal(lsv, MatGfIdentity) {
			t.Errorf("Expected product of matrices in cycle to be identity, got %v", product)
		}
		return true
	}
	FindCycles(lsv, lsv.Generators(), 4, handleCycle)
	want := 168
	if n != want {
		t.Errorf("Expected %d cycles, got %d", want, n)
	}
}
