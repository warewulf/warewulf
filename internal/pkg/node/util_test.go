package node

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func NewUtilTestNode() (NodeYaml, error) {
	var data = `
nodeprofiles:
  default:
    comment: This profile is automatically included for each node
nodes:
  test_node:
    comment: Node Comment
    profiles:
    - default
    network devices:
      net0:
        default: true
        hwaddr: 00:00:00:00:12:34
        ipaddr: 1.2.3.4
        device: eth0
      net1:
        default: false
        hwaddr: ab:cd:ef:00:12:34
        ipaddr: 1.2.3.4
        device: eth1
      net2:
        default: false
        hwaddr: aB:Cd:eF:12:34:56
        ipaddr: 1.2.3.4
        device: eth2
  test_node_IPv6:
    profiles:
    - default
    network devices:
      net1:
        default: false
        ipaddr: fd1a:2b3c:4d5e:06f0:1234:5678:90ab:cdef
`
	var ret NodeYaml
	err := yaml.Unmarshal([]byte(data), &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func Test_nodeYaml_FindByHwaddr(t *testing.T) {
	c, _ := NewUtilTestNode()
	//type fields struct {
	//	NodeProfiles map[string]*NodeConf
	//	Nodes        map[string]*NodeConf
	//}
	type args struct {
		hwa string
	}
	tests := []struct {
		name string
		//fields  fields
		config  NodeYaml
		args    args
		want    string
		wantErr bool
	}{
		{"emptyString", c, args{hwa: ""}, "", true},
		{"noIpString", c, args{hwa: "this is not a MAC"}, "", true},
		{"intString", c, args{hwa: "4294967296"}, "", true},
		{"invalidMAC", c, args{hwa: "xx:00:00:00:12:34"}, "", true},
		{"validMACNotFound", c, args{hwa: "aa:FF:ee:65:43:21"}, "", true},
		{"validMAC", c, args{hwa: "ab:cd:ef:00:12:34"}, "test_node", false},
		{"validMAC2", c, args{hwa: "aB:Cd:eF:00:12:34"}, "test_node", false},
		{"validMAC3", c, args{hwa: "Ab:cD:Ef:12:34:56"}, "test_node", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			got, err := config.FindByHwaddr(tt.args.hwa)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByHwaddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !(got.Id.Get() == tt.want) {
				t.Errorf("FindByHwaddr() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeYaml_FindByIpaddr(t *testing.T) {
	c, _ := NewUtilTestNode()
	type args struct {
		ipaddr string
	}
	tests := []struct {
		name    string
		config  NodeYaml
		args    args
		want    string
		wantErr bool
	}{
		{"emptyString", c, args{ipaddr: ""}, "", true},
		{"noIpString", c, args{ipaddr: "this is not an IP"}, "", true},
		{"intString", c, args{ipaddr: "4294967296"}, "", true},
		{"invalidIPv4", c, args{ipaddr: "1.2.3.256"}, "", true},
		{"invalidIPv6", c, args{ipaddr: "xd1a:2b3c:4d5e:06f0:1234:5678:90ab:cdef"}, "", true},
		{"validIPv4NotFound", c, args{ipaddr: "1.1.1.1"}, "", true},
		{"validIPv6NotFound", c, args{ipaddr: "fd1a:2b3c:4d5e:06f0:1234:5678:90ab:fedc"}, "", true},
		{"validIPv4", c, args{ipaddr: "1.2.3.4"}, "test_node", false},
		{"validIPv6", c, args{ipaddr: "fd1a:2b3c:4d5e:06f0:1234:5678:90ab:cdef"}, "test_node_IPv6", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			got, err := config.FindByIpaddr(tt.args.ipaddr)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByIpaddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !(got.Id.Get() == tt.want) {
				t.Errorf("FindByHwaddr() got = %v, want %v", got, tt.want)
			}
		})
	}
}
