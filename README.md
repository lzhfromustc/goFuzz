# goFuzz

**How to use goFuzz**
1. Put goFuzz/gooracle under the same GOPATH of the target application

    After this step, gooracle is copied to $GOPATH/src/gooracle
    
2. Use goFuzz/runtime to overwrite the original runtime 
   
   Note: goFuzz/runtime is based on the runtime of go-1.14.2
   
   Remember to have a backup of the original runtime.
       
3. Use goFuzz/cmd/instrument to insert necessary function 
calls into the target application

    You can run the following command:
    `goFuzz/cmd/instrument/run.sh /PATH/TO/THE/UNIT/TEST`
    
4. Run goFuzz/cmd/fuzz

    After goFuzz/cmd/fuzz is installed, you can run $GOPATH/bin/fuzz with arguments
    
    For example:
    `$GOPATH/bin/fuzz -path=/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram -GOPATH=/data/ziheng/shared/gotest/gotest -test=TestF1 -globalTuple -output=/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram/myoutput.txt`
    
    This indicates: 
    
    run fuzzer on unit test "TestF1", 
    which is in "/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram", 
    and its GOPATH is "/data/ziheng/shared/gotest/gotest".
    Print the output to "/data/ziheng/shared/gotest/gotest/src/goFuzz/testdata/toyprogram/myoutput.txt"
    . And use global tuple strategy
    