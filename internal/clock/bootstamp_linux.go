//go:build linux

package clock

import (
	"errors"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const bootIDPath = "/proc/sys/kernel/random/boot_id"
const timeNSPath = "/proc/self/ns/time"
const timeChildrenNSPath = "/proc/self/ns/time_for_children"
const offsetsPath = "/proc/self/timens_offsets"

type bootProbe interface {
	procFS() error
	read(string, int) ([]byte, error)
	namespace(string) (BootNamespace, error)
	bootTime() (unix.Timespec, error)
}
type systemBootProbe struct{}
type systemBootSource struct{ probe bootProbe }

// SystemBootSource reads only supported Linux procfs/boottime inputs.
func SystemBootSource() BootSource { return systemBootSource{probe: systemBootProbe{}} }
func (systemBootProbe) procFS() error {
	var s unix.Statfs_t
	if e := unix.Statfs("/proc", &s); e != nil {
		return e
	}
	if s.Type != unix.PROC_SUPER_MAGIC {
		return ErrBootStamp
	}
	return nil
}
func (systemBootProbe) read(path string, max int) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(max)+1))
	e = errors.Join(e, f.Close())
	if e != nil {
		return nil, e
	}
	if len(b) > max {
		return nil, ErrBootStamp
	}
	return b, nil
}
func (systemBootProbe) namespace(path string) (BootNamespace, error) {
	var s unix.Stat_t
	if e := unix.Stat(path, &s); e != nil {
		return BootNamespace{}, e
	}
	n := BootNamespace{Device: uint64(s.Dev), Inode: s.Ino}
	if n.Device == 0 || n.Inode == 0 {
		return BootNamespace{}, ErrBootStamp
	}
	return n, nil
}
func (systemBootProbe) bootTime() (unix.Timespec, error) {
	var s unix.Timespec
	e := unix.ClockGettime(unix.CLOCK_BOOTTIME, &s)
	return s, e
}
func parseBootID(b []byte) (string, error) {
	s := strings.TrimSuffix(string(b), "\n")
	if !validBootID(s) {
		return "", ErrBootStamp
	}
	return s, nil
}
func parseZeroOffsets(b []byte) error {
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if len(lines) != 2 {
		return ErrBootStamp
	}
	seen := map[string]bool{}
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 3 || (parts[0] != "monotonic" && parts[0] != "boottime") || seen[parts[0]] {
			return ErrBootStamp
		}
		seen[parts[0]] = true
		for _, v := range parts[1:] {
			if v == "" {
				return ErrBootStamp
			}
			for _, c := range v {
				if c < '0' || c > '9' {
					return ErrBootStamp
				}
			}
			n, e := strconv.ParseUint(v, 10, 63)
			if e != nil || n != 0 {
				return ErrBootStamp
			}
		}
	}
	if !seen["monotonic"] || !seen["boottime"] {
		return ErrBootStamp
	}
	return nil
}
func timespecDuration(s unix.Timespec) (time.Duration, error) {
	sec := int64(s.Sec)
	ns := int64(s.Nsec)
	if sec < 0 || ns < 0 || ns >= int64(time.Second) || sec > math.MaxInt64/int64(time.Second) {
		return 0, ErrBootStamp
	}
	value := sec * int64(time.Second)
	if ns > math.MaxInt64-value {
		return 0, ErrBootStamp
	}
	return time.Duration(value + ns), nil
}
func (s systemBootSource) Snapshot() (BootStamp, error) {
	if s.probe == nil {
		return BootStamp{}, ErrBootStamp
	}
	if e := s.probe.procFS(); e != nil {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	first, e := s.probe.read(bootIDPath, 64)
	if e != nil {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	boot, e := parseBootID(first)
	if e != nil {
		return BootStamp{}, e
	}
	current, e := s.probe.namespace(timeNSPath)
	if e != nil {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	children, e := s.probe.namespace(timeChildrenNSPath)
	if e != nil || current != children {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	offsets, e := s.probe.read(offsetsPath, 1024)
	if e != nil {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	if e = parseZeroOffsets(offsets); e != nil {
		return BootStamp{}, e
	}
	sample, e := s.probe.bootTime()
	if e != nil {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	up, e := timespecDuration(sample)
	if e != nil {
		return BootStamp{}, e
	}
	currentAfter, e := s.probe.namespace(timeNSPath)
	if e != nil {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	childrenAfter, e := s.probe.namespace(timeChildrenNSPath)
	if e != nil || currentAfter != current || childrenAfter != children {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	last, e := s.probe.read(bootIDPath, 64)
	if e != nil {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	bootAfter, e := parseBootID(last)
	if e != nil || bootAfter != boot {
		return BootStamp{}, errors.Join(ErrBootStamp, e)
	}
	result := BootStamp{BootID: boot, Namespace: current, Uptime: up}
	if !validStamp(result) {
		return BootStamp{}, ErrBootStamp
	}
	return result, nil
}
