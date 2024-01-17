package golsv

import (
	"fmt"
	"log"
)

// xxx rename CalGCayleyComplex?
// xxx genericise to use any group with comparable element type?
type CalGCayleyExpander struct {
	gens       []ElementCalG
	maxDepth   int
	curDepth   int
	verbose    bool
	modulus    *F2Polynomial
	quotient   bool
	attendance map[ElementCalG]vertexWrapper
	todo       calGTodoQueue
	edgeSet    map[ZEdge[ElementCalG]]any
	congruenceSubgroupElements []ElementCalG
	observer   CalGObserver
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
	id  int
	// when a new vertex is created (i.e. we compute a new group
	// element that has not yet been seen, i.e. is not in the
	// attendance map), we store the generator that was used to
	// compute it (by index).  this is useful for computing inverses
	// of a group element by following the links back to the identity.
	generator int // index in E.gens
}

func NewCalGCayleyExpander(generators []ElementCalG, maxDepth int,
	verbose bool, modulus *F2Polynomial, quotient bool, observer CalGObserver) *CalGCayleyExpander {
	if quotient {
		k := modulus.MaxYFactor()
		if k > 0 {
			panic("f is divisible by y")
		}
	}
	return &CalGCayleyExpander{
		gens:       generators,
		maxDepth:   maxDepth,
		verbose:    verbose,
		modulus:    modulus,
		quotient:   quotient,
		attendance: make(map[ElementCalG]vertexWrapper),
		todo:       NewCalGTodoQueueSlice(),
		edgeSet:    make(map[ZEdge[ElementCalG]]any),
		observer:   observer,
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

func (E *CalGCayleyExpander) Expand() {
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
	// xxx experimental for debugging;
// 	checkCongruenceConsistency1 := true
// 	if checkCongruenceConsistency1 {
// 		log.Printf("checking for multiple congruence subgroup identities")
// 		for g, id := range E.attendance {
// 			if g.IsIdentityModf(*E.modulus) {
// 				log.Printf("found congruence subgroup identity: g=(%v) id=%d", g, id)
// 			}
// 		}
// 	}
}

func (E *CalGCayleyExpander) Complex(sortBases bool) *ZComplex[ElementCalG] {
	vertexBasis := E.vertexBasis()
	edgeBasis := E.edgeBasis()
	triangleBasis := E.triangleBasis(vertexBasis)
	if E.observer != nil {
		E.observer.Edges(edgeBasis)
		E.observer.Triangles(triangleBasis)
		E.observer.End()
	}
	return NewZComplex(vertexBasis, edgeBasis, triangleBasis, sortBases, E.verbose)
}

func (E *CalGCayleyExpander) initialVertex() (hId int, hRep ElementCalG) {
	identity := NewElementCalGIdentity()
	hId, _ = E.getOrSetVertex(identity, -1, 0)
	return hId, identity
}

func (E *CalGCayleyExpander) getOrSetVertex(u ElementCalG, genIndex int, uDepth int) (uId int, added bool) {
	// log.Printf("xxx getOrSetVertex; u=%v genIndex=%d", u, genIndex)

	wrapper, ok := E.attendance[u]
	if ok {
		return wrapper.id, false
	}
	uId = len(E.attendance)
	E.attendance[u] = vertexWrapper{
		id: uId,
		generator: genIndex,
	}
	// check for congruence subgroup element, unless we are doing a quotient
	if E.modulus != nil && !E.quotient && !u.IsIdentity() {
		c := u.Modf(*E.modulus)
		if c.IsIdentity() {
			E.congruenceSubgroupElements = append(E.congruenceSubgroupElements, u)
			log.Printf("depth=%d found=%d; subgroup element: %v", uDepth, len(E.congruenceSubgroupElements), u)
		}
	}
	if E.observer != nil {
		E.observer.Vertex(u, uId, uDepth)
	}
	return uId, true
}

func (E *CalGCayleyExpander) elementInverse(u ElementCalG) (ElementCalG, ZPath[ElementCalG]) {
	// walk the edges back to the identity, calculating the inverse as
	// a cumulative product of inverses of the generators for each
	// edge.
	// log.Printf("xxx computing inverse path for u=%v", u)
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
		// log.Printf("xxx inverse: i=%d v=%v gen=%d gInv=%v", i, v, wrapper.generator, gInv)
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
	// log.Printf("xxx inverse: path=%v", inverseEdges)
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
				break
			}
		}
		for i, g := range E.gens {
			u := NewElementCalGIdentity()
			u.Mul(task.hRep, g)
			if E.quotient {
				u = u.Modf(*E.modulus)
			}
			E.applyFilter(u, task.hRep, i, task.depth+1)
		}
		n++
		if n%(100*1000) == 0 {
			logStatus()
		}
	}
	logStatus()
}

// context: we are processing neighbor u of h via generator g.  u = h * g.
func (E *CalGCayleyExpander) applyFilter(u ElementCalG, h ElementCalG, genIndex int, uDepth int) {
	// xxx hack 
	// x := NewElementCalGFromString("(1,101,0)(1,0,101)(1,101,0)")
// 	x := NewElementCalGFromString("(11,0,1)(101,1,1)(101,1,0)")
// 	if u.Equal(x) {
// 		log.Printf("xxx getOrSetVertex; u=%v genIndex=%d h=%v", u, genIndex, h)
// 	}
	uId, added := E.getOrSetVertex(u, genIndex, uDepth)
	E.setEdge(h, u)
	if added {
		E.todo.Enqueue(&calGNeighborsTask{uId, u, uDepth})
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
	// log.Printf("xxx edge: %v -- %v", hRep, uRep)
	E.edgeSet[NewZEdge[ElementCalG](h, u)] = nil
}

func (E *CalGCayleyExpander) NumVertices() int {
	return len(E.attendance)
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

func (E *CalGCayleyExpander) vertexBasis() []ZVertex[ElementCalG] {
	basis := make([]ZVertex[ElementCalG], 0, len(E.attendance))
	for h, _ := range E.attendance {
		basis = append(basis, h)
	}
	return basis
}

func (E *CalGCayleyExpander) edgeBasis() []ZEdge[ElementCalG] {
	edges := make([]ZEdge[ElementCalG], 0, len(E.edgeSet))
	for edge := range E.edgeSet {
		edges = append(edges, edge)
	}
	return edges
}

// note that if the expansion was truncated, this method may return
// triangles referencing vertices that are not in the vertex basis.
func (E *CalGCayleyExpander) triangleBasis(vertexBasis []ZVertex[ElementCalG]) []ZTriangle[ElementCalG] {
	if E.verbose {
		log.Printf("computing triangle basis")
	}		
	//
	//       f     g     h
	//    u --- v --- w --- x
	//
	triangleSet := make(map[ZTriangle[ElementCalG]]any)
	for _, u := range vertexBasis {
		u := u.(ElementCalG)
		for _, f := range E.gens {
			v := NewElementCalGIdentity()
			v.Mul(u, f)
			if E.quotient {
				v = v.Modf(*E.modulus)
			}
			for _, g := range E.gens {
				w := NewElementCalGIdentity()
				w.Mul(v, g)
				if E.quotient {
					w = w.Modf(*E.modulus)
				}
				if w.Equal(u) {
					continue
				}
				for _, h := range E.gens {
					x := NewElementCalGIdentity()
					x.Mul(w, h)
					if E.quotient {
						x = x.Modf(*E.modulus)
					}
					if x.Equal(v) {
						continue
					} else if x.Equal(u) {
						triangleSet[NewZTriangle[ElementCalG](u, v, w)] = nil
					}
				}
			}
		}
	}
    triangles := make([]ZTriangle[ElementCalG], 0, len(triangleSet))
	for t := range triangleSet {
		triangles = append(triangles, t)
	}
	return triangles
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

