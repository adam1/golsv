package golsv

// xxx this file should be considered experimental and deprecated

// this file is similar to bf_cayley_expander.go except instead of
// expanding edges incident to vertices, it expands triangles incident
// to edges.
//
// xxx is this equivalent to generating/expanding the triangle-edge graph?
// in particular, is the triangle-edge graph a Cayley graph an induced
// set of generators?

import (
	"log"
)

type TriangleExpander struct {
	lsv *LsvContext
	initialEdges []Edge
	maxDepth int
    attendance map[Triangle]any
	queue *triangleQueue
	trianglesAtOrigin []Triangle
	verbose bool
	exhausted bool
}
	
func NewTriangleExpander(lsv *LsvContext, edges []Edge, maxDepth int, verbose bool) *TriangleExpander {
	return &TriangleExpander{
		lsv: lsv,
		initialEdges: edges,
		maxDepth: maxDepth,
		attendance: make(map[Triangle]any),
		queue: NewTriangleQueue(),
		trianglesAtOrigin: LsvTrianglesAtOrigin(lsv),
		verbose: verbose,
	}
}

// xxx test; test E.exhausted logic in particular
func (E *TriangleExpander) Expand() bool {
	if E.verbose {
		log.Printf("expanding triangles to depth %d", E.maxDepth)
		log.Printf("%d initial edges", len(E.initialEdges))
		log.Printf("triangles at origin: %d", len(E.trianglesAtOrigin))
	}
	E.exhausted = true
	E.computeInitialTriangles()
 	E.processNeighborsTasks()
	if E.verbose {
		log.Printf("done with triangle expansion; exhausted=%v", E.exhausted)
	}
	return E.exhausted
}

// xxx test
func (E *TriangleExpander) computeInitialTriangles() {
	for _, e := range E.initialEdges {
		// log.Printf("finding triangles on initial edge: %v", e)
		triangles := AllTrianglesContainingEdge(E.lsv, E.trianglesAtOrigin, e)
		for _, t := range triangles {
			E.applyFilter(t, 0)
		}
	}
}

func (E *TriangleExpander) processNeighborsTasks() {
	n := 0
	curDepth := 0
	logStatus := func() {
		if E.verbose {
			log.Printf("processed %d tasks; triangles=%d depth=%d todo=%d", n, len(E.attendance), curDepth, E.queue.Len())
		}
	}
	for {
		task := E.queue.Dequeue()
		if task == nil {
			break
		}
		if task.depth > curDepth {
			curDepth = task.depth
			if E.verbose {
				logStatus()
			}
		}
		neighbors := AllTrianglesSharingEdgeWithTriangle(E.lsv, E.trianglesAtOrigin, task.t)
		for _, u := range neighbors {
			E.applyFilter(u, task.depth)
		}
		n++
		if n%(100*1000) == 0 {
			logStatus()
		}
	}
	logStatus()
}

func (E *TriangleExpander) applyFilter(t Triangle, depth int) {
	// log.Printf("xxx applyFilter: %v", t)
	if _, ok := E.attendance[t]; ok {
		return
	}
	E.attendance[t] = nil
	if E.maxDepth >= 0 {
		if depth < E.maxDepth {
			E.queue.Enqueue(&triangleTask{t, depth+1})
		} else {
			// log.Printf("xxx suppressing task at depth %d", depth+1)
			E.exhausted = false
		}
	} else {
		E.queue.Enqueue(&triangleTask{t, depth+1})
	}
}

func (E *TriangleExpander) Triangles() []Triangle {
	triangles := make([]Triangle, 0, len(E.attendance))
	for t, _ := range E.attendance {
		triangles = append(triangles, t)
	}
	return triangles
}

type triangleTask struct {
	t     Triangle
	depth int
}

type triangleQueue struct {
	slice []*triangleTask
}

func NewTriangleQueue() *triangleQueue {
	return &triangleQueue{
		slice: make([]*triangleTask, 0),
	}
}

func (S *triangleQueue) Enqueue(task *triangleTask) {
	S.slice = append(S.slice, task)
}

func (S *triangleQueue) Dequeue() *triangleTask {
	if len(S.slice) == 0 {
		return nil
	}
	task := S.slice[0]
	S.slice = S.slice[1:]
	return task
}

func (S *triangleQueue) Len() int {
	return len(S.slice)
}
