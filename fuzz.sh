#!/bin/bash -e
cd "$(dirname "$0")"

TARGET_GO_MOD_DIR=$1
OUTPUT_DIR=$2
PARALLEL=$3
shift 3

if [ -z "$PARALLEL" ]
then
    PARALLEL=4
fi

# prepare target directory that used in dockerfile
cp -R $TARGET_GO_MOD_DIR ./target-tmp

docker build -t gofuzz:latest .

# clean target directory
rm -rf ./target-tmp

docker run -it \
-v $OUTPUT_DIR:/fuzz/output \
gofuzz:latest /fuzz/target /fuzz/output $PARALLEL $@