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

sed -i -e 's@//#define	CONSOLE_SERIAL@#define	CONSOLE_SERIAL@' config/console.h
sed -i -e 's@//#define	CONSOLE_FRAMEBUFFER@#define	CONSOLE_FRAMEBUFFER@' config/console.h

sed -i -e 's@//#define	IMAGE_ZLIB@#define	IMAGE_ZLIB@' config/general.h
sed -i -e 's@//#define	IMAGE_GZIP@#define	IMAGE_GZIP@' config/general.h

make -j 8 $PCBIOS
make -j 8 $EFI


cp $PCBIOS $PCBIOS_output
cp $EFI $EFI_output


rm -rf "$TMPDIR"
