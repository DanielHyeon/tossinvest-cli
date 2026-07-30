package main

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

type consoleEngineBoot struct {
	svc *config.Service
}

func newConsoleEngineBoot(root *rootOptions) *consoleEngineBoot {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return &consoleEngineBoot{svc: svc}
}

func (s consoleEngineBoot) Load() (bool, error) {
	return s.svc.LoadEngineAutostart()
}

func (s consoleEngineBoot) Save(on bool) error {
	before, beforeErr := s.svc.LoadEngineAutostart()
	if err := s.svc.SaveEngineAutostart(on); err != nil {
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
	// Match the other console setting seams: the config save is already durable
	// and cannot be rolled back if a later append fails. The engine's own startup
	// audit still records what configuration it actually accepted.
	_ = log.Record(audit.Entry{
		Action:  audit.ActionEngineAutostart,
		Setting: "engine.autostart",
		Old:     old,
		New:     boolText(on),
		Detail:  "operator console, 엔진 자동 시작 저장",
	})
	return nil
}

// runConfiguredEngineAutostart is the fail-closed boot decision. It receives
// functions so tests can prove every branch without constructing a broker or
// spawning a process.
func runConfiguredEngineAutostart(
	load func() (bool, error),
	start func() (string, error),
) string {
	if load == nil {
		return ""
	}
	on, err := load()
	if err != nil {
		return "엔진 자동 시작 안 함 — 설정을 읽을 수 없다: " + err.Error()
	}
	if !on {
		return ""
	}
	if start == nil {
		return "엔진 자동 시작 거부 — 이 빌드에는 엔진 기동 배선이 없다"
	}

	note, err := start()
	if err != nil {
		result := "엔진 자동 시작 거부: " + err.Error()
		if strings.TrimSpace(note) != "" {
			result += "\n" + note
		}
		return result
	}
	if strings.TrimSpace(note) == "" {
		note = "엔진을 시작했다."
	}
	return fmt.Sprintf("엔진 자동 시작: %s", note)
}
