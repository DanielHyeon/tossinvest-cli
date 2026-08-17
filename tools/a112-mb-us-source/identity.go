package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

const prescribedBuildCommand = "<verified-absolute-go> build -mod=readonly -trimpath -buildvcs=false -o <verified-temp>/a112-mb-us-source ./tools/a112-mb-us-source"

const maxIdentityInputBytes int64 = 256 << 20

// identitySnapshot is deliberately source/build oriented. It has no wall-clock
// field: local time never becomes source-observation or candle-finality proof.
type identitySnapshot struct {
	BaseCommit           string            `json:"base_commit"`
	RepoRoot             string            `json:"repo_root"`
	BuildCommand         string            `json:"build_command"`
	ExecutableSHA256     string            `json:"executable_sha256"`
	GoBinary             string            `json:"go_binary"`
	GoVersion            string            `json:"go_version"`
	GoToolchainSHA256    string            `json:"go_toolchain_sha256"`
	GoDistributionSHA256 string            `json:"go_distribution_sha256"`
	GitBinarySHA256      string            `json:"git_binary_sha256"`
	TargetGOOS           string            `json:"target_goos"`
	FrozenInputSHA256    string            `json:"frozen_input_sha256"`
	SourceSHA256         map[string]string `json:"source_sha256"`
	CompiledSourceSHA256 map[string]string `json:"compiled_source_sha256"`
	TrackedDiffSHA256    string            `json:"tracked_diff_sha256"`
	Untracked            []string          `json:"untracked"`
	UntrackedContent     map[string]string `json:"untracked_content_sha256"`
	UntrackedSHA256      string            `json:"untracked_sha256"`
	WorktreeHEAD         string            `json:"worktree_head"`
	goExecution          goBinaryIdentity
	gitExecution         gitBinaryIdentity
}

func (s *identitySnapshot) close() error {
	if s == nil {
		return nil
	}
	return errors.Join(s.goExecution.close(), s.gitExecution.close())
}

func (s identitySnapshot) equal(other identitySnapshot) bool {
	if s.BaseCommit != other.BaseCommit || s.RepoRoot != other.RepoRoot || s.BuildCommand != other.BuildCommand || s.ExecutableSHA256 != other.ExecutableSHA256 || s.GoBinary != other.GoBinary ||
		s.GoVersion != other.GoVersion || s.GoToolchainSHA256 != other.GoToolchainSHA256 || s.GoDistributionSHA256 != other.GoDistributionSHA256 || s.GitBinarySHA256 != other.GitBinarySHA256 || s.TargetGOOS != other.TargetGOOS || s.FrozenInputSHA256 != other.FrozenInputSHA256 || s.TrackedDiffSHA256 != other.TrackedDiffSHA256 || s.UntrackedSHA256 != other.UntrackedSHA256 || s.WorktreeHEAD != other.WorktreeHEAD || len(s.SourceSHA256) != len(other.SourceSHA256) || len(s.CompiledSourceSHA256) != len(other.CompiledSourceSHA256) || len(s.Untracked) != len(other.Untracked) || len(s.UntrackedContent) != len(other.UntrackedContent) {
		return false
	}
	for path, digest := range s.CompiledSourceSHA256 {
		if other.CompiledSourceSHA256[path] != digest {
			return false
		}
	}
	for path, digest := range s.UntrackedContent {
		if other.UntrackedContent[path] != digest {
			return false
		}
	}
	for index := range s.Untracked {
		if s.Untracked[index] != other.Untracked[index] {
			return false
		}
	}
	for path, digest := range s.SourceSHA256 {
		if other.SourceSHA256[path] != digest {
			return false
		}
	}
	return true
}

type identityProvider interface {
	snapshot(config) (identitySnapshot, error)
}

type contextualIdentityProvider interface {
	snapshotContext(context.Context, config) (identitySnapshot, error)
}

type systemIdentity struct {
	executable func() (string, error)
}

// goBinaryIdentity is the caller-selected build authority. It is deliberately
// derived from --go-binary, never from the running runtime's Go-root lookup,
// ambient GOROOT, PATH, or inherited GO* controls.
type goBinaryIdentity struct {
	Path         string
	SHA256       string
	execution    *executionSnapshot
	distribution *goDistributionSnapshot
}

func (g goBinaryIdentity) commandPath() (string, error) {
	if g.execution == nil {
		return "", errors.New("selected Go execution snapshot is unavailable")
	}
	return g.execution.commandPath()
}

func (g goBinaryIdentity) close() error {
	return errors.Join(g.execution.close(), g.distribution.close())
}

func (g goBinaryIdentity) environment(cache, moduleCache, home string) ([]string, error) {
	if g.distribution == nil {
		return nil, errors.New("private Go distribution snapshot is unavailable")
	}
	root, err := g.distribution.rootPath()
	if err != nil {
		return nil, err
	}
	return selectedGoEnvironmentWithRoot(cache, moduleCache, home, root), nil
}

type gitBinaryIdentity struct {
	Path      string
	SHA256    string
	execution *executionSnapshot
}

func (g gitBinaryIdentity) commandPath() (string, error) {
	if g.execution == nil {
		return "", errors.New("Git execution snapshot is unavailable")
	}
	return g.execution.commandPath()
}

func (g gitBinaryIdentity) close() error {
	if g.execution == nil {
		return nil
	}
	return g.execution.close()
}

// buildRunner is injected only by tests; production must use the fixed
// prescribed builder, never a caller-supplied command string.
type buildRunner interface {
	rebuild(context.Context, string, goBinaryIdentity) ([]byte, string, error)
}

type systemBuildRunner struct{}

func (systemBuildRunner) rebuild(ctx context.Context, root string, goBinary goBinaryIdentity) ([]byte, string, error) {
	if err := goBinary.revalidate(); err != nil {
		return nil, "", err
	}
	executable, err := goBinary.commandPath()
	if err != nil {
		return nil, "", err
	}
	moduleCache, err := verifiedModuleCache(ctx, goBinary)
	if err != nil {
		return nil, "", err
	}
	temporary, err := os.MkdirTemp("/tmp", "a112-mb-us-build-")
	if err != nil {
		return nil, "", err
	}
	if err := os.Chmod(temporary, 0o700); err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(temporary)
	cache := filepath.Join(temporary, "gocache")
	if err := os.Mkdir(cache, 0o700); err != nil {
		return nil, "", err
	}
	output := filepath.Join(temporary, "a112-mb-us-source")
	if err := verifyOfflineModuleClosure(ctx, goBinary, root, cache, moduleCache); err != nil {
		return nil, "", err
	}
	environment, err := prescribedBuildEnvironment(goBinary, cache, moduleCache)
	if err != nil {
		return nil, "", err
	}
	command := exec.CommandContext(ctx, executable, "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", output, "./tools/a112-mb-us-source")
	command.Dir = root
	command.Env = environment
	if err := command.Run(); err != nil {
		return nil, "", err
	}
	if err := goBinary.revalidate(); err != nil {
		return nil, "", err
	}
	data, err := readRegularNoFollow(output)
	if err != nil {
		return nil, "", err
	}
	return data, prescribedBuildCommand, nil
}

func verifyOfflineModuleClosure(ctx context.Context, goBinary goBinaryIdentity, root, cache, moduleCache string) error {
	if err := goBinary.revalidate(); err != nil {
		return err
	}
	executable, err := goBinary.commandPath()
	if err != nil {
		return err
	}
	environment, err := prescribedBuildEnvironment(goBinary, cache, moduleCache)
	if err != nil {
		return err
	}
	verify := exec.CommandContext(ctx, executable, "mod", "verify")
	verify.Dir = root
	verify.Env = environment
	if err := verify.Run(); err != nil {
		return err
	}
	if err := goBinary.revalidate(); err != nil {
		return err
	}
	command := exec.CommandContext(ctx, executable, "list", "-mod=readonly", "-deps", "./tools/a112-mb-us-source")
	command.Dir = root
	command.Env = environment
	if err := command.Run(); err != nil {
		return err
	}
	return goBinary.revalidate()
}

type listedPackage struct {
	Dir                           string
	GoFiles, CgoFiles, EmbedFiles []string
}

func compiledSourceClosure(ctx context.Context, root string, goBinary goBinaryIdentity, targetGOOS string) (map[string]string, error) {
	git, err := verifiedGitBinary()
	if err != nil {
		return nil, err
	}
	defer func() { _ = git.close() }()
	return compiledSourceClosureWith(ctx, root, goBinary, git, targetGOOS)
}

func compiledSourceClosureWith(ctx context.Context, root string, goBinary goBinaryIdentity, git gitBinaryIdentity, targetGOOS string) (map[string]string, error) {
	moduleCache, err := verifiedModuleCache(ctx, goBinary)
	if err != nil {
		return nil, err
	}
	temporary, err := os.MkdirTemp("/tmp", "a112-mb-us-closure-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temporary)
	if err := os.Chmod(temporary, 0o700); err != nil {
		return nil, err
	}
	cache := filepath.Join(temporary, "gocache")
	if err := os.Mkdir(cache, 0o700); err != nil {
		return nil, err
	}
	if err := verifyOfflineModuleClosure(ctx, goBinary, root, cache, moduleCache); err != nil {
		return nil, err
	}
	if err := goBinary.revalidate(); err != nil {
		return nil, err
	}
	executable, err := goBinary.commandPath()
	if err != nil {
		return nil, err
	}
	environment, err := prescribedBuildEnvironment(goBinary, cache, moduleCache)
	if err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, executable, "list", "-mod=readonly", "-json", "-deps", "./tools/a112-mb-us-source")
	command.Dir = root
	command.Env = environment
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	if err := goBinary.revalidate(); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	out := map[string]string{}
	for {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if !withinRoot(root, pkg.Dir) {
			continue
		}
		for _, name := range append(append(pkg.GoFiles, pkg.CgoFiles...), pkg.EmbedFiles...) {
			path := filepath.Join(pkg.Dir, name)
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil, err
			}
			if err := validateCompiledInputForGOOSWith(root, filepath.ToSlash(rel), targetGOOS, git); err != nil {
				return nil, err
			}
			data, err := readSourceNoFollow(root, filepath.ToSlash(rel))
			if err != nil {
				return nil, err
			}
			out[filepath.ToSlash(rel)] = digestBytes(data)
		}
	}
	return out, nil
}

func withinRoot(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func validateCompiledInput(path string) error {
	root, err := canonicalRepoRoot()
	if err != nil {
		return err
	}
	return validateCompiledInputAt(root, path)
}

func validateCompiledInputAt(root, path string) error {
	return validateCompiledInputForGOOSAt(root, path, "linux")
}

func validateCompiledInputForGOOSAt(root, path, targetGOOS string) error {
	git, err := verifiedGitBinary()
	if err != nil {
		return err
	}
	defer func() { _ = git.close() }()
	return validateCompiledInputForGOOSWith(root, path, targetGOOS, git)
}

// Tracked files are deliberately accepted here: the enclosing identity binds
// their frozen base, exact HEAD diff, and each compiled input's SHA-256.
// Only untracked production inputs are restricted to the exact GOOS allowlist.
func validateCompiledInputForGOOSWith(root, path, targetGOOS string, git gitBinaryIdentity) error {
	if _, err := approvedSourcePath(root, path); err != nil {
		return err
	}
	if runGitWith(root, git, "check-ignore", "-q", "--", path) == nil {
		return errors.New("compiled input is ignored")
	}
	if runGitWith(root, git, "ls-files", "--error-unmatch", "--", path) == nil {
		return nil
	}
	if _, allowed := frozenProductionInputs(targetGOOS)[path]; allowed {
		return nil
	}
	return errors.New("compiled input is not approved")
}

func frozenProductionInputs(targetGOOS string) map[string]struct{} {
	paths := []string{"internal/official/a112_mbus_read.go", "tools/a112-mb-us-source/cli.go", "tools/a112-mb-us-source/execution_snapshot.go", "tools/a112-mb-us-source/go_distribution_snapshot.go", "tools/a112-mb-us-source/identity.go", "tools/a112-mb-us-source/main.go", "tools/a112-mb-us-source/official_reader.go"}
	if targetGOOS == "windows" {
		paths = append(paths, "internal/official/a112_mbus_read_unsupported.go", "tools/a112-mb-us-source/execution_snapshot_unsupported.go", "tools/a112-mb-us-source/go_distribution_snapshot_unsupported.go", "tools/a112-mb-us-source/identity_read_unsupported.go", "tools/a112-mb-us-source/receipt_unsupported.go", "tools/a112-mb-us-source/signals_unsupported.go")
	} else {
		paths = append(paths, "internal/official/a112_mbus_read_unix.go", "tools/a112-mb-us-source/execution_snapshot_unix.go", "tools/a112-mb-us-source/go_distribution_snapshot_unix.go", "tools/a112-mb-us-source/identity_read_unix.go", "tools/a112-mb-us-source/receipt_unix.go", "tools/a112-mb-us-source/signals_unix.go")
	}
	out := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		out[path] = struct{}{}
	}
	return out
}

func frozenInputDigest(targetGOOS string) string {
	paths := make([]string, 0, len(frozenProductionInputs(targetGOOS)))
	for path := range frozenProductionInputs(targetGOOS) {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return digestBytes([]byte(targetGOOS + "\x00" + strings.Join(paths, "\x00")))
}

func verifiedGoToolchain(path string) (goBinaryIdentity, error) {
	return verifiedGoToolchainContext(context.Background(), path)
}

func verifiedGoToolchainContext(ctx context.Context, path string) (goBinaryIdentity, error) {
	snapshot, err := snapshotExecutionBinary(path)
	if err != nil {
		return goBinaryIdentity{}, err
	}
	distribution, err := snapshotGoDistributionContext(ctx, path, snapshot.digest)
	if err != nil {
		_ = snapshot.close()
		return goBinaryIdentity{}, err
	}
	return goBinaryIdentity{Path: path, SHA256: snapshot.digest, execution: snapshot, distribution: distribution}, nil
}

func (g goBinaryIdentity) revalidate() error {
	if g.execution == nil {
		if g.Path == "" && g.SHA256 == "" {
			return nil // dependency-injected rebuild tests do not execute Go.
		}
		return errors.New("selected Go execution snapshot is unavailable")
	}
	if g.SHA256 == "" || g.execution.digest != g.SHA256 {
		return errors.New("Go execution snapshot identity drift")
	}
	if _, err := g.execution.commandPath(); err != nil {
		return err
	}
	if g.distribution == nil || g.distribution.digest == "" {
		return errors.New("private Go distribution snapshot is unavailable")
	}
	if _, err := g.distribution.rootPath(); err != nil {
		return err
	}
	return nil
}

func verifiedModuleCache(ctx context.Context, goBinary goBinaryIdentity) (string, error) {
	if err := goBinary.revalidate(); err != nil {
		return "", err
	}
	currentUser, err := user.Current()
	if err != nil || currentUser.HomeDir == "" || !filepath.IsAbs(currentUser.HomeDir) {
		return "", errors.New("current user home is unavailable")
	}
	executable, err := goBinary.commandPath()
	if err != nil {
		return "", err
	}
	environment, err := goBinary.environment("", "", currentUser.HomeDir)
	if err != nil {
		return "", err
	}
	command := exec.CommandContext(ctx, executable, "env", "GOMODCACHE")
	command.Env = environment
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	cache, err := filepath.EvalSymlinks(strings.TrimSpace(string(output)))
	if err != nil || !filepath.IsAbs(cache) {
		return "", errors.New("offline module cache is not resolvable")
	}
	info, err := os.Stat(cache)
	if err != nil || !info.IsDir() {
		return "", errors.New("offline module cache is unavailable")
	}
	if err := goBinary.revalidate(); err != nil {
		return "", err
	}
	return cache, nil
}

func prescribedBuildEnvironment(goBinary goBinaryIdentity, cache, moduleCache string) ([]string, error) {
	return goBinary.environment(cache, moduleCache, "/nonexistent")
}

func selectedGoEnvironmentWithRoot(cache, moduleCache, home, goRoot string) []string {
	return []string{"HOME=" + home, "GOROOT=" + goRoot, "GOENV=off", "GOFLAGS=", "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "GOCACHE=" + cache, "GOMODCACHE=" + moduleCache}
}

func verifyPrescribedBinary(ctx context.Context, snapshot identitySnapshot, builder buildRunner) error {
	if snapshot.RepoRoot == "" || snapshot.ExecutableSHA256 == "" || builder == nil {
		return errors.New("rebuild inputs are unavailable")
	}
	goBinary := snapshot.goExecution
	if goBinary.Path == "" && goBinary.SHA256 == "" && snapshot.GoBinary == "" && snapshot.GoToolchainSHA256 == "" {
		goBinary = goBinaryIdentity{} // in-memory-only test builder; no Go process runs.
	}
	if goBinary.Path != snapshot.GoBinary || goBinary.SHA256 != snapshot.GoToolchainSHA256 {
		return errors.New("selected Go execution snapshot does not bind receipt identity")
	}
	if goBinary.Path != "" && (goBinary.distribution == nil || goBinary.distribution.digest != snapshot.GoDistributionSHA256) {
		return errors.New("private Go distribution snapshot does not bind receipt identity")
	}
	if goBinary.Path != "" && (goBinary.SHA256 == "" || goBinary.revalidate() != nil) {
		return errors.New("selected Go binary changed before prescribed rebuild")
	}
	rebuilt, command, err := builder.rebuild(ctx, snapshot.RepoRoot, goBinary)
	if err != nil || command != prescribedBuildCommand || digestBytes(rebuilt) != snapshot.ExecutableSHA256 {
		return errors.New("running executable does not match the prescribed rebuild")
	}
	if goBinary.Path != "" && goBinary.revalidate() != nil {
		return errors.New("selected Go binary changed after prescribed rebuild")
	}
	return nil
}

func validateFrozenBase(base string) error {
	root, err := canonicalRepoRoot()
	if err != nil {
		return err
	}
	return validateFrozenBaseAt(root, base)
}

func validateFrozenBaseAt(root, base string) error {
	git, err := verifiedGitBinary()
	if err != nil {
		return err
	}
	defer func() { _ = git.close() }()
	return validateFrozenBaseWith(root, base, git)
}

func validateFrozenBaseWith(root, base string, git gitBinaryIdentity) error {
	if len(base) != 40 {
		return errors.New("frozen base is not a full commit hash")
	}
	for _, char := range base {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return errors.New("frozen base is not lowercase hexadecimal")
		}
	}
	if _, err := gitOutputBytesWith(root, git, "rev-parse", "--verify", base+"^{commit}"); err != nil {
		return errors.New("frozen base commit does not exist")
	}
	if err := runGitWith(root, git, "merge-base", "--is-ancestor", base, "HEAD"); err != nil {
		return errors.New("frozen base is not an ancestor of HEAD")
	}
	return nil
}

func (s systemIdentity) snapshot(cfg config) (identitySnapshot, error) {
	return s.snapshotContext(context.Background(), cfg)
}

func (s systemIdentity) snapshotContext(ctx context.Context, cfg config) (identitySnapshot, error) {
	root, err := canonicalRepoRoot()
	if err != nil {
		return identitySnapshot{}, err
	}
	if cwd, cwdErr := os.Getwd(); cwdErr != nil || filepath.Clean(cwd) != root {
		return identitySnapshot{}, errors.New("collector must execute from the canonical repository root")
	}
	executablePath := os.Executable
	if s.executable != nil {
		executablePath = s.executable
	}
	executable, err := executablePath()
	if err != nil || executable == "" || strings.Contains(filepath.Clean(executable), string(filepath.Separator)+"go-build"+string(filepath.Separator)) {
		return identitySnapshot{}, errors.New("ad-hoc executable cannot attest identity")
	}
	executableBytes, err := readRegularNoFollow(executable)
	if err != nil {
		return identitySnapshot{}, err
	}
	goBinary, err := verifiedGoToolchainContext(ctx, cfg.goBinary)
	if err != nil {
		return identitySnapshot{}, err
	}
	keepGoExecution := false
	defer func() {
		if !keepGoExecution {
			_ = goBinary.close()
		}
	}()
	gitBinary, err := verifiedGitBinary()
	if err != nil {
		return identitySnapshot{}, err
	}
	keepGitExecution := false
	defer func() {
		if !keepGitExecution {
			_ = gitBinary.close()
		}
	}()
	goVersion, err := selectedGoEnv(ctx, goBinary, "GOVERSION")
	if err != nil || goVersion == "" {
		return identitySnapshot{}, errors.New("selected Go binary did not provide GOVERSION")
	}
	targetGOOS, err := selectedGoEnv(ctx, goBinary, "GOOS")
	if err != nil || targetGOOS == "" {
		return identitySnapshot{}, errors.New("selected Go binary did not provide GOOS")
	}
	base, err := readSourceSingleLine(root, "openspec/changes/a112-run-four-strategy-families-independently/base-commit.txt")
	if err != nil || base == "" {
		return identitySnapshot{}, errors.New("frozen A112 base commit is unavailable")
	}
	if err := validateFrozenBaseWith(root, base, gitBinary); err != nil {
		return identitySnapshot{}, err
	}
	headBytes, err := gitOutputBytesWith(root, gitBinary, "rev-parse", "HEAD")
	if err != nil {
		return identitySnapshot{}, err
	}
	head := strings.TrimSpace(string(headBytes))
	diff, err := gitOutputBytesWith(root, gitBinary, "diff", "--no-ext-diff", "--binary", "HEAD")
	if err != nil {
		return identitySnapshot{}, err
	}
	untracked, err := gitOutputBytesWith(root, gitBinary, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return identitySnapshot{}, err
	}
	sources, err := measurementSourceDigests(root)
	if err != nil {
		return identitySnapshot{}, err
	}
	compiled, err := compiledSourceClosureWith(ctx, root, goBinary, gitBinary, targetGOOS)
	if err != nil {
		return identitySnapshot{}, err
	}
	set := splitNULPaths(untracked)
	sort.Strings(set)
	content, err := untrackedContentDigests(root, set)
	if err != nil {
		return identitySnapshot{}, err
	}
	snapshot := identitySnapshot{BaseCommit: base, RepoRoot: root, BuildCommand: prescribedBuildCommand, ExecutableSHA256: digestBytes(executableBytes), GoBinary: goBinary.Path, GoVersion: goVersion, GoToolchainSHA256: goBinary.SHA256, GoDistributionSHA256: goBinary.distribution.digest, GitBinarySHA256: gitBinary.SHA256, TargetGOOS: targetGOOS, FrozenInputSHA256: frozenInputDigest(targetGOOS),
		SourceSHA256: sources, CompiledSourceSHA256: compiled, TrackedDiffSHA256: digestBytes(diff), Untracked: set, UntrackedContent: content, UntrackedSHA256: digestBytes([]byte(strings.Join(flattenDigests(content), "\x00"))), WorktreeHEAD: head, goExecution: goBinary, gitExecution: gitBinary}
	keepGoExecution = true
	keepGitExecution = true
	return snapshot, nil
}

func canonicalRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := gitOutputAt(cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(root)
}

func selectedGoEnv(ctx context.Context, goBinary goBinaryIdentity, key string) (string, error) {
	if err := goBinary.revalidate(); err != nil {
		return "", err
	}
	currentUser, err := user.Current()
	if err != nil || currentUser.HomeDir == "" || !filepath.IsAbs(currentUser.HomeDir) {
		return "", errors.New("current user home is unavailable")
	}
	executable, err := goBinary.commandPath()
	if err != nil {
		return "", err
	}
	environment, err := goBinary.environment("", "", currentUser.HomeDir)
	if err != nil {
		return "", err
	}
	command := exec.CommandContext(ctx, executable, "env", key)
	command.Env = environment
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	if err := goBinary.revalidate(); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func readSourceSingleLine(root, path string) (string, error) {
	data, err := readSourceNoFollow(root, path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func approvedSourcePath(root, path string) (string, error) {
	if root == "" || !filepath.IsAbs(root) || path == "" || filepath.IsAbs(path) {
		return "", errors.New("source path is not an approved root-relative path")
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("source path escapes canonical repository root")
	}
	full := filepath.Join(root, clean)
	if !withinRoot(root, full) {
		return "", errors.New("source path escapes canonical repository root")
	}
	return full, nil
}

func readSourceNoFollow(root, path string) ([]byte, error) {
	full, err := approvedSourcePath(root, path)
	if err != nil {
		return nil, err
	}
	return readRegularNoFollow(full)
}

func gitOutputAt(root string, args ...string) (string, error) {
	output, err := gitOutputBytesAt(root, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func gitOutputBytesAt(root string, args ...string) ([]byte, error) {
	git, err := verifiedGitBinary()
	if err != nil {
		return nil, err
	}
	defer func() { _ = git.close() }()
	return gitOutputBytesWith(root, git, args...)
}

func gitOutputBytesWith(root string, git gitBinaryIdentity, args ...string) ([]byte, error) {
	if err := git.revalidate(); err != nil {
		return nil, err
	}
	executable, err := git.commandPath()
	if err != nil {
		return nil, err
	}
	command := exec.Command(executable, args...)
	command.Dir = root
	command.Env = sanitizedGitEnvironment()
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	if err := git.revalidate(); err != nil {
		return nil, err
	}
	return output, nil
}

func runGit(root string, args ...string) error {
	git, err := verifiedGitBinary()
	if err != nil {
		return err
	}
	defer func() { _ = git.close() }()
	return runGitWith(root, git, args...)
}

func runGitWith(root string, git gitBinaryIdentity, args ...string) error {
	if err := git.revalidate(); err != nil {
		return err
	}
	executable, err := git.commandPath()
	if err != nil {
		return err
	}
	command := exec.Command(executable, args...)
	command.Dir = root
	command.Env = sanitizedGitEnvironment()
	if err := command.Run(); err != nil {
		return err
	}
	return git.revalidate()
}

func verifiedGitBinary() (gitBinaryIdentity, error) {
	for _, candidate := range []string{"/usr/bin/git", "/bin/git"} {
		snapshot, err := snapshotExecutionBinary(candidate)
		if err != nil {
			continue
		}
		return gitBinaryIdentity{Path: candidate, SHA256: snapshot.digest, execution: snapshot}, nil
	}
	return gitBinaryIdentity{}, errors.New("vetted absolute git binary is unavailable")
}

func (g gitBinaryIdentity) revalidate() error {
	if g.execution == nil || g.SHA256 == "" || g.execution.digest != g.SHA256 {
		return errors.New("Git execution snapshot is unavailable or drifted")
	}
	_, err := g.execution.commandPath()
	return err
}

func sanitizedGitEnvironment() []string {
	return []string{"HOME=/nonexistent", "PATH=/usr/bin:/bin", "LC_ALL=C", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null", "GIT_OPTIONAL_LOCKS=0"}
}

func splitNULPaths(data []byte) []string {
	parts := bytes.Split(data, []byte{0})
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) != 0 {
			paths = append(paths, string(part))
		}
	}
	return paths
}

func measurementSourceDigests(root string) (map[string]string, error) {
	paths := []string{"go.mod", "go.sum", "openspec/changes/a112-run-four-strategy-families-independently/base-commit.txt", "internal/official/a112_mbus_read.go", "internal/official/a112_mbus_read_test.go", "internal/official/a112_mbus_read_unix.go", "internal/official/a112_mbus_read_unsupported.go", "internal/official/a112_mbus_static_test.go"}
	err := filepath.WalkDir(filepath.Join(root, "tools/a112-mb-us-source"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".go") {
			relative, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			paths = append(paths, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	digests := make(map[string]string, len(paths))
	for _, path := range paths {
		data, err := readSourceNoFollow(root, path)
		if err != nil {
			return nil, err
		}
		digests[path] = digestBytes(data)
	}
	return digests, nil
}

func untrackedContentDigests(root string, paths []string) (map[string]string, error) {
	out := make(map[string]string, len(paths))
	for _, path := range paths {
		data, err := readSourceNoFollow(root, path)
		if err != nil {
			return nil, err
		}
		out[path] = digestBytes(data)
	}
	return out, nil
}

func flattenDigests(digests map[string]string) []string {
	paths := make([]string, 0, len(digests))
	for path := range digests {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for index, path := range paths {
		paths[index] = path + "=" + digests[path]
	}
	return paths
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
