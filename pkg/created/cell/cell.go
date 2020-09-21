package cell

type Cell struct {
	CellId  string
	Players []Player
}

func (cell *Cell) AppendPlayer(player Player) {
	cell.Players = append(cell.Players, player)
}
