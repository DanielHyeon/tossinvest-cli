package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenGenerationFenceRunsAfterOldSoakExitBeforeSpawn(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{found: []int{4242}, aliveFor: map[int]int{4242: 3}}
	f.install(t)
	prepared := false
	spawn := soakSpawnDetached
	soakSpawnDetached = func(binary, log string) error {
		if !prepared {
			t.Fatal("new soak spawned before the token-generation fence")
		}
		return spawn(binary, log)
	}

	if _, err := restartSoak(record, func() error {
		if f.aliveFor[4242] > 0 {
			t.Fatal("token-generation fence ran before the old soak exited")
		}
		prepared = true
		return nil
	}); err != nil {
		t.Fatalf("restartSoak: %v", err)
	}
}

func TestTokenGenerationFenceFailureBlocksSoakSpawn(t *testing.T) {
	record := filepath.Join(t.TempDir(), "capability-soak.jsonl")
	f := &procFakes{}
	f.install(t)

	_, err := restartSoak(record, func() error { return errors.New("token cache busy") })
	if err == nil || !strings.Contains(err.Error(), "token") {
		t.Fatalf("restart error=%v, want token-generation fence failure", err)
	}
	if len(f.spawned) != 0 {
		t.Fatalf("soak spawned after token-generation fence failure: %v", f.spawned)
	}
}
