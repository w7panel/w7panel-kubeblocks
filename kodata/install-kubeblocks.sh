#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

# Check if KO_DATA_PATH is set
if [ -z "$KO_DATA_PATH" ]; then
    echo "Error: KO_DATA_PATH environment variable is not set"
    exit 1
fi

# Execute kubectl apply for CRDs 1.0.1
echo "Executing: kubectl apply -f $KO_DATA_PATH/crds --server-side"
echo "---"
kubectl apply -f "$KO_DATA_PATH/crds/" --server-side
if [ $? -ne 0 ]; then
    echo "Error: Failed to apply CRDs 1.0.1"
    exit 1
fi
echo "CRDs 1.0.1 applied successfully"
echo "---"



# Execute helm upgrade
echo "Executing: helm upgrade kubeblocks $KO_DATA_PATH/charts/kubeblocks-1.0.1.tgz -n kb-system --create-namespace --install"
echo "---"
helm upgrade kubeblocks "$KO_DATA_PATH/charts/kubeblocks-1.0.1.tgz" -n kb-system --create-namespace --install
if [ $? -ne 0 ]; then
    echo "Error: Helm command failed"
    exit 1
fi

echo "KubeBlocks installed successfully"
