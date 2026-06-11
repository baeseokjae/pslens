package pscan

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
)

// GetAllProcesses parses `ps` output and returns all processes.
func GetAllProcesses() []Process {
	raw := runSilent("ps", "-eo", "pid,state,stat,rss,%mem,user,comm", "--no-headers")
	if raw == "" {
		raw = runSilent("ps", "-eo", "pid,state,stat,rss,%mem,user,comm")
	}
	var procs []Process
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "PID") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		pid, _ := strconv.Atoi(fields[0])
		rss, _ := strconv.Atoi(fields[3])
		memPct, _ := strconv.ParseFloat(fields[4], 64)
		procs = append(procs, Process{
			PID:    pid,
			State:  fields[1],
			Stat:   fields[2],
			RSS_KB: rss,
			MemPct: memPct,
			User:   fields[5],
			Comm:   fields[6],
		})
	}
	return procs
}

// GetProcessArgs returns the full command line of a process.
func GetProcessArgs(pid int) string {
	return runSilent("ps", "-p", strconv.Itoa(pid), "-o", "args=")
}

// GetProcessCWD returns the current working directory of a process.
func GetProcessCWD(pid int) string {
	pidStr := strconv.Itoa(pid)
	out := runSilent("lsof", "-p", pidStr, "-a", "-d", "cwd")
	if out == "" {
		return ""
	}
	for _, line := range strings.Split(out, "\n") {
		fs := strings.Fields(line)
		if len(fs) >= 2 {
			lpid, _ := strconv.Atoi(fs[1])
			if lpid == pid && len(fs) >= 9 {
				return fs[8]
			}
		}
	}
	return ""
}

// GetPorts returns all listening TCP ports and their owning processes.
func GetPorts() []PortInfo {
	out := runSilent("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n")
	if out == "" {
		return nil
	}

	var ports []PortInfo
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "COMMAND") || !strings.Contains(line, "(LISTEN)") {
			continue
		}
		fs := strings.Fields(line)
		if len(fs) < 9 {
			continue
		}
		pid, _ := strconv.Atoi(fs[1])
		addr := fs[8]
		ports = append(ports, PortInfo{
			Process: fs[0],
			PID:     pid,
			Addr:    addr,
			Port:    addr[strings.LastIndex(addr, ":")+1:],
		})
	}
	return ports
}
func IsGhost(p Process) bool {
	return p.RSS_KB < 1024 && (p.State == "S" || p.State == "I")
}

// runSilent runs a command and returns trimmed stdout, ignoring errors.
func runSilent(cmd string, args ...string) string {
	out, _ := exec.Command(cmd, args...).Output()
	return strings.TrimSpace(string(out))
}