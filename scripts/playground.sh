# Start an docker container with goFuzz patched and source code to play around
# Default output folder is /playground/output


docker build -f playground.Dockerfile -t gofuzzpg:latest .

CWD=$(pwd)
docker run --rm -it \
-v "${CWD}/playground":"/playground" \
gofuzzpg:latest bash