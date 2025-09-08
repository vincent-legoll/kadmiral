#!/bin/sh

# Reset a k8s cluster using kadmiral

set -e
set -x

usage() {
    cat << EOD
Usage: $(basename "$0") [options]
Available options:
  -h            This message

Reset k8s cluster with kadmiral

EOD
}

# Get the options
while getopts h c ; do
    case $c in
        h) usage ; exit 0 ;;
        \?) usage ; exit 2 ;;
    esac
done
shift "$((OPTIND-1))"

if [ $# -ne 0 ] ; then
    usage
    exit 2
fi

echo "Reset k8s cluster"
echo "-----------------"
kadmiral reset

