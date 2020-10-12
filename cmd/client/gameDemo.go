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
	"bufio"
	"context"
	"fmt"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	NS "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	OBJ "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
	port        = int32(50052)
)

func main() {
	// Set up a connection to the server.\

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NS.NewCellManagerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	port, err := strconv.Atoi(os.Args[2])

	if err != nil {
		log.Fatalf("invalid port argument: ./gameDemo ip port")
	}

	c.AddPlayerToCellWithPositions(ctx, &NS.PlayerInCellRequestWithPositions{Ip: os.Args[1], Port: int32(port), PosX: 0, PosY: 0})

	cm, err := c.RequestCellMasterWithPositions(ctx, &NS.Position{PosX: 0, PosY: 0})

	if err != nil {
		log.Fatalf("No cell master :(")
	}

	conn2, err2 := grpc.Dial(objects.ToAddress(*cm), grpc.WithInsecure(), grpc.WithBlock())
	defer conn2.Close()
	if err2 != nil {
		log.Fatalf("did not connect: %v", err2)
	}

	cmConn := OBJ.NewCellMasterClient(conn)
	gameLoop(cmConn)
}

type Player struct {
	posXKey  string
	posX     int
	posYKey  string
	posY     int
	objectId string
}

type PlayerServer struct {
	OBJ.PlayerClient
}

func PlayerConstructor(posX int, posY int) Player {
	return Player{posX: posX, posY: posY, objectId: fmt.Sprint(time.Now().UnixNano()), posXKey: "posX", posYKey: "posY"}
}

var playerList = make(map[string]*Player, 0)

var player = PlayerConstructor(0, 0)

const PlayerObjectType = "player"

func (c *PlayerServer) SendUpdate(ctx context.Context, in *OBJ.MultipleObjects, opts ...grpc.CallOption) (*OBJ.EmptyReply, error) {
	for _, object := range (*in).Objects {
		if val, ok := playerList[object.ObjectId]; ok {
			for keyIndex, key := range object.UpdateKey {
				updatePlayer(key, val, object, keyIndex)
			}

		} else {
			playerList[object.ObjectId] = PlayerFromObject(object)
		}
	}
	printMap()
	return &OBJ.EmptyReply{}, nil
}

func PlayerFromObject(object *OBJ.SingleObject) *Player {
	player := PlayerConstructor(0, 0)
	for keyIndex, key := range object.UpdateKey {
		updatePlayer(key, &player, object, keyIndex)
	}
	player.objectId = object.ObjectId
	return &player
}

func updatePlayer(key string, val *Player, object *OBJ.SingleObject, keyIndex int) {
	switch key {
	case val.posXKey:
		newX, err := strconv.Atoi(object.NewValue[keyIndex])
		if err != nil {
			log.Fatalf("invalid posx")
		}
		playerList[object.ObjectId].posX = newX
	case val.posYKey:
		newY, err := strconv.Atoi(object.NewValue[keyIndex])
		if err != nil {
			log.Fatalf("invalid posy")
		}
		playerList[object.ObjectId].posY = newY
	}
}

func gameLoop(cm OBJ.CellMasterClient) {
	reader := bufio.NewReader(os.Stdin)

	playerList[player.objectId] = &player

	// Setup player server
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
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

	printMap()
	for {
		input, _ := reader.ReadString('\n')
		readInput(input)

		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		//TODO check so that defer is not needed
		//defer cancel()
		cm.SendUpdate(ctx, &OBJ.SingleObject{ObjectType: PlayerObjectType, ObjectId: player.objectId, UpdateKey: []string{player.posXKey, player.posYKey}, NewValue: []string{fmt.Sprint(player.posX), fmt.Sprint(player.posY)}})

		printMap()
	}
}

func readInput(input string) {
	if input[0] == 'w' {
		player.posY--
	} else if input[0] == 's' {
		player.posY++
	} else if input[0] == 'a' {
		player.posX--
	} else if input[0] == 'd' {
		player.posX++
	}
}

func printMap() {
	const MAP_SIZE = 5
	for row := 0; row < MAP_SIZE; row++ {
		for column := 0; column < MAP_SIZE; column++ {
			printPosition(row, column)
		}
		print("\n")
	}
}

func printPosition(row int, column int) {
	printedPlayer := false
	if row == player.posY && column == player.posX {
		print("P ")
		printedPlayer = true
	} else {
		for _, player := range playerList {
			if row == player.posY && column == player.posX {
				print("O ")
				printedPlayer = true
				break
			}
		}
	}
	if !printedPlayer {
		print("* ")
	}
}
