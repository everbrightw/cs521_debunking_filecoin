#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Step 1: Set environment variables
export LOTUS_PATH=~/.lotus-local-net
export LOTUS_MINER_PATH=~/.lotus-miner-local-net
export LOTUS_SKIP_GENESIS_CHECK=_yes_
export CGO_CFLAGS_ALLOW="-D__BLST_PORTABLE__"
export CGO_CFLAGS="-D__BLST_PORTABLE__"

# Step 2: Clone the lotus-local-net repository
git clone https://github.com/filecoin-project/lotus lotus-local-net
cd lotus-local-net

# Step 3: Checkout to the latest stable branch
# Replace <latest-version> with the actual latest version tag
git checkout <latest-version>

# Step 4: Remove any existing genesis sectors
rm -rf ~/.genesis-sectors

# Step 5: Build the 2k binary for Lotus
make 2k

# Step 6: Fetch the proving parameters for a 2048-byte sector size
./lotus fetch-params 2048

# Step 7: Build lotus-shed (required for Fil+ features)
make lotus-shed

# Step 8: Create BLS addresses to serve as root key holders
ROOT_KEY_1=$(./lotus-shed keyinfo new bls)
echo "Root Key 1: $ROOT_KEY_1"

ROOT_KEY_2=$(./lotus-shed keyinfo new bls)
echo "Root Key 2: $ROOT_KEY_2"

# Step 9: Pre-seal 2 sectors for the genesis block
./lotus-seed pre-seal --sector-size 2KiB --num-sectors 2

# Step 10: Create the genesis block
./lotus-seed genesis new localnet.json

# Step 11: Set the root key holders in the genesis block with a threshold of 2
./lotus-seed genesis set-signers --threshold=2 --signers $ROOT_KEY_1 --signers $ROOT_KEY_2 localnet.json

# Step 12: Add a pre-miner and an address with some funds
./lotus-seed genesis add-miner localnet.json ~/.genesis-sectors/pre-seal-t01000.json

# Step 13: Start the Lotus daemon
echo "Starting lotus daemon in background..."
nohup ./lotus daemon --lotus-make-genesis=devgen.car --genesis-template=localnet.json --bootstrap=false > lotus.log 2>&1 &
sleep 30  # Wait for lotus to initialize

# Step 14: Import the genesis miner key
./lotus wallet import --as-default ~/.genesis-sectors/pre-seal-t01000.key

# Step 15: Initialize the genesis miner
./lotus-miner init --genesis-miner --actor=t01000 --sector-size=2KiB \
  --pre-sealed-sectors=~/.genesis-sectors \
  --pre-sealed-metadata=~/.genesis-sectors/pre-seal-t01000.json --nosync

# Step 16: Start the lotus-miner
echo "Starting lotus-miner in background..."
nohup ./lotus-miner run --nosync > lotus-miner.log 2>&1 &
sleep 30

# Step 17: Import root key holder addresses
echo "$ROOT_KEY_1" > bls-root-key-1.keyinfo
echo "$ROOT_KEY_2" > bls-root-key-2.keyinfo

./lotus wallet import bls-root-key-1.keyinfo
./lotus wallet import bls-root-key-2.keyinfo