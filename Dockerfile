FROM bci/bci-init:latest

LABEL Description="Warewulf Base Container"
LABEL maintainer="Christian Goll <cgoll@suse.com>"


RUN zypper -n install \
  cpio \
  gzip \
  pigz \
  rsync \
  openssh-clients \
  less \
  dhcp-server \
  tftp \
  go1.18 \
  git \
  && \
  zypper -n install -t pattern devel_basis && \
  zypper clean -a && \
  systemctl enable dhcpd && \
  systemctl enable tftp.socket 

# now build the warewulf
COPY . /warewulf-src
RUN cd /warewulf-src &&\
  make clean &&\
  make lint &&\ 
  make

# Our dhcpd will listen on ANY interface, limits will be handled by container runtime
RUN export DHCPDCONF=/etc/sysconfig/dhcpd; test -e $DHCPDCONF && \ 
	sed -i 's/^DHCPD_INTERFACE=""/DHCPD_INTERFACE="ANY"/' $DHCPDCONF && \
	sed -i 's/^DHCPD_RUN_CHROOTED="yes"/DHCPD_RUN_CHROOTED="no"/' $DHCPDCONF && \
  WW4CONF=/etc/warewulf/warewulf.conf; test -e $WW4CONF && \
  sed -i 's/^ipaddr:.*/ipaddr: EMPTY/' $WW4CONF && \
  sed -i 's/^netmask:.*/netmask: EMPTY/' $WW4CONF && \
  sed -i 's/^network:.*/network: EMPTY/' $WW4CONF && \
  sed -i 's/^  range start:.*/  range start: EMPTY/' $WW4CONF && \
  sed -i 's/^  range end:.*/  range end: EMPTY/' $WW4CONF 


# We need the configs on the host as these files are quite important
RUN mkdir -p /container/warewulf &&  cp -rv /etc/warewulf/* /container/warewulf 
COPY warewulf.service \
 warewulf-container-manage.sh \
 wwctl \
 label-install \
 label-uninstall \
 label-purge \
 /container

# Add a service which will create a porper config on the startup
COPY ww4-config.service \
  /etc/systemd/system/

RUN systemctl enable ww4-config
 

RUN chmod +x \
  /container/wwctl \
  /container/warewulf-container-manage.sh \
  /container/label-*

# need systemd for tftp and dhcpd
#ENTRYPOINT [ "/container/label-run" ]
CMD  [ "/usr/sbin/init" ]

#EXPOSE 67/udp 68/udp 69/udp 9873

LABEL INSTALL="/usr/bin/docker run --env IMAGE=IMAGE --rm --privileged -v /:/host IMAGE /bin/bash /container/label-install"
LABEL UNINSTALL="/usr/bin/docker run --rm --privileged -v /:/host IMAGE /bin/bash /container/label-uninstall"
LABEL PURGE="/usr/bin/docker run -ti --rm --privileged -v /:/host IMAGE /bin/bash /container/label-purge"
LABEL RUN="/usr/bin/docker run -d --replace --name \${NAME} --privileged --net=host -v /:/host -v /etc/warewulf:/etc/warewulf -v /var/lib/warewulf/:/var/lib/warewulf/ -e NAME=\${NAME} -e IMAGE=\${IMAGE} \${IMAGE}"

