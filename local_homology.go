package golsv

// xxx test; needed?
// func CycleIsInD2Image(alpha Path) bool {
// 	expander := NewTriangleExpander(alpha, -1)
// 	expander.Expand()
// 	relevantTriangles := expander.Triangles()
// 	Y := ComplexFromTriangles(relevantTriangles, false)
// 	log.Printf("Y: %v", Y)
// 	if Y.D2ImageContainsPath(alpha) {
// 		log.Printf("d_2|Y image contains alpha")
// 		return true
// 	} else {
// 		log.Printf("d_2|Y image does not contain alpha")
// 		log.Printf("systole is at most %d", len(alpha))
// 		return false
// 	}
// }

// xxx test; deprecate?
// func CycleIsInD2ImageGraded(lsv *LsvContext, alpha Path, sortBases bool, verbose bool) bool {
// 	for maxDepth := 0; maxDepth < 1000; maxDepth++ {
// 		expander := NewTriangleExpander(lsv, alpha, maxDepth, verbose)
// 		exhausted := expander.Expand()
// 		relevantTriangles := expander.Triangles()
// 		Y := ComplexFromTriangles(relevantTriangles, sortBases, verbose)
// 		if verbose {
// 			log.Printf("Y[grade %d]: %v", maxDepth, Y)
// 			log.Printf("exhausted: %v", exhausted)
// 		}
// 		// log.Printf("bases: %v", Y.DumpBases())
// 		if Y.D2ImageContainsPath(alpha) {
// 			if verbose {
// 				log.Printf("d_2|Y[%d] image contains alpha", maxDepth)
// 			}
// 			return true
// 		} else {
// 			if verbose {
// 				log.Printf("d_2|Y[%d] image does not contain alpha", maxDepth)
// 			}
// 			if exhausted {
// 				if verbose {
// 					log.Printf("systole is at most %d", len(alpha))
// 				}
// 				return false
// 			}
// 		}
// 	}
// 	return false
// }

// xxx test; needed?
// func ComplexFromGradedTriangles(alpha Path, maxDepth int, sortBases bool) *Complex {
// 	expander := NewTriangleExpander(alpha, maxDepth)
// 	expander.Expand()
// 	relevantTriangles := expander.Triangles()
// 	return ComplexFromTriangles(relevantTriangles, sortBases)
// }

// xxx test
func FindTriangleContainingEdge(lsv *LsvContext, T []Triangle, e Edge) *Triangle {
	for _, t := range T {
		if g, found := t.OrbitContainsEdge(lsv, e); found {
			translation := t.Translate(lsv, g)
			return &translation
		}
	}
	return nil
}

func AllTrianglesContainingEdge(lsv *LsvContext, T []Triangle, e Edge) []Triangle {
	triangles := make(map[Triangle]any)
	for _, t := range T {
		if g, found := t.OrbitContainsEdge(lsv, e); found {
			translation := t.Translate(lsv, g)
			triangles[translation] = nil
		}
	}
	// log.Printf("found %d triangles containing edge %v", len(triangles), e)
	result := make([]Triangle, len(triangles))
	i := 0
	for t, _ := range triangles {
		result[i] = t
		i++
	}
	return result
}

func AllTrianglesSharingEdgeWithTriangle(lsv *LsvContext, T []Triangle, t Triangle) []Triangle {
	triangles := make(map[Triangle]any)
	for _, e := range t.Edges() {
		for _, u := range AllTrianglesContainingEdge(lsv, T, e) {
			if t.Equal(lsv, &u) {
				continue
			}
			triangles[u] = nil
		}
	}
	result := make([]Triangle, len(triangles))
	i := 0
	for u, _ := range triangles {
		result[i] = u
		i++
	}
	return result
}
