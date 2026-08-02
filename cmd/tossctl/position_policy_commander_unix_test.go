//go:build unix

package main

import (
	"context"
	"os"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

type runtimeValueReader struct {
	value positionpolicy.ManagementRuntime
}

func (r runtimeValueReader) Runtime(context.Context) (positionpolicy.ManagementRuntime, error) {
	return r.value, nil
}

func TestPositionPolicyRuntimeDescriptorReaderRecoversAfterLateStartAndRestart(t *testing.T) {
	engineDir, err := os.MkdirTemp("/tmp", "runtime-descriptor-reader-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(engineDir) })
	if err := os.Chmod(engineDir, 0o700); err != nil {
		t.Fatal(err)
	}

	reader := positionPolicyRuntimeDescriptorReader{
		descriptorPath: positionpolicyrpc.RuntimeDescriptorPath(engineDir),
	}
	if _, err := reader.Runtime(t.Context()); err == nil {
		t.Fatal("Runtime() before engine start error=nil")
	}

	first := positionpolicy.ManagementRuntime{
		EffectiveKnown: true,
		Effective:      positionpolicy.NewAdoptionSettings(true, .03, []string{"IONQ"}, nil, ""),
	}
	server, err := engine.StartPositionPolicyRuntimeServer(engineDir, runtimeValueReader{value: first})
	if err != nil {
		t.Fatal(err)
	}
	got, err := reader.Runtime(t.Context())
	if err != nil || len(got.Effective.IncludeSymbols) != 1 || got.Effective.IncludeSymbols[0] != "IONQ" {
		t.Fatalf("late-start runtime=%+v err=%v", got, err)
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.Runtime(t.Context()); err == nil {
		t.Fatal("Runtime() between engine instances error=nil")
	}

	second := positionpolicy.ManagementRuntime{
		EffectiveKnown: true,
		Effective:      positionpolicy.NewAdoptionSettings(true, .05, []string{"TSLA"}, nil, ""),
	}
	restarted, err := engine.StartPositionPolicyRuntimeServer(engineDir, runtimeValueReader{value: second})
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	got, err = reader.Runtime(t.Context())
	if err != nil || len(got.Effective.IncludeSymbols) != 1 || got.Effective.IncludeSymbols[0] != "TSLA" {
		t.Fatalf("restarted runtime=%+v err=%v", got, err)
	}
}
