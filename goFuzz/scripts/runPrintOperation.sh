#WORKPATH=/data/ziheng/shared/gotest/stubs/prometheus/src/github.com/prometheus/prometheus
WORKPATH=$1 # Absolute path to the already instrumented application
OUTPUT=$2 # Absolute path to a file that you want to print out the results. Positions of all channel operations in the application will be printed in this file
for f in $(find $WORKPATH -iname "*.go"); do /data/ziheng/shared/gotest/gotest/src/goFuzz/goFuzz/bin/printOperation -file=$f -output=$2; done