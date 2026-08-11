package main

// soakautostart.go keeps the read-only capability survey across a container
// replacement (openspec change a101).
//
// # What went wrong
//
// Measured 2026-08-11: `docker compose up -d` recreated the container, and
// engine, console and httpapi came back while the survey did not. Nothing
// restarts it, because nothing records that it should be running — the operator
// pressed a button on a process, and the process is what the deploy replaced.
//
// The attestation the automation gate is interlocked on expires, and the survey
// is the only thing that renews it. So a routine deploy quietly stops the clock
// that keeps unattended trading permitted, with no warning anywhere.
//
// # Why this is not the script a060 retired
//
// A host-side `soak-autostart.sh` used to do this and was retired on 2026-08-03
// because the console had moved into the container. The hazard in it was that it
// spawned its child with no flags, so it could start a *second* survey on the
// default profile. Nothing here can: it calls the console's own restart seam,
// which passes the profile flags (soakArgs) and judges ownership by the record
// the survey appends to (soakFindProcesses) — the ownership rule operator-console
// already requires, and whose text names autostart explicitly.

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

type consoleSoakBoot struct {
	svc *config.Service
}

func newConsoleSoakBoot(root *rootOptions) *consoleSoakBoot {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return &consoleSoakBoot{svc: svc}
}

func (s consoleSoakBoot) Load() (bool, error) {
	return s.svc.LoadSoakAutostart()
}

// Save records the approval and audits it. It mirrors consoleEngineBoot.Save,
// including its ordering note: the config write is already durable and cannot be
// rolled back if the audit append then fails.
func (s consoleSoakBoot) Save(on bool) error {
	before, beforeErr := s.svc.LoadSoakAutostart()
	if err := s.svc.SaveSoakAutostart(on); err != nil {
		return err
	}

	old := "unknown"
	if beforeErr == nil {
		old = boolText(before)
	}
	log := openAuditLog()
	if log == nil {
		return nil
	}
	_ = log.Record(audit.Entry{
		Action:  audit.ActionSoakAutostart,
		Setting: "soak.autostart",
		Old:     old,
		New:     boolText(on),
		Detail:  "operator console, soak 재시작이 남긴 승인",
	})
	return nil
}

// runConfiguredSoakAutostart is the boot decision, shaped like
// runConfiguredEngineAutostart and for the same reason: runConsole is measured at
// 0.0% coverage, so a judgement written inside it is a judgement no test can
// reach. Taking the two seams as arguments puts every branch here instead.
//
// It returns a line for the operator and never an error. The survey is optional
// machinery — the console runs without it, shows its record either way, and can
// start it from a button — so nothing here may become a reason for there to be
// no operator screen.
func runConfiguredSoakAutostart(
	load func() (bool, error),
	start func() (string, error),
) string {
	if load == nil {
		return ""
	}
	on, err := load()
	if err != nil {
		return "soak 자동 시작 안 함 — 설정을 읽을 수 없다: " + err.Error()
	}
	if !on {
		return ""
	}
	if start == nil {
		return "soak 자동 시작 거부 — 이 빌드에는 soak 기동 배선이 없다"
	}

	note, err := start()
	if err != nil {
		result := "soak 자동 시작 실패: " + err.Error()
		if strings.TrimSpace(note) != "" {
			result += "\n" + note
		}
		return result
	}
	if strings.TrimSpace(note) == "" {
		note = "서베이를 시작했다."
	}
	return fmt.Sprintf("soak 자동 시작: %s", note)
}

// rememberSoakApproval persists "this profile runs the survey" after a restart
// that actually produced one.
//
// The failure handling is the whole point of it being a separate function. The
// survey is already running by the time save is called, so a bookkeeping failure
// must not come back as a restart failure: the dashboard would tell the operator
// the button did not work, and pressing it again would interrupt the survey that
// just started. The failure is appended to the note instead, where it is visible
// and harmless.
func rememberSoakApproval(save func(bool) error, note string, err error) (string, error) {
	if err != nil {
		// Nothing is running, so there is nothing to approve.
		return note, err
	}
	if save == nil {
		return note, nil
	}
	if saveErr := save(true); saveErr != nil {
		return note + "\n다음 기동에 자동으로 다시 시작하도록 기록하지는 못했다: " + saveErr.Error(), nil
	}
	return note, nil
}
