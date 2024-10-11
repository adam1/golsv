package golsv

import (
	"log"
	"math/rand"
	"reflect"
	"testing"
)

func TestEdgeContains(t *testing.T) {
	trials := 10
	for i := 0; i < trials; i++ {
		a := RandomInvertibleMatGF(lsv)
		b := RandomInvertibleMatGF(lsv)
		e := NewEdge(*a, *b)
		if !e.Contains(lsv, *a) {
			t.Errorf("e should contain a")
		}
		if !e.Contains(lsv, *b) {
			t.Errorf("e should contain b")
		}
		var c *MatGF
		for {
			c = RandomInvertibleMatGF(lsv)
			if c.Equal(lsv, a) || c.Equal(lsv, b) {
				continue
			} else {
				break
			}
		}
		if e.Contains(lsv, *c) {
			t.Errorf("e should not contain c")
		}
	}
}

func TestEdgeContainsOrigin(t *testing.T) {
	trials := 10
	for i := 0; i < trials; i++ {
		a := MatGfIdentity
		b := RandomInvertibleMatGF(lsv)
		if rand.Int() % 2 == 0 {
			a, b = b, a
		}
		e := NewEdge(*a, *b)
		if !e.ContainsOrigin(lsv) {
			t.Errorf("e should contain origin")
		}
	}
}

func TestEdgeEqual(t *testing.T) {
	e := NewEdge(S3Generators[0], S3Generators[1])
	f := NewEdge(S3Generators[0], S3Generators[1])
	if !e.Equal(lsv, f) {
		t.Errorf("e and f should be equal")
	}
}

func TestEdgeEqualIgnoresOrder(t *testing.T) {
	e := NewEdge(S3Generators[0], S3Generators[1])
	f := NewEdge(S3Generators[1], S3Generators[0])
	if !e.Equal(lsv, f) {
		t.Errorf("e and f should be equal")
	}
}

func TestEdgeGenerator(t *testing.T) {
	trials := 10
	for i := 0; i < trials; i++ {
		v := randomVertex()
		h := randomGenerator()
		e :=  NewEdge(v, *v.Multiply(lsv, h))
		g := e.Generator(lsv)
		if e[1].Equal(lsv, &v) {
			g = g.Inverse(lsv)
		}
		if !g.Equal(lsv, h) {
			t.Errorf("g=%v h=%v", g, h)
		}
	}
}

func TestEdgeLess(t *testing.T) {
    tests := []struct {
		e Edge
		f Edge
		want bool
	}{
		{
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
			true,
		},
		{
			NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			false,
		},
		{
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
			true,
		},
	}
	for _, test := range tests {
		got := test.e.Less(test.f)
		if got != test.want {
			t.Errorf("e=%v f=%v got=%v want=%v", test.e, test.f, got, test.want)
		}
	}
}

func TestNewPathFromEdges(t *testing.T) {
	id := MatGfIdentity
	r := &c3Matrix
	rsq := c3Matrix.Multiply(lsv, &c3Matrix)
	p := NewPathFromEdges(lsv, []Edge{
		NewEdge(*id, *r),
		NewEdge(*rsq, *r),
		NewEdge(*r, *rsq),
	})
	// note that edges have been sorted
	want := Path{
		NewEdge(*id, *r),
		NewEdge(*r, *rsq),
		NewEdge(*rsq, *r),
	}
	if !p.Equal(lsv, want) {
		t.Errorf("p=%v want=%v", p, want)
	}
}

func TestNewPathFromEdgesDiscontiguous(t *testing.T) {
	id := MatGfIdentity
	r := &c3Matrix
	rsq := c3Matrix.Multiply(lsv, &c3Matrix)
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("should have panicked")
		}
	} ()
	NewPathFromEdges(lsv, []Edge{
		NewEdge(*id, *r),
		NewEdge(*rsq, *rsq),
	})
}

func TestPathChords(t *testing.T) {
	gens := lsv.Generators()
	g := gens[0]
	g2 := *g.Multiply(lsv, &g)
	g3 := *g.Multiply(lsv, &g2)
	g4 := *g.Multiply(lsv, &g3)
	c32 := c3Matrix.Multiply(lsv, &c3Matrix)

	tests := []struct {
		p Path
		gens []MatGF
		want []Edge
	}{
		{
			Path{
				NewEdge(*MatGfIdentity, g),
			},
			gens,
			[]Edge{},
		},
		{
			Path{
				NewEdge(*MatGfIdentity, g),
				NewEdge(g, g2),
			},
			gens,
			[]Edge{},
		},
		{
			Path{
				NewEdge(*MatGfIdentity, g),
				NewEdge(g, g2),
			},
			append(gens, g2),
			[]Edge{
				NewEdge(*MatGfIdentity, g2),
			},
		},
		{
			Path{
				NewEdge(*MatGfIdentity, g),
				NewEdge(g, g2),
				NewEdge(g2, g3),
			},
			append(gens, g2),
			[]Edge{
				NewEdge(*MatGfIdentity, g2),
				NewEdge(g, g3),
			},
		},
		{
			Path{
				NewEdge(*MatGfIdentity, g),
				NewEdge(g, g2),
				NewEdge(g2, g3),
				NewEdge(g3, g4),
			},
			append(gens, g2),
			[]Edge{
				NewEdge(*MatGfIdentity, g2),
				NewEdge(g, g3),
				NewEdge(g2, g4),
			},
		},
		// cycle
		{
			Path{
				NewEdge(*MatGfIdentity, c3Matrix),
				NewEdge(c3Matrix, *c32),
				NewEdge(*c32, *MatGfIdentity),
			},
			append(gens, *c32),
			[]Edge{},
		},
		// cycle
		{
			Path{
				NewEdge(*MatGfIdentity, g),
				NewEdge(g, g2),
				NewEdge(g2, g3),
				NewEdge(g3, *MatGfIdentity),
			},
			append(gens, g2),
			[]Edge{
				NewEdge(*MatGfIdentity, g2),
				NewEdge(g, g3),
			},
		},
	}
	for n, test := range tests {
		got := test.p.Chords(lsv, test.gens)
		if len(got) != len(test.want) {
			t.Fatalf("test=%d p=%v len(got)=%d len(want)=%d", n, test.p, len(got), len(test.want))
		}
		for i := 0; i < len(got); i++ {
			if !got[i].Equal(lsv, test.want[i]) {
				t.Errorf("p=%v got=%v want=%v", test.p, got, test.want)
			}
		}
	}
}

func TestPathEqual(t *testing.T) {
	gens := lsv.Generators()
	p := Path{
		NewEdge(gens[0], gens[1]),
		NewEdge(gens[1], gens[2]),
	}
	q := Path{
		NewEdge(gens[0], gens[1]),
		NewEdge(gens[1], gens[2]),
	}
	if !p.Equal(lsv, q) {
		t.Errorf("p and q should be equal")
	}

	// different lengths
	q = Path{
		NewEdge(gens[0], gens[1]),
	}
	if p.Equal(lsv, q) {
		t.Errorf("p and q should not be equal")
	}
	q = Path{
		NewEdge(gens[0], gens[1]),
		NewEdge(gens[1], gens[3]),
	}
	if p.Equal(lsv, q) {
		t.Errorf("p and q should not be equal")
	}
}

// xxx tbd
// func TestPathSplitByChord(t *testing.T) {
	
// 	tests := []struct {
// 		p Path
// 		chord Edge
// 		wantA, wantB Path
// 	}{
// 		{
// 			Path{
// 				NewEdge(*MatGfIdentity, g),
// 				NewEdge(g, g2),
// 			},
// 			NewEdge(*MatGfIdentity, g2),
// 			Path{
// 				NewEdge(*MatGfIdentity, g),
				
// 		},
// 	}
// 	for n, test := range tests {
// 		gotA, gotB := test.p.SplitByChord(lsv, test.chord)
// 		if !gotA.Equal(lsv, test.wantA) {
// 			t.Errorf("test=%d gotA=%v wantA=%v", n, gotA, test.wantA)
// 		}
// 		if !gotB.Equal(lsv, test.wantB) {
// 			t.Errorf("test=%d gotB=%v wantB=%v", n, gotB, test.wantB)
// 		}
// 	}
// }

func TestPathStart(t *testing.T) {
	trials := 100
	for i := 0; i < trials; i++ {
		start, p := randomPath(10)
		got := p.Start(lsv)
		if !got.Equal(lsv, &start) {
			t.Errorf("trial=%d got=%v start=%v", i, got, start)
		}
	}
}

func TestPathWord(t *testing.T) {
	//     f     g           h
	// id --- f --- u = fg  --- v = uh
	//                
	gens := lsv.Generators()
	f, g, h := gens[0], gens[1], gens[2]
	id := MatGfIdentity
	u := f.Multiply(lsv, &g).MakeCanonical(lsv)
	v := u.Multiply(lsv, &h).MakeCanonical(lsv)
	p := NewPathFromEdges(lsv, []Edge{
		NewEdge(*id, f),
		NewEdge(f, *u),
		NewEdge(*u, *v),
	})
	gotWord, gotVertices := p.Word(lsv)
	wantWord := []MatGF{f, g, h}
	wantVertices := []MatGF{*id, f, *u, *v}
	if !reflect.DeepEqual(gotWord, wantWord) {
		t.Errorf("gotWord=%v wantWord=%v", gotWord, wantWord)
	}
	if !reflect.DeepEqual(gotVertices, wantVertices) {
		t.Errorf("gotVertices=%v wantVertices=%v", gotVertices, wantVertices)
	}
}

func TestEdgeTranslate(t *testing.T) {
	// dumpProduct("[1 5 14 11 10 2 14 9 6]", "[1 5 14 11 10 8 14 3 12]") // b0*b1; this gives us edge (b0, b0*b1)
    // translate by b2
	// dumpProduct("[1 7 8 15 14 3 8 12 5]", "[1 5 14 11 10 2 14 9 6]")
	// dumpProduct("[1 7 8 15 14 3 8 12 5]", "[1 0 7 13 6 9 0 9 5]")
	tests := []struct {
		e Edge
		g MatGF
		want Edge
	}{
		{
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1},
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
		},
		{
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,},
			NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 13, 13, 13, 14, 15, 0, 2, 12,}),
		},
		{
			// b0 = [1 5 14 11 10 2 14 9 6]
			// b1 = [1 5 14 11 10 8 14 3 12]
			// edge (b0, b0*b1)
			NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 0, 7, 13, 6, 9, 0, 9, 5,}),
			// translate by b2 = [1 7 8 15 14 3 8 12 5]
			MatGF{1, 7, 8, 15, 14, 3, 8, 12, 5,},
			// edge (b2*b0, b2*b0*b1)
			// ([1 7 4 9 15 0 9 14 12], [1 12 4 12 11 4 6 12 15]
			NewEdge(MatGF{1, 7, 4, 9, 15, 0, 9, 14, 12,}, MatGF{1, 12, 4, 12, 11, 4, 6, 12, 15,}),
		},
	}
	for _, test := range tests {
		got := test.e.Translate(lsv, &test.g)
		if !got.Equal(lsv, test.want) {
			t.Errorf("got=%v want=%v", got, test.want)
		}
	}
}

func TestEdgeFromString(t *testing.T) {
	tests := []struct {
		s string
		want Edge
	}{
		{
			"[[1 0 0 7 1 9 13 4 8] [1 5 14 11 10 2 14 9 6]]",
			NewEdge(MatGF{1, 0, 0, 7, 1, 9, 13, 4, 8}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6}),
		},
	}
	for _, test := range tests {
		got := NewEdgeFromString(test.s)
		if !got.Equal(lsv, test.want) {
			t.Errorf("NewEdgeFromString(%v) got=%v want=%v", test.s, got, test.want)
		}
	}
}

func TestEdgeOrbitContainsEdge(t *testing.T) {
	tests := []struct {
		E, f Edge
		wantContains bool
		wantMatGF *MatGF
	}{
		{
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			true,
			&MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1},
		},
		{
			NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
			NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 13, 13, 13, 14, 15, 0, 2, 12}),
			true,
			&MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6},
		},
	}
	for _, test := range tests {
		gotMatGF, gotContains := test.E.OrbitContainsEdge(lsv, test.f)
		if gotContains != test.wantContains {
			t.Errorf("gotContains=%v wantContains=%v", gotContains, test.wantContains)
		}
		if test.wantMatGF != nil {
			if gotMatGF == nil {
				t.Errorf("gotMatGF=nil wantMatGF=%v", test.wantMatGF)
			} else if !gotMatGF.Equal(lsv, test.wantMatGF) {
				t.Errorf("gotMatGF=%v wantMatGF=%v", gotMatGF, test.wantMatGF)
			}
		}
	}
	
	// take some edge e.
	// translate by some group element g.
	// translated edge g*e should be in orbit of e.
	trials := 100
	for i := 0; i < trials; i++ {
		e := randomEdge()
		g := randomVertex()
		ge := e.Translate(lsv, &g)
		h, ok := e.OrbitContainsEdge(lsv, ge)
		if !ok {
			t.Errorf("e.OrbitContainsEdge(ge) failed for e=%v g=%v ge=%v", e, g, ge)
		}
		if !h.Equal(lsv, &g) {
			t.Errorf("e.OrbitContainsEdge(ge) failed for e=%v g=%v ge=%v h=%v", e, g, ge, h)
		}
	}
}

func randomEdge() Edge {
	v := randomVertex()
	h := randomGenerator()
	return NewEdge(v, *v.Multiply(lsv, h))
}

func randomVertex() Vertex {
	return *RandomInvertibleMatGF(lsv)
}

func randomGenerator() *MatGF {
	gens := lsv.Generators()
	return &gens[rand.Intn(len(gens))]
}

func randomPath(length int) (start Vertex, path Path) {
	path = make(Path, length)
	start = randomVertex()
	curVertex := start
	for i := 0; i < length; i++ {
		var v *Vertex
		var e Edge
		// repeat to disallow self-loops
		for {
			g := randomGenerator()
			v = curVertex.Multiply(lsv, g).MakeCanonical(lsv)
			e = NewEdge(curVertex, *v)
			if i == 0 || !e.Equal(lsv, path[i-1]) {
				break
			}
		}
		path[i] = e
		curVertex = *v
	}
	return start, path
}

func TestExpectedEdgeOrbits(t *testing.T) {
	// check that the orbits of edges at the origin are distinct,
	// except an edge and its inverse are in the same orbit.
	gens := lsv.Generators()
	for i, g := range gens {
		e := NewEdge(*MatGfIdentity, g)
		for j, h := range gens {
			_, contains := e.OrbitContainsEdge(lsv, NewEdge(*MatGfIdentity, h))
			if g.Equal(lsv, &h) || g.Inverse(lsv).Equal(lsv, &h) {
				if !contains {
					t.Errorf("edge orbit test %v %v failed", i, j)
				}
			} else {
				if contains {
					t.Errorf("edge orbit test %v %v failed", i, j)
				}
			}
		}
	}
}

// this example comes from bin/cycles
// 2023/05/10 08:18:58 cycle: [[[1 0 0 0 1 0 0 0 1] [1 5 14 11 10 2 14 9 6]] [[1 5 14 11 10 2 14 9 6] [1 10 10 5 2 0 10 0 13]] [[1 0 0 0 1 0 0 0 1] [1 10 10 5 2 0 10 0 13]]]
var exampleTriangle = NewTriangle(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1},
		MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,},
		MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,})

func TestTriangleEdges(t *testing.T) {
	got := exampleTriangle.Edges()
	want := [3]Edge{
		NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}),
		NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
		NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got=%v want=%v", got, want)
	}
}

// xxx TODO NEXT; rewrite using randomness  
func TestTriangleOrbitContainsEdge(t *testing.T) {
	// wip
// 	T := LsvTrianglesAtOrigin()
// 	t, u := trianglesNotSharingEdge(T)
}

// xxx maybe
// func TestFindExampleTwoTrianglesSharingEdgeOrbit(t *testing.T) {
// }

func TestTrianglesAtOriginShareZeroOrOneEdge(t *testing.T) {
	T := LsvTrianglesAtOrigin(lsv)
	m := 0
	n := 0
	for i, u := range T {
		for j, v := range T {
			edges := u.SharedEdges(lsv, v)
			if len(edges) == 0 {
				m++
			} else {
				if i == j {
					if len(edges) != 3 {
						t.Errorf("i=%v j=%v u=%v v=%v edges=%v", i, j, u, v, edges)
					}
				} else {
					n++
					if len(edges) > 1 {
						t.Errorf("i=%v j=%v u=%v v=%v edges=%v", i, j, u, v, edges)
					}
				}
			}
		}
	}
	m = m / 2
	n = n / 2
	// there are 21 triangles at the origin.  21 choose 2 is 210.  we
	// expect m pairs of triangles to share no edges, and n pairs to
	// share one edge. xxx prove this without using computation, so
	// that this test is not simply self-validating.
	expectedM := 168
	expectedN := 42
	if m != expectedM {
		t.Errorf("m=%d expected=%d", m, expectedM)
	}
	if n != expectedN {
		t.Errorf("n=%d expected=%d", n, expectedN)
	}
}

func TestTriangleEdgesAreInDistinctOrbits(t *testing.T) {
	T := LsvTrianglesAtOrigin(lsv)
	for _, u := range T {
		edges := u.Edges()
		e, f, g := edges[0], edges[1], edges[2]
		if _, ok := e.OrbitContainsEdge(lsv, f); ok {
			t.Errorf("e=%v f=%v g=%v", e, f, g)
		}
		if _, ok := e.OrbitContainsEdge(lsv, g); ok {
			t.Errorf("e=%v f=%v g=%v", e, f, g)
		}
		if _, ok := f.OrbitContainsEdge(lsv, g); ok {
			t.Errorf("e=%v f=%v g=%v", e, f, g)
		}
	}
}

func TestExpectedTriangleOrbits(t *testing.T) {
	// check that the orbits of triangles at the origin are distinct.
	T := LsvTrianglesAtOrigin(lsv)
	for i, u := range T {
		for j, v := range T {
			_, ok := u.OrbitContainsTriangle(lsv, v)
			if i == j {
				if !ok {
					t.Errorf("i=%d j=%d u=%v v=%v", i, j, u, v)
				}
			} else {
				if ok {
					t.Errorf("i=%d j=%d u=%v v=%v", i, j, u, v)
				}
			}
		}
	}
}

func TestTriangleTranslate(t *testing.T) {
    T := LsvTrianglesAtOrigin(lsv)
	trials := 100
	for i := 0; i < trials; i++ {
		u := T[rand.Intn(len(T))]
        g := randomVertex()
		v := u.Translate(lsv, &g)
		if g.IsIdentity(lsv) {
			if !u.Equal(lsv, &v) {
				t.Errorf("u=%v g=%v v=%v", u, g, v)
			}
		} else {
			if u.Equal(lsv, &v) {
				t.Errorf("u=%v g=%v v=%v", u, g, v)
			}
		}
		h := randomVertex()
		w := v.Translate(lsv, &h)
		if h.IsIdentity(lsv) {
			if !v.Equal(lsv, &w) {
				t.Errorf("v=%v h=%v w=%v", v, h, w)
			}
		} else {
			if v.Equal(lsv, &w) {
				t.Errorf("v=%v h=%v w=%v", v, h, w)
			}
		}
		hInv := h.Inverse(lsv)
		x := w.Translate(lsv, hInv)
		if !x.Equal(lsv, &v) {
			t.Errorf("v=%v h=%v w=%v x=%v", v, h, w, x)
		}
		gInv := g.Inverse(lsv)
		y := x.Translate(lsv, gInv)
		if !y.Equal(lsv, &u) {
			t.Errorf("u=%v g=%v x=%v y=%v", u, g, x, y)
		}
	}
}

func dumpProduct(a, b string) {
	ma := NewMatGFFromString(a)
	mb := NewMatGFFromString(b)
	mab := ma.Multiply(lsv, &mb).MakeCanonical(lsv)
	log.Printf("%v * %v = %v", ma, mb, *mab)
}

func exampleTriangleToComplex() *Complex {
	edges := make(map[Edge]any)
	vertices := make(map[Vertex]any)	

	for _, edge := range exampleTriangle.Edges() {
		edges[edge] = nil
		vertices[edge[0]] = nil
		vertices[edge[1]] = nil
	}
	vertexBasis := make([]Vertex, len(vertices))
	i := 0
	for vertex, _ := range vertices {
		vertexBasis[i] = vertex
		i++
	}
	edgeBasis := make([]Edge, len(edges))
	i = 0
	for edge, _ := range edges {
		edgeBasis[i] = edge
		i++
	}
	triangleBasis := []Triangle{exampleTriangle}
	sortBases := true // to make test deterministic
	return NewComplex(vertexBasis, edgeBasis, triangleBasis, sortBases, false)
}

func TestComplexD1(t *testing.T) {
	C := exampleTriangleToComplex()
	got := C.D1()
	want := NewSparseBinaryMatrixFromRowInts([][]uint8{
		{1, 1, 0},
		{1, 0, 1},
		{0, 1, 1},
	})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%v\nwant:\n%v\n", got, want)
	}
}

func TestComplexD2Transpose(t *testing.T) {
	C := exampleTriangleToComplex()
	got := C.D2Transpose()
	want := NewSparseBinaryMatrixFromRowInts([][]uint8{
		{1, 1, 1},
	})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%v\nwant:\n%v\n", got, want)
	}
}

func TestComplexPathToEdgeVector(t *testing.T) {
	C := exampleTriangleToComplex()
	got := C.PathToEdgeVector(Path{
		NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
		NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
	})
	want := NewBinaryVectorFromInts([]uint8{0, 1, 1})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%v\nwant:\n%v\n", got, want)
	}
}

func TestComplexEdgeVectorToPath(t *testing.T) {
	C := exampleTriangleToComplex()
	v := NewSparseBinaryMatrixFromRowInts([][]uint8{
		{0},
		{1},
		{1},
	})
	got := C.EdgeVectorToPath(lsv, v)
	want := Path{
		NewEdge(MatGF{1, 0, 0, 0, 1, 0, 0, 0, 1}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
		NewEdge(MatGF{1, 5, 14, 11, 10, 2, 14, 9, 6,}, MatGF{1, 10, 10, 5, 2, 0, 10, 0, 13,}),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%v\nwant:\n%v\n", got, want)
	}
}

func TestReadWriteVertexFile(t *testing.T) {
	trials := 10
	maxSize := 1000
	file := ".testRWVertex.txt~"
	for i := 0; i < trials; i++ {
		n := rand.Intn(maxSize) + 1
		vertices := make([]Vertex, n)
		for j := 0; j < n; j++ {
			vertices[j] = *RandomMatGF(lsv)
		}
		WriteVertexFile(vertices, file)
		got := ReadVertexFile(file)
		if !reflect.DeepEqual(got, vertices) {
			t.Errorf("got:\n%v\nwant:\n%v\n", got, vertices)
		}
	}
}

func TestReadWriteEdgeFile(t *testing.T) {
	trials := 10
	maxSize := 1000
	file := ".testRWEdge.txt~"
	for i := 0; i < trials; i++ {
		n := rand.Intn(maxSize) + 1
		edges := make([]Edge, n)
		for j := 0; j < n; j++ {
			edges[j] = NewEdge(*RandomMatGF(lsv), *RandomMatGF(lsv))
		}
		WriteEdgeFile(edges, file)
		got := ReadEdgeFile(file)
		if !reflect.DeepEqual(got, edges) {
			t.Errorf("got:\n%v\nwant:\n%v\n", got, edges)
		}
	}
}

func TestReadWriteTriangleFile(t *testing.T) {
	trials := 10
	maxSize := 1000
	file := ".testRWTriangle.txt~"
	for i := 0; i < trials; i++ {
		n := rand.Intn(maxSize) + 1
		triangles := make([]Triangle, n)
		for j := 0; j < n; j++ {
			triangles[j] = NewTriangle(*RandomMatGF(lsv), *RandomMatGF(lsv), *RandomMatGF(lsv))
		}
		WriteTriangleFile(triangles, file)
		got := ReadTriangleFile(file)
		if !reflect.DeepEqual(got, triangles) {
			t.Errorf("got:\n%v\nwant:\n%v\n", got, triangles)
		}
	}
}
