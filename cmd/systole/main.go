package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"golsv"
)

// Usage:
//
//   systole -trials N -B B.txt -U U.txt -min min.txt
//
// The program reads BinaryMatrix B and U in sparse column support
// format.  First, it checks the weight of each column of U.  The
// minimum weight is an upper bound for the minimum weight of a
// nonzero vector Z_1 \setminus B_1. (see the align program.)
// 
// If N is zero, the program exhaustively enumerates all sums of
// columns of B and U, and records the minimum weight to min.txt.
// 
// If N > 0, then for N iterations, it picks a random column vector
// from U and adds random vector from the column space of B.  It
// measures the weight of the resulting vector.  Repeatedly doing
// this, it records the minimum found to min.txt, which is overwritten
// each time a new minimum is found.

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading matrix B from %s", args.B)
	var B golsv.BinaryMatrix
	B = golsv.ReadSparseBinaryMatrixFile(args.B)
	log.Printf("done; read %s", B)

	log.Printf("reading matrix U from %s", args.U)
	var U golsv.BinaryMatrix
	U = golsv.ReadSparseBinaryMatrixFile(args.U)
	log.Printf("done; read %s", U)

	log.Printf("computing dense form of B")
	Bdense := B.DenseSubmatrix(0, B.NumRows(), 0, B.NumColumns())
	log.Printf("done; dense form of B is %s", Bdense)

	log.Printf("computing dense form of U")
	Udense := U.DenseSubmatrix(0, U.NumRows(), 0, U.NumColumns())
	log.Printf("done; dense form of U is %s", Udense)

	var minWeight int
	if args.trials <= 0 {
		minWeight = golsv.SystoleExhaustiveSearch(Udense, Bdense, args.verbose)
	} else {
		minWeight = golsv.SystoleRandomSearch(Udense, Bdense, args.trials, args.verbose)
	}
	if args.min != "" {
		writeMinWeight(minWeight, args.min)
	}
	log.Printf("done; minimum nonzero weight found is %d", minWeight)
}

func writeMinWeight(minWeight int, minFile string) {
	// log.Printf("writing min weight to %s", minFile)
	err := os.WriteFile(minFile, []byte(fmt.Sprintf("%d\n", minWeight)), 0644)
	if err != nil {
		panic(err)
	}
}

type Args struct {
	golsv.ProfileArgs
	B string
	U string
	min string
	trials int
	verbose bool
}

func parseFlags() *Args {
	args := Args{
		verbose: true,
		trials: 1000,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.verbose, "verbose", args.verbose, "verbose logging")
	flag.StringVar(&args.B, "B", args.B, "matrix B input file (sparse column support txt format)")
	flag.StringVar(&args.U, "U", args.U, "matrix U input file (sparse column support txt format)")
	flag.StringVar(&args.min, "min", args.min, "minimum weight output file (text)")
	flag.IntVar(&args.trials, "trials", args.trials, "number of trials")
	flag.Parse()
	if args.B == "" {
		flag.Usage()
		log.Fatal("missing required -B flag")
	}
	if args.U == "" {
		flag.Usage()
		log.Fatal("missing required -U flag")
	}
	return &args
}
