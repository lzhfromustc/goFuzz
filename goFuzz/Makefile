build:
	go mod tidy
	go build -o bin/instrument goFuzz/cmd/instrument
	go build -o bin/fuzz goFuzz/cmd/fuzz
	go build -o bin/ranktest goFuzz/cmd/ranktest
	go build -o bin/printOperation goFuzz/cmd/printOperation

test:
	go test -v goFuzz/fuzzer
