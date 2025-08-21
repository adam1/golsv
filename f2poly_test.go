package golsv

import (
	"testing"
)

// xxx DEPRECATED
// func TestF2PolynomialNewFromCoefficients(t *testing.T) {
// 	tests := []struct {
// 		coefficients []byte
// 		want string
// 	}{
// 		{[]byte{}, "0"},
// 		{[]byte{0}, "0"},
// 		{[]byte{1}, "1"},
// 		{[]byte{1,1}, "1+y"},
// 		{[]byte{1,1,1}, "1+y+y^2"},
// 		{[]byte{0,1,0,1}, "y^2+y"},
// 	}
// 	for n, test := range tests {
// 		f := NewF2PolynomialFromCoefficients(test.coefficients)
// 		if f.String() != test.want {
// 			t.Errorf("test %d: f != want", n)
// 		}
// 	}
// }

func TestF2PolynomialAdd(t *testing.T) {
	tests := []struct {
		a, b, want F2Polynomial
	}{
		{F2PolynomialZero, F2PolynomialZero, F2PolynomialZero},
		{F2PolynomialOne, F2PolynomialOne, F2PolynomialZero},
		{F2PolynomialOne, F2PolynomialZero, F2PolynomialOne},
		{F2PolynomialZero, F2PolynomialOne, F2PolynomialOne},
		{F2PolynomialOne, F2PolynomialY, F2PolynomialOnePlusY},
		{NewF2Polynomial("101"), F2PolynomialY, NewF2Polynomial("111")},
		{NewF2Polynomial("1011"), NewF2Polynomial("001000000000000001"), NewF2Polynomial("100100000000000001")},
	}
	for n, test := range tests {
		c := test.a.Add(test.b)
		if !c.Equal(test.want) {
			t.Errorf("test %d: a=(%v) b=(%v) got=(%v) want=(%v)", n, test.a, test.b, c, test.want)
		}
	}
}

func TestF2PolynomialAddMonomial(t *testing.T) {
	tests := []struct {
		f F2Polynomial
		m int
		want F2Polynomial
	}{
		{F2PolynomialOne, 0, F2PolynomialZero},
		{F2PolynomialOne, 1, F2PolynomialOnePlusY},
		{F2PolynomialOne, 2, NewF2Polynomial("101")},
		{F2PolynomialOnePlusY, 0, F2PolynomialY},
		{F2PolynomialOnePlusY, 1, F2PolynomialOne},
		{F2PolynomialOnePlusY, 17, NewF2Polynomial("110000000000000001")},
	}
	for n, test := range tests {
		got := test.f.AddMonomial(test.m)
		if !got.Equal(test.want) {
			t.Errorf("test %d:\nf: %v\nm: %v\ngot: %v\nwant: %v", n, test.f, test.m, got, test.want)
		}
	}
}

func TestF2PolynomialCoefficient(t *testing.T) {
	tests := []struct {
		f F2Polynomial
		m int
		want int
	}{
		{F2PolynomialOne, 0, 1},
		{F2PolynomialOne, 1, 0},
		{F2PolynomialOne, 2, 0},
		{F2PolynomialOnePlusY, 0, 1},
		{F2PolynomialOnePlusY, 1, 1},
		{F2PolynomialOnePlusY, 2, 0},
		{NewF2Polynomial("101"), 0, 1},
		{NewF2Polynomial("101"), 1, 0},
		{NewF2Polynomial("101"), 2, 1},
	}
	for n, test := range tests {
		got := test.f.Coefficient(test.m)
		if got != test.want {
			t.Errorf("test %d: f: %v\nm: %v\ngot: %v\nwant: %v", n, test.f, test.m, got, test.want)
		}
	}
}

func TestF2PolynomialDegree(t *testing.T) {
	tests := []struct {
		f F2Polynomial
		want int
	}{
		{F2PolynomialZero, -1},
		{F2PolynomialOne, 0},
		{F2PolynomialY, 1},
		{F2PolynomialOnePlusY, 1},
		{NewF2Polynomial("111"), 2},
		{NewF2Polynomial("0101"), 3},
	}
	for n, test := range tests {
		if test.f.Degree() != test.want {
			t.Errorf("test %d: f.Degree() != want", n)
		}
	}
}

func TestF2PolynomialDiv(t *testing.T) {
	tests := []struct {
		p, f, wantQuotient, wantRemainder F2Polynomial
	}{
		{F2PolynomialOne, F2PolynomialOne, F2PolynomialOne, F2PolynomialZero},
		{F2PolynomialOne, F2PolynomialY, F2PolynomialZero, F2PolynomialOne},
		{F2PolynomialY, F2PolynomialY, F2PolynomialOne, F2PolynomialZero},
		{NewF2Polynomial("111"), F2PolynomialOnePlusY, F2PolynomialY, F2PolynomialOne},
		{NewF2Polynomial("11011"), NewF2Polynomial("11001"), F2PolynomialOne, NewF2Polynomial("0001")},

		// former DivY(n) tests
		{F2PolynomialOne, F2PolynomialY.Pow(0), F2PolynomialOne, F2PolynomialZero},
		{F2PolynomialY, F2PolynomialY.Pow(0), F2PolynomialY, F2PolynomialZero},
		{F2PolynomialY, F2PolynomialY.Pow(1), F2PolynomialOne, F2PolynomialZero},
		{F2PolynomialOnePlusY, F2PolynomialY.Pow(0), F2PolynomialOnePlusY, F2PolynomialZero},
		{NewF2Polynomial("011"), F2PolynomialY.Pow(0), NewF2Polynomial("011"), F2PolynomialZero},
		{NewF2Polynomial("011"), F2PolynomialY.Pow(1), F2PolynomialOnePlusY, F2PolynomialZero},
		{NewF2Polynomial("00100000000000000001"), F2PolynomialY.Pow(2), NewF2Polynomial("100000000000000001"), F2PolynomialZero},

		// former Div1PlusY() tests
		{F2PolynomialOne, F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialOne},
		{F2PolynomialOnePlusY, F2PolynomialOnePlusY, F2PolynomialOne, F2PolynomialZero},
		{NewF2PolynomialFromSupport(0,1,2), F2PolynomialOnePlusY, F2PolynomialY, F2PolynomialOne},
		{NewF2PolynomialFromSupport(0,3), F2PolynomialOnePlusY, NewF2PolynomialFromSupport(0,1,2), F2PolynomialZero},
		{NewF2PolynomialFromSupport(0,1,3,4), F2PolynomialOnePlusY, NewF2PolynomialFromSupport(0,3), F2PolynomialZero},
		{NewF2PolynomialFromSupport(1,2,3,5), F2PolynomialOnePlusY, NewF2PolynomialFromSupport(1,3,4), F2PolynomialZero},

	}
	for n, test := range tests {
		quotient, remainder := test.p.Div(test.f)
		if !quotient.Equal(test.wantQuotient) {
			t.Errorf("test %d: (%v)/(%v) want quotient=(%v) got quotient=(%v)", n, test.p, test.f, test.wantQuotient, quotient)
		}
		if !remainder.Equal(test.wantRemainder) {
			t.Errorf("test %d: (%v)/(%v) want remainder=(%v) got remainder=(%v)", n, test.p, test.f, test.wantRemainder, remainder)
		}
	}
}

func TestF2PolynomialDivRand(t *testing.T) {
	trials := 10
	maxDegree := 100
	for i := 0; i < trials; i++ {
		p := NewF2PolynomialRandom(maxDegree)
		f := NewF2PolynomialRandom(maxDegree)
		if f.IsZero() {
			continue
		}
		quotient, remainder := p.Div(f)
		sum := quotient.Mul(f).Add(remainder)
		if !sum.Equal(p) {
			t.Errorf("(%v)/(%v) = (%v),(%v) but quotient * divisor + remainder = (%v)",
				 p, f, quotient, remainder, sum)
		}
	}
}

func TestF2PolynomialDup(t *testing.T) {
	p := NewF2PolynomialFromSupport(0,1,2)
	q := p.Dup()
	q.AddMonomial(0)
	if !p.Equal(NewF2PolynomialFromSupport(0,1,2)) {
		t.Errorf("p != (0,1,2)")
	}
}

func TestF2PolynomialEqual(t *testing.T) {
	tests := []struct {
		a, b F2Polynomial
		want bool
	}{
		{F2PolynomialOne, F2PolynomialOne, true},
		{F2PolynomialOne, F2PolynomialY, false},
		{F2PolynomialY, F2PolynomialOne, false},
		{F2PolynomialY, F2PolynomialY, true},
		{NewF2PolynomialFromSupport(1, 2), NewF2PolynomialFromSupport(1, 2), true},
		{NewF2PolynomialFromSupport(1, 2), NewF2PolynomialFromSupport(2, 1), true},
		{NewF2PolynomialFromSupport(1, 2), NewF2PolynomialFromSupport(1, 2, 3), false},
		{NewF2PolynomialFromSupport(1, 2, 3), NewF2PolynomialFromSupport(1, 2), false},
		{F2PolynomialZero, F2PolynomialZero, true},
	}
	for n, test := range tests {
		if got := test.a.Equal(test.b); got != test.want {
			t.Errorf("test %d:\na: %v\nb: %v\ngot: %v\nwant: %v", n, test.a, test.b, got, test.want)
		}
	}
}

func TestF2PolynomialInverseModf(t *testing.T) {
	tests := []struct {
		a, f, want F2Polynomial
	}{
		{F2PolynomialOne, F2Polynomial111, F2PolynomialOne},
		{F2PolynomialY, F2Polynomial111, F2PolynomialOnePlusY},
		{F2PolynomialOnePlusY, F2Polynomial111, F2PolynomialY},
		{F2PolynomialOne, NewF2Polynomial("1101"), F2PolynomialOne},
		{F2PolynomialY, NewF2Polynomial("1101"), NewF2Polynomial("101")},
		{F2PolynomialOnePlusY, NewF2Polynomial("1101"), NewF2Polynomial("011")},
		{NewF2Polynomial("001"), NewF2Polynomial("1101"), NewF2Polynomial("111")},
		{NewF2Polynomial("101"), NewF2Polynomial("1101"), F2PolynomialY},
		{NewF2Polynomial("011"), NewF2Polynomial("1101"), F2PolynomialOnePlusY},
		{NewF2Polynomial("111"), NewF2Polynomial("1101"), NewF2Polynomial("001")},
	}
	for n, test := range tests {
		got := test.a.InverseModf(test.f)
		if !got.Equal(test.want) {
			t.Errorf("test %d: got %v want %v", n, got, test.want)
		}
		p := test.a.Mul(got).Modf(test.f)
		if !p.Equal(F2PolynomialOne) {
			t.Errorf("test %d: inverse failed", n)
		}
	}
	moduli := []F2Polynomial { NewF2Polynomial("111"), NewF2Polynomial("1101") }
	for _, f := range moduli {
		polys := EnumerateF2Polynomials(f.Degree() - 1)
		for _, a := range polys {
			if a.IsZero() {
				continue
			}
			got := a.InverseModf(f)
			if !a.Mul(got).Modf(f).Equal(F2PolynomialOne) {
				t.Errorf("invert %v mod %v: got %v", a, f, got)
			}
		}
	}
}

func TestF2PolynomialIsOne(t *testing.T) {
	one := F2PolynomialOne
	if !one.IsOne() {
		t.Errorf("IsOne() returned false")
	}
	onePlusY := F2PolynomialOnePlusY
	if onePlusY.IsOne() {
		t.Errorf("IsOne() returned true")
	}
}

func TestF2PolynomialIsZero(t *testing.T) {
	zero := F2PolynomialZero
	if !zero.IsZero() {
		t.Errorf("IsZero() returned false for F2PolynomialZero")
	}
	g := NewF2Polynomial("")
	if !g.IsZero() {
		t.Errorf("IsZero() returned false for empty F2Polynomial")
	}
	h := NewF2Polynomial("0")
	if !h.IsZero() {
		t.Errorf("IsZero() returned false for F2Polynomial(\"0\")")
	}
	i := NewF2Polynomial("00")
	if !i.IsZero() {
		t.Errorf("IsZero() returned false for F2Polynomial(\"00\")")
	}
	j := NewF2PolynomialFromSupport()
	if !j.IsZero() {
		t.Errorf("IsZero() returned false for F2PolynomialFromSupport()")
	}
	k := NewF2PolynomialFromSupport(0)
	if k.IsZero() {
		t.Errorf("IsZero() returned true for F2PolynomialFromSupport(0)")
	}
}

func TestF2PolynomialIsZeroModf(t *testing.T) {
	tests := []struct {
		p, f F2Polynomial
		want bool
	}{
		{F2PolynomialZero, F2PolynomialOne, true},
		{F2PolynomialZero, F2PolynomialY, true},
		{F2PolynomialOne, F2PolynomialOne, true},
		{F2PolynomialOne, F2PolynomialY, false},
		{F2PolynomialY, F2PolynomialY, true},
		{F2PolynomialOnePlusY, F2PolynomialOnePlusY, true},
		{F2PolynomialOnePlusY, F2PolynomialY, false},
		{NewF2Polynomial("111"), F2Polynomial111, true},
		{NewF2Polynomial("111"), F2PolynomialY, false},
		{NewF2PolynomialFromSupport(0,1,3,4), NewF2PolynomialFromSupport(0,1,4), false},
		{NewF2PolynomialFromSupport(1,11,17), NewF2PolynomialFromSupport(0,1,2), false},
	}
	for n, test := range tests {
		got := test.p.IsZeroModf(test.f)
		if got != test.want {
			t.Errorf("test %d: (%v).IsZeroModf(%v) got: %v want: %v", n, test.p, test.f, got, test.want)
		}
	}
}

func TestF2PolynomialIsOneModf(t *testing.T) {
	tests := []struct {
		p, f F2Polynomial
		want bool
	}{
		{F2PolynomialOne, F2PolynomialY, true},
		{F2PolynomialZero, F2PolynomialY, false},
		{F2PolynomialY, F2PolynomialY, false},
		{F2PolynomialOne, F2PolynomialOnePlusY, true},
		{F2PolynomialZero, F2PolynomialOnePlusY, false},
		{F2PolynomialOnePlusY, F2PolynomialOnePlusY, false},
		{NewF2Polynomial("101"), NewF2Polynomial("100"), true},
		{NewF2Polynomial("110"), NewF2Polynomial("111"), false},
		{NewF2Polynomial("010"), NewF2Polynomial("111"), false},
		{NewF2Polynomial("0111001"), NewF2Polynomial("1111001"), true},
	}
	for n, test := range tests {
		got := test.p.IsOneModf(test.f)
		if got != test.want {
			t.Errorf("test %d: (%v).IsOneModf(%v) got: %v want: %v", n, test.p, test.f, got, test.want)
		}
	}
}

func TestF2PolynomialLess(t *testing.T) {
	p := NewF2PolynomialFromSupport(1,2,3,48)
	q := NewF2PolynomialFromSupport(1,2,3,48,70,127)
	r := NewF2PolynomialFromSupport(1,2,3,48,49,70,127)
	if p.Less(p) {
		t.Errorf("Less() returned true for %v and %v", p, p)
	}
	if !p.Less(q) {
		t.Errorf("Less() returned false for %v and %v", p, q)
	}
	if !q.Less(r) {
		t.Errorf("Less() returned false for %v and %v", q, r)
	}
	if !p.Less(r) {
		t.Errorf("Less() returned false for %v and %v", p, r)
	}
	if q.Less(p) {
		t.Errorf("Less() returned true for %v and %v", q, p)
	}
	if r.Less(q) {
		t.Errorf("Less() returned true for %v and %v", r, q)
	}
	if r.Less(p) {
		t.Errorf("Less() returned true for %v and %v", r, p)
	}
}

func TestF2PolynomialMaxYFactor(t *testing.T) {
	tests := []struct {
		f F2Polynomial
		want int
	}{
		{F2PolynomialOne, 0},
		{F2PolynomialY, 1},
		{F2PolynomialOnePlusY, 0},
		{NewF2PolynomialFromSupport(0,1,2), 0},
		{NewF2PolynomialFromSupport(1,2,3), 1},
		{NewF2PolynomialFromSupport(2,3,4), 2},
		{NewF2PolynomialFromSupport(3,7), 3},
	}
	for n, test := range tests {
		if got := test.f.MaxYFactor(); got != test.want {
			t.Errorf("test %d: f: %v\n got: %d want: %d", n, test.f, got, test.want)
		}
	}
}

func TestF2PolynomialMax1PlusYFactor(t *testing.T) {
	tests := []struct {
		f F2Polynomial
		want int
	}{
		{F2PolynomialOne, 0},
		{F2PolynomialY, 0},
		{F2PolynomialOnePlusY, 1},
		{NewF2PolynomialFromSupport(0,1,2), 0},
		{NewF2PolynomialFromSupport(1,2,3), 0},
		{NewF2PolynomialFromSupport(0,1,2,4), 1},
		{NewF2PolynomialFromSupport(0,2), 2},
		{NewF2PolynomialFromSupport(0,2,3,5), 3},
		{NewF2PolynomialFromSupport(1,3,4,6), 3},
	}
	for n, test := range tests {
		if got := test.f.Max1PlusYFactor(); got != test.want {
			t.Errorf("test %d: f=%v got=%d want=%d", n, test.f, got, test.want)
		}
	}
}

func TestF2PolynomialModf(t *testing.T) {
	tests := []struct {
		f, g, want F2Polynomial
	}{
		{F2PolynomialOne, F2PolynomialOne, F2PolynomialZero},
		{F2PolynomialOnePlusY, F2PolynomialOne, F2PolynomialZero},
		{F2PolynomialOnePlusY, F2PolynomialY, F2PolynomialOne},
		{NewF2PolynomialFromSupport(0,1,3,4), NewF2PolynomialFromSupport(0,1,4), NewF2PolynomialFromSupport(3)},
		{NewF2PolynomialFromSupport(1,11,17), NewF2PolynomialFromSupport(0,1,2), F2PolynomialY},
	}
	for n, test := range tests {
		if got := test.f.Modf(test.g); !got.Equal(test.want) {
			t.Errorf("test %d: (%v) mod (%v) got: (%v) want: (%v)", n, test.f, test.g, got, test.want)
		}
	}
}

func TestF2PolynomialMul(t *testing.T) {
	tests := []struct {
		a, b, want F2Polynomial
	}{
		{F2PolynomialOne, F2PolynomialOne, F2PolynomialOne},
		{F2PolynomialOne, F2PolynomialY, F2PolynomialY},
		{F2PolynomialY, F2PolynomialOne, F2PolynomialY},
		{F2PolynomialY, F2PolynomialY, NewF2PolynomialFromSupport(2)},
		{F2PolynomialY, NewF2PolynomialFromSupport(17), NewF2PolynomialFromSupport(18)},
		{NewF2PolynomialFromSupport(17), F2PolynomialY, NewF2PolynomialFromSupport(18)},
		{F2PolynomialOnePlusY, F2PolynomialY, NewF2PolynomialFromSupport(1,2)},
		{F2PolynomialY, F2PolynomialOnePlusY, NewF2PolynomialFromSupport(1,2)},
		{F2PolynomialOnePlusY, NewF2PolynomialFromSupport(17), NewF2PolynomialFromSupport(17,18)},
		{F2PolynomialOnePlusY, F2PolynomialOnePlusY, NewF2PolynomialFromSupport(0,2)},
		{F2PolynomialOnePlusY, NewF2PolynomialFromSupport(0,1,2), NewF2PolynomialFromSupport(0,3)},
		{NewF2PolynomialFromSupport(1,17), NewF2PolynomialFromSupport(2,4), NewF2PolynomialFromSupport(3,5,19,21)},
		
	}
	for i, test := range tests {
		got := test.a.Mul(test.b)
		if !got.Equal(test.want) {
			t.Errorf("test %d: (%v)*(%v)\ngot: %v\nwant: %v", i, test.a, test.b, got, test.want)
		}
	}
}

func TestF2PolynomialNew(t *testing.T) {
	support := make([]int, 3)
	support[0], support[1], support[2] = 1, 2, 3
	p := NewF2PolynomialFromSupport(support...)
	if !p.Equal(NewF2PolynomialFromSupport(1, 2, 3)) {
		t.Errorf("p != NewF2PolynomialFromSupport(1,2,3)")
	}
	support = support[:2]
	if !p.Equal(NewF2PolynomialFromSupport(1, 2, 3)) {
		t.Errorf("p != NewF2PolynomialFromSupport(1,2,3)")
	}
}

func TestF2PolynomialPow(t *testing.T) {
	tests := []struct {
		a    F2Polynomial
		n    int
		want F2Polynomial
	}{
		{F2PolynomialOne, 0, F2PolynomialOne},
		{F2PolynomialOne, 1, F2PolynomialOne},
		{F2PolynomialOne, 2, F2PolynomialOne},
		{F2PolynomialY, 0, F2PolynomialOne},
		{F2PolynomialY, 1, F2PolynomialY},
		{F2PolynomialY, 2, NewF2PolynomialFromSupport(2)},
		{F2PolynomialY, 3, NewF2PolynomialFromSupport(3)},
		{NewF2PolynomialFromSupport(2), 0, F2PolynomialOne},
		{NewF2PolynomialFromSupport(2), 1, NewF2PolynomialFromSupport(2)},
		{NewF2PolynomialFromSupport(2), 2, NewF2PolynomialFromSupport(4)},
		{NewF2PolynomialFromSupport(2), 3, NewF2PolynomialFromSupport(6)},
		{F2PolynomialOnePlusY, 0, F2PolynomialOne},
		{F2PolynomialOnePlusY, 1, F2PolynomialOnePlusY},
		{F2PolynomialOnePlusY, 2, NewF2PolynomialFromSupport(0,2)},
		{F2PolynomialOnePlusY, 3, NewF2PolynomialFromSupport(0,1,2,3)},
	}
	for i, test := range tests {
		got := test.a.Pow(test.n)
		if !got.Equal(test.want) {
			t.Errorf("test %d: a=%v n=%d got=%v want=%v", i, test.a, test.n, got, test.want)
		}
	}
}

func TestF2PolynomialLatex(t *testing.T) {
	tests := []struct {
		f F2Polynomial
		want string
	}{
		{F2PolynomialZero, "0"},
		{F2PolynomialOne, "1"},
		{F2PolynomialY, "v"},
		{F2PolynomialOnePlusY, "1+v"},
		{NewF2Polynomial("101"), "1+v^{2}"},
	}
	for n, test := range tests {
		if got := test.f.Latex("v"); got != test.want {
			t.Errorf("test %d: got %v want %v", n, got, test.want)
		}
	}
}

func TestF2PolynomialEnumerate(t *testing.T) {
	d := 4
	want := 1 << (d + 1)
	polys := EnumerateF2Polynomials(d)
	if len(polys) != want {
		t.Errorf("len(polys)=%d want=%d", len(polys), want)
	}
	seen := make(map[F2Polynomial]bool)
	for _, p := range polys {
		if seen[p] {
			t.Errorf("duplicate polynomial")
		}
		seen[p] = true
	}
}

