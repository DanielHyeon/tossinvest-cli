package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

func TestStrictVerifyActivityFailsClosedForFreshOrUnclassifiableEvidence(t *testing.T) {
	now := time.Date(2026, 7, 30, 7, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	path := filepath.Join(dir, "verify-run.lock")

	if err := strictVerifyActivity(path, now); err != nil {
		t.Fatalf("missing marker should mean idle: %v", err)
	}
	if err := os.WriteFile(path, []byte("active"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, now, now); err != nil {
		t.Fatal(err)
	}
	if err := strictVerifyActivity(path, now); err == nil {
		t.Fatal("fresh verification marker was treated as idle")
	}
	stale := now.Add(-runlock.StaleAfter - time.Second)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}
	if err := strictVerifyActivity(path, now); err != nil {
		t.Fatalf("stale marker should mean idle: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "missing-target"), path); err != nil {
		t.Fatal(err)
	}
	if err := strictVerifyActivity(path, now); err == nil {
		t.Fatal("symlink verification evidence was treated as idle")
	}
}

func TestRunConsoleWiresFixedSystemUpdaterAndBothActivityGuards(t *testing.T) {
	fields := consoleOptionFields(t)
	for _, field := range []string{
		"SystemUpdater",
		"ReleaseDownloader",
		"ReleaseCandidateStager",
		"AcquireUpdateEngineLock",
		"CheckUpdateVerifyActivity",
	} {
		if !fields[field] {
			t.Errorf("console.Options is built without %s", field)
		}
	}

	src := readSource(t, "console.go")
	for _, want := range []string{
		"binstamp.SelfPath()",
		"assembleConsoleSystemUpdate(",
		`filepath.Join(filepath.Dir(cachePath), "sigstore")`,
		"enginelock.Acquire(engineDir)",
		"strictVerifyActivity(verifyRunLockPath(verifyRecord)",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("runConsole does not contain fixed update wiring %q", want)
		}
	}
}

func TestAssembleConsoleSystemUpdateUsesRealTossctlBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the real tossctl command")
	}
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "tossctl")
	build := exec.Command("go", "build", "-o", binary, "./cmd/tossctl")
	build.Dir = repoRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building real tossctl: %v\n%s", err, output)
	}

	updater, downloader, err := assembleConsoleSystemUpdate(
		binary, filepath.Join(t.TempDir(), "sigstore"), "dev")
	if err != nil {
		t.Fatalf("assembling production system update: %v", err)
	}
	if updater == nil || downloader == nil {
		t.Fatalf("assembly = updater %v, downloader %v; both must be wired", updater, downloader)
	}
	view := updater.Inspect()
	if view.Current.Path != binary {
		t.Fatalf("updater current path = %q, want built binary %q", view.Current.Path, binary)
	}
	if view.CandidatePath != binary+".candidate" {
		t.Fatalf("candidate path = %q, want fixed sibling", view.CandidatePath)
	}

	updater, downloader, err = assembleConsoleSystemUpdate(binary, "relative-cache", "dev")
	if err == nil {
		t.Fatal("relative Sigstore cache unexpectedly assembled")
	}
	if updater == nil || downloader != nil {
		t.Fatalf("degraded assembly = updater %v, downloader %v; local review must remain wired",
			updater, downloader)
	}
}

func TestLegacyUpdateWarnsOnlyAfterCheckOnlyReturnAndBeforeMutation(t *testing.T) {
	src := readSource(t, "update.go")
	checkReturn := strings.Index(src,
		"if checkOnly {\n\t\t\t\treturn writeUpdateCheckResult")
	warning := strings.Index(src, "`tossctl update`는 레거시 checksum-only 경로")
	mutation := strings.Index(src, "selfupdate.Run(")
	if checkReturn < 0 || warning < 0 || mutation < 0 {
		t.Fatalf("legacy update control points missing: check=%d warning=%d mutation=%d",
			checkReturn, warning, mutation)
	}
	if !(checkReturn < warning && warning < mutation) {
		t.Fatalf("legacy warning ordering is unsafe: check=%d warning=%d mutation=%d",
			checkReturn, warning, mutation)
	}
}

func TestVerificationEntryPointsRefuseWhileSystemUpdateOwnsEngineExclusion(t *testing.T) {
	testEntry := func(t *testing.T, run func(*rootOptions) error) {
		t.Helper()
		root := &rootOptions{configDir: t.TempDir()}
		updateLock, err := enginelock.Acquire(root.configDir)
		if err != nil {
			t.Fatalf("acquiring update exclusion: %v", err)
		}
		defer updateLock.Release()

		brokerBuilds := 0
		previous := verifyBrokerFactory
		verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
			brokerBuilds++
			return nil, "", errors.New("broker construction must not be reached")
		}
		t.Cleanup(func() { verifyBrokerFactory = previous })

		err = run(root)
		if !errors.Is(err, enginelock.ErrAlreadyRunning) {
			t.Fatalf("verification start error = %v, want engine/update exclusion", err)
		}
		if brokerBuilds != 0 {
			t.Fatalf("verification built its broker %d time(s) while update held exclusion", brokerBuilds)
		}
	}

	t.Run("standalone verify run", func(t *testing.T) {
		testEntry(t, func(root *rootOptions) error {
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			return runVerifyRun(cmd, root, &verifyOptions{market: verifylive.MarketKR})
		})
	})

	t.Run("console verify starter", func(t *testing.T) {
		testEntry(t, func(root *rootOptions) error {
			_, _, err := consoleVerifyStarter(root)(
				context.Background(),
				func(verifylive.Batch) error { return nil },
				io.Discard,
				verifylive.MarketKR,
				nil,
			)
			return err
		})
	})
}
