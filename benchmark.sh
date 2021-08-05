#!/bin/bash -e

docker build -f benchmark.Dockerfile -t gfuzzbenchmark:latest .
docker run -it --rm gfuzzbenchmark:latest