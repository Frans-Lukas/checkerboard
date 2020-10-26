package constants

const CellManagerPort = ":50051"
const CellManagerAddress = "localhost" + CellManagerPort
const DebugMode = false
const MAP_SIZE = 10
const SplitCellRequirement = 8
const MergeCellRequirement = 1
const SplitCellInterval = 3
const MergeAgeRequirement = 30
const ClientImage = "client.png"
const PlayerImage = "player.png"
const IconSize = int(500 / MAP_SIZE)
const AliveCheckInterval = 5
const DialTimeoutMilli = 100
const RemovedKey = "REMOVE_KEY"
