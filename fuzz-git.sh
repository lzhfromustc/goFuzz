#!/bin/bash -e
cd "$(dirname "$0")"

TARGET_GIT=$1
TARGET_GIT_COMMIT=$2
OUTPUT_DIR=$3
shift 3



# prepare target directory that used in dockerfile
rm -rf ./target-tmp
git clone $TARGET_GIT target-tmp
cd target-tmp && git checkout $TARGET_GIT_COMMIT && cd ..

docker build -t gofuzz:latest .

# clean target directory
rm -rf ./target-tmp

docker run -it \
-v $OUTPUT_DIR:/fuzz/output \
gofuzz:latest /fuzz/target /fuzz/output 5 $@