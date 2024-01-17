package bench_galoisfield

import (
	"fmt"
	"testing"

	"github.com/cloud9-tools/go-galoisfield"
)

// Package galiosfield supports GF(2^1) through GF(2^8).  Here we use GF(16)
// with the built-in modulus t^4 + t + 1 (the same as used in the main LSV example).
// The field elements are represented as bit vectors, stored in a single byte, with:
//
//   0x01 = 1
//   0x02 = t
//   0x04 = t^2
//   0x08 = t^3.
//
// we use hard-coded 3x3 matrices flattened into a single array.
// the code is derived from
// https://rosettacode.org/wiki/Matrix_multiplication#Library_go.matrix

var GF *galoisfield.GF = galoisfield.Poly410_g2

const stride = 3

type matrix [9]byte

var identity = &matrix{
	0x01, 0x00, 0x00,
	0x00, 0x01, 0x00,
	0x00, 0x00, 0x01,
}

func (m *matrix) String() (s string) {
	s = "{\n"
	for i := 0; i < len(m); i += stride {
		for j := 0; j < stride; j++ {
			s += fmt.Sprintf("0x%02x,", m[i+j])
		}
		s += "\n"
	}
	s += "}"
	return s
}

func (m1 *matrix) multiply(m2 *matrix) (m3 *matrix) {
	m3 = &matrix{}
	for m1c0, m3x := 0, 0; m1c0 < len(m1); m1c0 += stride {
		for m2r0 := 0; m2r0 < stride; m2r0++ {
			for m1x, m2x := m1c0, m2r0; m2x < len(m2); m2x += stride {
				m3[m3x] = GF.Add(m3[m3x], GF.Mul(m1[m1x], m2[m2x]))
				m1x++
			}
			m3x++
		}
	}
	return m3
}

func (a *matrix) equal(b *matrix) bool {
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (m *matrix) canonical_representative() (b *matrix) {
	b = &matrix{}
	a := m.first_nonzero_entry()
	for i := 0; i < len(m); i++ {
		b[i] = GF.Div(m[i], a)
	}
	return b
}

func (m *matrix) first_nonzero_entry() byte {
	for i := 0; i < len(m); i++ {
		if m[i] != 0 {
			return m[i]
		}
	}
	panic("matrix is all zero")
}

func TestCanonicalRepresentative(t *testing.T) {
	var m matrix = LSVExampleGenerators[0]
	// fmt.Printf("m: %s\n", m.String())
	m_rep := m.canonical_representative()
	// fmt.Printf("m_rep: %s\n", m_rep.String())
	if m_rep.String() != `{
0x01,0x05,0x0e,
0x0b,0x0a,0x02,
0x0e,0x09,0x06,
}` {
		t.Fail()
	}
}

var LSVExampleGenerators = [...]matrix{
	// printed from test_cayley_graph_generator.py
	{
		0x0a, 0x04, 0x06,
		0x02, 0x08, 0x07,
		0x06, 0x05, 0x09,
	},
	{
		0x0f, 0x06, 0x05,
		0x03, 0x0c, 0x01,
		0x05, 0x02, 0x08,
	},
	{
		0x0d, 0x05, 0x02,
		0x07, 0x0a, 0x04,
		0x02, 0x03, 0x0c,
	},
	{
		0x0e, 0x02, 0x03,
		0x01, 0x0f, 0x06,
		0x03, 0x07, 0x0a,
	},
	{
		0x09, 0x03, 0x07,
		0x04, 0x0d, 0x05,
		0x07, 0x01, 0x0f,
	},
	{
		0x08, 0x07, 0x01,
		0x06, 0x0e, 0x02,
		0x01, 0x04, 0x0d,
	},
	{
		0x0c, 0x01, 0x04,
		0x05, 0x09, 0x03,
		0x04, 0x06, 0x0e,
	},
}

var LSVExampleGeneratorsCanonicalSymmetric = [...]matrix{
	{
		0x01, 0x05, 0x0e,
		0x0b, 0x0a, 0x02,
		0x0e, 0x09, 0x06,
	},
	{
		0x01, 0x0c, 0x08,
		0x00, 0x07, 0x0c,
		0x0c, 0x0c, 0x03,
	},
	{
		0x01, 0x05, 0x0e,
		0x0b, 0x0a, 0x08,
		0x0e, 0x03, 0x0c,
	},
	{
		0x01, 0x0a, 0x0a,
		0x05, 0x02, 0x00,
		0x0a, 0x00, 0x0d,
	},
	{
		0x01, 0x07, 0x08,
		0x0f, 0x0e, 0x03,
		0x08, 0x0c, 0x05,
	},
	{
		0x01, 0x09, 0x00,
		0x0e, 0x0f, 0x09,
		0x0d, 0x0d, 0x0c,
	},
	{
		0x01, 0x06, 0x05,
		0x03, 0x02, 0x0a,
		0x05, 0x09, 0x0d,
	},
	{
		0x01, 0x0d, 0x0d,
		0x00, 0x0c, 0x04,
		0x07, 0x0e, 0x0f,
	},
	{
		0x01, 0x06, 0x0e,
		0x08, 0x09, 0x0a,
		0x0e, 0x02, 0x0d,
	},
	{
		0x01, 0x0b, 0x05,
		0x09, 0x04, 0x02,
		0x09, 0x00, 0x08,
	},
	{
		0x01, 0x0b, 0x0f,
		0x04, 0x05, 0x0d,
		0x0f, 0x09, 0x07,
	},
	{
		0x01, 0x09, 0x00,
		0x05, 0x08, 0x0b,
		0x0b, 0x09, 0x04,
	},
	{
		0x01, 0x0a, 0x0e,
		0x04, 0x05, 0x0d,
		0x0e, 0x09, 0x06,
	},
	{
		0x01, 0x07, 0x04,
		0x09, 0x0f, 0x00,
		0x09, 0x0e, 0x0c,
	},
}

func TestMatrixMultiplicationIdentity(t *testing.T) {
	u := identity.multiply(identity)
	if !u.equal(identity) {
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

func BenchmarkMatrix(b *testing.B) {
	G := galoisfield.Poly410_g2
	fmt.Println("G=", G)
	if G.Size() != 16 {
		b.Fail()
	}
	gens := LSVExampleGeneratorsCanonicalSymmetric

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(gens); j++ {
			h := identity
			n := 0
			for {
				n++
				u := h.multiply(&gens[j])
				u_rep := u.canonical_representative()
				if u_rep.equal(identity) {
					// fmt.Printf("Order of g is %d\n", n)
					break
				}
				h = u_rep
			}
		}

	}
}
