

    Instructions for setting up warewulf development environment under KVM

1. Create Centos 7 development virtual machine under KVM

    # KVM is running on server called master1 which is not my desktop

    ssh -X master1

    # On master1 server

    wget -P /global/downloads/centos http://mirror.mobap.edu/centos/7.8.2003/isos/x86_64/CentOS-7-x86_64-Everything-2003.iso

    qemu-img create -o preallocation=metadata -f qcow2 /global/images/centos-7.qcow2

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



2. Turn off default network dhcp on server master1

    # shutdown all VMs

    $ sudo virsh net-destroy default

    $ sudo virsh net-edit default

        # remove dhcp lines from XML

    $ sudo virsh net-start default


3. Build and install warewulf on wwdev

    ssh wwdev


    # Fedora prerequisites
    sudo dnf -y install tftp-server tftp
    sudo dnf -y install dhcp
    sudo dnf install singularity

    # Centos prerequisites
    sudo yum -y install tftp-server tftp
    sudo yum -y install dhcp
    sudo yum install http://repo.ctrliq.com/packages/rhel7/ctrl-release.rpm
    sudo yum install singularityplus

    # follow README.md instructions

    cd projects/ctrliq/warewulf

    vi nodes.yaml.local

    make -f Makefile.local install

    # build VNFS

    sudo singularity build --sandbox /global/chroots/centos-7 centos-7.def

    vi /etc/warewulf/overlays/generic/etc/sysconfig/network-scripts/ifcfg-* 

        # add IP addresses

    sudo ./wwbuild vnfs

    sudo ./wwbuild kernel

    sudo ../wwbuild overlay

    sudo ./warewulfd

4. Boot your node and watch the console and the output of the Warewulfd process

~



