#!/bin/env bash

PKG='github.com/deltachat-bot/deltabot-cli-go'

echo "Checking code with gofmt..."
OUTPUT=`gofmt -d .`
if [ -n "$OUTPUT" ]
then
    echo "$OUTPUT"
    exit 1
fi

echo "Checking code with golangci-lint..."
if ! command -v golangci-lint &> /dev/null
then
    echo "golangci-lint not found, installing..."
    # binary will be $(go env GOPATH)/bin/golangci-lint
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0
fi

cd v2
if ! golangci-lint run
then
    exit 1
fi
cd ..

if ! command -v deltachat-rpc-server &> /dev/null
then
    echo "deltachat-rpc-server not found, installing..."
    curl -L https://github.com/chatmail/core/releases/download/v2.44.0/deltachat-rpc-server-x86_64-linux --output deltachat-rpc-server
    chmod +x deltachat-rpc-server
    export PATH=`pwd`:"$PATH"
fi

if ! command -v courtney &> /dev/null
then
    echo "courtney not found, installing..."
    go install github.com/dave/courtney@master
fi

# test examples
for i in examples/*
do
    echo "Testing: $i"
    cd "$i"
    go mod edit -replace=$PKG/v2=../../v2
    go mod tidy
    if ! golangci-lint run
    then
        exit 1
    fi
    if ! go build -v
    then
        exit 1
    fi
    if ! go test -v
    then
        exit 1
    fi
    go mod edit -dropreplace $PKG/v2
    cd ../..
done
echo "Done testing examples"

cd v2
# add -t="-parallel=1" to avoid running tests in parallel
courtney -v -t="./..." -o coverage.out
go tool cover -func=coverage.out -o=../coverage-percent.out
