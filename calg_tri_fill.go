package golsv

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

type CalGTriangleFiller struct {
	edgeBasis   []ZEdge[ElementCalG]
	edgeChecks  bool
	gens        []ElementCalG
	graph       *ZComplex[ElementCalG]
	modulus     *F2Polynomial
	quotient    bool
	verbose     bool
	vertexBasis []ZVertex[ElementCalG]
}

func NewCalGTriangleFiller(vertexBasis []ZVertex[ElementCalG], edgeBasis []ZEdge[ElementCalG],
	generators []ElementCalG, verbose bool, modulus *F2Polynomial, quotient bool, edgeChecks bool) *CalGTriangleFiller {
	return &CalGTriangleFiller{
		edgeBasis: edgeBasis,
		edgeChecks: edgeChecks, // nb. only used for unit testing when we generate a Cayley graph only up to limited depth
		gens: generators,
		modulus: modulus,
		quotient: quotient,
		verbose: verbose,
		vertexBasis: vertexBasis,
	}
}

func (F *CalGTriangleFiller) Complex() *ZComplex[ElementCalG] {
	// We don't want ZComplex to lexically sort the bases using the
	// ZVertex[T] interface, because we have already sorted them in
	// the order we want, namely, by distance to the origin.
	resortBases := false
	F.graph = NewZComplex(F.vertexBasis, F.edgeBasis, nil, resortBases, F.verbose)
	triangleBasis := F.triangleBasis()
	return NewZComplex(F.vertexBasis, F.edgeBasis, triangleBasis, resortBases, F.verbose)
}


func (F *CalGTriangleFiller) triangleBasis() []ZTriangle[ElementCalG] {
	if F.verbose {
		log.Printf("computing triangle basis")
	}
	trianglesAtOrigin := F.trianglesAtVertex(F.vertexBasis[0])
	
	numWorkers := runtime.NumCPU() - 1
	if numWorkers < 1 {
		numWorkers = 1
	}
	numVertices := len(F.vertexBasis)
	verticesPerWorker := numVertices / numWorkers
	
	workers := make([]triangleWorkerData, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i].id = i
		workers[i].startIdx = i * verticesPerWorker
		if i == numWorkers-1 {
			workers[i].endIdx = numVertices
		} else {
			workers[i].endIdx = (i + 1) * verticesPerWorker
		}
		workers[i].triangles = make([]ZTriangle[ElementCalG], 0)
		workers[i].currentIdx = workers[i].startIdx
	}
	
	if F.verbose {
		for i, worker := range workers {
			log.Printf("worker %d: vertices [%d:%d)", i, worker.startIdx, worker.endIdx)
		}
	}
	basis := make([]ZTriangle[ElementCalG], 0)
	triangleSet := make(map[ZTriangle[ElementCalG]]any)
	var mutex sync.Mutex
	
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerIdx int) {
			defer wg.Done()
			F.triangleWorker(&workers[workerIdx], trianglesAtOrigin, triangleSet, &basis, &mutex)
		}(i)
	}
	
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	startTime := time.Now()
	statInterval := 10 * time.Second
	ticker := time.NewTicker(statInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-done:
			goto workersComplete
		case <-ticker.C:
			if F.verbose {
				totalProgress := 0
				for _, worker := range workers {
					totalProgress += worker.currentIdx - worker.startIdx
				}
				now := time.Now()
				rate := float64(totalProgress) / now.Sub(startTime).Seconds()
				estimatedHoursRemaining := float64(numVertices-totalProgress) / rate / 3600
				msg := fmt.Sprintf("triangleBasis; progress=%d/%d rate=%1.1f ehr=%1.1f",
					totalProgress, numVertices, rate, estimatedHoursRemaining)
				log.Println(msg)
			}
		}
	}
	
workersComplete:
	if F.verbose {
		log.Printf("done computing triangle basis; found %d triangles", len(basis))
		log.Printf("sorting triangles")
	}
	sort.Slice(basis, func(i, j int) bool {
		return F.triangleLessByVertexBasis(basis[i], basis[j])
	})
	if F.verbose {
		log.Printf("done sorting triangles")
	}
	return basis
}

func (F *CalGTriangleFiller) trianglesAtVertex(uVertex ZVertex[ElementCalG]) (triangles []ZTriangle[ElementCalG]) {
	edgeIndex := F.graph.EdgeIndex()
	//
	//       f     g     h
	//    u --- v --- w --- x
	//
	u := uVertex.(ElementCalG)
	for _, f := range F.gens {
		v := NewElementCalGIdentity()
		v.Mul(u, f)
		if F.quotient {
			v = v.Modf(*F.modulus)
		}
		if F.edgeChecks {
			uv := NewZEdge[ElementCalG](u, v)
			if _, ok := edgeIndex[uv]; !ok {
				continue
			}
		}
		for _, g := range F.gens {
			w := NewElementCalGIdentity()
			w.Mul(v, g)
			if F.quotient {
				w = w.Modf(*F.modulus)
			}
			if w.Equal(u) {
				continue
			}
			if F.edgeChecks {
				vw := NewZEdge[ElementCalG](v, w)
				if _, ok := edgeIndex[vw]; !ok {
					continue
				}
			}
			for _, h := range F.gens {
				x := NewElementCalGIdentity()
				x.Mul(w, h)
				if F.quotient {
					x = x.Modf(*F.modulus)
				}
				if x.Equal(v) {
					continue
				} else if x.Equal(u) {
					if F.edgeChecks {
						wu := NewZEdge[ElementCalG](w, u)
						if _, ok := edgeIndex[wu]; !ok {
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

func (F *CalGTriangleFiller) trianglesAtVertex2(uVertex ZVertex[ElementCalG], trianglesAtOrigin []ZTriangle[ElementCalG]) (triangles []ZTriangle[ElementCalG]) {
	u := uVertex.(ElementCalG)
	for _, t := range trianglesAtOrigin {
		a := t[0].(ElementCalG)
		b := t[1].(ElementCalG)
		c := t[2].(ElementCalG)
		var ua, ub, uc ElementCalG
		ua.Mul(u, a)
		ub.Mul(u, b)
		uc.Mul(u, c)
		if F.quotient {
			ua = ua.Modf(*F.modulus)
			ub = ub.Modf(*F.modulus)
			uc = uc.Modf(*F.modulus)
		}
		if F.edgeChecks {
			edgeIndex := F.graph.EdgeIndex()
			uaub := NewZEdge(ua, ub)
			if _, ok := edgeIndex[uaub]; !ok {
				continue
			}
			uauc := NewZEdge(ua, uc)
			if _, ok := edgeIndex[uauc]; !ok {
				continue
			}
			ubuc := NewZEdge(ub, uc)
			if _, ok := edgeIndex[ubuc]; !ok {
				continue
			}
		}
		s := NewZTriangle[ElementCalG](ua, ub, uc)
		triangles = append(triangles, s)
	}
	return triangles
}

type triangleWorkerData struct {
	id         int
	startIdx   int
	endIdx     int
	triangles  []ZTriangle[ElementCalG]
	currentIdx int
}

func (F *CalGTriangleFiller) triangleWorker(worker *triangleWorkerData, trianglesAtOrigin []ZTriangle[ElementCalG], triangleSet map[ZTriangle[ElementCalG]]any, basis *[]ZTriangle[ElementCalG], mutex *sync.Mutex) {
	drainThresholdBase := 10000
	drainThreshold := drainThresholdBase + rand.Intn(5001)
	
	drain := func() {
		if len(worker.triangles) == 0 {
			return
		}
		mutex.Lock()
		if F.verbose {
			log.Printf("worker %d synchronizing %d triangles", worker.id, len(worker.triangles))
		}
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
		u := F.vertexBasis[i]
		localTriangles := F.trianglesAtVertex2(u, trianglesAtOrigin)
		worker.triangles = append(worker.triangles, localTriangles...)
		
		if len(worker.triangles) >= drainThreshold {
			drain()
			drainThreshold = drainThresholdBase + rand.Intn(5001) // new random threshold
		}
	}
	drain()
}

func (F *CalGTriangleFiller) triangleLessByVertexBasis(s, t ZTriangle[ElementCalG]) bool {
	sx := F.triangleToSortedVertexIndices(s)
	tx := F.triangleToSortedVertexIndices(t)
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

func (F *CalGTriangleFiller) triangleToSortedVertexIndices(t ZTriangle[ElementCalG]) [3]int {
	vertexIndex := F.graph.VertexIndex()
	x := [3]int{}
	for i := 0; i < 3; i++ {
		if n, ok := vertexIndex[t[i].(ElementCalG)]; ok {
			x[i] = n
		} else {
			panic(fmt.Sprintf("vertex %v not found", t[i]))
		}
	}
	return x
}

