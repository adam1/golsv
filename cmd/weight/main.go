package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"golsv"
)

// Usage:
//
//   weight -min -in U.txt
//
// The program reads BinaryMatrix U from U.txt in sparse column
// support format, and computes the minimum weight of a nonzero vector
// in the column space of U.
//
// xxx Problem? a linear combination of columns of U is not
// necessarily outside of B_1.  For example, add a vector to itself.
// If that is correct, delete this file.

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading matrix U from %s", args.in)
	var U golsv.BinaryMatrix
	U = golsv.ReadSparseBinaryMatrixFile(args.in)
	log.Printf("done; read %s", U)

	log.Printf("computing dense form of U")
	Udense := U.DenseSubmatrix(0, U.NumRows(), 0, U.NumColumns())
	log.Printf("done; dense form of U is %s", Udense)

	if args.min {
		log.Printf("computing minimum nonzero weight")
		minWeight := math.MaxInt
		golsv.EnumerateBinaryVectorSpace(Udense, func(v golsv.BinaryMatrix) bool {
			weight := v.(*golsv.DenseBinaryMatrix).ColumnWeight(0)
			if weight > 0 && weight < minWeight {
				minWeight = weight
				log.Printf("new min weight: %d", minWeight)
			}
			return true
		})
		log.Printf("done; min weight is %d", minWeight)
		if args.out != "" {
			log.Printf("writing min weight to %s", args.out)
			err := os.WriteFile(args.out, []byte(fmt.Sprintf("%d\n", minWeight)), 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

type Args struct {
	golsv.ProfileArgs
	in string
	out string
	min bool
	verbose bool
}

func parseFlags() *Args {
	args := Args{
		verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.verbose, "verbose", args.verbose, "verbose logging")
	flag.BoolVar(&args.min, "min", args.min, "compute minimum weight")
	flag.StringVar(&args.in, "in", args.in, "matrix input file (sparse column support txt format)")
	flag.StringVar(&args.out, "out", args.out, "minimum weight output file (text)")
	flag.Parse()
	if args.in == "" {
		flag.Usage()
		log.Fatal("missing required -in flag")
	}
	if !args.min {
		flag.Usage()
		log.Fatal("only -min is supported")
	}
	return &args
}
