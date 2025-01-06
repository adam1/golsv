package golsv

import (
	"github.com/emicklei/dot"
)

type ZComplexToGraphviz[T any] struct {
	complex *ZComplex[T]
}

func NewZComplexToGraphviz[T any](C *ZComplex[T]) *ZComplexToGraphviz[T] {
	return &ZComplexToGraphviz[T]{complex: C}
}

func (viz *ZComplexToGraphviz[T]) Graphviz() (G *dot.Graph, err error) {
	g := dot.NewGraph(dot.Undirected)

	g.NodeInitializer(func(n dot.Node) {
		n.Attr("shape", "circle")
		n.Attr("width", "0.05")
		n.Attr("height", "0.05")
		n.Attr("fixedsize", "true")
		n.Attr("label", "")
		n.Attr("style", "filled")
		n.Attr("color", "black")
	})
	g.EdgeInitializer(func(e dot.Edge) {
		e.Attr("label", "")
	})

	for _, v := range viz.complex.VertexBasis() {
		g.Node(v.String())
	}
	//edgeIndex := viz.complex.EdgeIndex()
	for _, e := range viz.complex.EdgeBasis() {
		// with label
		// 		ei, ok := edgeIndex[e]
		// 		if !ok {
		// 			panic("edge not found in edge index")
		// 		}
		// g.Edge(g.Node(e[0].String()), g.Node(e[1].String())).Attr("label", fmt.Sprintf("%d", ei))

		// without label
		g.Edge(g.Node(e[0].String()), g.Node(e[1].String()))
	}
	return g, nil
}
