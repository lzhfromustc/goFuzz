#WORKPATH=/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc/internal/transport
WORKPATH=$1
OUTPUT=$2
for f in $(find $WORKPATH -iname "*.go"); do /data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/bin/printOperation -file=$f -output=$OUTPUT; done