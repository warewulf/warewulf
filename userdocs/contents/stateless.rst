======================
Stateless Provisioning
======================

Why is Provisioning Important
=============================

Clusters are pools of servers bundled together to do a particular job or set of jobs. While there are a number of different use cases for clustering today, Warewulf was originally designed out of necessity.

Back in 2000, when Linux clustering was growing up for HPC, the issues of scale became apparent. Of course in HPC, there are many scalability factors which needed to be overcome as we continued to scale up clusters. Pretty early on was the "administrative scaling" which is the factor that full time systems administrators could only maintain so many servers. While homogenous configurations were able to help that, we still had the problem that every installed server became a point of administration, version creep, and debugging. The larger the cluster, the harder this problem was to solve.

Warewulf was created to help with exactly this.

Provisioning Overview
=====================

Provisioning in this definition is the process of putting an operating system onto a system. There are many ways to provision operating system images, from copying hard drives, to scripted installs, to automated installs. There are many valuable tools to facilitate this and they all helped to solve this problem.

In a cluster environment, this means one could group all of the nodes together, to be installed in bulk. Previous to cluster provisioning system administrators would go around to each cluster node, and install it from scratch, with an ISO or USB thumb drive. This obviously is not scalable. But being able to automatically install hundreds or thousands of computers in parallel and automate the management of these systems completely changed the paradigm.

There were several cluster provision and management toolkits already available when Warewulf was created and while these tools absolutely helped, there was even more optimization to be had. Stateless computing.

Stateless Provisioning
======================

The next step past automated installs is to just skip the installation completely; boot directly into the runtime operating system without ever doing an installation.

This is Warewulf.

Stateless provisioning is realizing you never have to install another compute node. Think of it like booting a LiveOS or LiveISO on nodes over the network. This means that no node individually is a point of system administration, but rather the entire cluster, inclusively is administrated as a single unit.

If all cluster nodes are booting the same OS image (or set of OS images), then any individual nodes that have problems is hardware. Debugging software and doing system administration in single points within a cluster is not needed. There is no version drift, because it is not possible for nodes to fall out of sync. Every reboot makes it exactly the same as its neighbors.

Warewulf provisions the operating system by default to system memory. There is no need for hard drives with Warewulf.

Previous versions of Warewulf had the ability to write the operating system to hard disk as well as do hybrid provisioning (the core operating system in memory, other pieces overlaid over NFS) but these have been obsoleted in Warewulf v4 as there are easier ways to accomplish the same thing (e.g. use of swap space).

> note: If you wish to provision to the hard drive, we might add that feature back based on user requests but in the mean time, you may wish to look at an automated or scripted installation platform instead of a cluster provisioning system.

In our experience, the Warewulf provisioning model is by far the most advantageous, simplest, and most flexible and scalable cluster provisioning platform available.