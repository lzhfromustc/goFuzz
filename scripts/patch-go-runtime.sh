#!/bin/bash -e
cd "$(dirname "$0")"/.. 

GOROOT=$(go env GOROOT)
RUNTIME=$GOROOT/src/runtime

echo "Copying files into /usr/local/go/src"


cp runtime/my* $RUNTIME
cp runtime/select.go $RUNTIME/select.go
cp runtime/chan.go $RUNTIME/chan.go
cp runtime/runtime2.go $RUNTIME/runtime2.go
cp runtime/proc.go $RUNTIME/proc.go
cp -r goFuzz/gooracle $RUNTIME/..
cp -r sync $RUNTIME/..
cp -r `time` $RUNTIME/..