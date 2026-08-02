package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

type fakePositionPolicyLifecycle struct{}

func (fakePositionPolicyLifecycle) List(context.Context) ([]positionpolicy.State, error) {
	return []positionpolicy.State{{PositionID: "pos-1"}}, nil
}

func TestPositionPolicyRuntimeDescriptorReaderFailsClosedWhenEngineIsLate(t *testing.T) {
	reader := positionPolicyRuntimeDescriptorReader{
		descriptorPath: filepath.Join(t.TempDir(), "runtime.json"),
	}
	if _, err := reader.Runtime(context.Background()); err == nil {
		t.Fatal("Runtime() error=nil, want unavailable descriptor error")
	}
}
func (fakePositionPolicyLifecycle) Preview(context.Context,
	positionpolicy.Request) (positionpolicy.Preview, error) {
	return positionpolicy.Preview{}, nil
}
func (fakePositionPolicyLifecycle) Apply(context.Context,
	positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	return positionpolicy.State{}, nil
}

type fakePositionPolicyRuntime struct {
	value positionpolicy.ManagementRuntime
}

func (f fakePositionPolicyRuntime) Runtime(context.Context) (positionpolicy.ManagementRuntime, error) {
	return f.value, nil
}

func TestConsolePositionPolicyCommanderKeepsRuntimeAuthoritySeparate(t *testing.T) {
	want := positionpolicy.ManagementRuntime{EffectiveKnown: true}
	commander := &consolePositionPolicyCommander{
		lifecycle: fakePositionPolicyLifecycle{}, runtime: fakePositionPolicyRuntime{value: want},
	}
	if got, err := commander.Runtime(context.Background()); err != nil || !got.EffectiveKnown {
		t.Fatalf("runtime=%+v err=%v", got, err)
	}
	withoutRuntime := &consolePositionPolicyCommander{lifecycle: fakePositionPolicyLifecycle{}}
	if _, err := withoutRuntime.Runtime(context.Background()); !errors.Is(err, errPositionPolicyRuntimeUnavailable) {
		t.Fatalf("runtime error=%v", err)
	}
}
