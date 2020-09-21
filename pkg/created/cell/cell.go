package cell

type Cell struct {
	CellId  string
	Players []Player
	Locked bool
}

func (cell *Cell) AppendPlayer(player Player) {
	cell.Players = append(cell.Players, player)
}
