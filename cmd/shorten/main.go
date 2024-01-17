package main

import (
	"flag"
	"log"
	"math"

	"golsv"
)

// Usage:
//
//   shorten -d2 d2.txt -U U.txt -vertex-basis vbasis.txt -edge-basis ebasis.txt -triangle-basis tbasis.txt
//
// The program reads column vectors from U.txt, which by previous
// preparations are known to be in Z_1 \setminus B_1, and thus
// potentially systolic. It chooses a column vector v with minimum
// weight.
//
// The program attempts to shorten the path alpha corresponding to
// vector v, only by adding or removing triangle boundaries.  The latter
// are obtained as columns of the d2 matrix.
//

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	lsv := golsv.NewLsvContext(args.BaseField)

	log.Printf("reading matrix U from %s", args.U)
	U := golsv.ReadSparseBinaryMatrixFile(args.U).(*golsv.Sparse)
	log.Printf("done; read %s", U)

	log.Printf("reading vertex basis from %s", args.vertexBasis)
	vertexBasis := golsv.ReadVertexFile(args.vertexBasis)
	log.Printf("done; read %d vertices", len(vertexBasis))

	log.Printf("reading edge basis from %s", args.edgeBasis)
	edgeBasis := golsv.ReadEdgeFile(args.edgeBasis)
	log.Printf("done; read %d edges", len(edgeBasis))

	log.Printf("reading triangle basis from %s", args.triangleBasis)
	triangleBasis := golsv.ReadTriangleFile(args.triangleBasis)
	log.Printf("done; read %d triangles", len(triangleBasis))

	sortBases := false
	C := golsv.NewComplex(vertexBasis, edgeBasis, triangleBasis, sortBases, false)
	log.Printf("complex C: %s", C)

	v := columnWithMinWeight(U).(*golsv.Sparse)
	weight := v.ColumnWeight(0)
	log.Printf("found column with min weight %d", weight)

	log.Printf("reading matrix d2 from %s", args.d2)
	d2 := golsv.ReadSparseBinaryMatrixFile(args.d2).(*golsv.Sparse)
	log.Printf("done; read %s", d2)

	alpha := C.EdgeVectorToPath(lsv, v)
	log.Printf("alpha: path with %d edges", len(alpha))

	gens := lsv.Generators()

// 	var word, vertices []golsv.MatGF
// 	word, vertices = alpha.Word(lsv)

	// EXPERIMENTAL: simple shortening
	//
	// for kicks, check all column vectors from U to see if they can
	// be simply shortened.
// 	tightenUbyZeroEdges(lsv, d2, U)
// 	tightenUbyOneEdge(lsv, d2, U)


	// EXPERIMENTAL: triangle expander
	//
	// try the TriangleExpander on alpha to see if all triangles are pulled in.
	// the TriangleExpander is not necessarily trusted.
// 	expander := golsv.NewTriangleExpander(lsv, alpha, -1, true)
// 	expander.Expand()

	// EXPERIMENTAL: chordal cycles
	chordalCycles(lsv, gens, C, d2, U)
}

func chordalCycles(lsv *golsv.LsvContext, gens []golsv.MatGF, C *golsv.Complex, d2, U *golsv.Sparse) {
	log.Printf("checking chordal cycles")
	for j := 0; j < U.NumColumns(); j++ {
		testChordsForCycle(lsv, gens, C, d2, U, j)
	}
}

func testChordsForCycle(lsv *golsv.LsvContext, gens []golsv.MatGF, C *golsv.Complex, d2, U *golsv.Sparse, col int) {
	log.Printf("checking U column=%d weight=%d", col, U.ColumnWeight(col))
	v := U.Submatrix(0, U.NumRows(), col, col+1).(*golsv.Sparse)
	alpha := C.EdgeVectorToPath(lsv, v)
	// note that the vertices here are stored with the beginning and
	// end repeated in the case that alpha is a cycle, which it should
	// be.

	// xxx factor out Path.Chords()
	chords := alpha.Chords(lsv, gens)
	log.Printf("xxx found %d chords", len(chords))
	for _, chord := range chords {
		beta, gamma := alpha.SplitByChord(lsv, chord)
		log.Printf("xxx beta: %s", beta)
		log.Printf("xxx gamma: %s", gamma)
	}


// 	word, vertices := alpha.Word(lsv)
// 	if !vertices[0].Equal(lsv, &vertices[len(vertices)-1]) {
// 		panic("alpha is not a cycle")
// 	}
// 	log.Printf("considering column %d; weight=%d len(word)=%d len(vertices)=%d",
// 		col, v.ColumnWeight(0), len(word), len(vertices))
// 	for i := 0; i < len(vertices) - 1; i++ { // -1 since the last vertex is repeated
// 		u := vertices[i]


// 		// note that we wrap around here, and we ignore vertex i
// 		// (chord start), and the vertices immediately before and
// 		// after i (since those already have edges in alpha).  so the
// 		// total number of vertices considered is len(vertices) - 3.
// 		// skip redundant chords.
// 		for k := 0; k < len(vertices) - 3; k++ {
// 			j := (i + k + 2) % len(vertices)
// 			if j < i {
// 				continue
// 			}
// 			// log.Printf("xxx considering chord from %d to %d", i, j)
// 			for _, g := range gens {
// 				if u.Translate(lsv, &g).Equal(lsv, &vertices[j]) {
// 					log.Printf("found chord from %d to %d", i, j)
// 					// use the chord to split alpha into two cycles
// 					beta, gamma := splitCycle(lsv, alpha, vertices, i, j)
// 					// and then check whether each cycle is in B_1
// 					// xxx
// 					if cycleIsBoundary(lsv, beta) {
// 						log.Printf("beta is a boundary")
// 						// xxx shorten
// 					}
// 					if cycleIsBoundary(lsv, gamma) {
// 						log.Printf("gamma is a boundary")
// 						// xxx shorten
// 					}
// 				}
// 			}
// 		}
// 	}
}

// func splitCycle(lsv *golsv.LsvContext, alpha golsv.Path, vertices []MatGF,  i, j int) (golsv.Path, golsv.Path) {
// 	// xxx beta: start at i, go across to j, then continue to i
// 	savings := j - i
// 	if i > j {
// 		savings = i - j
// 	}
// 	lenBeta := j - i + 1
// 	beta := make([]golsv.Edge, lenBeta)

// 	// xxx gamma: start at i, go around until j, then across to i
// }

func tightenUbyZeroEdges(lsv *golsv.LsvContext, d2, U *golsv.Sparse) {
	for j := 0 ; j < U.NumColumns(); j++ {
		weight := U.ColumnWeight(j)
		removeIndividualTrianglesFromColumn(lsv, d2, U, j)
		log.Printf("U column %d: weight %d -> %d", j, weight, U.ColumnWeight(j))
	}
}

func tightenUbyOneEdge(lsv *golsv.LsvContext, d2, U *golsv.Sparse) {
	for j := 0 ; j < U.NumColumns(); j++ {
		tightenColumnByOneEdge(lsv, d2, U, j)
	}
}

func tightenColumnByOneEdge(lsv *golsv.LsvContext, d2, U *golsv.Sparse, col int) {
	// Given a column vector v from U, which necessarily corresponds
	// to a cycle in Z_1 \setminus B_1.  For adjacent edges f and g in
	// v with vertices a, b c, determine if edge h from a to c exists
	// (corresponds to a generator). If so, we can remove f and g and
	// replace them with h.  The latter corresponds to adding a
	// boundary vector from B_1, namely adding the triangle abc, and
	// thus does not change the homology class of v.
	//
	//       b
	//    f / \ g
	//     /   \
	//    a ... c
	//       h
	//
	// This could be done purely at the linear algebra level, or by
	// computing with the vertices and edges in the underlying group
	// as we do here.

	// xxx

}

func removeIndividualTrianglesFromColumn(lsv *golsv.LsvContext, d2 *golsv.Sparse, U *golsv.Sparse, col int) {
	for j := 0; j < d2.NumColumns(); j++ {
		support := d2.Support(j)
		if len(support) != 3 {
			panic("expected triangle boundary")
		}
		match := 0
		for p := range d2.Support(j) {
			if U.Get(p, col) == 1 {
				match++
			}
		}
		if match == 3 {
			log.Printf("removing triangle %d", j)
			for p := range d2.Support(j) {
				U.Set(p, col, 0)
			}
		}
	}
}

func columnWithMinWeight(U golsv.BinaryMatrix) golsv.BinaryMatrix {
	minWeight := math.MaxInt
	column := -1
	for j := 0; j < U.NumColumns(); j++ {
		weight := U.(*golsv.Sparse).ColumnWeight(j)
		if weight < minWeight {
			minWeight = weight
			column = j
		}
	}
	return U.Submatrix(0, U.NumRows(), column, column+1)
}


type Args struct {
	golsv.ProfileArgs
	BaseField string
	d2 string
	U string
	edgeBasis string
	vertexBasis string
	triangleBasis string
	verbose bool
}

func parseFlags() *Args {
	args := Args{
		verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.BaseField, "base-field", args.BaseField, "base field: F4 or F16")
	flag.BoolVar(&args.verbose, "verbose", args.verbose, "verbose logging")
	flag.StringVar(&args.d2, "d2", args.d2, "matrix d2 input file (sparse column support txt format)")
	flag.StringVar(&args.U, "U", args.U, "matrix U input file (sparse column support txt format)")
	flag.StringVar(&args.edgeBasis, "edge-basis", args.edgeBasis, "edge basis input file (text)")
	flag.StringVar(&args.vertexBasis, "vertex-basis", args.vertexBasis, "vertex basis input file (text)")
	flag.StringVar(&args.triangleBasis, "triangle-basis", args.triangleBasis, "triangle basis input file (text)")
	flag.Parse()
	if args.BaseField == "" {
		flag.Usage()
		log.Fatal("missing required -base-field flag")
	}
	if args.d2 == "" {
		flag.Usage()
		log.Fatal("missing required -d2 flag")
	}
	if args.U == "" {
		flag.Usage()
		log.Fatal("missing required -U flag")
	}
	if args.edgeBasis == "" {
		flag.Usage()
		log.Fatal("missing required -edge-basis flag")
	}
	if args.vertexBasis == "" {
		flag.Usage()
		log.Fatal("missing required -vertex-basis flag")
	}
	if args.triangleBasis == "" {
		flag.Usage()
		log.Fatal("missing required -triangle-basis flag")
	}
	return &args
}
