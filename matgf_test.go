package golsv

import (
	"fmt"
	"testing"

	"github.com/cloud9-tools/go-galoisfield"
)

func TestCanonicalRepresentative(t *testing.T) {
	var m MatGF = lsv.Generators()[0]
	m.MakeCanonical(lsv)
	expected := "[1 5 14 11 10 2 14 9 6]"
	got := fmt.Sprintf("%v", m)
	if got != expected {
		t.Errorf("expected: %s\ngot: %s", expected, got)
	}
}

func TestMatGFMultiplicationIdentity(t *testing.T) {
	u := MatGfIdentity.Multiply(lsv, MatGfIdentity)
	if !u.Equal(lsv, MatGfIdentity) {
		t.Fail()
	}
}

func TestGFMultiplicationByIdentity(t *testing.T) {
	G := galoisfield.Poly410_g2
	id := byte(0x01)
	for i := 0; i < 16; i++ {
		a := byte(i)
		b := G.Mul(a, id)
		if a != b {
			t.Fail()
		}
	}
}

func TestEqualCanonicalizes(t *testing.T) {
	a := MatGF{1, 5, 14, 11, 10, 8, 14, 3, 12}
	scalar := MatGF{2, 0, 0, 0, 2, 0, 0, 0, 2}
	b := a.Multiply(lsv, &scalar)
	if !a.Equal(lsv, b) {
		t.Fail()
	}
}

func TestGFMultiplication(t *testing.T) {
	// verify that we understand how to do multiplication in this field
	G := galoisfield.Poly410_g2

	if G.Mul(0x02, 0x04) != 0x08 {
		t.Fail()
	}

	if G.Mul(0x08, 0x08) != 0x0c {
		t.Fail()
	}
}

func BenchmarkMatGF(b *testing.B) {
	G := galoisfield.Poly410_g2
	fmt.Println("G=", G)
	if G.Size() != 16 {
		b.Fail()
	}
	gens := lsv.Generators()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(gens); j++ {
			h := MatGfIdentity
			n := 0
			for {
				n++
				u := h.Multiply(lsv, &gens[j])
				u.MakeCanonical(lsv)
				if u.Equal(lsv, MatGfIdentity) {
					// fmt.Printf("Order of g is %d\n", n)
					break
				}
				h = u
			}
		}

	}
}

func TestProduct1(t *testing.T) {
	c := C2Generators[0]
	list := []MatGF{c, c}
	p := Product(lsv, list)
	if !p.Equal(lsv, MatGfIdentity) {
		t.Fail()
	}
}

func TestProduct2(t *testing.T) {
	c := C3Generators[0]
	list := []MatGF{c, c, c}
	p := Product(lsv, list)
	if !p.Equal(lsv, MatGfIdentity) {
		t.Fail()
	}
}

func TestProduct3(t *testing.T) {
	c := C2Generators[0]
	list := []MatGF{c, c, c}
	p := Product(lsv, list)
	if !p.Equal(lsv, &c) {
		t.Fail()
	}
}

func TestSwapRows(t *testing.T) {
	var m = MatGF{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
	m.SwapRows(0, 1)
	var expected = MatGF{
		0, 1, 0,
		1, 0, 0,
		0, 0, 1,
	}
	if !m.Equal(lsv, &expected) {
		t.Fail()
	}
	m.SwapRows(0, 1)
	if !m.Equal(lsv, MatGfIdentity) {
		t.Fail()
	}
}

func TestInverseIdentity(t *testing.T) {
	var m = MatGF{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
	var expected = MatGF{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
	if !m.Inverse(lsv).Equal(lsv, &expected) {
		t.Fail()
	}
}

func TestInverseGenerators(t *testing.T) {
	for _, g := range lsv.Generators() {
		if !(&g).Inverse(lsv).Multiply(lsv, &g).Equal(lsv, MatGfIdentity) {
			t.Fail()
		}
	}
}

func TestInverseRandom(t *testing.T) {
	trials := 100
	inverted := 0
	for i := 0; i < trials; i++ {
		ok := testRandomMatGF(t)
		if ok {
			inverted++
		}
	}
	// fmt.Printf("inverted %d matrices in %d trials\n", inverted, trials)
}

func testRandomMatGF(t *testing.T) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	g := RandomMatGF(lsv)
	if !g.Inverse(lsv).Multiply(lsv, g).Equal(lsv, MatGfIdentity) {
		t.Fail()
	}
	return true
}

func TestLess(t *testing.T) {
	a := &MatGF{1, 2, 3}
	b := &MatGF{1, 2, 4}
	if !a.Less(b) {
		t.Fail()
	}
	if b.Less(a) {
		t.Fail()
	}
	a = &MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13}
	b = &MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6}
	if a.Less(b) {
		t.Fail()
	}
}

func TestLsvGeneratingSetSquareFree(t *testing.T) {
	index := make(map[MatGF]any)
	gens := lsv.Generators()
	for _, g := range gens {
		index[g] = nil
	}
	for _, g := range gens {
		g2 := g.Multiply(lsv, &g)
		if _, found := index[*g2]; found {
			t.Fail()
		}
	}
}
