package cell

import "github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"
import generated "github.com/Frans-Lukas/checkerboard/pkg/generated/cellmanager"

type Cell struct {
	CellId     string
	Players    []objects.Client
	CellMaster *objects.Client
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
		Players: make([]objects.Client, 0),
	}
}

func (cell *Cell) AppendPlayer(player objects.Client) {
	cell.Players = append(cell.Players, player)
}
func (cell *Cell) CollidesWith(in *generated.Position) bool {
	return cell.PosX <= in.PosX && cell.PosX+cell.Width >= in.PosX &&
		cell.PosY <= in.PosY && cell.PosY+cell.Height >= in.PosY
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

func (cell *Cell) DeletePlayer(playerToRemove objects.Client) {
	for index, player := range cell.Players {
		if player.Ip == playerToRemove.Ip && player.Port == playerToRemove.Port {
			cell.Players[index] = cell.Players[len(cell.Players)-1]
			cell.Players = cell.Players[:len(cell.Players)-1]
		}
	}
}
