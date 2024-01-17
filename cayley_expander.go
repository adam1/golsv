package golsv

import (
	"container/list"
	"log"
)

// xxx rename CayleyComplex?
type CayleyExpander struct {
	lsv        *LsvContext
	gens       []MatGF
	maxDepth   int
	verbose    bool
	attendance map[MatGF]int // map to vertex id
	todo       todoQueue
	edgeSet    map[Edge]any
}

func NewCayleyExpander(lsv *LsvContext, generators []MatGF, maxDepth int, verbose bool ) *CayleyExpander {
	return &CayleyExpander{
		lsv:        lsv,
		gens:       generators,
		maxDepth:   maxDepth,
		verbose:    verbose,
		attendance: make(map[MatGF]int),
		todo:       NewTodoQueueSlice(),
		edgeSet:    make(map[Edge]any),
	}
}

// general nomenclature:
//
//     g: a generator
//     h: a group element
//     u = g * h: a group element that is a neighbor of h via generator g

type neighborsTask struct {
	hId   int
	hRep  *MatGF
	depth int
}

func (E *CayleyExpander) Expand() {
	if E.verbose {
		log.Printf("expanding Cayley graph with %d generators", len(E.gens))
	}
	identityId, identity := E.initialVertex()
	E.todo.Enqueue(&neighborsTask{identityId, identity, 0})
	E.processNeighborsTasks()
	if E.verbose {
		log.Printf("done with Cayley expansion")

	}
}

func (E *CayleyExpander) Complex(sortBases bool) *Complex {
	vertexBasis := E.vertexBasis()
	return NewComplex(vertexBasis, E.edgeBasis(), E.triangleBasis(vertexBasis), sortBases, E.verbose)
}

func (E *CayleyExpander) initialVertex() (hId int, hRep *MatGF) {
	hId, _ = E.ensureGroupElementKnown(MatGfIdentity)
	return hId, MatGfIdentity
}

func (E *CayleyExpander) ensureGroupElementKnown(uRep *MatGF) (uId int, added bool) {
	uId, added = E.getOrSetVertex(uRep)
	return
}

func (E *CayleyExpander) getOrSetVertex(hRep *MatGF) (hId int, added bool) {
	hId, ok := E.attendance[*hRep]
	if ok {
		return hId, false
	}
	hId = len(E.attendance)
	E.attendance[*hRep] = hId
	return hId, true
}

func (E *CayleyExpander) processNeighborsTasks() {
	n := 0
	curDepth := 0
	logStatus := func() {
		if E.verbose {
			log.Printf("processed %d tasks; vertices=%d depth=%d todo=%d", n, len(E.attendance), curDepth, E.todo.Len())
		}
	}
	for {
		task := E.todo.Dequeue()
		if task == nil {
			break
		}
		if task.depth > curDepth {
			curDepth = task.depth
			if E.verbose {
				logStatus()
			}
			if E.maxDepth > 0 && curDepth >= E.maxDepth {
				if E.verbose {
					log.Printf("exiting after depth %d", E.maxDepth)
				}
				break
			}
		}
		// xxx save and recycle task struct?
		for _, g := range E.gens {
			u := task.hRep.Multiply(E.lsv, &g)
			u.MakeCanonical(E.lsv)
			E.applyFilter(task.hRep, task.hId, u, task.depth+1)
		}
		n++
		if n%(100*1000) == 0 {
			logStatus()
		}
	}
	logStatus()
}

func (E *CayleyExpander) applyFilter(hRep *MatGF, hId int, uRep *MatGF, uDepth int) {
	uId, added := E.ensureGroupElementKnown(uRep)
	E.setEdge(hRep, uRep)
	if added {
		E.todo.Enqueue(&neighborsTask{uId, uRep, uDepth})
	}
}

func (E *CayleyExpander) setEdge(hRep, uRep *MatGF) {
	E.edgeSet[NewEdge(*hRep, *uRep)] = nil
}

func (E *CayleyExpander) NumVertices() int {
	return len(E.attendance)
}

func (E *CayleyExpander) vertexBasis() []Vertex {
	basis := make([]Vertex, 0, len(E.attendance))
	for hRep, _ := range E.attendance {
		basis = append(basis, hRep)
	}
	return basis
}

func (E *CayleyExpander) edgeBasis() []Edge {
	edges := make([]Edge, 0, len(E.edgeSet))
	for edge := range E.edgeSet {
		edges = append(edges, edge)
	}
	return edges
}

func (E *CayleyExpander) triangleBasis(vertexBasis []Vertex) []Triangle {
	if E.verbose {
		log.Printf("computing triangle basis")
	}		
	//
	//       f     g     h
	//    u --- v --- w --- x
	//
	triangleSet := make(map[Triangle]any)
	for _, u := range vertexBasis {
		for _, f := range E.gens {
			v := u.Multiply(E.lsv, &f)
			for _, g := range E.gens {
				w := v.Multiply(E.lsv, &g)
				if w.Equal(E.lsv, &u) {
					continue
				}
				for _, h := range E.gens {
					x := w.Multiply(E.lsv, &h)
					if x.Equal(E.lsv, v) {
						continue
					} else if x.Equal(E.lsv, &u) {
						triangleSet[NewTriangle(u, *v, *w)] = nil
					}
				}
			}
		}
	}
    triangles := make([]Triangle, 0, len(triangleSet))
	for t := range triangleSet {
		triangles = append(triangles, t)
	}
	return triangles
}

type todoQueue interface {
	Enqueue(*neighborsTask)
	Dequeue() *neighborsTask
	Len() int
}

type todoQueueList struct {
	list *list.List
}

func NewTodoQueueList() *todoQueueList {
	return &todoQueueList{
		list: list.New(),
	}
}

func (L *todoQueueList) Enqueue(task *neighborsTask) {
	L.list.PushBack(task)
}

func (L *todoQueueList) Dequeue() *neighborsTask {
	item := L.list.Front()
	if item == nil {
		return nil
	}
	L.list.Remove(item)
	return item.Value.(*neighborsTask)
}

func (L *todoQueueList) Len() int {
	return L.list.Len()
}

type todoQueueSlice struct {
	slice []*neighborsTask
}

func NewTodoQueueSlice() *todoQueueSlice {
	return &todoQueueSlice{
		slice: make([]*neighborsTask, 0),
	}
}

func (S *todoQueueSlice) Enqueue(task *neighborsTask) {
	S.slice = append(S.slice, task)
}

func (S *todoQueueSlice) Dequeue() *neighborsTask {
	if len(S.slice) == 0 {
		return nil
	}
	task := S.slice[0]
	S.slice = S.slice[1:]
	return task
}

func (S *todoQueueSlice) Len() int {
	return len(S.slice)
}

