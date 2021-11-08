package warewulfd

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"reflect"
	"testing"
)

func TestGetNode(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    node.NodeInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNode(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadNodeDB(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadNodeDB(); (err != nil) != tt.wantErr {
				t.Errorf("LoadNodeDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// NodeList testing
//  ||
//  \/
func TestNodeList_AddTestNodes(t *testing.T) {
	var testDB = make(NodeList)
	type args struct {
		c int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"zero", args{0}, 0},
		{"2 nodes", args{2}, 2},
		{"3 nodes", args{3}, 3},
		{"2 nodes again", args{2}, 3},
		{"4 nodes", args{4}, 4},
		{"negative number", args{-5}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.AddTestNodes(tt.args.c)
			got := len(testDB)
			if got != tt.want {
				t.Errorf("AddTestNodes() failed to add requested number of test nodes %v, want %v", got, tt.want)
			}
		})
	}
}

// helpers
func makeNodeList(n int) NodeList {
	var ndb = make(NodeList)
	for i := 0; i < n; i++ {
		s := fmt.Sprintf("n_%d", i)
		ndb[s] = &node.NodeInfo{}
		ndb[s].Id.Set(s)
		ndb[s].LastSeen = (int64(n - i))
	}
	return ndb
}

func makeNodeSlice(l NodeList) NodeSlice {
	s := make(NodeSlice, 0, len(l))

	for _, d := range l {
		s = append(s, d)
	}
	return s
}

//func TestNodeList_JsonSend(t *testing.T) {
//	type args struct {
//		w http.ResponseWriter
//	}
//	tests := []struct {
//		name string
//		n    NodeList
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//		})
//	}
//}

func TestNodeList_Sort(t *testing.T) {
	tests := []struct {
		name string
		n    NodeList
		want []int64
	}{
		{"t1", makeNodeList(3), []int64{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.Sort()
			res := []int64{}
			for _, v := range got {
				res = append(res, v.LastSeen)
			}

			if got := res; !reflect.DeepEqual(res, tt.want) {
				t.Errorf("Sort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeSlice_Len(t *testing.T) {
	tests := []struct {
		name string
		d    NodeSlice
		want int
	}{
		{"len=3", makeNodeSlice(makeNodeList(3)), 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeSlice_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		d    NodeSlice
		args args
		want bool
	}{
		// note, node.LastSeen are in reserved order!
		{"2,1 is true", makeNodeSlice(makeNodeList(3)), args{i: 2, j: 1}, true},
		{"0,1 is false", makeNodeSlice(makeNodeList(3)), args{i: 0, j: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("Less() = %v, want %v, data: %v", got, tt.want, tt.d)
			}
		})
	}
}
