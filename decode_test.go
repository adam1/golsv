package golsv

// import (
// 	"reflect"
// 	"testing"
// )

// xxx deprecated
// func TestEnumeratePowerSet(t *testing.T) {
// 	tests := []struct {
// 		keys []int
// 		want [][]int
// 	}{
// 		{
// 			keys: []int{},
// 			want: [][]int{{}},
// 		},
// 		{
// 			keys: []int{1},
// 			want: [][]int{{}, {1}},
// 		},
// 		{
// 			keys: []int{1, 2},
// 			want: [][]int{{}, {2}, {1}, {1, 2}},
// 		},
// 	}
// 	for n, tt := range tests {
// 		var got [][]int
// 		enumeratePowerSet(tt.keys, func(keys []int) bool {
// 			got = append(got, keys)
// 			return true
// 		})
// 		if !reflect.DeepEqual(got, tt.want) {
// 			t.Errorf("%d: got %v, want %v", n, got, tt.want)
// 		}
// 	}
// }

// xxx deprecated
// func TestCoboundaryOfEdges(t *testing.T) {
// 	tests := []struct {
// 		incidenceMap map[int][]int
// 		edges []int
// 		want map[int]any
// 	}{
// 		{
// 			incidenceMap: map[int][]int{
// 				1: {1, 2},
// 				2: {2, 3},
// 				3: {3, 4},
// 			},
// 			edges: []int{1},
// 			want: map[int]any{
// 				1: struct{}{},
// 				2: struct{}{},
// 			},
// 		},
// 		{
// 			incidenceMap: map[int][]int{
// 				1: {1, 2},
// 				2: {2, 3},
// 				3: {3, 4},
// 			},
// 			edges: []int{1, 2},
// 			want: map[int]any{
// 				1: struct{}{},
// 				3: struct{}{},
// 			},
// 		},
// 	}
// 	for n, tt := range tests {
// 		if got := coboundaryOfEdges(tt.incidenceMap, tt.edges); !reflect.DeepEqual(got, tt.want) {
// 			t.Errorf("%d: got %v, want %v", n, got, tt.want)
// 		}
// 	}
// }
