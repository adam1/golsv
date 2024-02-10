package main

import (
	"flag"
	"log"
	"golsv"
)

// Usage:
//
//   automorphism -dim n -ops ops.txt -out M.txt
//
// The program generates the identity matrix of size n, and then
// applies the operations in ops.txt to it.  The result is written as
// a sparse matrix to M.txt.
func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	// xxx we could avoid the transpose: create M in row-major order,
	// apply the ops as row ops (still in reverse order). this should
	// be fast and not require too much work.  we could then crop
	// columns while converting to Sparse at the same time.
	//
	// however, the subsequent crop/sparse still amounts in the same
	// amount of work.  hence is not worth the effort.

 	M := golsv.NewDenseBinaryMatrixIdentity(args.Dim)
	var reader golsv.OperationReader
	if args.Reverse {
		reader = golsv.OpenOperationFileReverse(args.In)
	} else {
		reader = golsv.OpenOperationFile(args.In)
	}
	streamer := golsv.NewOpsFileMatrixStreamer(reader, M, args.Verbose)
	streamer.Stream()
	log.Printf("done computing matrix M: %v", M)

	if args.Transpose {
		log.Printf("transposing matrix")
		M = M.Transpose()
	}
	if args.CropStart > 0 || args.CropEnd < args.Dim {
		log.Printf("cropping to columns %d-%d", args.CropStart, args.CropEnd)
		M = M.Submatrix(0, M.NumRows(), args.CropStart, args.CropEnd)
	}
	log.Printf("computing sparse form")
	S := M.Sparse()

	if args.Density {
		log.Printf("checking density")
		density := S.Density(0, 0)
		log.Printf("density: %f", density)
	}
	if args.Out != "" {
		log.Printf("writing sparse matrix to %s", args.Out)
		S.WriteFile(args.Out)
	}
	log.Printf("done")
}

type Args struct {
	golsv.ProfileArgs
	Density bool
	Dim int
	In string
	Out string
	CropStart int
	CropEnd int
	Reverse bool
	Transpose bool
	Verbose bool
}

func parseFlags() *Args {
	args := Args{
		ProfileArgs: golsv.ProfileArgs{
			ProfileType: "cpu",
			Toggled: true,
		},
		CropStart: -1,
		CropEnd: -1,
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.IntVar(&args.CropStart, "crop-start", args.CropStart, "crop columns before this index (-1 for no crop)")
	flag.IntVar(&args.CropEnd, "crop-end", args.CropEnd, "crop columns at this index and higher (-1 for no crop)")
	flag.BoolVar(&args.Density, "density", args.Density, "compute density")
	flag.IntVar(&args.Dim, "dim", args.Dim, "dimension of matrix")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.BoolVar(&args.Reverse, "reverse", args.Reverse, "reverse the order of the operations")
	flag.BoolVar(&args.Transpose, "transpose", args.Transpose, "transpose the matrix")
	flag.StringVar(&args.In, "in", args.In, "ops input file (custom txt format)")
	flag.StringVar(&args.Out, "out", args.Out, "matrix output file (Sparse txt format)")
	flag.Parse()
	if args.Dim < 0 {
		flag.Usage()
		log.Fatal("-dim must be greater than 0")
	}
	if args.CropStart == -1 {
		args.CropStart = 0
	}
	if args.CropEnd == -1 {
		args.CropEnd = args.Dim
	}
	if args.In == "" {
		flag.Usage()
		log.Fatal("missing required -in flag")
	}
	return &args
}
