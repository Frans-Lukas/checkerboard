package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cellmanager"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	"testing"
)

func TestCreateCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	request := generated.CellRequest{CellId: "testId"}
	_, err := cm.CreateCell(context.Background(), &request)
	failIfNotNull(err, "could not create cell")
	if (*cm.Cells)[0].CellId != "testId" {
		fatalFail(errors.New("cell was not inserted with CreateCell"))
	}
}

func TestDeleteCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	request := generated.CellRequest{CellId: "testId"}
	_, err := cm.CreateCell(context.Background(), &request)
	failIfNotNull(err, "could not create cell")
	if (*cm.Cells)[0].CellId != "testId" {
		fatalFail(errors.New("cell was not inserted with CreateCell"))
	}

	_, err = cm.DeleteCell(context.Background(), &request)
	failIfNotNull(err, "could not delete cell")
	if len(*cm.Cells) != 0 {
		fatalFail(errors.New("cell was not deleted with DeleteCell"))
	}
}

func TestUnregisterCellMaster(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	(*cm.Cells)[0].CellMaster = &objects.Client{Ip: "testIp", Port: 1337, TrustLevel: 0}
	status, err := cm.UnregisterCellMaster(
		context.Background(), &generated.CellMasterRequest{CellId: "testId1"},
	)
	failIfNotNull(err, "could not unregister cell")
	if status.WasUnregistered == true && (*cm.Cells)[0].CellMaster == nil {
		return
	} else {
		fatalFail(errors.New("CellMaster was not unregistered with UnregisterCellMaster"))
	}
}

func TestUnregisterCellMasterReturnsOnFail(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	(*cm.Cells)[0].CellMaster = &objects.Client{Ip: "testIp", Port: 1337, TrustLevel: 0}
	status, err := cm.UnregisterCellMaster(
		context.Background(), &generated.CellMasterRequest{CellId: "invalidId"},
	)
	failIfNotNull(err, "could not unregister cell")
	if status.WasUnregistered == false {
		return
	}
	fatalFail(errors.New("unregister succeeded when it should not have"))
}

func TestCreateCellCreatesEmptyPlayerList(t *testing.T) {
	cm := cellmanager.NewCellManager()
	request := generated.CellRequest{CellId: "testId"}
	_, err := cm.CreateCell(context.Background(), &request)
	failIfNotNull(err, "could not create cell")
	if (*cm.Cells)[0].Players == nil {
		fatalFail(errors.New("cell did not create empty player list with CreateCell"))
	}
}

func TestListCells(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	cm.AppendCell(objects.Cell{CellId: "testId2"})
	cellList, err := cm.ListCells(context.Background(), &generated.ListCellsRequest{})
	failIfNotNull(err, "could not list cells")
	testId1Exists := false
	testId2Exists := false
	for _, cellId := range cellList.CellId {
		if cellId == "testId1" {
			testId1Exists = true
		}
		if cellId == "testId2" {
			testId2Exists = true
		}
	}
	if !testId1Exists || !testId2Exists {
		fatalFail(errors.New("cells were not returned from ListCells"))
	}
}

func TestListPlayersInCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	testIp := "192.168.16.1"
	(*cm.Cells)[0].AppendPlayer(objects.Client{Ip: testIp, Port: 1337})
	playerList, err := cm.ListPlayersInCell(
		context.Background(), &generated.ListPlayersRequest{CellId: "testId1"},
	)
	failIfNotNull(err, "could not list players object cell")
	if len(playerList.Port) == 0 || len(playerList.Ip) == 0 {
		fatalFail(errors.New("players were not returned from ListPlayersInCell"))
	}
	if playerList.Ip[0] == testIp && playerList.Port[0] == 1337 {
		return
	}
	fatalFail(errors.New("incorrect players were returned from ListPlayersInCell"))
}

func TestAddPlayerToCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Players: make([]objects.Client, 0)})
	testIp := "192.168.16.1"
	status, err := cm.AddPlayerToCell(
		context.Background(),
		&generated.PlayerInCellRequest{CellId: "testId1", Ip: testIp, Port: 1337},
	)
	failIfNotNull(err, "could not add player to cell")
	addedPlayer := (*cm.Cells)[0].Players[0]
	if status.Succeeded && addedPlayer.Port == 1337 && addedPlayer.Ip == testIp {
		return
	}
	fatalFail(errors.New("player was not added correctly"))
}

func TestAddPlayerToCellWithPositionsBoundary(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Players: make([]objects.Client, 0), PosY: 0, PosX: 0, Width: 100, Height: 100})
	testIp := "192.168.16.1"
	status, err := cm.AddPlayerToCellWithPositions(
		context.Background(),
		&generated.PlayerInCellRequestWithPositions{PosX: 100, PosY: 0, Ip: testIp, Port: 1337},
	)
	failIfNotNull(err, "could not add player to cell")
	addedPlayer := (*cm.Cells)[0].Players[0]
	if status.Succeeded && addedPlayer.Port == 1337 && addedPlayer.Ip == testIp {
		return
	}
	fatalFail(errors.New("player was not added correctly"))
}

func TestAddPlayerToCellWithPositionsCenter(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Players: make([]objects.Client, 0), PosY: 0, PosX: 0, Width: 100, Height: 100})
	testIp := "192.168.16.1"
	status, err := cm.AddPlayerToCellWithPositions(
		context.Background(),
		&generated.PlayerInCellRequestWithPositions{PosX: 50, PosY: 50, Ip: testIp, Port: 1337},
	)
	failIfNotNull(err, "could not add player to cell")
	addedPlayer := (*cm.Cells)[0].Players[0]
	if status.Succeeded && addedPlayer.Port == 1337 && addedPlayer.Ip == testIp {
		return
	}
	fatalFail(errors.New("player was not added correctly"))
}

func TestAddPlayerToCellWithPositionsShouldFailOnInvalidPositions(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Players: make([]objects.Client, 0), PosY: 0, PosX: 0, Width: 100, Height: 100})
	testIp := "192.168.16.1"
	status, err := cm.AddPlayerToCellWithPositions(
		context.Background(),
		&generated.PlayerInCellRequestWithPositions{PosX: -50, PosY: 50, Ip: testIp, Port: 1337},
	)
	if !status.Succeeded && err != nil {
		return
	}
	fatalFail(errors.New("player was added incorrectly"))
}

func TestAddPlayerToCellThrowsIfInvalidCellId(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Players: make([]objects.Client, 0)})
	testIp := "192.168.16.1"
	status, err := cm.AddPlayerToCell(
		context.Background(),
		&generated.PlayerInCellRequest{CellId: "invalidTestId", Ip: testIp, Port: 1337},
	)
	if status.Succeeded == false && err != nil {
		return
	} else {
		fatalFail(errors.New("AddPlayerToCell did not throw on invalid cellId"))
	}
}

func TestSetWorldSize(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.SetWorldSize(context.Background(), &generated.WorldSize{Width: 100, Height: 100})
	if cm.WorldHeight == 100 && cm.WorldWidth == 100 {
		return
	} else {
		fatalFail(errors.New("SetWorldSize does not set world size"))
	}
}

func TestSetWorldSizeCreatesBigCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.SetWorldSize(context.Background(), &generated.WorldSize{Width: 100, Height: 100})
	if (*cm.Cells)[0].Width == 100 && (*cm.Cells)[0].Height == 100 &&
		(*cm.Cells)[0].PosX == 0 && (*cm.Cells)[0].PosY == 0 {
		return
	} else {
		fatalFail(errors.New("SetWorldSize does not initialize world cell"))
	}
}

func TestSetWorldSizeFailsIfACellExists(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.SetWorldSize(context.Background(), &generated.WorldSize{Width: 100, Height: 100})
	status, _ := cm.SetWorldSize(context.Background(), &generated.WorldSize{Width: 100, Height: 100})
	if status.Succeeded == false {
		return
	} else {
		fatalFail(errors.New("SetWorldSize should fail if cells exists"))
	}
}

func TestPlayerLeftCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId2"})
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	testIp := "192.168.16.1"
	testIp2 := "192.168.16.2"
	(*cm.Cells)[1].AppendPlayer(objects.Client{Ip: testIp, Port: 1337})
	(*cm.Cells)[1].AppendPlayer(objects.Client{Ip: testIp2, Port: 1337})
	reply, err := cm.PlayerLeftCell(
		context.Background(),
		&generated.PlayerInCellRequest{Port: 1337, Ip: testIp, CellId: "testId1"},
	)
	failIfNotNull(err, "could not list players object cell")
	if !reply.PlayerLeft {
		fatalFail(errors.New("PlayerLeft bool is invalid"))
	}
	player2Exists := false
	for _, player := range (*cm.Cells)[0].Players {
		if player.Ip == testIp2 {
			player2Exists = false
		}
	}
	if player2Exists {
		fatalFail(errors.New("player was not removed from cell object playerleftcell"))
	}
}

func TestLockCells(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	cm.AppendCell(objects.Cell{CellId: "testId2"})
	cm.AppendCell(objects.Cell{CellId: "testId3"})

	ids := []string{"testId1", "testId2"}
	request := generated.LockCellsRequest{CellId: ids, SenderCellId: "tester"}
	reply, err := cm.LockCells(context.Background(), &request)
	failIfNotNull(err, "could not lock cells")
	if !reply.Locked {
		fatalFail(errors.New("locked bool is invalid"))
	}

	if !(*cm.Cells)[0].Locked {
		fatalFail(errors.New("cell testId1 is not locked"))
	} else if !(*cm.Cells)[1].Locked {
		fatalFail(errors.New("cell testId2 is not locked"))
	} else if (*cm.Cells)[2].Locked {
		fatalFail(errors.New("cell testId3 is locked"))
	}

	if (*cm.Cells)[0].Lockee != "tester" {
		fatalFail(errors.New("cell testId1 has wrong lockee"))
	} else if (*cm.Cells)[1].Lockee != "tester" {
		fatalFail(errors.New("cell testId2 has wrong lockee"))
	}
}

func TestCannotLockWhenACellIsLocked(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	cm.AppendCell(objects.Cell{CellId: "testId2", Locked: true})
	cm.AppendCell(objects.Cell{CellId: "testId3"})

	ids := []string{"testId1", "testId2"}
	request := generated.LockCellsRequest{CellId: ids, SenderCellId: "tester"}
	reply, err := cm.LockCells(context.Background(), &request)
	failIfNotNull(err, "could not lock cells")
	if reply.Locked {
		fatalFail(errors.New("locked bool is invalid"))
	}

	if (*cm.Cells)[0].Locked {
		fatalFail(errors.New("cell testId1 is locked"))
	} else if !(*cm.Cells)[1].Locked {
		fatalFail(errors.New("cell testId2 is not locked"))
	} else if (*cm.Cells)[2].Locked {
		fatalFail(errors.New("cell testId3 is locked"))
	}
}

func TestUnlockCells(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Locked: true})
	cm.AppendCell(objects.Cell{CellId: "testId2", Locked: true})
	cm.AppendCell(objects.Cell{CellId: "testId3"})

	ids := []string{"testId1", "testId2"}
	request := generated.LockCellsRequest{CellId: ids}
	reply, err := cm.UnlockCells(context.Background(), &request)
	failIfNotNull(err, "could not lock cells")
	if reply.Locked {
		fatalFail(errors.New("locked bool is invalid"))
	}

	if (*cm.Cells)[0].Locked {
		fatalFail(errors.New("cell testId1 is not unlocked"))
	} else if (*cm.Cells)[1].Locked {
		fatalFail(errors.New("cell testId2 is not unlocked"))
	} else if (*cm.Cells)[2].Locked {
		fatalFail(errors.New("cell testId3 is locked"))
	}
}

func TestCannotUnlockWhenACellIsNotLocked(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Locked: true})
	cm.AppendCell(objects.Cell{CellId: "testId2"})
	cm.AppendCell(objects.Cell{CellId: "testId3"})

	ids := []string{"testId1", "testId2"}
	request := generated.LockCellsRequest{CellId: ids}
	reply, err := cm.UnlockCells(context.Background(), &request)
	failIfNotNull(err, "could not lock cells")
	if !reply.Locked {
		fatalFail(errors.New("locked bool is invalid"))
	}

	if !(*cm.Cells)[0].Locked {
		fatalFail(errors.New("cell testId1 is unlocked"))
	} else if (*cm.Cells)[1].Locked {
		fatalFail(errors.New("cell testId2 is locked"))
	} else if (*cm.Cells)[2].Locked {
		fatalFail(errors.New("cell testId3 is locked"))
	}
}

func TestCannotUnlockWhenACellIsLockedBySomeoneElse(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1", Locked: true, Lockee: "tester"})
	cm.AppendCell(objects.Cell{CellId: "testId2", Locked: true, Lockee: "hacker"})
	cm.AppendCell(objects.Cell{CellId: "testId3"})

	ids := []string{"testId1", "testId2"}
	request := generated.LockCellsRequest{CellId: ids}
	reply, err := cm.UnlockCells(context.Background(), &request)
	failIfNotNull(err, "could not lock cells")
	if !reply.Locked {
		fatalFail(errors.New("locked bool is invalid"))
	}

	if !(*cm.Cells)[0].Locked {
		fatalFail(errors.New("cell testId1 is unlocked"))
	} else if !(*cm.Cells)[1].Locked {
		fatalFail(errors.New("cell testId2 is unlocked"))
	} else if (*cm.Cells)[2].Locked {
		fatalFail(errors.New("cell testId3 is locked"))
	}
}

func TestRequestCellMaster(t *testing.T) {
	cellMaster := objects.Client{Ip: "randomIp", Port: 1337}
	mainCell := objects.Cell{CellId: "testId2", CellMaster: &cellMaster}

	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.Cell{CellId: "testId1"})
	cm.AppendCell(mainCell)

	request := generated.CellMasterRequest{CellId: "testId2"}
	reply, err := cm.RequestCellMaster(context.Background(), &request)
	failIfNotNull(err, "could not lock cells")
	if reply.Ip == "" {
		fatalFail(errors.New("returned empty cellMaster"))
	}
	if reply.Ip != "randomIp" {
		fatalFail(errors.New("returned wrong Ip"))
	}
	if reply.Port != 1337 {
		fatalFail(errors.New("returned wrong Port"))
	}
}

func TestRequestCellMasterSelectsNewCellMaster(t *testing.T) {
	cellMaster := objects.Client{Ip: "randomIp", Port: 1337}
	mainCell := objects.NewCell("testId2")

	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.NewCell("testId1"))
	cm.AppendCell(mainCell)
	(*cm.Cells)[1].Players = append((*cm.Cells)[1].Players, cellMaster)

	request := generated.CellMasterRequest{CellId: "testId2"}
	reply, err := cm.RequestCellMaster(context.Background(), &request)
	failIfNotNull(err, "Error")
	if reply.Ip == "" {
		fatalFail(errors.New("returned empty cellMaster"))
	}
	if reply.Ip != "randomIp" {
		fatalFail(errors.New("returned wrong Ip"))
	}
	if reply.Port != 1337 {
		fatalFail(errors.New("returned wrong Port"))
	}
}

func TestRequestCellMasterFailsOnEmptyCell(t *testing.T) {
	cellMaster := objects.Client{Ip: "randomIp", Port: 1337}
	mainCell := objects.NewCell("testId2")

	cm := cellmanager.NewCellManager()
	cm.AppendCell(objects.NewCell("testId1"))
	cm.AppendCell(mainCell)
	(*cm.Cells)[0].Players = append((*cm.Cells)[0].Players, cellMaster)

	request := generated.CellMasterRequest{CellId: "testId2"}
	_, err := cm.RequestCellMaster(context.Background(), &request)
	if err != nil {
		return
	}
	fatalFail(errors.New("should have failed"))
}

func TestRequestCellMasterWithPositions(t *testing.T) {
	cellMaster := objects.Client{Ip: "randomIp", Port: 1337}
	mainCell := objects.NewCell("testId2")
	mainCell.PosX = 0
	mainCell.PosY = 0
	mainCell.Width = 100
	mainCell.Height = 100

	cm := cellmanager.NewCellManager()
	cm.AppendCell(mainCell)

	(*cm.Cells)[0].CellMaster = &cellMaster

	request := generated.Position{PosX: 50, PosY: 50}
	newCm, err := cm.RequestCellMasterWithPositions(context.Background(), &request)
	if err != nil {
		fatalFail(errors.New("error on requesting CM"))
	}

	if newCm.Port == cellMaster.Port {
		return
	}

	fatalFail(errors.New("wrong cell master returned"))
}

func TestRequestCellMasterWithPositionsFailsIfOutOfBounds(t *testing.T) {
	cellMaster := objects.Client{Ip: "randomIp", Port: 1337}
	mainCell := objects.NewCell("testId2")
	mainCell.PosX = 0
	mainCell.PosY = 0
	mainCell.Width = 100
	mainCell.Height = 100

	cm := cellmanager.NewCellManager()
	cm.AppendCell(mainCell)

	(*cm.Cells)[0].CellMaster = &cellMaster

	request := generated.Position{PosX: 150, PosY: 50}
	_, err := cm.RequestCellMasterWithPositions(context.Background(), &request)
	if err != nil {
		return
	}
	fatalFail(errors.New("should have failed"))
}
