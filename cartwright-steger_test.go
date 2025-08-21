package golsv

import (
	"log"
	"math/rand"
	"testing"
)

func TestCSElementFqdInverse(t *testing.T) {
	fieldElements := FqdAllElements()
	for _, a := range fieldElements {
		if a.IsZero() {
			continue
		}
		b := a.Inverse()
		c := a.Mul(b)
		if !c.IsIdentity() {
			t.Errorf("a=%v b=%v c=%v", a, b, c)
		}
		d := b.Mul(a)
		if !d.IsIdentity() {
			t.Errorf("a=%v b=%v d=%v", a, b, d)
		}
	}
}

func TestCSFqdNormalMultTable(t *testing.T) {
	// verify that w_i^2 = w_{i+1 mod 3}
	if FqdNormalMultTable[0][0] != [3]byte{0,1,0} {
		t.Errorf("FqdNormalMultTable[0][0] != {0,1,0}")
	}
	if FqdNormalMultTable[1][1] != [3]byte{0,0,1} {
		t.Errorf("FqdNormalMultTable[1][1] != {0,0,1}")
	}
	if FqdNormalMultTable[2][2] != [3]byte{1,0,0} {
		t.Errorf("FqdNormalMultTable[2][2] != {1,0,0}")
	}
	// verify that the table is symmetric
	if FqdNormalMultTable[0][1] != FqdNormalMultTable[1][0] {
		t.Errorf("FqdNormalMultTable[0][1] != FqdNormalMultTable[1][0]")
	}
	if FqdNormalMultTable[0][2] != FqdNormalMultTable[2][0] {
		t.Errorf("FqdNormalMultTable[0][2] != FqdNormalMultTable[2][0]")
	}
	if FqdNormalMultTable[1][2] != FqdNormalMultTable[2][1] {
		t.Errorf("FqdNormalMultTable[1][2] != FqdNormalMultTable[2][1]")
	}
}

func TestCSFqdElementToStandardBasis(t *testing.T) {
	all := FqdAllElements()
	found := make(map[F2Polynomial]bool)
	for _, a := range all {
		b := a.ToStandardBasis()
		found[b] = true
	}
	if len(found) != len(all) {
		t.Errorf("len(found)=%d len(all)=%d", len(found), len(all))
	}
}

func TestCSElementCalGNewFromFieldElement(t *testing.T) {
	g := NewElementCalGFromFieldElement(ElementFqd{1,1,1})
	if !g.Equal(NewElementCalGIdentity()) {
		t.Errorf("g != NewElementCalGIdentity()")
	}
}

func TestCSElementCalGNewFromString(t *testing.T) {
	tests := []struct {
		s    string
		want ElementCalG
	}{
		{"(11,0,1)(101,1,1)(101,1,0)", NewElementCalG(
			NewF2Polynomial("11"), NewF2Polynomial("0"), NewF2Polynomial("1"),
			NewF2Polynomial("101"), NewF2Polynomial("1"), NewF2Polynomial("1"),
			NewF2Polynomial("101"), NewF2Polynomial("1"), NewF2Polynomial("0")),
		},
	}
	for i, test := range tests {
		got := NewElementCalGFromString(test.s)
		if !got.Equal(test.want) {
			t.Errorf("test %d: got=%v want=%v", i, got, test.want)
		}
	}
}

func TestCSElementCalGadd(t *testing.T) {
	var zeroGR [9]F2Polynomial
	tests := []struct {
		a, b, want ElementCalG
	}{
		//		{zeroGR, zeroGR, zeroGR},
		{NewElementCalGIdentity(), NewElementCalGIdentity(), zeroGR},
		{NewElementCalGIdentity(), zeroGR, NewElementCalGIdentity()},
		{
			newElementCalGArrayNotNormalized(
				[9]F2Polynomial{
					F2PolynomialOne, F2PolynomialY, NewF2PolynomialFromSupport(2),
					NewF2PolynomialFromSupport(3), NewF2PolynomialFromSupport(4), NewF2PolynomialFromSupport(5),
					NewF2PolynomialFromSupport(6), NewF2PolynomialFromSupport(7), NewF2PolynomialFromSupport(8)}),
			newElementCalGArrayNotNormalized(
				[9]F2Polynomial{
					NewF2PolynomialFromSupport(0,2), NewF2PolynomialFromSupport(1,2), F2PolynomialZero,
					F2PolynomialZero, NewF2PolynomialFromSupport(46,100), NewF2PolynomialFromSupport(7,24),
					NewF2PolynomialFromSupport(5), NewF2PolynomialFromSupport(6,12), NewF2PolynomialFromSupport(47)}),
			newElementCalGArrayNotNormalized(
				[9]F2Polynomial{
					NewF2PolynomialFromSupport(2), NewF2PolynomialFromSupport(2), NewF2PolynomialFromSupport(2),
					NewF2PolynomialFromSupport(3), NewF2PolynomialFromSupport(4,46,100), NewF2PolynomialFromSupport(5,7,24),
					NewF2PolynomialFromSupport(5,6), NewF2PolynomialFromSupport(6,7,12), NewF2PolynomialFromSupport(8,47)}),
		},
	}
	for i, test := range tests {
		got := NewElementCalGIdentity()
		got.add(test.a, test.b)
		if !got.Equal(test.want) {
			t.Errorf("test %d: a: %v\nb: %v\ngot: %v\nwant: %v", i, test.a, test.b, got.Dump(), test.want.Dump())
		}
	}
}

func TestCSElementCalGCopy(t *testing.T) {
	tests := []struct {
		from, to ElementCalG
	}{
		{NewElementCalGIdentity(), NewElementCalGIdentity()},
		{
			NewElementCalG(
				F2PolynomialOne, F2PolynomialY, NewF2PolynomialFromSupport(2),
				NewF2PolynomialFromSupport(3), NewF2PolynomialFromSupport(4), NewF2PolynomialFromSupport(5),
				NewF2PolynomialFromSupport(6), NewF2PolynomialFromSupport(7), NewF2PolynomialFromSupport(8)),
			NewElementCalGIdentity(),
		},
		{
			// w_2 z^0
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero),
			NewElementCalGIdentity(),
		},
		{
			// w_2 z^0
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero),
			// w_0 z^0 + w_1 z^0 + w_2 z^0
			NewElementCalG(
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero),
		},
	}
	for i, test := range tests {
		test.to.Copy(test.from)
		if !test.to.Equal(test.from) {
			t.Errorf("test %d: from: %v\nto: %v", i, test.from, test.to)
		}
		if !test.from.Equal(test.to) {
			t.Errorf("test %d: from: %v\nto: %v", i, test.from, test.to)
		}
	}
}

func TestCSElementCalGScale(t *testing.T) {
	g := NewElementCalG(
		F2PolynomialOne, F2PolynomialZero, NewF2PolynomialFromSupport(2),
		NewF2PolynomialFromSupport(3,4), F2PolynomialY, F2PolynomialZero,
		NewF2PolynomialFromSupport(6), NewF2PolynomialFromSupport(7), NewF2PolynomialFromSupport(8))
	p := NewF2PolynomialFromSupport(0, 1, 3)
	g.scale(p)
	want := NewElementCalG(
		NewF2PolynomialFromSupport(0, 1, 3), F2PolynomialZero, NewF2PolynomialFromSupport(2, 3, 5),
		NewF2PolynomialFromSupport(3, 5, 6, 7), NewF2PolynomialFromSupport(1, 2, 4), F2PolynomialZero,
		NewF2PolynomialFromSupport(6,7,9), NewF2PolynomialFromSupport(7, 8, 10), NewF2PolynomialFromSupport(8, 9, 11))
	if !g.Equal(want) {
		t.Errorf("g != want")
	}
}

func TestCSElementCalGcomputeBasisProduct(t *testing.T) {
	tests := []struct {
		i, j, k, l int
		want ElementCalG
	}{
		// spot checks
		//
		// w_0 z^0 * w_0 z^0 = w_0^2 z^0 = w_1 z^0
		{
			0, 0, 0, 0,
			newElementCalGArrayNotNormalized([...]F2Polynomial{
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero}),
		},
		// w_0 z^1 * w_0 z^0 = w_0 w_1 z^1 z^0
		//                           = (w_0 + w_2) z^1
		//                           = w_0 z^1 + w_2 z^1
		{
			0, 1, 0, 0,
			newElementCalGArrayNotNormalized([...]F2Polynomial{
				F2PolynomialZero, F2PolynomialOne, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialOne, F2PolynomialZero}),
		},
		// w_0 z^2 * w_0 z^0 = w_0 w_2 z^2
		//						     = (w_1 + w_2) z^2
		//						     = w_1 z^2 + w_2 z^2
		{
			0, 2, 0, 0,
			newElementCalGArrayNotNormalized([...]F2Polynomial{
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialOne,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialOne}),
		},
		// w_0 z^2 * w_1 z^1 = w_0 w_0 z^3
		//						     = (1+y)w_1 z^0
		{
			0, 2, 1, 1,
			newElementCalGArrayNotNormalized([...]F2Polynomial{
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				NewF2PolynomialFromSupport(0, 1), F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero}),
		},
		// w_0 z^2 * w_0 z^1 = w_0 w_2 z^3 = (1+y)(w_1 + w_2) z^0
		{
			0, 2, 0, 1,
			newElementCalGArrayNotNormalized([...]F2Polynomial{
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				NewF2PolynomialFromSupport(0, 1), F2PolynomialZero, F2PolynomialZero,
				NewF2PolynomialFromSupport(0, 1), F2PolynomialZero, F2PolynomialZero}),
		},
	}
	for i, test := range tests {
		got := computeCalGBasisProduct(test.i, test.j, test.k, test.l)
		if !got.Equal(test.want) {
			t.Errorf("test %d: got: %v, want: %v", i, got, test.want)
		}
	}
}

func TestCSElementCalGIdentityMul(t *testing.T) {
	g := NewElementCalG(
		F2PolynomialOne, F2PolynomialZero, NewF2PolynomialFromSupport(2),
		NewF2PolynomialFromSupport(3,4), F2PolynomialY, F2PolynomialZero,
		NewF2PolynomialFromSupport(6), NewF2PolynomialFromSupport(7), NewF2PolynomialFromSupport(8))
	one := NewElementCalGIdentity()
	prod := NewElementCalGIdentity()
	prod.Mul(g, one)
	if !prod.Equal(g) {
		t.Errorf("prod != g\ng: %v\none: %v\nprod: %v", g, one, prod)
	}
}

func TestCSElementCalGMultZ(t *testing.T) {
	// z = 1_{F_q^d} * z = (w_0 + w_1 + w_2) * z
	z := NewElementCalG(
		F2PolynomialZero, F2PolynomialOne, F2PolynomialZero,
		F2PolynomialZero, F2PolynomialOne, F2PolynomialZero,
		F2PolynomialZero, F2PolynomialOne, F2PolynomialZero)
	// compute z^3
	a := NewElementCalGIdentity()
	a.Mul(z, z)
	b := NewElementCalGIdentity()
	b.Mul(a, z)
	// check that z^3 = 1+y
	//                = 1+y (w_0 + w_1 + w_2) z^0
	//
	// actually, projectively this product is simply 1; the
	// constructor normalizes its input, so the test succeeds.
	want := NewElementCalG(
		F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialZero,
		F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialZero,
		F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialZero)
	if !b.Equal(want) {
		t.Errorf("b != want")
	}
}

func TestCSElementCalGEmbeddedFieldMult(t *testing.T) {
	// verify that each field element, embedded in CalG, has order 7.
	fieldElements := FqdAllElements()
	if len(fieldElements) != 8 {
		t.Errorf("len(fieldElements) != 8")
	}
	for _, u := range fieldElements {
		if u.IsZero() {
			continue
		}
		g_u := NewElementCalGFromFieldElement(u)
		order := g_u.Order()
		if u.IsIdentity() {
			if order != 1 {
				t.Errorf("u.IsIdentity() && order != 1; u: %v", u)
			}
		} else {
			if order != 7 {
				t.Errorf("u.Order() != 7; u: %v", u)
			}
		}
	}
}

func TestCSElementCalGEmbeddedFieldInverses(t *testing.T) {
	for _, u := range FqdAllElements() {
		if u.IsZero() {
			continue
		}
		g_u := NewElementCalGFromFieldElement(u)
		uInv := u.Inverse()
		g_uInv := NewElementCalGFromFieldElement(uInv)
		p := NewElementCalGIdentity()
		p.Mul(g_u, g_uInv)
		if !p.IsIdentity() {
			t.Errorf("p != identity\ng_u: %v\ng_uInv: %v", g_u, g_uInv)
		}
	}
}

func TestCSElementCalGIsIdentity(t *testing.T) {
	g := NewElementCalGIdentity()
	if !g.IsIdentity() {
		t.Errorf("id.IsIdentity() != true")
	}
	g.zero()
	if g.IsIdentity() {
		t.Errorf("g.IsIdentity() != false")
	}
}

func TestCSElementCalGIsIdentityModf(t *testing.T) {
	tests := []struct {
		g    ElementCalG
		f	F2Polynomial
		want bool
	}{
		{
			NewElementCalGIdentity(),
			F2Polynomial111,
			true,
		},
		{
			NewElementCalGFromString("(1,0,0)(011,0,0)(1,0,0)"),
			F2Polynomial111,
			true,
		},
		{
			NewElementCalGFromString("(011,0,0)(011,0,0)(1,0,0)"),
			F2Polynomial111,
			true,
		},
		{
			NewElementCalGFromString("(011,0,1)(011,0,0)(1,0,0)"),
			F2Polynomial111,
			false,
		},
		// xxx some tests temporarily disabled; normalizeModf only supports the case
		// where f = 1+y+y^2.
// 		{
// 			NewElementCalGIdentity(),
// 			F2PolynomialY,
// 			true,
// 		},
// 		{
// 			NewElementCalGIdentity(),
// 			NewF2PolynomialFromSupport(1,3,7),
// 			true,
// 		},
// 		{
// 			NewElementCalG(
// 				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
// 				F2PolynomialY, F2PolynomialZero, F2PolynomialZero,
// 				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
// 			NewF2PolynomialFromSupport(1,3,7),
// 			false,
// 		},
// 		{
// 			NewElementCalG(
// 				NewF2PolynomialFromSupport(0,1,3,7), F2PolynomialZero, F2PolynomialZero,
// 				NewF2PolynomialFromSupport(0,1,3,7), F2PolynomialZero, F2PolynomialZero,
// 				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero),
// 			NewF2PolynomialFromSupport(1,3,7),
// 			true,
// 		},
	}
	for i, test := range tests {
		got := test.g.IsIdentityModf(test.f)
		if got != test.want {
			t.Errorf("test %d:\ng: %v got: %v want: %v", i, test.g, got, test.want)
		}
	}
}

func TestCSElementCalGnormalizeModf(t *testing.T) {
	tests := []struct {
		g    ElementCalG
		f	F2Polynomial
		want ElementCalG
	}{
		{
			NewElementCalGFromString("(01,0,11)(01,0,11)(01,0,11)"),
			F2Polynomial111,
			NewElementCalGFromString("(1,0,01)(1,0,01)(1,0,01)"),
		},
	}
	for i, test := range tests {
		got := test.g.normalizeModf(test.f)
		if !got.Equal(test.want) {
			t.Errorf("test %d:\ng: %v\nf: %v\ngot: %v\nwant: %v", i, test.g, test.f, got, test.want)
		}
	}
}

func TestCSElementCalGnormalize(t *testing.T) {
	tests := []struct {
		g    ElementCalG
		want ElementCalG
	}{
		{
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
		},
		{
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialY, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
		},
		{
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
		},
		{
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				NewF2PolynomialFromSupport(1,2), F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
		},
		{
			NewElementCalG(
				F2PolynomialY, NewF2PolynomialFromSupport(2), NewF2PolynomialFromSupport(3),
				NewF2PolynomialFromSupport(4), NewF2PolynomialFromSupport(5), NewF2PolynomialFromSupport(6),
				NewF2PolynomialFromSupport(7), NewF2PolynomialFromSupport(8), NewF2PolynomialFromSupport(9)),
			NewElementCalG(
				F2PolynomialOne, F2PolynomialY, NewF2PolynomialFromSupport(2),
				NewF2PolynomialFromSupport(3), NewF2PolynomialFromSupport(4), NewF2PolynomialFromSupport(5),
				NewF2PolynomialFromSupport(6), NewF2PolynomialFromSupport(7), NewF2PolynomialFromSupport(8)),
		},
		{
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, NewF2PolynomialFromSupport(3,5),
				NewF2PolynomialFromSupport(2,3), NewF2PolynomialFromSupport(3,4), NewF2PolynomialFromSupport(4,5),
				NewF2PolynomialFromSupport(2,3), NewF2PolynomialFromSupport(6,7), NewF2PolynomialFromSupport(9,10)),
			NewElementCalG(
				F2PolynomialZero, F2PolynomialZero, NewF2PolynomialFromSupport(1,2),
				F2PolynomialOne, F2PolynomialY, NewF2PolynomialFromSupport(2),
				F2PolynomialOne, NewF2PolynomialFromSupport(4), NewF2PolynomialFromSupport(7)),
		},
	}
	for i, test := range tests {
		got := test.g.normalize()
		if !got.Equal(test.want) {
			t.Errorf("test %d:\ng: %v\ngot: %v\nwant: %v", i, test.g, got, test.want)
		}
	}
}

func TestCSElementCalGFrobenius(t *testing.T) {
	// set w = w_0 z^0
	w := NewElementCalG(
		F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
		F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
		F2PolynomialZero, F2PolynomialZero, F2PolynomialZero)
	// check that w^2 = w_1
	a := NewElementCalGIdentity()
	a.Mul(w, w)
	if !a.Equal(
		NewElementCalG(
			F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialZero, F2PolynomialZero, F2PolynomialZero)) {
		t.Errorf("w^2 != w_1")
	}
	// check that w^4 = w_2
	b := NewElementCalGIdentity()
	b.Mul(a, a)
	if !b.Equal(
		NewElementCalG(
			F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialOne, F2PolynomialZero, F2PolynomialZero)) {
		t.Errorf("w^4 != w_2")
	}
	// check that w^8 = w_0
	c := NewElementCalGIdentity()
	c.Mul(b, b)
	if !c.Equal(
		NewElementCalG(
			F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialZero, F2PolynomialZero, F2PolynomialZero)) {
		t.Errorf("w^8 != w_0")
	}
}

func TestCSElementCalGAsMapKey(t *testing.T) {
	g := NewElementCalGIdentity()
	h := NewElementCalGIdentity()
	m := make(map[ElementCalG]bool)
	m[g] = true
	if !m[h] {
		t.Errorf("m[h] != true")
	}

	// xxx test whether reduction mod f results in the same map key -
	// fear it doesn't, perhaps due to underlying F2Polynomial string
	// not being trimmed?
}

func TestCSElementCalGGenB(t *testing.T) {
	b := cartwrightStegerGenB()
	bInv := cartwrightStegerGenBInv()
	c := NewElementCalGIdentity()
	c.Mul(b, bInv)
	if !c.IsIdentity() {
		t.Errorf("expected identity for b * bInv; got:\n%v", c)
	}
}

func TestCSElementCalGLatex(t *testing.T) {
	tests := []struct {
		g    ElementCalG
		want string
	}{
		{
			NewElementCalGIdentity(),
			"\\zeta_0 + \\zeta_1 + \\zeta_2",
		},
		{
			cartwrightStegerGenB(),
			"(1+y) \\zeta_0 + \\zeta_0 z^{2} + (1+y) \\zeta_1 + \\zeta_1 z^{2} + (1+y) \\zeta_2 + \\zeta_2 z^{2}",
		},
	}
	for i, test := range tests {
		got := test.g.Latex()
		if got != test.want {
			t.Errorf("test %d:\ng: %v\ngot:  %v\nwant: %v", i, test.g, got, test.want)
		}
	}
}

func TestCSElementCalGModf(t *testing.T) {
	tests := []struct {
		g   ElementCalG
		f 	F2Polynomial
		want ElementCalG
	}{
// xxx temporarily disabled; normalizeModf only supports the case
// where f = 1+y+y^2.
// 		{
// 			NewElementCalG(
// 				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
// 				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
// 				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
// 			F2PolynomialY,
// 			NewElementCalG(
// 				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero,
// 				F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
// 				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
// 		},
		{
			NewElementCalG(
				NewF2PolynomialFromSupport(0,1,2), NewF2PolynomialFromSupport(2), NewF2PolynomialFromSupport(3),
				F2PolynomialOne, F2PolynomialZero, NewF2PolynomialFromSupport(3,4,7),
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
			NewF2PolynomialFromSupport(0,1,2),
			NewElementCalG(
				F2PolynomialZero, F2PolynomialOne, F2PolynomialY,
				F2PolynomialY, F2PolynomialZero, F2PolynomialY,
				F2PolynomialZero, F2PolynomialZero, F2PolynomialZero),
		},
	}
	for i, test := range tests {
		got := test.g.Modf(test.f)
		if !got.Equal(test.want) {
			t.Errorf("test %d:\ng: %v\nf: %v\ngot: %v\nwant: %v", i, test.g, test.f, got, test.want)
		}
	}
		
		
}

func TestCSCartwrightStegerGenerators(t *testing.T) {
	gens := CartwrightStegerGenerators()
	numExpected := 14
	if len(gens) != numExpected {
		t.Errorf("len(gens)=%d expected=%d", len(gens), numExpected)
	}
	seen := make(map[ElementCalG]bool)
	for _, g := range gens {
		if seen[g] {
			t.Errorf("duplicate generator")
		}
		seen[g] = true
	}
}

func TestCSCartwrightStegerGeneratorsInverse(t *testing.T) {
	gens := CartwrightStegerGenerators()
	for i, g := range gens {
		gInv := CartwrightStegerGeneratorsInverse(gens, i)
		p := NewElementCalGIdentity()
		p.Mul(g, gInv)
		if !p.IsIdentity() {
			t.Errorf("i: %d\ng: %v\ngInv: %v\ngot: %v\nwant: identity", i, g, gInv, p)
		}
	}
}

func TestCSCartwrightStegerGeneratorsOrderModf(t *testing.T) {
	gens := CartwrightStegerGenerators()
	f := F2Polynomial111
	expectedOrder := 21
	for i, g := range gens {
		g = g.Modf(f)
		order := g.OrderModf(f)
		if order != expectedOrder {
			t.Errorf("i: %d\ng: %v\norder: %d", i, g, order)
		}
	}
}

// xxx
func DisabledTestCSCartwrightStegerExperiment(t *testing.T) {
	gens := CartwrightStegerGenerators()
	g := NewElementCalGIdentity()
	tmp := NewElementCalGIdentity()
	for i := 0; i < 1000; i++ {
		h := gens[rand.Intn(len(gens))]
		tmp.Mul(g, h)
		g.Copy(tmp)
		log.Printf("g = %v", g)
	}
}

// xxx debugging 
func TestCSCosetRep(t *testing.T) {
	g := cartwrightStegerGenB()
	h := NewElementCalG(
		NewF2Polynomial("1011"), NewF2Polynomial("111"), F2PolynomialZero,
		NewF2Polynomial("1011"), F2PolynomialZero, F2PolynomialZero,
		NewF2Polynomial("0101"), NewF2Polynomial("111"), NewF2Polynomial("111"))
	p := NewElementCalGIdentity()
	p.Mul(g, h)

	modulus := NewF2Polynomial("111")
	gModf := g.Modf(modulus)
	hModf := h.Modf(modulus)
	pModf := p.Modf(modulus)

	q := NewElementCalGIdentity()
	q.Mul(gModf, hModf)
	qModf := q.Modf(modulus)

	if !gModf.Equal(pModf) {
		log.Printf("g:     %v", g)
		log.Printf("h:     %v", h)
		log.Printf("p:     %v", p)
		log.Printf("gModf: %v", gModf)
		log.Printf("hModf: %v", hModf)
		log.Printf("pModf: %v", pModf)
		log.Printf("q:     %v", q)
		log.Printf("qModf: %v", qModf)
		t.Errorf("gModf != pModf")
	}
	// check whether g and p are in the same coset of by checking
	// whether g^{-1} * p is the identity mod f
	gInv := cartwrightStegerGenBInv()
	s := NewElementCalGIdentity()
	s.Mul(gInv, p)
	if !s.IsIdentityModf(modulus) {
		t.Errorf("g and p are not in the same coset")
	}
}

// experimenting with calculating the dim d^2 matrix representation of
// the Cartwright-Steger group. the output can be manually checked
// against the table in the paper. xxx not sure if this currently
// matches the table in the paper; and we don't currently use this
// representation anyway.
func DisabledTestCSMatrixRepD2(t *testing.T) {
	gens := CartwrightStegerGenerators()
	p := NewElementCalGIdentity()
	q := NewElementCalGIdentity()
	for k, b_u := range gens {
		b_uInv := CartwrightStegerGeneratorsInverse(gens, k)
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				var array [9]F2Polynomial
				array[i*3+j] = F2PolynomialOne
				h := newElementCalGArrayNotNormalized(array)
				p.Mul(b_u, h)
				q.Mul(p, b_uInv)
			}
		}
	}
}

func TestCSMatrixRepFieldElement(t *testing.T) {
	id := F2PolynomialOne
	mId := cartwrightStegerMatrixRepFieldElement(id)
	wantMId := MatF2PolyIdentity
	if !mId.Equal(wantMId) {
		t.Errorf("mId=%v want=%v", mId, wantMId)
	}
	// quick check: powers of a non-identity element should generate
	// the whole multiplicative group of F_8.
	u := NewF2Polynomial("110")
	mU := cartwrightStegerMatrixRepFieldElement(u)
	v := mId
	for i := 0; i <= 7; i++ {
		if v.Equal(MatF2PolyIdentity) {
			if i != 0 && i != 7 {
				t.Errorf("unexpected identity at i=%d", i)
			}
		} else {
			if i == 7 {
				t.Errorf("unexpected non-identity at i=%d", i)
			}
		}
		v = v.Mul(mU)
	}
}

func TestCSMatrixRepBeta(t *testing.T) {
	beta := cartwrightStegerEmbeddingBeta
	wantBeta := NewF2Polynomial("11")
	if !beta.Equal(wantBeta) {
		t.Errorf("beta=%v want=%v", beta, wantBeta)
	}
	// let lambda be the linear map on F_8 defined by multiplication
	// by beta.
	lambda := cartwrightStegerMatrixRepFieldElement(beta)
	wantLambda := NewMatF2PolyFromString("[1 0 1 1 1 1 0 1 1]")
	if !lambda.Equal(wantLambda) {
		t.Errorf("lambda=%v want=%v", lambda, wantLambda)
	}
	tr := lambda.Trace()
	if tr.Equal(F2PolynomialZero) {
		t.Errorf("trace(lambda)=%v want=nonzero", tr)
	}
}

func TestCSMatrixRepOnePlusBetaX(t *testing.T) {
	mat := cartwrightStegerMatrixRepOnePlusBetaX()
	wantMat := NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]")
	if !mat.Equal(wantMat) {
		t.Errorf("mat=%v want=%v", mat, wantMat)
	}
}

func TestCSMatrixRepOneTensorPhi(t *testing.T) {
	mat := cartwrightStegerMatrixRepOneTensorPhi()
	wantMat := NewMatF2PolyFromString("[1 0 0 0 0 1 0 1 1]")
	if !mat.Equal(wantMat) {
		t.Errorf("mat=%v want=%v", mat, wantMat)
	}
}

func TestCSMatrixRepZ(t *testing.T) {
	zMat := cartwrightStegerMatrixRepZ()
	wantZMat := NewMatF2PolyFromString("[11 01 01 01 01 1 0 11 1]")
	if !zMat.Equal(wantZMat) {
		t.Errorf("zMat=%v want=%v", zMat, wantZMat)
	}
	// verify that z^3 = 1 + y(x)
	// where y(x) = x + x^3
	y := NewF2Polynomial("0101")
	onePlusY := MatF2PolyIdentity.Add(MatF2PolyIdentity.Scale(y))
	zCubed := zMat.Pow(3)
	if !zCubed.Equal(onePlusY) {
		t.Errorf("zCubed=%v want=%v", zCubed, onePlusY)
	}
}

func TestCSMatrixRepB(t *testing.T) {
	bMat := cartwrightStegerMatrixRepB()
	wantBMat := NewProjMatF2PolyFromString("[0101 001 011 01 0001 111 011 101 1001]")
	if !bMat.Equal(wantBMat) {
		t.Errorf("bMat=%v want=%v", bMat, wantBMat)
	}
}

func TestCSMatrixRepBInv(t *testing.T) {
	bMat := cartwrightStegerMatrixRepB()
	bInvMat := cartwrightStegerMatrixRepBInverse()
	c := bMat.Mul(bInvMat)
	if !c.Equal(ProjMatF2Poly(MatF2PolyIdentity)) {
		t.Errorf("c=%v want=I", c)
	}
}

func TestCSGeneratorsMatrixReps(t *testing.T) {
	gens, _ := CartwrightStegerGeneratorsMatrixReps()
	want := 14
	if len(gens) != want {
		t.Errorf("len(gens)=%d want=%d", len(gens), want)
	}
	for i := 0; i < len(gens); i += 2 {
		g := gens[i]
		gInv := gens[i+1]
		c := g.Mul(gInv)
		if !c.Equal(ProjMatF2Poly(MatF2PolyIdentity)) {
			t.Errorf("c=%v want=I", c)
		}
	}
}

func TestCSCartwrightStegerEmbedPolynomial(t *testing.T) {
	tests := []struct {
		name   string
		input  F2Polynomial
		want   F2Polynomial
	}{
		{
			name:  "zero polynomial",
			input: F2PolynomialZero,
			want:  F2PolynomialZero,
		},
		{
			name:  "constant 1",
			input: F2PolynomialOne,
			want:  F2PolynomialOne,
		},
		{
			name:  "y (degree 1)",
			input: F2PolynomialY,
			want:  cartwrightStegerEmbeddingY, // x + x^3
		},
		{
			name:  "1 + y",
			input: F2PolynomialOnePlusY,
			want:  F2PolynomialOne.Add(cartwrightStegerEmbeddingY), // 1 + x + x^3
		},
		{
			name:  "y^2",
			input: NewF2Polynomial("001"),
			want:  cartwrightStegerEmbeddingY.Pow(2), // (x + x^3)^2
		},
		{
			name:  "y^3",
			input: NewF2Polynomial("0001"),
			want:  cartwrightStegerEmbeddingY.Pow(3), // (x + x^3)^3
		},
		{
			name:  "1 + y + y^2",
			input: NewF2Polynomial("111"),
			want:  F2PolynomialOne.Add(cartwrightStegerEmbeddingY).Add(cartwrightStegerEmbeddingY.Pow(2)),
		},
		{
			name:  "y + y^3",
			input: NewF2Polynomial("0101"),
			want:  cartwrightStegerEmbeddingY.Add(cartwrightStegerEmbeddingY.Pow(3)),
		},
		{
			name:  "polynomial with higher degree",
			input: NewF2Polynomial("10101"),
			want:  F2PolynomialOne.Add(cartwrightStegerEmbeddingY.Pow(2)).Add(cartwrightStegerEmbeddingY.Pow(4)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cartwrightStegerEmbedPolynomial(tt.input)
			if !got.Equal(tt.want) {
				t.Errorf("cartwrightStegerEmbedPolynomial(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCSCartwrightStegerEmbedPolynomialLinearity(t *testing.T) {
	testPolys := []F2Polynomial{
		F2PolynomialOne,
		F2PolynomialY,
		NewF2Polynomial("001"), // y^2
		NewF2Polynomial("111"), // 1 + y + y^2
	}

	for i, a := range testPolys {
		for j, b := range testPolys {
			if i >= j {
				continue // avoid duplicate tests
			}
			sum := a.Add(b)
			embeddedSum := cartwrightStegerEmbedPolynomial(sum)
			embeddedA := cartwrightStegerEmbedPolynomial(a)
			embeddedB := cartwrightStegerEmbedPolynomial(b)
			expectedSum := embeddedA.Add(embeddedB)

			if !embeddedSum.Equal(expectedSum) {
				t.Errorf("linearity test failed: embed(%v + %v) = %v, but embed(%v) + embed(%v) = %v", 
					a, b, embeddedSum, a, b, expectedSum)
			}
		}
	}
}
