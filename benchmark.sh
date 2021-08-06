#!/bin/bash -e


# Examples
# To benchmark native speed:   
# $ ./benchmark.sh native

# To benchmark gooracle speed: 
# $ ./benchmark.sh inst


docker build -f benchmark.Dockerfile -t gfuzzbenchmark:latest .
docker run -it --rm gfuzzbenchmark:latest $1