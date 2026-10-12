#!/usr/bin/env python3
"""a131 FLM/BTM 작성기 — ast.json 의 분기·반환 좌표를 **그대로** 표에 옮기고(손으로 좌표를 적지 않음),
함수별 산문(역할·입력·편집 경계·시험)은 아래 NOTES 에서 붙인다. 좌표 칸에는 소스 줄 원문을 붙여 사람이 대조할 수 있게 함.

사용: python3 flm_write.py <repo-root> [phase]   (phase: before|after — 머리말 문구만 바뀜)
"""
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(sys.argv[1]).resolve()
PHASE = sys.argv[2] if len(sys.argv) > 2 else "before"
CHANGE = ROOT / "openspec/changes/a131-no-account-in-any-log/analysis/function-logic"
NOTES = json.loads((Path(__file__).with_name("flm_notes.json")).read_text(encoding="utf-8"))


def line_text(src_lines, n):
    t = src_lines[n - 1].strip() if 0 < n <= len(src_lines) else ""
    return t.replace("|", "\\|").replace("`", "'")[:110]


for bundle in sorted(CHANGE.iterdir()):
    ast_path = bundle / "ast.json"
    if not ast_path.is_file():
        continue
    a = json.loads(ast_path.read_text())
    key = bundle.name
    note = NOTES[key]
    if PHASE == "before":
        # 편집 전 번들의 좌표는 base 의 줄 — 소스 줄 원문도 base 에서 읽음(작업 트리는 이미 밀렸을 수 있음).
        base = (ROOT / "openspec/changes/a131-no-account-in-any-log/base-commit.txt").read_text().strip()
        src = subprocess.run(["git", "-C", str(ROOT), "show", f"{base}:{a['file']}"], capture_output=True,
                             text=True, check=True).stdout.splitlines()
    else:
        src = (ROOT / a["file"]).read_text(encoding="utf-8").splitlines()
    qual = (a.get("receiver", "") + "." if a.get("receiver") else "") + a["function"]
    branches = a.get("branches") or []
    returns = a.get("returns") or []
    s, e = a["start"]["line"], a["end"]["line"]
    stage = "편집 전(base)" if PHASE == "before" else "편집 뒤"
    out = []
    out.append(f"# Function Logic Map: `{qual}`\n")
    out.append(f"- Source: `{a['file']}` ({s}-{e})")
    out.append(f"- Qualified: `{qual}`")
    out.append(f"- AST evidence: `ast.json` ({stage}, `source_sha256` {a['source_sha256'][:16]}…) — AST branches {len(branches)} · returns {len(returns)}")
    out.append("- Risk scan: `risk-pattern-report.md`")
    out.append(f"- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)\n")
    out.append(f"**역할.** {note['role']}\n")
    out.append("## Inputs and invariants\n")
    out.append("| Input/state | Valid range | Source of truth | Failure behavior |")
    out.append("|---|---|---|---|")
    for row in note["inputs"]:
        out.append("| " + " | ".join(row) + " |")
    out.append("\n## Branches and early returns\n")
    out.append("| Branch | AST kind | Source line | a131 |")
    out.append("|---|---|---|---|")
    edited = set(note.get("edited_branches", []))
    if not branches:
        out.append(f"| B1 | branchless happy path at {s}:1 | `{line_text(src, s)}` | {note.get('b1', '본문 전체')} |")
    for b in branches:
        ln, col = b["at"]["line"], b["at"]["column"]
        mark = "편집 영향" if b["id"] in edited else "무편집"
        out.append(f"| {b['id']} | {b['kind']} at {ln}:{col} | `{line_text(src, ln)}` | {mark} |")
    out.append("\nExact AST return positions: " + ", ".join(f"{r['at']['line']}:{r['at']['column']}" for r in returns) if returns else "\n반환문 없음(AST returns 0).")
    out.append("\n## Calls and live bindings\n")
    out.append("| Callee | Why called | Error/timeout/retry contract | Evidence |")
    out.append("|---|---|---|---|")
    for row in note["calls"]:
        out.append("| " + " | ".join(row) + " |")
    out.append("\n## State mutations and fallbacks\n")
    for m in note["mutations"]:
        out.append(f"- {m}")
    out.append("\n## Safety conclusion\n")
    out.append(f"- Safe edit boundary: {note['boundary']}")
    out.append(f"- High-risk impact: {note['highrisk']}")
    (bundle / "function-logic-map.md").write_text("\n".join(out) + "\n", encoding="utf-8")

    t = []
    t.append(f"# Branch Test Map: `{qual}`\n")
    t.append(f"- Source: `{a['file']}` ({s}-{e}) — {stage}\n")
    t.append("| Branch | Anchor | Scenario | Test | a131 |")
    t.append("|---|---|---|---|---|")
    tests = note.get("tests", {})
    default = note.get("default_test", "a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify)")
    if not branches:
        t.append(f"| B1 | branchless happy path at {s}:1 | 본문 | {tests.get('B1', default)} | 편집 |")
    for b in branches:
        ln, col = b["at"]["line"], b["at"]["column"]
        t.append(f"| {b['id']} | {b['kind']} at {ln}:{col} | `{line_text(src, ln)}` | {tests.get(b['id'], default)} | {'편집 영향' if b['id'] in edited else '무편집'} |")
    t.append("\n## a131 시험\n")
    for x in note.get("a131_tests", []):
        t.append(f"- {x}")
    (bundle / "branch-test-map.md").write_text("\n".join(t) + "\n", encoding="utf-8")
    print("wrote", bundle.name)
