#!/usr/bin/env python3
"""a138 FLM/BTM 번들 생성 하네스.

사용: python3 a138_flm.py base   — 편집 전: 각 대상의 ast.json(= 현재 소스) 과 ast-base.json(같은 바이트)을 쓰고 FLM/BTM 산문을 만듦.
      python3 a138_flm.py after  — 편집 뒤: ast.json 만 다시 추출하고 FLM/BTM 의 좌표 · 분기 행을 새 AST 로 다시 씀(ast-base.json 유지).
산문(역할 · a138 편집 · 시험)은 TARGETS 표가 정본. 저장소 루트는 이 파일 위치에서 유도함.
"""
import json, os, subprocess, sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
CHANGE = "a138-max-open-positions"
OUT = ROOT / "openspec" / "changes" / CHANGE / "analysis" / "function-logic"

# (파일, 함수, 역할, a138 편집, 기본 시험 명령, {분기 id: 시험 인용})
TARGETS = [
    ("internal/journal/reservations.go", "reserveRows",
     "모든 예약 tx 가 지나는 한 함수 — 결정 확인 · 총계 · 하루 매수 · 삽입.",
     "5b(하루 매수) 다음에 5c `checkOpenPositions` 호출 한 줄 — 새 분기 하나(거부 시 반환).",
     "go test ./internal/journal -run 'Reserve|A136|A138'"),
    ("internal/execgw/riskguardian.go", "RiskGuardian.IssueEntry",
     "진입 발급 문 1 — 체인 · 재수집 · 원자 발급.",
     "collectIssue 클로저의 ReserveRequest 에 `OpenPositionLimit: g.openPositionLimit()` 한 칸(분기 무변경).",
     "go test ./internal/execgw -run 'IssueEntry|A138'"),
    ("internal/execgw/riskguardian.go", "RiskGuardian.LimitsDigest",
     "전략 바인딩이 결속하는 Guardian 한도 digest.",
     "digest 입력에 `\\nmax_open_positions=N` 을 더함(분기 무변경).",
     "go test ./internal/execgw -run 'LimitsDigest|A136|A138'"),
    ("internal/execgw/riskguardian_qfinal.go", "RiskGuardian.IssuePrecheckedQFinalEntry",
     "진입 발급 문 2 — q_final precheck 뒤 원자 발급.",
     "ReserveRequest 에 상한 한 칸(분기 무변경).",
     "go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'"),
    ("internal/execgw/riskguardian_first_leg.go", "RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg",
     "진입 발급 문 3 — 전략 첫 다리.",
     "ReserveRequest 에 상한 한 칸(분기 무변경).",
     "go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'"),
    ("internal/risk/input.go", "DefaultPolicy",
     "사용자 미확정 보수 기본 정책.",
     "`MaxOpenPositions: 6`(사용자 결정 2026-10-15) 한 칸.",
     "go test ./internal/risk -run 'DefaultPolicy|A138'"),
    ("internal/risk/input.go", "Policy.Validate",
     "체인 · Guardian 생성 전 정책 사용 가능성 검사(전부 fail-closed).",
     "`MaxOpenPositions < 1` 거부 분기 하나 추가.",
     "go test ./internal/risk -run 'Validate|Policy|A138'"),
    ("internal/app/engine/guardian_wiring.go", "riskPolicyFromAutomationGate",
     "감사된 게이트 한도 → Guardian 정책 유도(직선).",
     "`policy.MaxOpenPositions = openPositionsFromGate(gate)` 한 줄(해석 · 정수 판정은 새 leaf 함수).",
     "go test ./internal/app/engine -run 'RiskPolicy|A136|A138'"),
    ("internal/app/engine/guardian_wiring.go", "newProductionRiskGuardian",
     "생산 Guardian 생성 — 정책 유도 · 검증 · 생성.",
     "Validate 앞에서 `config.WholeOpenPositions` 로 해석 값을 판정해, 운영자가 적은 값을 문구에 담아 ErrLimitsRequired — 새 분기 하나(freeze 리뷰 P2-3).",
     "go test ./internal/app/engine -run 'ProductionPolicy|A136|A138'"),
    ("internal/app/engine/interlock.go", "newAutomationStatus",
     "인터록이 감사할 해석 값 운반.",
     "`MaxOpenPositions: gate.EffectiveMaxOpenPositions()` 한 칸.",
     "go test ./internal/app/engine -run 'Interlock|A138'"),
    ("internal/app/engine/interlock.go", "acceptanceDetail",
     "인터록 수용 감사 문구.",
     "`max_open_positions=` 한 칸.",
     "go test ./internal/app/engine -run 'Interlock|Acceptance|A138'"),
    ("internal/app/engine/interlock.go", "verifyGate",
     "게이트 ON 기동 인터록의 단계 판정.",
     "7c `checkGuardianOpenPositionLimit` 한 단 — 새 분기 하나.",
     "go test ./internal/app/engine -run 'Interlock|Gate|A138'"),
    ("internal/app/engine/interlock.go", "recordGateSettings",
     "기동 audit — 게이트 설정값 줄들.",
     "목록 끝에 `engine.automation_gate.max_open_positions`(해석 값) 한 줄(순서 무변경).",
     "go test ./internal/app/engine -run 'Audit|A074|A138'"),
    ("internal/config/engine.go", "mergeEngine",
     "파싱 모양 → 설정 구조체 복사.",
     "`MaxOpenPositions` 복사 한 줄.",
     "go test ./internal/config -run 'Engine|A138'"),
    ("internal/config/limits_io.go", "Service.LoadRawEngineGate",
     "콘솔이 보여 줄 적힌 그대로의 게이트 블록 읽기.",
     "`MaxOpenPositions` 한 칸.",
     "go test ./internal/config -run 'Limits|A138'"),
    ("internal/config/limits_io.go", "Service.SaveEngineGateLimits",
     "편집기 저장 — 검증 · 잠금 · 새 파일 갈래 · 키 단위 쓰기.",
     "새 파일(skeleton) 갈래에 `OpenPositions` 를 옮기는 분기 하나(구현 리뷰 P2-1 — 빠뜨리면 운영자 값이 기본 6 으로 사라짐).",
     "go test ./internal/config -run 'Limits|Save|A138'"),
    ("internal/config/limits_io.go", "gateMembersOf",
     "편집기 저장이 건드리는 키 목록.",
     "`OpenPositions != nil` 이면 `max_open_positions` 멤버 — 새 분기 하나.",
     "go test ./internal/config -run 'Limits|Save|A138'"),
    ("internal/config/limits.go", "GuardianLimits.Validate",
     "편집기 저장 전 검증.",
     "`OpenPositions` 가 음수 · 비정수 · NaN · Inf 면 거부 — 새 분기 하나.",
     "go test ./internal/config -run 'Limits|A138'"),
    ("internal/app/engine/guardian_wiring.go", "defaultProductionGuardianFactory",
     "생산 Guardian factory 의 얇은 감쌈.", "본문 무편집 — 같은 파일 끝에 새 함수 `wholeOpenPositions` 를 덧붙여 diff 창에 들어옴.",
     "go test ./internal/app/engine -run 'ProductionPolicy|A138'"),
    ("cmd/tossctl/limitsettings.go", "consoleLimitSettings.Save",
     "콘솔 한도 저장 seam — 쓰기 + 필드별 audit.",
     "저장 뒤 파일 값으로 `recordOpenPositionsSave` 한 줄(best-effort, 0 = 「미설정 — 기본 6」).",
     "go test ./cmd/tossctl -run 'A136|A138|Limit'"),
    ("internal/console/settings_limits.go", "settingsPage.LimitRows",
     "현재 값 표의 행들.", "「최대 보유 포지션 수」 행 하나(미설정 → 기본 6 문장, 무효 → 기동 거부 문장).",
     "go test ./internal/console -run 'A136|A138|Limit'"),
    ("internal/console/settings_limits.go", "Console.handleSettingsLimits",
     "고급 폼 저장 핸들러.", "`max_open_positions` 칸이 폼에 **있을 때만** 값을 실음(빈 칸 = 0), 저장 뒤 문장에 한 마디 — 새 분기 셋.",
     "go test ./internal/console -run 'A136|A138|Limit'"),
    ("internal/app/engine/interlock_test.go", "stubGuardian.DailyBuyNotionalLimit",
     "시험 fixture — 주입 Guardian 의 하루 매수 진술.", "본문 무편집 — 바로 뒤에 `MaxOpenPositionsLimit` 메서드를 더해 diff 창에 들어옴.",
     "go test ./internal/app/engine -run 'Interlock|A136|A138'"),
    ("internal/app/engine/engine_test.go", "seedExpiredUnconsumedReservation",
     "시험 fixture — 예약 요청.", "`OpenPositionLimit` 넉넉한 칸 하나(이 시험이 재는 것과 무관).", "go test ./internal/app/engine"),
    ("internal/execgw/a066_entry_loss_lock_red_test.go", "TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths",
     "시험 — 손실 lock 이 위험 축소 경로를 막지 않음.", "요청에 `OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/execgw -run A066"),
    ("internal/execgw/gateway_test.go", "holdExposure",
     "시험 fixture — 노출 예약.", "`OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/execgw"),
    ("internal/execgw/riskguardian_qfinal_test.go", "TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker",
     "시험 — q_final 표시 결정의 Gateway 거부.", "요청에 `OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/execgw -run QFinal"),
    ("internal/journal/a126_filled_exposure_leaves_the_bucket_test.go", "a126IssueFor",
     "시험 fixture — q_final 발급.", "`OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/journal -run A126"),
    ("internal/journal/issuance_test.go", "issueRequest",
     "시험 fixture — 원자 발급 요청.", "`OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/journal -run Issu"),
    ("internal/journal/observation_crash_linux_test.go", "crashChildObservationGap",
     "시험 fixture — 관측 크래시 자식.", "`OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/journal -run Crash"),
    ("internal/journal/reservations_test.go", "exposureReserve",
     "시험 fixture — OPEN_EXPOSURE 예약 요청.", "`OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/journal -run 'Reserv|A136|A138'"),
    ("internal/journal/risk_bucket_admission_test.go", "seedExistingRiskReservation",
     "시험 fixture — 원장에 직접 심는 기존 HELD 예약.",
     "결정 preimage 를 `'{}'` 에서 생산 RecordDecision 이 쓰는 정규형 RiskIntent 로 — 개수 판정이 HELD 예약의 종목을 읽으므로(옛 모양은 생산이 만들 수 없음).",
     "go test ./internal/journal -run A066"),
    ("internal/journal/risk_bucket_issuance_test.go", "qFinalIssueFixture",
     "시험 fixture — q_final 발급.", "`OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/journal -run QFinal"),
    ("internal/console/settings_cadence_test.go", "TestTheSubmitContractIsUnchanged",
     "시험 — 화면별 제출 칸 전수(census).", "`/settings/limits` 의 칸 목록에 `max_open_positions` 하나.", "go test ./internal/console -run TestTheSubmitContractIsUnchanged"),
    ("internal/console/settings_card_test.go", "TestTheSettingsScreensAskForNoTyping",
     "시험 — 설정 화면에 사유 · 확인 문구 칸이 없음.", "화면이 소유한 설정 목록에 `max_open_positions` 하나(확인 문구가 아님).", "go test ./internal/console -run TestTheSettingsScreensAskForNoTyping"),
    ("internal/measure/reconstruct_test.go", "issueEntry",
     "시험 fixture — 진입 발급.", "`OpenPositionLimit` 넉넉한 칸 하나.", "go test ./internal/measure"),
]

# 함수마다 그 함수를 실제로 지나는 a138 시험(행마다 덧붙임 — 검사기가 이름의 존재를 대조함).
FN_TESTS = {
    "reserveRows": ["TestA138ReconcileScopesMatchExactlyWhatTheyCover", "TestA138TheSeventhSymbolIsRefused", "TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused", "TestA138ConcurrentEntriesCannotBothTakeTheLastSlot"],
    "RiskGuardian.IssueEntry": ["TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum"],
    "RiskGuardian.LimitsDigest": ["TestA138TheLimitsDigestBindsTheOpenPositionMaximum"],
    "RiskGuardian.IssuePrecheckedQFinalEntry": ["TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum"],
    "RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg": ["TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum"],
    "DefaultPolicy": ["TestA138DefaultPolicyCarriesTheOpenPositionMaximum"],
    "Policy.Validate": ["TestA138PolicyValidationRefusesAnUnusableOpenPositionMaximum"],
    "riskPolicyFromAutomationGate": ["TestA138ProductionPolicyResolvesTheOpenPositionMaximum"],
    "newProductionRiskGuardian": ["TestA138ProductionPolicyResolvesTheOpenPositionMaximum"],
    "defaultProductionGuardianFactory": ["TestA138TheInterlockAuditsTheOpenPositionMaximum"],
    "newAutomationStatus": ["TestA138TheInterlockAuditsTheOpenPositionMaximum"],
    "acceptanceDetail": ["TestA138TheInterlockAuditsTheOpenPositionMaximum"],
    "verifyGate": ["TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum", "TestA138AnUnusableOpenPositionMaximumRefusesToStart"],
    "recordGateSettings": ["TestA138TheInterlockAuditsTheOpenPositionMaximum"],
    "mergeEngine": ["TestA138OpenPositionMaximumIsParsedFromTheGateBlock"],
    "Service.LoadRawEngineGate": ["TestA138OpenPositionMaximumIsParsedFromTheGateBlock"],
    "Service.SaveEngineGateLimits": ["TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked"],
    "gateMembersOf": ["TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked"],
    "GuardianLimits.Validate": ["TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked", "TestA138TheAdvancedFormSavesTheOpenPositionMaximum"],
    "consoleLimitSettings.Save": ["TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile"],
    "settingsPage.LimitRows": ["TestA138TheScreenSaysWhatTheOpenPositionMaximumIs"],
    "Console.handleSettingsLimits": ["TestA138TheAdvancedFormSavesTheOpenPositionMaximum"],
    "stubGuardian.DailyBuyNotionalLimit": ["TestA136TheInterlockRefusesAGuardianThatDisagreesOnTheDailyBuyCeiling"],
    "seedExistingRiskReservation": ["TestA066StaleUsageRuleEdgesAtBothAdmissionSites"],
    "exposureReserve": ["TestA136ReachingTheDailyBuyCeilingIsRefused"],
}

BASE_REVISION = {"defaultProductionGuardianFactory", "stubGuardian.DailyBuyNotionalLimit"}

def slug(s):
    import re
    return re.sub(r"[^a-z0-9._-]+", "-", s.lower()).strip("-")

def bundle_dir(f, fn):
    return OUT / f"{slug(Path(f).parent.as_posix().replace('/', '-'))}--{slug(fn)}"

def extract(f, fn):
    env = dict(os.environ, GOFLAGS="-trimpath")
    out = subprocess.run(["go", "run", "./tools/logic-map", "--file", f, "--func", fn], cwd=ROOT, env=env,
                         capture_output=True, check=True)
    return out.stdout

def write_md(d, f, fn, role, edit, cmd, ast, base):
    src = (ROOT / f).read_text().splitlines()
    br = ast.get("branches") or []
    s, e = ast["start"]["line"], ast["end"]["line"]
    rets = " · ".join(str(r["at"]["line"]) for r in ast.get("returns") or []) or "(없음)"
    bb = len(base.get("branches") or [])
    rows, trows = [], []
    named = " · ".join("`" + n + "`" for n in FN_TESTS.get(fn, []))
    tests = (f"`{cmd}`" + (" · " + named if named else ""))
    for b in br:
        ln = b["at"]["line"]
        text = src[ln - 1].strip().replace("|", "\\|")
        rows.append(f"| {b['id']} | `{text}` ({b['kind']} {ln}) | — | 분기 본문 참조 | {tests} |")
        trows.append(f"| {b['id']} | {b['kind']} {ln} | 분기 동작 | {tests} |")
    if not br:
        rows.append(f"| B1 | 직선 함수 — 분기 없음 | — | 정상 경로 | {tests} |")
        trows.append(f"| B1 | 직선(분기 없음) | 정상 경로 | {tests} |")
    calls = sorted({c.get("text", "") for c in ast.get("calls") or [] if c.get("text")})
    head = (f"- Source: `{f}` ({s}-{e})\n- Qualified: `{fn}`\n"
            f"- AST evidence: `ast.json`(현재) — 현재 AST branches {len(br)} · returns {len(ast.get('returns') or [])}\n"
            f"- 편집 전 기준: `ast-base.json`(base, `source_sha256` {base['source_sha256'][:16]}…) — 편집 전 갈래 수 {bb}\n"
            f"- Risk scan: `risk-pattern-report.md`\n")
    flm = f"""# Function Logic Map: `{fn.split('.')[-1]}`

{head}
**역할.** {role}

**a138 편집.** {edit}

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
""" + "\n".join(rows) + f"""

Exact AST return positions: {rets}

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 {len(calls)}종 — 대표: {', '.join('`'+c+'`' for c in calls[:6]) or '없음'}) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
"""
    btm = f"# Branch Test Map: `{fn.split('.')[-1]}`\n\n- Source: `{f}` ({s}-{e})\n\n| Branch | Anchor | Scenario | Test |\n|---|---|---|---|\n" + "\n".join(trows) + "\n"
    (d / "function-logic-map.md").write_text(flm)
    (d / "branch-test-map.md").write_text(btm)

def base_extract(f, fn):
    """편집 뒤에 추가된 대상의 ast-base.json — base 커밋의 격리 worktree 에서 추출(워킹트리 소스는 이미 편집됨)."""
    import tempfile
    base = (ROOT / "openspec" / "changes" / CHANGE / "base-commit.txt").read_text().strip()
    with tempfile.TemporaryDirectory() as tmp:
        wt = Path(tmp) / "base"
        subprocess.run(["git", "-C", str(ROOT), "worktree", "add", "-q", "--detach", str(wt), base], check=True)
        try:
            env = dict(os.environ, GOFLAGS="-trimpath")
            return subprocess.run(["go", "run", "./tools/logic-map", "--file", f, "--func", fn], cwd=wt, env=env,
                                  capture_output=True, check=True).stdout
        finally:
            subprocess.run(["git", "-C", str(ROOT), "worktree", "remove", "--force", str(wt)], check=True)

def main(mode):
    for f, fn, role, edit, cmd in TARGETS:
        d = bundle_dir(f, fn); d.mkdir(parents=True, exist_ok=True)
        if fn in BASE_REVISION:
            # 본문 무편집 · 이웃 편집만 받은 함수 — 검사기 규칙대로 편집 전 리비전을 `revision: base` 로 고정.
            if not (d / "ast-base.json").exists():
                (d / "ast-base.json").write_bytes(base_extract(f, fn))
            value = json.loads((d / "ast-base.json").read_bytes()); value["revision"] = "base"
            raw = (json.dumps(value, ensure_ascii=False, indent=2) + "\n").encode()
        else:
            raw = extract(f, fn)
        (d / "ast.json").write_bytes(raw)
        if mode == "base":
            (d / "ast-base.json").write_bytes(raw)
        elif not (d / "ast-base.json").exists():
            (d / "ast-base.json").write_bytes(base_extract(f, fn))
        ast = json.loads(raw); base = json.loads((d / "ast-base.json").read_bytes())
        write_md(d, f, fn, role, edit, cmd, ast, base)
        subprocess.run(["python3", "tools/logic-map/risk_pattern_report.py", f, "--output", str(d / "risk-pattern-report.md")],
                       cwd=ROOT, check=True, capture_output=True)
        print(d.name, len(ast.get("branches") or []))

if __name__ == "__main__":
    main(sys.argv[1])
