WORKPATH=/data/ziheng/shared/gotest/stubs/toy/src/fuzzer-toy
#WORKPATH=$1
for f in $(find $WORKPATH -iname "*.go"); do /data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/bin/instrument -file=$f ; done