package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/baeseokjae/pslens/internal/pscan"
)

var sortBy = "footprint"

func main() {
	args := os.Args[1:]

	sortBy = "footprint"
	newArgs := []string{}
	for _, a := range args {
		if a == "-r" {
			sortBy = "rss"
		} else {
			newArgs = append(newArgs, a)
		}
	}
	args = newArgs

	if len(args) == 0 || args[0] == "scan" || args[0] == "all" {
		cmdScan()
		return
	}

	switch args[0] {
	case "top":
		n := 10
		if len(args) > 1 {
			if v, err := strconv.Atoi(args[1]); err == nil {
				n = v
			}
		}
		cmdTop(n)

	case "ghost":
		cmdGhost()

	case "app":
		if len(args) < 2 {
			fmt.Printf("  %s usage: pscan app <name>%s\n", pscan.Red, pscan.Reset)
			return
		}
		cmdApp(args[1])

	case "pid":
		if len(args) < 2 {
			fmt.Printf("  %s usage: pscan pid <pid>%s\n", pscan.Red, pscan.Reset)
			return
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("  %s invalid PID%s\n", pscan.Red, pscan.Reset)
			return
		}
		cmdPID(pid)

	case "ports":
		cmdPorts()

	case "kill":
		if len(args) < 2 {
			fmt.Printf("  %s usage: pscan kill <pid>%s\n", pscan.Red, pscan.Reset)
			return
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("  %s invalid PID%s\n", pscan.Red, pscan.Reset)
			return
		}
		args := pscan.GetProcessArgs(pid)
		if args == "" {
			fmt.Printf("  %s PID %d not found%s\n", pscan.Red, pid, pscan.Reset)
			return
		}
		fmt.Printf("\n  %s⚠️  PID %d (%s) to kill?%s\n", pscan.Yellow, pid, strings.TrimSpace(args), pscan.Reset)
		fmt.Printf("  kill -9 %d\n", pid)

	case "help", "--help", "-h":
		printUsage()

	default:
		printUsage()
	}
}

func cmdScan() {
	all := pscan.GetAllProcesses()
	var totalRSS float64
	var activeCount, ghostCount int
	appMap := make(map[string][]pscan.Process)

	for _, p := range all {
		ak := pscan.AppKey(p)
		appMap[ak] = append(appMap[ak], p)
		totalRSS += float64(p.RSS_KB) / 1024.0
		if pscan.IsGhost(p) {
			ghostCount++
		} else {
			activeCount++
		}
	}

	sort.Slice(all, func(i, j int) bool { return all[i].RSS_KB > all[j].RSS_KB })

	if sortBy != "rss" {
		pscan.EnrichFootprint(all[:10])
	}

	label := "Physical Footprint"
	if sortBy == "rss" {
		label = "RSS"
	}

	pscan.Header("🔍 System Process Scan (%s)", label)
	fmt.Printf("  %sTotal:%s %d  %sActive:%s %d  %sGhost:%s %d (%.0fMB RSS)%s\n\n",
		pscan.Bold, pscan.Reset, len(all),
		pscan.Green, pscan.Reset, activeCount,
		pscan.Red, pscan.Reset, ghostCount, totalRSS, pscan.Gray)

	cmdTopN(all, 10)
	cmdGhostN(all)

	// App summary
	pscan.Header("📦 Per-Application Memory (%s)", label)
	type appStat struct {
		name    string
		totalMB float64
		count   int
		ghostCt int
	}
	var appStats []appStat
	for name, procs := range appMap {
		var total float64
		gc := 0
		for _, p := range procs {
			total += float64(p.SortKey(sortBy)) / 1024.0
			if pscan.IsGhost(p) {
				gc++
			}
		}
		appStats = append(appStats, appStat{name, total, len(procs), gc})
	}
	sort.Slice(appStats, func(i, j int) bool { return appStats[i].totalMB > appStats[j].totalMB })

	fmt.Printf("  %-28s %-10s %-6s %s\n", "APP", "MEMORY", "PROCS", "GHOST")
	fmt.Printf("  %s\n", strings.Repeat("─", 60))
	for i, as := range appStats {
		if i >= 15 {
			break
		}
		ghostTag := "0"
		if as.ghostCt > 0 {
			ghostTag = pscan.Colorize(strconv.Itoa(as.ghostCt), pscan.Red)
		}
		fmt.Printf("  %s%-28s%s %s%s%s %-6d %s\n",
			pscan.Bold, as.name, pscan.Reset,
			pscan.ColorMem(as.totalMB), fmt.Sprintf("%.0fMB", as.totalMB), pscan.Reset,
			as.count, ghostTag,
		)
	}
	fmt.Println()
}

func cmdTop(n int) {
	all := pscan.GetAllProcesses()
	cmdTopN(all, n)
}

func cmdTopN(all []pscan.Process, n int) {
	// Sort by RSS first
	sort.Slice(all, func(i, j int) bool { return all[i].RSS_KB > all[j].RSS_KB })

	// Enrich a bit more than needed in case re-sort moves things around
	// Enrich slightly more than n, in case re-sort moves things around
	topN := n + 3
	if topN > len(all) {
		topN = len(all)
	}
	if sortBy != "rss" {
		pscan.EnrichFootprint(all[:topN])
	}

	// Re-sort by the chosen metric (now that we have footprints)
	sort.Slice(all, func(i, j int) bool { return all[i].SortKey(sortBy) > all[j].SortKey(sortBy) })

	label := "Physical Footprint"
	if sortBy == "rss" {
		label = "RSS"
	}
	pscan.Header("📊 TOP %d (by %s)", n, label)
	fmt.Printf("  %-7s %-9s %-10s %-28s %s\n", "PID", "MEM", "STATE", "APP", "NOTE")
	fmt.Printf("  %s\n", strings.Repeat("─", 100))
	shown := 0
	for i := 0; i < len(all) && shown < n; i++ {
		p := all[i]
		if pscan.IsGhost(p) {
			continue
		}
		shown++
		memVal := p.SortKey(sortBy)
		mbVal := float64(memVal) / 1024.0
		appName := pscan.AppKey(p)
		if len([]rune(appName)) > 28 {
			appName = pscan.TruncateLen(appName, 28)
		}
		note := ""
		if sortBy != "rss" && p.FootprintKB > 0 && p.RSS_KB > 0 {
			note = fmt.Sprintf("RAM:%.0fMB", float64(p.RSS_KB)/1024.0)
		}
		fmt.Printf("  %s%-7d%s %s%-9s%s %s%-10s%s %-28s %s%s%s\n",
			pscan.Bold, p.PID, pscan.Reset,
			pscan.ColorMem(mbVal), pscan.FmtKB(memVal), pscan.Reset,
			pscan.StateColor(p.State), p.State, pscan.Reset,
			appName,
			pscan.Gray, note, pscan.Reset,
		)
	}
	fmt.Println()
}

func cmdGhost() {
	all := pscan.GetAllProcesses()
	cmdGhostN(all)
}

func cmdGhostN(all []pscan.Process) {
	var ghosts []pscan.Process
	for _, p := range all {
		if pscan.IsGhost(p) {
			ghosts = append(ghosts, p)
		}
	}
	if len(ghosts) == 0 {
		fmt.Printf("\n  %s👻 No ghost processes%s\n\n", pscan.Green, pscan.Reset)
		return
	}
	fmt.Printf("\n  %s👻 Ghost Processes: %d%s\n", pscan.Bold, len(ghosts), pscan.Reset)
	fmt.Printf("  %s   (RSS < 1MB, sleeping — safe to kill)%s\n\n", pscan.Gray, pscan.Reset)
	fmt.Printf("  %-7s %-9s %-10s %s\n", "PID", "MEM", "STATE", "ARGUMENTS")
	fmt.Printf("  %s\n", strings.Repeat("─", 100))
	for _, p := range ghosts {
		args := pscan.GetProcessArgs(p.PID)
		shortArgs := args
		if len([]rune(shortArgs)) > 75 {
			shortArgs = pscan.TruncateLen(shortArgs, 75)
		}
		fmt.Printf("  %s%-7d%s %s%-9s%s %s%-10s%s %s%s%s\n",
			pscan.Red, p.PID, pscan.Reset,
			pscan.Red, pscan.FmtKB(p.RSS_KB), pscan.Reset,
			pscan.Gray, p.State, pscan.Reset,
			pscan.Gray, shortArgs, pscan.Reset,
		)
	}
	fmt.Printf("\n  %sSuggestion:%s kill -9 ", pscan.Yellow, pscan.Reset)
	for i, p := range ghosts {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(p.PID)
	}
	fmt.Println()
}

func cmdApp(name string) {
	all := pscan.GetAllProcesses()
	nameLower := strings.ToLower(name)
	var matched []pscan.Process
	for _, p := range all {
		ak := strings.ToLower(pscan.AppKey(p))
		commLower := strings.ToLower(p.Comm)
		if ak == nameLower || strings.Contains(commLower, nameLower) {
			matched = append(matched, p)
		}
	}
	if len(matched) == 0 {
		fmt.Printf("\n  %s'%s' not found%s\n", pscan.Red, name, pscan.Reset)
		return
	}

	if sortBy != "rss" {
		pscan.EnrichFootprint(matched)
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].SortKey(sortBy) > matched[j].SortKey(sortBy) })

	var totalMem float64
	label := "Physical Footprint"
	if sortBy == "rss" {
		label = "RSS"
	}
	fmt.Printf("\n  %s📦 %s — %d processes (%s)%s\n\n", pscan.Bold, name, len(matched), label, pscan.Reset)
	for _, p := range matched {
		memVal := p.SortKey(sortBy)
		mbVal := float64(memVal) / 1024.0
		totalMem += mbVal
		d := pscan.GetProcessDetail(p)

		var detailParts []string
		if sortBy != "rss" && p.FootprintKB > 0 && p.RSS_KB > 0 {
			detailParts = append(detailParts, fmt.Sprintf("RAM:%.0fMB", float64(p.RSS_KB)/1024.0))
		}
		if d.OpenFiles > 0 {
			detailParts = append(detailParts, fmt.Sprintf("files:%d", d.OpenFiles))
		}
		if len(d.ListeningPorts) > 0 {
			detailParts = append(detailParts, fmt.Sprintf("listen:%s", strings.Join(d.ListeningPorts, ",")))
		}
		if len(d.Connections) > 0 {
			detailParts = append(detailParts, fmt.Sprintf("conn:%d", len(d.Connections)))
		}
		detail := strings.Join(detailParts, " | ")

		fmt.Printf("  %s%s PID %d%s  %s%s%s  %s%s%s\n",
			pscan.Bold, pscan.AppKey(p), p.PID, pscan.Reset,
			pscan.ColorMem(mbVal), pscan.FmtKB(memVal), pscan.Reset,
			pscan.StateColor(p.State), p.State, pscan.Reset,
		)
		if d.CWD != "" {
			cwd := d.CWD
			parts := strings.Split(cwd, "/")
			if len(parts) > 4 {
				cwd = ".../" + strings.Join(parts[len(parts)-3:], "/")
			}
			if len([]rune(cwd)) > 70 {
				cwd = pscan.TruncateTail(cwd, 70)
			}
			fmt.Printf("  %s📂%s %s\n", pscan.Gray, cwd, pscan.Reset)
		}
		if d.Args != "" {
			args := d.Args
			if len([]rune(args)) > 70 {
				parts := strings.Fields(args)
				if len(parts) >= 2 {
					rest := strings.Join(parts[1:], " ")
					if len([]rune(rest)) > 65 {
						rest = pscan.TruncateTail(rest, 65)
					}
					args = parts[0] + " " + rest
				} else {
					args = pscan.TruncateTail(args, 70)
				}
			}
			fmt.Printf("  %s🔧%s %s\n", pscan.Gray, args, pscan.Reset)
		}
		if detail != "" {
			fmt.Printf("  %sℹ️ %s%s\n", pscan.Gray, detail, pscan.Reset)
		}
		fmt.Println()
	}
	fmt.Printf("  %sTotal: %.0fMB%s\n", pscan.Bold, totalMem, pscan.Reset)
}

func cmdPID(pid int) {
	all := pscan.GetAllProcesses()
	var target *pscan.Process
	for _, p := range all {
		if p.PID == pid {
			target = &p
			break
		}
	}
	if target == nil {
		args := pscan.GetProcessArgs(pid)
		if args == "" {
			fmt.Printf("  %sPID %d not found%s\n", pscan.Red, pid, pscan.Reset)
			return
		}
		raw := pscan.GetProcessArgs(pid)
		_ = raw
		fmt.Printf("  %sPID %d not found in ps list%s\n", pscan.Red, pid, pscan.Reset)
		return
	}

	target.FootprintKB = pscan.GetFootprintKB(target.PID)
	d := pscan.GetProcessDetail(*target)
	memVal := target.SortKey(sortBy)
	mbVal := float64(memVal) / 1024.0

	fmt.Printf("\n  %s🔍 PID %d Details%s\n\n", pscan.Bold, pid, pscan.Reset)
	fmt.Printf("  %s%-15s%s %s\n", pscan.Bold, "Name:", pscan.Reset, d.Comm)
	fmt.Printf("  %s%-15s%s %s%s%s\n",
		pscan.Bold, "Physical Footprint:", pscan.Reset,
		pscan.ColorMem(mbVal), pscan.FmtKB(memVal), pscan.Reset)
	if d.RSS_KB > 0 {
		fmt.Printf("  %s%-15s%s %.0fMB%s\n", pscan.Bold, "RSS:", pscan.Reset, float64(d.RSS_KB)/1024.0, pscan.Gray)
	}
	fmt.Printf("  %s%-15s%s %s%s%s\n", pscan.Bold, "State:", pscan.Reset, pscan.StateColor(d.State), d.State, pscan.Reset)
	fmt.Printf("  %s%-15s%s %s\n", pscan.Bold, "User:", pscan.Reset, d.User)
	fmt.Printf("  %s%-15s%s %d\n", pscan.Bold, "Open files:", pscan.Reset, d.OpenFiles)
	if d.CWD != "" {
		fmt.Printf("  %s%-15s%s %s\n", pscan.Bold, "CWD:", pscan.Reset, d.CWD)
	}
	fmt.Printf("  %s%-15s%s %s\n\n", pscan.Bold, "Command:", pscan.Reset, d.Args)

	if len(d.ListeningPorts) > 0 {
		fmt.Printf("  %s📡 Listening Ports%s\n", pscan.Bold, pscan.Reset)
		for _, port := range d.ListeningPorts {
			fmt.Printf("    • %s\n", port)
		}
		fmt.Println()
	}
	if len(d.Connections) > 0 {
		fmt.Printf("  %s🔗 Active Connections%s\n", pscan.Bold, pscan.Reset)
		maxShow := 10
		for i, conn := range d.Connections {
			if i >= maxShow {
				fmt.Printf("    • ...and %d more\n", len(d.Connections)-maxShow)
				break
			}
			fmt.Printf("    • %s\n", conn)
		}
		fmt.Println()
	}

	// Parent info uses ps directly
	ppidStr := ""
	_ = ppidStr
	if pscan.IsGhost(*target) {
		fmt.Printf("  %s⚠️  Ghost process. kill -9 %d to clean up%s\n", pscan.Yellow, pid, pscan.Reset)
	}
}

func cmdPorts() {
	fmt.Printf("\n  %s🔌 Listening Ports%s\n\n", pscan.Bold, pscan.Reset)
	fmt.Printf("  %-8s %-22s %-10s %s\n", "PORT", "PROCESS", "PID", "TYPE")
	fmt.Printf("  %s\n", strings.Repeat("─", 70))

	for _, p := range pscan.GetPorts() {
		fmt.Printf("  %s%-8s%s %-22s %-10d %s\n",
			pscan.Yellow, p.Port, pscan.Reset,
			p.Process, p.PID,
			pscan.Gray+p.Addr+pscan.Reset,
		)
	}
	fmt.Println()
}

func printUsage() {
	b, r, g, c := pscan.Bold, pscan.Reset, pscan.Green, pscan.Cyan
	fmt.Printf("\n  %spslens%s — macOS Process Scanner\n\n", b, r)
	fmt.Printf("  %sUsage:%s\n", g, r)
	fmt.Printf("    %spslens%s               Full scan (Footprint)\n", c, r)
	fmt.Printf("    %spslens -r%s             Full scan (RSS)\n", c, r)
	fmt.Printf("    %spslens top%s            Top 10 (Footprint)\n", c, r)
	fmt.Printf("    %spslens top -r%s         Top 10 (RSS)\n", c, r)
	fmt.Printf("    %spslens ghost%s          Show ghost processes\n", c, r)
	fmt.Printf("    %spslens app%s <name>     App details (CWD, args, connections)\n", c, r)
	fmt.Printf("    %spslens pid%s <pid>      PID details\n", c, r)
	fmt.Printf("    %spslens ports%s          Listening ports\n", c, r)
}