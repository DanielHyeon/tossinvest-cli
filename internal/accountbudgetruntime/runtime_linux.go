//go:build linux

// Package accountbudgetruntime constructs clients from a trusted launch manifest.
package accountbudgetruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"golang.org/x/sys/unix"
)

type Mode uint8

const (
	ModeOff Mode = iota + 1
	ModeRequired
)

type Config struct {
	Mode         Mode
	Manifest     string
	CleanupError func(error)
}
type manifest struct {
	Version   int                    `json:"version"`
	Root      string                 `json:"root"`
	Scope     string                 `json:"scope"`
	Instance  string                 `json:"instance"`
	Identity  accountbudget.Identity `json:"identity"`
	EventLock accountbudget.FileID   `json:"event_lock"`
}
type Factory struct {
	mode    Mode
	mu      sync.Mutex
	closed  bool
	budget  *accountbudget.Budget
	cleanup func(error)
}

// FromEnv treats an absent variable as OFF; required deployment enforcement is a launch responsibility.
func FromEnv(lookup func(string) (string, bool), key string, cleanup func(error)) (Config, error) {
	if lookup == nil || key == "" {
		return Config{}, accountbudget.ErrUnavailable
	}
	path, ok := lookup(key)
	if !ok {
		return Config{Mode: ModeOff}, nil
	}
	if path == "" {
		return Config{}, accountbudget.ErrUnavailable
	}
	return Config{Mode: ModeRequired, Manifest: path, CleanupError: cleanup}, nil
}
func Open(ctx context.Context, cfg Config) (*Factory, error) {
	if ctx == nil {
		return nil, accountbudget.ErrUnavailable
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	switch cfg.Mode {
	case ModeOff:
		if cfg.Manifest != "" {
			return nil, accountbudget.ErrUnavailable
		}
		return &Factory{mode: ModeOff}, nil
	case ModeRequired:
		if cfg.Manifest == "" || cfg.CleanupError == nil {
			return nil, accountbudget.ErrUnavailable
		}
	default:
		return nil, accountbudget.ErrUnavailable
	}
	m, e := readManifest(cfg.Manifest)
	if e == nil {
		rel, relErr := filepath.Rel(m.Root, cfg.Manifest)
		if relErr != nil || rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			e = errors.New("manifest must be outside budget root")
		}
	}
	if e != nil {
		return nil, errors.Join(accountbudget.ErrUnavailable, e)
	}
	b, e := accountbudget.OpenObserved(ctx, accountbudget.Config{Root: m.Root, Scope: m.Scope, Instance: m.Instance, Clock: clock.System()}, m.EventLock, clock.SystemBootSource())
	if e != nil {
		return nil, e
	}
	if e = b.VerifyIdentity(m.Identity); e != nil {
		return nil, errors.Join(e, b.Close())
	}
	return &Factory{mode: ModeRequired, budget: b, cleanup: cfg.CleanupError}, nil
}
func (f *Factory) NewClient(creds official.Credentials, cache string, opts ...official.Option) (*official.Client, error) {
	if f == nil {
		return nil, accountbudget.ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil, accountbudget.ErrUnavailable
	}
	if f.mode == ModeOff && f.budget == nil {
		return official.New(creds, cache, opts...), nil
	}
	if f.mode != ModeRequired || f.budget == nil || f.cleanup == nil {
		return nil, accountbudget.ErrUnavailable
	}
	return official.NewWithAccountBudget(creds, cache, f.budget, f.cleanup, opts...)
}
func (f *Factory) Close() error {
	if f == nil {
		return accountbudget.ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	if (f.mode == ModeOff && f.budget != nil) || (f.mode != ModeOff && (f.mode != ModeRequired || f.budget == nil || f.cleanup == nil)) {
		return accountbudget.ErrUnavailable
	}
	if f.budget != nil {
		if e := f.budget.Close(); e != nil {
			return e
		}
	}
	f.closed = true
	return nil
}

func readManifest(path string) (manifest, error) {
	var m manifest
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || path == "/" {
		return m, errors.New("invalid manifest path")
	}
	fd, e := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if e != nil {
		return m, e
	}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var ancestors []accountbudget.FileID
	recordDir := func(fd int) error {
		var st unix.Stat_t
		if err := unix.Fstat(fd, &st); err != nil {
			return err
		}
		ancestors = append(ancestors, accountbudget.FileID{Device: uint64(st.Dev), Inode: st.Ino})
		return nil
	}
	if e = recordDir(fd); e != nil {
		unix.Close(fd)
		return m, e
	}
	for _, part := range parts[:len(parts)-1] {
		next, err := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		unix.Close(fd)
		if err != nil {
			return m, err
		}
		fd = next
		if e = recordDir(fd); e != nil {
			unix.Close(fd)
			return m, e
		}
	}
	filefd, e := unix.Openat(fd, parts[len(parts)-1], unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	unix.Close(fd)
	if e != nil {
		return m, e
	}
	f := os.NewFile(uintptr(filefd), path)
	defer f.Close()
	var st unix.Stat_t
	if e = unix.Fstat(filefd, &st); e != nil {
		return m, e
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG || st.Mode&07777 != 0600 || st.Uid != uint32(os.Geteuid()) || st.Nlink != 1 || st.Size <= 0 || st.Size > 4096 {
		return m, errors.New("unsafe manifest")
	}
	data, e := io.ReadAll(io.LimitReader(f, 4097))
	if e != nil || len(data) > 4096 || int64(len(data)) != st.Size {
		return m, errors.New("manifest read failed")
	}
	var after unix.Stat_t
	if e = unix.Fstat(filefd, &after); e != nil || st.Dev != after.Dev || st.Ino != after.Ino || st.Size != after.Size || st.Mtim != after.Mtim {
		return m, errors.New("manifest changed")
	}
	if e = rejectDuplicates(data); e != nil {
		return m, e
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e = d.Decode(&m); e != nil {
		return m, e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return m, errors.New("trailing JSON")
	}
	for _, ancestor := range ancestors {
		if ancestor == m.Identity.Root {
			return m, errors.New("manifest is inside budget root")
		}
	}
	if m.Version != 2 || m.Root == "" || m.Scope == "" || m.Instance == "" || m.EventLock.Device == 0 || m.EventLock.Inode == 0 {
		return m, errors.New("invalid manifest fields")
	}
	return m, nil
}
func rejectDuplicates(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	allowed := map[string]map[string]bool{
		"":              {"version": true, "root": true, "scope": true, "instance": true, "identity": true, "event_lock": true},
		"identity":      {"root": true, "queue": true, "wire": true},
		"event_lock":    {"device": true, "inode": true},
		"identity.root": {"device": true, "inode": true}, "identity.queue": {"device": true, "inode": true}, "identity.wire": {"device": true, "inode": true},
	}
	var walk func(string) error
	walk = func(path string) error {
		tok, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				key, ok := k.(string)
				if !ok || seen[key] || !allowed[path][key] {
					return errors.New("duplicate or invalid field")
				}
				seen[key] = true
				child := key
				if path != "" {
					child = path + "." + key
				}
				if e = walk(child); e != nil {
					return e
				}
			}
			end, e := d.Token()
			if e != nil || end != json.Delim('}') {
				return errors.New("invalid object")
			}
		case '[':
			for d.More() {
				if e = walk(path); e != nil {
					return e
				}
			}
			end, e := d.Token()
			if e != nil || end != json.Delim(']') {
				return errors.New("invalid array")
			}
		default:
			return errors.New("unexpected delimiter")
		}
		return nil
	}
	return walk("")
}

// Budget returns a borrowed binding for trusted callers; Factory owns Close.
// A nil budget means OFF. This does not prove authority or readiness.
func (f *Factory) Budget() (*accountbudget.Budget, error) {
	if f == nil {
		return nil, accountbudget.ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil, accountbudget.ErrUnavailable
	}
	if (f.mode == ModeOff && f.budget != nil) || (f.mode == ModeRequired && (f.budget == nil || f.cleanup == nil)) || (f.mode != ModeOff && f.mode != ModeRequired) {
		return nil, accountbudget.ErrUnavailable
	}
	return f.budget, nil
}
