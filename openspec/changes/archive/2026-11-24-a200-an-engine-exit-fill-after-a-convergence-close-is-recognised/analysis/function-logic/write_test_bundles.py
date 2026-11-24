#!/usr/bin/env python3
"""a200 2.4 — 1단계 재현 시험 파일에서 지우거나(뒤집혀 새 파일로 옮김) 자리만 바뀐 시험 함수의 FLM/BTM 번들을 씀.

시험 함수의 분기는 전부 「단언 실패면 t.Fatalf」 모양이라 손으로 고를 것이 없다 — AST(ast.json)의 분기 목록을 그대로 표로 옮긴다.
지운 함수는 `revision: base` 번들(ast.json 은 base 28680af3d 의 같은 도구로 뽑음), 남은 함수는 `revision: current`.
`python3 <이 파일>` (저장소 어디서든). risk-pattern-report.md 는 따로 `risk_pattern_report.py` 로 씀.
"""
import json
import pathlib

HERE = pathlib.Path(__file__).resolve().parent
SOURCE = "internal/journal/a200_convergence_close_exit_fill_repro_test.go"

# 함수 → (판본, 뒤를 잇는 시험, 한 줄 설명)
BUNDLES = {
    "TestA200CurrentBehaviourAConvergenceCloseThenTheStopFillIsAnAttributionFailure": (
        "base", "TestA200AStopFillAfterAConvergenceCloseIsRecognised",
        "본 사고 모양의 편집 전 동작(ATTRIBUTION_FAILED X59) 고정 — 새 계약(인정)으로 뒤집혀 지움"),
    "TestA200CurrentBehaviourAnAdoptedPositionsOutcomeIsRefusedAsAPostAdoptionDisposal": (
        "base", "TestA200AnAdoptedStopFillIsRecognisedWithoutAnOutcome",
        "편입 포지션 같은 모양의 편집 전 동작 + 성과 탐침 — 새 계약으로 뒤집혀 지움(탐침은 새 시험이 이어받음)"),
    "TestA200CurrentBehaviourPartialStopFillsAfterAConvergenceClose": (
        "base", "TestA200PartialStopFillsAfterAConvergenceCloseAreEachRecognised",
        "조정 뒤 부분 체결 둘의 편집 전 동작(X58 행 1 · 둘째 무기록) — 새 계약으로 뒤집혀 지움"),
    "TestA200CurrentBehaviourTheSameStopFilledBeforeAnyAdjustmentFreezesTheOutcome": (
        "current", "TestA200CurrentBehaviourTheSameStopFilledBeforeAnyAdjustmentFreezesTheOutcome",
        "W0(조정 없이 체결 먼저) 정상 청산 기준 — 본문 무변, 앞의 함수를 지워 자리(줄)만 옮겨짐"),
    # impl-r1 P2-3: reconcile 재현 파일의 무관리 W4 · W5 는 a200 뒤에도 맞는 계약이라 이름만 사실대로 바꿈(본문 무변).
    "TestA200CurrentBehaviourW4FillBetweenConvergenceAndRefresh": (
        "base", "TestA200AnUnmanagedW4FillStaysAnAttributionFailure",
        "무관리 W4 의 귀속 실패 — 이름만 바꿈(본문 무변, 새 이름 시험이 같은 단언)"),
    "TestA200CurrentBehaviourW5FillBetweenRefreshAndPersist": (
        "base", "TestA200AnUnmanagedW5FillStaysAnAttributionFailure",
        "무관리 W5 의 귀속 실패 · Observe 오류 — 이름만 바꿈(본문 무변)"),
}


def main() -> None:
    for fn, (revision, successor, why) in BUNDLES.items():
        target = next(HERE.glob(f"*--{fn.lower()}"))
        ast = json.loads((target / "ast.json").read_text(encoding="utf-8"))
        SOURCE = ast["file"]
        assert ast.get("revision", "current") == revision, (fn, ast.get("revision"))
        start, end = ast["start"]["line"], ast["end"]["line"]
        branches = ast.get("branches") or []
        label = "base 28680af3d" if revision == "base" else "작업 트리"
        rows = []
        btm = []
        for b in branches:
            coord = f"{b['id']} {b['kind']}@{b['at']['line']}"
            rows.append(f"| {coord} | 단언 조건(원장 관측값) | 없음 | 참이면 t.Fatalf / t.Logf | 시험 자신 |")
            red = "해당 없음(시험 함수)"
            green = "지움 — 뒤를 잇는 시험이 통과" if revision == "base" else "yes — 본문 무변, PASS(`../../verify/selected-tests.txt`)"
            btm.append(f"| {b['id']} | 단언 갈래 | `{successor}` | {red} | {green} |")
        flm = f"""# Function Logic Map: `{fn}` ({'지운 base 판' if revision == 'base' else '자리만 옮겨진 판'})

- Source: `{SOURCE}` ({start}-{end}), {label}, source sha256 `{ast['source_sha256'][:12]}…`.
- AST evidence: `ast.json`(revision: {revision}, 분기 {len(branches)}).
- Risk scan: `risk-pattern-report.md`.
- 처분: {why}. 뒤를 잇는 시험 `{successor}`(internal/journal/a200_recognition_test.go 또는 같은 파일).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 임시 원장 · 운영 훅 리터럴 · 합성 픽스처 | `t.TempDir()` | `a200Journal` | 단언 실패면 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
{chr(10).join(rows) if rows else '| B1 | 분기 없음 | 없음 | 정상 종료 | 시험 자신 |'}

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 1단계 하네스(a200Journal · a200ArmStop · a200Converge · readA200Ledger) | 원장 모양 만들기 · 읽기 | 오류면 t.Fatal | AST |

## State mutations and fallbacks

- 시험 함수 — 생산 상태 변경 없음(임시 원장만).

## Safety conclusion

- Safe edit boundary: 시험 전용. 생산 코드 영향 없음.
- High-risk impact: no(시험 함수 — 다만 지운 기대는 새 계약 시험이 이어받음).
"""
        # 지운 함수의 이름은 표에 쓰지 않는다 — 게이트는 표에 적힌 시험 이름이 추적 파일에 있는지 보고, 지운 이름은 없다.
        title = f"`{fn}`" if revision == "current" else "지운 base 함수(이름은 function-logic-map.md · ast.json)"
        bmap = f"""# Branch Test Map: {title}

- Source: `{SOURCE}` ({start}-{end}), {label}.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
{chr(10).join(btm) if btm else f'| B1 | 정상 종료 | `{successor}` | 해당 없음 | yes |'}
"""
        (target / "function-logic-map.md").write_text(flm, encoding="utf-8")
        (target / "branch-test-map.md").write_text(bmap, encoding="utf-8")
        print("wrote", target.name, revision, len(branches))


if __name__ == "__main__":
    main()
