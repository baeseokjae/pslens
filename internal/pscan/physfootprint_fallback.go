//go:build !cgo || !darwin

package pscan

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	footprintMu    sync.Mutex
	footprintCache map[int]int
)

const footprintTimeout = 1 * time.Second
const maxConcurrent = 8

// GetFootprintKB returns the Physical Footprint (vmmap) for a PID in KB.
func GetFootprintKB(pid int) int {
	footprintMu.Lock()
	if footprintCache == nil {
		footprintCache = make(map[int]int)
	}
	if v, ok := footprintCache[pid]; ok {
		footprintMu.Unlock()
		return v
	}
	footprintMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), footprintTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "vmmap", "--summary", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		footprintMu.Lock()
		footprintCache[pid] = 0
		footprintMu.Unlock()
		return 0
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Physical footprint:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				kb := ParseMemToKB(parts[2])
				footprintMu.Lock()
				footprintCache[pid] = kb
				footprintMu.Unlock()
				return kb
			}
		}
	}

	footprintMu.Lock()
	footprintCache[pid] = 0
	footprintMu.Unlock()
	return 0
}

// EnrichFootprint fetches Physical Footprint for multiple processes in parallel.
// Skips processes with RSS < 50MB (their footprint ≈ RSS, vmmap overhead not worth it).
func EnrichFootprint(procs []Process) {
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	skipThreshold := 50 * 1024 // 50MB in KB

	for i := range procs {
		if procs[i].RSS_KB < skipThreshold {
			procs[i].FootprintKB = procs[i].RSS_KB
			continue
		}
		wg.Add(1)
		go func(p *Process) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			p.FootprintKB = GetFootprintKB(p.PID)
		}(&procs[i])
	}
	wg.Wait()
}