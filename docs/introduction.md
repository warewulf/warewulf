# Introduction
Warewulf is an operating system provisioning system for Linux. It was created in 2001 to facilitate and scale the 
administration of HPC clusters in a way that allowed for simple, turn key solutions as well as support flexibility 
and configurability at scale.

Since its initial release, Warewulf has become the most popular open source and vendor-agnostic provisioning system 
within the HPC ecosystem with worldwide adoption. Warewulf is known for its massive scalability and ease of use for
stateless (disk optional) provisioning.

In a nutshell, node operating system images are containers. These containers are built into a bootable format called 
a "Virtual Node File System" (VNFS) image which is provisioned out to nodes at boot. The VNFS and kernel can be 
configured for any number of nodes which means you could have all of your nodes booting the exact same container 
VNFS image. The subtle differences in operating system configurations between the nodes are handled in "overlays".

On boot, each node receives, in the following order:

1. iPXE boot image (tftp)
1. Linux kernel (http)
1. VNFS image (http)
1. Kernel module overlay (http)
1. System overlay (http)

*note: once the node has booted, it will request a runtime overlay (privileged http) at a periodic interval.*

All the images which are transferred for provisioning are prebuilt static components and the per node connection 
processing is minimal. This means that Warewulf is as scalable as the physical infrastructure.

### Warewulf Design Tenants
To enable simple, scalable, and flexible operating system management at scale.

1. **Lightweight**: Warewulf needs to do it's job and stay out of the way. There should be no underlying system 
   dependencies, changes, or "stack" for the controller or worker nodes.
   
1. **Simple**: Warewulf is used by hobbyists, researchers, scientists, as well as systems admnistrators and engineers.
   This means that Warewulf must be simple to use and intuitive to understand.
   
1. **Flexible**: Warewulf must remain highly flexible to be able to fit in any environment, from a computer lab with 
   graphical workstations, to an under-desk clusters, as well as massive centers providing traditional HPC 
   capabilities to thousands of users.
   
1. **Agnostic**: From the underlying Linux distribution to the underlying hardware, Warewulf should be as agnostic 
   and standards compliant as possible. From ARM to X86, Atos to Dell, Warewulf can provision it all equally well 
   with no favorites.
   
1. **Open Source**: It is imperative that Warewulf be and remain absolutely Open Source.
