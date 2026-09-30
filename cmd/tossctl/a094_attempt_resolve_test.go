package main

// a094 4.T — tossctl engine attempt-resolve 의 표면: mutating: true · 필수 입력 · 엔진 없으면 거절 · 거절 어휘 · 발의 해제 보고.

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attemptthaw"
)

type a094ThawClient struct {
	got    attemptthaw.Request
	result attemptthaw.Result
	err    error
}

func (c *a094ThawClient) ResolveParkedAttempt(_ context.Context, req attemptthaw.Request) (attemptthaw.Result, error) {
	c.got = req
	return c.result, c.err
}

func a094ThawCmd(client attemptThawClient, dialErr error) (*bytes.Buffer, func(args ...string) error) {
	out := &bytes.Buffer{}
	deps := attemptThawDeps{
		engineDir: func(*rootOptions) (string, error) { return "/nonexistent", nil },
		dial: func(context.Context, string) (attemptThawClient, error) {
			if dialErr != nil {
				return nil, dialErr
			}
			return client, nil
		},
	}
	return out, func(args ...string) error {
		cmd := newEngineAttemptResolveCmdWithDeps(&rootOptions{}, deps)
		cmd.SetOut(out)
		cmd.SetErr(out)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
}

var a094ThawArgs = []string{"--attempt", "a-1", "--target", "FAILED_CONFIRMED", "--operator", "op", "--approval", "OPS-1", "--note", "checked"}

func TestA094TheAttemptResolveCommandIsMutating(t *testing.T) {
	cmd := newEngineAttemptResolveCmd(&rootOptions{})
	if cmd.Annotations["mutating"] != "true" {
		t.Fatalf("annotations = %v — an interactive agent must not run this by itself", cmd.Annotations)
	}
	for _, flag := range []string{"attempt", "target", "operator", "approval", "note"} {
		f := cmd.Flags().Lookup(flag)
		if f == nil || f.Annotations["cobra_annotation_bash_completion_one_required_flag"] == nil {
			t.Errorf("--%s is not a required flag", flag)
		}
	}
}

func TestA094TheAttemptResolveCommandReportsTheRelease(t *testing.T) {
	client := &a094ThawClient{result: attemptthaw.Result{AttemptID: "a-1", State: "FAILED_CONFIRMED", PositionID: "pos-1", ProposalReleased: true}}
	out, run := a094ThawCmd(client, nil)
	if err := run(a094ThawArgs...); err != nil {
		t.Fatalf("run: %v", err)
	}
	if client.got.Approval != "OPS-1" || client.got.Operator != "op" || client.got.AttemptID != "a-1" {
		t.Errorf("request = %+v", client.got)
	}
	if !strings.Contains(out.String(), "released") {
		t.Errorf("output = %q, want the release reported", out.String())
	}
}

func TestA094TheAttemptResolveCommandRefusesWithoutAnEngine(t *testing.T) {
	_, run := a094ThawCmd(nil, errors.New("no descriptor"))
	if err := run(a094ThawArgs...); err == nil || !strings.Contains(err.Error(), "nothing was changed") {
		t.Fatalf("err = %v, want a refusal that says nothing changed", err)
	}
	client := &a094ThawClient{err: attemptthaw.ErrStale}
	_, run = a094ThawCmd(client, nil)
	if err := run(a094ThawArgs...); err == nil || !errors.Is(err, attemptthaw.ErrStale) {
		t.Fatalf("stale err = %v", err)
	}
	client = &a094ThawClient{result: attemptthaw.Result{AttemptID: "a-1", State: "FAILED_CONFIRMED", ReleaseError: "disk"}}
	_, run = a094ThawCmd(client, nil)
	if err := run(a094ThawArgs...); err == nil || !strings.Contains(err.Error(), "next engine start releases it") {
		t.Fatalf("release failure err = %v", err)
	}
}

// 4.T — 명령이 실제 바이너리의 `engine` 트리에 붙어 있음(조립만 되고 안 붙은 명령은 없는 명령과 같다 — park 원인 알림 본문이
// 이 명령을 가리킨다).
func TestA094TheAttemptResolveCommandIsAttachedToTheEngineTree(t *testing.T) {
	engine := newEngineCmd(&rootOptions{})
	for _, c := range engine.Commands() {
		if c.Name() == "attempt-resolve" {
			if c.Annotations["mutating"] != "true" {
				t.Fatalf("attached command annotations = %v", c.Annotations)
			}
			return
		}
	}
	t.Fatal("`tossctl engine attempt-resolve` is not attached — the park exit does not exist in the binary")
}
