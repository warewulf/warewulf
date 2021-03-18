---
id: node
title: wwctl node
---

Management of node settings

add
~~~
This command will add a new node to Warewulf.

-g, --group  Group to add nodes to
-c, --controller  Controller to add nodes to
-N, --netdevDefine  the network device to configure
-I, --ipaddrSet  the node's network device IP address
-M, --netmaskSet  the node's network device netmask
-G, --gatewaySet  the node's network device gateway
-H, --hwaddrSet  the node's network device HW address
--discoverable  Make this node discoverable

console
~~~~~~~
Start IPMI console for a singe node.

delete
~~~~~~
This command will remove a node from the Warewulf node configuration.

-f, --force  Force node delete
-g, --group  Set group to delete nodes from
-c, --controller  Controller to add nodes to

list
~~~~
This command will show you configured nodes.

-n, --net  Show node network configurations
-i, --ipmi  Show node IPMI configurations
-a, --all  Show all node configurations
-l, --long  Show long or wide format

sensors
~~~~~~~
Show IPMI sensors for a single node.
-F, --full  show detailed output

set
~~~
This command will allow you to set configuration properties for nodes.

--comment  Set a comment for this node
-C, --container  Set the container (VNFS) for this node
-K, --kernel  Set Kernel version for nodes
-A, --kernelargs  Set Kernel argument for nodes
-c, --cluster  Set the node's cluster group
-P, --ipxe  Set the node's iPXE template name
-i, --init  Define the init process to boot the container
--root  Define the rootfs
-R, --runtime  Set the node's runtime overlay
-S, --system  Set the node's system overlay
--ipmi  Set the node's IPMI IP address
--ipminetmask  Set the node's IPMI netmask
--ipmigateway  Set the node's IPMI gateway
--ipmiuser  Set the node's IPMI username
--ipmipass  Set the node's IPMI password
-p, --addprofile  Add Profile(s) to node
-r, --delprofile  Remove Profile(s) to node
-N, --netdev  Define the network device to configure
-I, --ipaddr  Set the node's network device IP address
-M, --netmask  Set the node's network device netmask
-G, --gateway  Set the node's network device gateway
-H, --hwaddr  Set the node's network device HW address
--netdel  Delete the node's network device
--netdefault  Set this network to be default
-a, --all  Set all nodes
-y, --yes  Set 'yes' to all questions asked
-f, --force  Force configuration (even on error)
--discoverable  Make this node discoverable
--undiscoverable  Remove the discoverable flag
