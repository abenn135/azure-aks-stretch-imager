#!/bin/bash

# This script, when run with a JSON file indicating where the source partition images are, starts an Azure VM, downloads those images, boots a QEMU instance on netboot that copies those images onto its local disk, and then reboots the VM, to demonstrate it will boot from the images.

set -euo pipefail

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <path to JSON file with image URLs>"
    exit 1
fi

CONFIG_FILE="$1"

BUILD_ARCH=$(jq -r .arch "$CONFIG_FILE")

GROUP=${GROUP:-alexbenn-test}
LOCATION=${LOCATION:-eastus2}
RUNNER_VM_NAME=${RUNNER_VM_NAME:-disk-spinner}
DELETE_VM=${DELETE_VM:-true}

if [ -z $RUNNER_VM_SKU ]; then
    if [ "$BUILD_ARCH" = "x86_64" ]; then
        RUNNER_VM_SKU=Standard_D4s_v7
    elif [ "$BUILD_ARCH" = "arm64" ]; then
        RUNNER_VM_SKU=Standard_D4ps_v6
    else
        echo "Error: Unsupported architecture $BUILD_ARCH. Supported values are x86_64 and arm64."
        exit 1
    fi
fi

az login --identity

if ! az vm show -g $GROUP -n $RUNNER_VM_NAME &>/dev/null; then
    echo "Creating VM $RUNNER_VM_NAME in resource group $GROUP..."
    az vm create -g $GROUP -l $LOCATION -n $RUNNER_VM_NAME \
        --size $RUNNER_VM_SKU \
        --image Ubuntu2404 \
        --assign-identity \
        --role reader \
        --scope /Subscriptions/$AZURE_SUBSCRIPTION_ID/resourceGroups/$GROUP
else
    echo "VM $RUNNER_VM_NAME already exists in resource group $GROUP. Skipping VM creation."
fi

VM_IP=$(az vm show -g $GROUP -n $RUNNER_VM_NAME -d --query 'publicIps' -o tsv)
echo "VM IP: $VM_IP"
echo "Waiting for VM to be ready for SSH..."
while ! nc -z "$VM_IP" 22; do
  sleep 5
done
echo "VM is ready for SSH."
ssh -o StrictHostKeyChecking=no $VM_IP exit

scp hack/test-fetch-boot-flow-runner.sh $VM_IP:~/
scp config.json $VM_IP:~/
ssh $VM_IP "chmod +x test-fetch-boot-flow-runner.sh && ./test-fetch-boot-flow-runner.sh config.json"

if [ "$DELETE_VM" = "true" ]; then
    echo "Deleting VM $RUNNER_VM_NAME in resource group $GROUP..."
    az vm delete -g $GROUP -n $RUNNER_VM_NAME --yes --no-wait
else
    echo "Not deleting VM $RUNNER_VM_NAME as DELETE_VM is set to false."
fi
