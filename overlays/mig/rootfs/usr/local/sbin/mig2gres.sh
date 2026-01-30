#!/usr/bin/bash

set -euo pipefail

# Generates NVIDIA MIG device entries for Slurm's gres.conf.
# Queries MIG GPU instances and maps them to /dev/nvidia-caps/ device files.
#
# Process: nvidia-smi -> extract GPU/GI/profile -> lookup minor in 
# /proc/driver/nvidia-caps/mig-minors -> map to device file -> output gres format


if ! command -v nvidia-smi &> /dev/null; then
  echo "Error: nvidia-smi command not found" >&2
  exit 1
fi

hostname=$(hostname -s)

nvidia-smi mig -lgi | grep -E '^\|' | grep -v '===' | grep -v 'GPU.*Name' | while read line; do
  gpu=$(echo "$line" | awk '{print $2}')
  name=$(echo "$line" | awk '{print $4}')
  gi=$(echo "$line" | awk '{print $6}')  # Instance ID is column 6

  [[ "$gpu" =~ ^[0-9]+$ ]] || continue

  pattern="gpu${gpu}/gi${gi}/access"
  minor=$(grep "^${pattern} " /proc/driver/nvidia-caps/mig-minors | awk '{print $2}')

  echo "NodeName=${hostname} Name=gpu Type=${name} File=/dev/nvidia-caps/nvidia-cap${minor}"
done