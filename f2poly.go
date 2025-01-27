package golsv

import (
	"fmt"
	"math/rand"
)

func init() {
	initInverseModfTable()
}

// \F_2[y]
//
// We would like the ElementCalG type to be comparable (in the native
// Go sense) so that we can use it as a map key.  since ElementCalG is
// an array of nine F2Polynomials, we need F2Polynomial to be
// comparable.

const f2PolynomialSize = 2

type F2Polynomial struct {
	w [f2PolynomialSize]uint64
}

func smallF2Polynomial(p uint64) F2Polynomial {
	q := F2Polynomial {}
	q.w[0] = p
	return q
}

// arrays cannot be constant, careful not to reassign constants
var F2PolynomialZero = smallF2Polynomial(0)

var F2PolynomialOne = smallF2Polynomial(1)
var F2PolynomialY = smallF2Polynomial(2)
var F2PolynomialOnePlusY = smallF2Polynomial(3)
var F2Polynomial111 = smallF2Polynomial(7)

func checkF2PolynomialDegree(degree int) {
	if degree >= f2PolynomialSize * 64 {
		panic(fmt.Sprintf("F2Polynomial degree %d too high", degree))
	}
}

func NewF2Polynomial(s string) F2Polynomial {
	if len(s) > f2PolynomialSize * 64 {
		panic("NewF2Polynomial: too many terms")
	}
	p := F2Polynomial {}
	for i, k := range s {
		b := uint64(0)
		if k == '1' {
			b = 1
		} else if k != '0' {
			panic(fmt.Sprintf("unrecognized rune %c", k))
		}
		p.w[i/64] |= (b << (i%64))
	}
	return p
}

func NewF2PolynomialFromSupport(support ...int) F2Polynomial {
	p := F2Polynomial {}
	for _, k := range support {
		checkF2PolynomialDegree(k)
		p.w[k/64] |= (uint64(1) << (k%64))
	}
	return p
}

func NewF2PolynomialRandom(maxDegree int) F2Polynomial {
	checkF2PolynomialDegree(maxDegree)
	p := F2Polynomial {}
	highestWord := maxDegree/64
	for i := 0; i < highestWord; i++ {
		p.w[i] = rand.Uint64()
	}
	p.w[highestWord] = rand.Uint64() >> (63-(maxDegree%64))
	return p
}

func EnumerateF2Polynomials(maxDegree int) []F2Polynomial {
	if maxDegree < 0 {
		panic("EnumerateF2Polynomials: negative degree")
	}
	n := 1 << (maxDegree + 1)
	ps := make([]F2Polynomial, n)
	for i := 0; i < n; i++ {
		for j := 0; j <= maxDegree + 1; j++ {
			if (i >> j) & 1 == 1 {
				ps[i] = ps[i].AddMonomial(j)
			}
		}
	}
	return ps
}

// returns a + b
func (p F2Polynomial) Add(q F2Polynomial) F2Polynomial {
	r := F2Polynomial {}
	for i := 0; i < f2PolynomialSize; i++ {
		r.w[i] = p.w[i] ^ q.w[i]
	}
	return r
}

func (p F2Polynomial) AddMonomial(degree int) F2Polynomial {
	checkF2PolynomialDegree(degree)
	r := p
	r.w[degree/64] ^= (uint64(1) << (degree%64))
	return r
}

func (p F2Polynomial) Coefficient(degree int) int {
	if degree < 0 || degree >= f2PolynomialSize * 64 {
		return 0
	}
	return int((p.w[degree/64] >> (degree%64)) & 1)
}

func highestOneBit(p uint64) int {
	for i := 63; i >= 0; i-- {
		if (p >> i) & 1 == 1 {
			return i
		}
	}
	return -1
}

func (p F2Polynomial) Degree() int {
	for i := f2PolynomialSize-1; i >= 0; i-- {
		if p.w[i] != 0 {
			return highestOneBit(p.w[i]) + i * 64
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

func (p F2Polynomial) Dup() F2Polynomial {
	return p
}

func (p F2Polynomial) Equal(q F2Polynomial) bool {
	return p == q
}

var inverseModfTable map[F2Polynomial]map[F2Polynomial]F2Polynomial

func initInverseModfTable() {
	// log.Printf("initializing InverseModf lookup table")
	fs := []struct {
		f, g F2Polynomial
	}{
		{smallF2Polynomial(7), F2PolynomialY},
		{smallF2Polynomial(11), F2PolynomialY},
	}
	inverseModfTable = make(map[F2Polynomial]map[F2Polynomial]F2Polynomial)
	for _, poly := range fs {
		f, g := poly.f, poly.g
		invg := g
		for i := 1; i < (1 << f.Degree())-2; i++ {
			invg = invg.Mul(g).Modf(f)
			if invg.Equal(F2PolynomialOne) {
				panic(fmt.Sprintf("%v is not a generator mod %v", g, f))
			}
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
	for i := 1; i < f2PolynomialSize; i++ {
		if p.w[i] != 0 {
			return false
		}
	}
	return p.w[0] == 1
}

func (p F2Polynomial) IsZero() bool {
	for i := 0; i < f2PolynomialSize; i++ {
		if p.w[i] != 0 {
			return false
		}
	}
	return true
}

func (p F2Polynomial) Latex() string {
	if p.IsZero() {
		return "0"
	}
	var s string
	for i := 0; i <= p.Degree(); i++ {
		if p.Coefficient(i) == 1 {
			if s != "" {
				s += "+"
			}
			if i == 0 {
				s += "1"
			} else {
				s += "v"
				if i > 1 {
					s += "^{" + fmt.Sprintf("%d", i) + "}"
				}
			}
		}
	}
	return s
}

func (p F2Polynomial) Less(q F2Polynomial) bool {
	for i := f2PolynomialSize-1; i >= 0; i-- {
		if p.w[i] < q.w[i] {
			return true
		} else if p.w[i] > q.w[i] {
			return false
		}
	}
	return false
}

// returns the highest power of y that divides p
func (p F2Polynomial) MaxYFactor() int {
	for i := 0; i < f2PolynomialSize * 64; i++ {
		if (p.w[i/64] >> (i%64)) & 1 == 1 {
			return i
		}
	}
	return 0
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
	if p.Degree() + q.Degree() >= f2PolynomialSize * 64 {
		panic("F2Polynomial product overflow")
	}
	r := F2Polynomial {}
	w := q.w
	for i := 0; i <= p.Degree(); i++ {
		if p.Coefficient(i) == 1 {
			for j := 0; j < f2PolynomialSize; j++ {
				r.w[j] ^= w[j]
			}
		}
		for j := f2PolynomialSize-1; j >= 1; j-- {
			w[j] = (w[j] << 1) | (w[j-1] >> 63)
		}
		w[0] = w[0] << 1
	}
	return r
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
	if p.IsZero() {
		return "0"
	}
	runes := make([]rune, p.Degree()+1)
	for i := 0; i <= p.Degree(); i++ {
		if p.Coefficient(i) == 0 {
			runes[i] = '0'
		} else {
			runes[i] = '1'
		}
	}
	return string(runes)
}
