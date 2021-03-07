---
id: architecture
title: Architecture
---

Warewulf's primary design goal is to facilitate software installation, configuration and management on the most common architecture for an HPC cluster nowadays. This popular cluster design originated 25 years ago with the seminal commodity-off-the-shelf (COTS) architecture commonly known as the "Beowulf".

## The Beowulf

The original [Beowulf Cluster](https://en.wikipedia.org/wiki/Beowulf_cluster) was developed in 1996 by Dr. Thomas Sterling and Dr. Donald Becker at NASA. The architecture is defined as a group of similar compute worker nodes all connected together using standard commodity equipment on a private network segment. The control node is dual homed (has two network interface cards) with one of these network interface cards attached to the upstream network and the other connected to the same private network which connects the compute worker nodes (as seen in the figure below).

.. image:: /images/beowulf_architecture.png

This architecture is very advantageous for creating a scalable HPC cluster resource, but it is an overly simple portrayal of the cluster ecosystem. To this thumbnail sketch we must add storage, scheduling and resource management, monitoring, interactive systems, and as the system grows, it may need to have groups of nodes with different features, architectures, vintages, memory configurations, GPUs or interconnects. As the system grows more complex, the need for a scalable control surface also increases which can result in even more complexity.

## Provisioning

Warewulf is designed to support the simplest of clusters, but also the most complicated, largest, and specialized resources that are in demand today. But when implementing all the various configurations, we rely on the basic architecture illustrated above. We can rely on such a simple schematic because Warewulf is designed to "own" the network broadcast domain and manage how these worker nodes boot.

As mentioned before, Warewulf is designed first and foremost to be a stateless provisioning subsystem. This means that the worker nodes maintain no state about their configuration, operating system or purpose when they are powered off. So, when these systems boot, they retain no knowledge of who they are, what they are supposed to do or how they got there. For these reasons, the worker nodes will need to boot via [PXE](https://en.wikipedia.org/wiki/Preboot_Execution_Environment) in order to retrieve a personality and a purpose.

## PXE

Most servers will have network interface cards that support PXE by default. In a nutshell, PXE will allow the network card to be seen by the BIOS as a bootable device. This means that the boot order may need to be configured in the system's BIOS on the worker nodes to allow it to boot. If there is also another bootable device on these systems, it might be necessary to set the network interface card to boot first.

When the system boots via PXE, it will begin a chain reaction of events:

1. The network card will register an option ROM into the BIOS
2. The BIOS will run through all of its functions and finish with boot devices
3. The boot devices are attempted to be booted in the defined order
4. When it gets to the network boot device, PXE is run from the firmware on the network card
5. PXE will request a BOOTP/DHCP address on the network which is handled by the controller node
6. If the DHCP response includes a boot file name, it will download this file (iPXE boot image) over TFTP
7. Once iPXE is downloaded, it is loaded by the network card and it will request a configuration from the 
   controller node
8. The configuration will tell iPXE what to download and load

Warewulf manages the entire process once the worker node's network device has begun the PXE process.

## Warewulf Server

Warewulf will configure the controller's DHCP and TFTP services and put all the required files into place. But once the iPXE file has been sent and loaded on the worker nodes, all network communication from this point on is handled by the Warewulf server on the controller node over HTTP. The order of events from this point on is as follows:

1. iPXE requests its configuration and Warewulf generates this on demand from the configured template
1. The default iPXE template tells iPXE to request a kernel, VNFS image, runtime kernel modules and a system overlay
1. Each of the requested files will be sent to the worker node

## Post PXE

Once the worker node has received all the required files, the kernel will boot, and the runtime components will be loaded into the kernel's initial RAM file system (initramfs). The order of these processes is important because each layer is dependent on the previous. The layers are implemented in the following order:

1. VNFS/Container image
1. Kernel modules
1. System Overlay
1. Runtime Overlay (repeated at a given interval)

As part of the provisioning process, the system boots after the System Overlay has been provisioned, which means that the Runtime Overlay occurs after `/sbin/init` has been called on the system. This delineation is important because it clearly defines what should be in the system versus runtime overlay. More on overlays in the next section.

Before calling `/sbin/init` (or `init` to override in the node configuration file), Warewulf must set up the system. Again, this process occurs after the System Overlay has been provisioned. This is where the system initializes and prepares for booting the runtime OS. Depending on the configuration, the system might boot directly in the initramfs file system as it stands, or it could migrate the root file system to a different mount point (e.g. `tmpfs`or at some point, hard drives).

Other things that get done at this stage are setting up and enabling SELinux, IPMI, making any needed changes or configurations to the file system before booting it and starting up the `wwclient` which is responsible for loading the runtime overlay.

Lastly, the `init` process is executed with PID 1 and thus "boots" the VNFS container. This is where SysVInit, Upstart, or Systemd takes over.

## Warewulf Overlays

As described above, Warewulf uses layers to provision worker nodes in phases. The first layer is static across any number of nodes, but each node may require some custom configurations, for example, network. This means there must be a method for leveraging a base file system (VNFS) that can be shared by many nodes and also be able to configure these file systems with custom per-node options at a large scale.

Typically, there are two major times this configuration needs to be done-- pre "boot" and post "boot". In this case, as described above, we can consider the call to `/sbin/init` the delineation point and the proper way to consider the two configurable overlays: system and runtime.

**System Overlay**: The system overlay is what will be present before `/sbin/init` is called. This gives the administrator the ability to control the configuration of the booting system itself. For example, network configuration must be addressed on every node, but each node must have a slightly different network configuration otherwise the IP addresses will clash. This must be set before Systemd brings up the network device, so the Warewulf system overlay is the right place to configure this.

**Runtime Overlays**: Some configurations happen after the system boots and continuously at periodic intervals. For example, user and group accounts. You probably don't want to reprovision a node, let alone hundreds of nodes, to add a user or change a runtime configuration, and this is where you should use the runtime overlay.

Both overlays leverage a similar file system template structure. Each overlay (you can create any number of them) can include text files, directories, links and templates. Templates allow you to dynamically customize any of the content within an overlay for each node that will be leveraging that template.
