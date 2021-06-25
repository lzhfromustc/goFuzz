#!/bin/bash -e
cd "$(dirname "$0")"

echo "Copying from /usr/local/go/src/runtime into goFuzz's runtime and gooracle"

#GOROOT=$(go env GOROOT)
GOROOT=/usr/local/go
RUNTIME=$GOROOT/src/runtime

cp $RUNTIME/my* ../../runtime
cp $RUNTIME/select.go ../../runtime/select.go
cp $RUNTIME/chan.go ../../runtime/chan.go
cp $RUNTIME/runtime2.go ../../runtime/runtime2.go
cp $RUNTIME/proc.go ../../runtime/proc.go
cp $GOROOT/src/gooracle/* ../gooracle