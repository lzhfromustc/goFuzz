echo "Copying files into /usr/local/go/src/runtime"
echo "select.go, chan.go, runtime2.go will be overwritten. They are copied to *.backup"
RUNTIME=/usr/local/go/src/runtime
cp ../../runtime/my* $RUNTIME
mv $RUNTIME/select.go $RUNTIME/select.go.backup
mv $RUNTIME/chan.go $RUNTIME/chan.go.backup
mv $RUNTIME/runtime2.go $RUNTIME/runtime2.go.backup
cp ../../runtime/select.go $RUNTIME/select.go
cp ../../runtime/chan.go $RUNTIME/chan.go
cp ../../runtime/runtime2.go $RUNTIME/runtime2.go