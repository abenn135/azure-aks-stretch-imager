#!/bin/bash

set -euo pipefail

# This script runs on the test VM.

STORAGE_ACCOUNT_NAME=$(jq -r .storageAccountName config.json)
STORAGE_CONTAINER_NAME=$(jq -r .storageContainerName config.json)
BUILD_VERSION=$(jq -r .buildVersion config.json)

if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root. Please run with sudo."
    exit 1
fi

apt-get update
apt-get install -y qemu-system-x86 qemu-utils
az storage blob download --account-name $STORAGE_ACCOUNT_NAME --container-name $STORAGE_CONTAINER_NAME --name ${BUILD_VERSION}_root.img.gz --file /tmp/compressed-root-partition.img.gz
az storage blob download --account-name $STORAGE_ACCOUNT_NAME --container-name $STORAGE_CONTAINER_NAME --name ${BUILD_VERSION}_boot.tgz --file /tmp/compressed-boot-partition.tgz
az storage blob download --account-name $STORAGE_ACCOUNT_NAME --container-name $STORAGE_CONTAINER_NAME --name ${BUILD_VERSION}_efi.tgz --file /tmp/compressed-efi-partition.tgz
echo "All images downloaded."

echo "Fetching alpine linux netboot image for QEMU..."
curl -L -o alpine-virt.iso https://dl-cdn.alpinelinux.org/alpine/v3.23/releases/x86_64/alpine-virt-3.23.3-x86_64.iso

echo "Starting QEMU to copy images to local disk..."
qemu-system-x86_64 \
    -m 4G \
    -netdev user,id=net0,hostfwd=tcp::2222-:22 \
    -smp 2 \
    -nographic \
    -cdrom alpine-virt.iso
