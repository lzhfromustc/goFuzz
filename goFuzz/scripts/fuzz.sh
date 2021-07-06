#!/bin/bash -e
cd "$(dirname "$0")"

TARGET_GO_MOD_DIR=$1
OUTPUT_DIR=$2
PARALLEL=$3

OP_OUT=$OUTPUT_DIR/op-out

# Instrument the target Go module source code
./instrument.py $TARGET_GO_MOD_DIR --op-out $OP_OUT

# Start fuzzing
../bin/fuzz -goModDir $TARGET_GO_MOD_DIR -chCover $OP_OUT -outputDir $OUTPUT_DIR -parallel $PARALLEL