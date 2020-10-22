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
	"github.com/Frans-Lukas/checkerboard/cmd/mapDrawer"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	NS "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	OBJ "github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
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

var thisPlayer = objects.NewPlayer(constants.SplitCellRequirement, constants.SplitCellInterval)

var isBot = true

var playersMap = mapDrawer.MapInfo{}

const PlayerObjectType = "player"

func PlayerConstructor(posX int64, posY int64) Player {
	return Player{posX: posX, posY: posY, objectId: fmt.Sprint(time.Now().UnixNano())}
}

func main() {
	// Set up a connection to the server.\

	rand.Seed(time.Now().UnixNano())

	port, err := strconv.Atoi(os.Args[2])

	// test map drawing
	playersMap = mapDrawer.SetupMap(strconv.Itoa(port), constants.MAP_SIZE*constants.IconSize, constants.MAP_SIZE*constants.IconSize)

	if len(os.Args) >= 4 {
		isBot = false
	}

	if err != nil {
		log.Fatalf("invalid port argument: ./gameDemo ip port")
	}

	//start cell master
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	playerServer := grpc.NewServer()

	thisPlayer.Port = port
	thisPlayer.Ip = "localhost"
	thisPlayer.PosX = int64(rand.Int() % constants.MAP_SIZE)
	thisPlayer.PosY = int64(rand.Int() % constants.MAP_SIZE)
	thisPlayer.ObjectId = objects.ToAddress(thisPlayer.Ip, int32(thisPlayer.Port))
	OBJ.RegisterPlayerServer(playerServer, thisPlayer)
	go func() {
		if err := playerServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve %v", err)
		}
	}()

	conn, err := grpc.Dial(constants.CellManagerAddress, grpc.WithInsecure(), grpc.WithTimeout(time.Millisecond*constants.DialTimeoutMilli))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	cellManager := NS.NewCellManagerClient(conn)

	go func() {
		updateWorld(thisPlayer, &cellManager)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = cellManager.AddPlayerToCellWithPositions(ctx, &NS.PlayerInCellRequestWithPositions{Ip: os.Args[1], Port: int32(port), PosX: thisPlayer.PosX, PosY: thisPlayer.PosY})

	if err != nil {
		log.Fatalf("Failed to add player to cell: ", err.Error())
	}

	println("requesting cm")
	RequestNewCellMaster(cellManager, thisPlayer)
	println("requested cm")

	println("my objectid is: ", thisPlayer.ObjectId)

	for {

		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		_, err = (*thisPlayer.CellMaster).SubscribePlayer(ctx, &OBJ.PlayerInfo{
			Ip:       "localhost",
			Port:     int32(port),
			PosX:     thisPlayer.PosX,
			PosY:     thisPlayer.PosY,
			ObjectId: thisPlayer.ObjectId,
		})
		if err == nil {
			break
		} else {
			RequestNewCellMaster(cellManager, thisPlayer)
			println("got error ", err.Error())
			time.Sleep(time.Second)
		}
	}
	gameLoop(thisPlayer, cellManager)
}

func gameLoop(thisPlayer *objects.Player, cellManager NS.CellManagerClient) {
	reader := bufio.NewReader(os.Stdin)

	printMap(thisPlayer)
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
			println("Players: ")
			for _, player := range playerList {
				println(player.objectId)
			}
			println()

		}

		//TODO check so that defer is not needed
		//defer cancel()

		for thisPlayer.CellMaster == nil {
			println("Requesting cellmaster")
			RequestNewCellMaster(cellManager, thisPlayer)
			if thisPlayer.CellMaster != nil {

				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				println("Got cellmaster, subscribing")
				_, err := (*thisPlayer.CellMaster).SubscribePlayer(ctx, &OBJ.PlayerInfo{
					Ip:       thisPlayer.Ip,
					Port:     int32(thisPlayer.Port),
					PosX:     thisPlayer.PosX,
					PosY:     thisPlayer.PosY,
					ObjectId: thisPlayer.ObjectId,
				})

				for err != nil {
					println("Failed to subscribe: ", err.Error())
					RequestNewCellMaster(cellManager, thisPlayer)

					if thisPlayer.CellMaster != nil {
						ctx, _ := context.WithTimeout(context.Background(), time.Second)
						_, err = (*thisPlayer.CellMaster).SubscribePlayer(ctx, &OBJ.PlayerInfo{
							Ip:       thisPlayer.Ip,
							Port:     int32(thisPlayer.Port),
							PosX:     thisPlayer.PosX,
							PosY:     thisPlayer.PosY,
							ObjectId: thisPlayer.ObjectId,
						})
					}
				}

				time.Sleep(time.Second)

			}
			time.Sleep(time.Second)
		}
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		_, err := (*thisPlayer.CellMaster).RequestObjectMutation(ctx, &OBJ.SingleObject{
			ObjectType: PlayerObjectType,
			ObjectId:   thisPlayer.ObjectId,
			PosX:       int64(thisPlayer.PosX),
			PosY:       int64(thisPlayer.PosY),
		})
		if err != nil {
			println("request object mutation failed: %v", err.Error())
		}
		checkForPlayerUpdates(thisPlayer)
		printMap(thisPlayer)
		println()
		println()
		println()
	}
}

func RequestNewCellMaster(cellManager NS.CellManagerClient, thisPlayer *objects.Player) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	_, err := cellManager.AddPlayerToCellWithPositions(ctx, &NS.PlayerInCellRequestWithPositions{Ip: thisPlayer.Ip, Port: int32(thisPlayer.Port), PosX: thisPlayer.PosX, PosY: thisPlayer.PosY})

	if err != nil {
		log.Println("RequestNewCellMaster: failed AddyerToCellWithPosition: ", err.Error())
		return
	}

	cm, err := cellManager.RequestCellMasterWithPositions(ctx, &NS.Position{PosX: thisPlayer.PosX, PosY: thisPlayer.PosY})
	if err != nil {
		log.Println("did not find new cell master: %v", err)
		return
	}
	conn, err2 := grpc.Dial(objects.ToAddress(cm.Ip, cm.Port), grpc.WithInsecure(), grpc.WithTimeout(time.Millisecond*constants.DialTimeoutMilli))
	if err2 != nil {
		log.Println("did not connect to new cell master: %v", err2)
	}
	cmConn := OBJ.NewPlayerClient(conn)
	thisPlayer.CellMaster = &cmConn
	thisPlayer.Connection = conn
	println(cmConn)
}

func botMove() {
	switch rand.Int() % 4 {
	case 0:
		if thisPlayer.PosX+1 < constants.MAP_SIZE {
			thisPlayer.PosX = thisPlayer.PosX + 1
		}
	case 1:
		if thisPlayer.PosX-1 >= 0 {
			thisPlayer.PosX = thisPlayer.PosX - 1
		}
	case 2:
		if thisPlayer.PosY-1 >= 0 {
			thisPlayer.PosY = thisPlayer.PosY - 1
		}
	case 3:
		if thisPlayer.PosY+1 < constants.MAP_SIZE {
			thisPlayer.PosY = thisPlayer.PosY + 1
		}
	}

}

func checkForPlayerUpdates(cellMaster *objects.Player) {
	for _, object := range *cellMaster.MutatedObjects {
		if _, ok := playerList[object.ObjectId]; ok {
			if object.ObjectId != thisPlayer.ObjectId {
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
			thisPlayer.PosY--
		} else if input[0] == 's' {
			thisPlayer.PosY++
		} else if input[0] == 'a' {
			thisPlayer.PosX--
		} else if input[0] == 'd' {
			thisPlayer.PosX++
		}
	}
}

func printMap(cellMaster *objects.Player) {
	playersMap.ClearMap()
	playersMap.DrawClient(int(thisPlayer.PosX), int(thisPlayer.PosY), true)
	playersMap.SaveMapAsPNG()

	for row := 0; row < constants.MAP_SIZE; row++ {
		if row < 10 {
			print(row, " ")
		} else {
			print(row)
		}
		//fmt.Fprintf("%3d", row)
		for column := 0; column < constants.MAP_SIZE; column++ {
			printPosition(int64(row), int64(column), cellMaster)
		}
		print(row, "\n")
	}
}

func printPosition(row int64, column int64, cellMaster *objects.Player) {
	printedPlayer := false
	printedMap := false
	if row == thisPlayer.PosY && column == thisPlayer.PosX {
		print("P ")
		printedPlayer = true
	} else {
		for _, player := range playerList {
			if row == player.posY && column == player.posX && player.objectId != thisPlayer.ObjectId {
				print("O ")
				printedPlayer = true
				break
			}
		}
	}

	if !printedPlayer {
		c := cellMaster.Cells
		if c != nil {
			if row == c.PosY && column == c.PosX {
				print("+-")
				printedMap = true
			} else if row == c.PosY+c.Height-1 && column == c.PosX+c.Width-1 {
				print("+ ")
				printedMap = true
			} else if row == c.PosY+c.Height-1 && column == c.PosX {
				print("+-")
				printedMap = true
			} else if row == c.PosY && column == c.PosX+c.Width-1 {
				print("+ ")
				printedMap = true
			} else if row == c.PosY || row == c.PosY+c.Height-1 {
				if column >= c.PosX && column < c.PosX+c.Width {
					print("--")
					printedMap = true
				}
			} else if column == c.PosX || column == c.PosX+c.Width-1 {
				if row >= c.PosY && row < c.PosY+c.Height {
					print("| ")
					printedMap = true
				}
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

func updateWorld(player *objects.Player, cellManager *NS.CellManagerClient) {
	// poll mutatingobjects
	for {

		objectsToCellMap := make(map[string][]*OBJ.SingleObject, 0)

		objectsToMutate := make([]OBJ.SingleObject, len(*player.MutatingObjects))
		copy(objectsToMutate, *player.MutatingObjects)
		player.MutatingObjects = new([]OBJ.SingleObject)
		for _, mutatingObject := range objectsToMutate {
			mutatedObject := performGameLogic(mutatingObject, cellManager)
			if len(mutatedObject.CellId) == 0 {
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

func performGameLogic(mutatingObject OBJ.SingleObject, cellManager *NS.CellManagerClient) OBJ.SingleObject {
	switch mutatingObject.ObjectType {
	case PlayerObjectType:
		return performPlayerUpdate(mutatingObject, cellManager)
	}
	return mutatingObject
}

func performPlayerUpdate(object OBJ.SingleObject, cellManager *NS.CellManagerClient) OBJ.SingleObject {
	//playerToUpdate := singleObjectToPlayer(object)
	//TODO: check for valid update
	thisPlayer.PlayerMightLeaveCellHandle(object, cellManager)
	return object
}
