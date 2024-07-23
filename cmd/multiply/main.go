package main

import (
	"flag"
	"fmt"
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

	APrime:= A
	if !args.NoDense {
		log.Printf("converting A to dense")
		APrime = A.Dense()
		log.Printf("done converting A to dense")
	}
	log.Printf("reading matrix B from %s", args.B)
	B := golsv.ReadSparseBinaryMatrixFile(args.B)
	log.Printf("done; read %s", B)

	log.Printf("computing C = A * B")
	C := APrime.MultiplyRight(B)
	
	if args.C != "" {

		log.Printf("converting C to sparse")
		Csparse := C.Sparse()

		log.Printf("writing matrix C to %s", args.C)
		Csparse.WriteFile(args.C)
		log.Printf("done; wrote %s", C)
	}
	if args.IntHack {
		var s string
		for i := 0; i < A.NumRows(); i++ {
			row := A.RowVector(i)
			for j := 0; j < B.NumColumns(); j++ {
				col := B.ColumnVector(j)
				m := row.IntegerDotProduct(col)
				s += fmt.Sprintf("%d ", m)
			}
			s += "\n"
		}
		log.Printf("integer product:\n%s", s)
	}
	log.Printf("done")
}

type Args struct {
	golsv.ProfileArgs
	A string
	B string
	C string
	NoDense bool
	Verbose bool
	IntHack bool
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
	flag.BoolVar(&args.NoDense, "no-dense", args.NoDense, "do not convert A to dense")
	flag.BoolVar(&args.IntHack, "int-hack", args.IntHack, "use int instead of bool for matrix entries")
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
