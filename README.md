# warewulf v4 WIP

This is built on CentOS-7. More needs to be done to make it work on other
distributions and versions specifically with the system service
components.

In a nutshell, to install and start provisioning nodes, do the following:

1. `make`
2. `sudo make install`
3. `vi /etc/warewulf/warewulf.conf`
4. `vi /etc/warewulf/nodes.conf`
5. `sudo singularity build --sandbox /var/chroots/centos-7 centos-7.def`
6. `sudo ./wwbuild vnfs`
7. `sudo ./wwbuild kernel`
8. `sudo ./wwbuild overlay`
9. `./warewulfd`
10. Boot your compute node

