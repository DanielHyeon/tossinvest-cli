//go:build linux

package accountbudgetruntime

import (
	"bytes"
	"context"
	"errors"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestA166OfflineProvisionAndCheck(t *testing.T) {
	base := t.TempDir()
	if err := os.Chmod(base, 0700); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "root")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(base, "manifest.json")
	if err := ProvisionOffline(context.Background(), root, manifest, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	state, err := os.ReadFile(filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	events, err := os.ReadFile(filepath.Join(root, "event.state"))
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckOffline(context.Background(), manifest); err != nil {
		t.Fatal(err)
	}
	stateAfter, err := os.ReadFile(filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	eventsAfter, err := os.ReadFile(filepath.Join(root, "event.state"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(state, stateAfter) || !bytes.Equal(events, eventsAfter) {
		t.Fatal("check changed state")
	}
	m, err := readManifest(manifest)
	if err != nil || m.Version != 2 || m.EventLock.Device == 0 || m.EventLock.Inode == 0 {
		t.Fatalf("manifest roundtrip %+v %v", m, err)
	}
	if err = ProvisionOffline(context.Background(), root, manifest, "scope", "instance"); err == nil {
		t.Fatal("re-provision accepted")
	}
}

func offlineFixture(t *testing.T) (string, string, string) {
	t.Helper()
	base := t.TempDir()
	if err := os.Chmod(base, 0700); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "root")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	return base, root, filepath.Join(base, "manifest")
}
func TestA166OfflineRefusesExistingAndNonempty(t *testing.T) {
	for _, mode := range []string{"manifest", "root", "inside", "root-symlink", "parent-symlink", "parent-mode", "root-mode"} {
		t.Run(mode, func(t *testing.T) {
			base, root, path := offlineFixture(t)
			switch mode {
			case "manifest":
				if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
			case "root":
				if err := os.WriteFile(filepath.Join(root, "keep"), []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
			case "inside":
				path = filepath.Join(root, "manifest")
			case "root-symlink":
				alias := filepath.Join(base, "alias")
				if err := os.Symlink(root, alias); err != nil {
					t.Fatal(err)
				}
				root = alias
			case "parent-symlink":
				alias := filepath.Join(base, "alias")
				if err := os.Symlink(base, alias); err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(alias, "manifest")
			case "parent-mode":
				if err := os.Chmod(base, 0755); err != nil {
					t.Fatal(err)
				}
			case "root-mode":
				if err := os.Chmod(root, 0755); err != nil {
					t.Fatal(err)
				}
			}
			if err := ProvisionOffline(context.Background(), root, path, "scope", "instance"); err == nil {
				t.Fatal("unsafe input accepted")
			}
			if mode == "manifest" {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != "keep" {
					t.Fatal("destination changed")
				}
			}
			if mode == "root" {
				got, err := os.ReadFile(filepath.Join(root, "keep"))
				if err != nil || string(got) != "keep" {
					t.Fatal("root changed")
				}
			}
		})
	}
}
func TestA166OfflineFaultBoundaries(t *testing.T) {
	for _, stage := range []string{"write", "file-sync", "close", "rename", "dir-sync"} {
		t.Run(stage, func(t *testing.T) {
			base, root, path := offlineFixture(t)
			sentinel := errors.New("fault")
			err := provisionOffline(context.Background(), root, path, "scope", "instance", func(s string) error {
				if s == stage {
					return sentinel
				}
				return nil
			})
			if !errors.Is(err, sentinel) {
				t.Fatalf("%s: %v", stage, err)
			}
			_, statErr := os.Stat(path)
			if stage == "dir-sync" {
				if statErr != nil {
					t.Fatal("published file missing")
				}
			} else if !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("unexpected manifest %v", statErr)
			}
			entries, err := os.ReadDir(base)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".a166-") {
					t.Fatal("temp file leaked")
				}
			}
			if err = ProvisionOffline(context.Background(), root, path, "scope", "instance"); err == nil {
				t.Fatal("partial root reused")
			}
		})
	}
}
func TestA166OfflineRootReplacementAndCancel(t *testing.T) {
	for _, stage := range []string{"after-provision", "before-final-root"} {
		t.Run(stage, func(t *testing.T) {
			_, root, path := offlineFixture(t)
			err := provisionOffline(context.Background(), root, path, "scope", "instance", func(s string) error {
				if s != stage {
					return nil
				}
				if err := os.Rename(root, root+"-old"); err != nil {
					return err
				}
				return os.Mkdir(root, 0700)
			})
			if err == nil {
				t.Fatal("root replacement accepted")
			}
			if _, err = os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("manifest published")
			}
		})
	}
	_, root, path := offlineFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	err := provisionOffline(ctx, root, path, "scope", "instance", func(s string) error {
		if s == "after-provision" {
			cancel()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel %v", err)
	}
	if _, err = os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancel published manifest")
	}
}
func TestA166OfflineRenameDoesNotReplace(t *testing.T) {
	_, root, path := offlineFixture(t)
	err := provisionOffline(context.Background(), root, path, "scope", "instance", func(stage string) error {
		if stage == "rename" {
			return os.WriteFile(path, []byte("other"), 0600)
		}
		return nil
	})
	if err == nil {
		t.Fatal("rename replaced destination")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "other" {
		t.Fatal("destination changed")
	}
}
func TestA166OfflineCheckFaultAndDirty(t *testing.T) {
	_, root, path := offlineFixture(t)
	if err := ProvisionOffline(context.Background(), root, path, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("clock failure")
	if err := checkOffline(context.Background(), path, func() (clock.BootStamp, error) { return clock.BootStamp{}, sentinel }); !errors.Is(err, sentinel) {
		t.Fatal("clock failure lost")
	}
	if err := os.WriteFile(filepath.Join(root, "event.dirty"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := CheckOffline(context.Background(), path); err == nil {
		t.Fatal("dirty event accepted")
	}
}

func TestA166OfflineCheckRejectsV1MissingAndCopiedEvent(t *testing.T) {
	root, path, m := setupGroup(t)
	for _, change := range []func(*manifest){func(x *manifest) { x.Version = 1 }, func(x *manifest) { x.EventLock = accountbudget.FileID{} }, func(x *manifest) { x.EventLock.Inode++ }} {
		bad := m
		change(&bad)
		writeManifest(t, path, bad)
		if err := CheckOffline(context.Background(), path); err == nil {
			t.Fatal("invalid manifest accepted")
		}
	}
	writeManifest(t, path, m)
	if err := os.Rename(filepath.Join(root, "event.lock"), filepath.Join(root, "old.lock")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "event.lock"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := CheckOffline(context.Background(), path); err == nil {
		t.Fatal("replaced event lock accepted")
	}
}
func TestA166OfflineCanceledBeforeProvision(t *testing.T) {
	_, root, path := offlineFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ProvisionOffline(ctx, root, path, "scope", "instance"); !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("root changed: %v", err)
	}
}
func TestA166OfflineNoClientOrHTTPCalls(t *testing.T) {
	for _, item := range []struct {
		file  string
		funcs map[string]bool
	}{{"offline_linux.go", map[string]bool{"ProvisionOffline": true, "provisionOffline": true, "CheckOffline": true, "checkOffline": true}}, {"runtime_linux.go", map[string]bool{"Open": true}}} {
		tree, err := parser.ParseFile(token.NewFileSet(), item.file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range tree.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if item.file == "offline_linux.go" && (path == "net/http" || strings.HasSuffix(path, "/official")) {
				t.Fatalf("forbidden import %s", path)
			}
		}
		for _, decl := range tree.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !item.funcs[fn.Name.Name] {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if ok && (sel.Sel.Name == "NewClient" || sel.Sel.Name == "NewWithAccountBudget" || sel.Sel.Name == "Do" || sel.Sel.Name == "RoundTrip") {
					t.Errorf("forbidden call %s", sel.Sel.Name)
				}
				return true
			})
		}
	}
}

func TestA166OfflineOwnerAndFilesystemValidation(t *testing.T) {
	st := unix.Stat_t{Uid: 123, Mode: unix.S_IFDIR | 0700}
	if !safeOfflineDirectory(st, 123) || safeOfflineDirectory(st, 124) {
		t.Fatal("directory owner validation")
	}
	st.Mode = unix.S_IFDIR | 0755
	if safeOfflineDirectory(st, 123) {
		t.Fatal("directory mode accepted")
	}
	if !supportedOfflineFS(unix.EXT4_SUPER_MAGIC) || !supportedOfflineFS(unix.TMPFS_MAGIC) || supportedOfflineFS(unix.NFS_SUPER_MAGIC) {
		t.Fatal("filesystem support boundary")
	}
}

func TestA166OfflineInvalidScopeLeavesRootEmpty(t *testing.T) {
	for _, pair := range [][2]string{{"bad/scope", "instance"}, {"scope", "bad instance"}, {strings.Repeat("a", 129), "instance"}} {
		_, root, path := offlineFixture(t)
		if err := ProvisionOffline(context.Background(), root, path, pair[0], pair[1]); err == nil {
			t.Fatal("invalid scope accepted")
		}
		entries, err := os.ReadDir(root)
		if err != nil || len(entries) != 0 {
			t.Fatalf("invalid scope mutated root: %v", err)
		}
	}
}
