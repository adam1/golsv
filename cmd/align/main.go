package main

import (
	"flag"
	"log"
	"golsv"
)

// Usage:
//
//   align -B1smith B1smith.txt -P P.txt -B1colops B1colops.txt -Z1 Z1.txt -out U.txt
//
// The align program computes the matrix U whose columns, together
// with the columns of B1, form a basis for Z1. Matrix P is the row
// operations matrix obtained during reduction of B1 to B1smith.
func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading B1smith from %s", args.B1smith)
	B1smith := golsv.ReadSparseBinaryMatrixFile(args.B1smith)
	log.Printf("read B1smith: %v", B1smith)

	log.Printf("reading P from %s", args.P)
	P := golsv.ReadSparseBinaryMatrixFile(args.P)
	log.Printf("read P: %v; computing dense form", P)

	Pdense := P.(*golsv.Sparse).DenseSubmatrix(0, P.NumRows(), 0, P.NumColumns())
	log.Printf("Pdense = %v", Pdense)

	log.Printf("reading B1colops from %s", args.B1colops)
	B1colops := golsv.ReadOperationsFile(args.B1colops)
	log.Printf("read B1colops: %d operations", len(B1colops))

	log.Printf("reading Z1 from %s", args.Z1)
	Z1 := golsv.ReadSparseBinaryMatrixFile(args.Z1)
	log.Printf("read Z1: %v", Z1)

	log.Printf("beginning alignment")
	aligner := golsv.NewAligner(B1smith.(*golsv.Sparse),
		Pdense,
		B1colops,
		Z1.(*golsv.Sparse),
		args.Verbose)
	U := aligner.Align()
	log.Printf("U = %v", U)

	log.Printf("writing U to %s", args.Out)
	U.(*golsv.Sparse).WriteFile(args.Out)
	log.Printf("done")
}

type Args struct {
	golsv.ProfileArgs
	B1smith string
	P string
	B1colops string
	Z1 string
	Out string
	Verbose bool
}

func parseFlags() *Args {
	args := Args{
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.StringVar(&args.B1smith, "B1smith", args.B1smith, "matrix input file (Sparse text format)")
	flag.StringVar(&args.P, "P", args.P, "P matrix input file (Sparse text format)")
	flag.StringVar(&args.B1colops, "B1colops", args.B1colops, "B1 column operations file")
	flag.StringVar(&args.Z1, "Z1", args.Z1, "matrix input file (Sparse text format)")
	flag.StringVar(&args.Out, "out", args.Out, "matrix output file (Sparse text format)")
	flag.Parse()
	if args.B1smith == "" {
		flag.Usage()
		log.Fatal("missing required -B1smith flag")
	}
	if args.P == "" {
		flag.Usage()
		log.Fatal("missing required -P flag")
	}
	if args.B1colops == "" {
		flag.Usage()
		log.Fatal("missing required -B1colops flag")
	}
	if args.Z1 == "" {
		flag.Usage()
		log.Fatal("missing required -Z1 flag")
	}
	if args.Out == "" {
		flag.Usage()
		log.Fatal("missing required -out flag")
	}
	return &args
}
