=============================
Development Environment (KVM)
=============================

Create CentOS 7 development virtual machine under KVM
=====================================================

.. code-block:: bash

   # KVM is running on server called master1 which is not my desktop
   ssh -X master1
   
   # On master1 server
   wget -P /global/downloads/centos http://mirror.mobap.edu/centos/7.8.2003/isos/x86_64/CentOS-7-x86_64-Everything-2003.iso
   
   qemu-img create -o preallocation=metadata -f qcow2 /global/images/centos-7.qcow2 32G
   
   # install wwdev Centos 7 development VM
   sudo virt-install --virt-type kvm --name centos7-wwdev --ram 8192 \
      --disk /global/images/centos-7.qcow2,format=qcow2 \
      --network network=default \
      --graphics vnc,listen=0.0.0.0 --noautoconsole \
      --os-type=linux --os-variant=rhel7.0 \
      --location=/global/downloads/centos/CentOS-7-x86_64-Everything-2003.iso
   
   # Complete installation using virt-manager
   
   # To start virt-manager on non-local server
   ssh -X master1
   
   sudo -E virt-manager
   
   # Login to VM and install @development group and go language
   ssh root@wwdev
   
   # Disable selinux by modifying /etc/sysconfig/selinux
   vi /etc/sysconfig/selinux
   
       SELINUX=disabled
   
   # disable firewall
   systemctl stop firewalld
   systemctl disable firewalld

Turn off default network dhcp on server master1
===============================================

.. code-block:: bash

   # shutdown all VMs
   sudo virsh net-destroy default
   
   sudo virsh net-edit default
   
       # remove dhcp lines from XML
   
   sudo virsh net-start default

Build and install warewulf on wwdev
===================================

.. code-block:: bash

   ssh wwdev
   
   # Fedora prerequisites
   sudo dnf -y install tftp-server tftp
   sudo dnf -y install dhcp
   sudo dnf -y install ipmitool
   sudo dnf install singularity
   sudo dnf install gpgme-devel
   sudo dnf install libassuan.x86_64 libassuan-devel.x86_64
   sudo dnf golang
   sudo dnf nfs-utils
   
   # Centos prerequisites
   sudo yum -y install tftp-server tftp
   sudo yum -y install dhcp
   sudo yum -y install ipmitool
   sudo yum install http://repo.ctrliq.com/packages/rhel7/ctrl-release.rpm
   sudo yum install singularityplus
   sudo yum install gpgme-devel
   sudo yum install libassuan.x86_64 libassuan-devel.x86_64
   sudo yum install https://packages.endpoint.com/rhel/7/os/x86_64/endpoint-repo-1.7-1.x86_64.rpm
   sudo yum install golang 
   sudo yum install nfs-utils
   
   # Install Warewulf and dependencies
   git clone https://github.com/hpcng/warewulf.git
   cd warewulf
   
   make all
   sudo make install
   
   # Configure the controller
   Edit the file /etc/warewulf/warewulf.conf and ensure that you've ser the approprite configuration parameters
   
   # Configure system service automatically
   sudo wwctl configure dhcp # Create the default dhcpd.conf file and start/enable service
   sudo wwctl configure tftp # Install the base tftp/PXE boot files and start/enable service
   sudo wwctl configure nfs  # Configure the exports and create an fstab in the default system overlay
   sudo wwctl configure ssh  # Build the basic ssh keys to be included by the default system overlay
   
   # Pull and build the VNFS container and kernel
   sudo wwctl container import docker://warewulf/centos-8 centos-8 --setdefault
   sudo wwctl kernel import build $(uname -r) --setdefault
   
   # Set up the default node profile
   sudo wwctl profile set default -K $(uname -r) -C centos-7
   sudo wwctl profile set default --netdev eth0 -M WW_server_subnet_mask -G WW_server_ip
   sudo wwctl profile list
   
   # Add a node and build node specific overlays
   sudo wwctl node add n0000.cluster --netdev eth0 -I n0000_ip --discoverable
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

Boot your node and watch the bash and the output of the Warewulfd process