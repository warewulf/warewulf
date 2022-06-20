#!/usr/bin/bash

VERSION=$1

dnf install -y epel-release
dnf install -y mock

tar -xf warewulf-${VERSION}.tar.gz warewulf-${VERSION}/warewulf.spec

RELEASE=$(grep 'Release: ' warewulf-${VERSION}/warewulf.spec | cut -d ':' -f2 | awk -F'%' '{print $1}' | tr -d ' ')
echo RELEASE=${RELEASE} >> $GITHUB_ENV

mock -r rocky+epel-8-x86_64 --rebuild --spec=warewulf-${VERSION}/warewulf.spec --sources=.
mv /var/lib/mock/rocky+epel-8-x86_64/result/warewulf-${VERSION}-${RELEASE}.el8.x86_64.rpm .

mock -r centos+epel-7-x86_64 --rebuild --spec=warewulf-${VERSION}/warewulf.spec --sources=.
mv /var/lib/mock/centos+epel-7-x86_64/result/warewulf-${VERSION}-${RELEASE}.el7.x86_64.rpm .

mock -r opensuse-leap-15.3-x86_64 --rebuild --spec=warewulf-${VERSION}/warewulf.spec --sources=.
mv /var/lib/mock/opensuse-leap-15.3-x86_64/result/warewulf-${VERSION}-${RELEASE}.suse.lp153.x86_64.rpm .