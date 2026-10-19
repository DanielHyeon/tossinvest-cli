//go:build linux

package strategyobserve

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

const maxLedger = 64 << 20
const maxRecords = 65536
const marker = "strategyobserve-v1\n"

type checkpoint struct {
	Schema string `json:"schema"`
	Count  int    `json:"count"`
	LastID string `json:"last_id"`
	Offset int64  `json:"offset"`
	Prefix string `json:"prefix"`
}
type Store struct {
	dir    int
	file   *os.File
	seen   map[string]string
	count  int
	offset int64
	last   string
	hash   hash.Hash
	failed bool
	fault  func(string) error
	write  func(*os.File, []byte) (int, error)
}

func safeParts(path string) ([]string, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("absolute clean output path required")
	}
	p := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(p) == 0 || p[0] == "" {
		return nil, errors.New("root is not an output directory")
	}
	for _, s := range p {
		n := strings.ToLower(s)
		if s == "" || s == "." || s == ".." || n == "openspec" || n == "r1" || strings.Contains(n, "a128") || strings.Contains(n, "a147") {
			return nil, errors.New("protected or invalid output path")
		}
	}
	return p, nil
}
func openDir(path string) (int, bool, error) {
	parts, err := safeParts(path)
	if err != nil {
		return -1, false, err
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, false, err
	}
	created := false
	for i, p := range parts {
		next, e := unix.Openat(fd, p, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if e == unix.ENOENT && i == len(parts)-1 {
			if e = unix.Mkdirat(fd, p, 0700); e == nil {
				created = true
				e = unix.Fsync(fd)
			}
			if e == nil {
				next, e = unix.Openat(fd, p, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
			}
		}
		unix.Close(fd)
		if e != nil {
			return -1, false, e
		}
		fd = next
	}
	return fd, created, nil
}
func openRegular(dir int, name string, flags int, mode uint32) (*os.File, error) {
	fd, err := unix.Openat(dir, name, flags|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, mode)
	if err != nil {
		return nil, err
	}
	var st unix.Stat_t
	if err = unix.Fstat(fd, &st); err != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 {
		unix.Close(fd)
		return nil, errors.New("unsafe file type or hardlink")
	}
	return os.NewFile(uintptr(fd), name), nil
}
func readFile(dir int, name string, limit int64) ([]byte, error) {
	f, err := openRegular(dir, name, unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("file exceeds limit")
	}
	return b, nil
}
func Open(path string) (*Store, error) { return open(path, nil) }

func open(path string, fault func(string) error) (*Store, error) {
	dir, created, err := openDir(path)
	if err != nil {
		return nil, err
	}
	s := &Store{dir: dir, seen: make(map[string]string), hash: sha256.New(), fault: fault}
	ok := false
	defer func() {
		if !ok {
			s.Close()
		}
	}()
	if err = unix.Flock(dir, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return nil, fmt.Errorf("directory locked: %w", err)
	}
	scanfd, err := unix.Dup(dir)
	if err != nil {
		return nil, err
	}
	scan := os.NewFile(uintptr(scanfd), "observation-dir")
	names, err := scan.ReadDir(-1)
	_ = scan.Close()
	if err != nil {
		return nil, err
	}
	if created || len(names) == 0 {
		if len(names) != 0 {
			return nil, errors.New("foreign directory")
		}
		m, e := openRegular(dir, "marker", unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL, 0600)
		if e != nil {
			return nil, e
		}
		n, e := io.WriteString(m, marker)
		if e == nil && n != len(marker) {
			e = io.ErrShortWrite
		}
		if e == nil {
			e = m.Sync()
		}
		closeErr := m.Close()
		if e != nil {
			return nil, e
		}
		if closeErr != nil {
			return nil, closeErr
		}
		s.file, e = openRegular(dir, "ledger.jsonl", unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_APPEND, 0600)
		if e != nil {
			return nil, e
		}
		if e = s.file.Sync(); e != nil {
			return nil, e
		}
		if e = unix.Fsync(dir); e != nil {
			return nil, e
		}
	} else {
		allowed := map[string]bool{"marker": true, "ledger.jsonl": true, "checkpoint.json": true}
		for _, n := range names {
			if !allowed[n.Name()] {
				return nil, fmt.Errorf("foreign file: %s", n.Name())
			}
		}
		b, e := readFile(dir, "marker", 256)
		if e != nil || string(b) != marker {
			return nil, errors.New("missing or invalid marker")
		}
		s.file, e = openRegular(dir, "ledger.jsonl", unix.O_RDWR|unix.O_APPEND, 0)
		if e != nil {
			return nil, e
		}
	}
	if err = s.recover(); err != nil {
		return nil, err
	}
	ok = true
	return s, nil
}
func (s *Store) recover() error {
	info, err := s.file.Stat()
	if err != nil {
		return err
	}
	if info.Size() > maxLedger {
		return errors.New("ledger limit")
	}
	data, err := io.ReadAll(io.LimitReader(io.NewSectionReader(s.file, 0, info.Size()), maxLedger+1))
	if err != nil {
		return err
	}
	var cp checkpoint
	raw, err := readFile(s.dir, "checkpoint.json", 4096)
	hasCP := err == nil
	if err != nil && !errors.Is(err, unix.ENOENT) {
		return err
	}
	if hasCP {
		d := json.NewDecoder(bytes.NewReader(raw))
		d.DisallowUnknownFields()
		var extra any
		if d.Decode(&cp) != nil || d.Decode(&extra) != io.EOF || cp.Schema != Schema || cp.Offset < 0 || cp.Offset > int64(len(data)) || cp.Count < 1 || cp.Count > maxRecords {
			return errors.New("invalid checkpoint")
		}
	}
	scanner := bufio.NewReader(bytes.NewReader(data))
	offset := int64(0)
	count := 0
	last := ""
	matched := !hasCP
	for {
		line, e := scanner.ReadBytes('\n')
		if len(line) > 0 {
			if e != nil {
				return errors.New("torn ledger")
			}
			entry, err := ValidateLine(line)
			if err != nil {
				return err
			}
			old, exists := s.seen[entry.ID]
			if exists {
				if old != entry.PayloadDigest {
					return errors.New("conflicting record id")
				}
				return errors.New("duplicate record in ledger")
			}
			s.seen[entry.ID] = entry.PayloadDigest
			count++
			if count > maxRecords {
				return errors.New("record limit")
			}
			offset += int64(len(line))
			last = entry.ID
			if hasCP && offset == cp.Offset {
				sum := sha256.Sum256(data[:offset])
				matched = cp.Count == count && cp.LastID == last && cp.Prefix == hex.EncodeToString(sum[:])
			}
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
	}
	if !matched {
		return errors.New("checkpoint mismatch")
	}
	s.count = count
	s.offset = offset
	s.last = last
	_, _ = s.hash.Write(data)
	if err := s.inject("recovery-sync"); err != nil {
		return err
	}
	if err := s.file.Sync(); err != nil {
		return err
	}
	if count > 0 && (!hasCP || cp.Count < count) {
		recovered := checkpoint{Schema: Schema, Count: count, LastID: last, Offset: offset, Prefix: hex.EncodeToString(s.hash.Sum(nil))}
		b, err := json.Marshal(recovered)
		if err != nil {
			return err
		}
		if err = s.writeCheckpoint(b); err != nil {
			return err
		}
	} else {
		if hasCP {
			f, err := openRegular(s.dir, "checkpoint.json", unix.O_RDONLY, 0)
			if err != nil {
				return err
			}
			err = f.Sync()
			closeErr := f.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		}
		if err := s.inject("recovery-directory-sync"); err != nil {
			return err
		}
		if err := unix.Fsync(s.dir); err != nil {
			return err
		}
	}

	return nil
}
func (s *Store) Append(line []byte) (bool, error) {
	if s.failed {
		return false, errors.New("writer failed")
	}
	entry, err := ValidateLine(line)
	if err != nil {
		return false, err
	}
	if old, ok := s.seen[entry.ID]; ok {
		if old != entry.PayloadDigest {
			s.failed = true
			return false, errors.New("conflicting record id")
		}
		return true, nil
	}
	if s.count >= maxRecords || s.offset+int64(len(line)) > maxLedger {
		s.failed = true
		return false, errors.New("ledger limit")
	}
	var n int
	if s.write != nil {
		n, err = s.write(s.file, line)
	} else {
		n, err = s.file.Write(line)
	}
	if err != nil || n != len(line) {
		s.failed = true
		return false, errors.New("append failed or short write")
	}
	if err = s.inject("append-sync"); err == nil {
		err = s.file.Sync()
	}
	if err != nil {
		s.failed = true
		return false, err
	}
	_, _ = s.hash.Write(line)
	s.offset += int64(len(line))
	s.count++
	s.last = entry.ID
	s.seen[entry.ID] = entry.PayloadDigest
	cp := checkpoint{Schema: Schema, Count: s.count, LastID: s.last, Offset: s.offset, Prefix: hex.EncodeToString(s.hash.Sum(nil))}
	b, err := json.Marshal(cp)
	if err == nil {
		err = s.writeCheckpoint(b)
	}
	if err != nil {
		s.failed = true
		return false, err
	}
	return false, nil
}
func (s *Store) inject(stage string) error {
	if s.fault != nil {
		return s.fault(stage)
	}
	return nil
}
func (s *Store) writeCheckpoint(b []byte) error {
	f, err := openRegular(s.dir, ".checkpoint.tmp", unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL, 0600)
	if err != nil {
		return err
	}
	n, err := f.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = s.inject("checkpoint-sync")
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = s.inject("checkpoint-rename"); err != nil {
		return err
	}
	if err = unix.Renameat(s.dir, ".checkpoint.tmp", s.dir, "checkpoint.json"); err != nil {
		return err
	}
	if err = s.inject("directory-sync"); err != nil {
		return err
	}
	return unix.Fsync(s.dir)
}
func (s *Store) Close() error {
	var err error
	if s.file != nil {
		err = s.file.Close()
		s.file = nil
	}
	if s.dir >= 0 {
		_ = unix.Flock(s.dir, unix.LOCK_UN)
		e := unix.Close(s.dir)
		s.dir = -1
		if err == nil {
			err = e
		}
	}
	return err
}
