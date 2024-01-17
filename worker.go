package golsv

import (
	"log"
	"runtime"
	"sync"
)

type WorkGroup struct {
	numWorkers int
	waitGroup sync.WaitGroup
	workers []*Worker
	verbose bool
}

type Work interface {
	Do()
}

func NewWorkGroup(verbose bool) *WorkGroup {
	G := &WorkGroup{
		numWorkers: runtime.NumCPU(),
		verbose: verbose,
	}
	G.workers = make([]*Worker, G.numWorkers)
	if G.verbose {
		log.Printf("spawning %d workers", G.numWorkers)
	}
	for i := 0; i < G.numWorkers; i++ {
		G.workers[i] = NewWorker(i, &G.waitGroup)
	}
	return G
}

func (G *WorkGroup) ProcessBatch(work []Work) {
	items := len(work)
	itemsPerWorker := items / len(G.workers)
	if itemsPerWorker == 0 {
		itemsPerWorker = 1
	}
	G.sendWork(work, itemsPerWorker)
	G.waitGroup.Wait()
}

func (G *WorkGroup) sendWork(work []Work, itemsPerWorker int) {
	k := len(work)
	n := len(G.workers)
	for w := 0; w < n; w++ {
		start := w * itemsPerWorker
		if start >= k {
			break
		}
		end := start + itemsPerWorker
		if w == n - 1 {
			end = k
		}
		// log.Printf("worker %d: %d-%d", w, start, end)
		G.workers[w].setWork(work[start:end])
	}
}

// func (G *WorkGroup) wait() {
// 	// xxx is G.started consumed before we're done queueing up the work?
// 	log.Printf("waiting for %d workers", G.started)
// 	for w := 0; w < G.started; w++ {
// 		G.waitGroup.Wait()
// 	}
// }

type Worker struct {
	index int
	waitGroup *sync.WaitGroup
	todo chan any
	work []Work
}

// xxx per docs waitgroup must not be copied after first use
func NewWorker(index int, group *sync.WaitGroup) *Worker {
	W := &Worker{
		index: index,
		waitGroup: group,
		todo: make(chan any),
		work: nil,
	}
	go W.run()
	return W
}

func (W *Worker) run() {
	for {
		<- W.todo
		W.doWork()
	}
}

func (W *Worker) doWork() {
	did := 0
	for _, work := range W.work {
		work.Do()
		did++
	}
	// log.Printf("worker %d: done; did %d", W.index, did)
	W.waitGroup.Done()
}

func (W *Worker) setWork(work []Work) {
	// log.Printf("xxx worker %d: set work %v", W.index, work)
	// xxx have to set all the work before we start the worker and begin waiting??
	W.work = work
	W.waitGroup.Add(1)
	W.todo <- nil
}
