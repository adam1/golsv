package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
)

// Usage:
//
//   random-complex -d1 d1.txt -d2 d2.txt -dimC0 100
//
func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	gen := golsv.NewRandomComplexGenerator(args.DimC0, args.Verbose)
	var err error
	var d_1, d_2 golsv.BinaryMatrix
	if args.Clique {
		d_1, d_2, err = gen.RandomCliqueComplex(args.ProbEdge)
	} else if args.Simplicial {
		d_1, d_2, err = gen.RandomSimplicialComplex()
	} else {
		d_1, d_2, err = gen.RandomComplex()
	}
	if err != nil {
		log.Fatalf("Error generating random complex: %v", err)
	}
	if args.Verbose {
		log.Printf("Generated complex: d_2=%v d_1=%v", d_2, d_1)
	}
	if args.D1File != "" {
		if args.Verbose {
			log.Printf("Writing d1 to %s", args.D1File)
		}
		d_1.Sparse().WriteFile(args.D1File)
	}
	if args.D2File != "" {
		if args.Verbose {
			log.Printf("Writing d2 to %s", args.D2File)
		}
		d_2.Sparse().WriteFile(args.D2File)
	}
}

type Args struct {
	D1File             string
	D2File             string
	DimC0              int
	Simplicial         bool
	Clique             bool
	ProbEdge           float64
	Verbose            bool
	golsv.ProfileArgs
}

func parseFlags() *Args {
	args := Args{
		DimC0: 10,
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.Clique, "clique", args.Clique, "Generate a clique complex over a random graph")
	flag.StringVar(&args.D1File, "d1", args.D1File, "d1 output file (sparse column support txt format)")
	flag.StringVar(&args.D2File, "d2", args.D2File, "d2 output file (sparse column support txt format)")
	flag.IntVar(&args.DimC0, "dimC0", args.DimC0, fmt.Sprintf("dim C_0 (default %d)", args.DimC0))
	flag.Float64Var(&args.ProbEdge, "p", args.ProbEdge, "probability of edge in random graph")
	flag.BoolVar(&args.Simplicial, "simplicial", args.Simplicial, "complex should be simplicial")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.Parse()
	return &args
}
