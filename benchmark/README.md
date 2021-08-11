/gofuzz/scripts/patch-go-runtime.sh
exclude_paths="(abi)|(fuzzer)" /gofuzz/scripts/gen-test-bins.sh /playground/out/go-eth/native
exclude_paths="(abi)|(fuzzer)" /gofuzz/scripts/gen-test-bins.sh /playground/out/go-eth/inst

exclude_paths="(abi)|(fuzzer)|(integration)" /gofuzz/scripts/gen-test-bins.sh /playground/out/grpc/native
exclude_paths="(abi)|(fuzzer)|(integration)" /gofuzz/scripts/gen-test-bins.sh /playground/out/grpc/inst

exclude_paths="(abi)|(fuzzer)|(integration)" /gofuzz/scripts/gen-test-bins.sh /playground/out/etcd/native

------

/benchmark/run.py custom --dir /playground/out/go-eth/native --mode native --bins-list-file /benchmark/go-eth-list
/benchmark/run.py custom --dir /playground/out/grpc/native --mode native --bins-list-file /benchmark/grpc-list
/benchmark/run.py custom --dir /playground/out/kubernetes/native --mode native --bins-list-file /benchmark/k8s-list
/benchmark/run.py custom --dir /playground/out/prometheus/native --mode native --bins-list-file /benchmark/prom-list
/benchmark/run.py custom --dir /playground/out/moby/native --mode native --bins-list-file /benchmark/docker-list
/benchmark/run.py custom --dir /playground/out/etcd/native --mode native --bins-list-file /benchmark/etcd-list

--------

/benchmark/run.py custom --dir /playground/out/prometheus/inst --mode inst --bins-list-file /benchmark/prom-list
/benchmark/run.py custom --dir /playground/out/tidb/inst --mode inst --bins-list-file /benchmark/tidb-list
/benchmark/run.py custom --dir /playground/out/grpc-go/inst --mode inst --bins-list-file /benchmark/grpc-list
/benchmark/run.py custom --dir /playground/out/go-ethereum/inst --mode inst --bins-list-file /benchmark/go-eth-list
/benchmark/run.py custom --dir /playground/out/kubernetes/inst --mode inst --bins-list-file /benchmark/k8s-list
/benchmark/run.py custom --dir /playground/out/moby/inst --mode inst --bins-list-file /benchmark/docker-list
/benchmark/run.py custom --dir /playground/out/etcd/inst --mode inst --bins-list-file /benchmark/etcd-list



# Kubernetes: files need to receover after instrumentation
vendor/k8s.io/apimachinery/pkg/api/resource/quantity.go
vendor/k8s.io/client-go/util/workqueue/metrics.go
vendor/sigs.k8s.io/structured-merge-diff/v4/value/list.go
vendor/k8s.io/apimachinery/pkg/util/net/interface.go
vendor/k8s.io/apimachinery/pkg/apis/meta/v1/duration.go
vendor/k8s.io/apimachinery/pkg/apis/meta/v1/micro_time.go
vendor/k8s.io/apimachinery/pkg/apis/meta/v1/time.go
vendor/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1/types_jsonschema.go
vendor/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1/types_jsonschema.go
pkg/controller/nodelifecycle/scheduler/taint_manager.go
vendor/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1/types_jsonschema.go
pkg/controller/nodelifecycle/scheduler/taint_manager.go

cp ../kubernetes-copy/

