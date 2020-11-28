# warewulf v4 WIP

This is built on CentOS-7/8. More needs to be done to make it work on other
distributions and versions specifically with the system service
components.

In a nutshell, to install and start provisioning nodes, do the following:

1. Build Warewulf and dependencies:

```
make all
```
    
1. Install Warewulf onto your host system:

```
sudo make install
```
    
1. Set the master's kernel and default VNFS into your "default" node profile:

```
sudo ./wwctl profile set -K $(uname -r) -V docker://warewulf/centos-8 default
```
    

1. Build the kernel, VNFS, and overlays:

```
sudo ./wwctl kernel build -a
sudo ./wwctl vnfs build -a
sudo ./wwctl overlay build -sa
```
    
1. Start the Warewulf daemon:

```
./warewulfd
```
    
1. Boot your compute node and watch it boot

