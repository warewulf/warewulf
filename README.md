# warewulf

This is built on CentOS-7. More needs to be done to make it work on other
distributions and versions specifically with the system service
compoenents.

In a nutshell, to use:

1. make install
2. build VNFS `sudo singularity build --sandbox /var/chroots/centos-7 centos-7.def`
3. Edit `/etc/warewulf/nodes.yaml`
4. Run the following commands:
   a. `./wwbuild vnfs`
   b. `./wwbuild kernel`
   c. `./wwbuild overlay`
   c. `./warewulfd`
5. Boot your node and watch the console and the output of the Warewulfd process
