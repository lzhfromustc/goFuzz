#!/bin/bash
OUTPUT_DIR=/builder

# build native part
cd /repos/grpc-go
exclude_paths="(abi)|(fuzzer)|(integration)" /gofuzz/scripts/gen-test-bins.sh $OUTPUT_DIR/grpc/native


# instrument runtime, code and do instrumentation part
/gofuzz/scripts/patch-go-runtime.sh
cd /repos/grpc-go
/gofuzz/goFuzz/scripts/instrument.py .
exclude_paths="(abi)|(fuzzer)|(integration)" /gofuzz/scripts/gen-test-bins.sh $OUTPUT_DIR/grpc/inst