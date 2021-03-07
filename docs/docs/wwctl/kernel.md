---
id: kernel
title: wwctl kernel
---

This command is for management of Warewulf Kernels to be used for bootstrapping nodes.

imprt
~~~~~
This will import a Kernel version from the control node into Warewulf for nodes to be configured to boot on.

-a, --all  Build all overlays (runtime and system)
-n, --node  Build overlay for a particular node(s)
-r, --root  Import kernel from root (chroot) directory
--setdefault  Set this kernel for the default profile

list, ls
~~~~~~~~
This command will list the kernels that have been imported into Warewulf.
