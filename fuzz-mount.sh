#!/bin/bash -e
cd "$(dirname "$0")"

TARGET_GO_MOD_DIR=$1
OUTPUT_DIR=$2
shift 2


docker build -f mount.Dockerfile -t gofuzz:latest .

# clean target directory
rm -rf ./target-tmp

docker run --rm -it \
-v $OUTPUT_DIR:/fuzz/output \
-v $TARGET_GO_MOD_DIR:/fuzz/target \
--memory-swap -1 \
gofuzz:latest /fuzz/target /fuzz/output 4 $@
