#!/bin/sh
# Copies a built program to the tablet: tools/install.sh out/hello. It writes only inside the
# program's own folder in AppLoad's home, never on the system partition. The tablet must be reachable
# over SSH: on USB it is 10.11.99.1, and REMARKABLE_HOST changes that.
#
# The files go to a folder next to the real one and are moved at the end, so a cut cable leaves the
# old version whole. tar needs no SFTP server on the tablet, unlike scp.
set -eu

if [ $# -ne 1 ] || [ ! -d "$1" ]; then
    echo "usage: tools/install.sh out/<program>" >&2
    exit 2
fi

built=${1%/}
name=$(basename "$built")
case $name in
    *[!a-z0-9-]* | '')
        echo "the program's name may only have lower-case letters, digits and hyphens: $name" >&2
        exit 2
        ;;
esac

host=${REMARKABLE_HOST:-root@10.11.99.1}
apps=/home/root/xovi/exthome/appload
incoming=$apps/.incoming-$name

COPYFILE_DISABLE=1 tar -C "$built" -cf - . | ssh "$host" \
    "test -d '$apps' && rm -rf '$incoming' && mkdir '$incoming' && tar -xf - -C '$incoming' && rm -rf '$apps/$name' && mv '$incoming' '$apps/$name'"
echo "installed $name on $host: reload AppLoad to see it"
