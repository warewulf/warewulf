============================
Using Warewulf on VMWare
============================


Sample Network Configuration
================

* Master node has 2 virtual NICs
** Public NIC with public IP
** Private NIC on private 10 subnet (In this case 10.85.0.0/16)
*** IP: 10.85.0.1
*** Subnet: 255.255.0.0

* Slave Nodes has 1 virtual NIC
** On same private 10 subnet
*** Slave Nodes to use IP pool 10.85.1.1 - 10.85.1.255


VMs won't ipxe boot
================
Make sure that "secure boot" for EFI is turned off (under Settings -> VM Options -> Boot Options)

Some users report that EFI must be disabled and instead set to BIOS, while other users have success with EFI and disabling 'secure boot'


VMs get wrong IP or has other DHCP issues
================
There is an issue where some versions of VMWare use their own DHCP server, even on the private network. This must be disabled. See https://knowledge.broadcom.com/external/article/311759/modifying-the-dhcp-settings-of-vmnet1-an.html



Issues with 'Failed to allocate memory for files'
================
Solution is to use Dracut (e.g. /contents/boot-management.html#booting-with-dracut)