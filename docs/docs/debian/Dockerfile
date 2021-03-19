FROM debian:buster-slim

RUN apt-get update

RUN apt-get -y install \
	kmod \
	systemd-sysv \
	openssh-server \
	openssh-client \
	isc-dhcp-client \
	pciutils \
	strace \
	nfs-common \
	ethtool\
	ifupdown

RUN apt-get -y install --no-install-recommends \
  linux-image-amd64

RUN rm -rf /var/lib/apt/lists/*

RUN echo "root:r00tr00t" | chpasswd
