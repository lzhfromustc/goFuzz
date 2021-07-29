#!/bin/bash -e
cd "$(dirname "$0")"

echo "Copying files into /usr/local/go/src"
echo "In runtime, select.go, chan.go, runtime2.go, proc.go will be overwritten. They are copied to *.backup"

GOROOT=$(go env GOROOT)
RUNTIME=$GOROOT/src/runtime

cp ../../runtime/my* $RUNTIME
mv $RUNTIME/select.go $RUNTIME/select.go.backup
mv $RUNTIME/chan.go $RUNTIME/chan.go.backup
mv $RUNTIME/runtime2.go $RUNTIME/runtime2.go.backup
mv $RUNTIME/proc.go $RUNTIME/proc.go.backup

cp ../../runtime/select.go $RUNTIME/select.go
cp ../../runtime/chan.go $RUNTIME/chan.go
cp ../../runtime/runtime2.go $RUNTIME/runtime2.go
cp ../../runtime/proc.go $RUNTIME/proc.go
cp ../../reflect/value.go $RUNTIME/../reflect/value.go

cp -r ../../time $RUNTIME/..
cp -r ../gooracle $RUNTIME/..
cp -r ../../sync $RUNTIME/..
