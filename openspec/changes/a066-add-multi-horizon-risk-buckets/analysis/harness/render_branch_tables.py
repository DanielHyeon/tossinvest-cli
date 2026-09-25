#!/usr/bin/env python3
"""branch_coverage_rows.py 의 JSONL 을 Branch Test Map 과 FLM 분기 표로 렌더링함.

사용: python3 render_branch_tables.py <rows.jsonl> <function> <suite-command> <coverage-head> <btm-out> <flm-rows-out>

손으로 쓰는 칸은 annotations 의 시나리오·시험 이름뿐이고, 나머지(조건·도입 커밋·커버리지)는 측정값임.
"""
import json
import sys


def main() -> None:
    rows_path, function, suite, head, btm_out, flm_out = sys.argv[1:7]
    rows = [json.loads(line) for line in open(rows_path, encoding="utf-8") if line.strip()]
    btm = [f"# Branch Test Map: `{function}`", "",
           f"Measured at HEAD `{head}` with `{suite}` (statement coverage; `covered` = the branch body ran at "
           "least once in the package suite, it does not say which test). Named tests are attributed per test "
           "with a single-test `-coverprofile` run. Harness: `analysis/harness/branch_coverage_rows.py`.", "",
           "| Branch | Scenario | Test | RED observed | GREEN observed |", "|---|---|---|---|---|"]
    flm = []
    for row in rows:
        # 좌표 칸은 `kind at L:C — 산문` 모양으로 둠 — role_check 가 이 칸을 AST 분기 좌표와 순서대로 대조함
        then = f"; then `{row['then']}`" if row.get("then") else ""
        source_fact = f"`{row['text']}`{then} (line last changed by `{row['commit']}`)"
        scenario = f"{row['kind']} at {row['at']} — " + (row["scenario"] or source_fact)
        test = row["tests"] or f"package suite `{suite}`"
        if row["red"]:
            red = row["red"]
        elif row.get("a066"):
            red = f"a066 commit `{row['commit']}` — no RED recorded for this row"
        else:
            red = f"n/a — branch line last changed by `{row['commit']}`, not an a066 commit"
        btm.append(f"| {row['id']} | {scenario} | {test} | {red} | {row['coverage']} at `{head}` |")
        relevance = ("a066: " + row["scenario"]) if row["scenario"] else ("a066 commit" if row.get("a066") else "not a066")
        flm.append(f"| {row['id']} | {row['kind']} at {row['at']} | {source_fact} | {relevance} | {row['coverage']} |")
    open(btm_out, "w", encoding="utf-8").write("\n".join(btm) + "\n")
    open(flm_out, "w", encoding="utf-8").write("\n".join(flm) + "\n")


if __name__ == "__main__":
    main()
