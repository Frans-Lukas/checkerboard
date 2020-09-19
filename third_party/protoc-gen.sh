#!/bin/bash

protoc --proto_path=api/proto --proto_path=third_party --go_out=plugins=grpc:pkg/generated ns.proto
