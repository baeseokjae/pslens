package pscan

import (
	"strconv"
	"strings"
)

// GetProcessDetail enriches a Process with full runtime details.
func GetProcessDetail(p Process) ProcessDetail {
	d := ProcessDetail{Process: p}
	d.Args = GetProcessArgs(p.PID)
	d.CWD = GetProcessCWD(p.PID)

	pidStr := strconv.Itoa(p.PID)

	lf := runSilent("lsof", "-p", pidStr)
	if lf != "" {
		lines := strings.Split(lf, "\n")
		count := 0
		for _, l := range lines {
			fs := strings.Fields(l)
			if len(fs) >= 2 {
				lpid, _ := strconv.Atoi(fs[1])
				if lpid == p.PID {
					count++
				}
			}
		}
		d.OpenFiles = count
	}

	listen := runSilent("lsof", "-p", pidStr, "-iTCP", "-sTCP:LISTEN", "-P", "-n")
	if listen != "" {
		lines := strings.Split(listen, "\n")
		for _, l := range lines {
			if strings.Contains(l, "(LISTEN)") {
				fs := strings.Fields(l)
				if len(fs) >= 9 {
					lpid, _ := strconv.Atoi(fs[1])
					if lpid == p.PID {
						d.ListeningPorts = append(d.ListeningPorts, fs[8])
					}
				}
			}
		}
	}

	conns := runSilent("lsof", "-p", pidStr, "-iTCP", "-sTCP:ESTABLISHED", "-P", "-n")
	if conns != "" {
		lines := strings.Split(conns, "\n")
		for _, l := range lines {
			if strings.Contains(l, "ESTABLISHED") {
				fs := strings.Fields(l)
				if len(fs) >= 9 {
					lpid, _ := strconv.Atoi(fs[1])
					if lpid == p.PID {
						d.Connections = append(d.Connections, fs[8])
					}
				}
			}
		}
	}
	return d
}