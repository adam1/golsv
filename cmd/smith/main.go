package main

import (
	"flag"
	"golsv"
	"log"
	"golang.org/x/sys/cpu"
)

// Usage:
//
//    smith -in M.txt -smith D.txt -rowops rowops.txt -colops colops.txt
//
// The program computes the Smith normal form of the matrix M.  We
// think of the matrix M as being factorized as
//
//   M = P D Q,
//
// where
//
//   P is an automorphism of the codomain of M,
//   D is a diagonal matrix, and
//   Q is an automorphism of the domain of M.
//
// Reduction is a potentially lengthy process, as is computing the
// matrices P and Q, hence rather than doing all of the computation in
// one go, we break it up smaller steps.  The first step is to do the
// reduction to Smith normal form and keep track of the elementary row
// and column operations performed.  This program does that and
// produces output files containing the reduced matrix D, the row
// operations, and the column operations.  The latter may converted to
// explicit matrices with the `automorphism` program.
//
// The D.txt file contains the reduced matrix D; it is in our custom
// Sparse text format. The rowops.txt and colops.txt files are text
// files.
func main() {
	if cpu.X86.HasAVX2 {
		log.Printf("AVX2 is supported")
	}
	if cpu.X86.HasSSE2 {
		log.Printf("SSE2 is supported")
	}
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()
	var M golsv.BinaryMatrix

	if args.M != "" {
		log.Printf("reading M from %s", args.M)
		M = golsv.ReadSparseBinaryMatrixFile(args.M)
		log.Printf("done; read %s", M)
	}

	// xxx testing; faster to reduce dense to begin with?
// 	log.Printf("xxx converting to dense matrix")
// 	M = M.(*golsv.Sparse).DenseSubmatrix(0, M.NumRows(), 0, M.NumColumns())
// 	log.Printf("xxx done converting to dense matrix: %v", M)

	R := golsv.NewDiagonalReducer(args.Verbose)
	D, rowOps, colOps := R.Reduce(M)

	D = D.Sparse() // in case we started dense above

	if args.D != "" {
		log.Printf("writing D to %s", args.D)
		D.(*golsv.Sparse).WriteFile(args.D)
	}
	if args.RowOps != "" {
		log.Printf("writing rowops to %s", args.RowOps)
		golsv.WriteOperationsFile(args.RowOps, rowOps)
	}
	if args.ColOps != "" {
		log.Printf("writing colops to %s", args.ColOps)
		golsv.WriteOperationsFile(args.ColOps, colOps)
	}
	log.Printf("done!")
}

type Args struct {
	golsv.ProfileArgs
	M string
	D string
	RowOps string
	ColOps string
	Verbose bool
}

func parseFlags() *Args {
	args := Args{
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.StringVar(&args.M, "in", args.M, "matrix input file (Sparse txt format)")
	flag.StringVar(&args.D, "smith", args.D, "smith normal form matrix output file (Sparse txt format)")
	flag.StringVar(&args.RowOps, "rowops", args.RowOps, "rowops output file (text)")
	flag.StringVar(&args.ColOps, "colops", args.ColOps, "colops output file (text)")
	flag.Parse()
	if args.M == "" {
		flag.Usage()
		log.Fatal("missing required -in flag")
	}
	return &args
}
