WORKPATH=/data/ziheng/shared/gotest/stubs/prometheus/src/github.com/prometheus/prometheus
#WORKPATH=$1
for f in $(find $WORKPATH -iname "*.go"); do /data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/bin/instrument -file=$f; done