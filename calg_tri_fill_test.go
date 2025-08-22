package golsv

import (
	"testing"
)

func TestCalGTriangleFillerTriangleOrder(t *testing.T) {
	gens := CartwrightStegerGenerators()
	maxDepth := 2
	verbose := false
	var modulus F2Polynomial = NewF2Polynomial("111")
	quotient := true
	var observer CalGObserver
	checkPSL := false
	E := NewCalGCayleyExpander(gens, maxDepth, verbose, &modulus, quotient, observer, checkPSL)
	E.Graph()
	edgeChecks := true
	F := NewCalGTriangleFiller(E.vertexBasis, E.edgeBasis, E.gens, E.verbose, E.modulus, E.quotient, edgeChecks)
	C := F.Complex()
	triangles := C.TriangleBasis()

	for i, r := range triangles {
		if i == len(triangles)-1 {
			break
		}
		s := triangles[i+1]
		if !F.triangleLessByVertexBasis(r, s) {
			t.Errorf("Triangle order failed: r=%v s=%v", r, s)
		}
	}
}

