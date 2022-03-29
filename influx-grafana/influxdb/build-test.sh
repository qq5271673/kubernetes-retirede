#! /bin/bash

pushd $( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
docker build -t vish/heapster_influxdb:e2e_test1 .
docker push vish/heapster_influxdb:e2e_test1
popd
