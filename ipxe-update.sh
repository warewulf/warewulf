#!/bin/sh
# Builds ipxe binaries from github sources or given tarball

VERSION=v1.21.1

ARCH=x86_64

PCBIOS=bin-${ARCH}-pcbios/undionly.kpxe
EFI=bin-${ARCH}-efi/ipxe.efi

PCBIOS_output=`pwd`/staticfiles/${ARCH}.kpxe
EFI_output=`pwd`/staticfiles/${ARCH}.efi

SYSBIOS="/usr/share/ipxe/undionly.kpxe"
SYSEFI="/usr/share/ipxe/ipxe-${ARCH}.efi"

TMPDIR=`mktemp -d /tmp/ipxebuild.XXXXXX`

usage() {
  echo "Usage: $(basename $0) 
        [-s] (install efi and bios from system location)
        [-g] (build from git repo and install built efi and bios)
        [-h] (help)
        [-f] ipxe_source_gz_file (build from given tar.gz ipxe file and install built efi and bios)
       "
  exit 1
}

sysbuild() {
  if [ -f "${SYSEFI}" ] ; then
    cp $SYSEFI $EFI_output
  else
    echo "No such file or directory: ${SYSEFI}, try 'dnf install ipxe-bootimgs-x86'"
    exit 1
  fi
  
  if [ -f "${SYSBIOS}" ] ; then
    cp $SYSBIOS $PCBIOS_output
  else
    echo "No such file or directory: ${SYSBIOS}, try 'dnf install ipxe-bootimgs-x86'"
    exit 1
  fi

  echo "copy ${SYSEFI} and ${SYSBIOS} done"
}

gitbuild() {
  cd "$TMPDIR"
  git clone --depth 1 --branch $VERSION https://github.com/ipxe/ipxe.git
  build
}

filebuild() {
  TARGET_PATH=$1
  case $TARGET_PATH in 
    /*) ;;
    *)
     TARGET_PATH=`pwd`/$TARGET_PATH
     ;;
  esac
  if [ -f "${TARGET_PATH}" ] ; then
    cd "$TMPDIR"
    tar xzf $TARGET_PATH
    build $TARGET_PATH
  else
    echo "No such file $TARGET_PATH"
    exit 1
  fi
}

build() {
  cd ipxe*/src

  sed -i.bak \
      -e 's,//\(#define.*CONSOLE_SERIAL.*\),\1,' \
      -e 's,//\(#define.*CONSOLE_FRAMEBUFFER.*\),\1,' \
      config/console.h

  sed -i.bak \
      -e 's,//\(#define.*IMAGE_ZLIB.*\),\1,' \
      -e 's,//\(#define.*IMAGE_GZIP.*\),\1,' \
      -e 's,//\(#define.*VLAN_CMD.*\),\1,' \
      config/general.h

  make -j 8 $PCBIOS "$@"
  make -j 8 $EFI "$@"

  cp $PCBIOS $PCBIOS_output
  cp $EFI $EFI_output

  echo "copy ${PCBIOS} and ${EFI} done"
}

while getopts 'sghdf:' c
do
 case $c in
  s) sysbuild 
     break 
     ;;
  g) gitbuild 
     break 
     ;;
  f) filebuild $OPTARG 
     break
     ;;
  h|*) usage 
     ;;
 esac
done

rm -rf "$TMPDIR"
