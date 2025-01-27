package golsv

import (
	"strings"
)

type MatF2Poly [9]F2Polynomial

var MatF2PolyIdentity = MatF2Poly{
	F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
	F2PolynomialZero, F2PolynomialOne, F2PolynomialZero,
	F2PolynomialZero, F2PolynomialZero, F2PolynomialOne,
}

func NewMatF2Poly(a, b, c, d, e, f, g, h, i F2Polynomial) MatF2Poly {
	return MatF2Poly{a, b, c, d, e, f, g, h, i}
}

func NewMatF2PolyFromString(s string) MatF2Poly {
	m := MatF2Poly{}
	entries := strings.Split(strings.Trim(s, " []"), " ")
	if len(entries) != len(m) {
		panic("wrong number of entries")
	}
	for i := 0; i < len(m); i++ {
		strings.TrimSpace(entries[i])
		m[i] = NewF2Polynomial(entries[i])
	}
	return m
}

func (M MatF2Poly) Add(N MatF2Poly) MatF2Poly {
	var R MatF2Poly
	for i := 0; i < 9; i++ {
		R[i] = M[i].Add(N[i])
	}
	return R
}

func (M MatF2Poly) Equal(N MatF2Poly) bool {
	for i := 0; i < 9; i++ {
		if !M[i].Equal(N[i]) {
			return false
		}
	}
	return true
}

func (M MatF2Poly) Latex() string {
	s := "\\begin{bmatrix}"
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			s += M[i*3+j].Latex()
			if j < 2 {
				s += " & "
			}
		}
		if i < 2 {
			s += " \\\\ "
		}
	}
	s += "\\end{bmatrix}"
	return s
}

func (M MatF2Poly) Mul(N MatF2Poly) MatF2Poly {
	var R MatF2Poly
	R[0] = M[0].Mul(N[0]).Add(M[1].Mul(N[3])).Add(M[2].Mul(N[6]))
	R[1] = M[0].Mul(N[1]).Add(M[1].Mul(N[4])).Add(M[2].Mul(N[7]))
	R[2] = M[0].Mul(N[2]).Add(M[1].Mul(N[5])).Add(M[2].Mul(N[8]))
	R[3] = M[3].Mul(N[0]).Add(M[4].Mul(N[3])).Add(M[5].Mul(N[6]))
	R[4] = M[3].Mul(N[1]).Add(M[4].Mul(N[4])).Add(M[5].Mul(N[7]))
	R[5] = M[3].Mul(N[2]).Add(M[4].Mul(N[5])).Add(M[5].Mul(N[8]))
	R[6] = M[6].Mul(N[0]).Add(M[7].Mul(N[3])).Add(M[8].Mul(N[6]))
	R[7] = M[6].Mul(N[1]).Add(M[7].Mul(N[4])).Add(M[8].Mul(N[7]))
	R[8] = M[6].Mul(N[2]).Add(M[7].Mul(N[5])).Add(M[8].Mul(N[8]))
	return R
}

func (M MatF2Poly) Pow(n int) MatF2Poly {
	R := MatF2PolyIdentity
	for i := 0; i < n; i++ {
		R = R.Mul(M)
	}
	return R
}

func (M MatF2Poly) Scale(c F2Polynomial) MatF2Poly {
	var R MatF2Poly
	for i := 0; i < 9; i++ {
		R[i] = M[i].Mul(c)
	}
	return R
}

func (M MatF2Poly) String() string {
	return "[" + M[0].String() + " " + M[1].String() + " " + M[2].String() + " " +
		M[3].String() + " " + M[4].String() + " " + M[5].String() + " " +
		M[6].String() + " " + M[7].String() + " " + M[8].String() + "]"
}

func (M MatF2Poly) Trace() F2Polynomial {
	return M[0].Add(M[4]).Add(M[8])
}

type ProjMatF2Poly MatF2Poly

func NewProjMatF2PolyFromString(s string) ProjMatF2Poly {
	return ProjMatF2Poly(NewMatF2PolyFromString(s))
}

var ProjMatF2PolyIdentity = ProjMatF2Poly(MatF2PolyIdentity)

func (M ProjMatF2Poly) Equal(N ProjMatF2Poly) bool {
	k := -1
	for i := 0; i < 9; i++ {
		if !M[i].IsZero() || !N[i].IsZero() {
			k = i
			break
		}
	}
	if k == -1 {
		return true // both matrices are zero
	}
	a, b := M[k], N[k]
	if (a.IsZero() && !b.IsZero()) || (!a.IsZero() && b.IsZero()) {
		return false
	}
	for i := 0; i < 9; i++ {
		if i == k {
			continue
		}
		c, d := M[i], N[i]
		if !a.Mul(d).Equal(b.Mul(c)) {
			return false
		}
	}
	return true
}

func (M ProjMatF2Poly) Latex() string {
	return MatF2Poly(M).Latex()
}

func (M ProjMatF2Poly) Mul(N ProjMatF2Poly) ProjMatF2Poly {
	return ProjMatF2Poly(MatF2Poly(M).Mul(MatF2Poly(N)))
}

func (M ProjMatF2Poly) ReduceModf(f F2Polynomial) ProjMatF2Poly {
	N := ProjMatF2Poly{}
	for i := 0; i < 9; i++ {
		N[i] = M[i].Modf(f)
	}
	return N
}

func (M ProjMatF2Poly) String() string {
	return MatF2Poly(M).String()
}

