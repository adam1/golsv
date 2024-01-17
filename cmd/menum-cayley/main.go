package main

import (
	"flag"
	"golsv"
)

func main() {
	var args *golsv.MatrixEnumCayleyExpanderArgs = parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	E := golsv.NewMatrixEnumCayleyExpander(args)
	E.Expand()
}

func parseFlags() *golsv.MatrixEnumCayleyExpanderArgs {
	args := &golsv.MatrixEnumCayleyExpanderArgs{}
	args.ProfileArgs.ConfigureFlags()
	flag.Parse()
	return args
}
