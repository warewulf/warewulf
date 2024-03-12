========
Glossary
========

Container
    Warewulf containers are the node images that it manages and provisions.
    The use of the term "container" alludes to Warewulf's support for importing OCI containers, OCI container archives, and Apptainer sandboxes to initialize its node images.

    Warewulf containers are maintained as an uncompressed "virtual node file system" or VNFS, (sometimes also referred to as a "chroot").
    These containers are then built as images which may then be used to provision a node.

    It is important to note, however, that Warewulf does not provision virtualized or nested "containers" in the common sense;
    Warewulf nodes run a decompressed image on "bare metal" loaded directly into system memory.

Controller
    The Warewulf controller runs the Warewulf daemon (``warewulfd``) and is responsible for the management, control, and administration of the cluster.
    This system may also sometimes be referred to as the "master," "head," or "admin" node.

    A typical Warewulf controller also runs a DHCP service and a TFTP service, and often an NFS service;
    though these services may be managed separately and on separate servers.

Kernel
    In addition to a container, Warewulf also requires a kernel (typically a Linux kernel) in order to provision a node.

    Kernels may be imported independently into Warewulf, either from the controller or from a container;
    however, recent versions of Warewulf (after v4.3.0) support automatically provisioning with a kernel detected and extracted from the container itself.
    In most cases, kernels may be installed in the container using normal system packages, and no special consideration is necessary.

Node
    Warewulf nodes are the systems that are being provisioned by Warewulf.
    The roles of these systems could be "compute", "storage", "GPU", "IO", etc.

nodes.conf
    One of two primary Warewulf configuration files, ``nodes.conf`` is a YAML document which records all configuration parameters for Warewulf's nodes and profiles.
    It does not contain the containers or overlays, but refers to them by name.

Overlay
    Warewulf overlays provide customization for the provisioned container image.
    Overlays may be configured on nodes or profiles, as either **system** or **runtime** overlays.

    **System overlays** are applied only once, when a node is first provisioned.

    **Runtime overlays** are applied when a node is first provisioned and periodically during the runtime of the node. (The default period is 1 minute.)

Profile
    Warewulf profiles are abstract nodes that carry the same configuration attributes but do not provision any specific node.
    Warewulf nodes may then refer to one or more such profiles for their configuration.
    In this way, profiles provide a simple mechanism for applying configuration to a group of nodes,
    and this configuration may be mixed with configuration from other profiles.

wwctl
    The main administrative interface for Warewulf is the ``wwctl`` command, which provides commands to manage nodes, profiles, containers, overlays, kernels, and more.

wwinit
    Warewulf performs some setup during the provisioning process before control is passed to the provisioned operating system.
    This process is referred to as "wwinit," and is implemented and configured by a script and overlay of the same name.

wwclient
    Warewulf adds a ``wwclient`` daemon to provisioned nodes.
    This daemon is responsible for periodically fetching and applying runtime overlays.
