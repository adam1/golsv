package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
	"golsv"
)

// Usage:
//
//   systole -trials N -B B.txt -U U.txt -min min.txt
//
// The program reads BinaryMatrix B and U in sparse column support
// format.  First, it checks the weight of each column of U.  The
// minimum weight is an upper bound for the minimum weight of a
// nonzero vector Z_1 \setminus B_1. (see the align program.)
// 
// Then, for N iterations, it picks a random column vector from U and
// adds random vector from the column space of B.  It measures the
// weight of the resulting vector.  Repeatedly doing this, it records
// the minimum found to min.txt, which is overwritten each time a new
// minimum is found.
//
// xxx add option for complete enumeration; then run at size small
// enough to run in reasonable time: 2^(k + m) where k = dim H_1 and m = dim B_1
// (= 2^t where t = dim Z_1)

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading matrix B from %s", args.B)
	var B golsv.BinaryMatrix
	B = golsv.ReadSparseBinaryMatrixFile(args.B)
	log.Printf("done; read %s", B)

	log.Printf("reading matrix U from %s", args.U)
	var U golsv.BinaryMatrix
	U = golsv.ReadSparseBinaryMatrixFile(args.U)
	log.Printf("done; read %s", U)

	log.Printf("computing dense form of B")
	Bdense := B.DenseSubmatrix(0, B.NumRows(), 0, B.NumColumns())
	log.Printf("done; dense form of B is %s", Bdense)

	log.Printf("computing dense form of U")
	Udense := U.DenseSubmatrix(0, U.NumRows(), 0, U.NumColumns())
	log.Printf("done; dense form of U is %s", Udense)

	log.Printf("sampling minimum nonzero weight for %d trials", args.trials)

	minWeight := math.MaxInt
	for j := 0; j < U.NumColumns(); j++ {
		weight := Udense.ColumnWeight(j)
		if weight == 0 {
			panic(fmt.Sprintf("column %d of U weight is zero", j))
		}
		log.Printf("column %d has weight %d", j, weight)
		if weight < minWeight {
			minWeight = weight
			log.Printf("new min weight: %d", minWeight)
			if args.min != "" {
				writeMinWeight(minWeight, args.min)
			}
		}
	}
	reportInterval := 10
	timeStart := time.Now()
	timeLast := timeStart
		
	for n := 0; n < args.trials; n++ {
		// pick a random column of U
		j := rand.Intn(U.NumColumns())
		a := Udense.DenseSubmatrix(0, Udense.NumRows(), j, j+1)
		// pick a random vector in the column space of B
		b := golsv.RandomLinearCombination(Bdense)
		// add b to a
		a.AddMatrix(b)
		// measure the weight
		weight := a.ColumnWeight(0)
		if weight == 0 {
			panic(fmt.Sprintf("random linear combination of B and U is zero"))
		}
		if weight < minWeight {
			minWeight = weight
			log.Printf("new min weight: %d", minWeight)
			if args.min != "" {
				writeMinWeight(minWeight, args.min)
			}
		}
		if n > 0 && n % reportInterval == 0 {
			timeNow := time.Now()
			timeElapsed := timeNow.Sub(timeStart)
			timeInterval := timeNow.Sub(timeLast)
			timeLast = timeNow
			log.Printf("trial %d/%d (%.2f%%) minwt=%d crate=%.2f trate=%.2f",
				n, args.trials, 100.0*float64(n)/float64(args.trials), minWeight,
				float64(reportInterval)/timeInterval.Seconds(),
				float64(n)/timeElapsed.Seconds())
		}
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
	B string
	U string
	min string
	trials int
	verbose bool
}

func parseFlags() *Args {
	args := Args{
		verbose: true,
		trials: 1000,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.verbose, "verbose", args.verbose, "verbose logging")
	flag.StringVar(&args.B, "B", args.B, "matrix B input file (sparse column support txt format)")
	flag.StringVar(&args.U, "U", args.U, "matrix U input file (sparse column support txt format)")
	flag.StringVar(&args.min, "min", args.min, "minimum weight output file (text)")
	flag.IntVar(&args.trials, "trials", args.trials, "number of trials")
	flag.Parse()
	if args.B == "" {
		flag.Usage()
		log.Fatal("missing required -B flag")
	}
	if args.U == "" {
		flag.Usage()
		log.Fatal("missing required -U flag")
	}
	return &args
}
