package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
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
	D1File                       string
	D2File                       string
	EdgeBasisFile                string
	GraphvizFile                 string
	SubcomplexNumTriangles       int
	SubcomplexByEdgesFile        string
	SubcomplexD1File             string
	SubcomplexD2File             string
	SubcomplexOutVertexBasisFile string
	SubcomplexOutEdgeBasisFile   string
	TriangleBasisFile            string
	Verbose                      bool
	VertexBasisFile              string
	golsv.ProfileArgs
}

func parseFlags() ComplexArgs {
	args := ComplexArgs{}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", "", "The file containing the D1 boundary matrix")
	flag.StringVar(&args.D2File, "d2", "", "The file containing the D2 boundary matrix")
	flag.StringVar(&args.EdgeBasisFile, "edge-basis", "", "The file containing the edge basis")
	flag.StringVar(&args.GraphvizFile, "graphviz", "", "The file to write the graphviz output to")
	flag.IntVar(&args.SubcomplexNumTriangles, "subcomplex-num-triangles", 0, "The number of triangles to include in the subcomplex")
	flag.StringVar(&args.SubcomplexByEdgesFile, "subcomplex-by-edges", "", "The input matrix file whose columns define edges to include in the subcomplex")
	flag.StringVar(&args.SubcomplexD1File, "subcomplex-d1", "", "The output file for the D1 boundary matrix of the subcomplex")
	flag.StringVar(&args.SubcomplexD2File, "subcomplex-d2", "", "The output file for the D2 boundary matrix of the subcomplex")
	flag.StringVar(&args.SubcomplexOutVertexBasisFile, "subcomplex-out-vertex-basis", "", "The output file for the vertex basis of the subcomplex")
	flag.StringVar(&args.SubcomplexOutEdgeBasisFile, "subcomplex-out-edge-basis", "", "The output file for the edge basis of the subcomplex")
	flag.StringVar(&args.TriangleBasisFile, "triangle-basis", "", "The file containing the triangle basis")
	flag.BoolVar(&args.Verbose, "verbose", false, "Print verbose output")
	flag.StringVar(&args.VertexBasisFile, "vertex-basis", "", "The file containing the vertex basis")
	flag.Parse()

	if args.D1File == "" && args.VertexBasisFile == "" {
		fmt.Println("At least one of d1 or vertex-basis is required")
		flag.PrintDefaults()
	}
	if args.SubcomplexNumTriangles > 0 && args.SubcomplexByEdgesFile != "" {
		fmt.Println("Only one of subcomplex-num-triangles or subcomplex-by-edges may be specified")
		flag.PrintDefaults()
	}
	return args
}

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	if args.VertexBasisFile != "" {
		handleComplexByBases(args)
		return
	}
	complex := complexFromBoundaryMatrices(args)
	if args.Verbose {
		log.Printf("loaded complex %s", complex)
	}
	if args.GraphvizFile != "" {
		golsv.WriteComplexGraphvizFile(complex, args.GraphvizFile, args.Verbose)
	}
	if args.SubcomplexNumTriangles > 0 {
		subcomplex := complex.SubcomplexByTriangles(args.SubcomplexNumTriangles)
		if args.Verbose {
			log.Printf("created subcomplex %s", subcomplex)
		}
		writeSubcomplex(subcomplex, args)
	}
}

func handleComplexByBases(args ComplexArgs) {
	complex := complexFromBasesFiles(args)
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
		writeSubcomplex(subcomplex, args)
		if args.GraphvizFile != "" {
			golsv.WriteComplexGraphvizFile(subcomplex, args.GraphvizFile, args.Verbose)
		}
		return
	}
	if args.GraphvizFile != "" {
		golsv.WriteComplexGraphvizFile(complex, args.GraphvizFile, args.Verbose)
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
	// xxx we could make D2 optional
	if args.D2File == "" {
		log.Fatalf("D2 file is required")
	}
	if args.Verbose {
		log.Printf("reading matrix d_1 from %s", args.D1File)
	}
	D1 := golsv.ReadSparseBinaryMatrixFile(args.D1File)
	if args.Verbose {
		log.Printf("done; read %s", D1)
	}
	if args.Verbose {
		log.Printf("reading matrix d_2 from %s", args.D2File)
	}
	D2 := golsv.ReadSparseBinaryMatrixFile(args.D2File)
	if args.Verbose {
		log.Printf("done; read %s", D2)
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

func writeSubcomplex[T any](subcomplex *golsv.ZComplex[T], args ComplexArgs) {
	if args.SubcomplexD1File != "" {
		if args.Verbose {
			log.Printf("writing subcomplex d_1 to %s", args.SubcomplexD1File)
		}
		subcomplex.D1().Sparse().WriteFile(args.SubcomplexD1File)
		if args.Verbose {
			log.Printf("done writing subcomplex d_1")
		}
	}
	if args.SubcomplexD2File != "" {
		if args.Verbose {
			log.Printf("writing subcomplex d_2 to %s", args.SubcomplexD2File)
		}
		subcomplex.D2().Sparse().WriteFile(args.SubcomplexD2File)
		if args.Verbose {
			log.Printf("done writing subcomplex d_2")
		}
	}
	if args.SubcomplexOutVertexBasisFile != "" {
		if args.Verbose {
			log.Printf("writing subcomplex vertex basis to %s", args.SubcomplexOutVertexBasisFile)
		}
		golsv.WriteStringFile(subcomplex.VertexBasis(), args.SubcomplexOutVertexBasisFile)
		if args.Verbose {
			log.Printf("done writing subcomplex vertex basis")
		}
	}
	if args.SubcomplexOutEdgeBasisFile != "" {
		if args.Verbose {
			log.Printf("writing subcomplex edge basis to %s", args.SubcomplexOutEdgeBasisFile)
		}
		golsv.WriteStringFile(subcomplex.EdgeBasis(), args.SubcomplexOutEdgeBasisFile)
		if args.Verbose {
			log.Printf("done writing subcomplex edge basis")
		}
	}
}
