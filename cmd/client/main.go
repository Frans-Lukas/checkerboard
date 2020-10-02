/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"fmt"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	NS "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	OBJ "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
	port = int32(50052)
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NS.NewCellManagerClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.CreateCell(ctx, &NS.CellRequest{CellId: "new id"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %t", r.WasPerformed)


	// Setup player server
	lis, err := net.Listen("tcp", ":" + fmt.Sprint(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	player1 := objects.NewPlayer()
	OBJ.RegisterPlayerServer(s, &player1)

	go func(s grpc.Server) {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve %v", err)
		}
	}(*s)

	// connect Player to nameServer
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := c.AddPlayerToCell(ctx, &NS.PlayerInCellRequest{Ip: "localhost", Port:port, CellId:"new id"})
	if err != nil {
		log.Fatalf("could not add player: %v", err)
	} else if !response.Succeeded {
		log.Fatalf("failed to add player to cellmanager")
	}

	// connect player and cellmaster
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cellMasterReply, err := c.RequestCellMaster(ctx, &NS.CellMasterRequest{CellId:"new id"})
	if err != nil {
		log.Fatalf("could not request cellmaster: %v", err)
	} else if cellMasterReply.Port == -1 {
		log.Fatalf("did not receive a cellmaster")
	}
}
