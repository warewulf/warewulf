#!/bin/bash
set -ex

# Check for KVM and dump CPU info
ls -la /dev/kvm 2>/dev/null || echo "WARNING: /dev/kvm not found — KVM unavailable"
lscpu

loop_command() {
	local retry_counter=0
	local max_retries=5

	while true; do
		((retry_counter += 1))
		if [ "${retry_counter}" -gt "${max_retries}" ]; then
			exit 1
		fi
		# shellcheck disable=SC2068
		$@ && break

		# In case it is a network error let's wait a bit.
		echo "Retrying attempt ${retry_counter}"
		sleep "${retry_counter}"
	done
}

# Clean dnf cache to avoid stale metadata in container images
loop_command dnf clean all

# Enable EPEL and CRB repositories
loop_command dnf -y install epel-release
loop_command dnf config-manager --set-enabled crb

# Install build and VM provisioning dependencies
loop_command dnf -y install \
	cpio \
	git \
	make \
	golang \
	gpgme-devel \
	python3-devel \
	qemu-kvm \
	ipxe-roms-qemu \
	edk2-ovmf \
	dnsmasq \
	iproute \
	iptables-nft \
	initscripts-service \
	yq \
	ipxe-bootimgs-x86 \
	ipxe-bootimgs-aarch64 \
	tftp-server \
	nfs-utils \
	policycoreutils-python-utils

# Build and install Warewulf
go mod vendor
make defaults PREFIX=/usr SYSCONFDIR=/etc
make build
make install
cp -f etc/warewulf.conf-el10 /etc/warewulf/warewulf.conf
systemctl daemon-reload

# Network configuration
export sms_ip=192.168.100.1
export internal_netmask=255.255.255.0
export internal_network=192.168.100.0
export eth_provision=br0

# Detect gateway and DNS from the host
export ipv4_gateway
ipv4_gateway=$(ip route | awk '/default/ {print $3; exit}')
export dns_servers
dns_servers=$(awk '/^nameserver/ {print $2; exit}' /etc/resolv.conf)

# Compute node definitions
export num_computes=2
c_ip[0]=192.168.100.100
c_ip[1]=192.168.100.101
c_mac[0]=52:54:00:00:01:00
c_mac[1]=52:54:00:00:01:01
c_name[0]=c0
c_name[1]=c1
c_bmc[0]=10.16.1.1
c_bmc[1]=10.16.1.2

# Set up bridge and TAP interfaces for VMs
ip link add br0 type bridge
ip addr add 192.168.100.1/24 dev br0
ip link set br0 up

for ((i = 0; i < num_computes; i++)); do
	ip tuntap add tap$i mode tap
	ip link set tap$i up
	ip link set tap$i master br0
done

# ---------------------------------------------------------------------------
# Install fake ipmitool
#
# The integration test calls ipmitool to power-reset compute nodes.
# We intercept "chassis power reset" and launch a QEMU VM on the
# corresponding TAP interface instead.
# ---------------------------------------------------------------------------
cat >/usr/local/bin/ipmitool <<'IPMIEOF'
#!/bin/bash
# Fake ipmitool: intercepts "chassis power reset" to launch QEMU VMs.

# BMC address -> compute index mapping
declare -A BMC_TO_IDX=(
	[10.16.1.1]=0
	[10.16.1.2]=1
)

# MAC addresses per compute index
declare -a MACS=(
	"52:54:00:00:01:00"
	"52:54:00:00:01:01"
)

# Compute node names (for log prefixing)
declare -a NAMES=(
	"c0"
	"c1"
)

# Parse the BMC host from arguments (-H <host>)
bmc_host=""
args=("$@")
for ((j = 0; j < ${#args[@]}; j++)); do
	if [[ "${args[$j]}" == "-H" ]]; then
		bmc_host="${args[$((j + 1))]}"
		break
	fi
done

# Check if this is a "chassis power reset" command
is_power_reset=0
for ((j = 0; j < ${#args[@]} - 2; j++)); do
	if [[ "${args[$j]}" == "chassis" && \
		"${args[$((j + 1))]}" == "power" && \
		"${args[$((j + 2))]}" == "reset" ]]; then
		is_power_reset=1
		break
	fi
done

if [[ "${is_power_reset}" -ne 1 ]] || [[ -z "${bmc_host}" ]]; then
	# Not a power reset or no host specified — try real ipmitool
	if command -v /usr/sbin/ipmitool > /dev/null 2>&1; then
		exec /usr/sbin/ipmitool "$@"
	fi
	echo "fake ipmitool: ignoring: $*" >&2
	exit 0
fi

idx="${BMC_TO_IDX[${bmc_host}]}"
if [[ -z "${idx}" ]]; then
	echo "fake ipmitool: unknown BMC host ${bmc_host}" >&2
	exit 1
fi

mac="${MACS[${idx}]}"
name="${NAMES[${idx}]}"
tap="tap${idx}"

echo "fake ipmitool: launching QEMU VM ${name} on ${tap} (mac=${mac})"

# Find the iPXE ROM
ROMFILE=""
for candidate in \
	/usr/share/ipxe/qemu/pxe-virtio.rom \
	/usr/share/ipxe/virtio-net.rom \
	/usr/share/qemu/pxe-virtio.rom; do
	if [[ -f "${candidate}" ]]; then
		ROMFILE="${candidate}"
		break
	fi
done

# Architecture-specific QEMU flags
ARCH_FLAGS=()
case "$(uname -m)" in
aarch64)
	ARCH_FLAGS=(
		-M virt,gic-version=max
		-drive "if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.fd"
	)
	;;
esac

NETDEV_OPTS="tap,id=net0,ifname=${tap},script=no,downscript=no"
DEVICE_OPTS="virtio-net-pci,netdev=net0,mac=${mac}"
if [[ -n "${ROMFILE}" ]]; then
	DEVICE_OPTS+=",romfile=${ROMFILE}"
fi

/usr/libexec/qemu-kvm \
	"${ARCH_FLAGS[@]}" \
	-pidfile "/tmp/vm-${idx}.pid" \
	-m 3072 -smp 2 \
	-accel kvm -cpu host \
	-boot n \
	-netdev "${NETDEV_OPTS}" \
	-device "${DEVICE_OPTS}" \
	-device virtio-rng-pci,rng=rng0 \
	-object rng-random,id=rng0,filename=/dev/urandom \
	-display none -vga none -machine graphics=off \
	-chardev "stdio,id=char0,mux=on" \
	-serial chardev:char0 \
	-device virtio-serial-pci \
	-device virtconsole,chardev=char0 \
	-mon chardev=char0 \
	< /dev/null 2>&1 |
	sed -u "s/\x1b\[[0-9;]*[a-zA-Z]//g; s/\r//g" |
	awk '{ printf "[%s %s]: %s\n", "'"${name}"'", strftime("%H:%M:%S"), $0; fflush() }' &

echo "fake ipmitool: QEMU VM ${name} launched (pid=$!)"
IPMIEOF

chmod +x /usr/local/bin/ipmitool

# Create the tftpboot directory (not created by dnsmasq)
install -d -m 0755 /var/lib/tftpboot

# Configure warewulf.conf
yq -i '.ipaddr = "'"${sms_ip}"'"' /etc/warewulf/warewulf.conf
yq -i '.netmask = "'"${internal_netmask}"'"' /etc/warewulf/warewulf.conf
yq -i '.network = "'"${internal_network}"'"' /etc/warewulf/warewulf.conf
yq -i '.dhcp["range start"] = "'"${internal_network}"'"' \
	/etc/warewulf/warewulf.conf
yq -i '.dhcp["range end"] = "static"' /etc/warewulf/warewulf.conf
yq -i '.dhcp.template = "static"' /etc/warewulf/warewulf.conf

# Configure nodes.conf
sed -i "s/defaults,noauto,nofail,ro/defaults,nofail,ro/" \
	/etc/warewulf/nodes.conf

# Turn on debugging messages
yq -i '.nodeprofiles.default.kernel.args -= ["quiet"]' \
	/etc/warewulf/nodes.conf
echo "log-debug" >>/etc/dnsmasq.d/ww4-debug.conf

# Enable and start warewulfd
systemctl enable --now warewulfd

# Create profiles and overlays
wwctl profile add nodes --profile default --comment "Nodes profile"
wwctl overlay create nodeconfig
wwctl profile set --yes nodes --system-overlays nodeconfig \
	--runtime-overlays syncuser

# Set default network configuration
wwctl profile set -y nodes --netname=default --netdev="${eth_provision}"
wwctl profile set -y nodes --netname=default \
	--netmask="${internal_netmask}"
wwctl profile set -y nodes --netname=default \
	--gateway="${ipv4_gateway}"
wwctl profile set -y nodes --netname=default \
	--nettagadd=DNS="${dns_servers}"

# Configure all Warewulf services
wwctl configure --all

# Generate SSH keys
bash /etc/profile.d/ssh_setup.sh

# Import the base image
wwctl image import docker://ghcr.io/warewulf/warewulf-rockylinux:10 \
	rocky-10 --syncuser

# Add compute nodes
for ((i = 0; i < num_computes; i++)); do
	wwctl node add --image=rocky-10 --profile=nodes --netname=default \
		--ipaddr="${c_ip[$i]}" --hwaddr="${c_mac[$i]}" \
		--ipmiaddr="${c_bmc[$i]}" "${c_name[$i]}"
done

wwctl profile set -y -A 'crashkernel=no,net.ifnames=1,console=hvc0,loglevel=5' default

# Rebuild image, overlays, and reconfigure
wwctl image build rocky-10
wwctl overlay build
wwctl configure --all

# Launch QEMU VMs via fake ipmitool
export PATH=/usr/local/bin:${PATH}
for ((i = 0; i < num_computes; i++)); do
	ipmitool -H "${c_bmc[$i]}" -U admin -P password \
		chassis power reset
done
BOOT_START=$SECONDS
echo "Started VMs"

# Wait for VMs to become reachable via SSH
MAX_RETRIES=60
SLEEP_INTERVAL=10

for ((i = 0; i < num_computes; i++)); do
	echo "Waiting for ${c_name[$i]} (${c_ip[$i]}) to become reachable..."
	for ((try = 1; try <= MAX_RETRIES; try++)); do
		if ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
			"${c_ip[$i]}" hostname 2>/dev/null; then
			echo "${c_name[$i]} is up after $(( SECONDS - BOOT_START )) seconds"
			break
		fi
		echo "  Attempt ${try}/${MAX_RETRIES} - ${c_name[$i]} not ready yet"
		sleep "${SLEEP_INTERVAL}"
	done
	if [ "${try}" -gt "${MAX_RETRIES}" ]; then
		echo "ERROR: ${c_name[$i]} did not become reachable"
		exit 1
	fi
done

echo "All compute nodes are up and reachable after $(( SECONDS - BOOT_START )) seconds."

# Verify OS on compute nodes
for ((i = 0; i < num_computes; i++)); do
	echo "OS release on ${c_name[$i]}:"
	ssh -o StrictHostKeyChecking=no "${c_ip[$i]}" cat /etc/os-release
done

# Shut down VMs
for ((i = 0; i < num_computes; i++)); do
	pidfile=/tmp/vm-$i.pid
	if [[ -f "${pidfile}" ]]; then
		kill "$(<"${pidfile}")" || true
	else
		echo "WARNING: PID file ${pidfile} not found; VM may have already exited"
	fi
done
