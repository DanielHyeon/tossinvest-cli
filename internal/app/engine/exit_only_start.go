package engine

// exit_only_start.go — a199: 낡은 capability 증명은 진입만 막고 청산은 막지 않는다.
//
// # 왜
//
// 게이트가 켜진 엔진은 attestation 이 만료됐거나 엔진이 쓰는 endpoint 를 덮지 않으면 기동을 통째로 거절했음. 그 거절은 진입뿐
// 아니라 exit 관측(손절 · 익절) · 대사 · 편입 · 체결 감지까지 끔 — 2026-11-23 재부팅 뒤 보유 전부가 손절 없이 남은 사고.
// 실패는 비대칭임: 증명이 낡은 경로로 **진입**하면 손절 없는 새 포지션이 생기지만, 그 경로로 **청산**하다 실패하면 남는 것은
// 기동을 거절했을 때와 같은 상태뿐. 그래서 낡음은 기동 조건이 아니라 진입 허가 조건으로 옮김(보호 미배선 조항과 같은 원리).
//
// # 닫힌 집합(design D1 · D1a · Manager Q-B)
//
// 판정은 classifyAttestation 하나 — 인터록(verifyGate)과 감시(judgeAttestation)가 같이 부름(판정 둘 금지, D7).
//   - Verify 성공 → 지금과 같은 수락. 아래 가드를 타지 않음(유효한 attestation 의 경로는 한 글자도 안 바뀜).
//   - Verify 실패가 만료 · endpoint 누락이고, 형식 번호가 1..이 빌드이며, endpoint 항목이 전부 비지 않았고, (만료면) Verify 와
//     같은 비교로 해석된 계좌를 가리키면 → 낡음(청산 전용).
//   - 그 밖 → 지금처럼 거절. 만료가 계좌 불일치를 가리지 못하게 만료일 때 계좌를 따로 물음(Verify 는 첫 실패만 돌려줌).
//
// # 집행(D4)
//
// 진입 차단은 기존 집행 지점 하나 — 게이트웨이의 진입 검사가 읽는 EntryGate 래치 — 에 새 사유 하나로 검. 래치는 메모리이고
// 매 기동 runInterlock 이 다시 세우므로 재시작으로 풀리지 않고, 파일을 고친 뒤의 기동에는 서지 않음. Verified 의 뜻(모든 조항
// 충족)은 바꾸지 않음 — 청산 쪽 소비자만 RunsExitSide 를 읽음. 놓친 소비자는 엄격한 쪽(Verified)에 남아 진입을 열지 못함.

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// exitOnlyRecovered 는 청산 전용으로 뜬 엔진의 파일이 다시 판정을 통과할 때의 회복 안내임(Manager Q-C) — 다른 차단(알림 미전달 ·
// 운영 모드 · RECONCILE)은 재시작으로 풀리지 않으므로 「attestation 사유의 진입 차단」만 말함.
const exitOnlyRecovered = "증명은 회복됐다 — 엔진을 재시작하면 attestation 사유의 진입 차단이 풀린다"

// 낡음의 종류(감사 · 로그 · 기동 출력의 사유 코드). 원장 · 감사에 그대로 남는 철자라 바꾸지 않음.
const (
	staleExpired   = "expired"
	staleScopeLost = "scope_lost"
)

// attestationStaleness 는 청산 전용으로 내려갈 수 있는 낡음의 사실임.
type attestationStaleness struct {
	expired bool
	missing []string // 필요하지만 덮이지 않은 endpoint(정렬됨) — 만료와 함께 있을 수도 있음
}

// code 는 낡음의 사유 코드임 — 만료와 누락이 같이 있으면 둘 다.
func (s attestationStaleness) code() string {
	var parts []string
	if s.expired {
		parts = append(parts, staleExpired)
	}
	if len(s.missing) > 0 {
		parts = append(parts, staleScopeLost)
	}
	return strings.Join(parts, ",")
}

// classifyAttestation 은 attestation 이 정상 / 청산 전용(낡음) / 거절 중 무엇인지 판정함(인터록 · 감시 공용, 판정 하나).
// 반환: (nil, nil) 정상 · (낡음, nil) 청산 전용 · (nil, 오류) 거절. 오류는 attest 의 sentinel 을 감쌈.
func classifyAttestation(att attest.Attestation, now time.Time, accountRef string, required []string) (*attestationStaleness, error) {
	verr := att.Verify(now, accountRef, required)
	if verr == nil {
		// 유효함 — 지금과 같은 수락. 낡음 가드는 「청산 전용으로 내릴지」에만 걸리므로 여기서 타지 않음(Manager Q-B).
		return nil, nil
	}
	expired := errors.Is(verr, attest.ErrExpired)
	if !expired && !errors.Is(verr, attest.ErrEndpointNotAttested) {
		// 불완전 · 계좌 불일치 등 낡음 밖의 결함 — 지금처럼 거절.
		return nil, verr
	}
	// 아래 셋은 「파일이 온전하고 이 계좌의 것」을 거르는 것뿐임 — 「이 계좌에서 한때 주문 실행(POST)까지 증명됐다」는 파일에
	// 남지 않아 여기서 검사하지 못함(감독 증거가 한 번도 없던 GET 만의 파일도 범위 상실로 청산 전용이 됨 — gstack review 공시).
	// 셋 중 하나라도 아니면 거절.
	if att.FormatVersion < 1 || att.FormatVersion > attest.FormatVersion {
		return nil, fmt.Errorf("%w: format_version %d is not a version this build accepts, so a stale attestation "+
			"cannot stand as the proof an exit-only start rests on (%v)", attest.ErrMalformed, att.FormatVersion, verr)
	}
	for _, e := range att.Endpoints {
		if strings.TrimSpace(e) == "" {
			return nil, fmt.Errorf("%w: it lists an empty endpoint item, so a stale attestation cannot stand as the "+
				"proof an exit-only start rests on (%v)", attest.ErrIncomplete, verr)
		}
	}
	// 만료가 계좌 불일치를 가리지 못하게 — Verify 는 만료(B5)를 계좌(B6)보다 먼저 보고 거기서 멈춤. 누락(B7)이면 계좌는 이미 통과.
	if expired && !att.NamesAccount(accountRef) {
		return nil, fmt.Errorf("%w: it attests %s but the credentials resolve to %s (and it has expired)",
			attest.ErrAccountMismatch, attest.Mask(att.AccountRef), attest.Mask(accountRef))
	}
	return &attestationStaleness{expired: expired, missing: att.MissingEndpoints(required)}, nil
}

// RunsExitSide 는 청산 쪽 루프(대사 · exit 관측 · 그 Guardian · 런타임 조립)를 돌려도 되는지임 — 검증됐거나 청산 전용이면 참.
// 진입 쪽(전략 진입 worker)은 이것이 아니라 Verified 를 읽음.
func (s AutomationStatus) RunsExitSide() bool { return s.Verified || s.ExitOnly }

// exitOnlyStart 는 인터록이 청산 전용 기동을 확정하는 갈래임: Verified · EntryPermitted 는 거짓으로 두고, 진입 래치를 세우고,
// 수락도 거절도 아닌 감사 한 줄과 구조 로그 한 줄을 남김.
func exitOnlyStart(status AutomationStatus, gate config.AutomationGate, log *audit.Log, logger *obs.Logger,
	entry *execgw.EntryGate,
) (AutomationStatus, error) {
	status.Verified = false
	status.EntryPermitted = false
	if entry == nil {
		// 래치가 진입 차단의 유일한 집행이므로, 세울 게이트가 없으면 청산 전용으로 세울 수 없음 — 거절(fail-closed).
		err := fmt.Errorf("%w: %w: the entry gate that carries the exit-only latch is not wired",
			ErrAutomationGateRefused, ErrGatewayRequired)
		_ = refuseStartup(log, gate, err)
		logGateDecision(logger, status, err)
		return status, err
	}
	entry.Block(execgw.ReasonCapabilityAttestationStale, exitOnlyLatchDetail(status))
	if err := log.Record(audit.Entry{
		Action:  audit.ActionGateExitOnly,
		Setting: "engine.automation_gate.enabled",
		Old:     "true",
		New:     "true",
		Detail:  exitOnlyDetail(status),
	}); err != nil {
		return status, fmt.Errorf("engine: recording the exit-only start: %w", err)
	}
	logGateDecision(logger, status, nil)
	return status, nil
}

// exitOnlyReason 은 구조 로그의 reason 칸 — 사유 코드와 낡음의 종류.
func exitOnlyReason(status AutomationStatus) string {
	return string(execgw.ReasonCapabilityAttestationStale) + ": " + status.AttestationStale
}

// exitOnlyDetail 은 감사 Detail — 사유 코드 · 낡음 종류 · 빠진 endpoint 이름 · 만료 시각. 계좌 · 경로 · 원문 오류는 싣지 않음(D8).
func exitOnlyDetail(status AutomationStatus) string {
	missing := status.AttestationMissing
	if missing == "" {
		missing = "-"
	}
	return fmt.Sprintf("reason=%s stale=%s missing=%s attestation_expires=%s",
		execgw.ReasonCapabilityAttestationStale, status.AttestationStale, missing,
		status.AttestationExpiresAt.UTC().Format(time.RFC3339))
}

// exitOnlyLatchDetail 은 진입 래치의 설명 — 상태 출력 · 알림 확인 명령에서 운영자가 읽음.
func exitOnlyLatchDetail(status AutomationStatus) string {
	return "capability attestation is stale (" + status.AttestationStale + "); the engine started exit-only — " +
		"stops and exits run, new entries are refused until the attestation is renewed and the engine restarted"
}

// ExitOnlyStartLine 은 `engine run` 이 청산 전용 기동 때 출력하는 한 줄임(D8 — 읽는 문장, 확인 입력 없음).
func ExitOnlyStartLine(status AutomationStatus) string {
	return fmt.Sprintf("exit-only        낡은 능력 증명(attestation: %s)으로 청산 전용 기동 — 손절 · 익절 · 대사는 돌고 새 진입은 막힌다. "+
		"정상으로 돌아가려면 `tossctl verify run` → `tossctl soak attest` → 엔진 재시작", status.AttestationStale)
}
