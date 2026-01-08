#!/bin/sh

#
# Resize root partition (if necessary)
#
dnf install -y cloud-utils-growpart util-linux bc

echo ""
echo "================================================================================"
echo "[Resizing Root Partition]"
echo ""

# Find the device that is mounted as root (/)
ROOT_DEVICE=$(findmnt -n -o SOURCE /)
echo "Detected root device: $ROOT_DEVICE"

# Resolve any symlinks or device-mapper names to the actual partition
ROOT_PARTITION=$(readlink -f "$ROOT_DEVICE")
echo "Detected root partition: $ROOT_PARTITION"

# Extract the parent disk and partition number from the partition path
PARENT_DISK="/dev/$(lsblk -no PKNAME "$ROOT_PARTITION")"
PARTITION_SUFFIX="${ROOT_PARTITION#$PARENT_DISK}"
PARTITION_NUMBER="${PARTITION_SUFFIX#p}"
echo "Parent disk: $PARENT_DISK"
echo "Partition number: $PARTITION_NUMBER"

# Check if there's actually room to grow
# Note: We use tail -n +2 to skip the first line, which is the disk itself.
# Lines 2 onward are the actual partitions.
DISK_SIZE_BYTES=$(lsblk -b -dn -o SIZE "$PARENT_DISK")
PARTITIONS_SIZE_BYTES=$(lsblk -b -ln -o SIZE "$PARENT_DISK" | tail -n +2 | paste -sd+ | bc)
UNALLOCATED_BYTES=$((DISK_SIZE_BYTES - PARTITIONS_SIZE_BYTES))
THRESHOLD_BYTES=$((100 * 1024 * 1024))

echo "Disk size: $((DISK_SIZE_BYTES / 1024 / 1024)) MiB"
echo "Allocated to partitions: $((PARTITIONS_SIZE_BYTES / 1024 / 1024)) MiB"
echo "Unallocated space: $((UNALLOCATED_BYTES / 1024 / 1024)) MiB"

if [ "$UNALLOCATED_BYTES" -gt "$THRESHOLD_BYTES" ]; then
    echo "Growing partition $PARENT_DISK partition $PARTITION_NUMBER..."
    growpart "$PARENT_DISK" "$PARTITION_NUMBER"

    FS_TYPE=$(lsblk -no FSTYPE "$ROOT_PARTITION")
    echo "Detected filesystem type: $FS_TYPE"

    case "$FS_TYPE" in
        xfs)
            xfs_growfs /
            ;;
        ext4|ext3|ext2)
            resize2fs "$ROOT_PARTITION"
            ;;
        *)
            echo "Warning: Unknown filesystem type '$FS_TYPE'. Cannot auto-resize."
            echo "Partition was grown, but you may need to manually resize the filesystem."
            ;;
    esac

    echo "Root partition successfully grown!"
    df -h /
else
    echo "No significant unallocated space available. Nothing to do."
fi

echo "================================================================================"
echo ""

echo "StreamLocalBindUnlink yes" > /etc/ssh/sshd_config.d/60-forward-cleanup.conf
systemctl reload sshd.service

dnf install -y libvirt-client python3-pip python3-libvirt ipmitool epel-release
pip3 install virtualbmc
dnf install -y apptainer

ww_version=$1

cat > /usr/local/bin/setup-wwnode <<EOF
#!/bin/sh

node=\$1
port=\$2

if [ ! -e /root/.vbmc/\$node ]; then
  /usr/local/bin/vbmc add --username admin --password password --port \$port --address 10.100.100.254 --libvirt-uri qemu:///system?socket=/var/tmp/libvirt.sock \$node
fi

/usr/local/bin/vbmc start \$node
EOF

chmod 755 /usr/local/bin/setup-wwnode

dnf install -y https://github.com/warewulf/warewulf/releases/download/v${ww_version}/warewulf-${ww_version}-1.el9.$(arch).rpm

cat > /etc/warewulf/warewulf.conf <<EOF
ipaddr: 10.100.100.254
netmask: 255.255.255.0
network: 10.100.100.0
warewulf:
    port: 9873
    secure: false
    update interval: 60
    autobuild overlays: true
    host overlay: true
    grubboot: false
api:
    enabled: false
    allowed subnets:
        - 127.0.0.0/8
        - ::1/128
dhcp:
    enabled: true
    template: default
    range start: 10.100.100.2
    range end: 10.100.100.9
    systemd name: dhcpd
tftp:
    enabled: true
    tftproot: /var/lib/tftpboot
    systemd name: tftp
    ipxe:
        00:0B: arm64-efi/snponly.efi
        "00:00": undionly.kpxe
        "00:07": ipxe-snponly-x86_64.efi
        "00:09": ipxe-snponly-x86_64.efi
nfs:
    enabled: true
    export paths:
        - path: /home
          export options: rw,sync
        - path: /opt
          export options: ro,sync,no_root_squash
    systemd name: nfs-server
ssh:
    key types:
        - ed25519
        - ecdsa
        - rsa
        - dsa
image mounts:
    - source: /etc/resolv.conf
      dest: /etc/resolv.conf
      readonly: true
paths:
    bindir: /usr/bin
    sysconfdir: /etc
    localstatedir: /var/lib
    cachedir: /var/cache
    ipxesource: /usr/share/ipxe
    srvdir: /var/lib
    firewallddir: /usr/lib/firewalld/services
    systemddir: /usr/lib/systemd/system
    datadir: /usr/share
    wwoverlaydir: /var/lib/warewulf/overlays
    wwchrootdir: /var/lib/warewulf/chroots
    wwprovisiondir: /var/lib/warewulf/provision
    wwclientdir: /warewulf
EOF

systemctl enable --now warewulfd
wwctl configure --all
systemctl restart warewulfd

cat > /etc/systemd/system/vbmcd.service <<EOF
[Install]
WantedBy = multi-user.target

[Service]
ExecReload = /bin/kill -HUP $MAINPID
ExecStart = /usr/local/bin/vbmcd --foreground
Group = root
Restart = on-failure
RestartSec = 2
TimeoutSec = 120
Type = simple
User = root
ExecStartPost = /usr/local/bin/setup-wwnode vagrant_wwnode1 6231
ExecStartPost = /usr/local/bin/setup-wwnode vagrant_wwnode2 6232

[Unit]
After = syslog.target
After = network.target
Description = vbmc service
EOF

systemctl daemon-reload

# shut down the firewall - not needed for this test environment:
sudo systemctl disable --now firewalld
