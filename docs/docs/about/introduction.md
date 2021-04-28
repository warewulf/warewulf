---
id: introduction
title: Introduction
slug: /
---

## Summary

Warewulf is an operating system provisioning system for Linux that is designed to produce simple, turnkey HPC deployment solutions that maintain flexibility and configurability at scale.

Since its initial release in 2001, Warewulf has become the most popular open source and vendor-agnostic provisioning system within the global HPC community. Warewulf is known for its massive scalability and simple management of stateless (disk optional) provisioning.

In a nutshell, cluster node operating system images are containers. These containers are built into a bootable format called a "Virtual Node File System" (VNFS) image which is provisioned out to nodes when they boot. A VNFS and kernel pair can be distributed to any number of nodes, which means you could have all of your nodes booting the exact same container VNFS image. To avoid the administrative headache of too many customized VNFS images, subtle differences between node operating system configurations are handled with "overlays". 

On boot, each node receives, in the following order:

1. iPXE boot image (tftp)
1. Linux kernel (http)
1. VNFS image (http)
1. Kernel module overlay (http)
1. System overlay (http)

Once the node has booted, it will request a runtime overlay (privileged http) at a periodic interval.

All the images, which are transferred for provisioning, are prebuilt static components and the per node connection processing is minimal. This means that Warewulf is as scalable as the physical infrastructure.

## Warewulf Design Tenants

To enable simple, scalable and flexible operating system management at scale.

- **Lightweight**: Warewulf needs to do its job and stay out of the way. There should be no underlying system dependencies, changes or "stack" for the controller or worker nodes.
   
- **Simple**: Warewulf is used by hobbyists, researchers, scientists, engineers and systems administrators. This means that Warewulf must be simple to use and understand.
   
- **Flexible**: Warewulf is highly flexible and can address the needs of any environment-- from a computer lab with graphical workstations, to under-the-desk clusters, to massive supercomputing centers providing traditional HPC capabilities to thousands of users.
   
- **Agnostic**: From the Linux distribution to the underlying hardware, Warewulf should be as agnostic and standards compliant as possible. From ARM to x86, Atos to Dell, Warewulf can provision it all equally well with no favorites.
   
- **Open Source**: It is imperative that Warewulf be and remain absolutely Open Source.
