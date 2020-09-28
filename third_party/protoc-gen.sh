#!/bin/bash

protoc --proto_path=api/proto --proto_path=third_party --go_out=plugins=grpc:pkg/generated/cellmanager ns.proto
protoc --proto_path=api/proto --proto_path=third_party --go_out=plugins=grpc:pkg/generated/objects objects.proto
