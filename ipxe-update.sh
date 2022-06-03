#!/bin/sh

ARCH=x86_64

PCBIOS=bin-${ARCH}-pcbios/undionly.kpxe
EFI=bin-${ARCH}-efi/ipxe.efi

PCBIOS_output=`pwd`/staticfiles/${ARCH}.kpxe
EFI_output=`pwd`/staticfiles/${ARCH}.efi

set -xe


TMPDIR=`mktemp -d`
cd "$TMPDIR"

git clone https://github.com/ipxe/ipxe.git
cd ipxe/src

sed -i.bak \
    -e 's,//\(#define.*CONSOLE_SERIAL.*\),\1,' \
    -e 's,//\(#define.*CONSOLE_FRAMEBUFFER.*\),\1,' \
    config/console.h

sed -i.bak \
    -e 's,//\(#define.*IMAGE_ZLIB.*\),\1,' \
    -e 's,//\(#define.*IMAGE_GZIP.*\),\1,' \
    config/general.h

make -j 8 $PCBIOS
make -j 8 $EFI


cp $PCBIOS $PCBIOS_output
cp $EFI $EFI_output


rm -rf "$TMPDIR"
