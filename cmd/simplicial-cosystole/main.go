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
//   - (optional) the U^1 (aka Uc1) matrix from the UB decomposition of the cochain complex
//
// Output:
//
//   - (log) the cosystole (weight)
//
type SimplicialCosystoleArgs struct {
	D1File            string
	D2File            string
	Z1File            string
	Zu1File           string
	B1File            string
	Bu1File           string
	Uu1File           string
	VertexBasisFile   string
	EdgeBasisFile     string
	TriangleBasisFile string
	DimUudSeq		  bool
	CohomologyOrbits  bool
	Modulus           string
	CheckUu1		  bool
	golsv.CosystoleSearchParams
	needD1, needD2, needZ1, needZu1, needB1, needBu1, needUu1,
	needVertexBasis, needEdgeBasis, needTriangleBasis,
	needModulus bool
	golsv.ProfileArgs
}

func parseFlags() SimplicialCosystoleArgs {
	args := SimplicialCosystoleArgs{}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", "", "D1 matrix file")
	flag.StringVar(&args.D2File, "d2", "", "D2 matrix file")
	flag.StringVar(&args.Z1File, "Z1", "", "Z1 matrix file")
	flag.StringVar(&args.Zu1File, "Z^1", "", "Z^1 matrix file")
	flag.StringVar(&args.B1File, "B1", "", "B1 matrix file")
	flag.StringVar(&args.Bu1File, "B^1", "", "B^1 matrix file")
	flag.StringVar(&args.Uu1File, "U^1", "", "U^1 matrix file")
	flag.StringVar(&args.VertexBasisFile, "vertex-basis", "", "Vertex basis file")
	flag.StringVar(&args.EdgeBasisFile, "edge-basis", "", "Edge basis file")
	flag.StringVar(&args.TriangleBasisFile, "triangle-basis", "", "Triangle basis file")
	flag.BoolVar(&args.PruneByCohomologyProjection, "prune-by-cohomology-projection", false, "Incrementally prune by cohomology projection (default: false)")
	flag.BoolVar(&args.InitialSupport, "initial-support", false, "Require cocycles to be supported on first triangle (default: false)")
	flag.BoolVar(&args.DimUudSeq, "dim-uud-seq", false, "Compute the sequence of dimension Uud (instead of the cosystole)")
	flag.BoolVar(&args.CohomologyOrbits, "cohomology-orbits", false, "Compute the cohomology orbits of the cosystole (instead of the cosystole)")
	flag.StringVar(&args.Modulus, "modulus", "", "F2Polynomial modulus for ElementCalG")
	flag.BoolVar(&args.CheckUu1, "check-U^1", false, "Sanity check U^1 matrix")
	flag.BoolVar(&args.Verbose, "verbose", false, "Verbose output")
	flag.Parse()

	if args.DimUudSeq {
		args.needD1 = true
		args.needD2 = true
		args.needZ1 = true
		args.needZu1 = true
		args.needBu1 = true
		args.needUu1 = true
	} else if args.CohomologyOrbits {
		args.needD1 = true
		args.needD2 = true
		args.needZ1 = true
		args.needB1 = true
		args.needUu1 = true
		args.needVertexBasis = true
		args.needEdgeBasis = true
		args.needTriangleBasis = true
		args.needModulus = true
	} else if args.CheckUu1 {
		args.needUu1 = true
		args.needZ1 = true
	} else {
		args.needD1 = true
		args.needD2 = true
		args.needZ1 = true
		args.needZu1 = true
		args.needBu1 = true
	}
	ok := true
	if args.needD1 && args.D1File == "" {
		fmt.Println("D1 matrix file is required")
		ok = false
	}
	if args.needD2 && args.D2File == "" {
		fmt.Println("D2 matrix file is required")
		ok = false
	}
	if args.needZ1 && args.Z1File == "" {
		fmt.Println("Z1 matrix file is required")
		ok = false
	}
	if args.needZu1 && args.Zu1File == "" {
		fmt.Println("Z^1 matrix file is required")
		ok = false
	}
	if args.needB1 && args.B1File == "" {
		fmt.Println("B1 matrix file is required")
		ok = false
	}
	if args.needBu1 && args.Bu1File == "" {
		fmt.Println("B^1 matrix file is required")
		ok = false
	}
	if args.needUu1 && args.Uu1File == "" {
		fmt.Println("U^1 matrix file is required")
		ok = false
	}
	if args.needVertexBasis && args.VertexBasisFile == "" {
		fmt.Println("Vertex basis file is required")
		ok = false
	}
	if args.needEdgeBasis && args.EdgeBasisFile == "" {
		fmt.Println("Edge basis file is required")
		ok = false
	}
	if args.needTriangleBasis && args.TriangleBasisFile == "" {
		fmt.Println("Triangle basis file is required")
		ok = false
	}
	if args.needModulus && args.Modulus == "" {
		fmt.Println("Modulus is required")
		ok = false
	}
	if !ok {
		flag.Usage()
		panic("missing required arguments")
	}
	return args
}

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	var D1, D2, Z1, Zu1, B1, Bu1, Uu1 golsv.BinaryMatrix

	if args.needD1 {
		if args.Verbose {
			log.Printf("reading matrix d_1 from %s", args.D1File)
		}
		D1 = golsv.ReadSparseBinaryMatrixFile(args.D1File)
		if args.Verbose {
			log.Printf("done; read %s", D1)
		}
	}
	if args.needD2 {
		if args.Verbose {
			log.Printf("reading matrix d_2 from %s", args.D2File)
		}
		D2 = golsv.ReadSparseBinaryMatrixFile(args.D2File)
		if args.Verbose {
			log.Printf("done; read %s", D2)
		}
	}
	if args.needZ1 {
		if args.Verbose {
			log.Printf("reading matrix Z_1 from %s", args.Z1File)
		}
		Z1 = golsv.ReadSparseBinaryMatrixFile(args.Z1File)
		if args.Verbose {
			log.Printf("done; read %s", Z1)
		}
	}
	if args.needZu1 {
		if args.Verbose {
			log.Printf("reading matrix Z^1 from %s", args.Zu1File)
		}
		Zu1 = golsv.ReadSparseBinaryMatrixFile(args.Zu1File)
		if args.Verbose {
			log.Printf("done; read %s", Zu1)
		}
	}
	if args.needB1 {
		if args.Verbose {
			log.Printf("reading matrix B1 from %s", args.B1File)
		}
		B1 = golsv.ReadSparseBinaryMatrixFile(args.B1File)
		if args.Verbose {
			log.Printf("done; read %s", B1)
		}
	}
	if args.needBu1 {
		if args.Verbose {
			log.Printf("reading matrix B^1 from %s", args.Bu1File)
		}
		Bu1 = golsv.ReadSparseBinaryMatrixFile(args.Bu1File)
		if args.Verbose {
			log.Printf("done; read %s", Bu1)
		}
	}
	if args.needUu1 {
		if args.Verbose {
			log.Printf("reading matrix U^1 from %s", args.Uu1File)
		}
		Uu1 = golsv.ReadSparseBinaryMatrixFile(args.Uu1File)
		if args.Verbose {
			log.Printf("done; read %s", Uu1)
		}
	}
	if args.CohomologyOrbits {
		complex := golsv.NewZComplexElementCalGFromBasisFiles(
			args.VertexBasisFile, args.EdgeBasisFile, args.TriangleBasisFile,
			args.Verbose)
		golsv.ComputeCohomologyOrbits(complex, Uu1, Z1, B1,
			golsv.NewF2Polynomial(args.Modulus), args.Verbose)
	} else if args.DimUudSeq {
		complex := golsv.NewZComplexFromBoundaryMatrices(D1, D2)
		golsv.ComputeDimUudBudSequence(complex, Uu1, Bu1, args.Verbose)
	} else if args.CheckUu1 {
		golsv.CheckUu1(Uu1, Z1)
	} else {
		complex := golsv.NewZComplexFromBoundaryMatrices(D1, D2)
		S := golsv.NewCosystoleSearch(complex, Z1, Zu1, Bu1, args.CosystoleSearchParams)
		cosys := S.Cosystole()
		log.Printf("cosystole: %d", cosys)
	}
}

