#!/bin/bash

protoc --proto_path=api/proto/v1 --proto_path=third_party --go_out=plugins=grpc:pkg/generated/v1 ns.proto
