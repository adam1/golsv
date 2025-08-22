package main

import (
	"bytes"
	"flag"
	"golsv"
	"log"
	"os"
	"text/template"
)

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	f := golsv.NewF2Polynomial(args.Modulus)
	gens := prepareGenerators(args, f)

	if args.SystolicCandidatesFile != "" {
		handleSystolicCandidatesMode(args, f, gens)
	} else if args.Graph {
		handleGraphMode(args, f, gens)
	} else if args.FillTriangles {
		handleFillTrianglesMode(args, f, gens)
	}
}

func handleGraphMode(args *CalGCayleyExpanderArgs, f golsv.F2Polynomial, gens []golsv.ElementCalG) {
	var observer golsv.CalGObserver
	if args.MeshFile != "" {
		observer = golsv.NewCalGVisualizer(args.MeshFile)
	}
	E := golsv.NewCalGCayleyExpander(gens,
		args.MaxDepth, args.Verbose, &f, args.Quotient,
		observer, args.PslFile != "")
	graph := E.Graph()
	writeGraphFiles(graph, args)
	if args.PslFile != "" {
		pslElements := E.PslGenerators()
		log.Printf("writing %d PSL elements to %s", len(pslElements), args.PslFile)
		golsv.WriteStringFile(pslElements, args.PslFile)
	}
	log.Printf("done")
}

func handleFillTrianglesMode(args *CalGCayleyExpanderArgs, f golsv.F2Polynomial, gens []golsv.ElementCalG) {
	if args.Verbose {
		log.Printf("reading vertex basis file %s", args.VertexBasisFile)
	}
	vertexBasis := golsv.ReadElementCalGVertexFile(args.VertexBasisFile)
	if args.Verbose {
		log.Printf("reading edge basis file %s", args.EdgeBasisFile)
	}
	edgeBasis := golsv.ReadElementCalGEdgeFile(args.EdgeBasisFile)

	edgeChecks := false
	F := golsv.NewCalGTriangleFiller(vertexBasis, edgeBasis, gens, args.Verbose, &f, args.Quotient, edgeChecks)
	complex := F.Complex()
	log.Printf("complex: %s", complex)

	if args.TriangleBasisFile != "" {
		log.Printf("writing triangle basis to %s", args.TriangleBasisFile)
		golsv.WriteStringFile(complex.TriangleBasis(), args.TriangleBasisFile)
	}
	if args.D2File != "" {
		D2sparse := complex.D2().(*golsv.Sparse)
		log.Printf("writing d2 to %s", args.D2File)
		D2sparse.WriteFile(args.D2File)
	}
	log.Printf("done")
}

func handleSystolicCandidatesMode(args *CalGCayleyExpanderArgs, f golsv.F2Polynomial, gens []golsv.ElementCalG) {
	log.Printf("computing systolic candidates")
	// first, expand the complex without quotient to limited depth to
	// find shortest candidates.  this finds the congruence subgroup
	// elements nearest to the identity on the Cayley graph,
	// i.e. those with the shortest reduced word expansion in the
	// generators.  their length is \ell_i in our terminology.
	quotient := false
	maxDepth := args.MaxDepth
	E := golsv.NewCalGCayleyExpander(gens, maxDepth, args.Verbose, &f, quotient, nil, args.PslFile != "")
	E.Graph()
	lifts := E.SystolicCandidateLifts()
	lens := make(map[int]int)
	for _, path := range lifts {
		lens[len(path)]++
	}
	if args.Verbose {
		log.Printf("found %d systolic candidate lifts: lengths=%v", len(lifts), lens)
		log.Printf("reading vertex basis file %s", args.VertexBasisFile)
	}
	vertexBasis := golsv.ReadElementCalGVertexFile(args.VertexBasisFile)
	if args.Verbose {
		log.Printf("reading edge basis file %s", args.EdgeBasisFile)
	}
	edgeBasis := golsv.ReadElementCalGEdgeFile(args.EdgeBasisFile)
	resortBases := false
	graph := golsv.NewZComplex(vertexBasis, edgeBasis, nil, resortBases, args.Verbose)

	// now, project the candidates to the quotient complex, i.e. take
	// each vertex mod f. by construction, this ought to be a no-op
	// except for the last vertex. (xxx verify)
	candidatePaths := make([]golsv.ZPath[golsv.ElementCalG], len(lifts))
	for i, path := range lifts {
		candidatePaths[i] = E.Project(path)
	}
	if args.Verbose {
		log.Printf("projected %d systolic candidates", len(candidatePaths))
		log.Printf("converting candidates to edge vectors and deduping")
	}
	vecs := make(map[string]golsv.BinaryVector)
	for _, path := range candidatePaths {
		vec := graph.PathToEdgeVector(path)
		vecs[vec.String()] = vec
	}
	candidatesMatrix := golsv.NewSparseBinaryMatrix(len(graph.EdgeBasis()), 0)
	for _, vec := range vecs {
		candidatesMatrix.AppendColumn(vec.Matrix())
	}
	if args.Verbose {
		log.Printf("writing (%v) candidates matrix to %s", candidatesMatrix, args.SystolicCandidatesFile)
	}
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

func writeGraphFiles(complex *golsv.ZComplex[golsv.ElementCalG], args *CalGCayleyExpanderArgs) {
	vertexBasis := complex.VertexBasis()
	edgeBasis := complex.EdgeBasis()

	if args.D1File != "" {
		D1sparse := complex.D1().(*golsv.Sparse)
		log.Printf("writing d1 to %s", args.D1File)
		D1sparse.WriteFile(args.D1File)
	}
	if args.VertexBasisFile != "" {
		log.Printf("writing vertex basis to %s", args.VertexBasisFile)
		golsv.WriteStringFile(vertexBasis, args.VertexBasisFile)
	}
	if args.EdgeBasisFile != "" {
		log.Printf("writing edge basis to %s", args.EdgeBasisFile)
		golsv.WriteStringFile(edgeBasis, args.EdgeBasisFile)
	}
}

func prepareGenerators(args *CalGCayleyExpanderArgs, f golsv.F2Polynomial) []golsv.ElementCalG {
	if args.Quotient {
		log.Printf("will reduce generators modulo %v", f)
	} else {
		f = golsv.F2PolynomialZero
	}
	genTable := golsv.CartwrightStegerGeneratorsWithMatrixReps(f)
	gens := make([]golsv.ElementCalG, 0)
	for _, inf := range genTable {
		gens = append(gens, inf.B_u, inf.B_uInv)
	}
		
	//log.Printf("prepared generators:\n%s", gens)
	if args.Determinant {
		printGeneratorsDeterminants(args, f, genTable)
	}
	if args.GeneratorsLatexFile != "" {
		produceGeneratorsLatexFile(args, genTable)
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

func printGeneratorsDeterminants(args *CalGCayleyExpanderArgs, f golsv.F2Polynomial, genTable []golsv.CartwrightStegerGenInfo) {
	for _, inf := range genTable {
		var b_uRepDet golsv.F2Polynomial
		var b_uInvRepDet golsv.F2Polynomial
		var yxModulus golsv.F2Polynomial
		if args.Quotient {
			yxModulus = golsv.CartwrightStegerEmbedPolynomial(f)
		}
		if args.Quotient {
			b_uRepDet = inf.B_uRep.Determinant().Modf(yxModulus)
		} else {
			b_uRepDet = inf.B_uRep.Determinant()
		}
		if args.Quotient {
			b_uInvRepDet = inf.B_uInvRep.Determinant().Modf(yxModulus)
		} else {
			b_uInvRepDet = inf.B_uInvRep.Determinant()
		}
		log.Printf("u = %s:", inf.U.String())
		log.Printf("  b_u = %s", inf.B_u.String())
		log.Printf("  matrix rep: %s", inf.B_uRep.String())
		log.Printf("  determinant: %s", b_uRepDet.String())
		log.Printf("  b_u^{-1} = %s", inf.B_uInv.String())
		log.Printf("  matrix rep: %s", inf.B_uInvRep.String())
		log.Printf("  determinant: %s", b_uInvRepDet.String())
		log.Printf("")
	}
}

func produceGeneratorsLatexFile(args *CalGCayleyExpanderArgs, genTable []golsv.CartwrightStegerGenInfo) {
	var latexTemplate string
	latexTemplate = `\begin{array}{|c|c|c|}
	\hline
	u \in \F_4^\times/\F_2^\times & b_u , b_u^{-1} \in {{GenSet}} & \rho(b_u), \rho(b_u^{-1}) \in {{RepSet}}\\
	\hline
	{{range .}}
	{{F2PolyLatexWithVar .U "v"}} & {{.B_u.LatexMatrix}} & {{ProjMatF2PolyLatexWithVar .B_uRep "x"}} \\
	                              & {{.B_uInv.LatexMatrix}} & {{ProjMatF2PolyLatexWithVar .B_uInvRep "x"}} \\
	{{end}}
	\hline
\end{array}`
	funcMap := template.FuncMap{
		"F2PolyLatexWithVar": func(p golsv.F2Polynomial, varName string) string {
			return p.Latex(varName)
		},
		"ProjMatF2PolyLatexWithVar": func(m golsv.ProjMatF2Poly, varName string) string {
			return m.Latex(varName)
		},
		"GenSet": func() string {
			if args.Quotient {
				return `\mathcal{G}(R/I)`
			} else {
				return `\mathcal{G}(R)`
			}
		},
		"RepSet": func() string {
			if args.Quotient {
				return `\PGL_3(\bar{R}/\bar{I})`
			} else {
				return `\PGL_3(\bar{R})`
			}
		},
	}
	tpl, err := template.New("gens").Funcs(funcMap).Parse(latexTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, genTable); err != nil {
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

type CalGCayleyExpanderArgs struct {
	D1File                  string
	D2File                  string
	Determinant             bool
	EdgeBasisFile           string
	FillTriangles           bool
	GeneratorsLatexFile     string
	Graph                   bool
	MaxDepth                int
	MeshFile                string
	Modulus                 string
	PslFile                 string
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
	flag.StringVar(&args.D1File, "d1", args.D1File, "d1 input/output file (sparse column support txt format)")
	flag.StringVar(&args.D2File, "d2", args.D2File, "d2 output file (sparse column support txt format)")
	flag.BoolVar(&args.Determinant, "determinant", args.Determinant, "print matrix representation and determinant of each generator")
	flag.StringVar(&args.EdgeBasisFile, "edge-basis", args.EdgeBasisFile, "edge basis output file (text)")
	flag.BoolVar(&args.FillTriangles, "fill-triangles", args.FillTriangles, "read vertex and edge bases, compute triangle basis and d2.txt")
	flag.StringVar(&args.GeneratorsLatexFile, "generators-latex-file", args.GeneratorsLatexFile, "write table of generators to this file (latex)")
	flag.BoolVar(&args.Graph, "graph", args.Graph, "do Cayley expansion to produce d1.txt")
	flag.IntVar(&args.MaxDepth, "max-depth", args.MaxDepth, "maximum depth")
	flag.StringVar(&args.MeshFile, "mesh", args.MeshFile, "mesh output file (OFF Object File Format text)")
	flag.StringVar(&args.Modulus, "modulus", args.Modulus, "modulus corresponding to a principle congruence subgroup")
	flag.StringVar(&args.PslFile, "psl", args.PslFile, "output file for PSL elements found during Cayley expansion")
	flag.BoolVar(&args.Quotient, "quotient", args.Quotient, "construct finite quotient complex by first reducing generators modulo the given modulus")
	flag.StringVar(&args.SystolicCandidatesFile, "systolic-candidates", args.SystolicCandidatesFile, "systolic candidates output file (text)")
	flag.StringVar(&args.TriangleBasisFile, "triangle-basis", args.TriangleBasisFile, "triangle basis output file (text)")
	flag.IntVar(&args.TruncateGenerators, "truncate-generators", args.TruncateGenerators, "truncate generators")
	flag.BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose logging")
	flag.StringVar(&args.VertexBasisFile, "vertex-basis", args.VertexBasisFile, "vertex basis output file (text)")
	flag.Parse()

	if !args.Graph && !args.FillTriangles && args.SystolicCandidatesFile == "" {
		flag.Usage()
		log.Fatal("Use one of: -graph -fill-triangles -systolic-candidates-file")
	}

	return &args
}
