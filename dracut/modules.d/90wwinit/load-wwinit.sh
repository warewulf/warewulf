#!/bin/bash

function setup_disks() {
	# format and prepare the disk(s)
        /usr/bin/ignition --root=/sysroot --platform=metal --stage=disks || die "warewulf: failed to partition/format disk"
	# mount /dev/disk/by-partlabel/root "$NEWROOT" || die "warewulf: failed to mount new root"
        /usr/bin/ignition --root=/sysroot --platform=metal --stage=mount || die "warewulf: failed to mount disk"

}


info "warewulf: Running warewulf v4 dracut init ${wwinit_persistent}"
archives="image system runtime"
if [ ${wwinit_persistent} -ne 1 ] ; then 
	info "warewulf: Mounting tmpfs at $NEWROOT"
	mount -t tmpfs -o mpol=interleave ${wwinit_tmpfs_size_option} tmpfs "$NEWROOT"
else
	info "warewulf: Using persistent setup ${wwinit_ignition}/ignition.json.ww?render=${wwinit_id}"
	# get the igintion config and store in /run/igintion.ign, don't use igntion as so we can
	# construct the download link
	curl --location --silent --get ${localport} \
            --retry 60 --retry-delay 1 \
            "${wwinit_ignition}/ignition.json.ww?render=${wwinit_id}" -o /run/ignition-rootfs.json || die "warewulf: failed to fetch ignition configuration from ${wwinit_ignition}/ignition.json.ww?render=${wwinit_id}"
	jq '.storage.filesystems |= map(select(.device=="/dev/disk/by-partlabel/rootfs").path="/")' /run/ignition-rootfs.json  > /run/ignition.json || die "warewulf: failed to rewrite ignition configuration"
	setup_disks
#	if [ -e ${NEWROOT}/warewulf/ww_${wwinit_imagename} ] ; then 
#		info "warewulf: found  ${NEWROOT}/warewulf/ww_${wwinit_imagename} running rsync for update"
#		# we need only the overlays as we update the root 
#		# with rsync when the same image is used
#		archives="system runtime"
#		rsync -aux --delete --exclude=/proc/ --exclude=/sys/ --exclude=/dev rsync://${wwinit_ip}/${wwinit_imagename} $NEWROOT || die "warewulf: failed to run rsync"
#	elif [  -e ${NEWROOT}/warewulf/ww_* ] ; then
#		info "warewulf: found $(ls ${NEWROOT}/warewulf/ww_*) but image is ${wwinit_imagename}, wiping $NEWROOT"
#		# we didn't have the wanted image but provisioned with warewulf, wipe the rootfs
#		/usr/bin/ignition --root=/sysroot --platform=metal --stage=umount || die "warewulf: failed to mount disk"
#		jq '.storage.filesystems |= map(select(.device=="/dev/disk/by-partlabel/rootfs").wipeFilesystem=true)' /run/ignition.json > /run/ignition.json.mod || die "couldn't reformation ignition configuration"
#		mv /run/ignition.json.mod /run/ignition.json || die "warewulf: couldn't mv ignition configuration back"
#		setup_disks
#	fi
fi
for stage in $archives ; do
   info "warewulf: Loading ${stage}"
   # Load runtime overlay from a static privledged port.
   # Others use default settings.
   localport=""
   if [[ "${stage}" == "runtime" ]] ; then
        localport="--local-port 1-1023"
   fi
   (
        curl --location --silent --get ${localport} \
            --retry 60 --retry-connrefused --retry-delay 1 \
            --data-urlencode "assetkey=${wwinit_assetkey}" \
            --data-urlencode "uuid=${wwinit_uuid}" \
            --data-urlencode "stage=${stage}" \
            --data-urlencode "compress=gz" \
            "${wwinit_uri}" \
        | gzip -d \
        | cpio -im --directory="${NEWROOT}"
    ) || die "Unable to load stage: ${stage}"
done

if [ ${wwinit_persistent} -eq 1 ] ; then
	# this avoids that ignition runs a second time
	info "warewulf: removing ignition config from wwinit image"
	rm -rf ${NEWROOT}/warewulf/ignition.json
	echo "Container name of persistent install: ${wwinit_imagename}" > ${NEWROOT}/warewulf/ww_${wwinit_imagename}
	echo "# created from ${wwinit_ignition_uri}/fstab.ww?render=${winit_id}" > ${NEWROOT}/etc/fstab
	curl --location --silent --get ${localport} \
            --retry 60 --retry-delay 1 \
            "${wwinit_ignition}/fstab.ww?render=${wwinit_id}" >> ${NEWROOT}/etc/fstab || die "warewulf: failed to write correct fstab"
fi
info "warewulf: Finished warewulf v4 dracut"

