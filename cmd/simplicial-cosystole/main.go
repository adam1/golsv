package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
)

// Input:
//
//   - the boundary matrices defining the LSV complex
//   - the Z1 matrix from the UB decomposition
//   - (optional) the U matrix from the UB decomposition
//
// Output:
//
//   - (log) the cosystole (weight)
//
type SimplicialCosystoleArgs struct {
	D1File             string
	D2File             string
	Z1File             string
	UFile              string
	Verbose            bool
	golsv.ProfileArgs
}

func parseFlags() SimplicialCosystoleArgs {
	args := SimplicialCosystoleArgs{}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", "", "D1 matrix file")
	flag.StringVar(&args.D2File, "d2", "", "D2 matrix file")
	flag.StringVar(&args.Z1File, "z1", "", "Z1 matrix file")
	flag.StringVar(&args.UFile, "u", "", "U matrix file")
	flag.BoolVar(&args.Verbose, "verbose", false, "Verbose output")
	flag.Parse()

	if args.D1File == "" {
		fmt.Println("D1 matrix file is required")
		flag.PrintDefaults()
	}
	if args.D2File == "" {
		fmt.Println("D2 matrix file is required")
		flag.PrintDefaults()
	}
	if args.Z1File == "" {
		fmt.Println("Z1 matrix file is required")
		flag.PrintDefaults()
	}
	return args
}

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	if args.Verbose {
		log.Printf("reading matrix d_1 from %s", args.D1File)
	}
	D1 := golsv.ReadSparseBinaryMatrixFile(args.D1File)
	if args.Verbose {
		log.Printf("done; read %s", D1)
	}
	if args.Verbose {
		log.Printf("reading matrix d_2 from %s", args.D2File)
	}
	D2 := golsv.ReadSparseBinaryMatrixFile(args.D2File)
	if args.Verbose {
		log.Printf("done; read %s", D2)
	}
	if args.Verbose {
		log.Printf("reading matrix z_1 from %s", args.Z1File)
	}
	Z1 := golsv.ReadSparseBinaryMatrixFile(args.Z1File)
	if args.Verbose {
		log.Printf("done; read %s", Z1)
	}
	var U golsv.BinaryMatrix
	if args.UFile != "" {
		if args.Verbose {
			log.Printf("reading matrix u from %s", args.UFile)
		}
		U = golsv.ReadSparseBinaryMatrixFile(args.UFile)
		if args.Verbose {
			log.Printf("done; read %s", U)
		}
	}
	C := golsv.NewZComplexFromBoundaryMatrices(D1, D2)
	S := golsv.NewCosystoleSearch(C, args.Verbose)
	if args.UFile != "" {
		S.Prefilter(U)
	}
	cosys := S.Cosystole(Z1)
	log.Printf("cosystole: %d", cosys)
}
