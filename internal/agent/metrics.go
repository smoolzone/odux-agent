package agent

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/smoolzone/odux-agent/internal/api"
)

// CollectMetrics snapshots the local machine. Linux reads /proc + statfs; other
// platforms return zeros (the caller omits the payload when ramTotal is 0).
func CollectMetrics() api.Metrics {
	var m api.Metrics
	if runtime.GOOS != "linux" {
		return m
	}
	m.Load1, m.Load5, m.Load15 = loadAvg()
	m.RamUsed, m.RamTotal = memInfo()
	m.DiskUsed, m.DiskTotal = diskUsage()
	m.CpuPct = cpuPct()
	return m
}

func loadAvg() (float64, float64, float64) {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0
	}
	f := strings.Fields(string(b))
	if len(f) < 3 {
		return 0, 0, 0
	}
	l1, _ := strconv.ParseFloat(f[0], 64)
	l5, _ := strconv.ParseFloat(f[1], 64)
	l15, _ := strconv.ParseFloat(f[2], 64)
	return l1, l5, l15
}

// memInfo returns (used, total) in bytes.
func memInfo() (usedBytes, totalBytes float64) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	var totalKB, availKB uint64
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(f[1], 10, 64)
		switch f[0] {
		case "MemTotal:":
			totalKB = v
		case "MemAvailable:":
			availKB = v
		}
	}
	return float64(totalKB-availKB) * 1024, float64(totalKB) * 1024
}

// diskUsage returns (used, total) in bytes for the root filesystem.
func diskUsage() (usedBytes, totalBytes float64) {
	var st syscall.Statfs_t
	if err := syscall.Statfs("/", &st); err != nil {
		return 0, 0
	}
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	return float64(total - free), float64(total)
}

type cpuSnap struct{ idle, total uint64 }

// cpuPct samples /proc/stat twice and returns busy percentage (0-100).
func cpuPct() float64 {
	s1 := cpuSample()
	time.Sleep(200 * time.Millisecond)
	s2 := cpuSample()
	idle := s2.idle - s1.idle
	total := s2.total - s1.total
	if total <= 0 {
		return 0
	}
	return (1 - float64(idle)/float64(total)) * 100
}

func cpuSample() cpuSnap {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuSnap{}
	}
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		f := strings.Fields(line)[1:] // drop the "cpu" token
		var idle, total uint64
		for i, tok := range f {
			v, _ := strconv.ParseUint(tok, 10, 64)
			total += v
			if i == 3 || i == 4 { // idle + iowait
				idle += v
			}
		}
		return cpuSnap{idle, total}
	}
	return cpuSnap{}
}
