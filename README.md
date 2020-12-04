# warewulf v4 (work in progress)

![Warewulf](warewulf-logo.png)

In a nutshell, to install and start provisioning nodes, do the following:

####Build Warewulf and dependencies:

Warewulf is programmed in GoLang, so you will need to also install a Go compiler
on your system. The easiest way to do this on RHEL and CentOS is by using the Go
packages that are included in EPEL. In addition there are some dependencies that
Warewulf will reuire to operate properly and because this is a quick HOWTO, we
disable the firewall so any provisioning issues we have won't be caused by
packets being dropped.

To install these and compile Warewulf, do the following:

```
sudo yum install epel-release
sudo yum install golang tftp-server dhcp

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
dhcp:
  enabled: true
  range start: 192.168.1.150
  range end: 192.168.1.200
  template: default
```

Once it has been configured, you can have Warewulf configure the services:

```
sudo ./wwctl service dhcp -c
# sudo ./wwctl service tftp -c
```
    
#### Use the controller's kernel and default VNFS into your "default" node profile:

Next we are going to start configuring Warewulf. Use the following two commands to
set the kernel and VNFS for the "default" profile which all nodes will utilize, and
then configure node specific information as you add a new node to the configuration:

```
sudo ./wwctl profile set default -K $(uname -r) -V docker://warewulf/centos-8
sudo ./wwctl node add n0000.cluster --netdev eth0 -I 192.168.1.100 -M 255.255.255.0 -G 192.168.1.1 -H 00:0c:29:23:8b:48
```
    

#### Build the kernel, VNFS, and overlays:

Once you have added the node, you can start building the needed bootable components.
There are three major groups of data to provision:

1. Kernel: This is the boot kernel and driver overlay pair. The `wwctl kernel build`
command will create these files. The `-a` option will scan your configuration looking
for all needed kernel images and then go through and build them all. The caveat is
that all these kernels must be installed to your host controller node.

1. VNFS: The "VNFS" is the "Virtual Node File System", and that is the template that
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
sudo ./wwctl kernel build -a
sudo ./wwctl vnfs build -a
sudo ./wwctl overlay build -sa
```
    
#### Start the Warewulf daemon:

Once the above provisioning images are built, you can check the provisioning "rediness"
and then begin booting nodes.

```
sudo ./wwctl ready
./warewulfd
```
    
#### Boot your compute node and watch it boot

