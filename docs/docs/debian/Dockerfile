FROM debian

# Disclaimer:
# this is WIP and subject to change, ant feedback is highly appreciated

# squelsh apt-get
ENV DEBIAN_FRONTEND noninteractive
ENV RUNLEVEL 1

# ----- install vital packages -----
# 'dbus' might makes sense as well
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

ENV DEBIAN_FRONTEND teletype
