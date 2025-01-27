package golsv

import (
	"log"
	"github.com/cloud9-tools/go-galoisfield"
)

type LsvContext struct {
	GF *galoisfield.GF
	modulus F2Polynomial
}

func NewLsvContext(baseField string) *LsvContext {
	ctx := &LsvContext{}
	switch baseField {
	case "F16":
		ctx.GF = galoisfield.Poly410_g2 // x^4 + x + 1
		ctx.modulus = NewF2Polynomial("11001")
	case "F8":
		// important: modulus "1101", i.e. x^3 + x + 1 does NOT work.
		ctx.GF = galoisfield.New(8, 0xd, 2) // x^3 + x^2 + 1
		ctx.modulus = NewF2Polynomial("1011")
	case "F4":
		ctx.GF = galoisfield.Poly210_g2 // x^2 + x + 1
		ctx.modulus = NewF2Polynomial("111")
	case "F2":
		panic("F2 not supported")
	default:
		log.Fatalf("unknown base field: %s", baseField)
	}
	// log.Printf("base field: %v", ctx.GF)
	return ctx
}

func LsvContextSupportedBaseFields() []string {
	// package galoisfield doesn't support F2.  we would have to work
	// around that. for example, we could replace MatGF with MatF2Poly
	// in CayleyExpander (or replace it with an interface and type
	// parameter).
	return []string{"F4", "F8", "F16"}
}

// xxx not sure this is accurate. [LSV] and [EKZ] are not clear.
// func (lsv *LsvContext) algebraSplits() bool {
// 	// verify by enumeration that we have a d^th root of 1+x in the
// 	// field extension.
// 	n := lsv.GF.Size()
// 	d := 3
// 	want := byte(3) // 1+x
// 	for i := uint(0); i < n; i++ {
// 		f := byte(i)
// 		g := gfpow(lsv.GF, f, d)
// 		if g == want {
// 			return true
// 		}
// 	}
// 	return false
// }

func gfpow(gf *galoisfield.GF, x byte, n int) byte {
	y := byte(1)
	for i := 0; i < n; i++ {
		y = gf.Mul(y, x)
	}
	return y
}

func (lsv *LsvContext) originalCartwrightStegerGenerators() []MatGF {
	// In the case of the [LSV section 10 example] with e=4, the
	// Cartwright-Steger generators are the same as the LSV
	// generators. But in general they're not.  The LSV generators are
	// the Cartwright-Steger generators mod some polynomial.  these
	// are labelled b_0, ..., b_6 in [LSV section 10].
	return []MatGF{
		// originally printed from test_cayley_graph_generator.py
		{
			10, 4, 6, 2, 8, 7, 6, 5, 9,
		},
		{
			15, 6, 5, 3, 12, 1, 5, 2, 8,
		},
		{
			13, 5, 2, 7, 10, 4, 2, 3, 12,
		},
		{
			14, 2, 3, 1, 15, 6, 3, 7, 10,
		},
		{
			9, 3, 7, 4, 13, 5, 7, 1, 15,
		},
		{
			8, 7, 1, 6, 14, 2, 1, 4, 13,
		},
		{
			12, 1, 4, 5, 9, 3, 4, 6, 14,
		},
	}
}

// take the original Cartwright-Steger generators, reduce their
// entries over the base field, find their canonical representatives,
// and return them as a symmetric set, meaning include their inverses.
func (lsv *LsvContext) Generators() []MatGF {
	var a [14]MatGF
	csGens := lsv.originalCartwrightStegerGenerators()
	for i, g := range csGens {
		h := NewReducedMatGF(lsv, &g)
		h.MakeCanonical(lsv)
		a[i] = *h
	}
	offset := len(csGens)
	for i := 0; i < offset; i++ {
		g := a[i]
		a[offset+i] = *g.Inverse(lsv)
	}
	return a[:]
}

// Here we use the Cartwright-Steger generators computed ourselves.
func (lsv *LsvContext) GeneratorsV2() []MatGF {
	gens, _ := CartwrightStegerGeneratorsMatrixReps()
	if len(gens) != 14 {
		log.Fatalf("unexpected number of generators: %d", len(gens))
	}
	// reduce by the modulus of the base field
	reduced := make([]ProjMatF2Poly, len(gens))
	for i, g := range gens {
		reduced[i] = g.ReduceModf(lsv.modulus)
	}
	// convert to MatGF normalized form (canonical representative).
	converted := make([]MatGF, len(reduced))
	for i, g := range reduced {
		mgf := NewMatGFFromProjMatF2Poly(lsv, g)
		if mgf.Determinant(lsv) == 0 {
			panic("generator has zero determinant")
		}
		mgf.MakeCanonical(lsv)
		converted[i] = mgf
	}
	return converted
}

// xxx test
func LsvGeneratorIndex(lsv *LsvContext, m *MatGF) int {
	for i, g := range lsv.Generators() {
		if g.Equal(lsv, m) {
			return i
		}
	}
	return -1
}

// xxx wip
// func LsvGeneratorsLatexReference() string {
// 	s := fmt.Sprintf("Field: $%v$\n\n", GF)
// 	s += "\\begin{tabular}{|l|l|l|}\n"
// 	s += "\\hline\n"
// 	s += "& MatGF & Canonical Representative \\\\\n"
// 	s += "\\hline\n"
// 	for i, b := range LsvExampleGenerators {
// 		c := NewReducedMatGF(&b)
// 		cLatexMatGF := c.Latex()
// 		cRep := c.Copy()
// 		cRep.MakeCanonical()
// 		cRepLatexMatGF := cRep.Latex()
// 		s += fmt.Sprintf("$b_{%d}$ & $%s$ & $%s$ \\\\\n", i, cLatexMatGF, cRepLatexMatGF)
// 	}
// 	s += "\\hline\n"
// 	s += "\\end{tabular}\n\n"
// 	return s
// }

