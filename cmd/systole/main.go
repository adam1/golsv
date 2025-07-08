package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
	"os"
)

// Usage 1: given matrices B and U (see systole.go for their meaning)
//
//   systole -trials N -B B.txt -U U.txt -systole S1.txt
//
// Usage 2: (given boundary matrices D1 and D2:
//
//   systole -d1 D1.txt -d2 D2.txt -systole S1.txt -cosystole S^1.txt
//
// For usage 1 and 2,
// the program reads BinaryMatrix B and U in sparse column support
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
//
// Usage 3: simplicial systole search at given vertex
//
//   systole -simplicial-at-vertex 0 -d1 D1.txt -d2 D2.txt -systole S1.txt
//
// Usage 4: simplicial systole search (global)
//
//   systole -simplicial -d1 D1.txt -d2 D2.txt -systole S1.txt
//

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	if args.UFile != "" {
		doSystoleSearchFromUB(args)
	} else if args.SimplicialAtVertex >= 0 {
		doSimplicialSystoleSearchAtVertex(args)
	} else if args.Simplicial {
		doSimplicialSystoleSearch(args)
	} else {
		doExhaustiveSystoleAndCosystoleSearchFromComplex(args)
	}
}

func doExhaustiveSystoleAndCosystoleSearchFromComplex(args *Args) {
	D1, D2 := readBoundaryMatrices(args)
	systole, _, _, _ := golsv.ComputeFirstSystole(D1, D2, args.Verbose)
	log.Printf("systole=%d", systole)
	if args.SystoleFile != "" {
		writeIntegerFile(systole, args.SystoleFile)
	}
	cosystole := golsv.ComputeFirstCosystole(D1, D2, args.Verbose)
	log.Printf("cosystole=%d", cosystole)
	if args.CosystoleFile != "" {
		writeIntegerFile(cosystole, args.CosystoleFile)
	}
}

func doSimplicialSystoleSearch(args *Args) {
	X := complexFromBoundaryMatrices(args)
	S := golsv.NewSimplicialSystoleSearch(X, args.Verbose)
	weight := S.Search()
	if args.SystoleFile != "" {
		writeIntegerFile(weight, args.SystoleFile)
	}
	log.Printf("done; minimum weight found is %d", weight)
}

func doSimplicialSystoleSearchAtVertex(args *Args) {
	X := complexFromBoundaryMatrices(args)
	S := golsv.NewSimplicialSystoleSearch(X, args.Verbose)
	weight := S.SearchAtVertex(X.VertexBasis()[args.SimplicialAtVertex])
	if args.SystoleFile != "" {
		writeIntegerFile(weight, args.SystoleFile)
	}
	log.Printf("done; minimum weight found is %d", weight)
}

func complexFromBoundaryMatrices(args *Args) *golsv.ZComplex[golsv.ZVertexInt] {
	D1, D2 := readBoundaryMatrices(args)
	return golsv.NewZComplexFromBoundaryMatrices(D1, D2)
}

func readBoundaryMatrices(args *Args) (D1, D2 golsv.BinaryMatrix) {
	if args.Verbose {
		log.Printf("reading matrix d_1 from %s", args.D1File)
	}
	D1 = golsv.ReadSparseBinaryMatrixFile(args.D1File)
	if args.Verbose {
		log.Printf("done; read %s", D1)
	}
	if args.Verbose {
		log.Printf("reading matrix d_2 from %s", args.D2File)
	}
	D2 = golsv.ReadSparseBinaryMatrixFile(args.D2File)
	if args.Verbose {
		log.Printf("done; read %s", D2)
	}
	return
}

func doSystoleSearchFromUB(args *Args) {
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
		minWeight = golsv.SystoleExhaustiveSearch(Udense, Bdense, args.Verbose)
	} else {
		minWeight = golsv.SystoleRandomSearch(Udense, Bdense, args.Trials, args.Verbose)
	}
	if args.SystoleFile != "" {
		writeIntegerFile(minWeight, args.SystoleFile)
	}
	log.Printf("done; minimum nonzero weight found is %d", minWeight)
}

func writeIntegerFile(n int, path string) {
	// log.Printf("writing min weight to %s", minFile)
	err := os.WriteFile(path, []byte(fmt.Sprintf("%d\n", n)), 0644)
	if err != nil {
		panic(err)
	}
}

type Args struct {
	golsv.ProfileArgs
	BFile              string
	CosystoleFile      string
	D1File             string
	D2File             string
	Simplicial         bool
	SimplicialAtVertex int
	SystoleFile        string
	Trials             int
	UFile              string
	Verbose            bool
}

func parseFlags() *Args {
	args := Args{
		SimplicialAtVertex: -1,
		Verbose:            true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.CosystoleFile, "cosystole", args.CosystoleFile, "cosystole output file (text)")
	flag.StringVar(&args.D1File, "d1", args.D1File, "boundary matrix d_1 input file (sparse column support txt format)")
	flag.StringVar(&args.D2File, "d2", args.D2File, "boundary matrix d_2 input file (sparse column support txt format)")
	flag.StringVar(&args.BFile, "B", args.BFile, "matrix B input file (sparse column support txt format)")
	flag.BoolVar(&args.Simplicial, "simplicial", args.Simplicial, "do simplicial systole search (global)")
	flag.IntVar(&args.SimplicialAtVertex, "simplicial-at-vertex", args.SimplicialAtVertex, "do simplicial systole search starting at given vertex index")
	flag.StringVar(&args.SystoleFile, "systole", args.SystoleFile, "systole output file (text)")
	flag.IntVar(&args.Trials, "trials", args.Trials, "number of samples of minimum weight search (0=exhaustive search)")
	flag.StringVar(&args.UFile, "U", args.UFile, "matrix U input file (sparse column support txt format)")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
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
