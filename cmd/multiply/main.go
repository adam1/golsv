package main

import (
	"flag"
	"log"
	"golsv"
)

// Usage:
//
//   multiply -A A.txt -B B.txt -out C.txt
//
// The program reads BinaryMatrix's from A.txt and B.txt in sparse
// column support format, computes the product C = A * B, and writes
// the result to C.txt in sparse column support format.

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading matrix A from %s", args.A)
	A := golsv.ReadSparseBinaryMatrixFile(args.A)
	log.Printf("done; read %s", A)

	log.Printf("converting A to dense")
	Adense := A.Dense()
	log.Printf("done converting A to dense")

	log.Printf("reading matrix B from %s", args.B)
	B := golsv.ReadSparseBinaryMatrixFile(args.B)
	log.Printf("done; read %s", B)

	log.Printf("computing C = A * B")
	C := Adense.MultiplyRight(B)
	
	if args.C != "" {

		log.Printf("converting C to sparse")
		Csparse := C.(*golsv.DenseBinaryMatrix).Sparse()

		log.Printf("writing matrix C to %s", args.C)
		Csparse.WriteFile(args.C)
	}
	log.Printf("done")
}

type Args struct {
	golsv.ProfileArgs
	A string
	B string
	C string
	Verbose bool
}

func parseFlags() *Args {
	args := Args{
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.StringVar(&args.A, "A", args.A, "matrix A input file (sparse column support txt format)")
	flag.StringVar(&args.B, "B", args.B, "matrix B input file")
	flag.StringVar(&args.C, "C", args.C, "matrix C output file")
	flag.Parse()
	if args.A == "" {
		flag.Usage()
		log.Fatal("missing required -A flag")
	}
	if args.B == "" {
		flag.Usage()
		log.Fatal("missing required -B flag")
	}
	return &args
}
