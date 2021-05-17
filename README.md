- [Go Fuzz Project](#go-fuzz-project)
  - [Project Structure](#project-structure)
  - [Prerequisite](#prerequisite)
  - [Dev Setup](#dev-setup)

# Go Fuzz Project

`Go Fuzz` aims find concurrency bugs in Golang program at runtime.

## Project Structure
- goFuzz: contains packages/utilities that used to/in instrument/instrumented program.
- runtime: patched Golang runtime (1.14) for recording necessary information to find bugs.

## Prerequisite
- Python 3.6+
- Golang 1.14+

## Dev Setup

1. Copy gooracle 
- Go Module: Put `goFuzz/gooracle` under the same root of go module of the target application
- Without Go Module: Put `goFuzz/gooracle` to $GOPATH/src/gooracle
    
2. Use goFuzz/runtime to overwrite the original runtime 
Note: goFuzz/runtime is based on the runtime of go-1.14.2
Remember to have a backup of the original runtime.

2. Build utilities

```bash
$ cd goFuzz

# This step will
# 1. Download dependencies by go mod tidy
# 2. Generate binary program 'instrument'
# 3. Generate binary program 'fuzz'
$ make build
```

4. Instrument target application
```bash
$ ./goFuzz/scripts/instrument.py [folder contains Golang source code]
```
    
5. Run goFuzz/cmd/fuzz
    
For example:
`./goFuzz/bin/fuzz -path=/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram -GOPATH=/data/ziheng/shared/gotest/gotest -test=TestF1 -globalTuple -output=/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram/myoutput.txt`

This indicates: 

run fuzzer on unit test "TestF1", 
which is in "/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram", 
and its GOPATH is "/data/ziheng/shared/gotest/gotest".
Print the output to "/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram/myoutput.txt"
. And use global tuple strategy
    