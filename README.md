# warewulf v4 (work in progress)

![Warewulf](warewulf-logo.png)

In a nutshell, to install and start provisioning nodes, do the following:

#### Build Warewulf and dependencies:

Warewulf is programmed in GoLang, so you will need to also install a Go compiler
on your system. The easiest way to do this on RHEL and CentOS is by using the Go
packages that are included in EPEL. In addition there are some dependencies that
Warewulf will reuire to operate properly and because this is a quick HOWTO, we
disable the firewall so any provisioning issues we have won't be caused by
packets being dropped.

To install these and compile Warewulf, do the following:

```
sudo yum install epel-release
sudo yum install --tolerant golang tftp-server dhcp dhcp-server

sudo systemctl stop firewalld
sudo systemctl disable firewalld

make all
```
    
#### Install Warewulf onto your host system:

The following command will install some of the Warewulf shared components into your
host system. The locations of these paths and files are not yet configurable, so
it can't be installed anywhere else.

The default install locations are:

* `/var/warewulf`: State data
* `/etc/warewulf`: Configuration data

```
sudo make install
```

#### Configure the controller:

Edit the file `/etc/warewulf/warewulf.conf` and ensure that you've set the
appropriate configuration paramaters. Here are some of the defaults for reference:

```
ipaddr: 192.168.1.1
netmask: 255.255.255.0
warewulf:
  port: 9873
  secure: true
  update interval: 60
dhcp:
  enabled: true
  range start: 192.168.1.150
  range end: 192.168.1.200
  template: default
  systemd name: dhcpd
tftp:
  enabled: true
  tftproot: /var/lib/tftpboot
  systemd name: tftp
```

Note: You may need to change the systemd service names for your distribution.

Once it has been configured, you can have Warewulf configure the services:

```
sudo ./wwctl service -a
```

#### Pull and build the VNFS container and kernel:

Once you have added the node, you can start building the needed bootable components.
There are three major groups of data to provision:

1. Kernel: This is the boot kernel and driver overlay pair. The `wwctl kernel build`
   command will create these files. The `-a` option will scan your configuration looking
   for all needed kernel images and then go through and build them all. The caveat is
   that all these kernels must be installed to your host controller node.

1. Container/VNFS: The "VNFS" is the "Virtual Node File System", and that is the template that
   nodes will be provisioned to boot into. Warewulf v4 can support standard "chroot"
   style VNFS formats (e.g. same as Warewulf 3 and/or Singularity Sandboxes) as well as
   OCI (Open Container Initiative) formats which include Docker containers and containers
   in Docker Hub. As you can see in the above `wwctl profile set` command, we configured
   the VNFS to be a container hosted in Docker Hub. This can also be a local path or a
   container in a `docker-daemon`.

1. Overlays: There are two types of overlays, "system" and "runtime". The difference is
   that the system overlay is provisioned before `/sbin/init` is called and the runtime
   overlay is provisioned after `/sbin/init` is called and is done from the booted operating
   system at periodic intervals (the time of this writing, it is ever 30 seconds).
   

```
sudo ./wwctl container pull docker://warewulf/centos-7 centos-7
sudo ./wwctl container build centos-7
sudo ./wwctl kernel build $(uname -r)
```

#### Set up the default node profile

```
sudo ./wwctl profile set default -K $(uname -r) -C centos-7
sudo ./wwctl profile set default --netdev eth0 -M 255.255.255.0 -G 192.168.1.1
sudo ./wwctl profile list
```
    
#### Add a node and build node specific overlays

```
sudo ./wwctl node add n0000.cluster --netdev eth0 -I 192.168.1.100 -H 00:0c:29:23:8b:48
sudo ./wwctl node list -a n0000
sudo ./wwctl overlay build -a
```
    
#### Start the Warewulf daemon:

Once the above provisioning images are built, you can check the provisioning "rediness"
and then begin booting nodes.

```
sudo ./wwctl ready
./warewulfd
```
    
#### Boot your compute node and watch it boot

