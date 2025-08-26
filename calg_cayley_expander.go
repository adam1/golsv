package golsv

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type CalGCayleyExpander struct {
	gens                       []ElementCalG
	maxDepth                   int
	curDepth                   int
	verbose                    bool
	modulus                    *F2Polynomial
	quotient                   bool
	attendance                 map[ElementCalG]vertexWrapper
	vertexBasis                []ZVertex[ElementCalG]
	todo                       calGTodoQueue
	edgeSet                    map[ZEdge[ElementCalG]]any
	edgeBasis                  []ZEdge[ElementCalG]
	congruenceSubgroupElements []ElementCalG
	observer                   CalGObserver
	checkPSL                   bool
	pslGenDepth                int
	pslElements                []ElementCalG
}

type CalGObserver interface {
	BeginVertices()
	End()
	EndVertices()
	Vertex(u ElementCalG, uId int, uDepth int)
	Edges([]ZEdge[ElementCalG])
	Triangles([]ZTriangle[ElementCalG])
}

type vertexWrapper struct {
	id int
	// when a new vertex is created (i.e. we compute a new group
	// element that has not yet been seen, i.e. is not in the
	// attendance map), we store the generator that was used to
	// compute it (by index).  this is useful for computing inverses
	// of a group element by following the links back to the identity.
	generator int // index in E.gens
}

func NewCalGCayleyExpander(generators []ElementCalG, maxDepth int,
	verbose bool, modulus *F2Polynomial, quotient bool, observer CalGObserver, checkPSL bool) *CalGCayleyExpander {
	if quotient {
		k := modulus.MaxYFactor()
		if k > 0 {
			panic("f is divisible by y")
		}
	}
	return &CalGCayleyExpander{
		gens:        generators,
		maxDepth:    maxDepth,
		verbose:     verbose,
		modulus:     modulus,
		quotient:    quotient,
		attendance:  make(map[ElementCalG]vertexWrapper),
		vertexBasis: make([]ZVertex[ElementCalG], 0),
		todo:        NewCalGTodoQueueSlice(),
		edgeSet:     make(map[ZEdge[ElementCalG]]any),
		edgeBasis:   make([]ZEdge[ElementCalG], 0),
		observer:    observer,
		checkPSL:    checkPSL,
		pslElements: make([]ElementCalG, 0),
	}
}

// general nomenclature:
//
//     g: a generator
//     h: a group element
//     u = h * g: a group element that is a neighbor of h via generator g

type calGNeighborsTask struct {
	hId   int
	hRep  ElementCalG
	depth int
}

func (E *CalGCayleyExpander) Graph() *ZComplex[ElementCalG] {
	if E.verbose {
		log.Printf("expanding Cayley graph with %d generators; modulus=%v quotient=%v maxDepth=%d",
			len(E.gens), E.modulus, E.quotient, E.maxDepth)
	}
	if E.observer != nil {
		E.observer.BeginVertices()
	}
	identityId, identity := E.initialVertex()
	E.todo.Enqueue(&calGNeighborsTask{identityId, identity, 0})
	E.processNeighborsTasks()
	if E.verbose {
		log.Printf("done with Cayley expansion")

	}
	if E.observer != nil {
		E.observer.EndVertices()
	}
	E.sortEdgeBasis()
	// We don't want ZComplex to lexically sort the bases using the
	// ZVertex[T] interface, because we have already sorted them in
	// the order we want, namely, by distance to the origin.
	resortBases := false
	return NewZComplex(E.vertexBasis, E.edgeBasis, nil, resortBases, E.verbose)
}

func (E *CalGCayleyExpander) Complex() *ZComplex[ElementCalG] {
	edgeChecks := false
	if E.maxDepth > 0 {
		edgeChecks = true
	}
	F := NewCalGTriangleFiller(E.vertexBasis, E.edgeBasis, E.gens, E.verbose, E.modulus, E.quotient, edgeChecks)
	return F.Complex()
}

func (E *CalGCayleyExpander) initialVertex() (hId int, hRep ElementCalG) {
	identity := NewElementCalGIdentity()
	hId, _ = E.getOrSetVertex(identity, -1, 0)
	return hId, identity
}

func (E *CalGCayleyExpander) getOrSetVertex(u ElementCalG, genIndex int, uDepth int) (uId int, added bool) {
	wrapper, ok := E.attendance[u]
	if ok {
		return wrapper.id, false
	}
	uId = len(E.attendance)
	wrapper = vertexWrapper{
		id:        uId,
		generator: genIndex,
	}
	E.attendance[u] = wrapper
	E.vertexBasis = append(E.vertexBasis, u)

	// check for congruence subgroup element, unless we are doing a quotient
	if E.modulus != nil && !E.quotient && !u.IsIdentity() {
		c := u.Modf(*E.modulus)
		if c.IsIdentity() {
			E.congruenceSubgroupElements = append(E.congruenceSubgroupElements, u)
			log.Printf("depth=%d found=%d; subgroup element: %v", uDepth, len(E.congruenceSubgroupElements), u)
		}
	}
	if E.checkPSL && E.quotient && (E.pslGenDepth == 0 || uDepth <= E.pslGenDepth) && !u.IsIdentity() {
		if E.elementIsInPSL(u) {
			if E.pslGenDepth == 0 {
				E.pslGenDepth = uDepth
			}
			//log.Printf("element at depth=%d is in PSL: %v", uDepth, u)
			E.pslElements = append(E.pslElements, u)
		}
	}
	if E.observer != nil {
		E.observer.Vertex(u, uId, uDepth)
	}
	return uId, true
}

func (E *CalGCayleyExpander) elementGeneratorPath(u ElementCalG) []int {
	// walk backwards from u to the identity, collecting generator indices
	v := u.Dup()
	generatorIndices := make([]int, 0)
	
	for {
		wrapper, ok := E.attendance[v]
		if !ok {
			panic(fmt.Sprintf("elementGeneratorPath: not in attendance: v=%v", v))
		}
		if wrapper.generator < 0 {
			// reached identity (generator = -1)
			break
		}
		generatorIndices = append(generatorIndices, wrapper.generator)
		
		// compute previous vertex toward identity
		gInv := CartwrightStegerGeneratorsInverse(E.gens, wrapper.generator)
		tmp := NewElementCalGIdentity()
		tmp.Mul(v, gInv)
		if E.quotient {
			tmp = tmp.Modf(*E.modulus)
		}
		v.Copy(tmp)
	}
	
	// reverse the slice to get forward path from identity to u
	for i, j := 0, len(generatorIndices)-1; i < j; i, j = i+1, j-1 {
		generatorIndices[i], generatorIndices[j] = generatorIndices[j], generatorIndices[i]
	}
	
	return generatorIndices
}

func (E *CalGCayleyExpander) elementIsInPSL(u ElementCalG) bool {
	// get the generator path from identity to u
	genPath := E.elementGeneratorPath(u)
	
	// get matrix representations of generators
	genTable := CartwrightStegerGeneratorsWithMatrixReps(*E.modulus)
	
	// compute the matrix representation of u by multiplying generator matrices
	matRep := ProjMatF2Poly(MatF2PolyIdentity)
	for _, genIndex := range genPath {
		pairIndex := genIndex / 2
		if pairIndex >= len(genTable) {
			panic(fmt.Sprintf("generator pair index %d out of range", pairIndex))
		}
		
		var genMatRep ProjMatF2Poly
		if genIndex%2 == 0 {
			// even index: use B_uRep (generator)
			genMatRep = genTable[pairIndex].B_uRep
		} else {
			// odd index: use B_uInvRep (generator inverse)
			genMatRep = genTable[pairIndex].B_uInvRep
		}
		
		// multiply on the right: matRep = matRep * generator
		matRep = matRep.Mul(genMatRep)
	}
	
	// compute determinant and check if it's 1 mod f
	det := matRep.Determinant()
	return det.IsOneModf(*E.modulus)
}

func (E *CalGCayleyExpander) elementInverse(u ElementCalG) (ElementCalG, ZPath[ElementCalG]) {
	// walk the edges back to the identity, calculating the inverse as
	// a cumulative product of inverses of the generators for each
	// edge.
	v := u.Dup()
	p := NewElementCalGIdentity()
	tmp := NewElementCalGIdentity()
	i := 0
	inverseEdges := make([]ZEdge[ElementCalG], 0)
	for {
		wrapper, ok := E.attendance[v]
		if !ok {
			panic(fmt.Sprintf("inverse: not in attendance: v=%v", v))
		}
		if wrapper.generator < 0 {
			break
		}
		gInv := CartwrightStegerGeneratorsInverse(E.gens, wrapper.generator)
		tmp.Mul(p, gInv)
		p.Copy(tmp)
		// now p = p * gInv

		// compute next vertex toward identity
		tmp.Mul(v, gInv)
		if E.quotient {
			tmp = tmp.Modf(*E.modulus)
		}
		edge := NewZEdge[ElementCalG](v, tmp)
		inverseEdges = append(inverseEdges, edge)
		v.Copy(tmp)
		// now v = v * gInv
		i++
	}
	return p, inverseEdges
}

func (E *CalGCayleyExpander) processNeighborsTasks() {
	n := 0
	E.curDepth = 0
	logStatus := func() {
		if E.verbose {
			log.Printf("processed %d tasks; vertices=%d depth=%d todo=%d", n, len(E.attendance), E.curDepth, E.todo.Len())
		}
	}
	exitingForDepth := false
	for {
		task := E.todo.Dequeue()
		if task == nil {
			break
		}
		if task.depth > E.curDepth {
			E.curDepth = task.depth
			if E.verbose {
				logStatus()
			}
			if E.maxDepth > 0 && E.curDepth >= E.maxDepth {
				if E.verbose {
					log.Printf("exiting after depth %d", E.maxDepth)
				}
				exitingForDepth = true
			}
		}
		for i, g := range E.gens {
			u := NewElementCalGIdentity()
			u.Mul(task.hRep, g)
			if E.quotient {
				u = u.Modf(*E.modulus)
			}
			E.applyFilter(u, task.hRep, i, task.depth+1, exitingForDepth)
		}
		n++
		if n%(100*1000) == 0 {
			logStatus()
		}
		if exitingForDepth {
			break
		}
	}
	logStatus()
}

// context: we are processing neighbor u of h via generator g.  u = h * g.
func (E *CalGCayleyExpander) applyFilter(u ElementCalG, h ElementCalG, genIndex int, uDepth int, exitingForDepth bool) {
	if exitingForDepth {
		// add edge h-u only if u is already known
		_, ok := E.attendance[u]
		if ok {
			E.setEdge(h, u)
		}
	} else {
		uId, added := E.getOrSetVertex(u, genIndex, uDepth)
		E.setEdge(h, u)
		if added {
			E.todo.Enqueue(&calGNeighborsTask{uId, u, uDepth})
		}
	}
}

// xxx test
// Note: may return duplicated edges
func (E *CalGCayleyExpander) Project(path ZPath[ElementCalG]) ZPath[ElementCalG] {
	result := make(ZPath[ElementCalG], len(path))
	for i, edge := range path {
		a := edge[0].(ElementCalG).Modf(*E.modulus)
		b := edge[1].(ElementCalG).Modf(*E.modulus)
		f := NewZEdge[ElementCalG](a, b)
		result[i] = f
	}
	return result
}

func (E *CalGCayleyExpander) setEdge(h, u ElementCalG) {
	e := NewZEdge[ElementCalG](h, u)
	if _, ok := E.edgeSet[e]; !ok {
		E.edgeSet[e] = true
		E.edgeBasis = append(E.edgeBasis, NewZEdge[ElementCalG](h, u))
	}
}

func (E *CalGCayleyExpander) NumVertices() int {
	return len(E.attendance)
}

func (E *CalGCayleyExpander) PslGenerators() []ElementCalG {
	// experimental mode #1 - there is a 1:3 (14:42) ratio between
	// generators of G (isomorphic to PGL) and elements of subgroup H
	// of G (isomorphic to PSL). we find these at depth 2. we expect
	// that for each generator s of G (depth 1), there are three
	// elements a, b, c of H at depth 2 adjacent to s.  (note that a,
	// b, c could and probably are adjacent to other generators s_j at
	// depth 1.)  the experiment is to take the first such element a
	// found for each generator s.  so we produce a list of 14
	// elements in H.  do they generate H?
	doExperiment1 := false
	if doExperiment1 {
		log.Printf("experiment 1: taking the first PSL element found at depth 2 for each generator s")
		result := make([]ElementCalG, 0)
		for i, s := range E.gens {
			found := false
			var h ElementCalG
			for _, t := range E.gens {
				prod := NewElementCalGIdentity()
				prod.Mul(s, t)
				if E.quotient {
					prod = prod.Modf(*E.modulus)
				}
				for _, p := range E.pslElements {
					if prod == p {
						found = true
						h = p
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				panic(fmt.Sprintf("did not find psl element for generator %d", i))
			}
			log.Printf("i=%d: %v", i, h)
			result = append(result, h)
		}
		return result
	}

	// experiment2: iterate over genTable since its paired by inverses.
	// attempt to form a symmetric set by taking each element and its inverse
	doExperiment2 := false
	if doExperiment2 {
		log.Printf("experiment 2: taking the first PSL element pair (with inverse) found at depth 2 for each generator s with inverse")
		result := make([]ElementCalG, 0)
		resultMap := make(map[ElementCalG]any)
		genTable := CartwrightStegerGeneratorsWithMatrixReps(*E.modulus)
		for i, pairA := range genTable {
			found := false
			var h, hInv ElementCalG
			for _, pairB := range genTable {
				// xxx we may need to check AB, ABInv, AInvB, AInvBInv?
				prod := NewElementCalGIdentity()
				prod.Mul(pairA.B_u, pairB.B_u)
				if E.quotient {
					prod = prod.Modf(*E.modulus)
				}
				for _, p := range E.pslElements {
					_, seen := resultMap[p]
					if prod == p && !p.IsIdentity() && !seen {
						found = true
						h = p
						hInv = NewElementCalGIdentity()
						hInv.Mul(pairB.B_uInv, pairA.B_uInv)
						if E.quotient {
							hInv = hInv.Modf(*E.modulus)
						}
						break
					}
				}
				if found {
					break
				}
				// xxx for the second generator in the word, we may need to test both B_u and B_uInv
				prod.Mul(pairA.B_u, pairB.B_uInv)
				if E.quotient {
					prod = prod.Modf(*E.modulus)
				}
				for _, p := range E.pslElements {
					_, seen := resultMap[p]
					if prod == p && !p.IsIdentity() && !seen {
						found = true
						h = p
						hInv = NewElementCalGIdentity()
						hInv.Mul(pairB.B_u, pairA.B_uInv)
						if E.quotient {
							hInv = hInv.Modf(*E.modulus)
						}
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				panic(fmt.Sprintf("did not find psl elements for generator pair %d", i))
			}
			resultMap[h] = nil
			resultMap[hInv] = nil
			result = append(result, h, hInv)
		}
		return result
	}

	return E.pslElements
}

// xxx test?
func (E *CalGCayleyExpander) SystolicCandidateLifts() []ZPath[ElementCalG] {
	paths := make([]ZPath[ElementCalG], 0)
	for _, g := range E.congruenceSubgroupElements {
		_, inversePath := E.elementInverse(g)
		paths = append(paths, inversePath)
	}
	return paths
}

func (E *CalGCayleyExpander) sortEdgeBasis() {
	if E.verbose {
		log.Printf("sorting basis of %d edges", len(E.edgeBasis))
	}
	sort.Slice(E.edgeBasis, func(i, j int) bool {
		return E.edgeLessByVertexAttendance(E.edgeBasis[i], E.edgeBasis[j])
	})
}

func (E *CalGCayleyExpander) edgeLessByVertexAttendance(e, f ZEdge[ElementCalG]) bool {
	ex := E.edgeToSortedVertexIndices(e)
	fx := E.edgeToSortedVertexIndices(f)
	if ex[0] < fx[0] {
		return true
	} else if ex[0] > fx[0] {
		return false
	} else if ex[1] < fx[1] {
		return true
	} else {
		return false
	}
}

func (E *CalGCayleyExpander) edgeToSortedVertexIndices(e ZEdge[ElementCalG]) [2]int {
	x := [2]int{}
	for i := 0; i < 2; i++ {
		if w, ok := E.attendance[e[i].(ElementCalG)]; ok {
			x[i] = w.id
		} else {
			panic(fmt.Sprintf("vertex %v not found", e[i]))
		}
	}
	return x
}

type calGTodoQueue interface {
	Enqueue(*calGNeighborsTask)
	Dequeue() *calGNeighborsTask
	Len() int
}

type calGTodoQueueSlice struct {
	slice []*calGNeighborsTask
}

func NewCalGTodoQueueSlice() *calGTodoQueueSlice {
	return &calGTodoQueueSlice{
		slice: make([]*calGNeighborsTask, 0),
	}
}

func (S *calGTodoQueueSlice) Enqueue(task *calGNeighborsTask) {
	S.slice = append(S.slice, task)
}

func (S *calGTodoQueueSlice) Dequeue() *calGNeighborsTask {
	if len(S.slice) == 0 {
		return nil
	}
	task := S.slice[0]
	S.slice = S.slice[1:]
	return task
}

func (S *calGTodoQueueSlice) Len() int {
	return len(S.slice)
}

// xxx awkward place for this
func NewZComplexElementCalGFromBasisFiles(vertexBasisFile, edgeBasisFile, triangleBasisFile string, verbose bool) *ZComplex[ElementCalG] {
	if verbose {
		log.Printf("reading vertex basis file %s", vertexBasisFile)
	}
	vertexBasis := ReadElementCalGVertexFile(vertexBasisFile)
	if verbose {
		log.Printf("reading edge basis file %s", edgeBasisFile)
	}
	edgeBasis := ReadElementCalGEdgeFile(edgeBasisFile)
	if verbose {
		log.Printf("reading triangle basis file %s", triangleBasisFile)
	}
	var triangleBasis []ZTriangle[ElementCalG]
	if triangleBasisFile != "" {
		triangleBasis = ReadElementCalGTriangleFile(triangleBasisFile)
	}
	sortBases := false
	return NewZComplex(vertexBasis, edgeBasis, triangleBasis, sortBases, verbose)
}

func ReadElementCalGVertexFile(filename string) []ZVertex[ElementCalG] {
	return ReadStringFile(filename, NewVertexElementCalGFromString)
}

func ReadElementCalGEdgeFile(filename string) []ZEdge[ElementCalG] {
	return ReadStringFile(filename, NewEdgeElementCalGFromString)
}

func ReadElementCalGTriangleFile(filename string) []ZTriangle[ElementCalG] {
	return ReadStringFile(filename, NewTriangleElementCalGFromString)
}

// xxx test
func NewVertexElementCalGFromString(s string) ZVertex[ElementCalG] {
	return NewElementCalGFromString(s)
}

func NewEdgeElementCalGFromString(s string) ZEdge[ElementCalG] {
	parts := strings.Split(strings.Trim(s, " []"), " ")
	if len(parts) != 2 {
		panic(fmt.Sprintf("edge string has %d parts; expected 2", len(parts)))
	}
	return NewZEdge[ElementCalG](NewElementCalGFromString(parts[0]), NewElementCalGFromString(parts[1]))
}

// xxx test
func NewTriangleElementCalGFromString(s string) ZTriangle[ElementCalG] {
	parts := strings.Split(strings.Trim(s, " []"), " ")
	if len(parts) != 3 {
		panic(fmt.Sprintf("triangle string has %d parts; expected 3", len(parts)))
	}
	return NewZTriangle[ElementCalG](NewElementCalGFromString(parts[0]), NewElementCalGFromString(parts[1]), NewElementCalGFromString(parts[2]))
}
