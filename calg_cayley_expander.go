package golsv

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// xxx rename CalGCayleyComplex?
// xxx genericise to use any group with comparable element type?
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
	verbose bool, modulus *F2Polynomial, quotient bool, observer CalGObserver) *CalGCayleyExpander {
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

func (E *CalGCayleyExpander) Complex() *ZComplex[ElementCalG] {
	E.sortEdgeBasis()
	triangleBasis := E.triangleBasis()
	if E.observer != nil {
		E.observer.Edges(E.edgeBasis)
		E.observer.Triangles(triangleBasis)
		E.observer.End()
	}
	// We don't want ZComplex to lexically sort the bases using the
	// ZVertex[T] interface, because we have already sorted them in
	// the order we want, namely, by distance to the origin.
	resortBases := false
	return NewZComplex(E.vertexBasis, E.edgeBasis, triangleBasis, resortBases, E.verbose)
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
	// log.Printf("xxx edge: %v -- %v", hRep, uRep)
	e := NewZEdge[ElementCalG](h, u)
	if _, ok := E.edgeSet[e]; !ok {
		E.edgeSet[e] = true
		E.edgeBasis = append(E.edgeBasis, NewZEdge[ElementCalG](h, u))
	}
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

func (E *CalGCayleyExpander) sortEdgeBasis() {
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

func (E *CalGCayleyExpander) triangleLessByVertexAttendance(s, t ZTriangle[ElementCalG]) bool {
	sx := E.triangleToSortedVertexIndices(s)
	tx := E.triangleToSortedVertexIndices(t)
	if sx[0] < tx[0] {
		return true
	} else if sx[0] > tx[0] {
		return false
	} else if sx[1] < tx[1] {
		return true
	} else if sx[1] > tx[1] {
		return false
	} else if sx[2] < tx[2] {
		return true
	} else {
		return false
	}
}

func (E *CalGCayleyExpander) triangleToSortedVertexIndices(t ZTriangle[ElementCalG]) [3]int {
	x := [3]int{}
	for i := 0; i < 3; i++ {
		if w, ok := E.attendance[t[i].(ElementCalG)]; ok {
			x[i] = w.id
		} else {
			panic(fmt.Sprintf("vertex %v not found", t[i]))
		}
	}
	return x
}

type triangleWorkerData struct {
	startIdx   int
	endIdx     int
	triangles  []ZTriangle[ElementCalG]
	currentIdx int
}

func (E *CalGCayleyExpander) triangleWorker(worker *triangleWorkerData, trianglesAtOrigin []ZTriangle[ElementCalG], edgeChecks bool, triangleSet map[ZTriangle[ElementCalG]]any, basis *[]ZTriangle[ElementCalG], mutex *sync.Mutex) {
	drainThreshold := 5000 + rand.Intn(5001) // 5000-10000
	
	drain := func() {
		if len(worker.triangles) == 0 {
			return
		}
		mutex.Lock()
		for _, triangle := range worker.triangles {
			if _, ok := triangleSet[triangle]; !ok {
				triangleSet[triangle] = nil
				*basis = append(*basis, triangle)
			}
		}
		mutex.Unlock()
		worker.triangles = worker.triangles[:0] // clear slice but keep capacity
	}
	
	for i := worker.startIdx; i < worker.endIdx; i++ {
		worker.currentIdx = i
		u := E.vertexBasis[i]
		localTriangles := E.trianglesAtVertex2(u, trianglesAtOrigin, edgeChecks)
		worker.triangles = append(worker.triangles, localTriangles...)
		
		if len(worker.triangles) >= drainThreshold {
			drain()
			drainThreshold = 5000 + rand.Intn(5001) // new random threshold
		}
	}
	drain()
}

func (E *CalGCayleyExpander) triangleBasis() []ZTriangle[ElementCalG] {
	if E.verbose {
		log.Printf("computing triangle basis")
	}
	edgeChecks := false
	if E.maxDepth > 0 {
		edgeChecks = true
	}
	trianglesAtOrigin := E.trianglesAtVertex(E.vertexBasis[0], edgeChecks)
	
	numWorkers := runtime.NumCPU()
	numVertices := len(E.vertexBasis)
	verticesPerWorker := numVertices / numWorkers
	
	workers := make([]triangleWorkerData, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i].startIdx = i * verticesPerWorker
		if i == numWorkers-1 {
			workers[i].endIdx = numVertices
		} else {
			workers[i].endIdx = (i + 1) * verticesPerWorker
		}
		workers[i].triangles = make([]ZTriangle[ElementCalG], 0)
		workers[i].currentIdx = workers[i].startIdx
	}
	
	if E.verbose {
		for i, worker := range workers {
			log.Printf("worker %d: vertices [%d:%d)", i, worker.startIdx, worker.endIdx)
		}
	}
	
	//
	//       f     g     h
	//    u --- v --- w --- x
	//
	basis := make([]ZTriangle[ElementCalG], 0)
	triangleSet := make(map[ZTriangle[ElementCalG]]any)
	var mutex sync.Mutex
	
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerIdx int) {
			defer wg.Done()
			E.triangleWorker(&workers[workerIdx], trianglesAtOrigin, edgeChecks, triangleSet, &basis, &mutex)
		}(i)
	}
	
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	startTime := time.Now()
	statInterval := 5 * time.Second
	ticker := time.NewTicker(statInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-done:
			goto workersComplete
		case <-ticker.C:
			if E.verbose {
				totalProgress := 0
				for _, worker := range workers {
					totalProgress += worker.currentIdx - worker.startIdx
				}
				now := time.Now()
				rate := float64(totalProgress) / now.Sub(startTime).Seconds()
				estimatedHoursRemaining := float64(numVertices-totalProgress) / rate / 3600
				msg := fmt.Sprintf("triangleBasis; progress=%d/%d rate=%1.3f ehr=%1.1f",
					totalProgress, numVertices, rate, estimatedHoursRemaining)
				log.Println(msg)
			}
		}
	}
	
workersComplete:
	if E.verbose {
		log.Printf("done computing triangle basis; found %d triangles", len(basis))
		log.Printf("sorting triangles")
	}
	sort.Slice(basis, func(i, j int) bool {
		return E.triangleLessByVertexAttendance(basis[i], basis[j])
	})
	if E.verbose {
		log.Printf("done sorting triangles")
	}
	return basis
}

func (E *CalGCayleyExpander) trianglesAtVertex(uVertex ZVertex[ElementCalG], edgeChecks bool) (triangles []ZTriangle[ElementCalG]) {
	u := uVertex.(ElementCalG)
	for _, f := range E.gens {
		v := NewElementCalGIdentity()
		v.Mul(u, f)
		if E.quotient {
			v = v.Modf(*E.modulus)
		}
		if edgeChecks {
			uv := NewZEdge[ElementCalG](u, v)
			if _, ok := E.edgeSet[uv]; !ok {
				continue
			}
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
			if edgeChecks {
				vw := NewZEdge[ElementCalG](v, w)
				if _, ok := E.edgeSet[vw]; !ok {
					continue
				}
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
					if edgeChecks {
						wu := NewZEdge[ElementCalG](w, u)
						if _, ok := E.edgeSet[wu]; !ok {
							continue
						}
					}
					triangle := NewZTriangle[ElementCalG](u, v, w)
					triangles = append(triangles, triangle)
				}
			}
		}
	}
	return triangles
}

func (E *CalGCayleyExpander) trianglesAtVertex2(uVertex ZVertex[ElementCalG], trianglesAtOrigin []ZTriangle[ElementCalG], edgeChecks bool) (triangles []ZTriangle[ElementCalG]) {
	u := uVertex.(ElementCalG)
	for _, t := range trianglesAtOrigin {
		a := t[0].(ElementCalG)
		b := t[1].(ElementCalG)
		c := t[2].(ElementCalG)
		var ua, ub, uc ElementCalG
		ua.Mul(u, a)
		ub.Mul(u, b)
		uc.Mul(u, c)
		if E.quotient {
			ua = ua.Modf(*E.modulus)
			ub = ub.Modf(*E.modulus)
			uc = uc.Modf(*E.modulus)
		}
		if edgeChecks {
			uaub := NewZEdge(ua, ub)
			if _, ok := E.edgeSet[uaub]; !ok {
				continue
			}
			uauc := NewZEdge(ua, uc)
			if _, ok := E.edgeSet[uauc]; !ok {
				continue
			}
			ubuc := NewZEdge(ub, uc)
			if _, ok := E.edgeSet[ubuc]; !ok {
				continue
			}
		}
		s := NewZTriangle[ElementCalG](ua, ub, uc)
		triangles = append(triangles, s)
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
