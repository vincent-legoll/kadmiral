#!/bin/sh

# Create an up and running k8s cluster using kadmiral

set -e
set -x

usage() {
    cat << EOD
Usage: $(basename "$0") [options]
Available options:
  -h            This message

Create k8s cluster with kadmiral

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

echo "Initialize k8s cluster"
echo "----------------------"
kadmiral init -v3

echo "Install CNI (Cilium)"
echo "-------------------"
kadmiral cni install cilium -v3

echo "Join worker nodes"
echo "-----------------"
kadmiral join node -v3
