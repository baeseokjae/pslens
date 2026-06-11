package pscan

// Process represents a single OS process with memory metrics.
type Process struct {
	PID         int
	State       string
	Stat        string
	RSS_KB      int     // from ps (resident set size — current RAM)
	FootprintKB int     // from vmmap (physical footprint — actual total usage)
	MemPct      float64
	User        string
	Comm        string
}

// ProcessDetail enriches a Process with additional runtime information.
type ProcessDetail struct {
	Process
	Args           string
	CWD            string
	OpenFiles      int
	ListeningPorts []string
	Connections    []string
}

// PortInfo represents a listening TCP port and its owning process.
type PortInfo struct {
	Port    string
	PID     int
	Process string
	Addr    string
}