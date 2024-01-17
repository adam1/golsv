package golsv

import (
	"log"
	"github.com/cloud9-tools/go-galoisfield"
)

type LsvContext struct {
	GF *galoisfield.GF
}

func NewLsvContext(baseField string) *LsvContext {
	ctx := &LsvContext{}
	switch baseField {
	case "F16":
		ctx.GF = galoisfield.Poly410_g2 // x^4 + x + 1
	case "F4":
		ctx.GF = galoisfield.Poly210_g2 // x^2 + x + 1
	default:
		log.Fatalf("unknown base field: %s", baseField)
	}
	// log.Printf("base field: %v", ctx.GF)
	return ctx
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

