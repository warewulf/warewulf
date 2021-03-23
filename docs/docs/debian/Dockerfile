FROM debian

ENV DEBIAN_FRONTEND noninteractive
ENV RUNLEVEL 1

RUN apt-get update && apt-get install -y --no-install-recommends \
	kmod \
	systemd-sysv \
	openssh-client \
	openssh-server \
	isc-dhcp-client \
	pciutils \
	strace \
	nfs-common \
	ethtool\
	ifupdown \
	linux-image-amd64 \
	ifmetric \
	netbase && \
	rm -rf /var/lib/apt/lists/*

RUN echo "root:r00tr00t" | chpasswd

ENV DEBIAN_FRONTEND teletype
