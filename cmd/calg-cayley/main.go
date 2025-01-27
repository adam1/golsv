package main

import (
	"bytes"
	"flag"
	"fmt"
	"golsv"
	"log"
	"os"
	"text/template"
)

// Usage:
//
//	calg-cayley -truncate-generators=0 -max-depth=0 -modulus=111 -quotient
func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	f := golsv.NewF2Polynomial(args.Modulus)
	gens := prepareGenerators(args, f)

	if args.SystolicCandidatesFile != "" {
		computeSystolicCandidates(args, f, gens)
		return
	}

	var observer golsv.CalGObserver
	if args.MeshFile != "" {
		observer = golsv.NewCalGVisualizer(args.MeshFile)
	}
	E := golsv.NewCalGCayleyExpander(gens,
		args.MaxDepth, args.Verbose, &f, args.Quotient,
		observer)
	E.Expand()

	log.Printf("computing complex")
	complex := E.Complex()

	writeComplexFiles(complex, args)
	log.Printf("done")
}

// xxx test?
func computeSystolicCandidates(args *CalGCayleyExpanderArgs, f golsv.F2Polynomial, gens []golsv.ElementCalG) {
	log.Printf("computing systolic candidates")
	// first, expand the complex without quotient to limited depth to
	// find shortest candidates.  this finds the congruence subgroup
	// elements nearest to the identity on the Cayley graph,
	// i.e. those with the shortest reduced word expansion in the
	// generators.  their length is \ell_i in our terminology.
	quotient := false
	maxDepth := args.MaxDepth
	E := golsv.NewCalGCayleyExpander(gens, maxDepth, args.Verbose, &f, quotient, nil)
	E.Expand()
	lifts := E.SystolicCandidateLifts()
	lens := make(map[int]int)
	for _, path := range lifts {
		lens[len(path)]++
	}
	log.Printf("found %d systolic candidate lifts: lengths=%v", len(lifts), lens)
	// log.Printf("xxx lifts=%v", lifts)

	// next, expand the quotient complex to full depth.
	quotient = true
	maxDepth = 0
	E = golsv.NewCalGCayleyExpander(gens, maxDepth, args.Verbose, &f, quotient, nil)
	E.Expand()

	// now, project the candidates to the quotient complex, i.e. take
	// each vertex mod f. by construction, this ought to be a no-op
	// except for the last vertex, but let's verify that to be
	// safe. (xxx verify)
	candidatePaths := make([]golsv.ZPath[golsv.ElementCalG], len(lifts))
	for i, path := range lifts {
		candidatePaths[i] = E.Project(path)
	}
	log.Printf("projected %d systolic candidates", len(candidatePaths))
	// log.Printf("xxx candidatePaths=%v", candidatePaths)

	// convert candidates to column vectors in the edge basis.
	log.Printf("computing complex")
	complex := E.Complex()
	writeComplexFiles(complex, args)

	// xxx despite the comment, this doesn't appear to be de-duping
	log.Printf("converting candidates to edge vectors and deduping")
	vecs := make(map[string]golsv.BinaryVector)
	for _, path := range candidatePaths {
		vec := complex.PathToEdgeVector(path)
		vecs[vec.String()] = vec
	}
	candidatesMatrix := golsv.NewSparseBinaryMatrix(len(complex.EdgeBasis()), 0)
	for _, vec := range vecs {
		candidatesMatrix.AppendColumn(vec.Matrix())
	}
	log.Printf("writing (%v) candidates matrix to %s", candidatesMatrix, args.SystolicCandidatesFile)
	candidatesMatrix.WriteFile(args.SystolicCandidatesFile)
}

func writeComplexFiles(complex *golsv.ZComplex[golsv.ElementCalG], args *CalGCayleyExpanderArgs) {
	vertexBasis := complex.VertexBasis()
	edgeBasis := complex.EdgeBasis()
	triangleBasis := complex.TriangleBasis()

	if args.D1File != "" {
		D1sparse := complex.D1().(*golsv.Sparse)
		log.Printf("writing d1 to %s", args.D1File)
		D1sparse.WriteFile(args.D1File)
	}
	if args.D2File != "" {
		D2sparse := complex.D2().(*golsv.Sparse)
		log.Printf("writing d2 to %s", args.D2File)
		D2sparse.WriteFile(args.D2File)
	}
	if args.VertexBasisFile != "" {
		log.Printf("writing vertex basis to %s", args.VertexBasisFile)
		golsv.WriteStringFile(vertexBasis, args.VertexBasisFile)
	}
	if args.EdgeBasisFile != "" {
		log.Printf("writing edge basis to %s", args.EdgeBasisFile)
		golsv.WriteStringFile(edgeBasis, args.EdgeBasisFile)
	}
	if args.TriangleBasisFile != "" {
		log.Printf("writing triangle basis to %s", args.TriangleBasisFile)
		golsv.WriteStringFile(triangleBasis, args.TriangleBasisFile)
	}
}

func prepareGenerators(args *CalGCayleyExpanderArgs, f golsv.F2Polynomial) []golsv.ElementCalG {
	gens := golsv.CartwrightStegerGenerators()
	a := dumpElements(gens)
	log.Printf("original generators:\n%s", a)
	if args.TruncateGenerators > 0 {
		log.Printf("truncating generators to %d", args.TruncateGenerators)
		gens = gens[:args.TruncateGenerators]
	}
	if args.Quotient {
		log.Printf("reducing generators modulo %v", f)
		for i := range gens {
			gens[i] = gens[i].Modf(f)
		}
	}
	b := dumpElements(gens)
	log.Printf("prepared generators (modified=%v):\n%s", a != b, b)

	if args.GeneratorsLatexFile != "" {
		produceGeneratorsLatexFile(args, gens)
	}
	return gens
}

func dumpElements(els []golsv.ElementCalG) string {
	s := ""
	for i, el := range els {
		if i > 0 {
			s += "\n"
		}
		s += el.String()
	}
	return s
}

func produceGeneratorsLatexFile(args *CalGCayleyExpanderArgs, gens []golsv.ElementCalG) {
	// combine the (algebraic) generators with their matrix representations
	// so as to produce a nice table.
	_, genMatrixReps := golsv.CartwrightStegerGeneratorsMatrixReps()
	log.Printf("Generator matrix reps:")
	combined := make([]genInfo, 0)
	for i, info := range genMatrixReps {
		fmt.Printf("u=%v b_u=%v b_uInv=%v\n", info.U, info.B_u, info.B_uInv)
		combined = append(combined, genInfo{info, gens[i], gens[i+1]})
	}
	const latexTemplate = `\begin{array}{|c|c|c|}
	\hline
	u & b_u, \quad b_u^{-1} & \rho(b_u), \quad \rho(b_u^{-1}\\
	\hline
	{{range .}}
	{{F2PolyLatexWithVar .MInfo.U "v"}} & {{.B_uCalG.Latex}} & {{ProjMatF2PolyLatexWithVar .MInfo.B_u "y"}} \\
	                                        & {{.B_uInvCalG.Latex}} & {{ProjMatF2PolyLatexWithVar .MInfo.B_uInv "y"}} \\
	{{end}}
	\hline
\end{array}
	`
	funcMap := template.FuncMap{
		"F2PolyLatexWithVar": func(p golsv.F2Polynomial, varName string) string {
			return p.Latex(varName)
		},
		"ProjMatF2PolyLatexWithVar": func(m golsv.ProjMatF2Poly, varName string) string {
			return m.Latex(varName)
		},

	}
	tpl, err := template.New("gens").Funcs(funcMap).Parse(latexTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, combined); err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(args.GeneratorsLatexFile, []byte(out.String()), 0644)
	if err != nil {
		panic(err)
	}
	if args.Verbose {
		log.Printf("wrote generator matrix info to %s", args.GeneratorsLatexFile)
	}
}

type genInfo struct {
	MInfo golsv.CartwrightStegerGenMatrixInfo
	B_uCalG, B_uInvCalG golsv.ElementCalG
}

type CalGCayleyExpanderArgs struct {
	D1File                  string
	D2File                  string
	EdgeBasisFile           string
	GeneratorsLatexFile     string
	MaxDepth                int
	MeshFile                string
	Modulus                 string
	Quotient                bool
	SystolicCandidatesFile  string
	TriangleBasisFile       string
	TruncateGenerators      int
	Verbose                 bool
	VertexBasisFile         string
	golsv.ProfileArgs
}

func parseFlags() *CalGCayleyExpanderArgs {
	args := CalGCayleyExpanderArgs{
		MaxDepth:           -1,
		TruncateGenerators: 0,
		Verbose:            true,
		Modulus:            "111",
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.D1File, "d1", args.D1File, "d1 output file (sparse column support txt format)")
	flag.StringVar(&args.D2File, "d2", args.D2File, "d2 output file (sparse column support txt format)")
	flag.StringVar(&args.EdgeBasisFile, "edge-basis", args.EdgeBasisFile, "edge basis output file (text)")
	flag.StringVar(&args.GeneratorsLatexFile, "generators-latex-file", args.GeneratorsLatexFile, "write table of generators to this file (latex)")
	flag.IntVar(&args.MaxDepth, "max-depth", args.MaxDepth, "maximum depth")
	flag.StringVar(&args.MeshFile, "mesh", args.MeshFile, "mesh output file (OFF Object File Format text)")
	flag.StringVar(&args.Modulus, "modulus", args.Modulus, "modulus corresponding to a principle congruence subgroup")
	flag.BoolVar(&args.Quotient, "quotient", args.Quotient, "construct finite quotient complex by first reducing generators modulo the given modulus")
	flag.StringVar(&args.SystolicCandidatesFile, "systolic-candidates", args.SystolicCandidatesFile, "systolic candidates output file (text)")
	flag.StringVar(&args.TriangleBasisFile, "triangle-basis", args.TriangleBasisFile, "triangle basis output file (text)")
	flag.IntVar(&args.TruncateGenerators, "truncate-generators", args.TruncateGenerators, "truncate generators")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.StringVar(&args.VertexBasisFile, "vertex-basis", args.VertexBasisFile, "vertex basis output file (text)")
	flag.Parse()
	return &args
}
