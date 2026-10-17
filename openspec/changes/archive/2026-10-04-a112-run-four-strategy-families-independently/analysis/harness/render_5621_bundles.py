#!/usr/bin/env python3
"""a112 5.6.2.1 — 편집 뒤 번들 셋(NewStrategyEntrySupervisor · Context.NewRefreshingPairedStrategyEntrySupervisor ·
execgw.AllReasonCodes)의 BTM/FLM 을 측정값(ast.json · coverage-post-5.6.2.1-engine.json)에서 채운다. 분기 좌표 · 시험 · 블록은
손으로 쓰지 않는다 — 저자가 쓰는 것은 분기의 뜻과 편집 설명뿐이다."""
from __future__ import annotations

import json
from pathlib import Path

A = Path(__file__).resolve().parents[1]
F = A / "function-logic"
COVREF = "`analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`"
cov = json.loads((A / "measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json").read_text())

SPECS = {
    "internal-app-engine--newstrategyentrysupervisor": dict(
        edit="(5.6.2.1) 반환하는 감독자 리터럴에 `entry: opts.EntryGate` 한 칸 — 분기 · 검증 규칙 불변(분기 18 전후 동일). "
             "`StrategyEntrySupervisorOptions.EntryGate`(좁은 인터페이스 `StrategyEntryBlocker` — `Block` 하나)를 감독자에 넘긴다.",
        red="해당 없음(분기 불변) — 변이 E08(칸 누락) CAUGHT",
        scen={}, safety="분기 · 검증 불변. 감독자가 진입을 닫을 수만 있고(열 수단 없음) 그 수단은 호출자가 준다.",
    ),
    "internal-app-engine--context.newrefreshingpairedstrategyentrysupervisor": dict(
        edit="(5.6.2.1) 새 B2 `if c.Entry == nil` — 진입 게이트 없는 Context 에서는 생산 감독자를 만들지 않는다(그 조립에서는 중앙 무결성 "
             "고장이 진입이 아니라 프로세스를 닫게 되므로). 감독자 옵션에 `EntryGate: c.Entry`. 편집 전 B2 · B3 → B3 · B4.",
        red="B2: `red-5.6.2.1.log`(`TestTheProductionStrategySupervisorRefusesAContextWithoutAnEntryGate` FAIL) · 변이 E07; "
            "옵션 전달: `TestTheProductionStrategySupervisorBlocksOnTheEnginesOwnEntryGate` FAIL · 변이 E06",
        scen={"B1": "nil Context · nil 시계 → 거절", "B2": "(새) 진입 게이트 없음 → `ErrRuntimeUnavailable`",
              "B3": "KR · US 권한 갱신 전용 worker 둘", "B4": "감독자 생성 실패 → 오류"},
        safety="생산 기동 순서에서 이 생성자 앞의 `Recovery` 가 이미 같은 게이트를 요구하므로(runtime_wiring.go) 생산 기동 동작 변화 0.",
    ),
    "internal-execgw--allreasoncodes": dict(
        edit="(5.6.2.1, 커밋 `3260f4eb`) 열거에 `ReasonStrategyCentralIntegrity` 한 줄 — 분기 없음. 골든 재생성 · a098 census 한 줄.",
        red="변이 E10(등록 누락) CAUGHT — `TestReasonCodeEnumIsStable`",
        scen={}, safety="어휘 추가(이름 바꾸기 아님) — 기존 기록 불변.",
    ),
}


def main() -> None:
    for bundle, spec in SPECS.items():
        ast = json.loads((F / bundle / "ast.json").read_text())
        rows = cov["bundles"].get(bundle, {})
        fn = f"{ast['receiver']}.{ast['function']}" if ast.get("receiver") else ast["function"]
        head = [f"# Branch Test Map: `{fn}`", "",
                f"- Source SHA-256: `{ast['source_sha256']}`; AST branch locations are authoritative.",
                f"- Revision: **modified (태스크 5.6.2.1, 2026-09-30).** {spec['edit']} 편집 전 번들은 "
                "`analysis/measurements/lot-5.6.2-5.2.2/pre-edit/`.",
                f"- 측정: {COVREF}(격리 사본, `./internal/app/engine` 시험 {cov['tests_run']}개를 하나씩)." if rows else
                "- 측정: 분기 없음 — 행동 증거는 RED 칸(`TestReasonCodeEnumIsStable` · 순서 시험).", "",
                "| Branch | Scenario anchor | Test | RED observed | GREEN observed |", "|---|---|---|---|---|"]
        flm_rows = []
        for br in ast.get("branches") or []:
            r = rows.get(br["id"], {"block": None, "entered_by": []})
            e = r["entered_by"]
            tests = ", ".join(f"`{x}`" for x in e[:2]) + (f" 외 {len(e) - 2}" if len(e) > 2 else "") if e else "(측정 표본 0)"
            green = f"yes (block {r['block']}, 시험 {len(e)}개)" if r["block"] and e else "측정 표본의 시험 0개(블록 좌표 없거나 미실행)"
            red = spec["red"] if br["id"] in ("B2",) and "B2" in spec["scen"] else "해당 없음(분기 불변)"
            s = spec["scen"].get(br["id"], "(분기 불변 — 편집 전 번들의 서술 그대로)")
            head.append(f"| {br['id']} | {br['kind']} at {br['at']['line']}:{br['at']['column']} — {s} | {tests} | {red} | {green} |")
            flm_rows.append(f"| {br['id']} | {br['kind']} (:{br['at']['line']}) | {s} | — | {tests} |")
        if not ast.get("branches"):
            head.append(f"| B1 | happy path — 분기 없는 열거 함수(정렬된 전체 목록) | `TestReasonCodeEnumIsStable` | {spec['red']} | 통과 |")
        (F / bundle / "branch-test-map.md").write_text("\n".join(head) + "\n")
        flm = [f"# Function Logic Map: `{fn}`", "", f"- Source: `{ast['file']}`",
               f"- AST evidence: `ast.json` — **편집 뒤**, :{ast['start']['line']}–{ast['end']['line']}, 분기 {len(ast.get('branches') or [])}, "
               f"source_sha256 `{ast['source_sha256'][:12]}…`.",
               "- Risk scan: `risk-pattern-report.md`", f"- 편집: {spec['edit']}", "",
               "## Inputs and invariants", "", "| Input/state | Valid range | Source of truth | Failure behavior |", "|---|---|---|---|",
               "| 편집이 더한 입력 | 진입 게이트(`*execgw.EntryGate`) 또는 없음 | 엔진 조립(`Context.Entry`) | 위 편집 설명 |", "",
               "## Branches and early returns", "", "| Branch | Condition | Mutation/side effect | Return/error | Required test |",
               "|---|---|---|---|---|"] + (flm_rows or ["| B1 | happy path(분기 없음) | 열거 | 정렬된 목록 | `TestReasonCodeEnumIsStable` |"]) + [
               "", "## Calls and live bindings", "", "| Callee expression | Position |", "|---|---|"] + [
               f"| `{c['text']}` | {c['at']['line']}:{c['at']['column']} |" for c in ast.get("calls") or []] + [
               "", "## State mutations and fallbacks", "", "- 위 편집 설명 외 없음.", "",
               "## Safety conclusion", "", f"- Safe edit boundary: {spec['safety']}",
               "- High-risk impact: yes(진입 게이트 경로) — 편집은 진입을 닫는 방향만 더함."]
        (F / bundle / "function-logic-map.md").write_text("\n".join(flm) + "\n")
        print("rendered", bundle)


if __name__ == "__main__":
    main()
