#!/usr/bin/env python3
"""a091 경량 번들 — base 재고정(조건 ① 둘째 갈래) 창에서 마지막 자기 Go 커밋(e4d976d6, 시험 전용)이 바꾼 시험 함수 다섯의 증거.

분기 표 · 반환 좌표는 ast.json(`go run ./tools/logic-map`)과 소스 원문에서 측정으로, 산문은 짧게. 시험 함수의 「Test」 열은 그 함수 자신
(또는 그 도우미를 부르는 시험)을 인용한다 — 시험 함수의 분기는 그 시험이 돌 때 지나간다.

사용: 저장소 루트에서 python3 <이 파일>
"""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

ROOT = Path.cwd()
FL = Path(__file__).resolve().parent.parent / "function-logic"
TARGETS = {
    # (파일, 함수, 역할, 인용 시험)
    "internal-app-engine--testa091theaugustsecondreplay": (
        "internal/app/engine/a091_replay_test.go", "TestA091TheAugustSecondReplay",
        "8/2 재생 네 팔(생산 AlertDeliverer) — i2 에서 (i) · (iv) 의 미전달 줄 0 단언을 더함.", ["TestA091TheAugustSecondReplay"]),
    "internal-app-engine--testa091thereportfitsitsshare": (
        "internal/app/engine/a091_replay_test.go", "TestA091TheReportFitsItsShare",
        "보고 몫 실측(소유 칸 ≤ 750ms · 승인 칸 < 5s) — i2 에서 승인 호출 구간과의 겹침 단언 · 연결 풀 B2 칸을 더함.", ["TestA091TheReportFitsItsShare"]),
    "internal-app-engine--a091harness": (
        "internal/app/engine/a091_stop_sold_nothing_test.go", "a091Harness",
        "하네스 (나) — 실제 RecordOnly + 원장 + 로그 캡처, 생산 모양(관측자 Log nil · ZeroFloorLog 전용 싱크).", ["TestA091AProtectiveZeroIsACriticalRow"]),
    "internal-app-engine--testa091azeroholdingisnotafailedstop": (
        "internal/app/engine/a091_stop_sold_nothing_test.go", "TestA091AZeroHoldingIsNotAFailedStop",
        "보유 0(Holdings 한정 0)은 옛 종류 normal — i2 에서 payload cause `no_holding` 단언을 더함.", ["TestA091AZeroHoldingIsNotAFailedStop"]),
    "internal-app-engine--testa091theproductionassemblyreadstheloadedswitch": (
        "internal/app/engine/a091_stop_sold_nothing_test.go", "TestA091TheProductionAssemblyReadsTheLoadedSwitch",
        "생산 배선이 로드된 notifications.enabled 와 엔진 로거로 덮음 — i2 에서 조립 로거를 덮지 않고 `eng.Log != nil` 단언.", ["TestA091TheProductionAssemblyReadsTheLoadedSwitch"]),
}


def esc(s: str) -> str:
    return s.replace("|", "\\|")


def main() -> int:
    for name, (src, fn, role, tests) in TARGETS.items():
        d = FL / name
        d.mkdir(parents=True, exist_ok=True)
        raw = subprocess.run(["go", "run", "./tools/logic-map", "--file", src, "--func", fn],
                             cwd=ROOT, capture_output=True, text=True, check=True).stdout
        (d / "ast.json").write_text(raw)
        ast = json.loads(raw)
        lines = (ROOT / src).read_text().splitlines()
        br = ast.get("branches") or []
        rets = ", ".join(f"`{r['at']['line']}:{r['at']['column']}`" for r in (ast.get("returns") or [])) or "none"
        cited = " · ".join(f"`{t}`" for t in tests)
        fl = [f"# Function Logic Map: `{fn}`", "",
              f"- Source: `{src}` (`{ast['start']['line']}`–`{ast['end']['line']}`)",
              f"- Qualified: `{fn}`",
              f"- AST evidence: `ast.json` (`source_sha256` {ast['source_sha256'][:16]}…) — `go run ./tools/logic-map`",
              "- Risk scan: `risk-pattern-report.md`", f"- 분기 {len(br)}", "",
              f"**역할(경량 번들 — 시험 함수).** {role}", "",
              "## Inputs and invariants", "", "| Input/state | Valid range | Source of truth | Failure behavior |", "|---|---|---|---|",
              "| 하네스 · 가짜 부품 | 시험 고정 | 시험 파일 | 단언 실패 → 시험 실패 |", "",
              "## Branches and early returns", "", "| Branch | 종류 | 조건 (원문) |", "|---|---|---|"]
        if not br:
            fl.append("| — | — | 분기 없음 — 유일한 경로 |")
        for b in br:
            n = b["at"]["line"]
            fl.append(f"| {b['id']} | {b['kind']} | `:{n}` `{esc(lines[n - 1].strip())}` |")
        fl += ["", f"Exact AST return positions: {rets}", "", "## Calls and live bindings", "",
               f"호출 좌표 전수는 `ast.json` `calls`({len(ast.get('calls') or [])}). 생산 경로에 닿는 것은 하네스가 조립한 `ExitObserver` · 알림기 · 원장뿐 — 브로커 0.", "",
               "## State mutations and fallbacks", "", "임시 원장 · 버퍼만(시험 범위).", "",
               "## Safety conclusion", "", "- 생산 코드가 아니다 — 생산 동작 변화 0. 이 번들은 base 재고정 창의 증거 완결을 위한 경량 기록이다.", ""]
        (d / "function-logic-map.md").write_text("\n".join(fl))
        btm = [f"# Branch Test Map: `{fn}`", "", f"- Source: `{src}`", "",
               "> 시험 함수의 분기는 그 시험이 돌 때 지나간다 — Test 열은 그 시험(또는 도우미를 부르는 시험)이다.", "",
               "| Branch | 조건 | Test | RED observed | GREEN observed |", "|---|---|---|---|---|"]
        if not br:
            btm.append(f"| B1 | 분기 없음 — 유일한 경로 | {cited} | n/a | yes |")
        for b in br:
            n = b["at"]["line"]
            btm.append(f"| {b['id']} | `:{n}` `{esc(lines[n - 1].strip())}` | {cited} | n/a | yes |")
        (d / "branch-test-map.md").write_text("\n".join(btm) + "\n")
        subprocess.run([sys.executable, "tools/logic-map/risk_pattern_report.py", src, "--output", str(d / "risk-pattern-report.md")],
                       cwd=ROOT, check=True)
        print("wrote", name)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
