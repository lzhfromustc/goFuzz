This is a modified GCatch that can analyze an application and print (append mode) into an output file: all structs that contain any sync primitive, directly or indirectly.

The output of this tool can be used to make gooracle/reflect.go faster.

Output format:
`v2store:watcher
v2store:watcherHub
v2store:store
v2store:EventHistory
v2store:node
v2store:ttlKeyHeap`
Each line is `pkgName:structName`. We don't print the full path of pkg, and don't care if the pkg is third party or not. structName is also purely a name.

How to run this tool:
export GOPATH=/GOPATH/To/This/Tool
cd GCatch/cmd/GCatch
go install
cd $GOPATH/bin
export GOPATH=/Path/To/App
export GO111MODULE=off
./GCatch -path=/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd -include=go.etcd.io/etcd 
 -output=/data/ziheng/shared/gotest/stubs/etcd/src/go.etcd.io/etcd/struct.txt 
 # you can add -r flag to recursively scan all child pkg in -path, but this may take lots of time and the output will have redundant lines
 