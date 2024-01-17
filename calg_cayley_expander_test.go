package golsv

import (
	"testing"
)

func TestCalGCayleyExpanderInverse(t *testing.T) {
	gens := CartwrightStegerGenerators()
	maxDepth := 2
	verbose := false
	var modulus *F2Polynomial = nil
	quotient := false
	E := NewCalGCayleyExpander(gens, maxDepth, verbose, modulus, quotient, nil)
	E.Expand()

	pathLenCounts := make([]int, maxDepth+1)
	
	for g, _ := range E.attendance {
		gInv, gInvPath := E.elementInverse(g)
		p := NewElementCalGIdentity()
		p.Mul(g, gInv)
		if !p.IsIdentity() {
			t.Errorf("Inverse failed;\ng: %v\ngInv: %v\np: %v", g, gInv, p)
		}
		n := len(gInvPath)
		pathLenCounts[n]++
		if g.IsIdentity() {
			continue
		}
		end := gInvPath[len(gInvPath)-1]
		if !end.Contains(NewElementCalGIdentity()) {
			t.Errorf("Inverse path end does not contain identity: g: %v\nend: %v", g, end)
		}
	}
	if pathLenCounts[0] != 1 {
		t.Errorf("Inverse path length 0 count not 1: %v", pathLenCounts[0])
	}
	if pathLenCounts[1] != 14 {
		t.Errorf("Inverse path length 1 count not 14: %v", pathLenCounts[1])
	}
	if pathLenCounts[2] != 98 {
		t.Errorf("Inverse path length 2 count not 98: %v", pathLenCounts[2])
	}
}

func TestCalGCayleyExpanderInverseModf(t *testing.T) {
	gens := CartwrightStegerGenerators()
	maxDepth := 2
	verbose := false
	var modulus F2Polynomial = NewF2Polynomial("111")
	quotient := true
	E := NewCalGCayleyExpander(gens, maxDepth, verbose, &modulus, quotient, nil)
	E.Expand()

	pathLenCounts := make([]int, maxDepth+1)

	for g, _ := range E.attendance {
		gInv, gInvPath := E.elementInverse(g)
		p := NewElementCalGIdentity()
		p.Mul(g, gInv)
		if !p.IsIdentityModf(modulus) {
			t.Errorf("Inverse failed;\ng: %v\ngInv: %v\np: %v", g, gInv, p)
		}
		n := len(gInvPath)
		pathLenCounts[n]++
		if g.IsIdentity() {
			continue
		}
		end := gInvPath[len(gInvPath)-1]
		if !end.Contains(NewElementCalGIdentity()) {
			t.Errorf("Inverse path end does not contain identity: g: %v\nend: %v", g, end)
		}
	}
	if pathLenCounts[0] != 1 {
		t.Errorf("Inverse path length 0 count not 1: %v", pathLenCounts[0])
	}
	if pathLenCounts[1] != 14 {
		t.Errorf("Inverse path length 1 count not 14: %v", pathLenCounts[1])
	}
	if pathLenCounts[2] != 98 {
		t.Errorf("Inverse path length 2 count not 98: %v", pathLenCounts[2])
	}
}
