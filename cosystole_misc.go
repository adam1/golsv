package golsv

import (
	"fmt"
	"log"
)

func ComputeDimUudBudSequence[T any] (complex *ZComplex[T], Uu1 BinaryMatrix, Bu1 BinaryMatrix, verbose bool) {
	if verbose {
		log.Printf("computing sequence dim U^d")
	}
	dimUu1 := Uu1.NumColumns()
	reachedDimUu1 := false
	seenEdgeIndices := make(map[int]struct{})
	for d, t := range complex.TriangleBasis() {
		d = d + 1 // naming convention
		for _, e := range t.Edges() {
			if k, ok := complex.edgeIndex[e]; ok {
				seenEdgeIndices[k] = struct{}{}
			} else {
				panic(fmt.Sprintf("edge %v not in edge index", e))
			}
		}
		if !reachedDimUu1 {
			Ucd := Uu1.(*Sparse).Project(func(i int) bool {
				_, ok := seenEdgeIndices[i]
				return ok
			})
			_, _, _, rank := smithNormalForm(Ucd, verbose)
			log.Printf("U^%d %v rank: %d", d, Ucd, rank)
			if rank == dimUu1 {
				reachedDimUu1 = true
			}
		}
		Bcd := Bu1.(*Sparse).Project(func(i int) bool {
			_, ok := seenEdgeIndices[i]
			return ok
		})
		_, _, _, rank := smithNormalForm(Bcd, verbose)
		log.Printf("B^%d %v rank: %d", d, Bcd, rank)
	}
}

// xxx experimental wip
func ComputeCohomologyOrbits(complex *ZComplex[ElementCalG], Uu1 BinaryMatrix, Z_1 BinaryMatrix, B_1 BinaryMatrix, modulus F2Polynomial, verbose bool) {
	if verbose {
		log.Printf("computing cohomology orbits")
	}
	Z_1TDense := Z_1.Transpose().Dense()
	vertices := complex.VertexBasis()
	for j := 0; j < Uu1.NumColumns(); j++ {
		orbit := make(map[int]struct{})
		u := Uu1.ColumnVector(j)
		uWeight := u.Weight()
		if verbose {
			log.Printf("computing orbit of u_%d (U^1 column %d) (weight: %d)", j, j, uWeight)
		}
		for k, g := range vertices {
			v := groupActionOnCochainSpace(complex, g.(ElementCalG), modulus, u)
			vWeight := v.Weight()
			if uWeight != vWeight {
				panic(fmt.Sprintf("uWeight %d != vWeight %d", uWeight, vWeight))
			}
			m := cosetRep(Uu1, Z_1TDense, B_1, v)
			orbit[m] = struct{}{}
			log.Printf("after group element %d: cumulative orbit size: %d", k, len(orbit))
		}
		log.Printf("orbit length of U^1 column %d: %d", j, len(orbit))
	}
}

// xxx test
func groupActionOnCochainSpace (complex *ZComplex[ElementCalG], g ElementCalG, modulus F2Polynomial, u BinaryVector) BinaryVector {
	v := NewBinaryVector(len(u))
	edgeBasis := complex.EdgeBasis()
	for i := 0; i < len(u); i++ {
		if u[i] == 0 {
			continue
		}
		e := edgeBasis[i]
		f := groupActionOnEdge(e, g, modulus)
		k, ok := complex.edgeIndex[f]
		if !ok {
			panic(fmt.Sprintf("edge %v not in edge index", f))
		}
		// log.Printf("edge %d -> %d", i, k)
		v[k] = 1
	}
	return v
}

// xxx test
func groupActionOnEdge (edge ZEdge[ElementCalG], g ElementCalG, modulus F2Polynomial) ZEdge[ElementCalG] {
	var a, b ElementCalG
	var ok bool
	if a, ok = edge[0].(ElementCalG); !ok {
		panic(fmt.Sprintf("edge[0] %v not an ElementCalG", edge[0]))
	}
	if b, ok = edge[1].(ElementCalG); !ok {
		panic(fmt.Sprintf("edge[1] %v not an ElementCalG", edge[1]))
	}
	var ga, gb ElementCalG
	ga.Mul(g, a)
	ga = ga.Modf(modulus)

	gb.Mul(g, b)
	gb = gb.Modf(modulus)
	return NewZEdge[ElementCalG](ga, gb)
}


func cosetRep(Uu1 BinaryMatrix, Z_1T *DenseBinaryMatrix, B_1 BinaryMatrix, v BinaryVector) int {
	Uu1Dense := Uu1.Dense()
	m := -1
	EnumerateBinaryVectorSpace(Uu1Dense, func(u BinaryMatrix, index int) (ok bool) {
		// u is dense here
		u.Add(v.Matrix())
		if isCoboundary2(u.Sparse(), Z_1T) {
			m = index
			log.Printf("xxx cosetRep match found; orbit with index: %d", m)
			return false
		}
		log.Printf("xxx cosetRep did not match orbit with index: %d", index)
		return true
	})
	if m < 0 {
		panic("cosetRep not found")
	}
	return m


// 	// xxx sanity check that diff is in Z^1
// 	for j := 0; j < Uu1.NumColumns(); j++ {
// 		u := Uu1.ColumnVector(j)
// 		diff := u.Add(v)
// 		if !isCocycle(diff.Matrix(), B_1) {
// 			panic(fmt.Sprintf("diff not in Z^1: %v", diff))
// 		}
// 	}
// 	log.Printf("xxx all diffs are in Z^1")
// 	// xxx could it be in B^1?
// 	if isCoboundary(v.Matrix(), Z_1) {
// 		panic("v is a coboundary!")
// 	}
// 	log.Printf("xxx v is not a coboundary")
// 	// xxx could it be a linear combination of columns of U^1?
// 	if isInColumnSpan(v, Uu1) {
// 		panic("v is in the column span of U^1")
// 	}
// 	log.Printf("xxx v is not in the column span of U^1")
	panic("cosetRep not found")
}

func isInColumnSpan(v BinaryVector, Uu1 BinaryMatrix) bool {
	Uu1T := Uu1.Transpose()
	verbose := true
	K := kernelBasis(Uu1T, verbose).Sparse()
	vTd := v.Matrix().Transpose().Dense()
	P := vTd.MultiplyRight(K)
	return P.IsZero()
}

func CheckUu1(Uu1, Z_1 BinaryMatrix) {
	// check that the difference of every column pair from U^1 is not
	// a coboundary
	for i := 0; i < Uu1.NumColumns(); i++ {
		for j := i + 1; j < Uu1.NumColumns(); j++ {
			log.Printf("Checking column pair (%d,%d)", i, j)
			diff := Uu1.ColumnVector(i).Add(Uu1.ColumnVector(j))
			if isCoboundary(diff.Matrix(), Z_1) {
				panic(fmt.Sprintf("U^1 columns %d and %d are in same coset of B^1", i, j))
			}
		}
	}
	log.Printf("Check passed: no column pair from U^1 is in the same coset of B^1.")
}
