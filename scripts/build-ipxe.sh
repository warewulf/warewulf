#!/bin/sh

set -e

TARGETS=${TARGETS:-"bin-x86_64-pcbios/undionly.kpxe bin-x86_64-efi/snponly.efi bin-arm64-efi/snponly.efi"}
IPXE_BRANCH=${IPXE_BRANCH:-master}
DESTDIR=${DESTDIR:-/usr/local/share/ipxe}

CPUS=$(grep 'processor.*:' /proc/cpuinfo | wc -l)


function usage {
  echo "Usage: $(basename $0)
        [-h] (help)

TARGETS: ${TARGETS}
IPXE_BRANCH: ${IPXE_BRANCH}
DESTDIR: ${DESTDIR}"
  exit 1
}


function main {
  local DESTDIR=$(readlink -f ${DESTDIR})

  while getopts 'h' c
  do
    case $c in
      h|*) usage
        ;;
    esac
  done

  TMPDIR=`mktemp -d /tmp/ipxebuild.XXXXXX`
  trap "rm -rf $TMPDIR" EXIT
  cd "$TMPDIR"

  git clone --branch "${IPXE_BRANCH}" https://github.com/ipxe/ipxe.git

  cd ipxe/src

  for target in ${TARGETS}
  do
    if $(echo "$target" | grep -q "\-arm64-")
    then
      if ! which aarch64-linux-gnu-gcc >/dev/null 2>&1
      then
        echo 1>&2 "aarch64-linux-gnu-gcc not found: not building for arm64"
        continue
      fi
      CROSS=aarch64-linux-gnu-
      configure_arm64
    else
      CROSS=""
      configure_x86_64
    fi
    destname=$(echo $target | tr / -)
    make -j $CPUS CROSS="${CROSS}" $target "$@" && cp -v $target ${DESTDIR}/${destname}
    restore_config
  done
}


function configure_arm64 {
  # CONSOLE_SERIAL causes build failure for aarch64, so omitting here
  # https://github.com/ipxe/ipxe/issues/658
  sed -i.bak \
      -e 's,//\(#define.*CONSOLE_FRAMEBUFFER.*\),\1,' \
      config/console.h

  sed -i.bak \
      -e 's,//\(#define.*IMAGE_ZLIB.*\),\1,' \
      -e 's,//\(#define.*IMAGE_GZIP.*\),\1,' \
      -e 's,//\(#define.*VLAN_CMD.*\),\1,' \
      config/general.h
}


function configure_x86_64 {
  sed -i.bak \
      -e 's,//\(#define.*CONSOLE_SERIAL.*\),\1,' \
      -e 's,//\(#define.*CONSOLE_FRAMEBUFFER.*\),\1,' \
      config/console.h

  sed -i.bak \
      -e 's,//\(#define.*IMAGE_ZLIB.*\),\1,' \
      -e 's,//\(#define.*IMAGE_GZIP.*\),\1,' \
      -e 's,//\(#define.*VLAN_CMD.*\),\1,' \
      config/general.h
}


function restore_config {
  cp config/console.h{.bak,}
  cp config/general.h{.bak,}
}


main "$@"
