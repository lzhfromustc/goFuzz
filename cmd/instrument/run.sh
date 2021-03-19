WORKPATH=/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd
for f in $(find $WORKPATH -iname "*.go"); do /data/ziheng/shared/gotest/gotest/bin/instrument -file=$f; done