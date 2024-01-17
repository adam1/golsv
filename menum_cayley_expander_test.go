package golsv

import (
	"testing"
)

func TestMenumExpandLimit(t *testing.T) {
	args := &MatrixEnumCayleyExpanderArgs{
		Limit: 32,
	}
	E := NewMatrixEnumCayleyExpander(args)
	E.Expand()
}
