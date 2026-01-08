# Local Development Environment with Vagrant

This document describes how to get started using `libvirt` and `Vagrant` to facilitate local dev/test environments for Warewulf. There are multiple moving parts involved to getting `libvirt`, `vagrant-libvirt` and `Vagrant` to work in harmony on a Linux host OS. For this reason this document is primarily focused on using Enterprise Linux variants as the host OS. Also note that this dev environment is primarily meant as an example. It's highly likely this example will need to be customized depending on what you are attempting to develop or test.

This document is broken down into the following major topics:

- [Prerequisites](#prerequisites)
- [Overview](#overview)
- [Quick Start](#quick-start) (Environment needs to be configured for `libvirt`/`vagrant-libvirt` first!!!)
- [System Setup](#system-setup-libvirt-and-vagrant-libvirt-installation)
- [Customization](#customization)
- [Development Workflow](#development-workflow)
- [Troubleshooting](#troubleshooting)

## Prerequisites

Before you begin, ensure your system meets these requirements:

### Hardware Requirements

- **CPU**: x86_64 processor with hardware virtualization support (Intel VT-x or AMD-V)
- **RAM**: Minimum 8 GB (16 GB recommended)
  - wwctl VM: 2 GB
  - wwnode1: 768 MB (for Alpine) or 2+ GB (for EL/SLES)
  - wwnode2: 768 MB (for Alpine) or 2+ GB (for EL/SLES)
- **Disk Space**: Minimum 20 GB free space in `/var/lib/libvirt/images`

### Software Requirements

- **Operating System**: Linux host OS (Enterprise Linux 9/10 or openSUSE Leap 16 preferred)
- **System Setup**: Complete the [System Setup](#system-setup-libvirt-and-vagrant-libvirt-installation) section before running Quick Start commands

### Working Directory

All commands in this guide should be run from this `vagrant/` directory within the Warewulf repository unless otherwise specified.

### Files in This Directory

This directory contains all the necessary files for the Vagrant environment:

- `Vagrantfile`: VM configuration defining wwctl and compute nodes
- `.vagrantplugins`: Custom Vagrant commands (`init-warewulf`, `cleanup-warewulf`)
- `export-libvirt-sock.sh`: SSH tunnel script for libvirt socket forwarding
- `alpine-bootable.def`: Apptainer definition for Alpine Linux container
- `alpine-boot.sh`: Script to build and deploy Alpine to compute nodes
- `init-scripts/`:
  - `init-wwctl.sh`: Warewulf server initialization script
  - `add-ww-nodes.sh`: Node registration script
  - `start-vbmcd.sh`: Virtual BMC startup script

## Overview

This Vagrant environment provides a self-contained Warewulf cluster for development and testing. It consists of three VMs orchestrated via libvirt and installs Warewulf version 4.6.4 by default (configurable via `WW_VERSION` environment variable):

- **wwctl** (Rocky Linux 9): Control node running Warewulf services (DHCP, TFTP, NFS) on network 10.100.100.0/24. Includes virtualbmc for IPMI emulation and apptainer for container support. Configuration files are provisioned to the home directory of the vagrant user.

- **wwnode1/wwnode2**: Network-boot-only compute nodes (no management network) configured to PXE boot from wwctl.

Key architectural components:

1. **SSH Reverse Tunnel**: An SSH tunnel (`export-libvirt-sock.sh`) exposes the host's libvirt socket (`/var/run/libvirt/libvirt-sock`) into the wwctl VM at `/var/tmp/libvirt.sock`, enabling the guest to manage sibling VMs.

2. **Virtual BMC**: Two vbmc instances run on wwctl, providing IPMI interfaces (ports 6231/6232) for power control of wwnode1/wwnode2 via the tunneled libvirt connection.

3. **Custom Vagrant Commands**: The `.vagrantplugins` file defines `init-warewulf` (orchestrates VM creation, vbmc setup, and node registration) and `cleanup-warewulf` (powers off nodes and destroys environment).

4. **Automated Provisioning**: The `init-wwctl.sh` script grows the root partition, installs Warewulf (version controlled via `WW_VERSION` env var), configures services, and sets up the vbmc systemd service. Nodes are registered via `add-ww-nodes.sh` with MAC addresses matching the Vagrantfile definitions.

The result is a functional Warewulf cluster where `wwctl power on/off n1/n2` commands control the compute nodes through IPMI, and the nodes boot their operating system images via network provisioning.

## Quick Start

**Note**: All commands in this section should be run from the `vagrant/` directory.

### Environment Setup and Teardown

The recommended way of setting up and tearing down the dev environment is by using the custom Vagrant commands: `init-warewulf` and `cleanup-warewulf`.

- `init-warewulf`: Command provisions the Vagrant environment with three virtual machines (a Warewulf Server and two Compute Nodes). **Note:** neither compute node has an OS deployed to them at this point.

- `cleanup-warewulf`: Command completely tears down the Warewulf development environment.

### Basic Workflow

1. Setup the development environment:

   ```shell
   vagrant init-warewulf
   ```

   This command will take 5-10 minutes to complete. It performs the following:
   - Downloads the Rocky Linux 9 base image (first time only)
   - Creates and provisions the wwctl VM with Warewulf services
   - Creates wwnode1 and wwnode2 VMs
   - Sets up Virtual BMC for IPMI emulation
   - Registers the compute nodes with Warewulf

   Once this command completes executing you will have three nodes running in vagrant:

   - `wwctl`: The Warewulf server node (fully configured and running)
   - `n1`: The first compute node (powered off, no OS deployed)
   - `n2`: The second compute node (powered off, no OS deployed)

2. Deploy Alpine Linux to the compute nodes:

   ```shell
   vagrant ssh wwctl -c "sudo ./alpine-boot.sh"
   ```

   This script will take 2-3 minutes to complete. It performs the following:

   - Creates an Alpine container image using apptainer from `alpine-bootable.def`
   - Imports the resulting image into Warewulf and builds it
   - Creates an Alpine specific network overlay
   - Creates an _"alpine"_ Warewulf profile that uses this new network overlay
   - Assigns the alpine image and profile to nodes `n1` and `n2`
   - Power-cycles both nodes (Warewulf provisioning takes over and deploys the OS via network boot)

   You should see the nodes power on and boot into Alpine Linux via PXE.

3. Verify the compute nodes are running:

   ```shell
   vagrant ssh wwctl -c "sudo wwctl node list"
   vagrant ssh wwctl -c "sudo wwctl node status n1,n2"
   ```

   Expected output from `node list`:
   - Both n1 and n2 should appear in the list with their configurations

   Expected output from `node status`:
   - Shows IPMI power state, network boot status, and system information
   - Nodes should show as "POWERED ON" if boot was successful

4. Cleanup the environment:

   ```shell
   vagrant cleanup-warewulf
   ```

### Power-cycling Compute Nodes

As a precondition, shell into the `wwctl` vagrant VM:

```
vagrant ssh wwctl
```

To then power-cycle the compute nodes you would do the following:

```shell
sudo wwctl power off n1
sudo wwctl power off n2
sudo wwctl power on n1
sudo wwctl power on n2
```

## System Setup: `libvirt`, `vagrant-libvirt`, and `Vagrant` Installation

**Note**: Vagrant bundles it's own embedded Ruby environment in Vagrant. The Vagrant package installer may not always recognize system dependencies. Instructions for working around this are included for `Enterprise Linux 9 and 10` below as well for `openSUSE Leap 16`. There are other options for running `vagrant-libvirt`. For more information please see the [vagrant-libvirt](https://vagrant-libvirt.github.io/vagrant-libvirt/) site.

### `libvirt` Install

#### EL 9/10 Instructions

```shell
# Verify KVM support (should show kvm_intel or kvm_amd)
lsmod | grep kvm

# Install virtualization packages
sudo dnf groupinstall -y "Virtualization Host"
sudo dnf install -y qemu-kvm libvirt virt-install virt-top libguestfs-tools

# Enable and start libvirtd
sudo systemctl enable --now libvirtd

# Add your user to the libvirt group (critical for non-root usage)
sudo usermod -aG libvirt $(whoami)

# Verify libvirt works before proceeding
sudo virsh list --all
```

#### openSUSE Leap 16 Instructions

```shell
# Verify KVM support (should show kvm_intel or kvm_amd)
lsmod | grep kvm

# Install virtualization packages
# Note: openSUSE uses patterns instead of groups
sudo zypper install -t pattern kvm_server kvm_tools

# Install additional virtualization tools
# (many are included in the patterns, but explicit install ensures they're present)
sudo zypper install -y qemu-kvm libvirt virt-install virt-top guestfs-tools

# Enable and start libvirtd
sudo systemctl enable --now libvirtd

# Add your user to the libvirt group (critical for non-root usage)
sudo usermod -aG libvirt $(whoami)

# Verify libvirt works before proceeding
sudo virsh list --all
```

**IMPORTANT**: After completing the above, log out and log back into the system (or run `newgrp libvirt`) to refresh your group membership. This is required for non-root access to libvirt.

### `libvirt` / `virsh` Configuration

```shell
# Check and start default network
sudo virsh net-list --all
sudo virsh net-start default 2>/dev/null || true
sudo virsh net-autostart default

# Create default storage pool if missing
sudo virsh pool-define-as default dir --target /var/lib/libvirt/images
sudo virsh pool-build default
sudo virsh pool-start default
sudo virsh pool-autostart default
```

To run read-only `virsh` commands without `sudo`, set this environment variable:

```shell
export LIBVIRT_DEFAULT_URI="qemu:///system"
```

To make this permanent, add the above line to your `~/.bashrc` or `~/.zshrc`.

### Vagrant Install

#### EL 9/10 Instructions

```shell
# Install repository management tools
sudo dnf install -y dnf-plugins-core

# Add HashiCorp repository
sudo dnf config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo

# Install Vagrant
sudo dnf install -y vagrant

# Verify installation
vagrant --version
```

#### openSUSE Leap 16 Instructions

```shell
# Set desired Vagrant version (check https://releases.hashicorp.com/vagrant/ for latest)
VAGRANT_VERSION="2.4.9"

# Download the RPM
wget https://releases.hashicorp.com/vagrant/${VAGRANT_VERSION}/vagrant-${VAGRANT_VERSION}-1.x86_64.rpm

# Install with zypper
sudo zypper install ./vagrant-${VAGRANT_VERSION}-1.x86_64.rpm

# Verify installation
vagrant --version
```

**Note**: Version 2.4.9 is known to work. Check the [Vagrant releases page](https://releases.hashicorp.com/vagrant/) for newer versions.

### `vagrant-libvirt` Configuration

**Precondition:** Install `libvirt-devel` to support building of the `vagrant-libvirt` plugin:

#### EL 9/10 Instructions

```shell
# Enable CRB repository (required for libvirt-devel)
sudo dnf config-manager --set-enabled crb

# Install development tools and all required libraries
sudo dnf groupinstall -y "Development Tools"
sudo dnf install -y \
    libvirt-devel \
    libxml2-devel \
    libxslt-devel \
    ruby-devel \
    gcc \
    gcc-c++ \
    make \
    cmake \
    zlib-devel \
    pkgconf-pkg-config \
    byacc \
    wget \
    rpm-build
```

Now, install `vagrant-libvirt`:

```shell
# Set CONFIGURE_ARGS to help locate libvirt
export CONFIGURE_ARGS="--with-libvirt-include=/usr/include/libvirt --with-libvirt-lib=/usr/lib64"

# Install the plugin
vagrant plugin install vagrant-libvirt
```

#### openSUSE Leap 16 Instructions

```shell
# Minimal essential set for vagrant-libvirt
sudo zypper install -y \
    libvirt-devel \
    libxml2-devel \
    libxslt-devel \
    ruby-devel \
    gcc \
    gcc-c++ \
    make
```

Now, install `vagrant-libvirt`:

_Note: Vagrant's embedded libraries are conflicting with the system shell on openSUSE Leap 16. When the build runs pkg-config, Vagrant sets LD_LIBRARY_PATH to its embedded libs, and its bundled libreadline.so.8 breaks /bin/sh with a symbol lookup error. To work around this we temporarily have to rename the `libreadline` library in Vagrant so that the system lib path is used._

```shell
# Temporarily rename the problematic library
sudo mv /opt/vagrant/embedded/lib/libreadline.so.8 /opt/vagrant/embedded/lib/libreadline.so.8.bak

# Set CONFIGURE_ARGS to help locate libvirt
export CONFIGURE_ARGS="--with-libvirt-include=/usr/include/libvirt --with-libvirt-lib=/usr/lib64"

# Install the plugin
vagrant plugin install vagrant-libvirt

# Restore the readline library
sudo mv /opt/vagrant/embedded/lib/libreadline.so.8.bak /opt/vagrant/embedded/lib/libreadline.so.8
```

### Verifying setup

First, let's verify Vagrant reports the plugin is now present:

```shell
vagrant plugin list
```

You should see `vagrant-libvirt` in the output.

Next - run a simple test to verify that Vagrant can use libvirt to deploy a VM:

1. Create a temporary test directory (outside of the warewulf vagrant directory):

   ```shell
   mkdir -p ~/vagrant-test
   cd ~/vagrant-test
   ```

2. Create a `Vagrantfile` with the following contents:

   ```ruby
   Vagrant.configure("2") do |config|
     config.vm.box = "rockylinux/9"
     config.vm.box_version = "5.0.0"
     config.vm.provider :libvirt do |libvirt|
       libvirt.cpus = 2
       libvirt.memory = 2048
       libvirt.nic_model_type = "virtio"
       libvirt.machine_virtual_size = 32
     end
   end
   ```

3. Start the VM (this will download the Rocky Linux 9 box if not already present):

   ```shell
   vagrant up --provider=libvirt
   ```

   You should see output indicating successful VM creation and provisioning.

4. Verify SSH access:

   ```shell
   vagrant ssh -c "cat /etc/os-release"
   ```

   You should see Rocky Linux 9 version information.

5. Clean up the test:

   ```shell
   vagrant destroy -f
   cd ~
   rm -rf ~/vagrant-test
   ```

If all steps completed successfully, your system is ready to use the Warewulf Vagrant environment!

## Customization

### Warewulf Version

By default, the environment installs Warewulf version 4.6.4. To use a different version, set the `WW_VERSION` environment variable before running `vagrant init-warewulf`:

```shell
export WW_VERSION="4.6.5"
vagrant init-warewulf
```

Check the [Warewulf releases page](https://github.com/warewulf/warewulf/releases) for available versions.

### Compute Node Memory

The default configuration allocates 768 MB per compute node, which is sufficient for Alpine Linux. If you plan to test with Enterprise Linux (Rocky, RHEL, AlmaLinux) or SLES, you'll need more memory.

Edit the `Vagrantfile` and modify the `domain.memory` value for wwnode1 and wwnode2:

```ruby
# For EL or SLES distributions
config.vm.define :wwnode1 do |node1|
  # ... other config ...
  node1.vm.provider :libvirt do |domain|
    domain.memory = 2048  # Increase from 768 to 2048 or higher
    # ... other settings ...
  end
end
```

### Network Configuration

The default network configuration uses 10.100.100.0/24:
- wwctl: 10.100.100.254
- DHCP range: 10.100.100.2 - 10.100.100.9

To customize the network, edit:
1. `Vagrantfile`: Update the IP address for the wwctl private network
2. `init-scripts/init-wwctl.sh`: Modify the `/etc/warewulf/warewulf.conf` network settings

### Adding More Compute Nodes

To add additional compute nodes, edit the `Vagrantfile` and add a new VM definition:

```ruby
config.vm.define :wwnode3 do |node3|
  node3.vm.network :private_network, :libvirt__network_name => 'netboot', :libvirt__mac => '006e6f646533'
  node3.vm.synced_folder '.', '/vagrant', disabled: true
  node3.vm.provider :libvirt do |domain|
    domain.memory = 768
    domain.cpus = 1
    domain.mgmt_attach = false
    boot_network = {'network' => 'netboot'}
    domain.boot boot_network
  end
end
```

**Note**: The MAC address pattern `006e6f646533` translates to "node3" in hex. Adjust the last digits accordingly for additional nodes.

You'll also need to:
1. Update `init-scripts/add-ww-nodes.sh` to register the new node
2. Update `init-scripts/init-wwctl.sh` to add a vbmc instance for the new node

## Development Workflow

### Testing Warewulf Code Changes

If you're developing Warewulf itself and want to test your changes:

#### Method 1: Using a Custom RPM

1. Build your modified Warewulf RPM locally
2. Copy the RPM to the vagrant directory
3. Modify `init-scripts/init-wwctl.sh` to install your local RPM instead of downloading from GitHub:

   ```shell
   # Replace this line:
   # dnf install -y https://github.com/warewulf/warewulf/releases/download/v${ww_version}/warewulf-${ww_version}-1.el9.$(arch).rpm

   # With:
   dnf install -y /vagrant/your-custom-warewulf.rpm
   ```

4. Re-provision the environment:

   ```shell
   vagrant destroy -f
   vagrant init-warewulf
   ```

#### Method 2: Live Development with Synced Folders

For faster iteration during development, you can enable synced folders to mount your Warewulf source code:

1. Edit `Vagrantfile` and enable the synced folder for wwctl:

   ```ruby
   wwctl.vm.synced_folder '/path/to/your/warewulf/source', '/home/vagrant/warewulf-dev'
   ```

2. Install Warewulf from your local source inside the VM:

   ```shell
   vagrant ssh wwctl
   cd ~/warewulf-dev
   # Build and install Warewulf from source
   ```

This allows you to edit code on your host and immediately test changes in the VM without rebuilding RPMs.

### Accessing Warewulf Logs

Common log locations on the wwctl node:

```shell
vagrant ssh wwctl -c "sudo journalctl -u warewulfd -f"           # Warewulf service logs
vagrant ssh wwctl -c "sudo journalctl -u dhcpd -f"                # DHCP logs
vagrant ssh wwctl -c "sudo journalctl -u tftp -f"                 # TFTP logs
vagrant ssh wwctl -c "sudo tail -f /var/log/messages"             # System logs
```

### Testing Different Container Images

The example uses Alpine Linux, but you can create and test other distributions:

1. Create an Apptainer definition file (e.g., `rocky-bootable.def`)
2. Build and import the container:

   ```shell
   vagrant ssh wwctl
   sudo apptainer build rocky-bootable.sif rocky-bootable.def
   sudo wwctl container import rocky-bootable.sif rocky
   sudo wwctl container build rocky
   ```

3. Assign to nodes:

   ```shell
   sudo wwctl node set n1,n2 --container rocky
   sudo wwctl power cycle n1,n2
   ```

### Snapshot and Restore

Libvirt supports VM snapshots, which can be useful for testing:

```shell
# Create a snapshot of the wwctl node
virsh snapshot-create-as vagrant_wwctl wwctl-snapshot1

# List snapshots
virsh snapshot-list vagrant_wwctl

# Revert to snapshot
virsh snapshot-revert vagrant_wwctl wwctl-snapshot1

# Delete snapshot
virsh snapshot-delete vagrant_wwctl wwctl-snapshot1
```

## Troubleshooting

### Common Issues

#### "Cannot connect to libvirt" or Permission Denied

**Problem**: You get permission errors when running vagrant commands.

**Solution**:
1. Verify you're in the `libvirt` group: `groups | grep libvirt`
2. If not, ensure you logged out and back in after adding yourself to the group
3. Alternatively, run `newgrp libvirt` to refresh group membership without logging out
4. Verify libvirtd is running: `sudo systemctl status libvirtd`

#### Vagrant Init Fails During Box Download

**Problem**: The Rocky Linux box download fails or times out.

**Solution**:
1. Check your internet connection
2. Manually download the box: `vagrant box add rockylinux/9 --box-version 5.0.0 --provider libvirt`
3. Try again: `vagrant init-warewulf`

#### Compute Nodes Won't Boot

**Problem**: After running `alpine-boot.sh`, nodes don't power on or fail to boot.

**Solution**:
1. Check IPMI connectivity:
   ```shell
   vagrant ssh wwctl -c "sudo ipmitool -I lanplus -H 10.100.100.254 -p 6231 -U admin -P password power status"
   ```
2. Verify vbmc is running:
   ```shell
   vagrant ssh wwctl -c "sudo systemctl status vbmcd"
   vagrant ssh wwctl -c "sudo vbmc list"
   ```
3. Check the libvirt SSH tunnel:
   ```shell
   ps aux | grep libvirt.sock
   ```
4. Check DHCP/TFTP logs:
   ```shell
   vagrant ssh wwctl -c "sudo journalctl -u dhcpd -f"
   vagrant ssh wwctl -c "sudo journalctl -u tftp -f"
   ```

#### "Network netboot not found" Error

**Problem**: Vagrant fails to create VMs with network errors.

**Solution**:
1. Verify the netboot network exists:
   ```shell
   virsh net-list --all
   ```
2. If missing, vagrant should create it automatically. Try:
   ```shell
   vagrant destroy -f
   vagrant init-warewulf
   ```

#### Out of Disk Space

**Problem**: VM provisioning fails due to disk space issues.

**Solution**:
1. Check available space: `df -h /var/lib/libvirt/images`
2. Clean up old vagrant boxes: `vagrant box prune`
3. Remove unused libvirt images:
   ```shell
   virsh vol-list default
   virsh vol-delete <volume-name> default
   ```

#### vagrant-libvirt Plugin Won't Install

**Problem**: `vagrant plugin install vagrant-libvirt` fails during compilation.

**Solution**:
1. Ensure all development libraries are installed (see System Setup section)
2. Check that `CONFIGURE_ARGS` is set correctly
3. For openSUSE, verify the libreadline workaround was applied
4. Check build errors for missing libraries and install them

#### Compute Nodes Stuck at "Powered Off"

**Problem**: `wwctl node status` shows nodes as powered off even after power on commands.

**Solution**:
1. Manually power on via IPMI:
   ```shell
   vagrant ssh wwctl -c "sudo ipmitool -I lanplus -H 10.100.100.254 -p 6231 -U admin -P password power on"
   ```
2. Check vbmc logs:
   ```shell
   vagrant ssh wwctl -c "sudo journalctl -u vbmcd -f"
   ```
3. Verify libvirt can see the nodes:
   ```shell
   vagrant ssh wwctl -c "sudo LIBVIRT_DEFAULT_URI='qemu:///system?socket=/var/tmp/libvirt.sock' virsh list --all"
   ```

#### Failed to Cleanup Environment

**Problem**: `vagrant cleanup-warewulf` hangs or fails.

**Solution**:
1. Manually power off nodes via libvirt:
   ```shell
   virsh destroy vagrant_wwnode1
   virsh destroy vagrant_wwnode2
   virsh destroy vagrant_wwctl
   ```
2. Force destroy:
   ```shell
   vagrant destroy -f
   ```
3. Clean up SSH tunnel:
   ```shell
   pkill -f "libvirt.sock"
   rm .export-libvirt-sock.pid
   ```

### Additional Resources

- **Warewulf Documentation**: https://warewulf.org/docs/v4.6.x/troubleshooting/troubleshooting.html
- **vagrant-libvirt Documentation**: https://vagrant-libvirt.github.io/vagrant-libvirt/
- **Libvirt Documentation**: https://libvirt.org/docs.html
- **Warewulf GitHub Issues**: https://github.com/warewulf/warewulf/issues
