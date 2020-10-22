package objects

import (
	generated "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"
	"github.com/Frans-Lukas/checkerboard/pkg/generated/objects"
)

type Cell struct {
	CellId     string
	Players    []Client
	CellMaster *Client
	PosX       int64
	PosY       int64
	Width      int64
	Height     int64
	Locked     bool
	Lockee     string
}

func NewCell(cellID string) Cell {
	return Cell{
		CellId:  cellID,
		Players: make([]Client, 0),
	}
}

func NewCellFromCells(cellID string, cell1 Cell, cell2 Cell) Cell {
	newCell := NewCell(cellID)

	if cell1.PosX <= cell2.PosX {
		newCell.PosX = cell1.PosX
	} else {
		newCell.PosX = cell2.PosX
	}

	if cell1.PosY <= cell2.PosY {
		newCell.PosY = cell1.PosY
	} else {
		newCell.PosY = cell2.PosY
	}

	if cell1.PosX+cell1.Width >= cell2.PosX+cell2.Width {
		newCell.Width = cell1.PosX + cell1.Width - newCell.PosX
	} else {
		newCell.Width = cell2.PosX + cell2.Width - newCell.PosX
	}

	if cell1.PosY+cell1.Height >= cell2.PosY+cell2.Height {
		newCell.Height = cell1.PosY + cell1.Height - newCell.PosY
	} else {
		newCell.Height = cell2.PosY + cell2.Height - newCell.PosY
	}

	return newCell
}

func (cell *Cell) ToGeneratedCell() objects.Cell {
	return objects.Cell{CellId: cell.CellId, PosX: cell.PosX, PosY: cell.PosY, Width: cell.Width, Height: cell.Height}
}

func (cell *Cell) AppendPlayer(player Client) {
	cell.Players = append(cell.Players, player)
}
func (cell *Cell) ContainsPlayer(player Client) bool {
	for _, p := range cell.Players {
		if ToAddress(p.Ip, p.Port) == ToAddress(player.Ip, int32(player.Port)) {
			return true
		}
	}
	return false
}

func (cell *Cell) CollidesWith(in *generated.Position) bool {
	return cell.PosX <= in.PosX && cell.PosX+cell.Width > in.PosX &&
		cell.PosY <= in.PosY && cell.PosY+cell.Height > in.PosY
}

func (cell *Cell) SelectNewCellMaster() int {
	bestTrustLevel := uint32(0)
	cmIndex := -1
	for index, player := range cell.Players {
		if player.TrustLevel >= bestTrustLevel {
			bestTrustLevel = player.TrustLevel
			cmIndex = index
		}
	}
	return cmIndex
}

func (cell *Cell) DeletePlayer(playerToRemove Client) {
	for index, player := range cell.Players {
		if player.Ip == playerToRemove.Ip && player.Port == playerToRemove.Port {
			cell.Players[index] = cell.Players[len(cell.Players)-1]
			cell.Players = cell.Players[:len(cell.Players)-1]
		}
	}
}
