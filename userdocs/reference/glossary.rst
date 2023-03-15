========
Glossary
========

Container
    Containers are used by Warewulf as the template for the VNFS
    image. Warewulf containers can be any type of OCI or Singularity
    standard image formats but maintained on disk as an "OCI
    bundle". Warewulf integrates with Docker, Docker Hub, any OCI
    registery, Singularity, standard chroots, etc.

Controller
    The controller node(s) are the resources responsible for
    management, control, and administration of the
    cluster. Historically these systems have been called "master",
    "head", or "administrative" nodes, but we feel the term
    "controller" is more appropriate and descriptive of the role of
    this system.

Initramfs

Kernel

Overlays

Virtual Node File System (VNFS)

Workers
    Worker nodes are the systems that are being provisioned by
    Warewulf. The roles of these systems could be "compute",
    "storage", "GPU", "IO", etc. which would typically be used as a
    prefix, for example: "**compute worker node**"
