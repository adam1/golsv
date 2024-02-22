package golsv

import (
	"testing"
)

// a shared context for testing
var lsv *LsvContext = NewLsvContext("F16")

func TestLsvGenerators(t *testing.T) {
	if len(lsv.originalCartwrightStegerGenerators()) != 7 {
		t.Error()
	}
	if len(lsv.Generators()) != 14 {
		t.Error()
	}
	// ensure that all of the generators in
	// LsvExampleGeneratorsCanonicalSymmetric are canonical
	for i, b := range lsv.Generators() {
		// fmt.Printf("b%d: %v\n", i, b)
		if !b.IsCanonical() {
			t.Errorf("b%d is not canonical: %v", i, b)
		}
	}
    gens := lsv.Generators()
	for i, b := range lsv.originalCartwrightStegerGenerators() {
		// verify that b is projectively equal to the corresponding
		// entry in LsvExampleGeneratorsCanonicalSymmetric
		c := gens[i]
		if !b.Equal(lsv, &c) {
			t.Errorf("b%d != c%d", i, i)
		}
		// verify that the corresponding entry in the second half of
		// gens is the inverse.
		cInv := gens[i+7]
		if !cInv.IsCanonical() {
			t.Errorf("cInv%d is not canonical", i+7)
		}
		if !c.Multiply(lsv, &cInv).IsIdentity(lsv) {
			t.Errorf("c%d * cInv%d != identity", i, i+7)
		}
	}
}

// xxx wip
func TestLsvGeneratorsInflationPair(t *testing.T) {
	gens := lsv.Generators()
	b := gens[0]
	points := make([]MatGF, 0)
    for _, g := range gens {
		gb := g.Multiply(lsv, &b)
		idx := LsvGeneratorIndex(lsv, gb)
		if idx >= 0 {
			// log.Printf("b: %v g: %v gb: %v idx: %d\n", b, g, *gb, idx)
			points = append(points, *gb)
		}
	}
}

// xxx not sure this is accurate. [LSV] and [EKZ] are not clear.
// func DisabledTestLsvContextAlgebraSplits(t *testing.T) {
// 	// verify that the chosen modulus satisfies the splitting
// 	// condition for the cartwright-steger algebra: the field
// 	// extension must have a d^th root of 1+y.
// 	baseFields := LsvContextSupportedBaseFields()
// 	for _, F := range baseFields {
// 		lsv := NewLsvContext(F)
// 		if !lsv.algebraSplits() {
// 			t.Errorf("algebra does not split for %s", F)
// 		}
// 	}
// }
