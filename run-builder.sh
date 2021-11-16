# This script will automatically generate the corresponding test binary files(with and without instrumented go runtime) to the ./tmp/builder

docker build -f docker/builder/Dockerfile -t gfuzzbuilder:latest .

CWD=$(pwd)
docker run --rm -it \
-v "${CWD}/tmp/builder":/builder \
-v "${CWD}/tmp/pkgmod":/go/pkg/mod \
gfuzzbuilder:latest