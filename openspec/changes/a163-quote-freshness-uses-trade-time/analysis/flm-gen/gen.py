"""a163 FLM 번들 Markdown 생성기(스크래치) — AST 분기 좌표·원문은 ast.json 과 소스에서 기계적으로 읽고, 의미 칸만 손으로 씀."""
import json, sys
from pathlib import Path

ROOT = Path(sys.argv[1])
FL = ROOT / "openspec/changes/a163-quote-freshness-uses-trade-time/analysis/function-logic"
spec = json.loads(Path(sys.argv[2]).read_text())

for name, s in spec.items():
    d = FL / name
    v = json.loads((d / "ast.json").read_text())
    lines = (ROOT / v["file"]).read_text().splitlines()
    fn = (v.get("receiver") + "." if v.get("receiver") else "") + v["function"]
    br = v.get("branches") or []
    rets = v.get("returns") or []
    out = [f"# Function Logic Map: `{fn}`", "",
           f"- Source: `{v['file']}` ({v['start']['line']}-{v['end']['line']})",
           "- AST evidence: `ast.json`",
           "- Risk scan: `risk-pattern-report.md`",
           f"- 기준: {s['revision_note']}",
           f"- AST 분기 {len(br)} · 반환 {len(rets)} · defer {len(v.get('defers') or [])} · go {len(v.get('go_statements') or [])}",
           f"- 위험 등급: {s['risk']}", "",
           "## Inputs and invariants", "",
           "| Input/state | Valid range | Source of truth | Failure behavior |", "|---|---|---|---|"]
    out += [f"| {a} | {b} | {c} | {e} |" for a, b, c, e in s["inputs"]]
    out += ["", "## Branches and early returns", "",
            "| Branch | Condition | Mutation/side effect | Return/error | Required test |", "|---|---|---|---|---|"]
    for b in br:
        bid = b["id"]
        row = s["branches"].get(bid, ["—", "—", "기존 시험"])
        text = lines[b["at"]["line"] - 1].strip().replace("|", "\\|")
        out.append(f"| {bid} | {b['kind']} {b['at']['line']} `{text}` | {row[0]} | {row[1]} | {row[2]} |")
    out += ["", "반환 좌표(AST): " + " · ".join(str(r["at"]["line"]) for r in rets) if rets else "반환 좌표(AST): 없음"]
    out += ["", "## Calls and live bindings", "",
            "| Callee | Why called | Error/timeout/retry contract | Evidence |", "|---|---|---|---|"]
    out += [f"| {a} | {b} | {c} | {e} |" for a, b, c, e in s["calls"]]
    out += ["", "## State mutations and fallbacks", ""] + [f"- {x}" for x in s["mutations"]]
    out += ["", "## a163 편집", ""] + [f"- {x}" for x in s["edit"]]
    out += ["", "## Safety conclusion", ""] + [f"- {x}" for x in s["safety"]]
    (d / "function-logic-map.md").write_text("\n".join(out) + "\n")

    bt = [f"# Branch Test Map: `{fn}`", "", s.get("btm_note", ""), "",
          "| Branch | Scenario | Test | RED observed | GREEN observed |", "|---|---|---|---|---|"]
    for b in br:
        bid = b["id"]
        row = s["tests"].get(bid) or s["tests"]["*"]
        bt.append(f"| {bid} | {b['kind']} {b['at']['line']} — {row[0]} | {row[1]} | {row[2]} | {row[3]} |")
    if not br:
        row = s["tests"]["*"]
        bt.append(f"| B1 | branchless happy path — {row[0]} | {row[1]} | {row[2]} | {row[3]} |")
    (d / "branch-test-map.md").write_text("\n".join(bt) + "\n")
    print("wrote", name)
