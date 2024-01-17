package golsv

import (
	"testing"
)

func TestAllTrianglesContainingEdge(t *testing.T) {
	T := LsvTrianglesAtOrigin(lsv)
    for _, u := range T {
		for _, e := range u.Edges() {
			triangles := AllTrianglesContainingEdge(lsv, T, e)
			if len(triangles) != 3 {
				t.Errorf("expected 3 triangles, got %d for edge %v", len(triangles), e)
			}
		}
	}
}

// xxx wip
func TestAllTrianglesSharingEdge(t *testing.T) {
	T := LsvTrianglesAtOrigin(lsv)
// 	log.Printf("|T|=%d", len(T))
// 	log.Printf("example triangle: %v", exampleTriangle)
	triangles := AllTrianglesSharingEdgeWithTriangle(lsv, T, exampleTriangle)
	// xxx check count
	ComplexFromTriangles(append(triangles, exampleTriangle), true, false)
// 	log.Printf("C: %v", C)
// 	log.Printf("bases:\n%v", C.DumpBases())
}

func TestTrianglesContainingEachEdgeAtOrigin(t *testing.T) {
	T := LsvTrianglesAtOrigin(lsv)
	allTriangles := make([]Triangle, 0)
	for _, g := range lsv.Generators() {
		e := NewEdge(*MatGfIdentity, g)
		triangles := AllTrianglesContainingEdge(lsv, T, e)
		if len(triangles) != 3 {
			t.Errorf("expected 3 triangles, got %d", len(triangles))
		}
		allTriangles = append(allTriangles, triangles...)
	}
	C := ComplexFromTriangles(allTriangles, true, false)
// 	log.Printf("C: %v", C)
// 	log.Printf("bases:\n%v", C.DumpBases())
	if C.NumVertices() != 15 {
		t.Errorf("expected 15 vertices, got %d", C.NumVertices())
	}
	if C.NumEdges() != 35 {
		t.Errorf("expected 35 edges, got %d", C.NumEdges())
	}
	if C.NumTriangles() != 21 {
		t.Errorf("expected 21 triangles, got %d", C.NumTriangles())
	}
}

// xxx disabled; takes too long
// func Test273Cycle(t *testing.T) {
// 	// let g be a generator from the set of generators for the LSV
// 	// cayley complex.  g generates a cyclic subgroup.  let alpha be
// 	// the corresponding path.  we want to know whether alpha is in
// 	// the image of d_2.  if it is not in the image of d_2, then
// 	// the systole is no more than the length of alpha.
// 	g := LsvExampleGeneratorsCanonicalSymmetric[0]
// 	var alpha Path = CyclicSubgroupPath(&g)

// 	// for checking whether some other arbitrary cycle alpha in in the
// 	// image of d_2, it would be necessary to either compute d_2
// 	// entirely, or to compute a local version of d_2 restricted to
// 	// only the relevant triangles.  our original algorithm to compute
// 	// the relevant triangles was flawed; the original algorithm finds
// 	// all triangles that share an edge with alpha, but this is not
// 	// sufficient.  we need to find all triangles that share an edge
// 	// with alpha, and then all triangles that share an edge with
// 	// those triangles, and so on.  the latter could be viewed as
// 	// examining the "triangle graph" as analogy to the standard "line
// 	// graph" of a graph.  we have yet to implement the latter; and it
// 	// is not clear that it would reduce the size of d_2 sufficiently
// 	// to be useful.
// 	CycleIsInD2ImageGraded(alpha, false)
// }

// xxx ungraded expansion leads to full expansion
// func TestAll4Cycles(t *testing.T) {
// 	cycles := FindLsvCycles(4, 4)
// 	log.Printf("found %d cycles", len(cycles))
// 	for _, alpha := range cycles {
// 		log.Printf("alpha: %v", alpha)
// 		CycleIsInD2Image(alpha)
// 	}
// }

// xxx succeeds but `exhausted` accounting is still suspect
// deprecated
// func TestAll4CyclesGraded(t *testing.T) {
// 	cycles := FindLsvCycles(lsv, 4, 4)
// 	// log.Printf("found %d cycles", len(cycles))
// 	for _, alpha := range cycles {
// 		verbose := false
// 		found := CycleIsInD2ImageGraded(lsv, alpha, true, verbose)
// 		if !found {
// 			t.Errorf("did not find alpha in image of d_2: %v", alpha)
// 		}
// 	}
// }

// xxx wip; testing hypothesis that there is a bug preventing the
// preimage of a 4-cycle from being found.
// func Test4CycleTriangulation(t *testing.T) {
// 	cycles := FindLsvCycles(4, 4)
//     alpha := cycles[0]
// 	log.Printf("alpha: %v", alpha)
// 	Y := ComplexFromGradedTriangles(alpha, 0, true)
// 	log.Printf("Y: %v", Y)
// 	log.Printf("bases:\n%v", Y.DumpBases())
// 	v := Y.PathToVector(alpha)
// 	log.Printf("v: %v", v)
// }
