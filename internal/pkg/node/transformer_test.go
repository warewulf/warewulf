package node

// import (
// 	"reflect"
// 	"testing"
// )

// func Test_nodeYaml_SetFrom(t *testing.T) {
// 	c, _ := NewTestNode()
// 	singleNodeConf := c.Nodes["test_node"]
// 	singleNodeInfo := NewInfo()
// 	singleNodeInfo.SetFrom(singleNodeConf)
// 	tests := []struct {
// 		name    string
// 		arg     string
// 		want    string
// 		wantErr bool
// 	}{
// 		{"Right comment", "Comment", "Node Comment", false},
// 		{"FieldName", "comment", "NodeComment", true},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := GetByName(&singleNodeInfo, tt.arg)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetByName(%s,%s) error = %v, wantErr %v",
// 					reflect.TypeOf(singleNodeConf), tt.arg, err, tt.wantErr)
// 				return
// 			}
// 			if (got != tt.want) != tt.wantErr {
// 				t.Errorf("GetByName(%s,%s) got = %v, want = %v",
// 					reflect.TypeOf(singleNodeConf), tt.arg, got, tt.want)
// 				return
// 			}
// 		})
// 	}
// }
