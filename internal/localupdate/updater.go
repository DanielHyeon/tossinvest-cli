// Package localupdate installs one staged sibling executable without accepting
// an operator-supplied path or executing candidate code.
package localupdate

import (
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	expectedModule  = "github.com/JungHoonGhae/tossinvest-cli"
	expectedCommand = expectedModule + "/cmd/tossctl"
	maxStagedBinary = 256 << 20
)

var (
	ErrCandidateChanged   = errors.New("localupdate: staged candidate changed after operator review")
	ErrCurrentChanged     = errors.New("localupdate: installed executable changed after this console started")
	ErrProvenanceMismatch = errors.New("localupdate: staged executable does not match verified release provenance")
	ErrUnsupported        = errors.New("localupdate: executable replacement is supported on Unix only")
)

type Metadata struct {
	Path          string
	Size          int64
	ModTime       time.Time
	SHA256        string
	GoVersion     string
	ModulePath    string
	ModuleVersion string
	CommandPath   string
	GOOS          string
	GOARCH        string
	VCSRevision   string
	VCSModified   bool
}

type Inspection struct {
	Current       Metadata
	Candidate     Metadata
	CandidatePath string
	Installable   bool
	Reason        string
}

type Result struct {
	OldSHA256    string
	NewSHA256    string
	RollbackPath string
}

type StageResult struct {
	Metadata     Metadata
	RecoveryPath string
}

type Updater struct {
	mu sync.Mutex

	currentPath     string
	candidatePath   string
	rollbackPath    string
	startedWith     Metadata
	expectedModule  string
	expectedCommand string
	goos            string
	goarch          string

	rename  func(string, string) error
	syncDir func(string) error
}

func New(currentPath string) (*Updater, error) {
	return newUpdater(currentPath, expectedModule, expectedCommand, runtime.GOOS, runtime.GOARCH)
}

func newUpdater(currentPath, modulePath, commandPath, goos, goarch string) (*Updater, error) {
	currentPath = filepath.Clean(strings.TrimSpace(currentPath))
	if currentPath == "." || !filepath.IsAbs(currentPath) {
		return nil, fmt.Errorf("localupdate: current executable path must be absolute")
	}
	started, err := inspectPath(currentPath, modulePath, commandPath, goos, goarch)
	if err != nil {
		return nil, fmt.Errorf("localupdate: inspecting current executable: %w", err)
	}
	return &Updater{
		currentPath:     currentPath,
		candidatePath:   currentPath + ".candidate",
		rollbackPath:    currentPath + ".rollback",
		startedWith:     started,
		expectedModule:  modulePath,
		expectedCommand: commandPath,
		goos:            goos,
		goarch:          goarch,
		rename:          os.Rename,
		syncDir:         syncDirectory,
	}, nil
}

func (u *Updater) Inspect() Inspection {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.inspectLocked()
}

func (u *Updater) inspectLocked() Inspection {
	view := Inspection{
		Current:       u.startedWith,
		CandidatePath: u.candidatePath,
	}
	candidate, err := inspectPath(
		u.candidatePath, u.expectedModule, u.expectedCommand, u.goos, u.goarch)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			view.Reason = "staged candidate 없음"
		} else {
			view.Reason = err.Error()
		}
		return view
	}
	view.Candidate = candidate
	view.Installable = replacementSupported()
	if !view.Installable {
		view.Reason = ErrUnsupported.Error()
	}
	return view
}

func (u *Updater) Install(reviewedSHA256 string, commitGuard func() error) (Result, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !replacementSupported() {
		return Result{}, ErrUnsupported
	}
	reviewedSHA256 = strings.ToLower(strings.TrimSpace(reviewedSHA256))
	if reviewedSHA256 == "" {
		return Result{}, fmt.Errorf("%w: no reviewed SHA-256 was supplied", ErrCandidateChanged)
	}

	dir := filepath.Dir(u.currentPath)
	prepared, preparedMeta, err := u.prepareCandidate(dir)
	if err != nil {
		return Result{}, err
	}
	defer os.Remove(prepared)
	if preparedMeta.SHA256 != reviewedSHA256 {
		return Result{}, fmt.Errorf("%w: reviewed %s, prepared %s",
			ErrCandidateChanged, reviewedSHA256, preparedMeta.SHA256)
	}

	current, err := openNoFollowExecutable(u.currentPath)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrCurrentChanged, err)
	}
	defer current.Close()
	currentMeta, err := inspectOpen(
		current, u.currentPath, u.expectedModule, u.expectedCommand, u.goos, u.goarch)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrCurrentChanged, err)
	}
	if currentMeta.SHA256 != u.startedWith.SHA256 {
		return Result{}, fmt.Errorf("%w: started %s, current %s",
			ErrCurrentChanged, u.startedWith.SHA256, currentMeta.SHA256)
	}
	if commitGuard != nil {
		if err := commitGuard(); err != nil {
			return Result{}, fmt.Errorf("localupdate: commit guard refused installation: %w", err)
		}
	}

	rollbackTemp, err := copyOpenToTemp(current, dir, "."+filepath.Base(u.currentPath)+".rollback-")
	if err != nil {
		return Result{}, fmt.Errorf("localupdate: preparing rollback: %w", err)
	}
	defer os.Remove(rollbackTemp)
	rollbackMeta, err := inspectPath(
		rollbackTemp, u.expectedModule, u.expectedCommand, u.goos, u.goarch)
	if err != nil {
		return Result{}, fmt.Errorf("localupdate: validating prepared rollback: %w", err)
	}
	if rollbackMeta.SHA256 != currentMeta.SHA256 {
		return Result{}, fmt.Errorf("%w: current bytes changed while preparing rollback",
			ErrCurrentChanged)
	}
	if err := sameOpenFile(current, u.currentPath); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrCurrentChanged, err)
	}
	if err := u.rename(rollbackTemp, u.rollbackPath); err != nil {
		return Result{}, fmt.Errorf("localupdate: publishing rollback: %w", err)
	}
	if err := u.syncDir(dir); err != nil {
		return Result{}, fmt.Errorf("localupdate: syncing rollback directory: %w", err)
	}
	if err := sameOpenFile(current, u.currentPath); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrCurrentChanged, err)
	}

	if err := u.rename(prepared, u.currentPath); err != nil {
		return Result{}, fmt.Errorf("localupdate: replacing current executable: %w", err)
	}
	if err := u.syncDir(dir); err != nil {
		restoreErr := u.restoreRollback(dir)
		if restoreErr != nil {
			return Result{}, fmt.Errorf(
				"localupdate: replacement directory sync failed: %v; rollback restoration also failed: %w",
				err, restoreErr)
		}
		return Result{}, fmt.Errorf(
			"localupdate: replacement directory sync failed and the previous executable was restored: %w", err)
	}

	return Result{
		OldSHA256:    currentMeta.SHA256,
		NewSHA256:    preparedMeta.SHA256,
		RollbackPath: u.rollbackPath,
	}, nil
}

// StageCandidate validates verified release bytes without executing them and
// atomically publishes only the updater's fixed sibling candidate. A non-empty
// expectedRevision binds the Go binary to the commit in the verified SLSA
// statement.
func (u *Updater) StageCandidate(reader io.Reader, expectedRevision string) (StageResult, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !replacementSupported() {
		return StageResult{}, ErrUnsupported
	}
	if reader == nil {
		return StageResult{}, errors.New("localupdate: staged executable reader is nil")
	}
	dir := filepath.Dir(u.currentPath)
	prepared, err := copyReaderToTemp(
		reader, dir, "."+filepath.Base(u.currentPath)+".download-", maxStagedBinary)
	if err != nil {
		return StageResult{}, fmt.Errorf("localupdate: preparing downloaded candidate: %w", err)
	}
	defer os.Remove(prepared)

	meta, err := inspectPath(
		prepared, u.expectedModule, u.expectedCommand, u.goos, u.goarch)
	if err != nil {
		return StageResult{}, fmt.Errorf("localupdate: validating downloaded candidate: %w", err)
	}
	expectedRevision = strings.ToLower(strings.TrimSpace(expectedRevision))
	if expectedRevision != "" &&
		(strings.ToLower(meta.VCSRevision) != expectedRevision || meta.VCSModified) {
		return StageResult{}, fmt.Errorf(
			"%w: binary revision=%q modified=%t, attested revision=%q",
			ErrProvenanceMismatch, meta.VCSRevision, meta.VCSModified, expectedRevision)
	}

	recovery := ""
	candidate, openErr := openNoFollowExecutable(u.candidatePath)
	switch {
	case openErr == nil:
		recovery, err = copyOpenToTemp(
			candidate, dir, "."+filepath.Base(u.currentPath)+".candidate-recovery-")
		closeErr := candidate.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			return StageResult{}, fmt.Errorf("localupdate: preserving prior candidate: %w", err)
		}
		if err := u.syncDir(dir); err != nil {
			_ = os.Remove(recovery)
			return StageResult{}, fmt.Errorf("localupdate: syncing candidate recovery: %w", err)
		}
	case errors.Is(openErr, os.ErrNotExist):
		// No earlier candidate needs preserving.
	default:
		return StageResult{}, fmt.Errorf("localupdate: opening prior candidate: %w", openErr)
	}

	if err := u.rename(prepared, u.candidatePath); err != nil {
		if recovery != "" {
			_ = os.Remove(recovery)
		}
		return StageResult{}, fmt.Errorf("localupdate: publishing candidate: %w", err)
	}
	if err := u.syncDir(dir); err != nil {
		return u.recoverCandidateAfterPublishFailure(dir, recovery, err)
	}
	if recovery != "" {
		_ = os.Remove(recovery)
	}
	meta.Path = u.candidatePath
	return StageResult{Metadata: meta}, nil
}

func (u *Updater) recoverCandidateAfterPublishFailure(
	dir, recovery string,
	publishErr error,
) (StageResult, error) {
	if recovery == "" {
		removeErr := os.Remove(u.candidatePath)
		syncErr := u.syncDir(dir)
		if removeErr == nil && syncErr == nil {
			return StageResult{}, fmt.Errorf(
				"localupdate: candidate directory sync failed; the new candidate was removed: %w",
				publishErr)
		}
		return StageResult{}, fmt.Errorf(
			"localupdate: candidate directory sync failed (%v) and cleanup is uncertain "+
				"(remove=%v sync=%v)", publishErr, removeErr, syncErr)
	}

	recoveryFile, err := openNoFollowExecutable(recovery)
	if err != nil {
		return StageResult{RecoveryPath: recovery}, fmt.Errorf(
			"localupdate: candidate directory sync failed (%v); opening recovery %s: %w",
			publishErr, recovery, err)
	}
	restore, copyErr := copyOpenToTemp(
		recoveryFile, dir, "."+filepath.Base(u.currentPath)+".restore-")
	closeErr := recoveryFile.Close()
	if copyErr == nil && closeErr != nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		return StageResult{RecoveryPath: recovery}, fmt.Errorf(
			"localupdate: candidate directory sync failed (%v); preparing recovery %s: %w",
			publishErr, recovery, copyErr)
	}
	defer os.Remove(restore)
	if err := u.rename(restore, u.candidatePath); err != nil {
		return StageResult{RecoveryPath: recovery}, fmt.Errorf(
			"localupdate: candidate directory sync failed (%v); restoring prior candidate "+
				"failed and recovery remains at %s: %w", publishErr, recovery, err)
	}
	if err := u.syncDir(dir); err != nil {
		return StageResult{RecoveryPath: recovery}, fmt.Errorf(
			"localupdate: candidate directory sync failed (%v); restored bytes could not be "+
				"durably synced and recovery remains at %s: %w", publishErr, recovery, err)
	}
	_ = os.Remove(recovery)
	return StageResult{}, fmt.Errorf(
		"localupdate: candidate directory sync failed and the prior candidate was restored: %w",
		publishErr)
}

func (u *Updater) prepareCandidate(dir string) (string, Metadata, error) {
	candidate, err := openNoFollowExecutable(u.candidatePath)
	if err != nil {
		return "", Metadata{}, fmt.Errorf("localupdate: opening staged candidate: %w", err)
	}
	defer candidate.Close()

	prepared, err := copyOpenToTemp(candidate, dir, "."+filepath.Base(u.currentPath)+".candidate-")
	if err != nil {
		return "", Metadata{}, fmt.Errorf("localupdate: preparing candidate: %w", err)
	}
	meta, err := inspectPath(
		prepared, u.expectedModule, u.expectedCommand, u.goos, u.goarch)
	if err != nil {
		_ = os.Remove(prepared)
		return "", Metadata{}, fmt.Errorf("localupdate: validating prepared candidate: %w", err)
	}
	meta.Path = u.candidatePath
	return prepared, meta, nil
}

func (u *Updater) restoreRollback(dir string) error {
	rollback, err := openNoFollowExecutable(u.rollbackPath)
	if err != nil {
		return err
	}
	defer rollback.Close()
	restore, err := copyOpenToTemp(rollback, dir, "."+filepath.Base(u.currentPath)+".restore-")
	if err != nil {
		return err
	}
	defer os.Remove(restore)
	if err := u.rename(restore, u.currentPath); err != nil {
		return err
	}
	return u.syncDir(dir)
}

func copyOpenToTemp(source *os.File, dir, pattern string) (path string, err error) {
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	temp, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	path = temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(path)
	}
	if _, err := io.Copy(temp, source); err != nil {
		cleanup()
		return "", err
	}
	if err := temp.Chmod(0o755); err != nil {
		cleanup()
		return "", err
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return "", err
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func copyReaderToTemp(
	source io.Reader,
	dir, pattern string,
	maxBytes int64,
) (path string, err error) {
	temp, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	path = temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(path)
	}
	written, err := io.Copy(temp, io.LimitReader(source, maxBytes+1))
	if err != nil {
		cleanup()
		return "", err
	}
	if written > maxBytes {
		cleanup()
		return "", fmt.Errorf("executable exceeds %d bytes", maxBytes)
	}
	if err := temp.Chmod(0o755); err != nil {
		cleanup()
		return "", err
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return "", err
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func inspectPath(path, modulePath, commandPath, goos, goarch string) (Metadata, error) {
	file, err := openNoFollowExecutable(path)
	if err != nil {
		return Metadata{}, err
	}
	defer file.Close()
	return inspectOpen(file, path, modulePath, commandPath, goos, goarch)
}

func inspectOpen(
	file *os.File,
	path, modulePath, commandPath, goos, goarch string,
) (Metadata, error) {
	info, err := file.Stat()
	if err != nil {
		return Metadata{}, err
	}
	if !info.Mode().IsRegular() {
		return Metadata{}, fmt.Errorf("%s is not a regular file", path)
	}
	if info.Mode().Perm()&0o111 == 0 {
		return Metadata{}, fmt.Errorf("%s에 실행 비트가 없다", path)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return Metadata{}, err
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return Metadata{}, err
	}
	bi, err := buildinfo.Read(file)
	if err != nil {
		return Metadata{}, fmt.Errorf("%s has no readable Go build information: %w", path, err)
	}
	if bi.Main.Path != modulePath {
		return Metadata{}, fmt.Errorf("%s module is %q, want %q", path, bi.Main.Path, modulePath)
	}
	if bi.Path != commandPath {
		return Metadata{}, fmt.Errorf("%s command is %q, want %q", path, bi.Path, commandPath)
	}
	settings := make(map[string]string, len(bi.Settings))
	for _, setting := range bi.Settings {
		settings[setting.Key] = setting.Value
	}
	if settings["GOOS"] != goos || settings["GOARCH"] != goarch {
		return Metadata{}, fmt.Errorf("%s platform is %s/%s, want %s/%s",
			path, settings["GOOS"], settings["GOARCH"], goos, goarch)
	}
	return Metadata{
		Path:          path,
		Size:          info.Size(),
		ModTime:       info.ModTime().UTC(),
		SHA256:        hex.EncodeToString(hash.Sum(nil)),
		GoVersion:     bi.GoVersion,
		ModulePath:    bi.Main.Path,
		ModuleVersion: bi.Main.Version,
		CommandPath:   bi.Path,
		GOOS:          settings["GOOS"],
		GOARCH:        settings["GOARCH"],
		VCSRevision:   settings["vcs.revision"],
		VCSModified:   settings["vcs.modified"] == "true",
	}, nil
}
