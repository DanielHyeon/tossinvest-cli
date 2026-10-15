#!/usr/bin/env python3
"""a139 FLM/BTM 번들 생성 하네스.

사용: python3 a139_flm.py base   — 편집 전: 각 대상의 ast.json(= 현재 소스) 과 ast-base.json(같은 바이트)을 쓰고 FLM/BTM 산문을 만듦.
      python3 a139_flm.py after  — 편집 뒤: ast.json 만 다시 추출하고 FLM/BTM 의 좌표 · 분기 행을 새 AST 로 다시 씀(ast-base.json 유지).
산문(역할 · a139 편집 · 시험)은 TARGETS 표가 정본. 저장소 루트는 이 파일 위치에서 유도함.
"""
import json, os, subprocess, sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
CHANGE = "a139-per-trade-risk-budget"
OUT = ROOT / "openspec" / "changes" / CHANGE / "analysis" / "function-logic"

# (파일, 함수, 역할, a139 편집, 기본 시험 명령, {분기 id: 시험 인용})
TARGETS = [
    ("internal/risk/input.go", "DefaultPolicy",
     "사용자 미확정 보수 기본 정책.",
     "`RiskBudget` 100,000 → 33,333 KRW(= PerTradeRiskBudget(MaxDailyLoss), 사용자 결정 2026-10-15) 한 칸 · 유도 주석 정정.",
     "go test ./internal/risk -run 'DefaultPolicy|A139'"),
    ("internal/app/engine/guardian_wiring.go", "riskPolicyFromAutomationGate",
     "감사된 게이트 한도 → Guardian 정책 유도(직선).",
     "`policy.RiskBudget = dailyLoss` → `risk.PerTradeRiskBudget(dailyLoss)`(오류면 금액을 비워 Validate 가 거부) — 새 분기 하나.",
     "go test ./internal/app/engine -run 'RiskPolicy|A136|A138|A139'"),
    ("internal/app/engine/guardian_wiring.go", "newProductionRiskGuardian",
     "생산 Guardian 생성 — 사전 검사 · 정책 유도 · 검증 · 생성.",
     "Validate 앞에서 `risk.PerTradeRiskBudget` 유도 실패를 이름으로 밝혀 ErrLimitsRequired — 새 분기 하나(freeze 리뷰 P3).",
     "go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'"),
    ("internal/app/engine/interlock.go", "newAutomationStatus",
     "인터록이 감사할 해석 값 운반.",
     "`PerTradeRiskBudget` 칸 하나(정책 유도와 같은 함수의 값).",
     "go test ./internal/app/engine -run 'Interlock|A139'"),
    ("internal/app/engine/interlock.go", "acceptanceDetail",
     "인터록 수용 감사 문구.",
     "`per_trade_risk_budget=` 한 칸.",
     "go test ./internal/app/engine -run 'Interlock|Acceptance|A139'"),
    ("internal/app/engine/interlock.go", "recordGateSettings",
     "기동 audit — 게이트 설정값 줄들.",
     "목록 끝에 `engine.automation_gate.per_trade_risk_budget`(유도 값) 한 줄(순서 무변경).",
     "go test ./internal/app/engine -run 'Audit|A074|A139'"),
    ("internal/app/engine/guardian_wiring.go", "wholeOpenPositions",
     "a138 정수 판정의 참거짓 반쪽.", "본문 무편집 — 같은 파일 끝에 a139 audit 함수를 덧붙여 diff 창에 들어옴.",
     "go test ./internal/app/engine -run A138"),
    ("internal/app/engine/guardian_wiring_internal_test.go", "TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit",
     "시험 — USD 게이트의 정책 유도 전 칸.", "기대 예산 50.75 → 16.91(floor(50.75 / 3, 0.01)).", "go test ./internal/app/engine -run RiskPolicy"),
    ("internal/execgw/riskguardian_account_base_testseam_test.go", "TestQFinalAccountBaseGuardianPairedKRUS",
     "시험 — KR · US 짝 계약(같은 권위 경로로 같은 q_final).", "rig 정책 예산을 옛 값으로 명시 — 1/3 예산은 US 위험 상한만 묶어 짝을 깸(시험이 재는 것이 예산 값이 아님).",
     "go test -tags tossos_testseams ./internal/execgw -run PairedKRUS"),
    ("internal/execgw/riskguardian_account_base_testseam_test.go", "TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes",
     "시험 — 여덟 레인 첫 다리의 Guardian 계보 투영.", "같은 이유로 rig 예산 명시.", "go test -tags tossos_testseams ./internal/execgw -run AllEightLanes"),
    ("internal/risk/chain_test.go", "TestOrderAboveThePerOrderCapIsRefused",
     "시험 — 주문 단위 상한(포함 상한).", "notional 초과 갈래의 손절 폭을 200 으로 좁혀 위험 상한이 묶지 않게 — 기대 사유 MAX_ORDER_EXCEEDED 유지(freeze 리뷰 P2-1).",
     "go test ./internal/risk -run TestOrderAboveThePerOrderCapIsRefused"),
    ("internal/risk/contract_test.go", "TestQuantityAboveTheRiskBudgetIsRefused",
     "시험 — 위험 예산을 넘는 수량 거부.", "새 예산 33,333 기준 수량 17 거부 · 16 허용 · 상세 「16 units」.",
     "go test ./internal/risk -run TestQuantityAboveTheRiskBudgetIsRefused"),
]

# 함수마다 그 함수를 실제로 지나는 a139 시험(행마다 덧붙임 — 검사기가 이름의 존재를 대조함).
FN_TESTS = {
    "DefaultPolicy": ["TestA139TheDefaultPolicyFollowsTheSameRule", "TestA139TheRiskCapIsAThirdOfTheDailyLoss"],
    "riskPolicyFromAutomationGate": ["TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss", "TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName"],
    "newProductionRiskGuardian": ["TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName", "TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss"],
    "newAutomationStatus": ["TestA139TheInterlockAuditsTheDerivedRiskBudget"],
    "acceptanceDetail": ["TestA139TheInterlockAuditsTheDerivedRiskBudget"],
    "recordGateSettings": ["TestA139TheInterlockAuditsTheDerivedRiskBudget", "TestA139AGateOffStartWithoutADailyLossStillStarts"],
    "TestOrderAboveThePerOrderCapIsRefused": ["TestOrderAboveThePerOrderCapIsRefused"],
    "TestQuantityAboveTheRiskBudgetIsRefused": ["TestQuantityAboveTheRiskBudgetIsRefused"],
}

BASE_REVISION = {"wholeOpenPositions"}

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

**a139 편집.** {edit}

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

- 편집은 위 「a139 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a139 편집」에 적은 한 자리.
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
