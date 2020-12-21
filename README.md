# warewulf v4 (work in progress)

![Warewulf](warewulf-logo.png)

In a nutshell, to install and start provisioning nodes, do the following:

#### Build Warewulf and dependencies:

Warewulf is programmed in GoLang, so you will need to also install a Go compiler
on your system. The easiest way to do this on RHEL and CentOS is by using the Go
packages that are included in EPEL. In addition there are some dependencies that
Warewulf will require to operate properly and because this is a quick HOWTO, we
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
sudo ./wwctl configure -a
```

note, at the time of this writing, there are additional services which are not included
in the `-a` option. Please do a `wwctl configure --help` to see all configurable services.

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
sudo ./wwctl container pull docker://warewulf/centos-7 centos-7 --setdefault
sudo ./wwctl kernel build $(uname -r) --setdefault
```

#### Set up the default node profile

The `--setdefault` arguments above will automatically set those entries in the default
profile, but if you wanted to set them by hand to something different, you can do the
following:

```
sudo ./wwctl profile set default -K $(uname -r) -C centos-7
```

Next we set some default networking configurations for the first ethernet device. On
modern Linux distributions, the name of the device is not critical, as it will be setup
according to the HW address. Because all nodes will share the netmask and gateway, we
can configure them in the default profile as follows:

```
sudo ./wwctl profile set default --netdev eth0 -M 255.255.255.0 -G 192.168.1.1
sudo ./wwctl profile list
```
    
#### Add a node and build node specific overlays

Adding nodes can be done while setting configurations in one command. Here we are setting
the IP address of `eth0` and setting this node to be discoverable, which will then
automatically have the HW address added to the configuration as the node boots.

Node names must be unique. If you have node groups and/or multiple clusters, designate
them using dot notation.

Note that the full node configuration comes from both cascading profiles and node
configurations which always supersede profile configurations.

```
sudo ./wwctl node add n0000.cluster --netdev eth0 -I 192.168.1.100 --discoverable
sudo ./wwctl node list -a n0000
```

### Warewulf Overlays

There are two types of overlays: system and runtime overlays.

System overlays are provisioned to the node before `/sbin/init` is called. This enables us
to prepopulate node configurations with content that is node specific like networking and
service configurations. When using the overlay subsystem, system overlays are never shown
by default. So when running `overlay` commands, you are always looking at runtime overlays
unless the `-s` option is passed.

Runtime overlays are provisioned after the node has booted and periodically during the
normal runtime of the node. Runtime overlays are also obtained using privileged source
ports such that non-root users can not obtain this from the Warewulf service (note: there
are other ways to secure the provisioned files like a provisioning VLan). Because these
overlays are provisioned at periodic intervals, they are very useful for content that
changes, like users and groups.

Overlays are generated from a template structure that is viewed using the `wwctl overlay`
commands. Files that end in the `.ww` suffix are templates and abide by standard
text/template rules. This supports loops, arrays, variables, and functions making overlays
extremely flexible.

All overlays are compiled before being provisioned. This accelerates the provisioning
process because there is less to do when nodes are being managed at scale.

Here are some of the common `overlay` commands:

```
sudo ./wwctl overlay list -l
sudo ./wwctl overlay list -ls
sudo EDITOR=vim ./wwctl overlay edit default /etc/hello_world.ww
sudo ./wwctl overlay build -a
```

#### Start the Warewulf daemon:

Once the above provisioning images are built, you can check the provisioning "rediness"
and then begin booting nodes.

```
sudo ./wwctl ready
sudo ./wwctl server start
sudo ./wwctl server status
```
    
#### Boot your compute node and watch it boot

