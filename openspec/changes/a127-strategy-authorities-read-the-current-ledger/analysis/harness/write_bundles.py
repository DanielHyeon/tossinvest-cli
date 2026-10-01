#!/usr/bin/env python3
"""a127 FLM · BTM 작성기 — 분기 표는 branch_table.py 가 ast · 소스 · 커버리지로 만들고, 산문(역할 · 입력 · 호출 · 상태 · 안전 · 분기별 시험)은 아래 사전.
사용(저장소 루트): python3 <this> <coverprofile> <pre|post>"""
import json, subprocess, sys
from pathlib import Path

C = Path("openspec/changes/a127-strategy-authorities-read-the-current-ledger/analysis")
PROSE = json.loads((C / "harness" / "bundle_prose.json").read_text())

def main():
    profile, phase = sys.argv[1], sys.argv[2]
    for d in sorted((C / "function-logic").iterdir()):
        name = d.name
        if name not in PROSE:
            continue
        p = PROSE[name]
        ast = json.loads((d / "ast.json").read_text())
        table = subprocess.run(["python3", str(C / "harness" / "branch_table.py"), str(d), profile], capture_output=True, text=True, check=True).stdout
        rows = [l for l in table.splitlines()[2:] if l.startswith("| B")]
        b = len(ast.get("branches") or []); r = len(ast.get("returns") or []); c = len(ast.get("calls") or [])
        when = "**편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)" if phase == "pre" else "**편집 뒤**(구현 로트, 커버리지 `analysis/impl/coverage-post-edit.out`)"
        flm = [f"# Function Logic Map: `{ast['function'] if not ast.get('receiver') else ast['receiver'] + '.' + ast['function']}`", "",
               f"- Source: `{ast['file']}` (`{ast['start']['line']}`–`{ast['end']['line']}`)",
               f"- Qualified: `{p['qualified']}`",
               f"- AST evidence: `ast.json` (`source_sha256` {ast['source_sha256'][:16]}…) — {when}",
               "- Risk scan: `risk-pattern-report.md`",
               f"- AST branches {b} · return {r} · 호출 {c}", "",
               f"**역할.** {p['role']}", "",
               "## Inputs and invariants", "", "| Input/state | Valid range | Source of truth | Failure behavior |", "|---|---|---|---|"] + p["inputs"] + ["",
               "## Branches and early returns", "",
               "> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.", "",
               table.rstrip(), "", "## Calls and live bindings", "", p.get("calls_" + phase, p["calls"]), "", "## State mutations and fallbacks", "", p["state"], "",
               "## Safety conclusion", "", f"- **Safe edit boundary**: {p['boundary_' + phase]}", f"- **High-risk impact**: {p['highrisk']}"]
        (d / "function-logic-map.md").write_text("\n".join(flm) + "\n")
        tests = p.get("tests_" + phase, {})
        btm = [f"# Branch Test Map: `{p['qualified']}`", "", p.get("btm_note_" + phase, ""), "",
               "| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |", "|---|---|---|---|---|---|"]
        for row in rows:
            cells = [x.strip() for x in row.strip("|").split("|")]
            bid, cond, entered = cells[0], cells[2], cells[4]
            t = tests.get(bid, tests.get("*", "기존 — 이 change 가 편집하지 않는 분기"))
            red = "yes" if bid in p.get("red_" + phase, []) else "n/a"
            btm.append(f"| {bid} | {cond} | {entered} | {t} | {red} | {'yes' if bid in p.get('red_' + phase, []) else 'n/a'} |")
        (d / "branch-test-map.md").write_text("\n".join(btm) + "\n")
        print(name, b, r, c)

if __name__ == "__main__":
    main()
