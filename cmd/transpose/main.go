package main

import (
	"flag"
	"log"
	"golsv"
)

// Usage:
//
//   transpose -in in.txt -out out.txt
//
// The program reads a BinaryMatrix from in.txt in sparse column
// support format, computes its transpose, column support format,
// computes the transpose, and writes the result to out.txt in sparse
// column support format.

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading matrix from %s", args.in)
	var M golsv.BinaryMatrix
	M = golsv.ReadSparseBinaryMatrixFile(args.in)
	log.Printf("done; read %s", M)
	M.(*golsv.Sparse).SetVerbose(args.Verbose)

	if args.dense {
		log.Printf("computing dense form")
		M = M.DenseSubmatrix(0, M.NumRows(), 0, M.NumColumns())
	}
	T := M.Transpose()
	log.Printf("computing sparse form")
	S := T.Sparse()
	
	if args.out != "" {
		log.Printf("writing matrix to %s", args.out)
		S.WriteFile(args.out)
	}
	log.Printf("done")
}

type Args struct {
	golsv.ProfileArgs
	in string
	out string
	dense bool
	Verbose bool
}

func parseFlags() *Args {
	args := Args{
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.BoolVar(&args.dense, "dense", args.dense, "do transpose in dense form intermediately")
	flag.StringVar(&args.in, "in", args.in, "matrix input file (sparse column support txt format)")
	flag.StringVar(&args.out, "out", args.out, "matrix output file")
	flag.Parse()
	if args.in == "" {
		flag.Usage()
		log.Fatal("missing required -in flag")
	}
	return &args
}
