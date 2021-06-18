#WORKPATH=/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc
WORKPATH=$1
export GOPATH=/data/ziheng/shared/gotest/gotest
go install
for f in $(find $WORKPATH -iname "*.go"); do /data/ziheng/shared/gotest/gotest/bin/instrument -file=$f; done