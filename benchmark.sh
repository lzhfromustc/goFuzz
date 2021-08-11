#!/bin/bash -e


# Examples

# docker build -f benchmark.Dockerfile -t gfuzzbenchmark:latest .


# $ ./benchmark.sh $(pwd)/playground/native native /testbins/github.com-ethereum-go-ethereum-core
# $ ./benchmark.sh $(pwd)/playground/inst inst /testbins/github.com-ethereum-go-ethereum-core
# custom --dir /testbins --mode $MODE --bins $@
docker build -f benchmark.Dockerfile -t gfuzzbenchmark:latest .
docker run -it --rm -v $(pwd)/playground:/playground \
-v $(pwd)/benchmark:/benchmark \
gfuzzbenchmark:latest 