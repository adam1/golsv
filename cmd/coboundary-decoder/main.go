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
//   - the Z_1 generators matrix
//   - error weight range
//   - number of samples per weight
//   - [OR] a file containing the errors to decode
//
// Output:
//
//   - (log) decoding success samples
//   - (optional) results json file
//
type CoboundaryDecoderArgs struct {
	D1File           string
	D2File           string
	Z_1File          string
	Verbose          bool
	MinErrorWeight   int
	MaxErrorWeight   int
	SamplesPerWeight int
	ErrorsFile       string
	ResultsFile	     string
	golsv.ProfileArgs
}

func parseFlags() CoboundaryDecoderArgs {
	args := CoboundaryDecoderArgs{
		MinErrorWeight: 0,
		MaxErrorWeight: 100,
		SamplesPerWeight: 100,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", "", "D1 matrix file")
	flag.StringVar(&args.D2File, "d2", "", "D2 matrix file")
	flag.StringVar(&args.Z_1File, "Z_1", "", "Z_1 matrix file")
	flag.IntVar(&args.MinErrorWeight, "min-error-weight", args.MinErrorWeight, "Minimum error weight")
	flag.IntVar(&args.MaxErrorWeight, "max-error-weight", args.MaxErrorWeight, "Maximum error weight")
	flag.IntVar(&args.SamplesPerWeight, "samples-per-weight", args.SamplesPerWeight, "Number of samples per weight")
	flag.StringVar(&args.ErrorsFile, "errors", args.ResultsFile, "Errors file. Decode these errors instead of sampling.")
	flag.StringVar(&args.ResultsFile, "results", "", "Results file")
	flag.BoolVar(&args.Verbose, "verbose", false, "Verbose output")
	flag.Parse()

	ok := true
	if args.D1File == "" {
		fmt.Println("D1 matrix file is required")
		ok = false
	}
	if args.D2File == "" {
		fmt.Println("D2 matrix file is required")
		ok = false
	}
	if args.Z_1File == "" {
		fmt.Println("Z_1 matrix file is required")
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

	var D1, D2 golsv.BinaryMatrix

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
	if args.Verbose {
		log.Printf("reading matrix Z_1 from %s", args.Z_1File)
	}
	Z_1 := golsv.ReadSparseBinaryMatrixFile(args.Z_1File)
	if args.Verbose {
		log.Printf("done; read %s", Z_1)
	}
	complex := golsv.NewZComplexFromBoundaryMatrices(D1, D2)

	decoder := golsv.NewCoboundaryDecoder(complex, Z_1, args.Verbose)

	if args.ErrorsFile != "" {
		decodeErrorsFromFile(decoder, args)
	} else {
		sampler := golsv.NewDecoderSampler(decoder, args.MinErrorWeight, args.MaxErrorWeight, args.SamplesPerWeight, args.ResultsFile, args.Verbose)
		sampler.Run()
	}
}

func decodeErrorsFromFile[T any](decoder *golsv.CoboundaryDecoder[T], args CoboundaryDecoderArgs) {
	errors := golsv.ReadSparseBinaryMatrixFile(args.ErrorsFile)
	log.Printf("read %s: %s", args.ErrorsFile, errors)
	for j := 0; j < errors.NumColumns(); j++ {
		errorVec := errors.ColumnVector(j)
		syndrome := decoder.Syndrome(errorVec)
		log.Printf("decoding column %d error weight=%d syndrome weight=%d", j, errorVec.Weight(), syndrome.Weight())
		err, _ := decoder.Decode(syndrome)
		if err != nil {
			log.Printf("error decoding column %d: %v", j, err)
		} else {
			log.Printf("decoded column %d", j)
		}
	}
}
