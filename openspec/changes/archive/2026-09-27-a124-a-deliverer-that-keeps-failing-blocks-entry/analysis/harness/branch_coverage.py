#!/usr/bin/env python3
"""a124 tasks 1.4 — 번들 분기마다 어떤 시험이 그 본문을 실행하는지 잰다.

손으로 「이 시험이 이 분기를 탄다」고 적으면 볼 곳을 고른 증거가 된다. 이 하네스는
`go test -c -cover -covermode=set` 바이너리를 시험 함수마다 따로 돌리고, 커버 프로필의
블록을 ast.json 의 분기 좌표에 붙인다.

분기 → 블록 규칙 (Go 커버 블록 모양에서 유도):
- if · case · for · range · else: 분기 좌표와 **같은 줄**에서 시작하고 열이 좌표보다 큰 첫 블록
  (`{` 또는 `:` 뒤의 본문). 본문이 비어 블록이 없으면 "블록 없음".
- switch: 분기 좌표를 **포함하는** 블록(스위치 문 자체를 실행했는가).

쓰기: python3 branch_coverage.py --workdir <저장소 사본> --pkg ./internal/app/engine \
        --bundle <번들 디렉터리> [--bundle …] --tests '<정규식>' [--union] --out out.json
- --tests 에 맞는 시험을 하나씩 돌린다. --union 이면 패키지 전체를 한 번 더 돌려 합집합을 잰다.
- 출력은 분기마다 {entered_by: [시험…], union: bool|None, block: "l.c-l.c"|None}.

실행은 사본(연결 워크트리)에서 한다 — 병행 로트의 미커밋 편집이 측정에 섞이지 않게.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

BLOCK = re.compile(r"^(.+?):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$")


def module_path(workdir: Path) -> str:
    for line in (workdir / "go.mod").read_text(encoding="utf-8").splitlines():
        if line.startswith("module "):
            return line.split()[1]
    raise SystemExit("go.mod has no module line")


def parse_profile(path: Path) -> dict[tuple[str, int, int, int, int], bool]:
    out: dict[tuple[str, int, int, int, int], bool] = {}
    if not path.exists():
        return out
    for line in path.read_text(encoding="utf-8").splitlines()[1:]:
        m = BLOCK.match(line)
        if not m:
            continue
        key = (m.group(1), int(m.group(2)), int(m.group(3)), int(m.group(4)), int(m.group(5)))
        out[key] = out.get(key, False) or int(m.group(7)) > 0
    return out


def branch_block(blocks, file_key: str, branch: dict):
    line, col = branch["at"]["line"], branch["at"]["column"]
    mine = [b for b in blocks if b[0] == file_key]
    if branch["kind"] == "switch":
        cands = [b for b in mine if (b[1], b[2]) <= (line, col) <= (b[3], b[4])]
        cands.sort(key=lambda b: ((b[3] - b[1]), b[4] - b[2]))
        return cands[0] if cands else None
    cands = sorted(b for b in mine if b[1] == line and b[2] > col)
    return cands[0] if cands else None


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--workdir", required=True)
    ap.add_argument("--pkg", required=True)
    ap.add_argument("--bundle", action="append", required=True)
    ap.add_argument("--tests", required=True)
    ap.add_argument("--union", action="store_true")
    ap.add_argument("--tags", default="")
    ap.add_argument("--timeout", default="600s")
    ap.add_argument("--out", required=True)
    args = ap.parse_args()

    workdir = Path(args.workdir).resolve()
    mod = module_path(workdir)
    pkg_import = mod + "/" + args.pkg.removeprefix("./").rstrip("/")
    env = dict(os.environ, GOFLAGS="-trimpath")
    tmp = Path(tempfile.mkdtemp(prefix="a124cov-"))
    binary = tmp / "pkg.test"
    build = ["go", "test", "-c", "-cover", "-covermode=set", "-coverpkg", pkg_import, "-o", str(binary)]
    if args.tags:
        build += ["-tags", args.tags]
    build.append(args.pkg)
    subprocess.run(build, cwd=workdir, env=env, check=True)
    pkg_dir = workdir / args.pkg

    listed = subprocess.run([str(binary), "-test.list", ".*"], cwd=pkg_dir, env=env,
                            capture_output=True, text=True, check=True).stdout.split()
    tests = [t for t in listed if re.search(args.tests, t) and t.startswith("Test")]

    bundles = []
    for b in args.bundle:
        ast = json.loads((Path(b) / "ast.json").read_text(encoding="utf-8"))
        bundles.append((Path(b).name, ast))

    per_test: dict[str, tuple[bool, dict]] = {}
    failures: dict[str, str] = {}
    for name in tests:
        prof = tmp / f"{name}.out"
        proc = subprocess.run([str(binary), "-test.run", f"^{name}$", "-test.count", "1",
                               "-test.timeout", args.timeout, "-test.coverprofile", str(prof)],
                              cwd=pkg_dir, env=env, capture_output=True, text=True)
        per_test[name] = (proc.returncode == 0, parse_profile(prof))
        if proc.returncode != 0:
            failures[name] = (proc.stdout + proc.stderr)[-2000:]

    union = None
    if args.union:
        prof = tmp / "union.out"
        proc = subprocess.run([str(binary), "-test.count", "1", "-test.timeout", "45m",
                               "-test.coverprofile", str(prof)], cwd=pkg_dir, env=env,
                              capture_output=True, text=True)
        union = (proc.returncode == 0, parse_profile(prof))
        if proc.returncode != 0:
            failures["<union>"] = "\n".join(
                l for l in (proc.stdout + proc.stderr).splitlines() if l.startswith(("--- FAIL", "FAIL", "panic")))[-4000:]

    any_blocks = next((p for _, p in per_test.values() if p), None) or (union[1] if union else {})
    result = {"package": pkg_import, "tests_run": len(tests),
              "failed_tests": sorted(n for n, (ok, _) in per_test.items() if not ok),
              "union_passed": None if union is None else union[0], "failures": failures, "bundles": {}}
    for bname, ast in bundles:
        file_key = mod + "/" + ast["file"]
        rows = {}
        for br in ast.get("branches") or []:
            blk = branch_block(list(any_blocks.keys()), file_key, br)
            entered = sorted(n for n, (_, p) in per_test.items() if blk and p.get(blk))
            rows[br["id"]] = {
                "kind": br["kind"], "line": br["at"]["line"], "column": br["at"]["column"],
                "block": None if blk is None else f"{blk[1]}.{blk[2]}-{blk[3]}.{blk[4]}",
                "entered_by": entered,
                "union": None if union is None or blk is None else bool(union[1].get(blk)),
            }
        result["bundles"][bname] = rows
    Path(args.out).write_text(json.dumps(result, ensure_ascii=False, indent=1), encoding="utf-8")
    print(f"tests={len(tests)} failed={len(result['failed_tests'])} out={args.out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
