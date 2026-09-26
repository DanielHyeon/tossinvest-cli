#!/usr/bin/env python3
"""task 6.4 보수(P0-A) — 새 가드가 거절할 **정상 입력**을 편집 전에 센다 ([[fail-closed-must-name-what-it-rejects]]).

새 가드: `base-commit.txt` 의 **지금 디스크 경로**가 HEAD 에 없으면, 같은 id 의 다른 자리(활성 `changes/<id>/` ·
HEAD 트리의 `changes/archive/<날짜>-<id>/`)를 HEAD 에서 찾아 값이 다르면 거절한다. 그러므로 거절은
"디스크 경로가 HEAD 에 없다" **그리고** "다른 자리의 커밋된 값이 디스크와 다르다" 둘 다일 때만 난다.

이 스크립트는 도구 코드를 부르지 않고 git 으로 직접 잰다(해독만 `_archived_change_id` 를 빌린다 — 셋째 정규식을
만들지 않는다). 모집단: 저장소 `openspec/changes/` 의 change 디렉터리 전부(활성 + 아카이브).

    python3 64r_census.py
"""
import subprocess
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))

import check_analysis  # noqa: E402


def git(*args: str) -> subprocess.CompletedProcess:
    return subprocess.run(["git", *args], cwd=REPO, capture_output=True, check=False)


head = git("rev-parse", "--verify", "HEAD^{commit}").stdout.decode().strip()
changes = REPO / "openspec" / "changes"
dirs = sorted([p for p in changes.iterdir() if p.is_dir() and p.name != "archive"]
              + [p for p in (changes / "archive").iterdir() if p.is_dir()])
listed = git("ls-tree", "-z", "-d", "--name-only", head, "--", "openspec/changes/archive/")
assert listed.returncode == 0, listed.stderr
archived_in_head = [entry.decode() for entry in listed.stdout.split(b"\0") if entry]


def blob(relative: str) -> bytes | None:
    shown = git("cat-file", "blob", f"{head}:{relative}")
    return shown.stdout if shown.returncode == 0 else None


rows = {"own path in HEAD, same bytes": 0, "own path in HEAD, different bytes": 0,
        "own path not in HEAD": 0, "no base-commit.txt on disk": 0}
elsewhere_multi = []
rejected = []
shape_off = []     # 40자리 소문자 + LF 한 줄이 아닌 것 — 6.4(b) 의 모양 가드가 거절
not_self = []      # `^{commit}` 이 벗긴 값이 적힌 값과 다른 것 — 6.4(b) 의 태그 가드가 거절
for directory in dirs:
    relative_dir = directory.relative_to(REPO).as_posix()
    change = check_analysis._archived_change_id(directory.name) if directory.parent.name == "archive" else directory.name
    base = directory / "base-commit.txt"
    if not base.exists():
        rows["no base-commit.txt on disk"] += 1
        continue
    disk = base.read_bytes()
    value = disk.decode("utf-8", "replace").strip()
    if disk != (value + "\n").encode() or not check_analysis.FULL_SHA.fullmatch(value):
        shape_off.append(relative_dir)
    elif git("rev-parse", "--verify", f"{value}^{{commit}}").stdout.decode().strip() != value:
        not_self.append(relative_dir)
    own = blob(f"{relative_dir}/base-commit.txt")
    # 같은 id 의 HEAD 자리 전부(자기 자리 포함) — 둘 이상이면 "어느 자리의 값과 견주나" 가 모호해진다.
    places = [f"openspec/changes/{change}/base-commit.txt"] + [
        f"{entry}/base-commit.txt" for entry in archived_in_head
        if check_analysis._archived_change_id(entry.rsplit("/", 1)[-1]) == change]
    present = {place: blob(place) for place in places}
    present = {place: value for place, value in present.items() if value is not None}
    if len(present) > 1:
        elsewhere_multi.append((relative_dir, sorted(present)))
    if own is None:
        rows["own path not in HEAD"] += 1
        others = {place: value for place, value in present.items() if place != f"{relative_dir}/base-commit.txt"}
        if any(value.strip() != disk.strip() for value in others.values()):
            rejected.append(relative_dir)
    elif own == disk:
        rows["own path in HEAD, same bytes"] += 1
    else:
        rows["own path in HEAD, different bytes"] += 1

print(f"HEAD {head} · change 디렉터리 {len(dirs)} · HEAD 트리의 아카이브 항목 {len(archived_in_head)}")
for key, value in rows.items():
    print(f"  {key}: {value}")
print(f"  같은 id 가 HEAD 에 두 자리 이상: {len(elsewhere_multi)} {elsewhere_multi}")
print(f"  모양(40자리 + LF) 밖: {len(shape_off)} {shape_off} · 자기 id 가 커밋이 아님: {len(not_self)} {not_self}")
print(f"새 가드가 거절: {len(rejected)} {rejected}")
