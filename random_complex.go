package golsv

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	mathrand "math/rand"
)	

type RandomComplexGenerator struct {
	dimC_0 int
	dimC_1 int
	dimC_2 int
	dimZ_1 int
	dimH_1 int
	dimB_1 int
	verbose bool
}

func NewRandomComplexGenerator(dimC_0 int, verbose bool) *RandomComplexGenerator {
	return &RandomComplexGenerator{
		dimC_0,
		-1,
		-1,
		-1,
		-1,
		-1,
		verbose,
	}
}

func (R *RandomComplexGenerator) RandomComplex() (d_1, d_2 BinaryMatrix, err error) {
	R.dimC_1, err = R.randomizeDimC_1(false)
	if err != nil {
		return nil, nil, err
	}
	R.dimC_2, err = R.randomizeDimC_2()
	if err != nil {
		return nil, nil, err
	}
	var kernelMatrix BinaryMatrix
	d_1, kernelMatrix = R.randomizeGeneral_d_1()
	if R.verbose {
		log.Printf("generated d_1: %v\n", d_1)
	}
	R.dimZ_1 = kernelMatrix.NumColumns()
	R.dimB_1, err = randomizeDimB_1(R.dimZ_1)
	if err != nil {
		return nil, nil, err
	}
	R.dimH_1 = R.dimZ_1 - R.dimB_1

// 	log.Printf("xxx dimC_0=%d dimC_1=%d dimC_2=%d dimZ_1=%d dimB_1=%d dimH_1=%d",
// 		R.dimC_0, R.dimC_1, R.dimC_2, R.dimZ_1, R.dimB_1, R.dimH_1)

	// shuffle the columns of kernelMatrix and truncate to dimB_1
	// columns
	if R.verbose {
		log.Printf("shuffling columns of kernel matrix")
	}
	shuffled := kernelMatrix.Sparse()
	shuffled.ShuffleColumns()
	kernelMatrix = shuffled.DenseSubmatrix(0, shuffled.NumRows(), 0, R.dimB_1)

	d_2 = R.randomizeGeneral_d_2(kernelMatrix)
	return d_1, d_2, nil
}

func (R *RandomComplexGenerator) RandomSimplicialComplex() (d_1, d_2 BinaryMatrix, err error) {
	R.dimC_1, err = R.randomizeDimC_1(true)
	if err != nil {
		return nil, nil, err
	}
	R.dimC_2, err = R.randomizeDimC_2()
	if err != nil {
		return nil, nil, err
	}
	// remove duplicate columns since we require there be at most one
	// edge between any two vertices.
	d_1 = NewRandomSparseBinaryMatrixWithColumnWeight(R.dimC_0, R.dimC_1, 2)
	d_1 = removeDuplicateColumns(d_1)

	// enumerate 3-cliques and randomly fill them in
	tmp_d_2 := NewSparseBinaryMatrix(0, 0)
	graph := NewZComplexFromBoundaryMatrices(d_1, tmp_d_2)

	triangleFillDensity := 0.5
	triangleBasis := make([]ZTriangle[ZVertexInt], 0)
	graph.BFWalk3Cliques(func(c [3]ZVertex[ZVertexInt]) {
		if randFloat() >= triangleFillDensity {
			t := NewZTriangle(c[0], c[1], c[2])
			triangleBasis = append(triangleBasis, t)
		}
	})
	verbose := false
	C := NewZComplex(graph.VertexBasis(), graph.EdgeBasis(), triangleBasis, true, verbose)
	return C.D1(), C.D2(), nil
}

func (R *RandomComplexGenerator) RandomCliqueComplex(probEdge float64) (d_1, d_2 BinaryMatrix, err error) {
	numVertices := R.dimC_0
	if R.verbose {
		log.Printf("Generating clique complex over %d vertices with edge probability %v", numVertices, probEdge)
	}
	if probEdge < 0 || probEdge > 1 {
		panic("pEdge must be between 0 and 1")
	}
	entropyBytesPerBit := 2 // nb. increase if more precision is needed in probEdge
	entropy := make([]byte, numVertices * entropyBytesPerBit)
	bytes := make([]byte, 8)
	maxInt := uint64(1)<<(entropyBytesPerBit * 8) - 1
	cutoff := uint64(probEdge * float64(maxInt))
	d_1Sparse := NewSparseBinaryMatrix(numVertices, 0).Sparse()
	numEdges := 0
	for i := 0; i < numVertices; i++ {
		numCols := numVertices - i - 1
		entropy = entropy[:numCols * entropyBytesPerBit]
		_, err := rand.Read(entropy)
		if err != nil {
			panic(err)
		}
		for k := 0; k < len(entropy); k += entropyBytesPerBit {
			copy(bytes, entropy[k:k+entropyBytesPerBit])
			num := binary.LittleEndian.Uint64(bytes)
			if num <= cutoff {
				M := NewSparseBinaryMatrix(numVertices, 1)
				M.Set(i, 0, 1)
				M.Set(i+k/entropyBytesPerBit+1, 0, 1)
				d_1Sparse.AppendColumn(M)
				numEdges++
			}
		}
	}
	C := NewZComplexFromBoundaryMatrices(d_1Sparse, NewSparseBinaryMatrix(numEdges, 0))
	if R.verbose {
		log.Printf("Generated d_1: %v\n", d_1Sparse)
		log.Printf("Filling cliques")
	}
	C.Fill3Cliques()
	return C.D1(), C.D2(), nil
}

func (R *RandomComplexGenerator) RandomRegularCliqueComplex(k int) (d_1, d_2 BinaryMatrix, err error) {
	return R.RandomRegularCliqueComplexWithRetries(k, 100)
}

func (R *RandomComplexGenerator) RandomRegularCliqueComplexWithRetries(k, maxRetries int) (d_1, d_2 BinaryMatrix, err error) {
	numVertices := R.dimC_0
	if R.verbose {
		log.Printf("Generating regular clique complex over %d vertices with regularity degree %d", numVertices, k)
	}
	if k < 0 || k >= numVertices {
		return nil, nil, fmt.Errorf("regularity degree %d must be between 0 and %d", k, numVertices-1)
	}
	if numVertices*k%2 != 0 {
		return nil, nil, fmt.Errorf("cannot create %d-regular graph on %d vertices (n*k must be even)", k, numVertices)
	}
	
	// Use configuration model: create k stubs per vertex, then randomly pair them
	stubs := make([]int, 0, numVertices*k)
	for v := 0; v < numVertices; v++ {
		for i := 0; i < k; i++ {
			stubs = append(stubs, v)
		}
	}
	
	mathrand.Shuffle(len(stubs), func(i, j int) {
		stubs[i], stubs[j] = stubs[j], stubs[i]
	})
	
	// Track adjacency to avoid multiple edges between same vertices
	adjacency := make([]map[int]bool, numVertices)
	for i := 0; i < numVertices; i++ {
		adjacency[i] = make(map[int]bool)
	}
	
	d_1Sparse := NewSparseBinaryMatrix(numVertices, 0).Sparse()
	
	// Pair up stubs to create edges
	for retry := 0; retry < maxRetries; retry++ {
		edges := make([][2]int, 0)
		adjacencyCopy := make([]map[int]bool, numVertices)
		for i := 0; i < numVertices; i++ {
			adjacencyCopy[i] = make(map[int]bool)
		}
		
		success := true
		for i := 0; i < len(stubs); i += 2 {
			v, u := stubs[i], stubs[i+1]
			
			// Skip self-loops and multiple edges
			if v == u || adjacencyCopy[v][u] {
				success = false
				break
			}
			
			adjacencyCopy[v][u] = true
			adjacencyCopy[u][v] = true
			if v > u {
				v, u = u, v
			}
			edges = append(edges, [2]int{v, u})
		}
		
		if success {
			// Create boundary matrix from successful edge list
			for _, edge := range edges {
				v, u := edge[0], edge[1]
				M := NewSparseBinaryMatrix(numVertices, 1)
				M.Set(v, 0, 1)
				M.Set(u, 0, 1)
				d_1Sparse.AppendColumn(M)
			}
			break
		}
		
		mathrand.Shuffle(len(stubs), func(i, j int) {
			stubs[i], stubs[j] = stubs[j], stubs[i]
		})
		
		if retry == maxRetries-1 {
			return nil, nil, fmt.Errorf("failed to generate %d-regular graph after %d attempts", k, maxRetries)
		}
	}
	
	numEdges := d_1Sparse.NumColumns()
	C := NewZComplexFromBoundaryMatrices(d_1Sparse, NewSparseBinaryMatrix(numEdges, 0))
	if R.verbose {
		log.Printf("Generated regular graph with %d edges", numEdges)
		log.Printf("Generated d_1: %v\n", d_1Sparse)
		log.Printf("Filling cliques")
	}
	C.Fill3Cliques()
	return C.D1(), C.D2(), nil
}

func (R *RandomComplexGenerator) randomizeGeneral_d_1() (d_1 BinaryMatrix, kernelMatrix BinaryMatrix) {
	density := 0.1
	for true {
		d_1 = NewRandomDenseBinaryMatrixWithDensity(R.dimC_0, R.dimC_1, density)
		verbose := false
		reducer := NewDiagonalReducer(verbose)
		// the copy here prevents the original d_1 from being modified
		// which causes a bug for some reason...
		rowOpWriter, colOpWriter := NewOperationSliceWriter(), NewOperationSliceWriter()
		D := reducer.Reduce(d_1.Copy(), rowOpWriter, colOpWriter)

		Dsparse := D.Sparse()
		smithNormal, _ := Dsparse.IsSmithNormalForm()
		if !smithNormal {
			panic(fmt.Errorf("D is not in Smith normal form"))
		}
		kernelMatrix := reducer.computeKernelBasis(colOpWriter.Slice())
		dimZ_1 := kernelMatrix.NumColumns()
		if dimZ_1 > 0 {
			return d_1, kernelMatrix
		}
		if verbose {
			log.Printf("dimZ_1=0, trying again")
		}
	}
	return nil, nil
}

func removeDuplicateColumns(M BinaryMatrix) BinaryMatrix {
	N := NewSparseBinaryMatrix(M.NumRows(), 0)
	seen := make(map[string]bool)
	cols := M.Columns()
	for _, col := range cols {
		key := col.String()
		if _, ok := seen[key]; !ok {
			seen[key] = true
			N.AppendColumn(col.Matrix())
		}
	}
	return N
}

func (R *RandomComplexGenerator) randomizeGeneral_d_2(kernelMatrix BinaryMatrix) (d_2 BinaryMatrix) {
	if R.verbose {
		log.Printf("generating d_2")
	}
	d_2Sparse := NewSparseBinaryMatrix(R.dimC_1, 0)
	for i := 0; i < R.dimC_2; i++ {
		col := RandomLinearCombination(kernelMatrix)
		d_2Sparse.AppendColumn(col)
	}
	if R.verbose {
		log.Printf("generated d_2: %v\n", d_2Sparse)
	}
	return d_2Sparse
}

func (R *RandomComplexGenerator) randomizeDimC_1(simplicial bool) (int, error) {
	if simplicial {
		if R.dimC_0 < 2 {
			return 0, nil
		}
	}
	min := R.dimC_0 / 2 // minimum number of edges in a connected graph
	if min < 1 {
		min = 1
	}
	max := R.dimC_0 * (R.dimC_0 - 1) / 2 // number of edges in a complete graph
	if max < 1 {
		max = 1
	}
	if max <= min {
		max = min + 1
	}
	d, err := rand.Int(rand.Reader, big.NewInt(int64(max - min)))
	if err != nil {
		return -1, err
	}
	return min + int(d.Int64()), nil
}

func (R *RandomComplexGenerator) randomizeDimC_2() (int, error) {
	min := R.dimC_1 / 3 // minimum number of triangles such that each edge could be in at least 1 triangle
	if min < 1 {
		min = 1
	}
	max := R.dimC_0 * (R.dimC_0 - 1) * (R.dimC_0 - 2) / 3 // number of triangles for a clique complex
	if max <= min {
		max = min + 1
	}
	d, err := rand.Int(rand.Reader, big.NewInt(int64(max - min)))
	if err != nil {
		return -1, err
	}
	return min + int(d.Int64()), nil
}

func randomizeDimB_1(dimZ_1 int) (int, error) {
	min := 0
	max := dimZ_1
	if max <= min {
		max = min + 1
	}
	d, err := rand.Int(rand.Reader, big.NewInt(int64(max - min)))
	if err != nil {
		return -1, err
	}
	if d.Int64() == 0 {
		return 1, nil
	}
	return min + int(d.Int64()), nil
}

func randFloat() float64 {
	d, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		panic(err)
	}
	return float64(d.Int64()) / 1000000.0
}
