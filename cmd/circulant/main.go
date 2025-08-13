package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
	"strconv"
	"strings"
)

// Usage:
//
//   circulant -d1 d1.txt -d2 d2.txt -n 100 -steps "1,2,5"
//
func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	complex, err := golsv.CirculantComplex(args.N, args.Steps, args.Verbose)
	if err != nil {
		log.Fatalf("Error generating circulant complex: %v", err)
	}

	d_1, d_2 := complex.D1(), complex.D2()

	if args.Verbose {
		log.Printf("Generated circulant complex: n=%d, steps=%v", args.N, args.Steps)
		log.Printf("Complex dimensions: d_1=%dx%d, d_2=%dx%d", 
			d_1.NumRows(), d_1.NumColumns(), d_2.NumRows(), d_2.NumColumns())
	}

	if args.D1File != "" {
		if args.Verbose {
			log.Printf("Writing d1 to %s", args.D1File)
		}
		d_1.Sparse().WriteFile(args.D1File)
	}
	if args.D2File != "" {
		if args.Verbose {
			log.Printf("Writing d2 to %s", args.D2File)
		}
		d_2.Sparse().WriteFile(args.D2File)
	}
}

type Args struct {
	D1File      string
	D2File      string
	N           int
	Steps       []int
	StepsString string
	Verbose     bool
	golsv.ProfileArgs
}

func parseFlags() *Args {
	args := Args{
		N:       10,
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", args.D1File, "d1 output file (sparse column support txt format)")
	flag.StringVar(&args.D2File, "d2", args.D2File, "d2 output file (sparse column support txt format)")
	flag.IntVar(&args.N, "n", args.N, fmt.Sprintf("number of vertices (default %d)", args.N))
	flag.StringVar(&args.StepsString, "steps", args.StepsString, "comma-separated list of step sizes (e.g., \"1,2,5\")")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.Parse()

	if args.StepsString == "" {
		log.Fatal("Error: -steps argument is required (e.g., -steps \"1,2,5\")")
	}

	steps, err := parseSteps(args.StepsString)
	if err != nil {
		log.Fatalf("Error parsing steps: %v", err)
	}
	args.Steps = steps

	return &args
}

func parseSteps(stepsString string) ([]int, error) {
	parts := strings.Split(stepsString, ",")
	steps := make([]int, len(parts))
	
	for i, part := range parts {
		step, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid step '%s': %v", part, err)
		}
		if step <= 0 {
			return nil, fmt.Errorf("step sizes must be positive, got %d", step)
		}
		steps[i] = step
	}
	
	return steps, nil
}