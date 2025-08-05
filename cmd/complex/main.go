package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
	"strings"
)

// Input:
//
//   - the boundary matrices defining a complex
//
// Output:
//
//   (optional) - a graphviz file representing the 1-skeleton of the
//   complex

type ComplexArgs struct {
	D1File                string
	D2File                string
	DualComplex           bool
	EdgeBasisFile         string
	DepthFiltration       bool
	GraphvizFile          string
	ReportRegularity      bool
	SubcomplexByDepth     int
	SubcomplexByEdgesFile string
	OutComplexD1File      string
	OutComplexD2File      string
	OutVertexBasisFile    string
	OutEdgeBasisFile      string
	OutTriangleBasisFile  string
	TriangleBasisFile     string
	Verbose               bool
	VertexBasisFile       string
	golsv.ProfileArgs
}

func parseFlags() ComplexArgs {
	args := ComplexArgs{
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", "", "The file containing the D1 boundary matrix")
	flag.StringVar(&args.D2File, "d2", "", "The file containing the D2 boundary matrix")
	flag.BoolVar(&args.DualComplex, "dual-complex", false, "Compute the dual complex (see zcomplex.go for precise definition)")
	flag.StringVar(&args.EdgeBasisFile, "edge-basis", "", "The file containing the edge basis")
	flag.BoolVar(&args.DepthFiltration, "depth-filtration", false, "Compute filtration subcomplexes by depth; output filenames will have '-d' appended with d=depth")
	flag.StringVar(&args.GraphvizFile, "graphviz", "", "The file to write the graphviz output to")
	flag.BoolVar(&args.ReportRegularity, "regular", false, "Report whether the complex is regular and its degree")
	flag.IntVar(&args.SubcomplexByDepth, "subcomplex-depth", -1, "The depth of the subcomplex to extract by BFS from first vertex")
	flag.StringVar(&args.SubcomplexByEdgesFile, "subcomplex-by-edges-matrix", "", "The input matrix file whose columns define edges to include in the subcomplex")
	flag.StringVar(&args.OutComplexD1File, "out-d1", "", "The output file for the D1 boundary matrix of the subcomplex")
	flag.StringVar(&args.OutComplexD2File, "out-d2", "", "The output file for the D2 boundary matrix of the subcomplex")
	flag.StringVar(&args.OutVertexBasisFile, "out-vertex-basis", "", "The output file for the vertex basis of the subcomplex")
	flag.StringVar(&args.OutEdgeBasisFile, "out-edge-basis", "", "The output file for the edge basis of the subcomplex")
	flag.StringVar(&args.OutTriangleBasisFile, "out-triangle-basis", "", "The output file for the triangle basis of the subcomplex")
	flag.StringVar(&args.TriangleBasisFile, "triangle-basis", "", "The file containing the triangle basis")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "Print verbose output")
	flag.StringVar(&args.VertexBasisFile, "vertex-basis", "", "The file containing the vertex basis")
	flag.Parse()

	if args.D1File == "" && args.VertexBasisFile == "" {
		fmt.Println("At least one of d1 or vertex-basis is required")
		flag.PrintDefaults()
	}
	if args.SubcomplexByDepth > 0 && args.SubcomplexByEdgesFile != "" {
		fmt.Println("subcomplex-depth and subcomplex-by-edges-matrix are mutually exclusive")
		flag.PrintDefaults()
	}
	if args.SubcomplexByDepth > 0 && args.DepthFiltration {
		fmt.Println("subcomplex-depth and depth-filtration are mutually exclusive")
		flag.PrintDefaults()
	}
	if args.SubcomplexByEdgesFile != "" && args.EdgeBasisFile == "" {
		fmt.Println("subcomplex-by-edges-matrix requires edge-basis")
		flag.PrintDefaults()
	}
	return args
}

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	// Nb. There are two ways to specify the complex: by boundary matrices or by bases.
	// These lead to different types of complexes. The former is a ZComplex of ZVertexInt
	// and the latter is a ZComplex of ElementCalG. At the moment, we lack a way to handle
	// these polymorphically in the same codepath, so we handle them separately, which
	// involves some redundancy.
	if args.VertexBasisFile != "" {
		handleComplexByBases(args)
	} else {
		handleComplexByBoundaryMatrices(args)
	}
}

func handleComplexByBoundaryMatrices(args ComplexArgs) {
	var complex *golsv.ZComplex[golsv.ZVertexInt] = complexFromBoundaryMatrices(args)
	if args.Verbose {
		log.Printf("loaded complex %s", complex)
	}
	if args.SubcomplexByDepth > 0 {
		if args.Verbose {
			log.Printf("creating subcomplex by depth %d", args.SubcomplexByDepth)
		}
		subcomplex := complex.SubcomplexByDepth(args.SubcomplexByDepth)
		if args.Verbose {
			log.Printf("created subcomplex %s", subcomplex)
		}
		writeOutputComplex(subcomplex, args)
		complex = subcomplex
	} else if args.DepthFiltration {
		if args.Verbose {
			log.Printf("creating depth filtration subcomplexes")
		}
		h := func(depth int, subcomplex *golsv.ZComplex[golsv.ZVertexInt], verticesAtDepth []golsv.ZVertex[golsv.ZVertexInt]) {
			writeDepthSubcomplex(subcomplex, args, depth)
		}
		complex.DepthFiltration(complex.VertexBasis()[0], h)
	}
	if args.GraphvizFile != "" {
		golsv.WriteComplexGraphvizFile(complex, args.GraphvizFile, args.Verbose)
	}
	if args.ReportRegularity {
		reportRegularity(complex)
	}
}

func handleComplexByBases(args ComplexArgs) {
	var complex *golsv.ZComplex[golsv.ElementCalG] = complexFromBasesFiles(args)
	if args.Verbose {
		log.Printf("loaded complex %s", complex)
	}
	if args.SubcomplexByEdgesFile != "" {
		edgeMatrix := golsv.ReadSparseBinaryMatrixFile(args.SubcomplexByEdgesFile)
		edges := rowSupports(edgeMatrix.Sparse())
		subcomplex := complex.SubcomplexByEdges(edges)
		if args.Verbose {
			log.Printf("created subcomplex %s", subcomplex)
		}
		writeOutputComplex(subcomplex, args)
		complex = subcomplex
	} else if args.SubcomplexByDepth > 0 {
		if args.Verbose {
			log.Printf("creating subcomplex by depth %d", args.SubcomplexByDepth)
		}
		subcomplex := complex.SubcomplexByDepth(args.SubcomplexByDepth)
		if args.Verbose {
			log.Printf("created subcomplex %s", subcomplex)
		}
		writeOutputComplex(subcomplex, args)
		complex = subcomplex
	} else if args.DepthFiltration {
		if args.Verbose {
			log.Printf("creating depth filtration subcomplexes")
		}
		h := func(depth int, subcomplex *golsv.ZComplex[golsv.ElementCalG], verticesAtDepth []golsv.ZVertex[golsv.ElementCalG]) {
			writeDepthSubcomplex(subcomplex, args, depth)
		}
		complex.DepthFiltration(complex.VertexBasis()[0], h)
	} else if args.DualComplex {
		if args.Verbose {
			log.Printf("computing dual complex")
		}
		dual := complex.DualComplex()
		if args.Verbose {
			log.Printf("computed dual complex: %s", dual)
		}
		writeOutputComplex(dual, args)
	}
	if args.GraphvizFile != "" {
		golsv.WriteComplexGraphvizFile(complex, args.GraphvizFile, args.Verbose)
	}
	if args.ReportRegularity {
		reportRegularity(complex)
	}
}

func complexFromBasesFiles(args ComplexArgs) *golsv.ZComplex[golsv.ElementCalG] {
	if args.VertexBasisFile == "" {
		log.Fatalf("vertex basis file is required")
	}
	if args.EdgeBasisFile == "" {
		log.Fatalf("edge basis file is required")
	}
	return golsv.NewZComplexElementCalGFromBasisFiles(
		args.VertexBasisFile, args.EdgeBasisFile, args.TriangleBasisFile,
		args.Verbose)
}

func complexFromBoundaryMatrices(args ComplexArgs) *golsv.ZComplex[golsv.ZVertexInt] {
	if args.D1File == "" {
		log.Fatalf("D1 file is required")
	}
	if args.Verbose {
		log.Printf("reading matrix d_1 from %s", args.D1File)
	}
	D1 := golsv.ReadSparseBinaryMatrixFile(args.D1File)
	if args.Verbose {
		log.Printf("done; read %s", D1)
	}
	var D2 golsv.BinaryMatrix
	if args.D2File == "" {
		D2 = golsv.NewSparseBinaryMatrix(D1.NumColumns(), 0)
	} else {
		if args.Verbose {
			log.Printf("reading matrix d_2 from %s", args.D2File)
		}
		D2 = golsv.ReadSparseBinaryMatrixFile(args.D2File)
		if args.Verbose {
			log.Printf("done; read %s", D2)
		}
	}
	return golsv.NewZComplexFromBoundaryMatrices(D1, D2)
}

func rowSupports(matrix *golsv.Sparse) (supports map[int]any) {
	supports = make(map[int]any)
	for j := 0; j < matrix.NumColumns(); j++ {
		support := matrix.Support(j)
		for _, i := range support {
			supports[i] = struct{}{}
		}
	}
	return supports
}

func writeOutputComplex[T any](outComplex *golsv.ZComplex[T], args ComplexArgs) {
	if args.OutComplexD1File != "" {
		if args.Verbose {
			log.Printf("writing output complex d_1 to %s", args.OutComplexD1File)
		}
		outComplex.D1().Sparse().WriteFile(args.OutComplexD1File)
		if args.Verbose {
			log.Printf("done writing")
		}
	}
	if args.OutComplexD2File != "" {
		if args.Verbose {
			log.Printf("writing output complex d_2 to %s", args.OutComplexD2File)
		}
		outComplex.D2().Sparse().WriteFile(args.OutComplexD2File)
		if args.Verbose {
			log.Printf("done writing")
		}
	}
	if args.OutVertexBasisFile != "" {
		if args.Verbose {
			log.Printf("writing output complex vertex basis to %s", args.OutVertexBasisFile)
		}
		golsv.WriteStringFile(outComplex.VertexBasis(), args.OutVertexBasisFile)
		if args.Verbose {
			log.Printf("done writing")
		}
	}
	if args.OutEdgeBasisFile != "" {
		if args.Verbose {
			log.Printf("writing output complex edge basis to %s", args.OutEdgeBasisFile)
		}
		golsv.WriteStringFile(outComplex.EdgeBasis(), args.OutEdgeBasisFile)
		if args.Verbose {
			log.Printf("done writing")
		}
	}
	if args.OutTriangleBasisFile != "" {
		if args.Verbose {
			log.Printf("writing output complex triangle basis to %s", args.OutTriangleBasisFile)
		}
		golsv.WriteStringFile(outComplex.TriangleBasis(), args.OutTriangleBasisFile)
		if args.Verbose {
			log.Printf("done writing")
		}
	}
}

func writeDepthSubcomplex[T any](subcomplex *golsv.ZComplex[T], args ComplexArgs, depth int) {
	if args.Verbose {
		log.Printf("created subcomplex %s at depth %d", subcomplex, depth)
	}
	cargs := args
	if args.OutComplexD1File != "" {
		cargs.OutComplexD1File = addTagDepthTag(args.OutComplexD1File, depth)
	}
	if args.OutComplexD2File != "" {
		cargs.OutComplexD2File = addTagDepthTag(args.OutComplexD2File, depth)
	}
	if args.OutVertexBasisFile != "" {
		cargs.OutVertexBasisFile = addTagDepthTag(args.OutVertexBasisFile, depth)
	}
	if args.OutEdgeBasisFile != "" {
		cargs.OutEdgeBasisFile = addTagDepthTag(args.OutEdgeBasisFile, depth)
	}
	if args.OutTriangleBasisFile != "" {
		cargs.OutTriangleBasisFile = addTagDepthTag(args.OutTriangleBasisFile, depth)
	}
	writeOutputComplex(subcomplex, cargs)
}

func addTagDepthTag(s string, d int) string {
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		return fmt.Sprintf("%s_%d.txt", parts[0], d)
	}
	parts[0] = fmt.Sprintf("%s_%d", parts[0], d)
	return strings.Join(parts, "-")
}

func reportRegularity[T any](complex *golsv.ZComplex[T]) {
	isRegular, degree := complex.IsRegular()
	if isRegular {
		fmt.Printf("Complex is regular with degree %d\n", degree)
	} else {
		fmt.Printf("Complex is not regular\n")
	}
}
