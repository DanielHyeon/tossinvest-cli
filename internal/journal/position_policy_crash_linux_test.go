//go:build linux

package journal

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

const (
	crashModePositionPolicyBeforeCommit = "position_policy_before_commit"
	crashModePositionPolicyAfterCommit  = "position_policy_after_commit"
	processModePositionPolicyCAS        = "position_policy_process_cas"
)

func TestPositionPolicyLifecycleAndAuditSurviveSIGKILLAtomically(t *testing.T) {
	switch os.Getenv(crashEnvMode) {
	case crashModePositionPolicyBeforeCommit:
		positionPolicyCrashChild(t, true)
		return
	case crashModePositionPolicyAfterCommit:
		positionPolicyCrashChild(t, false)
		return
	case processModePositionPolicyCAS:
		positionPolicyCASChild(t)
		return
	}
	for _, test := range []struct {
		name, mode string
		version    int64
		events     int
	}{
		{"before_commit", crashModePositionPolicyBeforeCommit, 0, 0},
		{"after_commit", crashModePositionPolicyAfterCommit, 1, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "journal.db")
			runCrashChild(t, "TestPositionPolicyLifecycleAndAuditSurviveSIGKILLAtomically", test.mode, path)
			assertCrashJournalArtifacts(t, path)
			j := openTestJournalAt(t, path)
			state, err := j.PositionPolicy(context.Background(), "p-policy")
			if err != nil || state.Version != test.version {
				t.Fatalf("recovered lifecycle=%+v err=%v", state, err)
			}
			events, err := j.PositionPolicyAudit(context.Background(), "p-policy")
			if err != nil || len(events) != test.events {
				t.Fatalf("recovered audit=%+v err=%v", events, err)
			}
		})
	}
}

func TestPositionPolicyCASAcrossCompetingProcessesHasOneWinner(t *testing.T) {
	if os.Getenv(crashEnvMode) != "" {
		return
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	j := openTestJournalAt(t, path)
	seedPolicyPosition(t, j, false)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	barrier := filepath.Join(dir, "go")
	outputs := []string{filepath.Join(dir, "one"), filepath.Join(dir, "two")}
	commands := make([]*exec.Cmd, 2)
	for index := range commands {
		commands[index] = exec.Command(os.Args[0], "-test.run=^TestPositionPolicyLifecycleAndAuditSurviveSIGKILLAtomically$")
		commands[index].Env = append(os.Environ(), crashEnvMode+"="+processModePositionPolicyCAS,
			crashEnvPath+"="+path, "TOSSOS_POLICY_CAS_BARRIER="+barrier,
			"TOSSOS_POLICY_CAS_RESULT="+outputs[index])
		if err := commands[index].Start(); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(barrier, []byte("go"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatalf("CAS child failed: %v", err)
		}
	}
	wins, stale := 0, 0
	for _, output := range outputs {
		body, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		switch string(body) {
		case "win":
			wins++
		case "stale":
			stale++
		default:
			t.Fatalf("unexpected child result %q", body)
		}
	}
	if wins != 1 || stale != 1 {
		t.Fatalf("process CAS winners=%d stale=%d", wins, stale)
	}
	reopened := openTestJournalAt(t, path)
	events, err := reopened.PositionPolicyAudit(context.Background(), "p-policy")
	if err != nil || len(events) != 1 {
		t.Fatalf("audit events=%d err=%v", len(events), err)
	}
}

func positionPolicyCASChild(t *testing.T) {
	t.Helper()
	barrier, output := os.Getenv("TOSSOS_POLICY_CAS_BARRIER"), os.Getenv("TOSSOS_POLICY_CAS_RESULT")
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(barrier); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("CAS barrier timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
	j := openCrashChildJournal()
	req := policyRequest(positionpolicy.ActionOverride, 0)
	if filepath.Base(output) == "one" {
		req.PolicyID = exitpolicy.CommonLadderBalanced
	} else {
		req.PolicyID = exitpolicy.CommonLadderRunner
	}
	_, err := j.ApplyPositionPolicy(context.Background(), req)
	result := "win"
	if errors.Is(err, positionpolicy.ErrVersionMismatch) {
		result = "stale"
	} else if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte(result), 0o600); err != nil {
		t.Fatal(err)
	}
}

func positionPolicyCrashChild(t *testing.T, beforeCommit bool) {
	t.Helper()
	j := openCrashChildJournal()
	seedPolicyPosition(t, j, false)
	req := policyRequest(positionpolicy.ActionOverride, 0)
	req.PolicyID = exitpolicy.CommonLadderBalanced
	if beforeCommit {
		j.exitWriteHook = func(stage string) error {
			if stage == "position_policy_after_audit" {
				kill()
			}
			return nil
		}
	}
	if _, err := j.ApplyPositionPolicy(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if !beforeCommit {
		kill()
	}
}
