package main

import (
	"flag"
	"fmt"
	"golsv"
	"log"
	"os"
	"strconv"
	"strings"
)

type MatrixArgs struct {
	InputFile         string
	OutputFile        string
	InputDebugFormat  bool
	Verbose           bool
	Dump              bool
	RowWeights        bool
	ColWeights        bool
	ProjectionRowMask map[int]any
	golsv.ProfileArgs
}

func parseFlags() MatrixArgs {
	args := MatrixArgs{
		Verbose: true,
		ProjectionRowMask: make(map[int]any),
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.InputFile, "in", "", "Input file")
	flag.StringVar(&args.OutputFile, "out", "", "Output file")
	flag.BoolVar(&args.InputDebugFormat, "input-debug-format", false, "Input file is in debug format instead of sparse format")
	flag.BoolVar(&args.Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&args.RowWeights, "row-weights", false, "Show row weights")
	flag.BoolVar(&args.ColWeights, "col-weights", false, "Show column weights")

	var rowIndicesStr string
	flag.StringVar(&rowIndicesStr, "project-rows", "", "Project rows to the given indices (comma-separated)")
	args.ProjectionRowMask = stringToIntSet(rowIndicesStr)

	flag.BoolVar(&args.Dump, "dump", false, "Dump the matrix to stdout")
	flag.Parse()
	if args.InputFile == "" {
		fmt.Fprintln(os.Stderr, "missing required input file")
		flag.Usage()
		os.Exit(1)
	}
	return args
}

func stringToIntSet(s string) map[int]any {
	set := make(map[int]any)
	if s == "" {
		return set
	}
	for _, str := range strings.Split(s, ",") {
		i, err := strconv.Atoi(str)
		if err != nil {
			log.Fatalf("error parsing integer: %v", err)
		}
		set[i] = nil
	}
	return set
}

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	if args.Verbose {
		log.Printf("reading matrix from %s", args.InputFile)
	}
	var M golsv.BinaryMatrix
	if args.InputDebugFormat {
		M = readDebugMatrix(args.InputFile)
	} else {
		M = golsv.ReadSparseBinaryMatrixFile(args.InputFile)
	}
	if args.Verbose {
		log.Printf("matrix: %s", M)
	}
	if len(args.ProjectionRowMask) > 0 {
		if args.Verbose {
			log.Printf("projecting rows")
		}
		M = M.Project(func(i int) bool {
			_, ok := args.ProjectionRowMask[i]
			return ok
		})
		if args.Verbose {
			log.Printf("projected matrix: %s", M)
		}
	}
	if args.Dump {
		fmt.Println(golsv.DumpMatrix(M))
	}
	if args.RowWeights {
		for i := 0; i < M.NumRows(); i++ {
			w := 0
			for j := 0; j < M.NumColumns(); j++ {
				if M.Get(i, j) == 1 {
					w++
				}
			}
			fmt.Printf("row %d: %d\n", i, w)
		}
	}
	if args.ColWeights {
		for j := 0; j < M.NumColumns(); j++ {
			w := M.ColumnWeight(j)
			fmt.Printf("col %d: %d\n", j, w)
		}
	}
	if args.OutputFile != "" {
		if args.Verbose {
			log.Printf("writing matrix to %s", args.OutputFile)
		}
		M.(*golsv.Sparse).WriteFile(args.OutputFile)
	}
}

func readDebugMatrix(filename string) golsv.BinaryMatrix {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading file %s: %v", filename, err)
	}
	return golsv.NewSparseBinaryMatrixFromString(string(content))
}

