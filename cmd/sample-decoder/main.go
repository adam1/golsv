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
//   - decoder type (coboundary or boundary)
//   - the Z_1 generators matrix (for coboundary decoder)
//   - error weight range
//   - number of samples per weight
//   - [OR] a file containing the errors to decode
//
// Output:
//
//   - (log) decoding success samples
//   - (optional) results json file
//
type DecoderSamplerArgs struct {
	D1File           string
	D2File           string
	Z_1File          string
	DecoderType      string  // "coboundary" or "boundary"
	Verbose          bool
	ErrorWeight      int
	SamplesPerWeight int
	ErrorsFile       string
	ResultsFile	     string
	FindThreshold    bool
	MinErrorWeight   int
	MaxErrorWeight   int
	golsv.ProfileArgs
}

func parseFlags() DecoderSamplerArgs {
	args := DecoderSamplerArgs{
		DecoderType:      "coboundary",
		ErrorWeight:      10,
		SamplesPerWeight: 1,
		MinErrorWeight:   0,
		MaxErrorWeight:   -1, // -1 means use code length
		Verbose:          true,
		FindThreshold:    true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", "", "D1 matrix file")
	flag.StringVar(&args.D2File, "d2", "", "D2 matrix file")
	flag.StringVar(&args.Z_1File, "Z_1", "", "Z_1 matrix file (required for coboundary decoder)")
	flag.StringVar(&args.DecoderType, "decoder", args.DecoderType, "Decoder type: 'coboundary' or 'boundary'")
	flag.IntVar(&args.ErrorWeight, "error-weight", args.ErrorWeight, "Error weight (for single weight sampling)")
	flag.IntVar(&args.SamplesPerWeight, "samples-per-weight", args.SamplesPerWeight, "Number of samples per weight")
	flag.StringVar(&args.ErrorsFile, "errors", args.ErrorsFile, "Errors file. Decode these errors instead of sampling.")
	flag.StringVar(&args.ResultsFile, "results", "", "Results file")
	flag.BoolVar(&args.FindThreshold, "find-threshold", args.FindThreshold, "Find error weight threshold using binary search")
	flag.IntVar(&args.MinErrorWeight, "min-error-weight", args.MinErrorWeight, "Minimum error weight for threshold search")
	flag.IntVar(&args.MaxErrorWeight, "max-error-weight", args.MaxErrorWeight, "Maximum error weight for threshold search (-1 = code length)")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "Verbose output")
	flag.Parse()

	ok := true
	if args.D1File == "" {
		fmt.Println("D1 matrix file is required")
		ok = false
	}
	if args.DecoderType != "coboundary" && args.DecoderType != "boundary" {
		fmt.Println("Decoder type must be 'coboundary' or 'boundary'")
		ok = false
	}
	if args.DecoderType == "coboundary" {
		if args.D2File == "" {
			fmt.Println("D2 matrix file is required for coboundary decoder")
			ok = false
		}
		if args.Z_1File == "" {
			fmt.Println("Z_1 matrix file is required for coboundary decoder")
			ok = false
		}
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

	var decoder golsv.Decoder

	if args.DecoderType == "coboundary" {
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
		decoder = golsv.NewCoboundaryDecoder(complex, Z_1, args.Verbose)
	} else {
		// For boundary decoder, we only need D1
		complex := golsv.NewZComplexFromBoundaryMatrices(D1, nil)
		decoder = golsv.NewSSFBoundaryDecoder(complex, args.Verbose)
	}

	if args.ErrorsFile != "" {
		decodeErrorsFromFile(decoder, args)
	} else if args.FindThreshold {
		maxErrorWeight := args.MaxErrorWeight
		if maxErrorWeight == -1 {
			maxErrorWeight = decoder.Length()
		}
		finder := golsv.NewDecoderThresholdFinder(decoder, args.MinErrorWeight, maxErrorWeight, args.SamplesPerWeight, args.Verbose)
		results := finder.FindThreshold()
		log.Printf("Threshold search results: threshold=%d, maxSuccess=%d, minFailure=%d, steps=%d, samples=%d", 
			results.ThresholdWeight, results.MaxSuccessWeight, results.MinFailureWeight, results.SearchSteps, results.TotalSamples)
	} else {
		sampler := golsv.NewDecoderSampler(decoder, args.ErrorWeight, args.SamplesPerWeight, args.ResultsFile, args.Verbose, false)
		sampler.Run()
	}
}

func decodeErrorsFromFile(decoder golsv.Decoder, args DecoderSamplerArgs) {
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
