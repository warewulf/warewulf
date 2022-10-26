#!/bin/sh
# Builds ipxe binaries from github sources or given tarball

VERSION=v1.21.1

ARCH=x86_64

PCBIOS=bin-${ARCH}-pcbios/undionly.kpxe
EFI=bin-${ARCH}-efi/ipxe.efi

PCBIOS_output=`pwd`/staticfiles/${ARCH}.kpxe
EFI_output=`pwd`/staticfiles/${ARCH}.efi

set -xe


TMPDIR=`mktemp -d /tmp/ipxebuild.XXXXXX`
if [ -f "$1" ] ; then
  WRKDIR=`pwd`
  cd "$TMPDIR"
  tar xzf $WRKDIR/$1
else
  cd "$TMPDIR"
  git clone --depth 1 --branch $VERSION https://github.com/ipxe/ipxe.git
fi

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


rm -rf "$TMPDIR"
