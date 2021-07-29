# docker build -f testbins.Dockerfile -t gofuzz:latest .
# docker run -it -v /Users/xsh/code/moby:/fuzz/target -v ~/docker-out:/fuzz/output  gofuzz:latest -testBinsDir /fuzz/target/testbins -chCover /fuzz/target/op-cov -outputDir /fuzz/output -parallel 5
FROM golang:1.16.4

RUN apt-get update && apt-get install -y --no-install-recommends \
		build-essential \
		curl \
		cmake \
		gcc \
		git \
		libapparmor-dev \
		libbtrfs-dev \
		libdevmapper-dev \
		libseccomp-dev \
		ca-certificates \
		e2fsprogs \
		iptables \
		pkg-config \
		pigz \
		procps \
		xfsprogs \
		xz-utils \
		\
		aufs-tools \
		vim-common \
	&& rm -rf /var/lib/apt/lists/*
    
WORKDIR /gofuzz

# copy source files to docker
COPY goFuzz ./goFuzz
RUN cd goFuzz && make build

WORKDIR /gofuzz/goFuzz

ENTRYPOINT [ "./bin/fuzz" ]