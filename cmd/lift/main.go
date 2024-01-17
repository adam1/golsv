package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log"
	"math"
	"os"

	"golsv"
)

// Usage:
//
//   lift -B B.txt -U U.txt -edge-basis ebasis.txt
//
// The program reads column vectors from U.txt, which by previous
// preparations are known to be in Z_1 \setminus B_1, and thus
// potentially systolic. It chooses a column vector v with minimum
// weight.
//
// Vector v represents a path alpha in the graph (which must be a
// cycle), via the edge basis. The program resolves v to alpha by
// index in the edge basis, producing a Path.  Then for each edge in
// the path it computes the generator corresponding to the edge,
// accumulating a product of generators, which ultimately be equal to
// the identity when multiplied out, i.e. the word is a relation in
// the group.

func main() {
	args := parseFlags()
	args.ProfileArgs.Start()
	defer args.ProfileArgs.Stop()

	lsv := golsv.NewLsvContext(args.BaseField)

	log.Printf("reading matrix U from %s", args.U)
	var U golsv.BinaryMatrix
	U = golsv.ReadSparseBinaryMatrixFile(args.U)
	log.Printf("done; read %s", U)

	log.Printf("reading vertex basis from %s", args.vertexBasis)
	vertexBasis := golsv.ReadVertexFile(args.vertexBasis)
	log.Printf("done; read %d vertices", len(vertexBasis))

	log.Printf("reading edge basis from %s", args.edgeBasis)
	edgeBasis := golsv.ReadEdgeFile(args.edgeBasis)
	log.Printf("done; read %d edges", len(edgeBasis))

	// note: we don't need the triangle basis, so we don't read it
	sortBases := false
	C := golsv.NewComplex(vertexBasis, edgeBasis, nil, sortBases, false)
	log.Printf("complex C has %d vertices and %d edges", C.NumVertices(), C.NumEdges())

	v := columnWithMinWeight(U).(*golsv.Sparse)
	weight := v.ColumnWeight(0)
	log.Printf("found column with min weight %d", weight)

	alpha := C.EdgeVectorToPath(lsv, v)
	log.Printf("alpha=%s", alpha)

	var word, vertices []golsv.MatGF
	word, vertices = alpha.Word(lsv)
	// log.Printf("word=%s", word)
	// print indices for generators in the word
	wordIds := ""
	for _, g := range word {
		id, ok := C.VertexIndex[g]
		if !ok {
			panic(fmt.Sprintf("vertex %s not found in index", g))
		}
		wordIds += fmt.Sprintf("%d\n", id)
	}
	log.Printf("word (vertex basis indices):\n%s", wordIds)

	// verify that the path is a cycle by multiplying out the matrix
	// product
	product := golsv.Product(lsv, word)
	if !product.IsIdentity(lsv) {
		panic(fmt.Sprintf("word product is not identity!!! product=%s", product))
	}

	if args.latexWord {
		fmt.Printf("$$\n")
		for i, g := range word {
			if i > 0 && i % 4 == 0 {
				fmt.Printf("$$\n$$\n")
			}
			fmt.Printf("%s", g.Latex(lsv))
		}
		fmt.Printf("$$\n")
	} else if args.latexVertices {
		fmt.Printf("$$\n")
		for i, v := range vertices {
			if i > 0 && i % 4 == 0 {
				fmt.Printf("$$\n$$\n")
			}
			fmt.Printf("%s", v.Latex(lsv))
		}
		fmt.Printf("$$\n")
	} else if args.animate {
		animatePath(lsv, word)
	} else if args.origin {
		log.Printf("reading matrix B from %s", args.B)
		var B golsv.BinaryMatrix
		B = golsv.ReadSparseBinaryMatrixFile(args.B)
		log.Printf("done; read %s", B)

		// find U \cap P where P is the set of cycles at the origin.
		log.Printf("computing UcapP")
		var UcapP golsv.BinaryMatrix
		UcapP = columnsAtOrigin(U, edgeBasis, lsv)
		log.Printf("UcapP = %s", UcapP)

		// similarly find B \cap P.
		log.Printf("computing BcapP")
		var BcapP golsv.BinaryMatrix
		BcapP = columnsAtOrigin(B, edgeBasis, lsv)
		log.Printf("BcapP = %s", BcapP)
	}
}

func columnWithMinWeight(U golsv.BinaryMatrix) golsv.BinaryMatrix {
	minWeight := math.MaxInt
	column := -1
	for j := 0; j < U.NumColumns(); j++ {
		weight := U.(*golsv.Sparse).ColumnWeight(j)
		if weight < minWeight {
			minWeight = weight
			column = j
		}
	}
	return U.Submatrix(0, U.NumRows(), column, column+1)
}

func columnsAtOrigin(M golsv.BinaryMatrix, edgeBasis []golsv.Edge, lsv *golsv.LsvContext) golsv.BinaryMatrix {
	N := golsv.NewSparseBinaryMatrix(M.NumRows(), 0)
	for j := 0; j < M.NumColumns(); j++ {
		if columnContainsOrigin(M, j, edgeBasis, lsv) {
			w := M.Submatrix(0, M.NumRows(), j, j+1)
			N.AppendColumn(w)
		}
	}
	return N
}

func columnContainsOrigin(M golsv.BinaryMatrix, j int, edgeBasis []golsv.Edge, lsv *golsv.LsvContext) bool {
	rows := M.(*golsv.Sparse).Support(j)
	for _, row := range rows {
		edge := edgeBasis[row]
		if edge.ContainsOrigin(lsv) {
			return true
		}
	}
	return false
}

func animatePath(lsv *golsv.LsvContext, word []golsv.MatGF) {
	vertices := golsv.CumulativeProducts(lsv, word)

	var images []*image.Paletted
	var delays []int

	for _, v := range vertices {
		img := drawMatrix(lsv, v)
		images = append(images, img)
		delays = append(delays, 15)
	}
	// xxx
	f, err := os.OpenFile("rgb.gif", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}

func drawMatrix(lsv *golsv.LsvContext, m golsv.MatGF) *image.Paletted {
	log.Printf("drawing matrix %s", m)
	w := 201
	h := w
	t := w / 3

	var palette = []color.Color{
		color.RGBA{0xff, 0xff, 0xff, 0xff}, // white
		color.RGBA{0x00, 0x00, 0x00, 0xff}, // black
		color.RGBA{0xff, 0x00, 0x00, 0xff}, // red
		color.RGBA{0x00, 0xff, 0xff, 0xff}, // cyan
	}
	img := image.NewPaletted(image.Rect(0, 0, w, h), palette)
	// xxx
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			poly := m.At(lsv, i, j)
			log.Printf("poly=%s", poly)
			color := polyToColorIndex(poly)
			log.Printf("color=%d", color)
			for x := i*t; x < (i+1)*t; x++ {
				for y := j*t; y < (j+1)*t; y++ {
					img.SetColorIndex(x, y, color)
				}
			}
		}
	}
	return img
}

func polyToColorIndex(poly string) uint8 {
	switch poly {
	case "0":
		return 0
	case "1":
		return 1
	case "x":
		return 2
	case "x + 1":
		return 3
	}
	panic(fmt.Sprintf("unknown poly %s", poly))
}

type Args struct {
	golsv.ProfileArgs
	BaseField string
	B string
	U string
	edgeBasis string
	vertexBasis string
	animate bool
	latexWord bool
	latexVertices bool
	origin bool
	verbose bool
}

func parseFlags() *Args {
	args := Args{
		verbose: true,
	}
	args.ProfileArgs.ConfigureFlags()
	flag.StringVar(&args.BaseField, "base-field", args.BaseField, "base field: F4 or F16")
	flag.BoolVar(&args.verbose, "verbose", args.verbose, "verbose logging")
	flag.StringVar(&args.B, "B", args.B, "matrix B input file (sparse column support txt format)")
	flag.StringVar(&args.U, "U", args.U, "matrix U input file (sparse column support txt format)")
	flag.StringVar(&args.edgeBasis, "edge-basis", args.edgeBasis, "edge basis input file (text)")
	flag.StringVar(&args.vertexBasis, "vertex-basis", args.vertexBasis, "vertex basis input file (text)")
	flag.BoolVar(&args.animate, "animate", args.animate, "animate the path")
	flag.BoolVar(&args.latexWord, "latex-word", args.latexWord, "output word in latex format")
	flag.BoolVar(&args.latexVertices, "latex-vertices", args.latexVertices, "output vertices in latex format")
	flag.BoolVar(&args.origin, "origin", args.origin, "reduce to origin")
	flag.Parse()
	if args.BaseField == "" {
		flag.Usage()
		log.Fatal("missing required -base-field flag")
	}
	if args.U == "" {
		flag.Usage()
		log.Fatal("missing required -U flag")
	}
	if args.edgeBasis == "" {
		flag.Usage()
		log.Fatal("missing required -edge-basis flag")
	}
	if args.vertexBasis == "" {
		flag.Usage()
		log.Fatal("missing required -vertex-basis flag")
	}
	if args.origin {
		if args.B == "" {
			flag.Usage()
			log.Fatal("missing required -B flag")
		}
	}
	return &args
}
