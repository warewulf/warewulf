# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- New documentation for the hostlist syntax. #611
- New documentation for development environment (Vagrant)

### Fixed
- More aggressive `make clean`.
- Replace deprecated `io.utils` functions with new `os` functions.
- The correct header is now displayed when `-al` flags are specified to overlay
  list.
- Added a missing `.ww` extension to the `70-ww4-netname.rules` template in the
  wwinit overlay.
- Restrict access to `/warewulf/config` to root only. (#728, #742)
- KERNEL VERSION column is too short. #690
- Add support for resolving absolute path automatically. #493
- The network device "OnBoot" parameter correctly configures the ONBOOT ifcfg
  parameter. (#644)
- Add support for listing profile/node via comma-separated values. #739
- Sort the node list returned entries by name. 
- 'wwctl node edit' inconsistent state with warewulfd.  #691
- Add `--parents` option to `overlay import` subcommand to create necessary
  parent folder.  #608
- Fix kernelargs are not printing properly in node list output. #828
- Fix build configuration on Quickstart guide #847
- Add Quickstart guide for EL9
- Add EL9 Quickstart guide to index.rst
- Container file gids are now updated properly during syncuser. #840
- Fix build for API.

### Changed

- The primary hostname and warewulf server fqdn are now the canonical name in
  `/etc/hosts`
- new subcommand `wwctl genconf` is available with following subcommands:
  * `completions` which will create the files used for bash-completion. Also
     fish an zsh completions can be generated
  * `defaults` which will generate a valid `defaults.conf`
  * `man` which will generate the man pages in the specified directory
  * `reference` which will generate a reference documentation for the wwctl commands
  * `warwulfconf print` which will print the used `warewulf.conf`. If there is no valid
     `warewulf.conf` a valid configuration is provided, prefilled with default values
     and an IP configuration derived from the network configuration of the host
- All paths can now be configured in `warewulf.conf`, check the paths section of of 
   `wwctl --emptyconf genconfig warewulfconf print` for the available paths.
- Added experimental dnsmasq support.
- Refactored `profile add` command to make it alike `node add`. #658 #659 
- The ifcfg ONBOOT parameter is no longer statically `true`, so unconfigured
  interfaces may not be enabled by default. (#644)

- new subcommand `wwctl genconf` is available with following subcommands:
  * `completions` which will create the files used for bash-completion. Also
     fish an zsh completions can be generated
  * `defaults` which will generate a valid `defaults.conf`
  * `man` which will generate the man pages in the specified directory
  * `reference` which will generate a reference documentation for the wwctl commands
  * `warwulfconf print` which will print the used `warewulf.conf`. If there is no valid
     `warewulf.conf` a valid configuration is provided, prefilled with default values
     and an IP configuration derived from the network configuration of the host
- All paths can now be configured in `warewulf.conf`, check the paths section of of 
   `wwctl --emptyconf genconfig warewulfconf print` for the available paths.
- Added experimental dnsmasq support.

- new subcommand `wwctl genconf` is available with following subcommands:
  * `completions` which will create the files used for bash-completion. Also
     fish an zsh completions can be generated
  * `defaults` which will generate a valid `defaults.conf`
  * `man` which will generate the man pages in the specified directory
  * `reference` which will generate a reference documentation for the wwctl commands
  * `warwulfconf print` which will print the used `warewulf.conf`. If there is no valid
     `warewulf.conf` a valid configuration is provided, prefilled with default values
     and an IP configuration derived from the network configuration of the host
- All paths can now be configured in `warewulf.conf`, check the paths section of of 
   `wwctl --emptyconf genconfig warewulfconf print` for the available paths.
- Added experimental dnsmasq support.
- Check for formal correct IP and MAC addresses for command line options and
  when reading in the configurations
- Write log messages to stderr rather than stdout. #768
- Updates to Makefile for clarity, notably removing genconfig and replacing
  test-it with test. #890

- realy reboot also without systemd

## [4.4.0] 2023-01-18

### Added

- New Docker container node image for CentOS 7. #621

### Fixed

- Replaced an invalid variable name in a NetworkManager overlay
  template. #626
- The 'nodes' alias now correctly refers to 'node' rather than
  'profile'.
- Fixed a typo in a log message. #631
- Boolean attributes now correctly account for profile and default
  values. #630
- Kernel version is shown correctly for symlink'd kernels #640
- Changing a profile always adds an empty default interface. #661

## [4.4.0rc3] 2022-12-23

### Added

- New `defaults.conf` man page. #593
- A new debug overlay includes a template which demonstrates accessing
  all available variables. #599
- Distribute a README along with staticfiles. #189
- Add a `-y` flag to `wwctl profile add`. #610
- Distribute a source RPM with GitHub releases. #614

### Changed

- No longer ask for confirmation when deleting 0 nodes. #603
- Ask for confirmation during `wwctl container delete`. #606

### Fixed

- `wwctl profile set` now indicates "profiles" in output where it
  previously mistakenly indicated "nodes." #600
- Set correct overlay permissions for a NetworkManager configuration
  file. #591

### Fixed

- Directories within overlays no longer lose group/other write permissions #584

## [4.4.0rc2] 2022-12-09

### Added

- The environment variable `WW_CONTAINER_SHELL` is defined in a `wwctl
  container shell` environment to indicate the container in use. #579
- Network interface configuration (`ifcfg`) files now include the
  interface name and type. #457

### Fixed

- Work-around for older versions of gzip that lack a `--keep` flag
  during `wwctl container build`. #580
- The default ipxe template is once again specified as a built-in
  default and in `defaults.conf`. #581
- `wwctl container list` no longer segfaults when a container chroot
  is present without a built image. #585
- `wwctl configure hostfile` now correctly detects the presence of the
  hostfile overlay template. #571
- `wwctl overlay build` no longer panics when rendering an template
  for a node which has tags set. #568
- Minor typographical fixes. #569

## [4.4.0rc1] 2022-10-27

### Added

- iPXE binaries included with Warewulf now support VLAN tagging. #563
- `wwctl container list` now shows the container creation date,
  modification date, and size. #537
- `wwctl node edit` supports directly editing or defining node
  configuration YAML in an editor. #540
- `wwctl node export` and `wwctl node import` support importing and
  exporting node definitions as YAML or (for import) CSV. The CSV file
  must have a header in where the first field must always be the
  nodename, and the rest of the fields are the same as the long
  commandline options. Network device must have the form
  `net.$NETNAME.$NETOPTION`. (e.g., `net.default.ipaddr`) #540
- The `warewulfd.service` systemd unit file now supports `execreload`
  and `execstop`. #550
- Network interfaces now accept an `mtu` attribute. #549
- The `wwinit` overlay now supports network interface configuration
  via NetworkManager for Ethernet and InfiniBand interfaces. #539
- Default node attribute values (e.g., for kernel arguments) are now
  read in from a `defaults.conf` configuration file. If this file
  is not present, built-in default values are used. #539
- [Warewulf documentation](https://warewulf.org/docs/) is now managed
  alongside the Warewulf source code in a single code repository so
  that documentation may be updated alongside code changes.
- New man pages for `warewulf.conf` and `nodes.conf` #510
- An initial cut of the [Warewulf API](API.md) #471
- `wwctl show --render` shows overlay templates as they would be
  rendered on a given target node. #467
- `wwctl ssh` now supports Bash completion. #466

### Changed

- `wwctl overlay edit` no longer saves a new template to the overlay
  if the template is not modified from its initial state. #522
- The wwinit overlay now only sets a name for a network interface if
  that interface has a MAC address defined. #553
- `wwctl container delete` now also deletes the built images
  associated with that container. #214
- Unified internal code paths for `wwctl profile` and `wwctl node`
  commands, and between the on disk YAML format and the in memory
  format, enabling the command-line options to be autogenerated from
  the datastructures and ensuring that profile and node capabilities
  remain in sync. Multiple command line arguments have been updated or
  changed. #495, #637
- `wwctl power` commands no longer separates node output with
  additional whitespace. #514

### Fixed

- `/etc/warewulf/excludes` (read from the node image) once again
  excludes files from being included in the node image. #532
- `wwctl ssh` always uses a node's primary interface. #544
- `wwctl container show` now correctly shows the kernel version. #542
- System users are no longer prevented from logging into compute
  nodes. #538
- `wwctl overlay chown` now correctly handles uid and gid arguments. #530
- `wwctl overlay chown` no longer sets gid to `0` when unspecified. #531
- Corrected the path for `.wwbackup` files in some situations. #524
- Bypass `imgextract` for legacy BIOS machines to avoid 32-bit memory
  limitations. #497
- `warewulfd` no longer panics when network interface tags are
  defined. #468
- The wwinit overlay now configures the network device type. #465
- Minor typographical fixes. #528, #519

## [4.3.0] 2022-06-25
### Added
- All configurations files for the host (/etc/exports, /etc/dhcpd.conf, /etc/hosts) are now
  populated from the (OVERLAYDIR/host/etc/{exports|dhcpd|hosts}.ww . Also other configuration
  files like prometheus.yml.ww or slurm.conf.ww which depend on the cluster nodes can be
  placed. Also the new templated functions {{ abort }}, {{ IncludeBlock }} abd {{ no_backup }}
  are allowed now.
- nodes and profiles can now have multiple system and runtime overlays, as a comma separated list.
  The overlays of the profile and the nodes are combined.
- simple ipv6 support is now enabled. In `warewulf.conf` the option `ipaddr6`/`Ipv6net` must
  be set to enable ipv6. If enabled on of these options is set a node will get a derived
  ipv6 in the scheme `ipv6net:ipaddr4`. This address can also be overwritten for every
  node
- Multiple files can now created from a single `template.ww` file with the `{{ file FILENAME }}`
  command in the template. The command is expanded to the magic template command
  `{{ /* file FILENAME */}}` which is picked up by wwctl and everything which comes after this
  magic comment will be written to the file `FILENAME`. This mechanism is leveraged in the
  configuration files for the network, see `ifcfg.xml.ww` and `ifcgf.ww`.
- Networks can now have arbitrary keys value pairs in the profiles and on the node, so that
  things like bridges and mtu sizes can be set
- oci container tar balls can be imported with the 'file://$PATH' scheme
- uids and gids of a container can now get synced at import time, so that at least users with the
  same name have the same uid. This is not necessarily needed for warewulf, but services like
  munge.
### Changed
- Provision interface is not tied to 'eth0' any more. The provision interface must be have the
  'primary' flag now. The file `nodes.conf' must be changed accordingly.
- the provisioning network is now called primary and not default
- Creating of '/etc/exports' can now be disabled, so that `wwctl configure -a` wont overwrite
  a existing '/etc/exports'.
- The yaml format for nodes has now sub-keys for ipmi and kernel, old nodes.conf files have to
  to be changed accordingly
- host overlays can globaly disbaled, but are enabled per default
- `wwctl overlay build -H` will only build the overlays which are assigned to the nodes


## [4.1.0] - 2021-07-29
### Added
- Support for ARM nodes
- firewalld service file for warewulfd
- `-y` option to skip "Are you sure" queries
- `wwctl kernel delete` command
- `wwctl vnfs` alias for `wwctl container`
- Support for authenticated OCI registries
- warewulfd can reload config on SIGHUP and when the config file changes
- Node database index to improve lookup speeds
- Kernels and containers can be imported from a chroot subdirectory
- Systemd service file

### Changed
- `wwctl node list` output beautification
- Log timestamps are more precise
- PID file and log files are now in `/var/run` and `/var/log`, respectively
- `make install` no longer overwrites preexisting configuration files
- Kernel modules and overlays are now compressed
- `rootfstype` now uses `rootfs` in default kernel arguments
- iPXE binaries updated
- Installed container directory is deleted when import fails
- Default iPXE script now reboots erroring nodes every 15 seconds
- Only open `/etc/hosts` when writing

### Removed
- `wwctl configure` `--persist` flags have been removed. `configure` commands persist changes by default unless `--show` is used
- In-repository documentation: has been moved to it's own repository

### Fixed
- Importing containers from directory
- Debug log verbosity option takes precedence over verbose option
- `wwctl node list -n` output is formatted corectly
- Container names can contain an underscore
- `wwctl overlay build --all` does not require an argument
- specfile date format works with older versions of rpmbuild
- Use SystemOverlay when building system overlay
- dhcpd template now references correct wwctl subcommand
- `wwctl node set kernelargs` and `wwctl profile set kernelargs` change kernel arguments
