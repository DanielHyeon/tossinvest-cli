#!/usr/bin/env python3
"""task 7.5.2.1 VERIFY — `main()` **출력 전체**의 판정 A/B. 순서를 번갈아 · 결정적 계수 · 이어 달리기.

7.5.2 의 A/B(`751_check_ab.py`)는 `check()` 의 반환만 비교했다. 7.5.2.1 은 창 줄의 커밋 수와 조언 줄
(`_commits_after` · `_recording_refusal`)도 고정한 `HEAD` 로 바꾸므로 `main()` 이 **찍는 줄 전부**와 rc 를
비교한다. 그리고 7.5.2 가 적은 시간 이득(2048.9s → 1934.5s)은 performance 전문가가 대부분 **실행 순서
표류**라고 짚었다 — 늘 before 를 먼저 돌렸기 때문이다. 여기서는 change 마다 순서를 번갈아 돌리고, 시간
대신 **결정적 계수**(git 프로세스 수 · `ast.json` 읽기 수)를 적는다. 시간은 참고로만 남긴다.

    python3 7521_main_ab.py [<before-sha>] [<초 예산>]
"""
import contextlib
import hashlib
import importlib.util
import io
import json
import shutil
import subprocess
import sys
import time
from pathlib import Path
from unittest import mock

# 루트는 **유도**한다 ([[renamed-checkout-strands-absolute-path-state]]).
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
BEFORE = next((a for a in sys.argv[1:] if len(a) >= 7 and not a.replace(".", "").isdigit()),
              "fc35eb2d013dccd0700836f857ad9d3e65bff0dc")
budget = next((float(a) for a in sys.argv[1:] if a.replace(".", "").isdigit()), 480.0)


def build(tag: str, revision: str | None):
    where = SP / f"7521_{tag}"
    if where.exists():
        shutil.rmtree(where)
    shutil.copytree(ROOT / "tools" / "logic-map", where,
                    ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:
        blob = subprocess.run(["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
                              cwd=ROOT, capture_output=True, check=True).stdout
        (where / "check_analysis.py").write_bytes(blob)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca7521_{tag}", where / "check_analysis.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[f"ca7521_{tag}"] = module
    spec.loader.exec_module(module)
    sys.path.pop(0)
    return module, (where / "check_analysis.py").read_bytes()


# 기준은 **풀어서** 쓴다 — 짧은 sha 를 그대로 두면 이어 달리기 기록의 열쇠가 철자에 묶인다.
BEFORE = subprocess.run(["git", "rev-parse", "--verify", f"{BEFORE}^{{commit}}"],
                        cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
before, before_src = build("before", BEFORE)
after, after_src = build("after", None)
# 대조군이 대조군인지는 **소스**로 확인한다 — 같으면 A/B 가 공허하다.
assert before_src != after_src, f"{BEFORE[:12]} 사본이 워킹트리와 같다 — 대조군 오염"
print(f"대조군 OK · 비교 기준 blob = {BEFORE[:12]}")

changes = ROOT / "openspec" / "changes"
ids = sorted(p.name for p in changes.iterdir() if p.is_dir() and p.name != "archive")
ids += sorted(after.ARCHIVED_CHANGE.fullmatch(p.name).group("change")
              for p in (changes / "archive").iterdir()
              if p.is_dir() and after.ARCHIVED_CHANGE.fullmatch(p.name))


def run_main(module, change: str) -> dict:
    """`main()` 한 번 — 찍은 줄 · rc · git 프로세스 수 · `ast.json` 읽기 수 · 시간."""
    counts = {"git": 0, "ast_reads": 0}
    real_run, real_bytes, real_text = subprocess.run, Path.read_bytes, Path.read_text

    def counting_run(*args, **kwargs):
        argv = args[0] if args else kwargs.get("args")
        if isinstance(argv, list) and argv and argv[0] == "git":
            counts["git"] += 1
        return real_run(*args, **kwargs)

    def counting_bytes(path):
        if path.name == "ast.json":
            counts["ast_reads"] += 1
        return real_bytes(path)

    def counting_text(path, *args, **kwargs):
        if path.name == "ast.json":
            counts["ast_reads"] += 1
        return real_text(path, *args, **kwargs)

    out = io.StringIO()
    argv = ["check_analysis.py", "--change", change, "--root", str(ROOT)]
    start = time.monotonic()
    with mock.patch.object(sys, "argv", argv), contextlib.redirect_stdout(out), \
            mock.patch.object(subprocess, "run", counting_run), \
            mock.patch.multiple(Path, read_bytes=counting_bytes, read_text=counting_text):
        try:
            code = module.main()
        except BaseException as exc:              # 예외도 판정이다 — 타입과 문장까지 비교한다
            code = f"RAISED {type(exc).__name__}: {exc}"
    return {"rc": code, "lines": out.getvalue().splitlines(), "s": time.monotonic() - start, **counts}


# 이어 달리기 기록은 **양쪽 소스**에 묶는다 (7.5.2 의 교훈 — 편집 도중 다시 돌리면 판본이 섞인다).
DONE = SP / f"7521_main_ab_done.{BEFORE[:12]}.{hashlib.sha256(after_src).hexdigest()[:12]}.json"
done = json.load(open(DONE)) if DONE.exists() else {}
spent = 0.0
for position, cid in enumerate(ids):
    if cid in done or spent > budget:
        continue
    # 순서를 **번갈아** 돈다 — 늘 before 를 먼저 돌리면 캐시 · 페이지 캐시가 after 를 돕는다.
    order = (("before", before), ("after", after)) if position % 2 == 0 else (("after", after), ("before", before))
    result = {tag: run_main(module, cid) for tag, module in order}
    spent += result["before"]["s"] + result["after"]["s"]
    same = (result["before"]["rc"], result["before"]["lines"]) == (result["after"]["rc"], result["after"]["lines"])
    done[cid] = {"same": same, "first": order[0][0], **{
        f"{tag}_{key}": result[tag][key] for tag in ("before", "after")
        for key in ("rc", "s", "git", "ast_reads")}}
    if not same:
        done[cid]["before_lines"] = result["before"]["lines"][:6]
        done[cid]["after_lines"] = result["after"]["lines"][:6]
        print(f"  DIFFERENT {cid}\n    before {result['before']['lines'][:3]}\n    after  {result['after']['lines'][:3]}")
    json.dump(done, open(DONE, "w"), indent=1, ensure_ascii=False)

rows = done.values()
same = sum(row["same"] for row in rows)
print(f"\n전수 {len(ids)} · 남은 것 {len(ids) - len(done)} · 비교 {len(done)} · "
      f"**SAME {same} · DIFFERENT {len(done) - same}**")
for key in ("git", "ast_reads"):
    print(f"{key}: 합계 {sum(r[f'before_{key}'] for r in rows)} → {sum(r[f'after_{key}'] for r in rows)}")
# 시간은 **참고**다 — 순서를 번갈아도 한 기계의 부하 표류는 남는다.
first = {tag: [r for r in rows if r["first"] == tag] for tag in ("before", "after")}
for tag, group in first.items():
    if group:
        print(f"먼저 돈 쪽 = {tag} ({len(group)}건): before {sum(r['before_s'] for r in group):.1f}s · "
              f"after {sum(r['after_s'] for r in group):.1f}s")
