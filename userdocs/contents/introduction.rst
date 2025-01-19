============
Introduction
============

The Warewulf Vision
===================

Warewulf has had a number of iterations over the last 20 years, but
its design tenets have always remained the same: a simple, scalable,
stateless (however some versions were able to provision stateful), and
very flexible operating system provisioning system for all types of
clusters. This is an overview of how Warewulf works.

About Warewulf
==============

Warewulf is an operating system provisioning platform for Linux that
is designed to produce secure, scalable, turnkey cluster deployments
that maintain flexibility and simplicity.

Since its initial release in 2001, Warewulf has become the most
popular open source and vendor-agnostic provisioning system within the
global HPC community. Warewulf is known for its massive scalability
and simple management of stateless (disk optional) provisioning.

Warewulf leverages a simple administrative model centralizing
administration around virtual node images which are used to provision
out to the cluster nodes. This means you can have hundreds or
thousands of cluster nodes all booting and running on the same,
identical virtual node file system image. As of Warewulf v4, the
virtual node image can be managed using any existing
container tooling and/or CI pipelines that are being used. This can be
as simple as DockerHub or your own private GitLab CI infrastructure.

With this architecture, Warewulf combines the best of High Performance
Computing (HPC), Cloud, Hyperscale, and Enterprise deployment
principals to create and maintain large scalable stateless clusters.

While Warewulf's roots are in HPC, it has been used for many different
types of tasks and use cases. Everything from clustered web servers,
to rendering farms, and even Kubernetes and cloud deployments,
Warewulf brings benefit in experience of general operating system
management at scale.

Features
========

* **Lightweight**: Warewulf provisions stateless operating system
  images and then gets out of the way. There should be no underlying
  system dependencies or changes to the provisioned cluster node
  operating systems.

* **Simple**: Warewulf is used by hobbyists, researchers, scientists,
  engineers and systems administrators because it is easy,
  lightweight, and simple.

* **Flexible**: Warewulf is highly flexible and can address the needs
  of any environment-- from a computer lab with graphical
  workstations, to under-the-desk clusters, to massive supercomputing
  centers providing traditional HPC capabilities to thousands of
  users.

* **Agnostic**: From the Linux distribution of choice to the
  underlying hardware, Warewulf is agnostic and standards
  compliant. From ARM to x86, Atos to Dell, Debian, SUSE, Rocky,
  CentOS, and RHEL, Warewulf can do it all.

* **Secure**: Warewulf is the only stateless provisioning system that
  will support SELinux out of the box. Just enable your node image to
  support SELinux, and let Warewulf do the rest!

* **Open Source**: For the last 20 years, Warewulf has remained open
  source and continues to be the golden standard for cluster
  provisioning.
