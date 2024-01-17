package golsv

type wordHandler func(word []MatGF, path Path, product MatGF) (Continue bool)

// we enumerate *reduced* words, meaning that we disallow an element
// to be immediately followed by its inverse.
func EnumerateWordsN(lsv *LsvContext, generators []MatGF, length int, f wordHandler) {
	if length == 0 {
		f(nil, nil, *MatGfIdentity)
		return
	}
	word := make([]MatGF, length)
	path := make(Path, length)
	enumerateWordsNRec(lsv, generators, word, 0, path, *MatGfIdentity, f)
}

func enumerateWordsNRec(lsv *LsvContext, generators []MatGF, word []MatGF, start int, path Path, cumulativeProduct MatGF, f wordHandler) bool {
	for _, g := range generators {
		word[start] = g
		// skip if g is the inverse of the previous generator
		if start > 0 && g.Equal(lsv, word[start-1].Inverse(lsv)) {
			continue
		}
		product := cumulativeProduct.Multiply(lsv, &g)
		product.MakeCanonical(lsv)
		path[start] = NewEdge(cumulativeProduct, *product)

// 		// xxx need a pruning mechanism here to avoid enumerating a
// 		// given branch.
// 		if prune(g, product

		if start == len(word)-1 {
			if !f(word, path, *product) {
				return false
			}
		} else {
			if !enumerateWordsNRec(lsv, generators, word, start+1, path, *product, f) {
				return false
			}
		}
	}
	return true
}

func FindCycles(lsv *LsvContext, generators []MatGF, length int, f wordHandler) {

	handleWord := func(word []MatGF, path Path, product MatGF) (Continue bool) {
		if product.IsIdentity(lsv) {
			Continue = f(word, path, product)
			if !Continue {
				return false
			}
		}
		return true
	}
	EnumerateWordsN(lsv, generators, length, handleWord)
}

// note: there may be duplicates in the returned list
// xxx re-test
func FindLsvCycles(lsv *LsvContext, minLength int, maxLength int) []Path {
	length := minLength
	totalCycles := 0
	paths := make([]Path, 0)
	gens := lsv.Generators()
	for {
		cyclesThisLength := 0
		handleCycle := func(cycle []MatGF, path Path, product MatGF) (Continue bool) {
			cyclesThisLength++
			totalCycles++
			// log.Printf("cycle: %v", path)
			paths = append(paths, path.Copy())
			return true
		}
		FindCycles(lsv, gens, length, handleCycle)
		// log.Printf("found length %d cycles: %d; total cycles: %d", length, cyclesThisLength, totalCycles)
		length++
		if length > maxLength {
			break
		}
	}
	return paths
}

func LsvTrianglesAtOrigin(lsv *LsvContext) []Triangle {
	triangles := make(map[Triangle]any)
	handleCycle := func(cycle []MatGF, path Path, product MatGF) (Continue bool) {
		triangle := path.Triangle(lsv)
		triangles[triangle] = nil
		return true
	}
	gens := lsv.Generators()
	FindCycles(lsv, gens, 3, handleCycle)
	ret := make([]Triangle, len(triangles))
	i := 0
	for t, _ := range triangles {
		ret[i] = t
		i++
	}
	return ret
}

func CyclicSubgroupPath(lsv *LsvContext, generator *MatGF) Path {
	// slightly hacky; we use the Cayley expander to find the length
	// of the cycle, and then the cycle finder to actually obtain the
	// path.  this is a workaround for the fact that the cayley
	// expander doesn't currently actually model paths.

	//	args := &CayleyExpanderBfArgs{}
	//args.TruncateGenerators = 1

	E := NewCayleyExpander(lsv, []MatGF{*generator}, -1, false)

	E.Expand()
	length := E.NumVertices()

	var ret Path
	handleCycle := func(cycle []MatGF, path Path, product MatGF) (Continue bool) {
		ret = path
		// xxx this appears to be the only place where we make use of
		// the boolean Continue return value
		return false
	}
	FindCycles(lsv, []MatGF{*generator}, length, handleCycle)
	return ret
}
