#!/usr/bin/env python3
"""a166 time-namespace 로트 — Namespace/BootStamp 자리 census.

입력: logic-map AST(json) 디렉터리와 저장소 루트. 각 AST 함수 범위 안에서
(1) AST 가 열거한 분기(if) 줄 중 조건 줄에 `Namespace` 가 있는 것,
(2) AST 가 열거한 호출 중 RemainingBootHint / validObservationStamp / validStamp / Snapshot / sample / namespace 호출,
(3) 함수 범위 안 `Namespace` 토큰이 나오는 모든 줄
을 센다. 분기 줄의 원문은 저장소 파일에서 읽는다(AST 가 조건 문자열을 내지 않으므로).
사용: census.py <repo> <ast-dir>
"""
import json
import pathlib
import sys

repo = pathlib.Path(sys.argv[1])
ast_dir = pathlib.Path(sys.argv[2])
WATCH = {"RemainingBootHint", "clock.RemainingBootHint", "validObservationStamp", "validStamp",
         "s.boot.Snapshot", "s.sample", "snapshot", "s.probe.namespace", "clock.SystemBootSource",
         "accountbudget.OpenObserved", "OpenObservation", "p.budget.PollRemaining", "obs.PollRemaining",
         "obs.Finish", "obs.Begin", "pruneObservationSlots", "validateObservationState", "s.read"}
for path in sorted(ast_dir.glob("*.json")):
    d = json.loads(path.read_text())
    src = (repo / d["file"]).read_text().splitlines()
    start, end = d["start"]["line"], d["end"]["line"]
    name = (d.get("receiver", "") + "." if d.get("receiver") else "") + d["function"]
    print(f"## {d['file']} {name} [{start}-{end}] branches={len(d["branches"] or [])} returns={len(d["returns"] or [])}")
    for b in d["branches"] or []:
        line = src[b["at"]["line"] - 1]
        if "Namespace" in line or "current != children" in line or "currentAfter" in line:
            print(f"  BRANCH {b['id']} L{b['at']['line']}: {line.strip()[:160]}")
    for c in d["calls"] or []:
        if c.get("text", "") in WATCH:
            print(f"  CALL L{c['at']['line']}: {c['text']}")
    for i in range(start, end + 1):
        if "Namespace" in src[i - 1]:
            print(f"  TOKEN L{i}: {src[i - 1].strip()[:160]}")
