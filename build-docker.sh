#!/bin/bash
set -e
set -x

HarborUrl=$1
ServiceName=$2
Version=$3

go env -w CGO_ENABLED=0
go env -w GOOS=linux
kratos build
go env -w CGO_ENABLED=1

docker build -t $ServiceName .
docker tag $ServiceName $HarborUrl/$ServiceName:$Version
docker tag $ServiceName $HarborUrl/$ServiceName:latest

docker push $HarborUrl/$ServiceName:$Version
docker push $HarborUrl/$ServiceName:latest

docker rmi $HarborUrl/$ServiceName:$Version
docker rmi $HarborUrl/$ServiceName:latest
docker rmi $ServiceName