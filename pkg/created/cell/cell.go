package cell

type Cell struct {
	CellId  string
	Players []Player
}

func (cell *Cell) AppendPlayer(player Player) {
	cell.Players = append(cell.Players, player)
}

func (cell *Cell) DeletePlayer(playerToRemove Player) {
	for index, player := range cell.Players {
		if player.Ip == playerToRemove.Ip && player.Port == playerToRemove.Port {
			cell.Players[index] = cell.Players[len(cell.Players)-1]
			cell.Players = cell.Players[:len(cell.Players)-1]
		}
	}
}
