package main

import (
	"fmt"
)

// Goal:
//
// Given an edge vector from an LSV complex, interpreted as a cochain,
// apply the [KKL] algorithm to find a locally minimal cochain
// representing the same cohomology class.
//
// Input:
//
//   - the boundary matrices defining the LSV complex
//   - a list of edge vectors
//
// Output:
//
//   - a list of edge vectors (locally minimal versions of the input)
//
// The primary example we are interested in is locally minimal
// versions of generators for the cohomology group Z^1.  In
// particular, we want to answer these questions:
//
//  - Given different generators for a particular cohomology class, do
//    they have the same locally minimal representative?
//
//  - Do different cohomology classes have locally minimal
//    representatives of different lengths?
//
//  - Is there a geometric interpretation of the sequence of boundary
//    cochains added by the algorithm to find the locally minimal
//    representative?




func main() {
// 	args := parseFlags()
// 	args.ProfileArgs.Start()
// 	defer args.ProfileArgs.Stop()

	fmt.Printf("Hello, world\n")
}
