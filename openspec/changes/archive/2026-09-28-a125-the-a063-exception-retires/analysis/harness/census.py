#!/usr/bin/env python3
"""a125 1.1 · 3.1 — 창 영향 전수. 활성 + 아카이브 전 change 를 **한 도구 판본**으로 판정해 JSON 으로 남긴다.

    python3 census.py --tool <check_analysis.py 가 있는 디렉터리> --root <고정한 연결 워크트리> --expect-head <sha> --out <json>

- 편집 전(1.1)은 `--tool <root>/tools/logic-map`, 편집 후(3.1)는 편집한 판본의 디렉터리를 준다. 두 판은 **같은 root · 같은 HEAD**
  여야 비교가 된다 — 시작 sha 를 단언하고 끝에서 다시 단언한다(병행 커밋이 측정 중 HEAD 를 옮기지 못하게).
- 판정 입력은 `check(id, root, context)` 하나다(게이트가 부르는 경로). 오류 줄은 전문을 남긴다 — 비교는 줄 단위다.
- `GOFLAGS=-trimpath` · 전용 `GOCACHE` 는 호출자가 준다(go 캐시가 경로마다 자란다).
"""
import argparse
import json
import os
import subprocess
import sys
import time
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("--tool", required=True)
parser.add_argument("--root", required=True)
parser.add_argument("--expect-head", required=True)
parser.add_argument("--out", required=True)
parser.add_argument("--only", default="")
args = parser.parse_args()
root = Path(args.root).resolve()
sys.path.insert(0, str(Path(args.tool).resolve()))
import check_analysis  # noqa: E402

assert Path(check_analysis.__file__).resolve().parent == Path(args.tool).resolve(), check_analysis.__file__


def head() -> str:
    return subprocess.run(["git", "-C", str(root), "rev-parse", "HEAD"], capture_output=True, text=True, check=True).stdout.strip()


assert head() == args.expect_head, (head(), args.expect_head)
changes = root / "openspec" / "changes"
ids = sorted(p.name for p in changes.iterdir() if p.is_dir() and p.name != "archive")
ids += sorted(check_analysis._archived_change_id(p.name) or p.name for p in (changes / "archive").iterdir() if p.is_dir())
if args.only:
    ids = [i for i in ids if i in args.only.split(",")]
rows = {}
for change in ids:
    started = time.time()
    context: dict[str, object] = {}
    try:
        errors = check_analysis.check(change, root, context)
    except Exception as exc:  # 게이트의 경계(main)는 GATE_FAULTS 만 줄로 바꾼다 — 그 밖은 여기서 이름으로 남긴다
        errors = [f"RAISED {type(exc).__name__}: {exc}"]
    rows[change] = {
        "errors": errors,
        "required": context.get("required_count"),
        "landing": context.get("landing"),
        "base": context.get("effective_base"),
        "keys": sorted(k for k in context if k not in ("head",)),
        "seconds": round(time.time() - started, 1),
    }
    print(f"{change} errors={len(errors)} required={rows[change]['required']} {rows[change]['seconds']}s", flush=True)
assert head() == args.expect_head, ("HEAD moved during the census", head())
Path(args.out).write_text(json.dumps({"head": args.expect_head, "tool": str(Path(args.tool).resolve()), "pid": os.getpid(), "rows": rows}, ensure_ascii=False, indent=1))
print("done", len(rows))
