
====================================
Development Environment (VirtualBox)
====================================

I have VirtualBox running on my desktop.

1. Create a NAT Network (a private vlan) to be used for the Warewlf Server and compute nodes inside the VirtualBox. Make sure to turnoff DHCP service within this NAT Network.

.. code-block:: console

   # On the host with VirtualBox execute below. In my example using 10.0.8.0/24 as the private vlan for my experiment with Warewulf
   
   VBoxManage natnetwork add --netname wwnatnetwork --network "10.0.8.0/24" --enable --dhcp off

2. Create a Centos 7 development Virtual machine (wwdev) to be used as the Warewulf Server. Enable two Network adapters one with a standard NAT and SSH port mapping such that you can access this VM from the host machine. Assign the second network adapter to the NAT Network created in step #1. Assign sufficient memory (e.g: 4GB) to the VM. 

.. code-block:: console

   # Download a Centos7 or SL7 ISO and mount it to the optical drive to boot and install OS for the wwdev VM.
   # Attach Network adapter #1 of the wwdev VM to the standard NAT via VM Settings -> Network option. 
   # By default VirtualBox puts the Network Adapter into 10.0.2.0/24 network and assigns 10.0.2.15 IP address.
   
   # Also add a rule to the port forwarding table under the standard NAT configuration to allow SSH 
   # from localhost (127.0.0.1) some high port e.g 2222 to the guest IP 10.0.2.15 port 22 such that      
   # you can SSH from your host/desktop to the wwdev VM. 
   
   # Next attach the second Network adapter #2 to the NAT Network and you should be able to choose 
   # the 'wwnatnetwork' created above in step #1 from the drop down list.

3. Build and install warewulf on wwdev

.. code-block:: console
   
   # Login to wwdev VM and install @development group and go language
   
   ssh localhost -p 2222 #(should prompt for a user account password on wwdev VM)
   
   # Disable selinux by modifying /etc/sysconfig/selinux
   vi /etc/sysconfig/selinux
   
       SELINUX=disabled
   
   # Disable firewall
   systemctl stop firewalld
   systemctl disable firewalld
   
   # Centos prerequisites
   sudo yum -y install tftp-server tftp
   sudo yum -y install dhcp
   sudo yum -y install ipmitool
   sudo yum install http://repo.ctrliq.com/packages/rhel7/ctrl-release.rpm
   sudo yum install singularityplus
   sudo yum install gpgme-devel
   sudo yum install libassuan.x86_64 libassuan-devel.x86_64
   
   # Upgrade git to v2+
   sudo yum install https://packages.endpoint.com/rhel/7/os/x86_64/endpoint-repo-1.7-1.x86_64.rpm
   sudo yum install git
   sudo yum install golang
   sudo yum install nfs-utils
   
   # Install Warewulf and dependencies
   git clone https://github.com/hpcng/warewulf.git
   cd warewulf
   
   make all
   sudo make install
   
   # Static assign an IP to adapter #2 which is in the wwnatnetwork.
   $ Edit the file /etc/sysconfig/networking-scripts/ifcfg-enp0s9 # adapter name at the end might be different for you
   # Add lines like to below to assign an ip in 10.0.8.0/24 wwnatnetwork, I choose 10.0.8.4
   BOOTPROTO=static
   ONBOOT=yes
   NAME=enp0s9
   DEVICE=enp0s9
   IPADDR=10.0.8.4
   NETMASK=255.255.255.0
   GATEWAY=10.0.8.1
   # Bring the enp0s9 interface online and verify ip assignment
   
   # Configure the Warewulf controller
   $ Edit the file /etc/warewulf/warewulf.conf and ensure that you've set the approprite configuration parameters. 
   # My conf file looks like below:
       ipaddr: 10.0.8.4
       netmask: 255.255.255.0
       warewulf:
         port: 9873
         secure: true
         update interval: 60
       dhcp:
         enabled: true
         range start: 10.0.8.150
         range end: 10.0.8.200
         template: default
         systemd name: dhcpd
       tftp:
         enabled: true
         tftproot: /var/lib/tftpboot
         systemd name: tftp
       nfs:
         systemd name: nfs-server
         exports:
         - /home
         - /var/warewulf
   
   # Configure system service automatically
   sudo wwctl configure dhcp --persist # Create the default dhcpd.conf file and start/enable service
   sudo wwctl configure tftp --persist # Install the base tftp/PXE boot files and start/enable service
   sudo wwctl configure nfs  --persist # Configure the exports and create an fstab in the default system overlay
   sudo wwctl configure ssh  --persist # Build the basic ssh keys to be included by the default system overlay
   
   # Pull and build the VNFS container and kernel
   sudo wwctl container import docker://warewulf/centos-7 centos-7 --setdefault
   sudo wwctl kernel import build $(uname -r) --setdefault
   
   # Set up the default node profile
   sudo wwctl profile set default -K $(uname -r) -C centos-7
   sudo wwctl profile set default --netdev eth0 -M 255.255.255.0 -G 10.0.8.4
   sudo wwctl profile list
   
   # Add a node and build node specific overlays
   # IP address of my nodes start from 150 as set in the warewulf.conf file above
   sudo wwctl node add n0000.cluster --netdev eth0 -I 10.0.8.150 --discoverable
   sudo wwctl node list -a n0000
   
   # Review Warewulf overlays
   sudo wwctl overlay list -l
   sudo wwctl overlay list -ls
   sudo wwctl overlay edit default /etc/hello_world.ww
   sudo wwctl overlay build -a
   
   # Start the Warewulf daemon
   sudo wwctl ready
   sudo wwctl server start
   sudo wwctl server status

4. Create a new guest VM instance inside the VirtualBox to be the warewulf client/compute node. Under the system configuration make sure to select the optical and network options only for the boot order. The default iPXE used by VirtualBox does not come with bzImage capability which is needed for warewulf. Download the ipxe.iso available at ipxe.org and mount the ipxe.iso to the optical drive. Enable one Network adapter for this VM and assign it to the NAT Network created in step #1 above. 

.. code-block:: console

   # Download ipxe.so available at http://boot.ipxe.org/ipxe.iso
   # VM Settings -> System disable Floppy, Hard Disk from Boot order. Enable Optical and Network options.
   # VM Settings -> Storage and mount the above download ipxe.so to the Optical Drive.
   # VM Settings -> Network Enable adapter #1, attach to 'Nat Network' and choose 'wwnatnetwork' from the drop down list.

Boot your node and watch the console and the output of the Warewulfd process