FROM opensuse/tumbleweed:latest as builder

RUN zypper  -n install --no-recommends git go1.18 libgpgme-devel &&\
  zypper -n install -t pattern devel_basis 

# now build the warewulf
COPY . /warewulf-src

RUN cd /warewulf-src &&\
  make contclean Defaults.mk \
    PREFIX=/usr \
    BINDIRa=/usr/bin \
    SYSCONFDIR=/etc \
    DATADIR=/usr/share \
    LOCALSTATEDIR=/var/lib \
    SHAREDSTATEDIR=/var/lib \
    MANDIR=/usr/share/man \
    INFODIR=/usr/share/info \
    DOCDIR=/usr/share/doc \
    SRVDIR=/var/lib \
    TFTPDIR=/srv/tftpboot \
    SYSTEMDDIR=/usr/lib/systemd/system \
    BASHCOMPDIR=/etc/warewulf/bash_completion.d/ \
    FIREWALLDDIR=/usr/lib/firewalld/services \
    WWCLIENTDIR=/warewulf &&\
  make lint &&\ 
  make &&\
  make install

# now all again but just the bare install for running the stuff
FROM opensuse/tumbleweed:latest

LABEL Description="Warewulf Base Container"
LABEL maintainer="Christian Goll <cgoll@suse.com>"

COPY --from=builder /usr/bin/wwctl /usr/bin/wwctl
COPY --from=builder /var/lib/warewulf /var/lib/warewulf
COPY --from=builder /usr/share/warewulf /usr/share/warewulf
COPY --from=builder /usr/lib/systemd/system/warewulfd.service /container-scripts/warewulfd.service
COPY --from=builder /etc/warewulf /etc/warewulf
COPY --from=builder /warewulf-src/container-scripts /container-scripts 

RUN zypper  -n install \
  cpio \
  gzip \
  pigz \
  rsync \
  openssh-clients \
  less \
  dhcp-server \
  iproute2 \
  vim \
  yq \
  tftp \
  systemd \
  && \
  zypper clean -a && \
  systemctl enable dhcpd && \
  systemctl enable tftp.socket &&\
  export DHCPDCONF=/etc/sysconfig/dhcpd; test -e $DHCPDCONF && \ 
	sed -i 's/^DHCPD_INTERFACE=""/DHCPD_INTERFACE="ANY"/' $DHCPDCONF && \
	sed -i 's/^DHCPD_RUN_CHROOTED="yes"/DHCPD_RUN_CHROOTED="no"/' $DHCPDCONF && \
  WW4CONF=/etc/warewulf/warewulf.conf; test -e $WW4CONF && \
  yq e '.ipaddr |= "EMPTY"' -i $WW4CONF && \
  yq e '.netmask |= "EMPTY"' -i $WW4CONF && \
  yq e '.network |= "EMPTY"' -i $WW4CONF && \
  yq e '.dhcp.["range start"] |= "EMPTY"'  -i $WW4CONF && \
  yq e '.dhcp.["range end"] |= "EMPTY"'  -i $WW4CONF && \
  yq e '.nfs.enabled |= false' -i $WW4CONF && \
  mkdir -p /container && \
  cp -vr container-scripts/label-* \
  container-scripts/wwctl \
  container-scripts/warewulf.service \
  container-scripts/warewulf-container-manage.sh \
  container-scripts/config-warewulf \
  /container &&\
  mkdir -p /usr/share/bash_completion/completions/ &&\
  cp /etc/warewulf/bash_completion.d/warewulf /usr/share/bash_completion/completions/warewulf &&\
  mv -v container-scripts/ww4-config.service /etc/systemd/system/ &&\
  mv -v container-scripts/warewulfd.service /etc/systemd/system/ &&\
  systemctl enable ww4-config &&\
  systemctl enable warewulfd

CMD  [ "/usr/sbin/init" ]

EXPOSE 67/udp 68/udp 69/udp 80 9873

LABEL INSTALL="/usr/bin/docker run --env IMAGE=IMAGE --rm --privileged -v /:/host IMAGE /bin/bash /container/label-install"
LABEL UNINSTALL="/usr/bin/docker run --rm --privileged -v /:/host IMAGE /bin/bash /container/label-uninstall"
LABEL PURGE="/usr/bin/docker run -ti --rm --privileged -v /:/host IMAGE /bin/bash /container/label-purge"
LABEL RUN="/usr/bin/docker run -d --replace --name \${NAME} --privileged --net=host -v /:/host -v /etc/warewulf:/etc/warewulf -v /var/lib/warewulf/:/var/lib/warewulf/ -e NAME=\${NAME} -e IMAGE=\${IMAGE} \${IMAGE}"

