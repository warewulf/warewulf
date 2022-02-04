package dhcp

import (
	"testing"
)

func TestDhcpTemplateFile(t *testing.T) {
	tests := []struct{
		parameter string
		expected string
	}{
		{"", "/usr/local/etc/warewulf/dhcp/default-dhcpd.conf"},
		{"default", "/usr/local/etc/warewulf/dhcp/default-dhcpd.conf"},
		{"static", "/usr/local/etc/warewulf/dhcp/static-dhcpd.conf"},
		{"/test/absolute/path.conf", "/test/absolute/path.conf"},
	}
	for _, tt := range tests {
		actual := dhcpTemplateFile(tt.parameter)
		if actual != tt.expected {
			t.Errorf("dhcpTemplateFile(%v) expected: %v, actual: %v",
				tt.parameter, tt.expected, actual)
		}
	}
}

