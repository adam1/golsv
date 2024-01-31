package golsv

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
)

func init() {
	initInverseModfTable()
}

// \F_2[y]
//
// We would like the ElementCalG type to be comparable (in the native
// Go sense) so that we can use it as a map key.  since ElementCalG is
// an array of nine F2Polynomials, we need F2Polynomial to be
// comparable.  in principle, the degree of such a polynomial not
// bounded; a polynomial can contain an arbitrarily large amount of
// information. the only natively comparable type that supports an
// unbounded length is string. so we store polynomials as strings, in
// particular a string of 0s and 1s.  this is not the most efficient
// representation, but it is simple and it works.

// F2Polynomial is a polynomial in y with coefficients in \F_2.
// it is represented as a string of 0s and 1s, where the i-th
// character is the coefficient of y^i.
type F2Polynomial string

// nb. the zero polynomial is represented by the empty string. it
// could also be represented by "0", but we choose to use strings
// normalized in the sense of removing trailing zeros.  this is to
// make equality equivalent to native golang string equality.
const F2PolynomialZero = F2Polynomial("")

const F2PolynomialOne = F2Polynomial("1")
const F2PolynomialY = F2Polynomial("01")
const F2PolynomialOnePlusY = F2Polynomial("11")
const F2Polynomial111 = F2Polynomial("111")

func NewF2Polynomial(s string) F2Polynomial {
	return F2Polynomial(s).normalize()
}

// xxx DEPRECATED and broken
// coeffs is a slice of 0s and 1s
// func NewF2PolynomialFromCoefficients(coeffs []int) F2Polynomial {
// 	runes := make([]rune, len(coeffs))
// 	for i, c := range coeffs {
// 		switch c {
// 		case 0:
// 			runes[i] = '0'
// 		case 1:
// 			runes[i] = '1'
// 		default:
// 			panic("NewF2PolynomialFromCoefficients: bad coefficient")
// 		}
// 		runes[i] = rune(c)
// 	}
// 	return F2Polynomial(runes)
// }

func NewF2PolynomialFromSupport(support ...int) F2Polynomial {
	if len(support) == 0 {
		return F2PolynomialZero
	}
	supp := make([]int, len(support))
	copy(supp, support)
	sort.Ints(supp)
	runes := make([]rune, supp[len(supp)-1]+1)
	for i := 0; i < len(runes); i++ {
		runes[i] = '0'
	}
	for _, k := range support {
		runes[k] = '1'
	}
	return NewF2Polynomial(string(runes))
}

func NewF2PolynomialRandom(maxDegree int) F2Polynomial {
	runes := make([]rune, maxDegree+1)
	for i := 0; i <= maxDegree; i++ {
		if rand.Intn(2) == 0 {
			runes[i] = '0'
		} else {
			runes[i] = '1'
		}
	}
	return NewF2Polynomial(string(runes))
}

// returns a + b
func (p F2Polynomial) Add(q F2Polynomial) F2Polynomial {
	n := len(p)
	if len(q) > n {
		n = len(q)
	}
	runes := make([]rune, n)
	pb := []rune(p)
	qb := []rune(q)
	for i := 0; i < n; i++ {
		if i < len(pb) && i < len(qb) {
			if pb[i] == qb[i] {
				runes[i] = '0'
			} else {
				runes[i] = '1'
			}
		} else if i < len(pb) {
			runes[i] = pb[i]
		} else {
			runes[i] = qb[i]
		}
	}
	return NewF2Polynomial(string(runes))
}

func (p F2Polynomial) AddMonomial(degree int) F2Polynomial {
	var runes []rune
	if degree + 1 > len(p) {
		runes = make([]rune, degree+1)
		copy(runes, []rune(p))
		for i := len(p); i < degree; i++ {
			runes[i] = '0'
		}
	} else {
		runes = []rune(p)
	}
	if runes[degree] == '1' {
		runes[degree] = '0'
	} else {
		runes[degree] = '1'
	}
	return NewF2Polynomial(string(runes))
}

func (p F2Polynomial) Coefficient(degree int) int {
	if degree < 0 || degree >= len(p) {
		return 0
	}
	if p[degree] == '0' {
		return 0
	}
	return 1
}

// xxx do we even use this?  DEPRECATED
// func (p F2Polynomial) Coefficients() []int {
// 	coeffs := make([]int, len(p))
// 	for i, c := range []rune(p) {
// 		switch c {
// 		case '0':
// 			coeffs[i] = 0
// 		case '1':
// 			coeffs[i] = 1
// 		default:
// 			panic("F2Polynomial.Coefficients: bad coefficient")
// 		}
// 	}
// 	return coeffs
// }

// xxx deprecated; strings are immutable
// func (p F2Polynomial) Copy(q *F2Polynomial) {
// 	if p == nil {
// 		if q.IsZero() {
// 			return
// 		}
// 		panic("F2Polynomial.Copy: p is nil")
// 	}
// 	if q.IsZero() {
// 		p.support = p.support[:0]
// 		return
// 	}
// 	p.support = append(p.support[:0], q.support...)
// }

func (p F2Polynomial) Degree() int {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '1' {
			return i
		}
	}
	return -1
}

func (p F2Polynomial) Div(f F2Polynomial) (quotient, remainder F2Polynomial) {
	if f.IsZero() {
		panic("F2Polynomial.Div: divide by zero")
	}
	q := F2PolynomialZero
	r := p
	dr := r.Degree()
	df := f.Degree()
	for !r.IsZero() && dr >= df {
		t := dr - df
		q = q.AddMonomial(t)
		u := NewF2PolynomialFromSupport(t)
		c := f.Mul(u)
		r = r.Add(c)
		dr = r.Degree()
	}
	return q, r
}

// xxx deprecate
// // sets p = p * y^{-n}.  panics if power would become negative.
// func (p *F2Polynomial) DivY(n int) {
// 	if n == 0 {
// 		return
// 	}
// 	for i, _ := range p.support {
// 		p.support[i] -= n
// 		if p.support[i] < 0 {
// 			panic("F2Polynomial.DivY: negative power")
// 		}
// 	}
// }

// // sets p = p / (1+y).  returns quotient and remainder.
// func (p *F2Polynomial) Div1PlusY() (quotient, remainder *F2Polynomial) {
// 	// if p has an odd number of terms, it is not divisible by
// 	// 1+y. use this fact to avoid computing an infinite series.
// 	if len(p.support) % 2 == 1 {
// 		return nil, p
// 	}
// 	coeffs := p.Coefficients()
// 	quotientCoeffs := make([]int, len(coeffs))
// 	remainderCoeffs := make([]int, len(coeffs))
// 	copy(remainderCoeffs, coeffs)

// 	for i := 0; i < len(coeffs); i++ {
// 		c := remainderCoeffs[i]
// 		if c == 1 {
// 			quotientCoeffs[i] = 1
// 			remainderCoeffs[i] = 0
// 			remainderCoeffs[i+1] ^= 1
// 		}
// 	}
// 	quotient = NewF2PolynomialFromCoefficients(quotientCoeffs)
// 	remainder = NewF2PolynomialFromCoefficients(remainderCoeffs)
// 	return
// }

func (p F2Polynomial) Dump() string {
	return "(" + string(p) + ")"
}

func (p F2Polynomial) Dup() F2Polynomial {
	return NewF2Polynomial(string(append([]byte(nil), p...)))
}

// it's important that this equality match ordinary string equality,
// so that F2Polynomial is a "comparable" type, e.g. can be used as a
// map key.
func (p F2Polynomial) Equal(q F2Polynomial) bool {
	return p == q
}

var inverseModfTable map[F2Polynomial]map[F2Polynomial]F2Polynomial

func initInverseModfTable() {
	log.Printf("initializing InverseModf lookup table")
	fs := []struct {
		f, g F2Polynomial
	}{
		{"111", F2PolynomialY},
		{"1101", F2PolynomialY},
	}
	inverseModfTable = make(map[F2Polynomial]map[F2Polynomial]F2Polynomial)
	for _, poly := range fs {
		f, g := poly.f, poly.g
		invg := F2PolynomialZero
		for i := 1; i < (1 << f.Degree()); i++ {
			a := F2PolynomialZero
			for j := 0; j < f.Degree(); j++ {
				if (i >> j) & 1 != 0 {
					a = a.AddMonomial(j)
				}
			}
			if a.Mul(g).Modf(f).Equal(F2PolynomialOne) {
				invg = a
			}
		}
		if invg == F2PolynomialZero {
			panic(fmt.Sprintf("no inverse found for generator %v mod %v", g, f))
		}
		if !g.Mul(invg).Modf(f).Equal(F2PolynomialOne) {
			panic(fmt.Sprintf("%v and %v are not inverses mod %v", g, invg, f))
		}
		inverseModfTable[f] = make(map[F2Polynomial]F2Polynomial)
		e, inve := F2PolynomialOne, F2PolynomialOne
		for i := 0; i < (1 << f.Degree()) - 1; i++ {
			inverseModfTable[f][e] = inve
			e = e.Mul(g).Modf(f)
			inve = inve.Mul(invg).Modf(f)
		}
	}
}

func (p F2Polynomial) InverseModf(f F2Polynomial) F2Polynomial {
	if p.IsZero() {
		panic("not a unit")
	}
	lut, ok := inverseModfTable[f]
	if !ok {
		panic(fmt.Sprintf("not implemented: mod f=%v", f))
	}
	invp, ok := lut[p]
	if !ok {
		panic(fmt.Sprintf("%v not reduced mod f=%v", p, f))
	}
	return invp
}

func (p F2Polynomial) IsOne() bool {
	if p.Degree() == 0 {
		if len(p) >= 1 {
			runes := []rune(p)
			return runes[0] == '1'
		}
	}
	return false
}

func (p F2Polynomial) IsZero() bool {
	if len(p) == 0 {
		return true
	}
	for _, b := range p {
		if b == '1' {
			return false
		}
	}
	return true
}

// xxx test
func (p F2Polynomial) Less(q F2Polynomial) bool {
	return p < q
}

// returns the highest power of y that divides p
func (p F2Polynomial) MaxYFactor() int {
	minPower := -1
	for i, c := range p {
		if c == '1' {
			minPower = i
			break
		}
	}
	if minPower < 1 {
		return 0
	}
	return minPower
}

// returns the highest power of 1+y that divides p
func (p F2Polynomial) Max1PlusYFactor() int {
	s := p.Dup()
	k := 0
	for {
		quotient, remainder := s.Div(F2PolynomialOnePlusY)
		// if the remainder is zero, keep going
		if remainder.IsZero() {
			s = quotient
			k++
			continue
		}
		// if the remainder is not zero, we're done
		break
	}
	return k
}

func (p F2Polynomial) Modf(f F2Polynomial) F2Polynomial {
	_, remainder := p.Div(f)
	return remainder
}

// return p * q
func (p F2Polynomial) Mul(q F2Polynomial) F2Polynomial {
	n := p.Degree()
	m := q.Degree()
	runes := make([]rune, n+m+1)
	for i := 0; i < n+m+1; i++ {
		runes[i] = '0'
	}
	for i, a := range p {
		for j, b := range q {
			if a == '1' && b == '1' {
				if runes[i+j] == '1' {
					runes[i+j] = '0'
				} else {
					runes[i+j] = '1'
				}
			}
		}
	}
	return NewF2Polynomial(string(runes))
}

func (p F2Polynomial) normalize() F2Polynomial {
	return F2Polynomial(strings.TrimRight(string(p), "0"))
}

// returns p^n
func (p F2Polynomial) Pow(n int) F2Polynomial {
	q := F2PolynomialOne
	for i := 0; i < n; i++ {
		q = q.Mul(p)
	}
	return q
}

func (p F2Polynomial) String() string {
	return string(p)
}
