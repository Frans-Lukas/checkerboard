#!/bin/bash

protoc --proto_path=api/proto/v1 --proto_path=third_party --go-grpc_out=pkg/api/v1 --go_out=plugins=grpc:pkg/api/v1 ns.proto
