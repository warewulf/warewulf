.PHONY: all

all: warewulfd wwbuild wwclient

files: all
	sudo install -d -m 0755 /var/warewulf/provision
	sudo install -d -m 0755 /var/warewulf/provision/kernels
	sudo install -d -m 0755 /var/warewulf/provision/overlays
	sudo install -d -m 0755 /var/warewulf/provision/bases
	sudo install -d -m 0755 /etc/warewulf/
	sudo install -d -m 0755 /var/lib/tftpboot/warewulf/ipxe/
	sudo install -m 0644 dhcpd.conf /etc/dhcp/dhcpd.conf
	sudo install -m 0644 nodes.yaml /etc/warewulf/nodes.yaml
	sudo cp -r tftpboot/* /var/lib/tftpboot/warewulf/ipxe/
	sudo cp -r overlays /var/warewulf/
	sudo chmod +x /var/warewulf/overlays/system/default/init
	sudo mkdir -p /var/warewulf/overlays/system/default/warewulf/bin/
	sudo cp wwclient /var/warewulf/overlays/system/default/warewulf/bin/

services: files
	sudo systemctl enable tftp
	sudo systemctl restart tftp
	sudo systemctl enable dhcpd
	sudo systemctl restart dhcpd

warewulfd:
	cd cmd/warewulfd; go build -o ../../warewulfd

wwbuild:
	cd cmd/wwbuild; go build -o ../../wwbuild

wwclient:
	cd cmd/wwclient; CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags -static' -o ../../wwclient

clean:
	rm -f warewulfd
	rm -f wwbuild
	rm -f wwclient

install: files services

