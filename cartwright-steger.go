package golsv

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"strings"
)

func init() {
	initCalGMultTable()
}

// We generally adhere to the notation in [LSV], except:
//   * we use w instead of zeta.
//
// Let q = 2; We work over \F_q.
// Let F = \F_q((y)), the field of formal Laurent series in y.
// Let d = 3.  We work with the affine building associated to PGL_d(F).
// Fix \F_{q^d} = \F_q[v]/f(v) with deg(f) = d and f(v) irreducible.
// We pick f(v) = v^3 + v + 1.
// Let \varphi be the Frobenius automorphism u \mapsto u^2.
// We choose normal basis
//   \zeta_0 = 1+v
//   \zeta_1 = 1+v^2
//   \zeta_2 = 1+v+v^2
// so
//   \varphi(\zeta_i) = \zeta_{i + 1 \mod 3}.
//
// 0,1 coefficients in the normal basis
type ElementFqd [3]byte

func (v ElementFqd) Inverse() ElementFqd {
	switch v {
	case ElementFqd{1,0,0}:
		return ElementFqd{1,1,0}
	case ElementFqd{0,1,0}:
		return ElementFqd{0,1,1}
	case ElementFqd{0,0,1}:
		return ElementFqd{1,0,1}
	case ElementFqd{1,1,0}:
		return ElementFqd{1,0,0}
	case ElementFqd{1,0,1}:
		return ElementFqd{0,0,1}
	case ElementFqd{0,1,1}:
		return ElementFqd{0,1,0}
	case ElementFqd{1,1,1}:
		return ElementFqd{1,1,1}
	}
	panic("ElementFqd.Inverse: zero")
}

func (v ElementFqd) IsIdentity() bool {
	return v[0] == 1 && v[1] == 1 && v[2] == 1
}

func (v ElementFqd) IsZero() bool {
	return v[0] == 0 && v[1] == 0 && v[2] == 0
}

// returns v * w
func (v ElementFqd) Mul(w ElementFqd) ElementFqd {
	var x ElementFqd
	x[0] = (v[0] * w[1]) ^ (v[1] * w[0]) ^ (v[1] * w[2]) ^ (v[2] * w[1]) ^ (v[2] * w[2])
	x[1] = (v[0] * w[0]) ^ (v[0] * w[2]) ^ (v[2] * w[0]) ^ (v[1] * w[2]) ^ (v[2] * w[1])
	x[2] = (v[0] * w[1]) ^ (v[1] * w[0]) ^ (v[0] * w[2]) ^ (v[2] * w[0]) ^ (v[1] * w[1])
	return x
}

func FqdAllElements() []ElementFqd {
	result := make([]ElementFqd, 0, 8)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				u := ElementFqd{byte(i), byte(j), byte(k)}
				result = append(result, u)
			}
		}
	}
	return result
}

// the i,j entry is the product w_i * w_j expressed in the
// normal basis.  note that this is a subset of the full
// multiplication table of F_{q^d}.
var FqdNormalMultTable = [3][3]ElementFqd{
	{ {0,1,0}, {1,0,1}, {0,1,1} },
	{ {1,0,1}, {0,0,1}, {1,1,0} },
	{ {0,1,1}, {1,1,0}, {1,0,0} },
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

func (p F2Polynomial) InverseModf(f F2Polynomial) F2Polynomial {
	if p.IsZero() {
		panic("not a unit")
	}
	switch f {
	case "111":
		switch p {
		case F2PolynomialOne:
			return F2PolynomialOne
		case F2PolynomialY:
			return F2PolynomialOnePlusY
		case F2PolynomialOnePlusY:
			return F2PolynomialY
		default:
			panic(fmt.Sprintf("%v not reduced mod f=%v", p, f))
		}
	case "1101":
		switch p {
		case F2PolynomialOne:
			return F2PolynomialOne
		case F2PolynomialY:
			return "101"
		case F2PolynomialOnePlusY:
			return "011"
		case "001":
			return "111"
		case "101":
			return F2PolynomialY
		case "011":
			return F2PolynomialOnePlusY
		case "111":
			return "001"
		default:
			panic(fmt.Sprintf("%v not reduced mod f=%v", p, f))
		}
	default:
		panic(fmt.Sprintf("not implemented: mod f=%s", f))
	}
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

// R = \F_q[y, 1/y, 1/(1+y)]

// xxx not actually used
//
// type ElementRingR struct {
// 	// this is a fraction with
// 	// numerator a polynomial in y, and
// 	// denominator a product of y and 1+y.
// 	//
// 	//        f(y)
// 	//     ----------
// 	//     y^a (1+y)^b

// 	f F2Polynomial
// 	a, b int
// }

// Define the algebra A(R) over R:
//
//   A(R) = \bigsum_{i,j=0}^{d-1} R w_i z^j
//
// with relations
//
//     z^d = 1+y
//     z w_i = \varphi(w_i) z.
//

// xxx not actually used
// type ElementAlgebraA struct {
// 	// w_0 z^0, w_0 z^1, w_0 z^2
// 	// w_1 z^0, w_1 z^1, w_1 z^2
// 	// w_2 z^0, w_2 z^1, w_2 z^2
// 	coeffs [9]ElementRingR
// }

// Define the group G(R) by
//
//   G(R) = A(R)^\times/{R^\times}.
//
// We hold the same data as the underlying algebra, but we handle
// projectivation and select a canonical representative of the
// projective class.
// 
// And since we are projectivizing, we can always clear out
// denominators, so we don't need to store them here.  So in fact, we
// just store the numerator, which is a polynomial in y.
type ElementCalG [9]F2Polynomial // xxx rename ElementGammaZero EltGammaZero?


//           z^0  z^1  z^2
//         ----------------
// w_0 | r00  r01  r02
// w_1 | r10  r11  r12
// w_2 | r20  r21  r22
func NewElementCalG(r00, r01, r02, r10, r11, r12, r20, r21, r22 F2Polynomial) ElementCalG {
	// note that we don't check whether g is invertible
	g := ElementCalG{
		r00, r01, r02,
		r10, r11, r12,
		r20, r21, r22,
	}
	return g.normalize()
}

func newElementCalGArrayNotNormalized(array [9]F2Polynomial) ElementCalG {
	return array
}

func NewElementCalGFromFieldElement(u ElementFqd) ElementCalG {
	a, b, c := F2PolynomialZero, F2PolynomialZero, F2PolynomialZero
	if u[0] == 1 {
		a = F2PolynomialOne
	}
	if u[1] == 1 {
		b = F2PolynomialOne
	}
	if u[2] == 1 {
		c = F2PolynomialOne
	}
	return NewElementCalG(
		a, F2PolynomialZero, F2PolynomialZero,
		b, F2PolynomialZero, F2PolynomialZero,
		c, F2PolynomialZero, F2PolynomialZero)
}

// Notice that the multiplicative identity 1_A, using the normal basis
// chosen above, is
//
//   1_A = w_0 z^0 + w_1 z^0 + w_2 z^0.
//
func NewElementCalGIdentity() ElementCalG {
	return newElementCalGArrayNotNormalized(
		[9]F2Polynomial{
			F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
			F2PolynomialOne, F2PolynomialZero, F2PolynomialZero})
}

func NewElementCalGFromString(s string) ElementCalG {
	// e.g. (11,0,1)(101,1,1)(101,1,0)
	re := regexp.MustCompile(`\(([01]+),([01]+),([01]+)\)`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) != 3 {
		panic("unrecognized input")
	}
	array := [9]F2Polynomial{}
	for i, match := range matches {
		array[i*3] = NewF2Polynomial(match[1])
		array[i*3+1] = NewF2Polynomial(match[2])
		array[i*3+2] = NewF2Polynomial(match[3])
	}
	g := newElementCalGArrayNotNormalized(array)
	return g.normalize()
}

// sets g = a + b.  this is an internal method since the result is not
// guaranteed to be an element of the group. result is not normalized.
func (g *ElementCalG) add(a, b ElementCalG) {
	for i := 0; i < 9; i++ {
		g[i] = a[i].Add(b[i])
	}
}

// sets g = h
func (g *ElementCalG) Copy(h ElementCalG) {
	for i := 0; i < 9; i++ {
		g[i] = h[i].Dup()
	}
}

func (g ElementCalG) Dump() string {
	s := ""
	for i := 0; i < 9; i++ {
		s += fmt.Sprintf("g[%d] = %s\n", i, g[i].Dump())
	}
	return s
}

func (g ElementCalG) Dup() ElementCalG {
	h := NewElementCalGIdentity()
	h.Copy(g)
	return h
}

// nb. we have carefully defined the data structure and normalization
// so that golang equality semantics correspond to group element
// equality. this prevents madness when using group elements as map
// keys.
func (g ElementCalG) Equal(h ElementCalG) bool {
	return g == h
}

func (g ElementCalG) firstNonzeroIndex() int {
	for i := 0; i < 9; i++ {
		if !g[i].IsZero() {
			return i
		}
	}
	return -1
}

func (g ElementCalG) IsIdentity() bool {
	id := ElementCalG{
		F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
		F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
		F2PolynomialOne, F2PolynomialZero, F2PolynomialZero,
	}
	return g.Equal(id)
}

func (g ElementCalG) IsIdentityModf(f F2Polynomial) bool {
	h := g.Modf(f)
	return h.IsIdentity()
}

func (g ElementCalG) isZero() bool {
	for i := 0; i < 9; i++ {
		if !g[i].IsZero() {
			return false
		}
	}
	return true
}

// xxx test
func (g ElementCalG) Less(h ElementCalG) bool {
	for i := 0; i < len(g); i++ {
		if g[i].Less(h[i]) {
			return true
		}
		if h[i].Less(g[i]) {
			return false
		}
	}
	return false
}

// returns the element with each entry taken mod f, normalized. (see
// the normalization comment below)
func (g ElementCalG) Modf(f F2Polynomial) ElementCalG {
	h := g.Dup()
	for i := 0; i < 9; i++ {
		h[i] = g[i].Modf(f)
	}
	return h.normalizeModf(f)
}

// sets g = a * b 
// xxx i hate this; just make it c = a.Mul(b)
func (g *ElementCalG) Mul(a, b ElementCalG) {
	g.mulNotNormalized(a, b)
	h := g.normalize()
	g.Copy(h)
}

func (g *ElementCalG) mulNotNormalized(a, b ElementCalG) {
	g.zero()
	u := NewElementCalGIdentity()
	tmp := NewElementCalGIdentity()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			r := a[i*3+j]
			if r.IsZero() {
				continue
			}
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					s := b[k*3+l]
					if s.IsZero() {
						continue
					}
					t := r.Mul(s)
					// lookup the product h := w_i z^j * w_k z^l.
					// the result may involve a linear combination of w's.
					h := calGMultTable[i*3+j][k*3+l]
					// log.Printf("xxx %p [%d%d%d%d] h: %v", g, i, j, k, l, h)
					u.Copy(h)
					u.scale(t)
					// log.Printf("xxx u:\n%v", u)
					// log.Printf("xxx %p [%d%d%d%d] u: %v", g, i, j, k, l, u)
					tmp.add(*g, u)
					g.Copy(tmp)
					// log.Printf("xxx %p [%d%d%d%d] g: %v", g, i, j, k, l, g)
				}
			}
		}
	}
}

// normalization: 
//
// we always have the following projective quotient:
//
//    A(R)^\times
//      |
//      |
//      v
//    A(R)^\times / R^\times
//
// so we are dealing with projective cosets of R^\times rather than
// group elements in A(R)^\times directly.  this is significant
// because we want to represent the projective cosets in a data type
// efficiently, and we want to be able to compare them for equality
// using the native == operator (native golang comparator semantics).
// in order to do this, we need to choose a canonical representative
// of each coset, a process which we call normalization.  
//
// additionally, we optionally have another quotient by modulus I = (f),
// (we refer to this a "modf" for short).
//
//
//    A(R/I)^\times
//      |
//      |
//      v
//    A(R/I)^\times / (R/I)^\times
//
//
// if we are doing a quotient Modf, we find the first unit entry
// and scale by its inverse.
// 
// xxx special case coded for now: f = 1+y+y^2.  then R/I = { 0, 1, y,
// 1+y }, and y*(1+y) = 1+y^2 = 1.  so all elements of R/I are
// invertible.  hence to normalize, just find the first nonzero entry
// and scale by its inverse.

// returns new element
func (g ElementCalG) normalizeModf(f F2Polynomial) ElementCalG {
	h := NewElementCalGIdentity()
	a := g.firstNonzeroIndex()
	if a < 0 {
		panic("not implemented")
	}
	p := g[a]
	pInv := p.InverseModf(f)
	for i := 0; i < 9; i++ {
		h[i] = g[i].Mul(pInv).Modf(f)
	}
	return h
}

func (g ElementCalG) normalize() ElementCalG {
	h := g.Dup()
	a := g.firstNonzeroIndex()
	if a < 0 {
		panic("zero cannot be normalized")
	}
	n := math.MaxInt
	for i := a; i < 9; i++ {
		if h[i].IsZero() {
			continue
		}
		k := h[i].MaxYFactor()
		// log.Printf("xxx g=%v i=%d k=%d", g, i, k)
		if k < n {
			n = k
		}
	}
	// log.Printf("xxx n=%d", n)
	for i := a; i < 9; i++ {
		if h[i].IsZero() {
			continue
		}
		for j := 0; j < n; j++ {
			quotient, remainder := h[i].Div(F2PolynomialY)
			if !remainder.IsZero() {
				panic("internal error: remainder is non-zero")
			}
			h[i] = quotient
		}
	}
	m := math.MaxInt
	for i := a; i < 9; i++ {
		if h[i].IsZero() {
			continue
		}
		k := h[i].Max1PlusYFactor()
		if k < m {
			m = k
		}
	}
	for i := a; i < 9; i++ {
		if h[i].IsZero() {
			continue
		}
		for j := 0; j < m; j++ {
			quotient, remainder := h[i].Div(F2PolynomialOnePlusY)
			if !remainder.IsZero() {
				panic("internal error: remainder is non-zero")
			}
			h[i] = quotient
		}
	}
	return h
}

func (g ElementCalG) Order() int {
	h := NewElementCalGIdentity()
	t := NewElementCalGIdentity()
	k := 0
	for {
		h.Mul(t, g)
		k++
		if h.IsIdentity() {
			return k
		}
		t.Copy(h)
	}
}

// sets g = p * g; internal method, result is not normalized.
func (g *ElementCalG) scale(p F2Polynomial) {
	for i := 0; i < 9; i++ {
		g[i] = g[i].Mul(p)
	}
}

// xxx new formula:
func (g ElementCalG) String() string {
	var buf bytes.Buffer
	for i := 0; i < 3; i++ {
		buf.WriteString("(")
		for j := 0; j < 3; j++ {
			if j > 0 {
				buf.WriteString(",")
			}
			f := g[i*3+j]
			if f.IsZero() {
				buf.WriteString("0")
			} else {
				buf.WriteString(f.String())
			}
		}
		buf.WriteString(")")
	}
	if buf.Len() == 0 {
		buf.WriteString("0")
	}
	return buf.String()
}

// xxx original:
func (g ElementCalG) OriginalString() string {
	var buf bytes.Buffer
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			f := g[i*3+j]
			if f.IsZero() {
				continue
			}
			if buf.Len() > 0 {
				buf.WriteString(" + ")
			}
			if f.IsOne() {
				fmt.Fprintf(&buf, "w_%d z^%d", i, j)
			} else {
				fmt.Fprintf(&buf, "(%v) w_%d z^%d", f, i, j)
			}
		}
	}
	if buf.Len() == 0 {
		buf.WriteString("0")
	}
	return buf.String()
}

// sets g = 0; this is an internal method since technically zero is
// not a valid element of the group.
func (g *ElementCalG) zero() {
	for i := 0; i < 9; i++ {
		g[i] = F2PolynomialZero
	}
}

var calGMultTable [9][9]ElementCalG

func initCalGMultTable() {
	log.Printf("initializing calG multiplication table")
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					calGMultTable[i*3+j][k*3+l] =
						computeCalGBasisProduct(i, j, k, l)
				}
			}
		}
	}
}

// type wCoord [2]int // i, j as in w_i z^j

// xxx experimental; test?
// func calGMultTablePreImage(z wCoord) []wCoord {
// 	result := make([]wCoord, 0)
// 	for i := 0; i < 3; i++ {
// 		for j := 0; j < 3; j++ {
// 			for k := 0; k < 3; k++ {
// 				for l := 0; l < 3; l++ {
// 					g := calGMultTable[i*3+j][k*3+l]
// 					if !g[z[0] + 3*z[1]].IsZero() {
// 						result = append(result, wCoord{i, j})
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return result
// }

// computes x = w_i z^j * w_k z^l
func computeCalGBasisProduct(i, j, k, l int) ElementCalG {
	// first, commute the w_k factors to the left across the z^l
	// factors using the rule:
	//
	//   z w_i = phi(w_i) z
	//            = w_{i+1 mod 3} z
	// thus
	//
	//   x = w_i z^j * w_k z^l
	//     = w_i w_{k+j mod 3} z^{j+l}
	//     = w_i w_m z^{j+l}
	m := (k + j) % 3
	// find w_i * w_m by lookup
	wVec := FqdNormalMultTable[i][m]
	// then
	//
	//   w_i * w_m = wVec[0] w_0 + wVec[1] w_1 + wVec[2] w_2
	//
	// where each wVec[s] is 0 or 1
	//
	// to compute z^{j+l}, we use the rule:
	//
	//   z^3 = 1+y
	//
	// let n = (j+l) mod 3
	//     p = int((j+l) / 3)
	n := (j + l) % 3
	p := (j + l) / 3
	//
	// then
	//
	//   z^{j+l} = z^{3p} z^n
	//           = (1+y)^p z^n
	// so
	//
	//   x = w_i  w_m z^{j+l}
	//     = (1+y)^p w_i w_m z^n
	//     = (1+y)^p (wVec[0] w_0 + wVec[1] w_1 + wVec[2] w_2) z^n
	//     = (1+y)^p (wVec[0] w_0 z^n + wVec[1] w_1 z^n + wVec[2] w_2 z^n)
	//
	// let
	//
	//   u = (1+y)^p
	//
	u := F2PolynomialOnePlusY.Pow(p)
	
	var polys [9]F2Polynomial
	if wVec[0] == 1 {
		polys[n] = u
	}
	if wVec[1] == 1 {
		polys[n+3] = u
	}
	if wVec[2] == 1 {
		polys[n+6] = u
	}
	return newElementCalGArrayNotNormalized(polys)
}

// Notice also that the element z^{-1} is
//
//   z^{-1} = 1_A * z^{d-1}/(1+y)
//
// To define the Cartwright-Steger generators, first we define the
// element b \in G(R) by
//
//   b = 1_A - z^{-1}
//
//     = (w_0 + w_1 + w_2) z^0 + (w_0 + w_1 + w_2) z^2/(1+y)
// 
// Making use of projectivization to clear denominators, technically
// multiplying by (1+y)(w_0 + w_1 + w_2) z^0, we have
//
//   b =   (1+y)w_0 z^0 + w_0 z^2
//       + (1+y)w_1 z^0 + w_1 z^2
//	     + (1+y)w_2 z^0 + w_2 z^2
//
func cartwrightStegerGenB() ElementCalG {
	return NewElementCalG(
		F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialOne,
		F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialOne,
		F2PolynomialOnePlusY, F2PolynomialZero, F2PolynomialOne)
}

func cartwrightStegerGenBInv() ElementCalG {
	return NewElementCalG(
		F2PolynomialOnePlusY, F2PolynomialOne, F2PolynomialOne,
		F2PolynomialOnePlusY, F2PolynomialOne, F2PolynomialOne,
		F2PolynomialOnePlusY, F2PolynomialOne, F2PolynomialOne)
}

func CartwrightStegerGenerators() []ElementCalG {
	gens := make([]ElementCalG, 0)
	b := cartwrightStegerGenB()
	bInv := cartwrightStegerGenBInv()
	fieldElements := FqdAllElements()
	for _, u := range fieldElements {
		if u.IsZero() {
			continue
		}
		// b_u = u b u^{-1}
		g_u := NewElementCalGFromFieldElement(u)
		tmp := NewElementCalGIdentity()
		tmp.Mul(g_u, b)
		uInv := u.Inverse()
		g_uInv := NewElementCalGFromFieldElement(uInv)
		b_u := NewElementCalGIdentity()
		b_u.Mul(tmp, g_uInv)
		gens = append(gens, b_u)
		// b_u^{-1} = u b^{-1} u^{-1}
		tmp.Mul(g_u, bInv)
		b_uInv := NewElementCalGIdentity()
		b_uInv.Mul(tmp, g_uInv)
		gens = append(gens, b_uInv)
	}
	return gens
}

func CartwrightStegerGeneratorsInverse(gens []ElementCalG, n int) ElementCalG {
	if n % 2 == 0 {
		if n + 1 >= len(gens) {
			panic("n too large")
		}
		return gens[n+1]
	}
	if n - 1 < 0 {
		panic("n too small")
	}
	return gens[n-1]
}
