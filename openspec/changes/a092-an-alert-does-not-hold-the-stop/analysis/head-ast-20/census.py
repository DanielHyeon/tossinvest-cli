#!/usr/bin/env python3
"""a092 20판 — 커밋된 FLM 번들(base 285c7619)을 현재 HEAD 소스와 대조하는 계수 하네스.

무엇을 하는가
  1. `tools/logic-map` 을 임시 디렉터리에 빌드함
  2. `analysis/function-logic/*/ast.json` 마다 같은 파일·함수를 현재 소스에서 다시 추출함
  3. 추가 대상(EXTRA — a098 이 착지시킨 함수, 번들 없음)도 추출함
  4. 줄·sha 를 뺀 구조가 같은지(SAME)·다른지(DIFF)를 셈

출력 디렉터리를 주면 추출 결과를 `<번들>.json` 으로 남김 — 20판이 인용하는 HEAD 산출물이 그것임.
판정 대상 바이트를 한 번만 읽음(추출 결과는 이 실행이 만든 바이트).

사용: python3 census.py [출력 디렉터리]   (저장소 루트 기준 경로를 씀 — 루트는 이 파일에서 유도함)
"""
import json
import subprocess
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
CHANGE = HERE.parent.parent
ROOT = CHANGE.parents[2]
BUNDLES = CHANGE / "analysis" / "function-logic"

# 번들이 없는 a098 착지 함수 — 20판이 분기를 근거로 쓰는 자리만 적음
EXTRA = [
    ("internal/app/engine/alertdelivery.go", "alertDeliverer.deliverOne"),
    ("internal/app/engine/alertdelivery.go", "alertDeliverer.cycle"),
    ("internal/app/engine/auxiliary.go", "blockEntryOnDeliveryStop"),
    ("internal/app/engine/auxiliary.go", "Runtime.runAuxiliary"),
]


def strip(o):
    """줄·열·sha 를 지운 구조 — 줄 이동만 있는 함수를 SAME 으로 가르기 위함."""
    if isinstance(o, dict):
        return {k: strip(v) for k, v in o.items()
                if k not in ("at", "line", "column", "start", "end", "source_sha256")}
    if isinstance(o, list):
        return [strip(x) for x in o]
    return o


def counts(j):
    return tuple(len(j.get(k) or []) for k in
                 ("branches", "returns", "calls", "assignments", "go_statements", "defers"))


def main() -> int:
    out = Path(sys.argv[1]) if len(sys.argv) > 1 else None
    if out:
        out.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory() as tmp:
        tool = Path(tmp) / "logicmap"
        subprocess.run(["go", "build", "-o", str(tool), "./tools/logic-map"], cwd=ROOT, check=True)

        def extract(file, func):
            r = subprocess.run([str(tool), "--file", file, "--func", func],
                               cwd=ROOT, check=True, capture_output=True)
            return r.stdout

        rows = []
        for d in sorted(p for p in BUNDLES.iterdir() if p.is_dir()):
            old = json.loads((d / "ast.json").read_bytes())
            func = (old["receiver"] + "." if old.get("receiver") else "") + old["function"]
            raw = extract(old["file"], func)
            new = json.loads(raw)
            if out:
                (out / f"{d.name}.json").write_bytes(raw)
            rows.append((d.name, old["source_sha256"] == new["source_sha256"],
                         strip(old) == strip(new), old["start"]["line"], new["start"]["line"],
                         counts(old), counts(new)))
        for file, func in EXTRA:
            raw = extract(file, func)
            if out:
                name = "extra--" + func.lower()
                (out / f"{name}.json").write_bytes(raw)
    sha_same = sum(1 for r in rows if r[1])
    logic_diff = [r for r in rows if not r[2]]
    shift_only = [r for r in rows if r[2] and not r[1]]
    print(f"bundles={len(rows)} sha_match={sha_same} line_shift_only={len(shift_only)} logic_diff={len(logic_diff)}")
    for r in rows:
        tag = "MATCH" if r[1] else ("SHIFT" if r[2] else "DIFF")
        print(f"{tag:5s} {r[0]:55s} start {r[3]}->{r[4]} counts {r[5]}->{r[6]}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
