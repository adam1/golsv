package main

import (
	"flag"
	"golsv"
	"log"
)

// Usage:
//
//   cayley -base-field=F4 -d1=d1.txt -d2=d2.txt -truncate-generators=0 -max-depth=0
//
//   cayley -base-field=F4 -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt
//
func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	lsv := golsv.NewLsvContext(args.BaseField)
	gens := prepareGenerators(lsv, args)

	E := golsv.NewCayleyExpander(lsv, gens, args.MaxDepth, args.Verbose)
	E.Expand()
	log.Printf("computing complex")
	complex := E.Complex(args.sortBases)
	log.Printf("complex: %s", complex)

	vertexBasis := complex.VertexBasis()
	edgeBasis := complex.EdgeBasis()
	triangleBasis := complex.TriangleBasis()

	D1sparse := complex.D1().(*golsv.Sparse)
	if args.D1File != "" {
		log.Printf("writing d1 to %s", args.D1File)
		D1sparse.WriteFile(args.D1File)
	}
	D2sparse := complex.D2().(*golsv.Sparse)
	if args.D2File != "" {
		log.Printf("writing d2 to %s", args.D2File)
		D2sparse.WriteFile(args.D2File)
	}
	if args.VertexBasisFile != "" {
		log.Printf("writing vertex basis to %s", args.VertexBasisFile)
		golsv.WriteVertexFile(vertexBasis, args.VertexBasisFile)
	}
	if args.EdgeBasisFile != "" {
		log.Printf("writing edge basis to %s", args.EdgeBasisFile)
		golsv.WriteEdgeFile(edgeBasis, args.EdgeBasisFile)
	}
	if args.TriangleBasisFile != "" {
		log.Printf("writing triangle basis to %s", args.TriangleBasisFile)
		golsv.WriteTriangleFile(triangleBasis, args.TriangleBasisFile)
	}
	log.Printf("done")
}

func prepareGenerators(lsv *golsv.LsvContext, args *CayleyExpanderArgs) []golsv.MatGF {
	gens := lsv.Generators()
	if args.TruncateGenerators > 0 {
		log.Printf("truncating generators to %d", args.TruncateGenerators)
		gens = gens[:args.TruncateGenerators]
	}
	return gens
}

type CayleyExpanderArgs struct {
	BaseField		   string
	TruncateGenerators int
    MaxDepth           int
	Verbose            bool
	D1File             string
	D2File             string
	VertexBasisFile    string
	EdgeBasisFile      string
	TriangleBasisFile  string
	sortBases          bool
	golsv.ProfileArgs
}

func parseFlags() *CayleyExpanderArgs {
	args := CayleyExpanderArgs{
		BaseField: "F4",
		MaxDepth: -1,
		Verbose: true,
		sortBases: false,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.BaseField, "base-field", args.BaseField, "base field: F4 or F16")
	flag.IntVar(&args.TruncateGenerators, "truncate-generators", 0, "truncate generators")
	flag.IntVar(&args.MaxDepth, "max-depth", args.MaxDepth, "maximum depth")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.BoolVar(&args.sortBases, "sort-bases", args.sortBases, "sort bases")
	flag.StringVar(&args.D1File, "d1", args.D1File, "d1 output file (sparse column support txt format)")
	flag.StringVar(&args.D2File, "d2", args.D2File, "d2 output file (sparse column support txt format)")
	flag.StringVar(&args.VertexBasisFile, "vertex-basis", args.VertexBasisFile, "vertex basis output file (text)")
	flag.StringVar(&args.EdgeBasisFile, "edge-basis", args.EdgeBasisFile, "edge basis output file (text)")
	flag.StringVar(&args.TriangleBasisFile, "triangle-basis", args.TriangleBasisFile, "triangle basis output file (text)")
	flag.Parse()
	return &args
}
