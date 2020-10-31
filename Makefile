

files:
	sudo install -d -m 0755 /var/warewulf/provision
	sudo install -d -m 0755 /var/warewulf/provision/kernels
	sudo install -d -m 0755 /var/warewulf/provision/overlays
	sudo install -d -m 0755 /var/warewulf/provision/bases
	sudo install -d -m 0755 /etc/warewulf/
	sudo install -d -m 0755 /var/lib/tftpboot/warewulf/ipxe/
	sudo install -m 0644 dhcpd.conf /etc/dhcp/dhcpd.conf
	sudo install -m 0644 nodes.yaml /etc/warewulf/nodes.yaml
	sudo cp -r tftpboot/* /var/lib/tftpboot/warewulf/ipxe/
	sudo cp -r overlays /etc/warewulf/
	sudo chmod +x /etc/warewulf/overlays/generic/init

services: files
	sudo systemctl enable tftp
	sudo systemctl restart tftp
	sudo systemctl enable dhcpd
	sudo systemctl restart dhcpd

build:
	go build cmd/warewulfd/warewulfd.go
	go build cmd/wwbuild/wwbuild.go

clean:
	rm -f warewulfd
	rm -f wwbuild

install: build files services

