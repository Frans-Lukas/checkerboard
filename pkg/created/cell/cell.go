package cell

import "github.com/Frans-Lukas/checkerboard/pkg/created/cell/objects"

type Cell struct {
	CellId     string
	Players    []objects.Client
	CellMaster *objects.Client
	Locked     bool
	Lockee     string
}

func (cell *Cell) AppendPlayer(player objects.Client) {
	cell.Players = append(cell.Players, player)
}

func (cell *Cell) DeletePlayer(playerToRemove objects.Client) {
	for index, player := range cell.Players {
		if player.Ip == playerToRemove.Ip && player.Port == playerToRemove.Port {
			cell.Players[index] = cell.Players[len(cell.Players)-1]
			cell.Players = cell.Players[:len(cell.Players)-1]
		}
	}
}
