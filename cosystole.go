package golsv

import (
	"fmt"
	"log"
	"math"
	"sort"
)

type CosystoleSearchParams struct {
	PruneByCohomologyProjection bool // require incremental cocycles to be projections of global non-coboundary cocycles
	PruneByGroupAction          bool // incrementally prune all but one representative from each orbit of a group action
	// xxx possibly store a reference to the group/action here in appropriate form
	InitialSupport              bool // require cocycles to be supported on first triangle
	Verbose                     bool
}

type CosystoleSearch[T any] struct {
	C *ZComplex[T] // xxx rename to X
	params CosystoleSearchParams
	triangles []ZTriangle[T]
	Z_1 BinaryMatrix
	Zu1 BinaryMatrix // aka Z^1 aka "Z upper 1"
	Bu1 BinaryMatrix // aka B^1 aka "B upper 1"
	edgesMd map[int]struct{}
	edgePredicateMd func(row int) bool
	templateF *Sparse
	template_rdZu1 *Sparse // the r_d projection of Z^1 with an extra blank column

	// xxx deprecated
// 	Bcd BinaryMatrix // B^d
// 	Ucd BinaryMatrix // U^d
}

func NewCosystoleSearch[T any](C *ZComplex[T], Z_1 BinaryMatrix, Zu1 BinaryMatrix, Bu1 BinaryMatrix, params CosystoleSearchParams) *CosystoleSearch[T] {
	return &CosystoleSearch[T]{
		C: C,
		Z_1: Z_1,
		Zu1: Zu1,
		Bu1: Bu1,
		triangles: make([]ZTriangle[T], 0),
		edgesMd: make(map[int]struct{}),
		params: params,
	}
}

func (S *CosystoleSearch[T]) prepare() {
	if S.params.Verbose {
		log.Printf("preparing triangle list")
	}
	S.triangles = S.C.TriangleBasis()
}

// returns the minimum weight of a cochain that is a cocycle but not a
// coboundary, i.e. the value S^1(C).  if B^1 = Z^1, or if there are
// no edges, returns zero.
func (S *CosystoleSearch[T]) Cosystole() int {
	// we want to enumerate Z^1 \setminus B^1, i.e. edge sets that
	// correspond to cocycles that are not coboundaries, i.e. cochains
	// that vanish on B_1 but not on Z_1.  we will do this by choosing
	// an edge set from each triangle, such that the edge set vanishes
	// on the boundary of the triangle, which happens precisely when
	// the edge set contains an even number of edges in the triangle,
	// i.e. zero or two edges.  since there are in general multiple
	// choices for each triangle, we build a tree of choices.  there
	// is one level in the tree for each triangle, plus the root.  the
	// levels in the tree correspond to the index in the overall
	// triangles list.  each vertex in the tree corresponds to a
	// choice of edge set for the triangle at that level.  an edge in
	// the tree from a vertex v down to a child vertex u corresponds
	// to a choice of edge set for the child triangle that is
	// compatible with the choices made for the ancestor triangles in
	// the tree (which in general may have shared edges).  in this
	// way, we are effectively coloring every edge in the complex
	// either ON or OFF.
	//
	// thus the algorithm enumerates cocycles of increasing distance
	// to the first triangle in the original list, meaning mininum
	// graph distance between vertices in two triangles.
	//
	// as we build the tree, the leaf vertices correspond to choices
	// of edge sets decided thus far that vanish on B_1. note that in
	// a particular branch of the tree, going from root to leaf, an
	// edge is never turned OFF once it is turned ON. hence even
	// midway through building the tree, we can already determine that
	// a particular cocycle corresponding to a leaf has a weight not
	// less than the number of edges turned ON so far in its branch.
	// hence if we happen to know a priori an upper bound on the
	// weight of a cosystolic vector, we can prune the tree early.
	// (in practice with small LSV complexes, we might know such an
	// upper bound from the linear algebra grinder. this type of
	// pruning optimization is TBD.)
	// 
	// when we reach the end of the triangles list, all edge colorings
	// have been decided, and each leaf vertex corresponds to a
	// possible edge set.  we can then test each edge set to see if it
	// is a coboundary, i.e. to see whether it does not vanish on
	// Z_1. taking the minimum weight edge set of those that are not
	// coboundaries gives us a cosystolic vector.
	//
	// this procedure is likely exponential in the number of edges in
	// the complex, although it depends on how the possibilities
	// branch.
	cocycles := S.Cocycles()
	if S.params.Verbose {
		log.Printf("cocycles: %v\n", len(cocycles))
		// log.Printf("Z_1: %v\n%s", S.Z_1, dumpMatrix(S.Z_1))
	}
	minWeight := -1
	for _, c := range cocycles {
		if c.IsZero() {
			// special case: in our algorithm, the zero cocycle
			// corresponds to a cochain that selects no edges from any
			// triangle.  sometimes such a cocycle can be turned into
			// a cosystolic cochain of weight 1 by selecting an edge
			// that is part of some cycle but not part of any
			// triangle.  this is possible whenever dim B_1 < dim Z_1
			// *and* there exists an edge that is not part of any
			// triangle.
			d_2 := S.C.D2()
			for j := 0; j < S.Z_1.NumRows(); j++ {
				if S.Z_1.RowIsZero(j) {
					continue
				}
				if !d_2.RowIsZero(j) {
					continue
				}
				return 1
			}
			continue
		}
		// if c is not a coboundary, then it is a cosystolic
		// candidate.  record the minimum weight of such candidates.
		cMatrix := c.Matrix()
		if !isCoboundary(cMatrix, S.Z_1) {
			weight := cMatrix.ColumnWeight(0)
			if minWeight < 0 || weight < minWeight {
				minWeight = weight
				if S.params.Verbose {
					log.Printf("new min weight of noncoboundary cocycle: %d\n", minWeight)
					// log.Printf("xxx c: %s", c.SupportString())
				}
			}
		}
	}
	if minWeight < 0 {
		minWeight = 0
	}
	if S.params.Verbose {
		log.Printf("cosystole: %d\n", minWeight)
	}
	return minWeight
}

func isCoboundary(c, Z1 BinaryMatrix) bool {
	// determine whether this cochain is a coboundary, i.e. whether it
	// vanishes on Z_1.  do this by the matrix multiplication
	//
	//   c^T * Z_1
	//
	// where c is the cocycle, and Z_1 is the matrix of basis vectors
	// of Z_1.
	cT := c.Transpose()
	dense := cT.Dense()
	result := dense.MultiplyRight(Z1)
// 	log.Printf("xxx result: %s w(result): %v isZero=%v", result,
// 		result.Transpose().ColumnWeight(0), result.IsZero())
	return result.IsZero()
}

func isCoboundary2(c *Sparse, Z_1T *DenseBinaryMatrix) bool {
	return Z_1T.MultiplyRight(c).IsZero()
}

func isCocycle(c, B1 BinaryMatrix) bool {
	cT := c.Transpose()
	dense := cT.Dense()
	result := dense.MultiplyRight(B1)
	return result.IsZero()
}

func (S *CosystoleSearch[T]) Cocycles() (cocycles []BinaryVector) {
	S.prepare()
	if S.params.Verbose {
		log.Printf("begin cocycle enumeration; triangles: %v", len(S.triangles))
	}
	// special case: no triangles.  in this case, every cochain
	// vacuously vanishes on the boundary, hence is a cocycle.
	if len(S.triangles) == 0 {
		return AllBinaryVectors(len(S.C.EdgeBasis()))
	}
	root := newStateNode(nil)
	leaves := []*StateNode{root}
	
	for level, t := range S.triangles {
		if S.params.Verbose {
			log.Printf("begin processing level %d: t=%v leaves=%d\n", level, t, len(leaves))
			// log.Printf(" t edges: %v\n", t.Edges())
		}
		// for this triangle, determine the branching at each current
		// leaf vertex.  the current leaf vertices represent the edge
		// set choices for all previous triangles.
		newLeaves := make([]*StateNode, 0)
		for _, leaf := range leaves {
			var newStates [][3]bool
			edgeStates := edgeStateForTriangle(S.triangles, leaf, level)
			// 			if S.params.Verbose {
			// 				log.Printf("  leaf %d.%d [%v]: edgeStates: %v", level, j, leaf, edgeStates)
			// 			}
			_, _, undecided := numPerState(edgeStates)
			{
				// temporary recovery for debugging
				// 				defer func() {
				// 					if r := recover(); r != nil {
				// 						log.Printf("panic: %v\n", r)
				// 						log.Printf("  level: %d\n", level)
				// 						log.Printf("  leaf: %v\n", leaf)
				// 						log.Printf("  edgeStates: %v\n", edgeStates)
				// 						log.Printf("  undecided: %d\n", undecided)
				// 						S.dumpBranch(leaf, level)
				// 						os.Exit(1)
				// 					}
				// 				}()
				switch undecided {
				case 0:
					newStates = transform0Undecided(edgeStates)
				case 1:
					newStates = transform1Undecided(edgeStates)
				case 2:
					newStates = transform2Undecided(edgeStates)
				case 3:
					newStates = transform3Undecided(edgeStates)
				}
			}
			for _, n := range newStates {
				{
					// sanity check: verify that the number of edges on is even
					sanityCheck := false
					if sanityCheck {
						on := numDecidedStatesOn(n)
						if on % 2 == 1 {
							panic("odd number of edges on")
						}
					}
				}
				q := leaf.addChild(n)
// 				if S.params.Verbose {
// 					log.Printf("    new leaf: [%v]\n", q)
// 				}
				newLeaves = append(newLeaves, q)
				sanityCheck := false
				if sanityCheck {
					// sanity check: verify that a cochain as
					// described up to this leaf vanishes on all of
					// the triangles handled up to this level.  do
					// this by truncating d2.
					truncatedD2 := S.C.D2().Submatrix(0, S.C.D2().NumRows(), 0, level+1)
					c := NewBinaryVector(len(S.C.EdgeBasis()))
					S.leafToVector(q, level, c)
					cT := c.Matrix().Transpose()
					W := cT.MultiplyRight(truncatedD2)
					if !W.IsZero() {
						log.Printf("cT: %v:\n%s", cT, dumpMatrix(cT))
						log.Printf("truncatedD2: %v:\n%s", truncatedD2, dumpMatrix(truncatedD2))
						log.Printf("W: %v:\n%s\n", W, dumpMatrix(W))
						S.dumpBranch(q, level)
						panic("cocycle does not vanish on B_1")
					}
				}
			}
		}
		if S.params.PruneByCohomologyProjection {
			if S.params.Verbose {
				log.Printf("pruning by cohomology projection")
			}
			S.incrementMd(level)
			// xxx disabled since it doesn't prune anything
			if false {
				leavesSansPruned := make([]*StateNode, 0)
				for _, leaf := range newLeaves {
					if S.leafIsProjectionOfNonzeroCohomologyClass(leaf, level) {
						// retain
						leavesSansPruned = append(leavesSansPruned, leaf)
					} else {
						// prune
					}
				}
				if S.params.Verbose {
					log.Printf("pruned %d leaves", len(newLeaves) - len(leavesSansPruned))
				}
				newLeaves = leavesSansPruned
			}
		}
		// xxx test; on small Cayley graph
		if level == 0 && S.params.InitialSupport {
			log.Printf("Pruning to require support on first triangle")
			leavesWithInitialSupport := make([]*StateNode, 0)
			for _, leaf := range newLeaves {
				c := NewBinaryVector(len(S.C.EdgeBasis()))
				S.leafToVector(leaf, level, c)
				if c.Weight() > 0 {
					leavesWithInitialSupport = append(leavesWithInitialSupport, leaf)
				}
			}
			if S.params.Verbose {
				log.Printf("pruned %d leaves", len(newLeaves) - len(leavesWithInitialSupport))
			}
			newLeaves = leavesWithInitialSupport
		}
		if (S.params.PruneByGroupAction) {
			newLeaves = S.pruneByGroupAction(newLeaves, level)
		}
		if len(newLeaves) == 0 {
			panic("All leaves were pruned!")
		}
		leaves = newLeaves
		weightDist, _, _ := S.weightDistribution(leaves, level)
		if S.params.Verbose {
			log.Printf("level %d weight distribution: %v", level, weightDist)
		}
		S.reviewWeightDistribution(weightDist, level)
	}
	for _, p := range leaves {
		c := NewBinaryVector(len(S.C.EdgeBasis()))
		S.leafToVector(p, len(S.triangles)-1, c)
		cocycles = append(cocycles, c)
	}
	return
}

func (S *CosystoleSearch[T]) pruneByGroupAction(leaves []*StateNode, level int) []*StateNode {
	newLeaves := make([]*StateNode, 0)
	// idea: iterate over the leaves, building a list of archetypes,
	// meaning a representative from each triangle orbit with specific
	// edge colorings. retain only the archetypes.

	// xxx tbd


	return newLeaves
}

func (S *CosystoleSearch[T]) reviewWeightDistribution(dist []weightSample, level int) {
	// xxx idea: forward-scan the lowest weight cochains in hopes of pruning them
}

type weightSample struct {
	weight int
	count int
}

func (S *CosystoleSearch[T]) weightDistribution(leaves []*StateNode, level int) (dist []weightSample, min int, max int) {
	weights := make(map[int]int)
	min = math.MaxInt64
	max = 0
	for _, p := range leaves {
		c := NewBinaryVector(len(S.C.EdgeBasis()))
		S.leafToVector(p, level, c)
		w := c.Weight()
		weights[w]++
		if w < min {
			min = w
		}
		if w > max {
			max = w
		}
	}
	for w, count := range weights {
		dist = append(dist, weightSample{w, count})
	}
	sort.Slice(dist, func(i, j int) bool {
		return dist[i].weight < dist[j].weight
	})
	return
}

func (S *CosystoleSearch[T]) leafIsProjectionOfNonzeroCohomologyClass(leaf *StateNode, level int) bool {
	c := NewBinaryVector(S.Z_1.NumRows())
	S.leafToVector(leaf, level, c)

	// let F = r^{-1}(\overline{c}}) where r is the restriction map
	// from C^1(X) to C^1(M_d), \bra{c} is the cocycle of M_d
	// corresponding to the leaf, and \overline{c} is the subspace
	// generated by \bra{c}.

	// first, do a quick check: whether \overline{c} \cap r(Z^1(X)) is
	// zero.  
// 	if S.params.Verbose {
// 		log.Printf("Checking E := \\overline{c} \\cap r(Z^1(X))")
// 	}
	cd := S.projectCochainToMd(c)
// 	if S.params.Verbose {
// 		log.Printf("xxx leaf: %v cd: %s", leaf, cd)
// 	}

	E := S.template_rdZu1.Copy().Sparse()
	m := E.NumColumns()
	E.SetColumn(m - 1, cd)
	// log.Printf("xxx E: %s\n%s", E, dumpMatrix(E))
	verbose := false
	_, _, _, rank := smithNormalForm(E, verbose)
	// E is full rank if and only if dim ker E = 0
	// if and only if \overline{c} \cap r(Z^1(X)) = 0.
// 	log.Printf("xxx rank=%d m=%d", rank, m)
	if rank == m {
		return false
	}

	// xxx --- below is temporarily disabled
	return true

	// second, compute basis for F = r^{-1}(\overline{c}}) where r
	// is the restriction map from C^1(X) to C^1(M_d), \bra{c} is
	// the cocycle of M_d corresponding to the leaf, and \overline{c}
	// is the subspace generated by \bra{c}.

	// the matrix F is identical for all leaves at this level, except
	// in the final column.  hence for performance, once per level we
	// prepare a template for F, and then overwrite the final column
	// here once per leaf.
	if S.params.Verbose {
		log.Printf("Preparing F")
	}
	k := S.templateF.NumColumns()
	S.templateF.SetColumn(k - 1, c)
	if S.params.Verbose {
		log.Printf("F: %v", S.templateF)
	}
	// compute the dimension of intersection of F with Z^1(X) by
	// forming the matrix G = (F | Z^1) and computing the dimension of its
	// kernel.
	if S.params.Verbose {
		log.Printf("Preparing G")
	}
	G := S.templateF.Copy().Sparse()
	G.AppendColumns(S.Zu1)
	if S.params.Verbose {
		log.Printf("Computing kernel of G: %v", G)
	}
	// xxx optimization below: we don't actually need to compute the
	// kernelBasis, only its length, which can be determined from
	// smithNormalForm.  this saves computing an automorphism which is
	// very costly.
	
	K := kernelBasis(G, S.params.Verbose)
	dimK := K.NumColumns()
	if S.params.Verbose {
		log.Printf("dimK: %d", dimK)
	}
	if dimK == 0 {
		return false
	}
	// also compute the dimension of the intersection of F with
	// B^1(X).
	if S.params.Verbose {
		log.Printf("Preparing J")
	}
	J := S.templateF.Copy().Sparse()
	J.AppendColumns(S.Bu1)
	if S.params.Verbose {
		log.Printf("Computing kernel of J: %v", J)
	}
	L := kernelBasis(J, S.params.Verbose)
	dimL := L.NumColumns()
	if S.params.Verbose {
		log.Printf("dimL: %d", dimL)
	}
	if dimL == dimK {
		return false
	}
	return true
}

func (S *CosystoleSearch[T]) projectCochainToMd(c BinaryVector) BinaryVector {
	return c.Project(len(S.edgesMd), S.edgePredicateMd)
}

func (S *CosystoleSearch[T]) dumpBranch(node *StateNode, level int) {
	p := node
	for level >= 0 {
		log.Printf("level %d triangle %v [%v]", level, S.triangles[level], p)
		p = p.parent
		level--
	}
}

func (S *CosystoleSearch[T]) leafToVector(leaf *StateNode, level int, v BinaryVector) {
	debug := false
	if debug && S.params.Verbose {
		log.Printf("--> leafToVector: %v level=%d", leaf, level)
	}
	// walk branch of the tree from leaf to root
	p := leaf
	for level >= 0 {
		t := S.triangles[level]
		if debug && S.params.Verbose {
			log.Printf("  level=%d p=%v t=%v", level, p, t)
		}
		edges := t.Edges()
		for i, e := range edges {
			if debug && S.params.Verbose {
				log.Printf("    edge %v: %v", i, e)
			}
			if p.edgeStates[i] {
				if j, ok := S.C.edgeIndex[e]; ok {
					if debug && S.params.Verbose {
						log.Printf("      edge is on; j=%d", j)
					}
					// xxx issue: the edge basis may have edges from
					// different triangles interleaved, e.g. the first
					// three edges in the edge basis do not
					// necessarily correspond to the first three edges
					// of the first triangle.
					v.Set(j, 1)
				} else {
					panic(fmt.Sprintf("edge %v not in edge index", e))
				}
			}
		}
		p = p.parent
		level--
	}
	if debug && S.params.Verbose {
		log.Printf("<-- leafToVector: leaf=%v vector=%v\n", leaf, v)
	}
}

func edgeStateForTriangle[T any](triangles []ZTriangle[T], leaf *StateNode, level int) (states [3]kEdgeState) {
	//	log.Printf("edgeStateForTriangle: level=%d\n", level)
	p := leaf
	// let t be the triangle at the current level.
	t := triangles[level]
	level--
	tEdges := t.Edges()
	for {
		// log.Printf("p=%v level=%d t=%v\n", p, level, t)
		if p == nil || level == -1 {
			break
		}
		// let u be the triangle corresponding to the current
		// StateNode.  if u shares an edge with t, copy the edge
		// state.  two triangles in a simplicial complex share at most
		// one edge.
		u := triangles[level]
		// log.Printf("  u=%v\n", u)
		uEdges := u.Edges()
		match := false
		for i, e := range tEdges {
			for j, f := range uEdges {
				if e.Equal(f) {
					// log.Printf("edge %d %v matches edge %d %v", i, e, j, f)
					// log.Printf("  edge state: %v", p.edgeStates[j])
					if p.edgeStates[j] {
						// log.Printf("edge %d %v is on", i, e)
						if states[i] == kEdgeOff {
							panic(fmt.Sprintf("inconsistent edge state: %v", states))
						}
						states[i] = kEdgeOn
					} else {
						// log.Printf("edge %d %v is off", i, e)
						if states[i] == kEdgeOn {
							panic(fmt.Sprintf("inconsistent edge state: %v", states))
						}
						states[i] = kEdgeOff
					}
					match = true
					break
				}
			}
			if match {
				break
			}
		}
		p = p.parent
		level--
	}
	return
}

func (S *CosystoleSearch[T]) incrementMd(level int) {
	t := S.triangles[level]
	for _, e := range t.Edges() {
		if k, ok := S.C.edgeIndex[e]; ok {
			S.edgesMd[k] = struct{}{}
		} else {
			panic(fmt.Sprintf("edge %v not in edge index", e))
		}
	}
	S.edgePredicateMd = func(i int) bool {
		_, ok := S.edgesMd[i]
		return ok
	}
	nd := len(S.edgesMd)
	k := S.Z_1.NumRows() - nd
	if S.params.Verbose {
		log.Printf("xxx incrementMd: level=%d edgesMd=%d", level, nd)
	}

	// update templateF templateF has one row for each edge in the
	// complex X (the columns are 1-cochains of X).  let k be the
	// number of edges in the complex X that are not in Md.  the first
	// k columns of templateF are the standard basis vectors
	// corresponding to the edges in X that are not in Md.  and there
	// is one additional column, that is blank (it is filled-in later
	S.templateF = NewSparseBinaryMatrix(S.Z_1.NumRows(), k + 1)
	j := 0
	for i := 0; i < S.templateF.NumRows() - 1; i++ {
		if _, ok := S.edgesMd[i]; ok {
			continue
		}
		S.templateF.Set(i, j, 1)
		j++
	}
	// log.Printf("xxx incrementMd: templateF=%v", S.templateF)
	verbose := false
	S.template_rdZu1 = ImageBasis(S.Zu1.Project(S.edgePredicateMd).Sparse(), verbose).Sparse()
	S.template_rdZu1.AppendColumn(NewSparseBinaryMatrix(S.template_rdZu1.NumRows(), 1))
}

// xxx deprecated
// func (S *CosystoleSearch[T]) incrementUcd(level int) {
// 	t := S.triangles[level]
// 	for _, e := range t.Edges() {
// 		if k, ok := S.C.edgeIndex[e]; ok {
// 			S.edgesUd[k] = struct{}{}
// 		} else {
// 			panic(fmt.Sprintf("edge %v not in edge index", e))
// 		}
// 	}
// 	S.edgePredicateMd = func(i int) bool {
// 		_, ok := S.edgesUd[i]
// 		return ok
// 	}
// 	{
// 		var s string
// 		for i := 0; i < S.Uu1.NumRows(); i++ {
// 			if S.edgePredicateMd(i) {
// 				s += fmt.Sprintf("* %d\n", i)
// 			} else {
// 				s += fmt.Sprintf("  %d\n", i)
// 			}
// 		}
// 		// log.Printf("xxx projected rows:\n%s", s)
// 	}
// 	// xxx tbd
// // 	S.Ucd = S.Uu1.Project(S.edgePredicateMd).IndependentColumns()
// // 	S.Bcd = S.Bu1.Project(S.edgePredicateMd).IndependentColumns()
// 	S.Ucd = S.Uu1.Project(S.edgePredicateMd)
// 	S.Bcd = S.Bu1.Project(S.edgePredicateMd)
// // 	log.Printf("xxx Ucd: %v\n%s", S.Ucd, dumpMatrix(S.Ucd))
// // 	log.Printf("xxx Bcd: %v\n%s", S.Bcd, dumpMatrix(S.Bcd))
// 	if S.params.Verbose {
// 		log.Printf("xxx Ucd: %v", S.Ucd)
// 		log.Printf("xxx Bcd: %v", S.Bcd)
// 	}
// }

func numPerState(states [3]kEdgeState) (off, on, undecided int) {
	for _, c := range states {
		switch c {
		case kEdgeOff:
			off++
		case kEdgeOn:
			on++
		case kEdgeUndecided:
			undecided++
		}
	}
	return
}

func edgeStatesToDecidedStates(states [3]kEdgeState) (bools [3]bool) {
	for i, c := range states {
		switch c {
		case kEdgeOff:
			bools[i] = false
		case kEdgeOn:
			bools[i] = true
		case kEdgeUndecided:
			panic("unexpected")
		}
	}
	return
}

func numDecidedStatesOn(bools [3]bool) int {
	n := 0
	for _, b := range bools {
		if b {
			n++
		}
	}
	return n
}

func transform0Undecided(edgeStates [3]kEdgeState) (newStates [][3]bool) {
	_, on, undecided := numPerState(edgeStates)
	if undecided != 0 {
		panic("unexpected")
	}
	if on == 1 || on == 3 {
		// this means that this branch is not possible and
		// should be pruned.
		// log.Printf("consistency error, pruning branch")
	} else {
		newStates = append(newStates, edgeStatesToDecidedStates(edgeStates))
	}
	return
}

func transform1Undecided(edgeStates [3]kEdgeState) (newStates [][3]bool) {
	off, on, undecided := numPerState(edgeStates)
	if undecided != 1 {
		panic("unexpected")
	}
	var newState [3]bool
	if off == 2 && on == 0 {
		// no-op
	} else if off == 1 && on == 1 {
		for i := 0; i < 3; i++ {
			switch edgeStates[i] {
			case kEdgeOff:
				newState[i] = false
			case kEdgeOn:
				newState[i] = true
			case kEdgeUndecided:
				newState[i] = true
			}
		}

	} else if off == 0 && on == 2 {
		for i := 0; i < 3; i++ {
			switch edgeStates[i] {
			case kEdgeOff:
				newState[i] = false
			case kEdgeOn:
				newState[i] = true
			case kEdgeUndecided:
				newState[i] = false
			}
		}
	} else {
		panic("consistency error")
	}
	newStates = append(newStates, newState)
	return
}

func transform2Undecided(edgeStates [3]kEdgeState) (newStates [][3]bool) {
	off, on, undecided := numPerState(edgeStates)
	if undecided != 2 {
		panic("unexpected")
	}
	var newState [3]bool
	if off == 1 && on == 0 {
		// two choices: all off or the two undecided are on
		newState = [3]bool{false, false, false}
		newStates = append(newStates, newState)

		for i := 0; i < 3; i++ {
			if edgeStates[i] == kEdgeUndecided {
				newState[i] = true
			}
		}
		newStates = append(newStates, newState)

	} else if off == 0 && on == 1 {
		// two choices: either of the undecideds is on
		firstUnd, secondUnd, onInd := -1, -1, -1
		for i, c := range edgeStates {
			if c == kEdgeUndecided {
				if firstUnd == -1 {
					firstUnd = i
				} else {
					secondUnd = i
				}
			} else if c == kEdgeOn {
				onInd = i
			}
		}
		newState[firstUnd] = true
		newState[secondUnd] = false
		newState[onInd] = true
		newStates = append(newStates, newState)

		newState[firstUnd] = false
		newState[secondUnd] = true
		newState[onInd] = true
		newStates = append(newStates, newState)
	} else {
		panic("consistency error")
	}
	return
}

func transform3Undecided(edgeStates [3]kEdgeState) (newStates [][3]bool) {
	_, _, undecided := numPerState(edgeStates)
	if undecided != 3 {
		panic("unexpected")
	}
	// four choices: all off, any two on
	newState := [3]bool{false, false, false}
	newStates = append(newStates, newState)

	for i := 0; i < 3; i++ {
		// index i will be off
		newState = [3]bool{false, false, false}
		newState[i] = false
		newState[(i+1)%3] = true
		newState[(i+2)%3] = true
		newStates = append(newStates, newState)
	}
	return
}

type kEdgeState int8

const (
	kEdgeOff kEdgeState = -1
	kEdgeUndecided kEdgeState = 0
	kEdgeOn kEdgeState = 1
)

// xxx tbd mem optimization: pack [3]bool into a single int8
type StateNode struct {
	edgeStates [3]bool
	parent *StateNode
}

func newStateNode(parent *StateNode) *StateNode {
	return &StateNode{
		parent: parent,
	}
}

func (p *StateNode) addChild(edgeStates [3]bool) *StateNode {
	q := newStateNode(p)
	q.edgeStates = edgeStates
	return q
}

func (p *StateNode) String() string {
	s := ""
	for _, c := range p.edgeStates {
		if c {
			s += "1"
		} else {
			s += "0"
		}
	}
	return s
}

