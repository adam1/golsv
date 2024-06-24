package golsv

import (
	"fmt"
	"log"
	"math"
)

type CosystoleSearch[T any] struct {
	C *ZComplex[T]
	triangles []ZTriangle[T]
	edgeMask map[ZEdge[T]]struct{}
	verbose bool
}

func NewCosystoleSearch[T any](C *ZComplex[T], verbose bool) *CosystoleSearch[T] {
	return &CosystoleSearch[T]{
		C: C,
		triangles: make([]ZTriangle[T], 0),
		verbose: verbose,
	}
}

func (S *CosystoleSearch[T]) prepare() {
	// in principle, our cosystole search algorithm could work with
	// any ordering of the triangles in the complex.  however, for
	// ease of interaction with the boundary matrices, which refer to
	// a specific triangle ordering (basis), we currently use the
	// triangle basis as our ordering.
	if len(S.triangles) == 0 {
		S.triangles = S.C.TriangleBasis()
		// default edge mask - all edge on
		S.edgeMask = make(map[ZEdge[T]]struct{})
		for _, e := range S.C.EdgeBasis() {
			S.edgeMask[e] = struct{}{}
		}
	}
}

// returns the minimum weight of a cochain that is a cocycle but not a
// coboundary, i.e. the value S^1(C).  if B^1 = Z^1, or if there are
// no edges, returns zero.
func (S *CosystoleSearch[T]) Cosystole(Z_1 BinaryMatrix) int {
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
	if S.verbose {
		log.Printf("cocycles: %v\n", len(cocycles))
		log.Printf("Z_1: %v\n%s", Z_1, dumpMatrix(Z_1))
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
			for j := 0; j < Z_1.NumRows(); j++ {
				if Z_1.RowIsZero(j) {
					continue
				}
				if !d_2.RowIsZero(j) {
					continue
				}
				return 1
			}
			continue
		}


		// determine whether this cocycle is a coboundary,
		// i.e. whether it vanishes on Z_1.  do this by the matrix
		// multiplication
		//
		//   c^T * Z_1
		//
		// where c is the cocycle, and Z_1 is the matrix of basis
		// vectors of Z_1. if the result is not zero, then c is a
		// cosystolic candidate.  record the minimum weight of such
		// candidates.
		cMatrix := c.Matrix()
		cT := cMatrix.Transpose()
		result := cT.MultiplyRight(Z_1)
// 		log.Printf("cT: %v:\n%s\nr: %v:\n%s\n",
// 			cT, dumpMatrix(cT), result, dumpMatrix(result))
		if !result.IsZero() {
			weight := cMatrix.ColumnWeight(0)
			if minWeight < 0 || weight < minWeight {
				minWeight = weight
				if S.verbose {
					log.Printf("new min weight: %d\n", minWeight)
				}
			}
		}
		// sanity check: verify that c vanishes on B_1
		sanityCheck := false
		if sanityCheck {
			W := cT.MultiplyRight(S.C.D2())
			if !W.IsZero() {
				log.Printf("W: %v:\n%s\n", W, dumpMatrix(W))
				panic("cocycle does not vanish on B_1")
			}
		}
	}
	if minWeight < 0 {
		minWeight = 0
	}
	if S.verbose {
		log.Printf("cosystole: %d\n", minWeight)
	}
	return minWeight
}

func (S *CosystoleSearch[T]) Cocycles() (cocycles []BinaryVector) {
	S.prepare()
	if S.verbose {
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
		if S.verbose {
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
			edgeStates = applyEdgeMask(S.triangles[level].Edges(), edgeStates, S.edgeMask)
			// 			if S.verbose {
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
// 				if S.verbose {
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
					cT := S.leafToVector(q, level).Matrix().Transpose()
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
		leaves = newLeaves
	}
	if S.verbose {
		log.Printf("leaves: %d\n", len(leaves))
	}
	for _, p := range leaves {
		cocycles = append(cocycles, S.leafToVector(p, len(S.triangles)-1))
	}
	return
}

func (S *CosystoleSearch[T]) dumpBranch(node *StateNode, level int) {
	p := node
	for level >= 0 {
		log.Printf("level %d triangle %v [%v]", level, S.triangles[level], p)
		p = p.parent
		level--
	}
}

func (S *CosystoleSearch[T]) leafToVector(leaf *StateNode, level int) BinaryVector {
// 	if S.verbose {
// 		log.Printf("--> leafToVector: %v level=%d", leaf, level)
// 	}
	v := NewBinaryVector(len(S.C.edgeBasis))
	// walk branch of the tree from leaf to root
	p := leaf
	for level >= 0 {
		t := S.triangles[level]
// 		if S.verbose {
// 			log.Printf("  level=%d p=%v t=%v", level, p, t)
// 		}
		edges := t.Edges()
		for i, e := range edges {
			// log.Printf("    edge %v: %v", i, e)
			if p.edgeStates[i] {
				if j, ok := S.C.EdgeIndex[e]; ok {
					// log.Printf("      edge is on; j=%d", j)
					v[j] = 1
				} else {
					panic(fmt.Sprintf("edge %v not in edge index", e))
				}
			}
		}
		p = p.parent
		level--
	}
// 	if S.verbose {
// 		log.Printf("<-- leafToVector: leaf=%v vector=%v\n", leaf, v)
// 	}
	return v
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

func applyEdgeMask[T any](edges [3]ZEdge[T], states [3]kEdgeState, edgeMask map[ZEdge[T]]struct{}) [3]kEdgeState {
	// apply edgeMask here to potentially force some edges to be off
	for i, e := range edges {
		_, ok := edgeMask[e]
		if !ok {
			states[i] = kEdgeOff
		}
	}
	return states
}

// xxx not tested yet; experimental
func (S *CosystoleSearch[T]) Prefilter(U BinaryMatrix) {
	S.prepare()
	log.Printf("Prefilter: %d triangles\n", len(S.triangles))
	// prepare list of triangle edge indices (in the edge basis)
	var triangleEdgeIndices [][3]int = make([][3]int, len(S.triangles))
	for i, t := range S.triangles {
		for j, e := range t.Edges() {
			if k, ok := S.C.EdgeIndex[e]; ok {
				triangleEdgeIndices[i][j] = k
			} else {
				panic(fmt.Sprintf("edge %v not in edge index", e))
			}
		}
	}
	// each column of U represents a cochain (a set of edges).
	minTris := math.MaxInt
	minTrisIndex := -1
	for j := 0; j < U.NumColumns(); j++ {
		log.Printf("column %d weight: %d", j, U.ColumnWeight(j))
		// determine the number of triangles that contain an edge on
		// which this column cochain is supported (i.e., has a 1).
		n := 0
		for _, inds := range triangleEdgeIndices {
			for _, k := range inds {
				if U.Get(k, j) == 1 {
					n++
					break
				}
			}
		}
		// keep track of the column with the fewest (distinct) triangles.
		log.Printf("column %d: %d triangles", j, n)
		if n < minTris {
			minTris = n
			minTrisIndex = j
		}

		// also find the max support index in this column
		m := -1
		for i := 0; i < U.NumRows(); i++ {
			if U.Get(i, j) == 1 {
				m = i
			}
		}
		log.Printf("column %d max support index: %d", j, m)
	}
	log.Printf("minTris=%d minTrisIndex=%d\n", minTris, minTrisIndex)

	S.edgeMask = make(map[ZEdge[T]]struct{})
	edgeBasis := S.C.EdgeBasis()
	m := 0
	for i := 0; i < U.NumRows(); i++ {
		if U.Get(i, minTrisIndex) == 1 {
			S.edgeMask[edgeBasis[i]] = struct{}{}
			m++
		}
	}
	log.Printf("Edges after filtering: %d", m)
}

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

