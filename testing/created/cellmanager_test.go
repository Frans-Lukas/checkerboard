package created

import (
	"context"
	"errors"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cell"
	"github.com/Frans-Lukas/checkerboard/pkg/created/cellmanager"
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated"
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
	cm.AppendCell(cell.Cell{CellId: "testId1"})
	cm.AppendCell(cell.Cell{CellId: "testId2"})
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
	cm.AppendCell(cell.Cell{CellId: "testId1"})
	testIp := "192.168.16.1"
	(*cm.Cells)[0].AppendPlayer(cell.Player{Ip: testIp, Port: 1337})
	playerList, err := cm.ListPlayersInCell(
		context.Background(), &generated.ListPlayersRequest{CellId: "testId1"},
	)
	failIfNotNull(err, "could not list players in cell")
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
	cm.AppendCell(cell.Cell{CellId: "testId1"})
	testIp := "192.168.16.1"
	(*cm.Cells)[0].AppendPlayer(cell.Player{Ip: testIp, Port: 1337})


	playerList, err := cm.ListPlayersInCell(
		context.Background(), &generated.ListPlayersRequest{CellId: "testId1"},
	)
	failIfNotNull(err, "could not list players in cell")
	if len(playerList.Port) == 0 || len(playerList.Ip) == 0 {
		fatalFail(errors.New("players were not returned from ListPlayersInCell"))
	}
	if playerList.Ip[0] == testIp && playerList.Port[0] == 1337 {
		return
	}
	fatalFail(errors.New("incorrect players were returned from ListPlayersInCell"))
}

func TestPlayerLeftCell(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(cell.Cell{CellId: "testId2"})
	cm.AppendCell(cell.Cell{CellId: "testId1"})
	testIp := "192.168.16.1"
	testIp2 := "192.168.16.2"
	(*cm.Cells)[1].AppendPlayer(cell.Player{Ip: testIp, Port: 1337})
	(*cm.Cells)[1].AppendPlayer(cell.Player{Ip: testIp2, Port: 1337})
	reply, err := cm.PlayerLeftCell(
		context.Background(),
		&generated.PlayerInCellRequest{Port: 1337, Ip: testIp, CellId: "testId1"},
	)
	failIfNotNull(err, "could not list players in cell")
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
		fatalFail(errors.New("player was not removed from cell in playerleftcell"))
	}
}

func TestLockCells(t *testing.T) {
	cm := cellmanager.NewCellManager()
	cm.AppendCell(cell.Cell{CellId: "testId1"})
	cm.AppendCell(cell.Cell{CellId: "testId2"})
	cm.AppendCell(cell.Cell{CellId: "testId3"})

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
	cm.AppendCell(cell.Cell{CellId: "testId1"})
	cm.AppendCell(cell.Cell{CellId: "testId2", Locked: true})
	cm.AppendCell(cell.Cell{CellId: "testId3"})

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
	cm.AppendCell(cell.Cell{CellId: "testId1", Locked: true})
	cm.AppendCell(cell.Cell{CellId: "testId2", Locked: true})
	cm.AppendCell(cell.Cell{CellId: "testId3"})

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
	cm.AppendCell(cell.Cell{CellId: "testId1", Locked: true})
	cm.AppendCell(cell.Cell{CellId: "testId2"})
	cm.AppendCell(cell.Cell{CellId: "testId3"})

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
	cm.AppendCell(cell.Cell{CellId: "testId1", Locked: true, Lockee: "tester"})
	cm.AppendCell(cell.Cell{CellId: "testId2", Locked: true, Lockee: "hacker"})
	cm.AppendCell(cell.Cell{CellId: "testId3"})

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
