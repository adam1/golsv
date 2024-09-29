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
//
// Output:
//
//   - (log) decoding success samples
//
type CoboundaryDecoderArgs struct {
	D1File           string
	D2File           string
	Verbose          bool
	MinErrorWeight   int
	MaxErrorWeight   int
	SamplesPerWeight int
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
	flag.IntVar(&args.MinErrorWeight, "min-error-weight", args.MinErrorWeight, "Minimum error weight")
	flag.IntVar(&args.MaxErrorWeight, "max-error-weight", args.MaxErrorWeight, "Maximum error weight")
	flag.IntVar(&args.SamplesPerWeight, "samples-per-weight", args.SamplesPerWeight, "Number of samples per weight")
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
	complex := golsv.NewZComplexFromBoundaryMatrices(D1, D2)

	decoder := golsv.NewCoboundaryDecoder(complex, args.Verbose)

	sampler := golsv.NewDecoderSampler(decoder, args.MinErrorWeight, args.MaxErrorWeight, args.SamplesPerWeight, args.Verbose)
	sampler.Run()
}

