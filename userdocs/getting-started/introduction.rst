============
Introduction
============

Warewulf is an operating system provisioning platform for Linux clusters. Since
its initial release in 2001, Warewulf has become the most popular open source
and vendor-agnostic provisioning system within the global HPC community.
Warewulf is known for its massive scalability and simple management of stateless
(disk optional) provisioning.

Warewulf leverages a simple administrative model centralizing administration
around virtual node images which are used to provision out to the cluster nodes.
This means you can have hundreds or thousands of cluster nodes all booting and
running on the same node image. As of Warewulf v4, the node image can be managed
using industry-standard container tooling and/or CI/CD pipelines. This can be as
simple as DockerHub or your own private GitLab CI infrastructure. With this
architecture, Warewulf combines the best of High Performance Computing (HPC),
Cloud, Hyperscale, and Enterprise deployment principals to create and maintain
large scalable stateless clusters.

Warewulf is used most prominently in High Performance Computing (HPC) clusters,
but its architecture is flexible enough to be used in most any clustered Linux
environment, including clustered web servers, rendering farms, and even
Kubernetes and cloud deployments.

Warewulf design
===============

Warewulf has had a number of iterations since its inception in 2001, but its
design tenets have always remained the same: a simple, scalable, stateless, and
flexible provisioning system for all types of clusters.

* **Lightweight**: Warewulf provisions stateless operating system images and
  then gets out of the way. There are no underlying system dependencies or
  requisite changes to the provisioned cluster node operating system.

* **Simple**: Warewulf is used by hobbyists, researchers, scientists, engineers
  and systems administrators alike.

* **Flexible**: Warewulf can address the needs of any environment--from a
  computer lab with graphical workstations, to under-the-desk clusters, to
  supercomputing centers providing HPC services to thousands of users.

* **Agnostic**: From the Linux distribution of choice to the underlying
  hardware, Warewulf is agnostic and standards compliant. From ARM to x86, Atos
  to Dell, Debian, SUSE, Rocky, CentOS, and RHEL, Warewulf can be used in most
  any environment.

* **Secure**: Warewulf support SELinux out-of-the-box. Just install SELinux in
  your node image and let Warewulf do the rest!

* **Open Source**: Warewulf is and has always been open source. It can be used
  in any environment, whether public, private, non-profit, or commercial. And
  the Warewulf project is always welcoming of contribution from its community of
  users, with major features often beginning as external contributions.

Warewulf architecture
=====================

Warewulf v4 has a simple but flexible base architecture:

A **Warewulf server** stores information about the cluster and the nodes in it,
and provides a command-line interface (``wwctl``) for managing nodes, their
images, and their overlays.

**Cluster nodes** are defined in a flexible `YAML
<https://en.wikipedia.org/wiki/YAML>`_ file, including their network
configuration and image and overlay assignments.

**Node profiles** provide a flexible abstraction for applying configuration to
multiple nodes.

**Node images** provide a bootable operating system image, including the kernel
that will be used to boot the cluster node. Node images provide a base operating
system and, by default, run entirely in memory. This means that when you
reboot the node, the node retains no information about Warewulf or how it
booted; but it also means that they return to their initial known-good state.

**Overlays** customize the provisioned operating system image with static files
and dynamic templates applied with the node image and, optionally, periodically
at runtime.

Beowulf overview
================

Warewulf is designed to support the original `Beowulf Cluster
<https://en.wikipedia.org/wiki/Beowulf_cluster>`_ concept. (Thus its name, a
soft\ **WARE** implementation of the beo\ **WULF**.) The architecture is
characterized by a group of similar cluster nodes all connected together using
standard commodity equipment on an internal cluster network. The server node
(often historically referred to as the "master" or "head" node) is "dual homed"
(i.e., it has two network interfaces) with one of these network interfaces
attached to an external network and the other connected to the internal cluster
network.

.. image:: beowulf_architecture.png
    :alt: Beowulf architecture

This simple topology is the foundation for creating a scalable HPC cluster
resource. Even today, almost 30 years after the inception of this architecture,
this is the baseline architecture that virtually all HPC systems are built to.

An HPC cluster often includes dedicated storage, scheduling and resource
management, monitoring, interactive systems, and other components. For smaller
systems, many of these components can be deployed to a single server node; but,
as the system scales, it may be better to have groups of nodes dedicated to
these different services.

Warewulf is flexible enough to start with a simple "head node" Beowulf style
cluster deployment and to grow as needs for the cluster and its environment
change.
