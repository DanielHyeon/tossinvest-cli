package main

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

type consoleExitPolicySettings struct{ svc *config.Service }

func newExitPolicySettingsSeam(root *rootOptions) *consoleExitPolicySettings {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return &consoleExitPolicySettings{svc: svc}
}

func (s consoleExitPolicySettings) Load() (config.ExitPolicy, error) {
	return s.svc.LoadRawEngineExitPolicy()
}

func (s consoleExitPolicySettings) Save(next config.ExitPolicy) error {
	before, beforeErr := s.svc.LoadRawEngineExitPolicy()
	if err := s.svc.SaveEngineExitPolicy(next); err != nil {
		return err
	}
	old := "unknown"
	if beforeErr == nil {
		old = before.CommonPolicy
	}
	if log := openAuditLog(); log != nil {
		// Best-effort like the other console config seams: the config write is
		// already durable and cannot be rolled back if the audit disk is full.
		_ = log.Record(audit.Entry{
			Action: audit.ActionExitPolicy, Setting: "engine.exit_policy.common_policy",
			Old: old, New: next.CommonPolicy,
			Detail: "operator console, 공통 익절·보호선 정책 저장; 다음 엔진 기동부터 신규 관리 포지션에 적용",
		})
	}
	return nil
}
