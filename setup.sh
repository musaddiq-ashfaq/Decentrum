#!/bin/bash
# save as setup.sh in decentralized-social-media directory

# Start from the project root
cd "$(dirname "$0")"

# Ensure fabric-samples is in the parent directory
if [ ! -d "../fabric-samples" ]; then
    echo "Error: fabric-samples directory not found in parent directory"
    exit 1
fi

# Start Fabric test network
cd ../fabric-samples/test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca

# Deploy chaincode
./network.sh deployCC -ccn social_media -ccp ../../decentralized-social-media/chaincode/social_media -ccl go

# Copy connection profile
mkdir -p ../../decentralized-social-media/network
cp organizations/peerOrganizations/org1.example.com/connection-org1.yaml ../../decentralized-social-media/network/

# Start IPFS daemon
# ipfs daemon &

# Setup backend
cd ../../decentralized-social-media/backend
go mod tidy
go run main.go