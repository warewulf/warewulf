# warewulf v4

![Warewulf](warewulf-logo.png)

#### Quick Links:

* [Documentation](docs/README.md)
* [GitHub](http://github.com/ctrliq/warewulf)

## About Warewulf

For over two decades, Warewulf has powered HPC systems around the world. From simple “under the desk” clusters to large
institutional systems at HPC centers as well as enterprises who rely on performance critical computing.

Through the evolution of Warewulf, we have seen various iterations provisioning models starting from CDROM / ISO images
to Etherboot (predecessor to PXE), then PXE, and more recently iPXE, but even during these different bootloaders,
Warewulf in it’s heart, has always been first and foremost a stateless provisioning system (e.g. the operating system
node image is not written to any persistent storage and rather it boots from the network directly into a runtime
system).

Warewulf v3 has been in production for over 6 years now as it has stabilized into a very solid and full featured
solution. But over the last few years, there have been many innovations in Enterprise technologies which can (and
should) be leveraged as part of Warewulf. Additionally, some of the lessons learned from Warewulf v3 architecture
should be rolled into an updated architecture for provisioning management.

At the core, Warewulf focuses on what has made Warewulf so widely utilized: simplicity, ultra scalable, lightweight,
and easy to manage solution built for entry level system administrators to be able to design a highly functional and
easy to maintain cluster no matter how big or small it is.

