package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
	"os"
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
	D1File                 string
	D2File                 string
	GraphvizFile           string
	SubcomplexNumTriangles int
	SubcomplexD1File       string
	SubcomplexD2File       string
	Verbose                bool
	golsv.ProfileArgs
}

func parseFlags() ComplexArgs {
	args := ComplexArgs{}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", "", "The file containing the D1 boundary matrix")
	flag.StringVar(&args.D2File, "d2", "", "The file containing the D2 boundary matrix")
	flag.StringVar(&args.GraphvizFile, "graphviz", "", "The file to write the graphviz output to")
	flag.IntVar(&args.SubcomplexNumTriangles, "subcomplex-num-triangles", 0, "The number of triangles to include in the subcomplex")
	flag.StringVar(&args.SubcomplexD1File, "subcomplex-d1", "", "The output file for the D1 boundary matrix of the subcomplex")
	flag.StringVar(&args.SubcomplexD2File, "subcomplex-d2", "", "The output file for the D2 boundary matrix of the subcomplex")
	flag.BoolVar(&args.Verbose, "verbose", false, "Print verbose output")
	flag.Parse()

	if args.D1File == "" {
		fmt.Println("D1 matrix file is required")
		flag.PrintDefaults()
	}
	if args.D2File == "" {
		fmt.Println("D2 matrix file is required")
		flag.PrintDefaults()
	}
	if args.SubcomplexNumTriangles > 0 {
		if args.SubcomplexD1File == "" {
			fmt.Println("Subcomplex D1 matrix file is required")
			flag.PrintDefaults()
		}
		if args.SubcomplexD2File == "" {
			fmt.Println("Subcomplex D2 matrix file is required")
			flag.PrintDefaults()
		}
	}
	return args
}

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

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
	C := golsv.NewZComplexFromBoundaryMatrices(D1, D2)
	if args.Verbose {
		log.Printf("done; created complex %s", C)
	}
	if args.GraphvizFile != "" {
		writeGraphvizFile(C, args)
	}
	if args.SubcomplexNumTriangles > 0 {
		writeSubcomplex(C, args)
	}
}

func writeGraphvizFile[T any](C *golsv.ZComplex[T], args ComplexArgs) {
	viz := golsv.NewZComplexToGraphviz(C)
	G, err := viz.Graphviz()
	if err != nil {
		log.Fatalf("error creating graphviz: %v", err)
	}
	if args.Verbose {
		log.Printf("writing graphviz file to %s", args.GraphvizFile)
	}
	f, err := os.Create(args.GraphvizFile)
	if err != nil {
		log.Fatalf("error creating graphviz file: %v", err)
	}
	defer f.Close()
	s := G.String()
	if err := os.WriteFile(args.GraphvizFile, []byte(s), 0644); err != nil {
		log.Fatalf("error writing graphviz file: %v", err)
	}
	if args.Verbose {
		log.Printf("done writing graphviz file")
	}
}

func writeSubcomplex[T any](C *golsv.ZComplex[T], args ComplexArgs) {
	subcomplex := C.SubcomplexByTriangles(args.SubcomplexNumTriangles)
	if args.Verbose {
		log.Printf("created subcomplex %s", subcomplex)
	}
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
}


	
