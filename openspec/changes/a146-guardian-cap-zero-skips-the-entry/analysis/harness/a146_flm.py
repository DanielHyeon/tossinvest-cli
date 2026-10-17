#!/usr/bin/env python3
"""a146 FLM/BTM 번들 생성 하네스(a139 하네스의 모양을 따름).

사용: python3 a146_flm.py base   — 편집 전: 각 대상의 ast.json(= 현재 소스) 과 ast-base.json(같은 바이트)을 쓰고 FLM/BTM 산문을 만듦.
      python3 a146_flm.py after  — 편집 뒤: ast.json 만 다시 추출하고 FLM/BTM 의 좌표 · 분기 행을 새 AST 로 다시 씀(ast-base.json 유지).
산문(역할 · a146 편집 · 분기 메모 · 시험)은 아래 표가 정본. 저장소 루트는 이 파일 위치에서 유도함.
"""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
CHANGE = "a146-guardian-cap-zero-skips-the-entry"
OUT = ROOT / "openspec" / "changes" / CHANGE / "analysis" / "function-logic"
ENGINE = "go test -tags tossos_testseams ./internal/app/engine"

# (파일, 함수, 역할, a146 편집, 기본 시험 명령)
TARGETS = [
    ("internal/app/engine/strategy_first_leg_admission.go", "strategyFirstLegAdmissionBridge.admit",
     "전략 첫 다리 승인 — 결과 검증 → 권한 수집 → 결속 검증 → 일일 손실 측정 → q_final precheck(거절 분류) → 원자 발급.",
     "precheck 거절 갈래에서 a137 두 갈래 뒤 · `AUTHORITY_MISMATCH` 앞에 분류 하나: `QFinalRefusal` 코드 `EXISTING_GUARDIAN_CAP` 이고 "
     "원인 사슬에 `risk.ErrStrategyQuantityZero` 면 `GUARDIAN_CAP_ZERO` + 원인 `*strategyEntrySkip` + ctx 기록자 호출. 새 분기 하나.",
     f"{ENGINE} -run 'A146|A137|StrategyFirstLegAdmission'"),
    ("internal/app/engine/strategy_market_handoff_delivery.go", "deliverEachStrategyHandoff",
     "한 시장 주기의 handoff 를 조정자 순서대로 몸통에 건넴 — 범위 거절은 건너뛰어 모으고, 일일 손실 표지는 주기를 조용히 끝내며, 그 밖은 멈춤.",
     "범위 거절 판정 다음에 `*strategyEntrySkip` 이면 `continue`(모으지 않음 — 주기 오류가 되지 않게). 새 분기 하나.",
     f"{ENGINE} -run 'A146|A137|Scope|Handoff|Delivery'"),
    ("internal/app/engine/strategy_entry_supervisor.go", "StrategyEntrySupervisor.runMarket",
     "시장 하나의 평가 루프 — 만료 · 허용 판정, 주기 실행(감시견), 결과에 따라 계속 · 삼킴 · 중앙 고장 · 시장 잠금.",
     "주기 ctx 에 a146 스킵 기록자를 싣는 문장 하나(a137 기록자 옆). 분기 · 잠금 · 삼킴 판정 무편집.",
     f"{ENGINE} -run 'A146|A137|Supervisor|Fault'"),
    ("internal/app/engine/strategy_entry_supervisor.go", "StrategyEntrySupervisor.Snapshot",
     "시장별 읽기 전용 관측 값.",
     "`GuardianCapZeroSkips` · `FirstGuardianCapZeroSkip` 두 칸 복사.",
     f"{ENGINE} -run 'A146|Snapshot'"),
    ("internal/app/engine/strategy_entry_supervisor.go", "NewStrategyEntrySupervisor",
     "닫힌 KR/US 조립 검증과 감독자 생성.",
     "옵션 `EntrySkipNotice` 를 감독자 칸으로 옮김(한 칸).",
     f"{ENGINE} -run 'A146|Supervisor'"),
    ("internal/app/engine/strategy_entry_supervisor.go", "Context.NewRefreshingPairedStrategyEntrySupervisor",
     "생산 감독자(권한 갱신 전용 worker 둘) 생성.",
     "`EntrySkipNotice: c.strategyEntrySkipNotice()` 배선 한 줄.",
     f"{ENGINE} -run 'A146|A137|ProductionStrategySupervisor'"),
    ("internal/app/engine/a112_qfinal_family_test.go", "TestAnExistingGuardianCapRefusalIsNotAScopeRefusal",
     "시험 — 버킷 고갈이 아닌 q_final 거절(예산 1 KRW → Guardian 상한 0)은 범위 거절이 아님.",
     "기대를 새 동작으로: 범위 거절이 아님은 그대로, 대신 주기가 멈추지 않고(오류 없음) 두 범위가 다 평가되며 브로커 0.",
     f"{ENGINE} -run TestAnExistingGuardianCapRefusalIsNotAScopeRefusal"),
    ("internal/strategyworker/policy_receipt_test.go", "TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter",
     "시험 — 엔진 시장 런타임 칸 목록이 「실패를 세지 않는다」는 임계값 1 의 영수증.",
     "목록에 a146 칸 넷(`guardianCapZeroSkips` · `firstGuardianCapZeroSkip` · `entrySkipNoticeDay` · `entrySkipNoticed`)과 근거 주석 — 쓰는 곳은 "
     "`recordGuardianCapZeroSkip` 하나, 읽는 곳은 `Snapshot` 하나, `latchMarket` 은 넷 중 어느 것도 보지 않음(세는 것은 실패가 아니라 정책 결과).",
     "go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter"),
]

# 함수마다 그 함수를 실제로 지나는 시험(검사기가 이름의 존재를 대조함).
FN_TESTS = {
    "strategyFirstLegAdmissionBridge.admit": [
        "TestA146ACapZeroPrecheckIsATypedSkip", "TestA146OtherGuardianCapRefusalsStayFaults",
        "TestA137ClassificationIsUnchangedForEveryPrecheckRefusal", "TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing"],
    "deliverEachStrategyHandoff": [
        "TestA146ASkipContinuesToTheNextScopeWithoutAnError", "TestA146ASkipIsNeverJoinedIntoTheCycleError",
        "TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops", "TestA137AMixedCycleStillReportsItsSkippedScopes"],
    "StrategyEntrySupervisor.runMarket": [
        "TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind", "TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday",
        "TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket"],
    "StrategyEntrySupervisor.Snapshot": ["TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol"],
    "NewStrategyEntrySupervisor": ["TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol"],
    "Context.NewRefreshingPairedStrategyEntrySupervisor": ["TestA146TheProductionSupervisorCarriesTheSkipNotice",
                                                           "TestA137TheProductionSupervisorCarriesTheNotice"],
    "TestAnExistingGuardianCapRefusalIsNotAScopeRefusal": ["TestAnExistingGuardianCapRefusalIsNotAScopeRefusal"],
    "TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter": ["TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter"],
}

# 분기 메모: (함수, 소스 줄에 든 문자열) → (부수 효과, 반환/오류). 맞는 것이 없으면 「편집 전과 같음」.
NOTES = {
    "strategyFirstLegAdmissionBridge.admit": [
        ("refusal.Code != \"\"", "—", "결과 무효 거절 그대로(권한 수집 전)"),
        ("b == nil || b.loader == nil", "—", "AUTHORITY_UNAVAILABLE"),
        ("validateStrategyFirstLegAuthority", "—", "AUTHORITY_MISMATCH(결속 불일치 — 진입 · 손절가를 조건에 묶음, design F2)"),
        ("unmeasurable := (*execgw.DailyLossUnmeasurable)", "—", "a137 측정 불가 갈래(무편집)"),
        ("issue.Stage == execgw.StageChain &&", "a137 강화(한도 갈래만)", "a137 DAILY_LOSS_UNMEASURABLE / DAILY_LOSS_LIMIT(무편집)"),
        ("RefusalExistingGuardianCap", "ctx 기록자 호출(`recordStrategyEntrySkip`)",
         "**a146** GUARDIAN_CAP_ZERO + 원인 `*strategyEntrySkip` — 코드 · sentinel 둘 다 맞을 때만, 나머지는 아래 AUTHORITY_MISMATCH"),
        ("RefusalBucketCapExhausted", "—", "버킷 고갈 → 원인을 범위 거절로(무편집)"),
        ("strategyOwnerKeyOf(authority.Result.Lineage)", "—", "범위 키가 있을 때만 범위 거절(무편집)"),
        ("IssuePrecheckedQFinalCampaignFirstLeg", "원장 원자 발급", "ATOMIC_ADMISSION_FAILED(무편집)"),
    ],
    "deliverEachStrategyHandoff": [
        ("range handoffs", "—", "조정자 순서대로 하나씩"),
        ("err == nil", "—", "다음 범위"),
        ("errors.As(err, &scope)", "skipped 에 모음", "다음 범위(끝에 Join 으로 보고 — 무편집)"),
        ("errors.As(err, &skip)", "— (흔적은 admit 의 기록자가 남김)", "**a146** 다음 범위 — skipped 에 넣지 않음"),
        ("errors.As(err, &blocked)", "—", "Join(skipped) — 일일 손실 계좌 차단은 주기를 조용히 끝냄(무편집)"),
        ("len(skipped) == 0", "—", "멈추는 결함 그대로 / skipped 와 Join(무편집)"),
    ],
    "StrategyEntrySupervisor.runMarket": [
        ("if expired", "latchMarket(만료)", "무편집"),
        ("if !allowed", "—", "continue(무편집)"),
        ("if abandoned", "markAbandoned", "무편집"),
        ("if cancelled", "—", "return(무편집)"),
        ("if err == nil", "—", "continue — **스킵만 있던 주기는 여기로 옴**(a146 의 효과 지점, 분기 자체는 무편집)"),
        ("if refreshOnly", "recordSwallowedCycleError", "continue(무편집 — 스킵은 오류가 아니므로 여기 오지 않음)"),
        ("isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity", "진입 차단", "중앙 신호(무편집)"),
        ("if isCentralStrategyIntegrity(err)", "—", "중앙 신호(무편집)"),
        ("waitMarketRestart", "—", "무편집"),
        ("ctx.Err() != nil", "—", "return(무편집)"),
    ],
}

BASE_REVISION = set()


def slug(s):
    return re.sub(r"[^a-z0-9._-]+", "-", s.lower()).strip("-")


def bundle_dir(f, fn):
    return OUT / f"{slug(Path(f).parent.as_posix().replace('/', '-'))}--{slug(fn)}"


def extract(f, fn, cwd=ROOT):
    env = dict(os.environ, GOFLAGS="-trimpath")
    return subprocess.run(["go", "run", "./tools/logic-map", "--file", f, "--func", fn], cwd=cwd, env=env,
                          capture_output=True, check=True).stdout


def note_for(fn, text):
    for needle, effect, ret in NOTES.get(fn, []):
        if needle.strip() and needle.strip() in text:
            return effect, ret
    return "—", "편집 전과 같음(소스 줄 참조)"


def write_md(d, f, fn, role, edit, cmd, ast, base):
    src = (ROOT / f).read_text().splitlines()
    br = ast.get("branches") or []
    s, e = ast["start"]["line"], ast["end"]["line"]
    rets = " · ".join(str(r["at"]["line"]) for r in ast.get("returns") or []) or "(없음)"
    bb = len(base.get("branches") or [])
    rows, trows = [], []
    named = " · ".join("`" + n + "`" for n in FN_TESTS.get(fn, []))
    tests = f"`{cmd}`" + (" · " + named if named else "")
    for b in br:
        ln = b["at"]["line"]
        text = src[ln - 1].strip()
        effect, ret = note_for(fn, text)
        quoted = text.replace("|", "\\|")
        rows.append(f"| {b['id']} | `{quoted}` ({b['kind']} {ln}) | {effect} | {ret} | {tests} |")
        trows.append(f"| {b['id']} | {b['kind']} {ln} | {ret} | {tests} |")
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

**a146 편집.** {edit}

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
| (호출 {len(calls)}종 — 대표: {', '.join('`' + c + '`' for c in calls[:6]) or '없음'}) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
"""
    btm = (f"# Branch Test Map: `{fn.split('.')[-1]}`\n\n- Source: `{f}` ({s}-{e})\n\n| Branch | Anchor | Scenario | Test |\n|---|---|---|---|\n"
           + "\n".join(trows) + "\n")
    (d / "function-logic-map.md").write_text(flm)
    (d / "branch-test-map.md").write_text(btm)


def base_extract(f, fn):
    """ast-base.json — base 커밋의 격리 worktree 에서 추출(워킹트리 소스는 이미 편집됐을 수 있음)."""
    base = (ROOT / "openspec" / "changes" / CHANGE / "base-commit.txt").read_text().strip()
    with tempfile.TemporaryDirectory() as tmp:
        wt = Path(tmp) / "base"
        subprocess.run(["git", "-C", str(ROOT), "worktree", "add", "-q", "--detach", str(wt), base], check=True)
        try:
            return extract(f, fn, cwd=wt)
        finally:
            subprocess.run(["git", "-C", str(ROOT), "worktree", "remove", "--force", str(wt)], check=True)


def main(mode):
    for f, fn, role, edit, cmd in TARGETS:
        d = bundle_dir(f, fn)
        d.mkdir(parents=True, exist_ok=True)
        raw = extract(f, fn)
        (d / "ast.json").write_bytes(raw)
        if mode == "base":
            (d / "ast-base.json").write_bytes(base_extract(f, fn))
        elif not (d / "ast-base.json").exists():
            (d / "ast-base.json").write_bytes(base_extract(f, fn))
        ast = json.loads(raw)
        base = json.loads((d / "ast-base.json").read_bytes())
        write_md(d, f, fn, role, edit, cmd, ast, base)
        subprocess.run(["python3", "tools/logic-map/risk_pattern_report.py", f, "--output", str(d / "risk-pattern-report.md")],
                       cwd=ROOT, check=True, capture_output=True)
        print(d.name, len(ast.get("branches") or []))


if __name__ == "__main__":
    main(sys.argv[1])
