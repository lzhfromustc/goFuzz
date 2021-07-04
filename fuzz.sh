#!/bin/bash -e
cd "$(dirname "$0")"

TARGET_GO_MOD_DIR=$1
OUTPUT_DIR=$2
PARALLEL=$3

if [ -z "$PARALLEL" ]
then
    PARALLEL=4
fi

docker build -t gofuzz:latest .

docker run -it \
-v $TARGET_GO_MOD_DIR:/fuzz/target \
-v $OUTPUT_DIR:/fuzz/output \
gofuzz:latest /fuzz/target /fuzz/output $PARALLEL