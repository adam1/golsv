package golsv

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
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
//   w_0 = 1+v
//   w_1 = 1+v^2
//   w_2 = 1+v+v^2
// so
//   \varphi(w_i) = w_{i + 1 \mod 3}.
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

func (w ElementFqd) ToStandardBasis() F2Polynomial {
	sum := F2PolynomialZero
	// w_0 \mapsto 1 + v
	// w_1 \mapsto 1 + v^2
	// w_2 \mapsto 1 + v + v^2
	if w[0] == 1 {
		sum = sum.Add(smallF2Polynomial(3))
	}
	if w[1] == 1 {
		sum = sum.Add(smallF2Polynomial(5))
	}
	if w[2] == 1 {
		sum = sum.Add(smallF2Polynomial(7))
	}
	return sum	
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
		s += fmt.Sprintf("g[%d] = %s\n", i, g[i])
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

func (g ElementCalG) Latex() string {
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
			if !f.IsOne() {
				fmt.Fprintf(&buf, "(%v) ", f.Latex("y"))
			}
			fmt.Fprintf(&buf, "\\zeta_%d", i)
			if j == 1 {
				fmt.Fprintf(&buf, " z")
			} else if j == 2 {
				fmt.Fprintf(&buf, " z^{%d}", j)
			}			
		}
	}
	if buf.Len() == 0 {
		buf.WriteString("0")
	}
	return buf.String()
}

func (g ElementCalG) LatexMatrix() string {
	var buf bytes.Buffer
	buf.WriteString("\\left \\langle \\begin{matrix}\n")
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			f := g[i*3+j]
			if j != 0 {
				buf.WriteString(" & ")
			}
			fmt.Fprint(&buf, f.Latex("y"))
		}
		buf.WriteString("\\\\\n")
	}
	buf.WriteString("\\end{matrix}\\right \\rangle \n")
	return buf.String()
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

func (g ElementCalG) OrderModf(f F2Polynomial) int {
	t := NewElementCalGIdentity()
	k := 0
	for {
		t.Mul(t, g)
		t = t.Modf(f)
		k++
		if t.IsIdentity() {
			return k
		}
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
	// log.Printf("initializing calG multiplication table")
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

// xxx deprecate
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

var cartwrightStegerModulus F2Polynomial = NewF2Polynomial("1101") // 1 + v + v^3

var cartwrightStegerEmbeddingBeta F2Polynomial = NewF2Polynomial("11") // beta = 1 + v

var cartwrightStegerEmbeddingY F2Polynomial = NewF2Polynomial("0101") // y(x) = x + x^3

// Construct the matrix representations of the Cartwright-Steger
// generators.  Here, we follow the notation of LSV section 10.
type CartwrightStegerGenMatrixInfo struct {
	U           F2Polynomial
	B_u, B_uInv ProjMatF2Poly
}

// xxx deprecate
func CartwrightStegerGeneratorsMatrixReps() (gens []ProjMatF2Poly, table []CartwrightStegerGenMatrixInfo) {
	// per the example in LSV section 10, we don't need to use the
	// normal basis of F_8; we can use the standard basis instead.
	repB := cartwrightStegerMatrixRepB()
	repBInv := cartwrightStegerMatrixRepBInverse()
 	gens = make([]ProjMatF2Poly, 0)
 	table = make([]CartwrightStegerGenMatrixInfo, 0)
	fieldElements := FqdAllElements()
	for _, uZeta := range fieldElements {
		if uZeta.IsZero() {
			continue
		}
		u := uZeta.ToStandardBasis()
		uRep := ProjMatF2Poly(cartwrightStegerMatrixRepFieldElement(u))
		uInvRep := ProjMatF2Poly(cartwrightStegerMatrixRepFieldElement(u.InverseModf(cartwrightStegerModulus)))
		// b_u = u b u^{-1}
		b_u := uRep.Mul(repB).Mul(uInvRep)
		gens = append(gens, b_u)
		// b_u^{-1} = u b^{-1} u^{-1}
		b_uInv := uRep.Mul(repBInv).Mul(uInvRep)
		gens = append(gens, b_uInv)
		table = append(table, CartwrightStegerGenMatrixInfo{u, b_u, b_uInv})
	}
	return gens, table
}

type CartwrightStegerGenInfo struct {
	U                 F2Polynomial
	B_u, B_uInv       ElementCalG
	B_uRep, B_uInvRep ProjMatF2Poly
}

// set modulus to zero to not take a quotient
func CartwrightStegerGeneratorsWithMatrixReps(modulus F2Polynomial) []CartwrightStegerGenInfo {
	table := make([]CartwrightStegerGenInfo, 0)
	b := cartwrightStegerGenB()
	bInv := cartwrightStegerGenBInv()
	repB := cartwrightStegerMatrixRepB()
	repBInv := cartwrightStegerMatrixRepBInverse()
	fieldElements := FqdAllElements()
	for _, uZeta := range fieldElements {
		if uZeta.IsZero() {
			continue
		}
		// first, the ElementCalG form
		// b_u = u b u^{-1}
		g_u := NewElementCalGFromFieldElement(uZeta)
		tmp := NewElementCalGIdentity()
		tmp.Mul(g_u, b)
		uInv := uZeta.Inverse()
		g_uInv := NewElementCalGFromFieldElement(uInv)
		b_u := NewElementCalGIdentity()
		b_u.Mul(tmp, g_uInv)
		if modulus != F2PolynomialZero {
			b_u = b_u.Modf(modulus)
		}
		// b_u^{-1} = u b^{-1} u^{-1}
		tmp.Mul(g_u, bInv)
		b_uInv := NewElementCalGIdentity()
		b_uInv.Mul(tmp, g_uInv)
		if modulus != F2PolynomialZero {
			b_uInv = b_uInv.Modf(modulus)
		}
		// next, the matrix reps
		uStd := uZeta.ToStandardBasis()
		uRep := ProjMatF2Poly(cartwrightStegerMatrixRepFieldElement(uStd))
		uInvRep := ProjMatF2Poly(cartwrightStegerMatrixRepFieldElement(uStd.InverseModf(cartwrightStegerModulus)))
		// b_u = u b u^{-1}
		b_uRep := uRep.Mul(repB).Mul(uInvRep)
		var yxModulus F2Polynomial = cartwrightStegerEmbedPolynomial(modulus)
		if yxModulus != F2PolynomialZero {
			b_uRep = b_uRep.ReduceModf(yxModulus)
		}
		// b_u^{-1} = u b^{-1} u^{-1}
		b_uInvRep := uRep.Mul(repBInv).Mul(uInvRep)
		if yxModulus != F2PolynomialZero {
			b_uInvRep = b_uInvRep.ReduceModf(yxModulus)
		}
		table = append(table, CartwrightStegerGenInfo{
			uStd,
			b_u, b_uInv,
			b_uRep, b_uInvRep,
		})
	}
	return table
}

func cartwrightStegerMatrixRepOnePlusBetaX() MatF2Poly {
	// We assume beta = 1 + v.  Note that for computing matrix
	// representations for other values of q or d, we would need to
	// find another appropriate value for beta.
	beta := cartwrightStegerEmbeddingBeta
	betaMat := cartwrightStegerMatrixRepFieldElement(beta)
	xBetaMat := betaMat.Scale(NewF2Polynomial("01"))
	return MatF2PolyIdentity.Add(xBetaMat)
}

func cartwrightStegerMatrixRepOneTensorPhi() MatF2Poly {
	// phi is assumed to be the Frobenius automophism v -> v^2.
	mod := cartwrightStegerModulus
	col0 := NewF2Polynomial("1").Pow(2).Modf(mod)
	col1 := NewF2Polynomial("01").Pow(2).Modf(mod)
	col2 := NewF2Polynomial("001").Pow(2).Modf(mod)
	if col0.Degree() > 2 || col1.Degree() > 2 || col2.Degree() > 2 {
		panic("degree too high")
	}
	return NewMatF2Poly(
		coeffToPoly(col0.Coefficient(0)), coeffToPoly(col1.Coefficient(0)), coeffToPoly(col2.Coefficient(0)),
		coeffToPoly(col0.Coefficient(1)), coeffToPoly(col1.Coefficient(1)), coeffToPoly(col2.Coefficient(1)),
		coeffToPoly(col0.Coefficient(2)), coeffToPoly(col1.Coefficient(2)), coeffToPoly(col2.Coefficient(2)))
}

func cartwrightStegerMatrixRepZ() MatF2Poly {
	onePlusBetaX := cartwrightStegerMatrixRepOnePlusBetaX()
	oneTensorPhi := cartwrightStegerMatrixRepOneTensorPhi()
	return onePlusBetaX.Mul(oneTensorPhi)
}

func cartwrightStegerMatrixRepB() ProjMatF2Poly {
	// b = 1 - z^{-1}
	//
	// since z^3 = 1 + y(x),
	//
	// z^{-1} = 1/(1 + y(x)) z^2
	//
	// repB = 1/(1 + y(x))((1 + y(x))I - repZ^2)
	//
	// and projectivizing, we drop the scalar factor
	//
	// projRepB = (1 + y(x))I - repZ^2
	//          = I + y(x)I + repZ^2
	repZ := cartwrightStegerMatrixRepZ()
	repZSq := repZ.Mul(repZ)
	id := MatF2PolyIdentity
	y := MatF2PolyIdentity.Scale(cartwrightStegerEmbeddingY)
	return (ProjMatF2Poly)(id.Add(y).Add(repZSq))
}

func cartwrightStegerMatrixRepBInverse() ProjMatF2Poly {
	// this was computed with Sage
	return NewProjMatF2PolyFromString("[1001 011 001 0 0101 011 011 011 0001]")
}

// Compute the matrix corresponding to multiplication by a field
// element.
func cartwrightStegerMatrixRepFieldElement(u F2Polynomial) MatF2Poly {
	// the columns of the matrix are the coefficients of
	//
	//   u*v^0  u*v^1  u*v^2
	//
	// in the {1, v, v^2} basis.
	mod := cartwrightStegerModulus
	col0 := u.Modf(mod)
	col1 := u.Mul(NewF2Polynomial("01")).Modf(mod)
	col2 := u.Mul(NewF2Polynomial("001")).Modf(mod)
	if col0.Degree() > 2 || col1.Degree() > 2 || col2.Degree() > 2 {
		panic("degree too high")
	}
	return NewMatF2Poly(
		coeffToPoly(col0.Coefficient(0)), coeffToPoly(col1.Coefficient(0)), coeffToPoly(col2.Coefficient(0)),
		coeffToPoly(col0.Coefficient(1)), coeffToPoly(col1.Coefficient(1)), coeffToPoly(col2.Coefficient(1)),
		coeffToPoly(col0.Coefficient(2)), coeffToPoly(col1.Coefficient(2)), coeffToPoly(col2.Coefficient(2)))
}

func coeffToPoly(c int) F2Polynomial {
	switch c {
	case 0:
		return NewF2Polynomial("0")
	case 1:
		return NewF2Polynomial("1")
	default:
		panic("invalid coefficient")
	}
}

func cartwrightStegerEmbedPolynomial(f_of_y F2Polynomial) (f_of_y_of_x F2Polynomial) {
	var sum F2Polynomial = F2PolynomialZero
	for i := 0; i <= f_of_y.Degree(); i++ {
		if f_of_y.Coefficient(i) == 1 {
			sum = sum.Add(cartwrightStegerEmbeddingY.Pow(i))
		}
	}
	return sum
}
