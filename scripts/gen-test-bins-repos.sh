MODE=$1

if [ $MODE = "native" ]; then
    for f in *; do
    cd $f
        go list ./...
        exclude_paths="(abi)|(fuzzer)|(integration)" /gofuzz/scripts/gen-test-bins.sh /playground/out/$f/native
    cd ..
    done
else
    /gofuzz/scripts/patch-go-runtime.sh
    for f in *; do
    cd $f
        go list ./...
        /gofuzz/goFuzz/scripts/instrument.py .
        exclude_paths="(abi)|(fuzzer)|(integration)" /gofuzz/scripts/gen-test-bins.sh /playground/out/$f/inst
    cd ..
    done
fi

