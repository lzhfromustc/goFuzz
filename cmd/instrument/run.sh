#WORKPATH=/data/ziheng/shared/gotest/stubs/grpc/grpc-last/src/google.golang.org/grpc
WORKPATH=$1
export GOPATH=/home/sly-psu/Desktop/GO/
go install
for f in $(find $WORKPATH -iname "*.go"); do $GOPATH/bin/instrument -file=$f; done
