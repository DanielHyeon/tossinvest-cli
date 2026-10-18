#!/usr/bin/env python3
"""a157 FLM/BTM 번들 생성 하네스(a146 하네스의 모양을 따름).

사용: python3 a157_flm.py base   — 편집 전: 각 대상의 ast.json(= 현재 소스) 과 ast-base.json(base 커밋 소스)을 쓰고 FLM/BTM 산문을 만듦.
      python3 a157_flm.py after  — 편집 뒤: ast.json 만 다시 추출하고 FLM/BTM 의 좌표 · 분기 행을 새 AST 로 다시 씀(ast-base.json 유지).
산문(역할 · a157 편집 · 분기 메모 · 시험)은 아래 표가 정본. 저장소 루트는 이 파일 위치에서 유도함.
분기마다 그 분기를 실제로 지나는 시험 이름을 적음(메모의 needle 이 분기 줄에 들어 있으면 그 메모의 시험) — 모든 행에 같은 목록을 쓰지 않음.
"""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
CHANGE = "a157-first-leg-binding-before-sizing"
OUT = ROOT / "openspec" / "changes" / CHANGE / "analysis" / "function-logic"
TAGS = "go test -tags tossos_testseams"

# (파일, 함수, 역할, a157 편집, 기본 시험 명령)
TARGETS = [
    ("internal/execgw/riskguardian_first_leg.go", "RiskGuardian.PrecheckQFinalCampaignFirstLeg",
     "전략 첫 다리 precheck — 진입 판정(`PrecheckQFinalEntry`: 사이징 · 버킷 · 체인 · 노출)과 첫 다리 고유 결속 판정 일곱(B2 정책 버전 · B3 소유자 "
     "prospective 세대 · B4 가격 단위 · B5 계보 결속 · B6 포지션 세대 · B7 attempt/활성화 · B8 주간 예약)을 거쳐 계보 투영(B9) 뒤 봉인을 돌려줌. 원장 · 브로커 호출 없음.",
     "design D1 — ① nil 가드(`PrecheckQFinalEntry` B1 과 같은 거절) ② B2~B8 을 편집 전 상대 순서로 **요청 값**에 대해 판정하고 첫 결함을 `bindErr` 로 잡음"
     "(입력 출처: `request.Entry.Admission.TransactionID` · `.Owner` · `request.Entry.QCandidate` — F3 · F4 로 같은 값) ③ `PrecheckQFinalEntry` 호출 한 번 "
     "④ 진입 거절이 있고 결속 결함이 없거나 진입 거절이 일일 손실 체인 거절(`execgw.DailyLossRefusalKindOf`)이면 진입 거절 ⑤ 결속 결함 ⑥ B9 투영(무편집). "
     "판정식 · 코드 · field · 문구 무변경.",
     f"{TAGS} ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'"),
    ("internal/app/engine/strategy_first_leg_admission.go", "strategyFirstLegAdmissionBridge.admit",
     "전략 첫 다리 승인 — 결과 검증 → 권한 수집 → 결속 검증 → 일일 손실 측정 → q_final precheck(거절 분류) → 원자 발급.",
     "precheck 거절 갈래의 a137 두 조건(측정 불가: 표지 + 체인 · 입력 불가 / 한도: 체인 · 한도 도달)을 execgw 판별 함수 "
     "`DailyLossRefusalKindOf` 의 결과로 가름 — 조건을 글자 그대로 옮겼으므로 분기 결과(코드 · 원인 · 강화 호출)는 같음(design D1 「판별은 한 자리」).",
     f"{TAGS} ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'"),
    ("internal/app/engine/strategy_entry_supervisor.go", "Context.NewPairedStrategyEntryProductionAssembly",
     "조립 감독자 생성 — 권한 파도(일정 · 후보 · 경로 · FX · 레인 · 제안 · 위험 · 계좌)를 모아 첫 다리 승인 · dispatch 주기 · 두 시장 worker 를 세우고 "
     "감독자를 만들어 투영을 발표함. 감독자는 투영에만 쓰이고 Run 되지 않음(design F9).",
     "394 행 감독자 옵션에 `DailyLossNotice: c.strategyDailyLossNotice()` · `EntrySkipNotice: c.strategyEntrySkipNotice()` 를 더하고, "
     "`c.Entry != nil` 일 때만 `opts.EntryGate = c.Entry`(타입 있는 nil 을 인터페이스에 넣지 않음 — design F10 · D4 · D5). 새 분기 하나.",
     f"{TAGS} ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'"),
]

# 함수마다 그 함수를 통째로 지나는 시험(분기 메모가 없을 때의 기본).
FN_TESTS = {
    "RiskGuardian.PrecheckQFinalCampaignFirstLeg": [
        "TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes", "TestA157AnAcceptedFirstLegSealsTheSameValues"],
    "strategyFirstLegAdmissionBridge.admit": [
        "TestStrategyFirstLegAdmissionClassifiesPairedFailures", "TestA137ClassificationIsUnchangedForEveryPrecheckRefusal"],
    "Context.NewPairedStrategyEntryProductionAssembly": [
        "TestTheMarketThatLeadsAWaveAlwaysPublishesIt", "TestAFailedWaveCarriesNoAssemblyForEitherMarket"],
}

# 편집 뒤에만 있는 시험(1.1 에서 더함) — after 모드에서만 인용.
AFTER_ONLY = {
    "TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing", "TestA157DailyLossRefusalKindOfTable",
    "TestA157TheAssemblySupervisorCarriesBothNotices", "TestA157TheAssemblyNoticesReachTheLog",
    "TestA157TheAssemblySupervisorEntryGate",
}

# 분기 메모: (함수, 소스 줄에 든 문자열) → (부수 효과, 반환/오류, 시험). 맞는 것이 없으면 「편집 전과 같음」 + 함수 기본 시험.
B_ALONE = "TestA157ABindingDefectAloneIsReportedAsItself"
CAPZ = "TestA157CapZeroAndABindingDefect"
DAILY = "TestA157DailyLossIsReportedBeforeABindingDefect"
NOTES = {
    "RiskGuardian.PrecheckQFinalCampaignFirstLeg": [
        ("bindErr := func() error", "—", "결속 판정 닫힌 함수의 즉시 호출 한 번(최상위 대입 · 진입 판정 앞 — 구조 시험)",
         ["TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing"]),
        ("g == nil", "—", "RISK_CALCULATION_INVALID/guardian(**a157 nil 가드** — 결속 판정이 `g.policyVersion` · `g.accountRef` 를 읽기 전)",
         ["TestA157ANilGuardianRefusesBeforeAnyBindingJudgement"]),
        ("bindErr == nil || DailyLossRefusalKindOf", "—",
         "**a157 ④** 진입 거절 반환 — 결속 결함이 없거나 진입 거절이 일일 손실 체인 거절(1-1)",
         [DAILY, "TestA157AnEntryRefusalAloneIsReportedAsItself", "TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect",
          "TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect", "TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect"]),
        ("bindErr != nil", "—", "**a157 ⑤** 결속 결함 반환(상한 0 · 버킷 · 그 밖의 진입 거절보다 먼저)",
         [CAPZ, "TestA157OtherEntryRefusalsAndABindingDefect", "TestA157OneInputThatTripsBothJudgements", "TestA157EngineACapZeroNoLongerHidesABindingDefect"]),
        # 「prev:」 메모는 분기 줄 바로 앞 세 줄에서 찾음 — `if err != nil {` 이 여럿이라 그 오류를 만든 호출로 가름.
        ("prev:g.PrecheckQFinalEntry", "—", "편집 전 B1: 진입 판정 거절 그대로(결속 판정은 돌지 않음 — a157 이 바꾸는 자리)",
         ["TestA157AnEntryRefusalAloneIsReportedAsItself", CAPZ]),
        ("prev:journal.QFinalPolicyVersion", "—", "B2 정책 버전 결속 오류(`journal.ErrInvalidRequest`) 그대로", [B_ALONE, CAPZ, DAILY]),
        ("prev:guardianWeeklyFirstLegBinding", "—", "B8 RISK_CALCULATION_INVALID/weekly_reservation(+주간 충돌 원인)", [B_ALONE, CAPZ, DAILY]),
        ("prev:ProjectAcceptedStrategyflowLineage", "—", "B9 RISK_CALCULATION_INVALID/first_leg_projection(무편집 — q_final 이 필요해 사이징 뒤)",
         ["TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes"]),
        ("ProspectiveGeneration", "—", "B3 OWNER_CONFLICT/prospective_generation", [B_ALONE, CAPZ, DAILY]),
        ("!entryOK || !stopOK || !targetOK", "—", "B4 RISK_CALCULATION_INVALID/first_leg_price_unit", [B_ALONE, CAPZ, DAILY]),
        ("request.Result.Quantity !=", "—", "B5 RISK_CALCULATION_INVALID/first_leg_lineage(q_candidate · 계좌 · 시장 · 종목 · 가격 · 소유자 · 캠페인)",
         [B_ALONE, CAPZ, DAILY, "TestA157OneInputThatTripsBothJudgements"]),
        ("currentGeneration < 0", "—", "B6 OWNER_CONFLICT/position_generation", [B_ALONE, CAPZ, DAILY]),
        ("request.AttemptID", "—", "B7 RISK_CALCULATION_INVALID/first_leg_strategy", [B_ALONE, CAPZ, DAILY]),
    ],
    "strategyFirstLegAdmissionBridge.admit": [
        ("refusal.Code != \"\"", "—", "결과 무효 거절 그대로(권한 수집 전)", ["TestStrategyFirstLegAdmissionRejectsInvalidBeforeAuthority"]),
        ("b == nil || b.loader == nil", "—", "AUTHORITY_UNAVAILABLE", ["TestStrategyFirstLegAdmissionDormantPairedKRUS"]),
        ("validateStrategyFirstLegAuthority", "—", "AUTHORITY_MISMATCH(결속 불일치 — B3~B7 꼴을 precheck 앞에서 막음, design F8)",
         ["TestStrategyFirstLegAdmissionRejectsAuthorityTamperPairedKRUS"]),
        ("DailyLossRefusalKindUnmeasurable", "—", "**a157** 판별 결과 측정 불가 → DAILY_LOSS_UNMEASURABLE + 원인(편집 전 B6-B7 과 같음)",
         ["TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault", "TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect"]),
        ("DailyLossRefusalKindLimitReached", "a137 강화(`EscalateDailyLossLimit`)",
         "**a157** 판별 결과 한도 도달 → DAILY_LOSS_LIMIT + 강화 결과 원인(편집 전 B8 과 같음)",
         ["TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault", "TestA137AFailedEscalationIsJoinedNotLost",
          "TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect", "TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect"]),
        ("switch execgw.DailyLossRefusalKindOf", "—", "**a157** 일일 손실 판별(한 자리) — 아니면 아래 갈래",
         ["TestA137ClassificationIsUnchangedForEveryPrecheckRefusal", "TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect"]),
        ("unmeasurable := (*execgw.DailyLossUnmeasurable)", "—", "편집 전 a137 측정 불가 갈래(표지) — a157 이 판별 함수로 옮김",
         ["TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault", "TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect"]),
        # 편집 전 B7 · B8 은 소스 줄이 글자까지 같음 — 바로 앞 줄(측정 불가 표지 갈래 안인지)로 가름.
        ("prev:unmeasurable := (*execgw.DailyLossUnmeasurable)", "—",
         "편집 전 a137 측정 불가 갈래의 체인 · 입력 불가 조건 → DAILY_LOSS_UNMEASURABLE — a157 이 판별 함수로 옮김",
         ["TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault", "TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect"]),
        ("issue.Stage == execgw.StageChain &&", "a137 강화(`EscalateDailyLossLimit`)",
         "편집 전 a137 한도 갈래(체인 · 한도 도달) → DAILY_LOSS_LIMIT — a157 이 판별 함수로 옮김",
         ["TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault", "TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect"]),
        ("prev:collectStrategyFirstLegAuthority", "—", "AUTHORITY_COLLECTION_FAILED + 원인(무편집)", ["TestStrategyFirstLegAdmissionClassifiesPairedFailures"]),
        ("prev:PrecheckQFinalCampaignFirstLeg(authority)", "—", "precheck 거절 → 아래 분류(일일 손실 · 상한 0 · 버킷 · 그 밖 AUTHORITY_MISMATCH)",
         ["TestA137ClassificationIsUnchangedForEveryPrecheckRefusal", "TestA157EngineABindingDefectAloneIsAnAuthorityMismatch"]),
        ("RefusalExistingGuardianCap", "ctx 기록자 호출(`recordStrategyEntrySkip`)", "a146 GUARDIAN_CAP_ZERO(무편집)",
         ["TestA146ACapZeroPrecheckIsATypedSkip", "TestA157EngineACapZeroAloneIsStillASkip"]),
        ("authority.Entry.FXAuthority.Digest()", "—", "스킵 알림의 FX 표시(무편집)", ["TestA146ACapZeroPrecheckIsATypedSkip"]),
        ("RefusalBucketCapExhausted", "—", "버킷 고갈 → 원인을 범위 거절로(무편집)", ["TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket"]),
        ("strategyOwnerKeyOf(authority.Result.Lineage)", "—", "범위 키가 있을 때만 범위 거절(무편집)",
         ["TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket"]),
        ("prev:IssuePrecheckedQFinalCampaignFirstLeg", "원장 원자 발급", "ATOMIC_ADMISSION_FAILED(무편집)", ["TestStrategyFirstLegAdmissionClassifiesPairedFailures"]),
    ],
    "Context.NewPairedStrategyEntryProductionAssembly": [
        ("c == nil", "—", "조립 불가 오류(무편집)", ["TestTheMarketThatLeadsAWaveAlwaysPublishesIt"]),
        ("c.Journal != nil", "—", "원장 경로 유도(무편집)", ["TestTheMarketThatLeadsAWaveAlwaysPublishesIt"]),
        ("journalPath != \"\"", "—", "증거 DB 경로 유도(무편집)", ["TestTheMarketThatLeadsAWaveAlwaysPublishesIt"]),
        ("err != nil", "—", "레인 · 감독자 · 발표 오류 그대로(무편집)", ["TestAFailedWaveCarriesNoAssemblyForEitherMarket"]),
        ("clk != nil", "dispatch 주기에 실시계", "무편집", ["TestTheProductionAssemblyGivesTheDispatchCycleTheRealClock"]),
        ("range []StrategyMarket", "worker 둘", "무편집", ["TestTheMarketThatLeadsAWaveAlwaysPublishesIt"]),
        ("c.Entry != nil", "`opts.EntryGate = c.Entry`", "**a157** 게이트가 있을 때만 배선 — 없으면 인터페이스 칸이 nil 로 남음(F10)",
         ["TestA157TheAssemblySupervisorEntryGate"]),
    ],
}


def slug(s):
    return re.sub(r"[^a-z0-9._-]+", "-", s.lower()).strip("-")


def bundle_dir(f, fn):
    return OUT / f"{slug(Path(f).parent.as_posix().replace('/', '-'))}--{slug(fn)}"


def extract(f, fn, cwd=ROOT):
    env = dict(os.environ, GOFLAGS="-trimpath")
    return subprocess.run(["go", "run", "./tools/logic-map", "--file", f, "--func", fn], cwd=cwd, env=env,
                          capture_output=True, check=True).stdout


def note_for(fn, text, prev, mode):
    for needle, effect, ret, tests in NOTES.get(fn, []):
        where, needle = (prev, needle[5:]) if needle.startswith("prev:") else (text, needle)
        if needle.strip() and needle.strip() in where:
            return effect, ret, [t for t in tests if mode == "after" or t not in AFTER_ONLY]
    return "—", "편집 전과 같음(소스 줄 참조)", FN_TESTS.get(fn, [])


def write_md(d, f, fn, role, edit, cmd, ast, base, mode):
    src = (ROOT / f).read_text().splitlines()
    br = ast.get("branches") or []
    s, e = ast["start"]["line"], ast["end"]["line"]
    rets = " · ".join(str(r["at"]["line"]) for r in ast.get("returns") or []) or "(없음)"
    bb = len(base.get("branches") or [])
    rows, trows = [], []
    for b in br:
        ln = b["at"]["line"]
        text = src[ln - 1].strip()
        prev = " ".join(line.strip() for line in src[max(0, ln - 4):ln - 1])
        effect, ret, tests = note_for(fn, text, prev, mode)
        cell = f"`{cmd}`" + "".join(" · `" + n + "`" for n in tests)
        quoted = text.replace("|", "\\|")
        rows.append(f"| {b['id']} | `{quoted}` ({b['kind']} {ln}) | {effect} | {ret} | {cell} |")
        trows.append(f"| {b['id']} | {b['kind']} {ln} | {ret} | {cell} |")
    if not br:
        cell = f"`{cmd}`" + "".join(" · `" + n + "`" for n in FN_TESTS.get(fn, []))
        rows.append(f"| B1 | 직선 함수 — 분기 없음 | — | 정상 경로 | {cell} |")
        trows.append(f"| B1 | 직선(분기 없음) | 정상 경로 | {cell} |")
    calls = sorted({c.get("text", "") for c in ast.get("calls") or [] if c.get("text")})
    stage = "편집 뒤" if mode == "after" else "편집 전(= base 소스)"
    head = (f"- Source: `{f}` ({s}-{e})\n- Qualified: `{fn}`\n"
            f"- AST evidence: `ast.json`({stage}) — 현재 AST branches {len(br)} · returns {len(ast.get('returns') or [])}\n"
            f"- 편집 전 기준: `ast-base.json`(base `7b1748481`, `source_sha256` {base['source_sha256'][:16]}…) — 편집 전 갈래 수 {bb}\n"
            f"- Risk scan: `risk-pattern-report.md`\n")
    flm = f"""# Function Logic Map: `{fn.split('.')[-1]}`

{head}
**역할.** {role}

**a157 편집.** {edit}

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD(설계 근거는 design F1~F10) | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
""" + "\n".join(rows) + f"""

Exact AST return positions: {rets}

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 {len(calls)}종 — 대표: {', '.join('`' + c + '`' for c in calls[:8]) or '없음'}) | 본문 참조 | 소스 참조 — 원장 · 브로커 · 파일 호출은 design F2 · F9 | AST |

## State mutations and fallbacks

- 편집은 위 「a157 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a157 편집」에 적은 자리.
- High-risk impact: yes(Guardian 첫 다리 판정 순서 · 일일 손실 분류 · 감독자 조립 — 받아들이는 요청 집합 불변(design D2), 주문을 늘리는 갈래 없음,
  일일 손실 강화 보존(1-1), 매도 · 손절 · 청산 경로 무관)
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
        if mode == "base" or not (d / "ast-base.json").exists():
            (d / "ast-base.json").write_bytes(base_extract(f, fn))
        ast = json.loads(raw)
        base = json.loads((d / "ast-base.json").read_bytes())
        write_md(d, f, fn, role, edit, cmd, ast, base, mode)
        subprocess.run(["python3", "tools/logic-map/risk_pattern_report.py", f, "--output", str(d / "risk-pattern-report.md")],
                       cwd=ROOT, check=True, capture_output=True)
        print(d.name, len(ast.get("branches") or []))


if __name__ == "__main__":
    main(sys.argv[1])
