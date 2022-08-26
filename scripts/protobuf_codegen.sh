#!/bin/bash

# any version changes here should also be bumped in Dockerfile.buf
BUF_VERSION='1.7.0'
PROTOC_GEN_GO_VERSION='v1.28.0'
PROTOC_GEN_GO_GRPC_VERSION='1.2.0'

# buf is required see:https://docs.buf.build/installation
# go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}

if ! [[ "$0" =~ scripts/protobuf_codegen.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

if [[ $(buf --version | cut -f2 -d' ') != "${BUF_VERSION}" ]]; then
  echo "could not find buf ${BUF_VERSION}, is it installed + in PATH?"
  exit 255
fi

if [[ $(protoc-gen-go --version | cut -f2 -d' ') != "${PROTOC_GEN_GO_VERSION}" ]]; then
  echo "could not find protoc-gen-go ${PROTOC_GEN_GO_VERSION}, is it installed + in PATH?"
  exit 255
fi

if [[ $(protoc-gen-go-grpc --version | cut -f2 -d' ') != "${PROTOC_GEN_GO_GRPC_VERSION}" ]]; then
  echo "could not find protoc-gen-go-grpc ${PROTOC_GEN_GO_GRPC_VERSION}, is it installed + in PATH?"
  exit 255
fi

TARGET=$PWD/proto
if [ -n "$1" ]; then 
  TARGET="$1"
fi

# move to api directory
cd $TARGET

echo "Running protobuf fmt..."

buf format -w

echo "Running protobuf lint check..."

buf lint

if [[ $? -ne 0 ]];  then
    echo "ERROR: protobuf linter failed"
    exit 1
fi

echo "Re-generating protobuf..."

buf generate

if [[ $? -ne 0 ]];  then
    echo "ERROR: protobuf generation failed"
    exit 1
fi
