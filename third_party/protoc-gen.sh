#!/bin/bash

protoc --proto_path=api/proto/v1 --proto_path=third_party --go-grpc_out=pkg/generated/v1 --go_out=plugins=grpc:pkg/generated/v1 ns.proto
