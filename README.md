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

3. Build utilities

```bash
$ cd goFuzz

# This step will
# 1. Download dependencies by go mod tidy
# 2. Generate binary program 'instrument'
# 3. Generate binary program 'fuzz'
$ make build
```

4. Overwrite your runtime

```bash
cd ./goFuzz/scripts
sudo ./editRuntime.sh
```

5. Instrument target application
```bash
# Shihao's tool
$ ./goFuzz/scripts/instrument.py [folder contains Golang source code]
# Ziheng's way
cd ./goFuzz/cmd/instrument
go install
cd $GOPATH/bin
./instrument -file=/Full/Path/Of/The/File/You/Want/To/Instrument
```
    
6. Run goFuzz/cmd/fuzz
    
For example:
`./fuzz -path=/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc -GOPATH=/data/ziheng/shared/gotest/stubs/grpc/grpc-last -output=/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc/myoutput.txt -test=TestStateTransitions_MultipleAddrsEntersReady`

This indicates: 

run fuzzer on unit test "TestStateTransitions_MultipleAddrsEntersReady()", 

which is in "/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc", 

and its GOPATH is "/data/ziheng/shared/gotest/stubs/grpc/grpc-last".

Print the output to "/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc/myoutput.txt". 

And use global tuple strategy

BTW, we need to remove "(s)" before "TestStateTransitions_MultipleAddrsEntersReady()" manually. This is a special problem with grpc
    