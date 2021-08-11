FROM golang:1.16.4

RUN git clone https://github.com/Z3Prover/z3 /repos/z3
WORKDIR /repos/z3
RUN python scripts/mk_make.py \ 
  && cd build \
  && make \
  && make install

RUN git clone https://github.com/system-pclub/GCatch.git /go/src/github.com/system-pclub/GCatch

RUN /go/src/github.com/system-pclub/GCatch/GCatch/install.sh

RUN git clone https://github.com/ethereum/go-ethereum.git /go/src/github.com/ethereum/go-ethereum
WORKDIR /repos

# GCatch -path=/go/src/github.com/ethereum/go-ethereum -include=github.com/ethereum/go-ethereum -output=./struct.txt -compile-error