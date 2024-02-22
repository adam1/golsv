package golsv

import (
	"fmt"
	"crypto/rand"
	"strconv"
	"strings"

	"github.com/cloud9-tools/go-galoisfield"
)

// Currently, this code is dedicated to computing a small number of
// specific examples - the example given in section 10 of the LSV
// paper[1], and smaller examples related to it.
//
// Reference:
//
// [1] Alexander Lubotzky, Beth Samuels, Uzi Vishne.  Explicit Constructions of Ramanujan Complexes. 2005.
// http://arxiv.org/abs/math/0406217v2
//
// Package galiosfield supports GF(2^1) through GF(2^8).  Here we use GF(16)
// with the built-in modulus t^4 + t + 1 (the same as used in the main LSV example).
// The field elements are represented as bit vectors, stored in a single byte, with:
//
//   0x01 = 1
//   0x02 = t
//   0x04 = t^2
//   0x08 = t^3.
//
// We use hard-coded 3x3 matrices flattened into a single array.  the
// code is derived from
// https://rosettacode.org/wiki/Matrix_multiplication#Library_go.Matrix
//
// This matrix type handles 3x3 matrices with byte entries from
// package galiosfield, hence models matrices over GF(2), GF(4),
// GF(8), or GF(16).
type MatGF [9]byte

const MatGFStride = 3

var MatGfIdentity = &MatGF{
	1, 0, 0,
	0, 1, 0,
	0, 0, 1,
}

func (m MatGF) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[")
	for i := 0; i < len(m); i ++ {
		fmt.Fprintf(&b, "%d", m[i])
		if i != len(m)-1 {
			fmt.Fprintf(&b, " ")
		}
	}
	fmt.Fprintf(&b, "]")
	return b.String()
}

func (m *MatGF) BigString() (s string) {
	s = "{\n"
	for i := 0; i < len(m); i += MatGFStride {
		for j := 0; j < MatGFStride; j++ {
			s += fmt.Sprintf("%d,", m[i+j])
		}
		s += "\n"
	}
	s += "}"
	return s
}

// xxx test
func NewMatGFFromString(s string) MatGF {
	m := &MatGF{}
	entries := strings.Split(strings.Trim(s, " []"), " ")
	if len(entries) != len(m) {
		panic("wrong number of entries")
	}
	for i := 0; i < len(m); i++ {
		strings.TrimSpace(entries[i])
		c, err := strconv.ParseUint(entries[i], 10, 4)
		if err != nil {
			panic(err)
		}
		m[i] = byte(c)
	}
	return *m
}

// xxx wip; 
func NewReducedMatGF(lsv *LsvContext, a *MatGF) *MatGF {
	b := &MatGF{}
	for i, d := range a {
		b[i] = byte(polyDiv(uint(d), lsv.GF.Polynomial()))
	}
	b.MakeCanonical(lsv)
	return b
}

func NewMatGFFromProjMatF2Poly(lsv *LsvContext, p ProjMatF2Poly) MatGF {
	m := MatGF{}
	gfSize := lsv.GF.Size()
	modDegree := log2(gfSize)
	for i := 0; i < len(m); i++ {
		if p[i].Degree() >= int(modDegree) {
			panic("degree too high")
		}
		m[i] = byteFromBinaryString(p[i].String())
	}
	return m
}

func log2(x uint) uint {
	var r uint
	for x > 1 {
		x >>= 1
		r++
	}
	return r
}
	

// ===================================================
// xxx wip; borrowed from go-galoisfield/gf.go
//
// polyDiv divides two polynomials and returns the remainder.
func polyDiv(dividend, divisor uint) uint {
	for m, n := degree(dividend), degree(divisor); m >= n; m-- {
		if (dividend & (1 << (m - 1))) != 0 {
			dividend ^= divisor << (m - n)
		}
	}
	return dividend
}
// degree returns the degree of the polynomial.  In this representation, the
// degree of a polynomial is:
//	[p == 0] 0
//	[p >  0] (k+1) such that (1<<k) is the highest 1 bit
func degree(p uint) uint {
	var d uint
	for p > 0 {
		d++
		p >>= 1
	}
	return d
}
// ===================================================

// xxx test
func (a *MatGF) Copy() *MatGF {
	b := &MatGF{}
	copy(b[:], a[:])
	return b
}

// xxx tbd: purge the extraneous use of MakeCanonical from elsewhere;
// make it private.
func (a *MatGF) Multiply(lsv *LsvContext, b *MatGF) (c *MatGF) {
	c = &MatGF{}
	f := lsv.GF
	c[0] = f.Add(f.Add(f.Mul(a[0], b[0]), f.Mul(a[1], b[3])), f.Mul(a[2], b[6]))
	c[1] = f.Add(f.Add(f.Mul(a[0], b[1]), f.Mul(a[1], b[4])), f.Mul(a[2], b[7]))
	c[2] = f.Add(f.Add(f.Mul(a[0], b[2]), f.Mul(a[1], b[5])), f.Mul(a[2], b[8]))
	c[3] = f.Add(f.Add(f.Mul(a[3], b[0]), f.Mul(a[4], b[3])), f.Mul(a[5], b[6]))
	c[4] = f.Add(f.Add(f.Mul(a[3], b[1]), f.Mul(a[4], b[4])), f.Mul(a[5], b[7]))
	c[5] = f.Add(f.Add(f.Mul(a[3], b[2]), f.Mul(a[4], b[5])), f.Mul(a[5], b[8]))
	c[6] = f.Add(f.Add(f.Mul(a[6], b[0]), f.Mul(a[7], b[3])), f.Mul(a[8], b[6]))
	c[7] = f.Add(f.Add(f.Mul(a[6], b[1]), f.Mul(a[7], b[4])), f.Mul(a[8], b[7]))
	c[8] = f.Add(f.Add(f.Mul(a[6], b[2]), f.Mul(a[7], b[5])), f.Mul(a[8], b[8]))
	return c.MakeCanonical(lsv)
}

// xxx test that this autocanonicalizes
func (a *MatGF) Equal(lsv *LsvContext, b *MatGF) bool {
	a.MakeCanonical(lsv)
	b.MakeCanonical(lsv)
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// we do this in-place to avoid allocating a new matrix
func (m *MatGF) MakeCanonical(lsv *LsvContext) *MatGF {
	a := m.FirstNonzeroEntry()
	for i := 0; i < len(m); i++ {
		m[i] = lsv.GF.Div(m[i], a)
	}
	return m
}

func (m *MatGF) IsIdentity(lsv *LsvContext) bool {
	return m.Equal(lsv, MatGfIdentity)
}

// xxx test
func (m *MatGF) IsCanonical() bool {
	a := m.FirstNonzeroEntry()
	return a == 1
}

// func (m *MatGF) canonicalRepresentative() (b *MatGF) {
// 	b = &MatGF{}
// 	a := m.firstNonzeroEntry()
// 	for i := 0; i < len(m); i++ {
// 		b[i] = GF.Div(m[i], a)
// 	}
// 	return b
// }

func (m *MatGF) FirstNonzeroEntry() byte {
	for i := 0; i < len(m); i++ {
		if m[i] != 0 {
			return m[i]
		}
	}
	panic("MatGF is all zero")
}

// xxx test
func (m *MatGF) IsInvertible(lsv *LsvContext) bool {
	return m.Determinant(lsv) != 0
}

// xxx test
func (m *MatGF) Determinant(lsv *LsvContext) byte {
	GF := lsv.GF
	// use the formula via determinant and adjugate
	// https://en.wikipedia.org/w/index.php?title=Invertible_matrix&oldid=1148689833
 	a, b, c := m[0], m[1], m[2]
	d, e, f := m[3], m[4], m[5]
	g, h, i := m[6], m[7], m[8]
	A := GF.Add(GF.Mul(e, i), GF.Mul(f, h))
	B := GF.Add(GF.Mul(d, i), GF.Mul(f, g))
	C := GF.Add(GF.Mul(d, h), GF.Mul(e, g))
	return GF.Add(GF.Add(GF.Mul(a, A), GF.Mul(b, B)), GF.Mul(c, C))
}

func (m *MatGF) Inverse(lsv *LsvContext) (mInv *MatGF) {
	GF := lsv.GF
	// use the formula via determinant and adjugate
	// https://en.wikipedia.org/w/index.php?title=Invertible_matrix&oldid=1148689833
 	a, b, c := m[0], m[1], m[2]
	d, e, f := m[3], m[4], m[5]
	g, h, i := m[6], m[7], m[8]
	A := GF.Add(GF.Mul(e, i), GF.Mul(f, h))
	B := GF.Add(GF.Mul(d, i), GF.Mul(f, g))
	C := GF.Add(GF.Mul(d, h), GF.Mul(e, g))
	D := GF.Add(GF.Mul(b, i), GF.Mul(c, h))
	E := GF.Add(GF.Mul(a, i), GF.Mul(c, g))
	F := GF.Add(GF.Mul(a, h), GF.Mul(b, g))
	G := GF.Add(GF.Mul(b, f), GF.Mul(c, e))
	H := GF.Add(GF.Mul(a, f), GF.Mul(c, d))
	I := GF.Add(GF.Mul(a, e), GF.Mul(b, d))

	det := GF.Add(GF.Add(GF.Mul(a, A), GF.Mul(b, B)), GF.Mul(c, C))
	detInv := GF.Inv(det)

	mInv = &MatGF{
		GF.Mul(A, detInv), GF.Mul(D, detInv), GF.Mul(G, detInv),
		GF.Mul(B, detInv), GF.Mul(E, detInv), GF.Mul(H, detInv),
		GF.Mul(C, detInv), GF.Mul(F, detInv), GF.Mul(I, detInv),
	}
	mInv.MakeCanonical(lsv)
	return mInv
}

func (m *MatGF) SwapRows(i, j int) {
	m[i*MatGFStride], m[j*MatGFStride] = m[j*MatGFStride], m[i*MatGFStride]
	m[i*MatGFStride+1], m[j*MatGFStride+1] = m[j*MatGFStride+1], m[i*MatGFStride+1]
	m[i*MatGFStride+2], m[j*MatGFStride+2] = m[j*MatGFStride+2], m[i*MatGFStride+2]
}

func (m *MatGF) Less(b *MatGF) bool {
	for i := 0; i < len(m); i++ {
		if m[i] < b[i] {
			return true
		}
		if m[i] > b[i] {
			return false
		}
	}
	return false
}

// xxx test;
func (m *MatGF) Latex(lsv *LsvContext) string {
	s := "\\begin{pmatrix}\n"
	for i := 0; i < len(m); i += MatGFStride {
		for j := 0; j < MatGFStride; j++ {
			if m[i+j] == 0 {
				s += "0 & "
			} else {
				coeffs := byteToBits(m[i+j])
				// xxx unfortunately this formats the polynomial high
				// degree to low whereas in the LSV paper they are
				// formatted low degree to high.
				p := galoisfield.NewPolynomial(lsv.GF, coeffs...)
				s += fmt.Sprintf("%v & ", p)
			}
		}
		s += "\\\\\n"
	}
	s += "\\end{pmatrix}\n"
	return s
}

// xxx test
func byteToBits(b byte) []byte {
	bits := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		bits[i] = byte((b >> i) & 1)
	}
	return bits
}

func byteFromBinaryString(s string) (b byte) {
	if len(s) > 8 {
		panic("binary string too long")
	}
	for i, c := range s {
		if c != '0' && c != '1' {
			panic("invalid binary string")
		}
		if c == '1' {
			b |= 1 << uint(i)
		}
	}
	return
}

// xxx returns a string representation of the polynomial at the given
// position.
func (m *MatGF) At(lsv *LsvContext, row int, col int) string {
	p := row * MatGFStride + col
	if m[p] == 0 {
		return "0"
	}
	coeffs := byteToBits(m[p])
	poly := galoisfield.NewPolynomial(lsv.GF, coeffs...)
	//log.Printf("xxx At(%d,%d): %s\n", row, col, poly)
	return  poly.String()
}

func EnumerateMatrices(F func(*MatGF) (Continue bool)) {
	var M MatGF
	for a := 0; a < 16; a++ {
		for b := 0; b < 16; b++ {
			for c := 0; c < 16; c++ {
				for d := 0; d < 16; d++ {
					for e := 0; e < 16; e++ {
						for f := 0; f < 16; f++ {
							for g := 0; g < 16; g++ {
								for h := 0; h < 16; h++ {
									for i := 0; i < 16; i++ {
										M = MatGF{
											byte(a), byte(b), byte(c),
											byte(d), byte(e), byte(f),
											byte(g), byte(h), byte(i),
										}
										if !F(&M) {
											return
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// xxx test or deprecate
func CumulativeProducts(lsv *LsvContext, list []MatGF) []MatGF {
	if len(list) == 0 {
		panic("empty list")
	}
	products := make([]MatGF, len(list))
	products[0] = list[0]
	for i := 1; i < len(list); i++ {
		products[i] = *products[i-1].Multiply(lsv, &list[i]).MakeCanonical(lsv)
	}
	return products
}

func Product(lsv *LsvContext, list []MatGF) MatGF {
	if len(list) == 0 {
		panic("empty list")
	}
	m := list[0]
	for i := 1; i < len(list); i++ {
		m = *m.Multiply(lsv, &list[i])
	}
	return *m.MakeCanonical(lsv)
}

func RandomMatGF(lsv *LsvContext) *MatGF {
	var data [9]byte
	_, err := rand.Read(data[:])
	if err != nil {
		panic(err)
	}
	var M MatGF = data
	return NewReducedMatGF(lsv, &M)
}

func RandomInvertibleMatGF(lsv *LsvContext) *MatGF {
	for {
		M := RandomMatGF(lsv)
		if M.IsInvertible(lsv) {
			return M
		}
	}
}

// these matrices are used for testing
var c2Matrix = MatGF{
	0, 1, 0,
	1, 0, 0,
	0, 0, 1,
}

var C2Generators = []MatGF{
	c2Matrix,
}

var c3Matrix = MatGF{
	0, 0, 1,
	1, 0, 0,
	0, 1, 0,
}

var C3Generators = []MatGF{
	c3Matrix,
}

var S3Generators = []MatGF{
	c2Matrix,
	c3Matrix,
}
