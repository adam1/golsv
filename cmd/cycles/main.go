package main

import (
	"flag"
	"golsv"
)

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	lsv := golsv.NewLsvContext(args.BaseField)
	golsv.FindLsvCycles(lsv, args.MinLength, args.MaxLength)
}

type Args struct {
	BaseField string
	MinLength int
	MaxLength int
	golsv.ProfileArgs
}

func parseFlags() *Args {
	args := &Args{
		BaseField: "F16",
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.BaseField, "base-field", "F16", "base field")
	flag.IntVar(&args.MinLength, "min", 3, "minimum cycle length")
	flag.IntVar(&args.MaxLength, "max", 3, "maximum cycle length (0 for no limit)")
	flag.Parse()
	return args
}
