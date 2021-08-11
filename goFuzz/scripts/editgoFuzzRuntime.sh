#!/bin/bash -e
cd "$(dirname "$0")"

echo "Copying from /usr/local/go/src/runtime into goFuzz's runtime, sync and gooracle"

GOROOT=$(go env GOROOT)
RUNTIME=$GOROOT/src/runtime
echo "Runtime is:\t$RUNTIME"

cp $RUNTIME/my* ../../runtime
cp $RUNTIME/select.go ../../runtime/select.go
cp $RUNTIME/chan.go ../../runtime/chan.go
cp $RUNTIME/runtime2.go ../../runtime/runtime2.go
cp $RUNTIME/proc.go ../../runtime/proc.go
cp $GOROOT/src/gooracle/* ../gooracle
cp $GOROOT/src/time/sleep.go ../../time/sleep.go
cp $GOROOT/src/reflect/value.go ../../reflect/value.go
cp $GOROOT/src/sync/* ../../sync