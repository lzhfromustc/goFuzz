#!/bin/bash -e
cd "$(dirname "$0")"

TEST_BINS_DIR=$1
OUT_DIR=$2
docker build -f testbins.Dockerfile -t gofuzz:latest .

docker run -it -v $TEST_BINS_DIR:/fuzz/target -v $OUT_DIR:/fuzz/output  gofuzz:latest -testBinsDir /fuzz/target -outputDir /fuzz/output -parallel 5
