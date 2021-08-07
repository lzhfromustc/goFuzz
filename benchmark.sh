#!/bin/bash -e


# Examples

# docker build -f benchmark.Dockerfile -t gfuzzbenchmark:latest .


# $ ./benchmark.sh $(pwd)/playground/native native /testbins/github.com-ethereum-go-ethereum-core
# $ ./benchmark.sh $(pwd)/playground/inst inst /testbins/github.com-ethereum-go-ethereum-core

TEST_BINS_DIR=$1
MODE=$2
shift 2
docker run -it --rm -v $TEST_BINS_DIR:/testbins gfuzzbenchmark:latest custom --dir /testbins --mode $MODE --bins $@