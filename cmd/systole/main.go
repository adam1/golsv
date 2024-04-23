package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"golsv"
)

// Usage 1 (given matrices B and U (see systole.go for their meaning):
//
//   systole -trials N -B B.txt -U U.txt -systole S1.txt
//
// Usage 2 (given boundary matrices D1 and D2:
//
//   systole -d1 D1.txt -d2 D2.txt -systole S1.txt -cosystole S^1.txt
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

	if args.D1File != "" {
		log.Printf("reading matrix d_1 from %s", args.D1File)
		var D1 golsv.BinaryMatrix
		D1 = golsv.ReadSparseBinaryMatrixFile(args.D1File)
		log.Printf("done; read %s", D1)

		log.Printf("reading matrix d_2 from %s", args.D2File)
		var D2 golsv.BinaryMatrix
		D2 = golsv.ReadSparseBinaryMatrixFile(args.D2File)
		log.Printf("done; read %s", D2)

		systole := golsv.ComputeFirstSystole(D1, D2, args.verbose)
		log.Printf("systole=%d", systole)
		if args.SystoleFile != "" {
			writeMinWeight(systole, args.SystoleFile)
		}
		cosystole := golsv.ComputeFirstCosystole(D1, D2, args.verbose)
		log.Printf("cosystole=%d", systole)
		if args.CosystoleFile != "" {
			writeMinWeight(cosystole, args.CosystoleFile)
		}
		return
	}
	log.Printf("reading matrix B from %s", args.BFile)
	var B golsv.BinaryMatrix
	B = golsv.ReadSparseBinaryMatrixFile(args.BFile)
	log.Printf("done; read %s", B)

	log.Printf("reading matrix U from %s", args.UFile)
	var U golsv.BinaryMatrix
	U = golsv.ReadSparseBinaryMatrixFile(args.UFile)
	log.Printf("done; read %s", U)

	log.Printf("computing dense form of B")
	Bdense := B.DenseSubmatrix(0, B.NumRows(), 0, B.NumColumns())
	log.Printf("done; dense form of B is %s", Bdense)

	log.Printf("computing dense form of U")
	Udense := U.DenseSubmatrix(0, U.NumRows(), 0, U.NumColumns())
	log.Printf("done; dense form of U is %s", Udense)

	var minWeight int
	if args.Trials <= 0 {
		minWeight = golsv.SystoleExhaustiveSearch(Udense, Bdense, args.verbose)
	} else {
		minWeight = golsv.SystoleRandomSearch(Udense, Bdense, args.Trials, args.verbose)
	}
	if args.SystoleFile != "" {
		writeMinWeight(minWeight, args.SystoleFile)
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
	BFile string
	UFile string
	D1File string
	D2File string
	SystoleFile string
	CosystoleFile string
	Trials int
	verbose bool
}

func parseFlags() *Args {
	args := Args{
		verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.verbose, "verbose", args.verbose, "verbose logging")
	flag.StringVar(&args.BFile, "B", args.BFile, "matrix B input file (sparse column support txt format)")
	flag.StringVar(&args.UFile, "U", args.UFile, "matrix U input file (sparse column support txt format)")
	flag.StringVar(&args.D1File, "d1", args.D1File, "boundary matrix d_1 input file (sparse column support txt format)")
	flag.StringVar(&args.D2File, "d2", args.D2File, "boundary matrix d_2 input file (sparse column support txt format)")
	flag.StringVar(&args.SystoleFile, "systole", args.SystoleFile, "systole output file (text)")
	flag.StringVar(&args.CosystoleFile, "cosystole", args.CosystoleFile, "cosystole output file (text)")
	flag.IntVar(&args.Trials, "trials", args.Trials, "number of samples of minimum weight search (0=exhaustive search)")
	flag.Parse()
	if (args.D1File != "" && args.D2File == "") || (args.D1File == "" && args.D2File != "") {
		flag.Usage()
		log.Fatal("Use either both -d1 and -d2 flags or both -B and -U flags")
	}
	if (args.BFile != "" && args.UFile == "") || (args.BFile == "" && args.UFile != "") {
		flag.Usage()
		log.Fatal("Use either both -d1 and -d2 flags or both -B and -U flags")
	}
	if args.D1File != "" && args.BFile != "" {
		flag.Usage()
		log.Fatal("Use either -d1 and -d2 flags or -B and -U flags, not both")
	}
	if args.D1File == "" && args.BFile == "" {
		flag.Usage()
		log.Fatal("Use either -d1 and -d2 flags or -B and -U flags")
	}
	return &args
}
