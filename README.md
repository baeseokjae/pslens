# pslens

**macOS Activity Monitor for your terminal.**

`pslens` shows you **Physical Footprint** — the same memory metric that Activity Monitor uses — not just RSS like `ps`, `top`, or `htop`. It groups processes by application, detects "ghost" processes (sleeping with <1MB RSS), and gives you deep per-process context (CWD, open files, listening ports, active connections).

> Why does Activity Monitor show Firefox using 1.2GB while `ps` says 400MB? Activity Monitor reports **Physical Footprint** (including compressed and swapped memory). `ps` only reports RSS (current RAM). `pslens` bridges this gap.

## Install

### Homebrew

```bash
brew install baeseokjae/tap/pslens
```

### Direct download

```bash
curl -fsSL https://github.com/baeseokjae/pslens/releases/latest/download/pslens_darwin_arm64.tar.gz \
  | tar xz -C /usr/local/bin pslens
```

### From source

```bash
go install github.com/baeseokjae/pslens@latest
```

## Quick start

```bash
pslens
```

Shows a system overview: total processes, ghost count, top 10 by Physical Footprint, and per-application memory usage.

## Commands

### `pslens` — Full system scan

Default command. Shows app-level grouping, top processes, and ghost detection in one view.

```bash
pslens           # Physical Footprint (Activity Monitor metric)
pslens -r        # RSS (current RAM usage — faster, no vmmap calls)
```

### `pslens top [N]` — Top processes

Lists the N most memory-hungry processes by Physical Footprint. Excludes ghosts.

```bash
pslens top       # Top 10 by Footprint
pslens top 20    # Top 20 by Footprint
pslens top -r    # Top 10 by RSS
```

Each row shows:
- **PID** and **MEM** (with color: red >500MB, yellow >200MB, cyan >50MB, green <50MB)
- **STATE** (green = sleeping, gray = idle, red = zombie)
- **NOTE** — actual RSS in MB when Footprint differs significantly

### `pslens ghost` — Find ghost processes

Ghost processes have RSS <1MB and are in Sleep state. They hold no real memory but clutter your process list. Safe to kill.

```bash
pslens ghost
```

Output includes a ready-to-paste `kill -9 <pid>` command.

### `pslens app <name>` — Application deep dive

Groups all processes belonging to an application, showing CWD, truncated arguments, open files, and network status for each.

```bash
pslens app claude    # All Claude processes
pslens app firefox   # All Firefox processes
pslens app codex     # All Codex processes
```

### `pslens pid <pid>` — Process deep dive

Detailed information about a single process: footprint, parent PID, children, open files, listening ports, and active connections.

```bash
pslens pid 52806
```

### `pslens ports` — Listening ports

All open TCP ports sorted by port number, with process name and PID.

```bash
pslens ports
```

### `pslens kill <pid>` — Confirm before killing

Shows the process name and the exact `kill` command, so you don't accidentally nuke the wrong process.

```bash
pslens kill 1234
```

## How it works

### Physical Footprint vs RSS

| Metric | Tool | What it measures |
|--------|------|------------------|
| **RSS** (Resident Set Size) | `ps`, `top`, `htop` | Pages currently in RAM |
| **Physical Footprint** | Activity Monitor, pslens | RSS + compressed memory + swapped memory |

macOS aggressively compresses inactive memory pages. A process may have 200MB compressed and only 50MB in RAM. Activity Monitor counts both; `ps` doesn't. `pslens` uses Apple's `vmmap --summary` to get the true Physical Footprint.

When you see something like:

```
PID    MEM        STATE  APP               NOTE
47883  2048.0MB   Ss     Virtualization     RAM:245MB
```

The process uses 2GB total (Footprint) but only 245MB is currently in RAM. The rest is compressed or swapped.

### Ghost processes

Processes can remain in the system after their work is done — sleeping with minimal RSS. `pslens` flags these as ghosts (RSS <1MB, Sleep/Idle state). They're safe to terminate. Common sources: completed background tasks, orphaned shell jobs, macOS launchd leftovers.

### Performance

- **RSS mode** (`-r`): Instant — reads `/proc` equivalent via `ps`
- **Footprint mode** (default): Uses `vmmap --summary` in parallel (up to 8 concurrent calls, 3-second timeout each)
- **Cache**: Footprint values are cached per run for fast `scan` and `top` commands

## FAQ

### How is pslens different from `btop` / `htop`?

`btop` and `htop` report RSS, which measures only the pages currently loaded in RAM. `pslens` reports Physical Footprint, which also includes compressed and swapped memory — matching what Activity Monitor shows. The two tools complement each other: use `btop`/`htop` for real-time CPU and process monitoring, and `pslens` when you want the full memory picture.

### Why does pslens call `vmmap`?

`vmmap --summary` provides the Physical Footprint value that Activity Monitor uses. Running `vmmap` directly for a single process is fine, but scanning all processes sequentially is slow (~0.5s per process). `pslens` extracts only the Footprint line from `vmmap`, runs up to 8 calls in parallel with a 3-second timeout per call, and presents the results in a readable table.

### Permission errors?

`vmmap` and `lsof` require access to process information. If you see permission errors, you can:

- Grant Terminal/your terminal app **Full Disk Access** in System Settings > Privacy & Security
- Run with `sudo` (not recommended for regular use)

## Roadmap

- [ ] `pslens watch` — Real-time refresh (similar to `btop`)
- [ ] `pslens --json` — JSON output for piping into `jq`
- [ ] `pslens export` — Export to CSV for analysis
- [ ] Linux support (via `/proc/self/smaps_rollup`)

## License

MIT — see [LICENSE](LICENSE).