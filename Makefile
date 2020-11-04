.PHONY: all

all: warewulfd wwbuild wwclient

files: all
	install -d -m 0755 /var/warewulf/
	install -d -m 0755 /etc/warewulf/
	install -d -m 0755 /var/lib/tftpboot/warewulf/ipxe/
	install -m 0640 etc/dhcpd.conf /etc/dhcp/dhcpd.conf
	install -m 0640 etc/nodes.conf /etc/warewulf/nodes.conf
	install -m 0640 etc/warewulf.conf /etc/warewulf/warewulf.conf
	cp -r tftpboot/* /var/lib/tftpboot/warewulf/ipxe/
	cp -r overlays /var/warewulf/
	chmod +x /var/warewulf/overlays/system/default/init
	mkdir -p /var/warewulf/overlays/system/default/warewulf/bin/
	cp wwclient /var/warewulf/overlays/system/default/warewulf/bin/

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

