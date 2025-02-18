========
Glossary
========

Cluster network
    A dedicated network for the Warewulf cluster. Used for provisioning
    communication between cluster nodes and the Warewulf server.

External services
    The Warewulf server can configure external services to support the
    provisioning process. For example, the Warewulf server typically deploys and
    configures a DHCP server (either ISC DHCP or dnsqmasq) and a TFTP server.

Image
    The node images that Warewulf manages and provisions. Images may be imported
    from OCI image registries, OCI image archives, Apptainer sandboxes, and
    manual chroot directories.

    Warewulf images are maintained as an uncompressed "virtual node file system"
    (VNFS, sometimes also referred to as a "chroot"). These virtual file systems
    are then built as single-file images which may be used to provision a node.

Kernel
    In addition to an image, Warewulf also requires a kernel (typically a Linux
    kernel) in order to provision a node.

    Warewulf (after v4.3.0) automatically provisions a kernel detected and
    extracted from the image itself. In most cases, kernels may be installed in
    the image using normal system packages, and no special consideration is
    necessary.

Node
    Warewulf nodes are the systems that are being provisioned by Warewulf. The
    roles of these systems could be "compute", "storage", "GPU", "IO", etc.

nodes.conf
    One of two primary Warewulf configuration files, ``nodes.conf`` is a YAML
    document which records all configuration parameters for Warewulf's nodes and
    profiles. It does not contain the images or overlays, but refers to them by
    name.

    This file is sometimes referred to as the "nodes database" or "node
    registry."

Overlay
    Warewulf overlays provide customization for the provisioned image. Overlays
    may be configured on nodes or profiles, as either **system** or **runtime**
    overlays.

    **System overlays** are applied only once, when a node is first provisioned.

    **Runtime overlays** are applied when a node is first provisioned and
    periodically during the runtime of the node. (The default period is 1
    minute.)

    Warewulf includes a number of **distribution overlays**; but additional
    **site overlays** can be added to a Warewulf environment.

Profile
    Warewulf profiles are abstract nodes that carry the same configuration
    attributes but do not provision any specific node. Warewulf nodes may then
    refer to one or more such profiles for their configuration. In this way,
    profiles provide a simple mechanism for applying configuration to a group of
    nodes, and this configuration may be mixed with configuration from other
    profiles.

Server, Warewulf
    The Warewulf controller runs the Warewulf daemon (``warewulfd``) and is
    responsible for the management, control, and administration of the cluster.
    This system may also sometimes be referred to as the "master," "head," or
    "admin" node.

    A typical Warewulf controller also runs a DHCP service and a TFTP service,
    and often an NFS service; though these services may be managed separately
    and on separate servers.

Two-stage boot
    A two-stage boot uses an intermediate image (often called "initrd," or
    "initramfs") to initialize hardware and load the final image. This contrasts
    with Warewulf's default "single stage" behavior, which effectively uses the
    final image as a large initial root file system.

    The Warewulf two-stage boot process currently supports Dracut-based images.

warewulf.conf
    One of two primary Warewulf configuration files, ``warewulf.conf`` is a YAML
    document which records all configuration parameters for the Warewulf server
    and its optional subservices.

wwclient
    Warewulf adds a ``wwclient`` daemon to provisioned nodes. This daemon is
    responsible for periodically fetching and applying runtime overlays.

wwctl
    The main administrative interface for Warewulf is the ``wwctl`` command,
    which provides commands to manage nodes, profiles, images, overlays,
    kernels, and more.

wwinit
    Warewulf performs some setup during the provisioning process before control
    is passed to the provisioned operating system. This process is referred to
    as "wwinit," and is implemented and configured by a script and overlay of
    the same name.
