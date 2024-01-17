package golsv

type MatrixEnumCayleyExpanderArgs struct {
	Limit       int
	ProfileArgs
}

type MatrixEnumCayleyExpander struct {
	Args *MatrixEnumCayleyExpanderArgs
}

func NewMatrixEnumCayleyExpander(args *MatrixEnumCayleyExpanderArgs) *MatrixEnumCayleyExpander {
	return &MatrixEnumCayleyExpander{
		Args: args,
	}
}

func (E *MatrixEnumCayleyExpander) Expand() {
	// log.Printf("Begin matrix enumeration")
	n := 0
	EnumerateMatrices(func(m *MatGF) (Continue bool) {
		//log.Printf("Matrix: %s", m.String())
		n++
		if n%(10*1000*1000) == 0 {
			// log.Printf("Enumerated %d matrices", n)
		}
		if E.Args.Limit > 0 && n >= E.Args.Limit {
			return false
		}
		return true
	})
	// log.Printf("Enumerated %d matrices", n)
}
