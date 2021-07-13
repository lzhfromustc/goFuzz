WORKPATH=/data/ziheng/shared/gotest/stubs/kubernetes/kubernetes-last/src/k8s.io/kubernetes
#WORKPATH=$1
for f in $(find $WORKPATH -iname "*.go"); do /data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/bin/instrument -file=$f ; done