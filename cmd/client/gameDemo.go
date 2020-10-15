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
	"github.com/Frans-Lukas/checkerboard/cmd/constants"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	NS "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	OBJ "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

const (
	defaultName = "world"
)
const posXKey = "posX"
const posYKey = "posY"

type Player struct {
	OBJ.PlayerServer
	posX     int64
	posY     int64
	objectId string
}

var playerList = make(map[string]*Player, 0)

var player = PlayerConstructor(0, 0)

var isBot = true

const PlayerObjectType = "player"

func PlayerConstructor(posX int64, posY int64) Player {
	return Player{posX: posX, posY: posY, objectId: fmt.Sprint(time.Now().UnixNano())}
}

func main() {
	// Set up a connection to the server.\

	rand.Seed(time.Now().UnixNano())
	player.posX = int64(rand.Int() % constants.MAP_SIZE)
	player.posY = int64(rand.Int() % constants.MAP_SIZE)

	port, err := strconv.Atoi(os.Args[2])
	splitCellRequirement := 5
	splitCellInterval := 5

	if err != nil {
		log.Fatalf("invalid port argument: ./gameDemo ip port")
	}

	//start cell master
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	playerServer := grpc.NewServer()
	cellMaster := objects.NewPlayer(splitCellRequirement, splitCellInterval)
	OBJ.RegisterPlayerServer(playerServer, &cellMaster)
	go func() {
		if err := playerServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve %v", err)
		}
	}()

	go func() {
		updateWorld(&cellMaster)
	}()

	conn, err := grpc.Dial(constants.CellManagerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NS.NewCellManagerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c.AddPlayerToCellWithPositions(ctx, &NS.PlayerInCellRequestWithPositions{Ip: os.Args[1], Port: int32(port), PosX: 0, PosY: 0})

	cm, err := c.RequestCellMasterWithPositions(ctx, &NS.Position{PosX: player.posX, PosY: player.posY})

	if err != nil {
		log.Fatalf("RequestCellMaster err: " + err.Error())
	}

	println("my cm has port: ", cm.Port)

	if err != nil {
		log.Fatalf("No cell master :(")
	}

	conn2, err2 := grpc.Dial(objects.ToAddress(cm.Ip, cm.Port), grpc.WithInsecure(), grpc.WithBlock())
	defer conn2.Close()
	if err2 != nil {
		log.Fatalf("did not connect: %v", err2)
	}

	cmConn := OBJ.NewPlayerClient(conn2)

	for {
		_, err = cmConn.SubscribePlayer(ctx, &OBJ.PlayerInfo{Ip: "localhost", Port: int32(port)})
		if err == nil {
			break
		}
	}

	gameLoop(cmConn, cellMaster)
}

func gameLoop(cm OBJ.PlayerClient, cellMaster objects.Player) {
	reader := bufio.NewReader(os.Stdin)

	playerList[player.objectId] = &player
	printMap(&cellMaster)
	println()
	println()
	println()
	for {
		if isBot {
			botMove()
			time.Sleep(time.Second * 3)
		} else {
			input, _ := reader.ReadString('\n')
			readInput(input)
		}

		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		//TODO check so that defer is not needed
		//defer cancel()

		_, err2 := cm.IsAlive(ctx, &OBJ.EmptyRequest{})
		if err2 != nil {
			log.Fatalf("isAlive: " + err2.Error())
		}

		_, err := cm.RequestObjectMutation(ctx, &OBJ.SingleObject{ObjectType: PlayerObjectType, ObjectId: player.objectId, PosX: int64(player.posX), PosY: int64(player.posY)})
		if err != nil {
			log.Fatalf(err.Error())
		}
		checkForPlayerUpdates(cellMaster)
		printMap(&cellMaster)
		println()
		println()
		println()
	}
}

func botMove() {
	switch rand.Int() % 4 {
	case 0:
		if player.posX+1 < constants.MAP_SIZE {
			player.posX = player.posX + 1
		}
	case 1:
		if player.posX-1 >= 0 {
			player.posX = player.posX - 1
		}
	case 2:
		if player.posY-1 >= 0 {
			player.posY = player.posY - 1
		}
	case 3:
		if player.posY+1 < constants.MAP_SIZE {
			player.posY = player.posY + 1
		}
	}

}

func checkForPlayerUpdates(cellMaster objects.Player) {
	for _, object := range *cellMaster.MutatedObjects {
		if _, ok := playerList[object.ObjectId]; ok {
			if object.ObjectId != player.objectId {
				updatePlayer(&object)
			}
		} else {
			playerList[object.ObjectId] = PlayerFromObject(&object)
		}
	}
	cellMaster.MutatedObjects = new([]OBJ.SingleObject)
}

func readInput(input string) {
	if len(input) > 0 {
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
}

func printMap(cellMaster *objects.Player) {
	for row := 0; row < constants.MAP_SIZE; row++ {
		for column := 0; column < constants.MAP_SIZE; column++ {
			printPosition(int64(row), int64(column), cellMaster)
		}
		print("\n")
	}
}

func printPosition(row int64, column int64, cellMaster *objects.Player) {
	printedPlayer := false
	printedMap := false
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
		for _, c := range *cellMaster.Cells {
			if row == c.PosY || row == c.PosY+c.Height-1 {
				print("--")
				printedMap = true
				break
			} else if column == c.PosX || column == c.PosX+c.Width-1 {
				print("| ")
				printedMap = true
				break
			}
		}
		if !printedMap {
			print("* ")
		}
	}
}

func PlayerFromObject(object *OBJ.SingleObject) *Player {
	player := PlayerConstructor(object.PosX, object.PosY)
	player.objectId = object.ObjectId
	return &player
}

func updatePlayer(object *OBJ.SingleObject) {
	playerList[object.ObjectId].posX = object.PosX
	playerList[object.ObjectId].posY = object.PosY
}

func updateWorld(player *objects.Player) {
	// poll mutatingobjects
	for {

		objectsToCellMap := make(map[string][]*OBJ.SingleObject, 0)

		objectsToMutate := make([]OBJ.SingleObject, len(*player.MutatingObjects))
		copy(objectsToMutate, *player.MutatingObjects)
		player.MutatingObjects = new([]OBJ.SingleObject)
		for _, mutatingObject := range objectsToMutate {
			mutatedObject := performGameLogic(mutatingObject)
			if constants.DebugMode {
				println("mutated object cellId: ", mutatedObject.CellId)
			}
			if objectList, ok := objectsToCellMap[mutatingObject.CellId]; ok {
				objectsToCellMap[mutatingObject.CellId] = append(objectList, &mutatedObject)
			} else {
				objectsToCellMap[mutatingObject.CellId] = make([]*OBJ.SingleObject, 0)
				objectsToCellMap[mutatingObject.CellId] = append((objectsToCellMap)[mutatingObject.CellId], &mutatedObject)
			}
		}

		for _, objectList := range objectsToCellMap {
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			//TODO check if defer is needed
			//defer cancel()
			player.BroadcastMutatedObjects(ctx, &OBJ.MultipleObjects{Objects: objectList})
		}
		player.MutatingObjects = new([]OBJ.SingleObject)
		time.Sleep(time.Second)
	}

	// if found -> apply game logic
	// broadcast update
}

type objectToUpdate struct {
	OBJ.SingleObject
}

func performGameLogic(mutatingObject OBJ.SingleObject) OBJ.SingleObject {
	switch mutatingObject.ObjectType {
	case PlayerObjectType:
		return performPlayerUpdate(mutatingObject)
	}
	return mutatingObject
}

func performPlayerUpdate(object OBJ.SingleObject) OBJ.SingleObject {
	//playerToUpdate := singleObjectToPlayer(object)
	//TODO: check for valid update
	return object
}

func singleObjectToPlayer(object OBJ.SingleObject) Player {
	playerToUpdate := PlayerConstructor(0, 0)
	playerToUpdate.posY = object.PosY
	playerToUpdate.posX = object.PosX
	return playerToUpdate
}
