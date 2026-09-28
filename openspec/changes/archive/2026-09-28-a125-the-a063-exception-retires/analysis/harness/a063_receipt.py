#!/usr/bin/env python3
"""a125 4.2 · a063 4.4.1 — a063 재고정 조건 ① 의 **양성 단언** 영수증 (codex freeze C1).

    cd <고정한 연결 워크트리> && python3 <이 파일> [--base <P>]

특례 없는 도구로 a063 을 옛 base P 에서 판정하고, 다음을 **전부** 단언한다(하나라도 틀리면 rc 1):

1. 판정이 끝까지 갔다 — 문맥의 `head` 가 지금 HEAD, `effective_base` 가 P, `landing` 이 "" 이고 required 가 정수다.
2. 귀속: P 뒤 이 change 의 자기 Go 비병합 커밋(`_self_repair_commits` → `_repairs_after`)이 고친 기존 함수마다
   `(소스, 함수) → 번들` 대응이 서고, 번들의 revision 이 창이 요구하는 것(`current_hash` 가 있으면 current, 아니면 base)과 같다.
3. 그 번들들의 **전체** 오류(번들 디렉터리명으로 나오는 해시 · revision · 분기 · 호출 · 시험 인용 줄과, 함수 이름으로 나오는
   누락 줄)가 0 이다.
4. 남는 오류는 전부 `missing evidence for modified function` — 귀속 밖(형제) 함수의 누락뿐이다.

귀속의 알려진 한계(문서 커밋과 Go 커밋을 나눈 경우 · 병합 자신의 변경)는 이 영수증이 못 본다 — 사람이 확인한다.
"""
import argparse
import json
import subprocess
import sys
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("--base", default="da80ce31b6a1ab5d443016768f970a82bab102db")
args = parser.parse_args()
root = Path(".").resolve()
sys.path.insert(0, str(root / "tools" / "logic-map"))
import check_analysis as c  # noqa: E402

CHANGE = "a063-align-attestation-renewal-profile"
head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=root, capture_output=True, text=True, check=True).stdout.strip()
assert (root / "openspec/changes" / CHANGE / "base-commit.txt").read_text().strip() == args.base, "base-commit.txt 가 P 가 아니다"
context: dict[str, object] = {}
errors = c.check(CHANGE, root, context)
failures = []
if not (context.get("head") == head and context.get("effective_base") == args.base and context.get("landing") == ""
        and isinstance(context.get("required_count"), int)):
    failures.append(f"판정이 끝까지 가지 않았거나 창이 다르다: {dict((k, context.get(k)) for k in ('head', 'effective_base', 'landing', 'required_count'))} · {errors[:1]}")
analysis = root / "openspec/changes" / CHANGE / "analysis" / "function-logic"
evidence = c._read_evidence(analysis)
bundles = {}
for path, raw in evidence.held.items():
    value = c._parsed(raw)
    bundles[(c.normalized_source(str(value["file"]), root)[1], c.qualified(value))] = (path.parent.name, value.get("revision", "current"))
attributed = c._repairs_after(root, args.base, c._self_repair_commits(root, analysis, head), head)
owned: dict = {}
for commit in attributed:
    owned.update(c.changed_existing_functions(root, f"{commit}^1", commit))
window = c.changed_existing_functions(root, args.base, "")
rows = []
for key in sorted(owned):
    bundle = bundles.get(key)
    expected = window.get(key)
    want = None if expected is None else ("current" if expected.get("current_hash") else "base")
    rows.append({"function": f"{key[0]}:{key[1]}", "bundle": bundle[0] if bundle else None,
                 "revision": bundle[1] if bundle else None, "window_requires": want})
    if bundle is None:
        failures.append(f"귀속 함수에 번들이 없다: {key}")
    elif want is not None and bundle[1] != want:
        failures.append(f"revision 이 창의 요구와 다르다: {key} {bundle[1]} != {want}")
names = {row["bundle"] for row in rows if row["bundle"]}
own_errors = [e for e in errors if any(e.startswith(n + ":") for n in names) or any(row["function"] in e for row in rows)]
other = [e for e in errors if e not in own_errors]
unclassified = [e for e in other if not e.startswith("missing evidence for modified function ")]
failures += [f"귀속 번들 오류: {e}" for e in own_errors] + [f"분류 안 된 오류: {e}" for e in unclassified]
receipt = {"HEAD": head, "base": args.base, "landing": context.get("landing"), "required": context.get("required_count"),
           "errors": len(errors), "attributed_commits": attributed, "own_functions": rows,
           "own_bundle_errors": own_errors, "sibling_missing": len(other) - len(unclassified), "unclassified": unclassified,
           "failures": failures, "verdict": "PASS" if not failures else "FAIL"}
print(json.dumps(receipt, ensure_ascii=False, indent=1))
raise SystemExit(0 if not failures else 1)
