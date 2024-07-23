package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"golsv"
)

// Usage:
//
//   weight -in U.txt -col
//
// The program reads BinaryMatrix U from U.txt in sparse column
// support format, computes the weight of each column, and the minimum
// weight of all columns.
//

type Args struct {
	golsv.ProfileArgs
	in string
	row bool
	col bool 
	verbose bool
}

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading matrix U from %s", args.in)
	var U golsv.BinaryMatrix
	U = golsv.ReadSparseBinaryMatrixFile(args.in)
	log.Printf("done; read %s", U)

	minWeight := math.MaxInt64
	if args.row {
		log.Printf("computing row weights")
		UT := U.Transpose()
		for i := 0; i < UT.NumColumns(); i++ {
			w := UT.ColumnWeight(i)
			fmt.Printf("row %d weight %d\n", i, w)
			if w < minWeight {
				minWeight = w
			}
		}
		log.Printf("min row weight: %d", minWeight)
	} else {
		log.Printf("computing column weights")
		for j := 0; j < U.NumColumns(); j++ {
			w := U.ColumnWeight(j)
			fmt.Printf("column %d weight %d\n", j, w)
			if w < minWeight {
				minWeight = w
			}
		}
		log.Printf("min column weight: %d", minWeight)
	}
}

func parseFlags() *Args {
	args := Args{
		verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.verbose, "verbose", args.verbose, "verbose logging")
	flag.StringVar(&args.in, "in", args.in, "matrix input file (sparse column support txt format)")
	flag.BoolVar(&args.row, "row", args.row, "compute row weight")
	flag.BoolVar(&args.col, "col", args.row, "compute col weight")
	flag.Parse()
	if args.in == "" {
		flag.Usage()
		log.Fatal("missing required -in flag")
	}
	if args.row && args.col {
		flag.Usage()
		log.Fatal("cannot specify both -row and -col")
	}
	if !args.row && !args.col {
		flag.Usage()
		log.Fatal("must specify either -row or -col")
	}
	return &args
}
