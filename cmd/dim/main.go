package main

import (
	"flag"
	"fmt"
	"log"
	"golsv"
)

// Usage:
//
//   dim -kernel -in smith.txt
//   dim -image -in smith.txt
//
// The program reads a BinaryMatrix from smith.txt in sparse column
// support format, checks that it is in Smith normal form, and then
// computes the dimension of the kernel or image.
func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	log.Printf("reading matrix from %s", args.In)
	M := golsv.ReadSparseBinaryMatrixFile(args.In).(*golsv.Sparse)
	log.Printf("done; read %s", M)
	if args.Dump {
		log.Printf("matrix:\n%s", golsv.DumpMatrix(M))
	}
	dimDomain := M.NumColumns()
	dimCodomain := M.NumRows()
	dimKernel := -1
	dimImage := -1
	log.Printf("domain=%d codomain=%d", dimDomain, dimCodomain)
	
	smith, rank := M.IsSmithNormalForm()
	if smith {
		log.Printf("matrix is in Smith normal form; rank=%d", rank)
		dimKernel = M.NumColumns() - rank
		dimImage = rank
		log.Printf("kernel=%d image=%d", dimKernel, dimImage)
	} else {
		log.Printf("matrix is not in Smith normal form")
		if args.Kernel || args.Image {
			panic("cannot continue")
		}
	}
	if args.CheckColsNonzero {
		found := 0
		for j := 0; j < M.NumColumns(); j++ {
			if M.ColumnWeight(j) == 0 {
				log.Printf("column %d is zero", j)
				found++
			}
		}
		if found > 0 {
			panic(fmt.Sprintf("found %d zero columns", found))
		}
		log.Printf("all columns are nonzero")
	}
	if args.Kernel {
		fmt.Printf("%d\n", dimKernel)
	} else if args.Image {
		fmt.Printf("%d\n", dimImage)
	} else if args.Domain {
		fmt.Printf("%d\n", dimDomain)
	} else if args.Codomain {
		fmt.Printf("%d\n", dimCodomain)
	} else if args.Density {
		fmt.Printf("%f\n", M.Density(0, 0))
	}
}

type Args struct {
	golsv.ProfileArgs
	In string
	Codomain bool
	Domain bool
	Image bool
	Kernel bool
	Density bool
	Verbose bool
	Dump bool
	CheckColsNonzero bool
}

func parseFlags() *Args {
	args := Args{
		Verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.BoolVar(&args.Codomain, "codomain", args.Codomain, "print dimension of codomain")
	flag.BoolVar(&args.Kernel, "kernel", args.Kernel, "compute dimension of kernel")
	flag.BoolVar(&args.Domain, "domain", args.Domain, "print dimension of domain")
	flag.BoolVar(&args.Image, "image", args.Image, "compute dimension of image")
	flag.StringVar(&args.In, "in", args.In, "matrix input file (sparse column support txt format)")
	flag.BoolVar(&args.Dump, "dump", args.Dump, "dump matrix to stderr")
	flag.BoolVar(&args.Density, "density", args.Density, "print density of matrix")
	flag.BoolVar(&args.CheckColsNonzero, "check-cols-nonzero", args.CheckColsNonzero, "check that all columns are nonzero")
	flag.Parse()
	if args.In == "" {
		flag.Usage()
		log.Fatal("missing required -in flag")
	}
	return &args
}
