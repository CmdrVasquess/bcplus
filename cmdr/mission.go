package cmdr

type Mission struct {
	Faction    string
	Title      string
	Reputation float32
	Dests      []int64
}
