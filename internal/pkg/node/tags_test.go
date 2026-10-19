package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ApplyTagChanges(t *testing.T) {
	tests := map[string]struct {
		profile Profile
		add     NodeConfAdd
		del     NodeConfDel
		want    Profile
	}{
		"no changes": {
			profile: Profile{},
			add:     NodeConfAdd{Net: "default"},
			want:    Profile{},
		},
		"tagadd does not touch network tags": {
			profile: Profile{
				NetDevs: map[string]*NetDev{"default": {Device: "eth0", Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
			add: NodeConfAdd{Net: "default", TagsAdd: map[string]string{"foo": "bar"}},
			want: Profile{
				Tags:    map[string]string{"foo": "bar"},
				NetDevs: map[string]*NetDev{"default": {Device: "eth0", Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
		},
		"tagadd does not create network tags": {
			profile: Profile{
				NetDevs: map[string]*NetDev{"default": {Device: "eth0"}},
			},
			add: NodeConfAdd{Net: "default", TagsAdd: map[string]string{"foo": "bar"}},
			want: Profile{
				Tags:    map[string]string{"foo": "bar"},
				NetDevs: map[string]*NetDev{"default": {Device: "eth0"}},
			},
		},
		"nettagadd does not touch tags": {
			profile: Profile{
				Tags:    map[string]string{"foo": "bar"},
				NetDevs: map[string]*NetDev{"default": {Device: "eth0"}},
			},
			add: NodeConfAdd{Net: "default", NetTagsAdd: map[string]string{"DNS1": "1.1.1.1"}},
			want: Profile{
				Tags:    map[string]string{"foo": "bar"},
				NetDevs: map[string]*NetDev{"default": {Device: "eth0", Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
		},
		"tagadd and nettagadd each go to their own map": {
			profile: Profile{
				NetDevs: map[string]*NetDev{"default": {Device: "eth0"}},
			},
			add: NodeConfAdd{
				Net:        "default",
				TagsAdd:    map[string]string{"foo": "bar"},
				NetTagsAdd: map[string]string{"DNS1": "1.1.1.1"},
			},
			want: Profile{
				Tags:    map[string]string{"foo": "bar"},
				NetDevs: map[string]*NetDev{"default": {Device: "eth0", Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
		},
		"nettagadd uses netname": {
			profile: Profile{
				NetDevs: map[string]*NetDev{
					"default": {Device: "eth0"},
					"ib0":     {Device: "ib0"},
				},
			},
			add: NodeConfAdd{Net: "ib0", TagsAdd: map[string]string{"foo": "bar"}, NetTagsAdd: map[string]string{"mtu": "9000"}},
			want: Profile{
				Tags: map[string]string{"foo": "bar"},
				NetDevs: map[string]*NetDev{
					"default": {Device: "eth0"},
					"ib0":     {Device: "ib0", Tags: map[string]string{"mtu": "9000"}},
				},
			},
		},
		"nettagadd on missing network device": {
			profile: Profile{
				NetDevs: map[string]*NetDev{"default": {Device: "eth0"}},
			},
			add: NodeConfAdd{Net: "ib0", NetTagsAdd: map[string]string{"mtu": "9000"}},
			want: Profile{
				NetDevs: map[string]*NetDev{"default": {Device: "eth0"}},
			},
		},
		"nettagdel on missing network device": {
			profile: Profile{
				NetDevs: map[string]*NetDev{"default": {Device: "eth0", Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
			add: NodeConfAdd{Net: "eth9"},
			del: NodeConfDel{NetTagsDel: []string{"DNS1"}},
			want: Profile{
				NetDevs: map[string]*NetDev{"default": {Device: "eth0", Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
		},
		"tagdel does not touch network tags": {
			profile: Profile{
				Tags:    map[string]string{"DNS1": "1.1.1.1"},
				NetDevs: map[string]*NetDev{"default": {Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
			add: NodeConfAdd{Net: "default"},
			del: NodeConfDel{TagsDel: []string{"DNS1"}},
			want: Profile{
				Tags:    map[string]string{},
				NetDevs: map[string]*NetDev{"default": {Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
		},
		"nettagdel does not touch tags": {
			profile: Profile{
				Tags:    map[string]string{"DNS1": "1.1.1.1"},
				NetDevs: map[string]*NetDev{"default": {Tags: map[string]string{"DNS1": "1.1.1.1"}}},
			},
			add: NodeConfAdd{Net: "default"},
			del: NodeConfDel{NetTagsDel: []string{"DNS1"}},
			want: Profile{
				Tags:    map[string]string{"DNS1": "1.1.1.1"},
				NetDevs: map[string]*NetDev{"default": {Tags: map[string]string{}}},
			},
		},
		"add and delete in one call": {
			profile: Profile{
				Tags:    map[string]string{"a": "1"},
				Ipmi:    &IpmiConf{Tags: map[string]string{"b": "2"}},
				NetDevs: map[string]*NetDev{"default": {Tags: map[string]string{"c": "3"}}},
			},
			add: NodeConfAdd{
				Net:         "default",
				TagsAdd:     map[string]string{"x": "1"},
				IpmiTagsAdd: map[string]string{"y": "2"},
				NetTagsAdd:  map[string]string{"z": "3"},
			},
			del: NodeConfDel{
				TagsDel:     []string{"a"},
				IpmiTagsDel: []string{"b"},
				NetTagsDel:  []string{"c"},
			},
			want: Profile{
				Tags:    map[string]string{"x": "1"},
				Ipmi:    &IpmiConf{Tags: map[string]string{"y": "2"}},
				NetDevs: map[string]*NetDev{"default": {Tags: map[string]string{"z": "3"}}},
			},
		},
		"ipmitagadd with nil ipmi": {
			profile: Profile{},
			add:     NodeConfAdd{Net: "default", IpmiTagsAdd: map[string]string{"y": "2"}},
			want: Profile{
				Ipmi: &IpmiConf{Tags: map[string]string{"y": "2"}},
			},
		},
		"ipmitagdel with nil ipmi": {
			profile: Profile{},
			add:     NodeConfAdd{Net: "default"},
			del:     NodeConfDel{IpmiTagsDel: []string{"y"}},
			want:    Profile{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.profile.ApplyTagChanges(tt.add, tt.del)
			assert.Equal(t, tt.want, tt.profile)
		})
	}
}
