# Warewulf on SUSE/openSUSE/ALP
This is the warewulf container including `tftp` and `dhcpd` services. It also shares the install run labels with the upstream container, but this container uses the rpm package of warewulf and not the upstream source.

The containers used for deployment, the overlays and compressed images are stored not in the warewulf container but under `var/lib/warewulf` on the host. Also the configuration directory `/etc/warewulf`is created on the host.

`dhpd` and `tftp` require systemd running also in the container. To provide these services the container must also run in the same network as the host and the container must also run in privileged mode.

Additionally the `warewulf-container-manage.sh` script is included to manage the container via podman and the `wwctl` command to manage the warewulf itself.

# Install the warewulf container

## Prepare the host system

It is heavlily advised that the host has a static ipv4 address. To configure this on ALP
you can use `nmcli` with the following command:
```
nmcli connection modify "$(nmcli -t device | awk -F: '/ethernet/{print$4;exit}')" \
  ipv4.method manual \
  ipv4.addresses 192.168.1.250/24 \
  ipv4.gateway 192.168.1.1\
  ipv4.dns 182.168.1.1
```

In order to run the warewulf container, you can just use the the '''runlable install''' on the container available in the registry.
```
# podman container runlabel install registry.opensuse.org/suse/alp/workloads/tumbleweed_containerfiles/suse/alp/workloads/warewlf:latest
```
This will create the directories `/var/lib/warewulf` and `/etc/warewulf` on the host system and populate it with necesarry configuraiton files, if this files are not present on the host.
On the first install the actual ip network settings are used as base values for the dynamic
network configuration in `warewulf.conf`.

## Create the container

The warewulf service is started with the command
```
# podman container runlabel run registry.opensuse.org/suse/alp/workloads/tumbleweed_containerfiles/suse/alp/workloads/warewlf:latest
```
Now the cluster can be managed with the `wwctl` command.

## Remove the Container

The container itself can be remove with the '''label uninstall''' which will remove the container and its scripts from the host.
```
# podman container runlabel uninstall registry.opensuse.org/suse/alp/workloads/tumbleweed_containerfiles/suse/alp/workloads/warewlf:latest
```

This step *doesn't* remove the configuration of warewulf under `/etc/warewulf` and the containers with the overlays under `/var/lib/warewulf`. You can use the purge label to remove these directories.

## Purge warewulf

All components of warewulf including `/etc/warewulf` and `/var/lib/warewulf` can be removed the the '''label purge''' which can be called with
```
# podman container runlabel purge registry.opensuse.org/suse/alp/workloads/tumbleweed_containerfiles/suse/alp/workloads/warewlf:latest
```
Please note that the '''label purge''' inherits the label '''label uninstall'''.



## More Info

* [Warewulf](https://github.com/warewulf/warewulf)
