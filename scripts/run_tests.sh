#!/bin/env bash

OUTPUT=`gofmt -d .`
if [ -n "$OUTPUT" ]
then
    echo "$OUTPUT"
    exit 1
fi

if ! command -v deltachat-rpc-server &> /dev/null
then
    echo "deltachat-rpc-server could not be found, installing..."
    curl -L "https://github.com/deltachat/deltachat-core-rust/releases/download/v1.112.6/deltachat-rpc-server-x86_64" -o ~/.cargo/bin/deltachat-rpc-server
    chmod +x ~/.cargo/bin/deltachat-rpc-server

fi

if ! command -v courtney &> /dev/null
then
    echo "courtney could not be found, installing..."
    go install github.com/dave/courtney@latest
fi

# test examples
for i in ./examples/*.go; do go build -v "$i"; done

courtney -v -t="./..."
go tool cover -func=coverage.out -o=coverage-percent.out
