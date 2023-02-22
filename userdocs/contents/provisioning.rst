=================
Node Provisioning
=================

Once the nodes are configured in Warewulf, they are ready to boot.

Node Hardware Setup
===================

The only thing that Warewulf requires to provision is that the node is
set to PXE boot. You may need to change the boot order if there is a
local disk present and bootable. This is a configuration change you
will have to make in the BIOS of the cluster node.

Each vendor does this differently and as a result we won't go into the
setup specifics here and if you can not find information on how to PXE
boot your nodes, please contact your hardware vendor support.

.. note::

   If you find that you are going to use Warewulf, or any other
   cluster provisioning tool, it is very helpful to require that
   hardware vendors preconfigure your cluster nodes with values of
   your choosing, and ask them to provide a text file that includes
   all of the HW/MAC addresses of the compute nodes in the order they
   are racked (which most creditable vendors will do). You can also
   ask them to certify their computing stack for the operating system
   you wish to use and the provisioning system. This helps hardware
   vendors to ensure their stack works with open source projects like
   Warewulf, Debian, OpenSuSE, and Rocky Linux.

The Provisioning Process
========================

When the cluster node boots, the following order of operations will
occur:

#. BIOS:
    #. The system BIOS will bootstrap the initialization of the
       hardware
    #. The network card will register its option ROM into the BIOS
    #. The BIOS will run through all of its functions and finish with
       boot devices
    #. The boot devices are attempted in order
    #. When it gets to the network boot device, PXE is run from the
       firmware on the network card
#. PXE:
    #. PXE will request a BOOTP/DHCP address on the network
    #. The Warewulf controller's DHCP server will respond with a
       network configuration and filename to try and boot
    #. PXE will attempt to download the filename referred to in the
       DHCP response via TFTP
    #. The downloaded file will execute an iPXE stack which will reach
       out to the Warewulf server for it's configuration
#. Bootstrap:
    #. The Warewulf server will generate the iPXE configuration which
       will include directions of what else is necessary to download
       and how to boot.
    #. The kernel, container image, kernel modules, and system overlay
       are all downloaded over REST HTTP from the Warewulf Server
    #. iPXE executes the kernel and processes the overlays to provide
       a unified root file system
    #. Warewulf bootstraps the initialization of cluster node's
       operating system
        #. File System (re)configuration
        #. SELinux
        #. ``wwclient`` is called as a background daemon and sleeps
           until network is ready
    #. The Warewulf bootstrap execs the container's ``/sbin/init``
#. Container:
    #. The container now boots exactly as any operating system would
       expect
