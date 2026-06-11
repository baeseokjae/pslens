package pscan

// MemKB returns the "best" memory value: footprint if available, else RSS.
func (p Process) MemKB() int {
	if p.FootprintKB > 0 {
		return p.FootprintKB
	}
	return p.RSS_KB
}

// SortKey returns the value to sort by (footprint or RSS depending on mode).
func (p Process) SortKey(sortBy string) int {
	if sortBy == "rss" {
		return p.RSS_KB
	}
	return p.MemKB()
}