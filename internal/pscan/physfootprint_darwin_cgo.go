//go:build cgo && darwin

package pscan

/*
#include <libproc.h>
#include <stdlib.h>
#include <string.h>

int get_phys_footprint(int pid, uint64_t *footprint) {
    struct rusage_info_v6 ri;
    memset(&ri, 0, sizeof(ri));
    int ret = proc_pid_rusage(pid, RUSAGE_INFO_CURRENT, (rusage_info_t *)&ri);
    if (ret == 0) {
        *footprint = ri.ri_phys_footprint;
        return 0;
    }
    return ret;
}
*/
import "C"
import (
	"sync"
)

var (
	footprintMu    sync.Mutex
	footprintCache map[int]int
)

const maxConcurrent = 8

// GetFootprintKB returns the Physical Footprint for a PID in KB,
// using the macOS libproc syscall (microsecond-fast).
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

	var fp C.uint64_t
	ret := C.get_phys_footprint(C.int(pid), &fp)
	if ret != 0 {
		footprintMu.Lock()
		footprintCache[pid] = 0
		footprintMu.Unlock()
		return 0
	}

	kb := int(fp) / 1024
	footprintMu.Lock()
	footprintCache[pid] = kb
	footprintMu.Unlock()
	return kb
}

// EnrichFootprint fetches Physical Footprint for multiple processes in parallel.
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