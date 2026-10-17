#!/usr/bin/env python3
"""task 7.5.13 · 7.5.26 영수증 — git 이 판정 diff 의 머리 줄에서 **무엇을** 인용하는가, 그리고 저장소가 그런 이름을 갖는가.

1. **인용 표.** 스크래치 저장소에 ASCII 1~127(`/` 제외)과 비ASCII 둘(`é` · U+2028)을 하나씩 든 `*.go` 를 만들어 편집하고,
   `_changed_existing_functions` 와 **같은 깃발**(`-c core.quotePath=false diff --src-prefix=a/ … --unified=0`)로 두 커밋을 견준다.
   각 이름의 `--- ` 머리 값을 적고, 워킹트리의 `check_analysis._git_header_path`(있으면)가 그 글자를 똑같이 만드는지 본다.
2. **센서스.** 이 저장소의 역사 전부(`git log --all --name-only -z`)와 추적 목록에서 `*.go` 이름 중 git 이 인용하는 바이트
   (`< 0x20` · `"` · `\\` · `0x7f`)가 든 것을 센다 — 거부 · 수용이 바뀌는 정상 입력의 모집단이다.

    python3 7513_quoted.py
"""
import os
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))
import check_analysis  # noqa: E402

ENV = {**os.environ, "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_NOSYSTEM": "1"}
QUOTED = set(range(1, 0x20)) | {0x22, 0x5C, 0x7F}


def git(root: Path, *args: str, **kwargs) -> subprocess.CompletedProcess:
    return subprocess.run(["git", "-C", str(root), *args], env=ENV, capture_output=True, check=True, **kwargs)


def table() -> None:
    encoder = getattr(check_analysis, "_git_header_path", None)
    names = {f"c{byte:03d}": "q" + chr(byte) + "q.go" for byte in range(1, 128) if chr(byte) != "/"}
    names.update({"e-acute": "qéq.go", "u2028": "q q.go"})
    with tempfile.TemporaryDirectory() as raw:
        root = Path(raw)
        git(root, "init", "-q")
        assert git(root, "rev-parse", "--show-toplevel").stdout.decode().strip() == str(root.resolve())
        for name in names.values():
            (root / name).write_bytes(b"package q\nfunc F() int { return 1 }\n")
        git(root, "add", "-A")
        git(root, "-c", "user.email=a@b", "-c", "user.name=a", "commit", "-qm", "base")
        for name in names.values():
            (root / name).write_bytes(b"package q\nfunc F() int { return 2 }\n")
        git(root, "add", "-A")
        git(root, "-c", "user.email=a@b", "-c", "user.name=a", "commit", "-qm", "edit")
        out = git(root, "-c", "core.quotePath=false", "diff", "--no-ext-diff", "--no-textconv", "--find-renames",
                  "--src-prefix=a/", "--dst-prefix=b/", "--no-color", "--unified=0", "--inter-hunk-context=0",
                  "HEAD~1", "HEAD").stdout.decode("utf-8").split("\n")
        headers = [check_analysis._header_name(line) for line in out if line.startswith("--- ")]
        by_name = {}
        for label, name in names.items():
            plain = "a/" + name
            by_name[label] = plain in headers
        quoted = sorted(label for label, plain in by_name.items() if not plain)
        expected = sorted(f"c{b:03d}" for b in QUOTED if chr(b) != "/")
        print(f"git {git(root, 'version').stdout.decode().split()[-1]} · 이름 {len(names)} · 머리 {len(headers)}")
        print(f"인용된 이름 {len(quoted)} · 예측(`< 0x20` · `\"` · `\\` · 0x7f) {len(expected)} · 일치 {quoted == expected}")
        print(f"비ASCII 인용: é {not by_name['e-acute']} · U+2028 {not by_name['u2028']}")
        if encoder is not None:
            rendered = {encoder("a/" + name) for name in names.values()}
            print(f"`_git_header_path` 가 git 의 머리 값 {len(headers)} 개를 전부 만든다: {rendered == set(headers)}"
                  + ("" if rendered == set(headers) else f" — 다름 {sorted(set(headers) ^ rendered)[:6]}"))
        else:
            print("`_git_header_path` 없음 — 이 리비전은 인용을 만들지 않는다")


def census() -> None:
    head = subprocess.run(["git", "rev-parse", "--short=12", "HEAD"], cwd=REPO, capture_output=True, text=True).stdout.strip()
    history = subprocess.run(["git", "log", "--all", "--format=", "--name-only", "-z"], cwd=REPO,
                             capture_output=True, check=True).stdout
    tracked = subprocess.run(["git", "ls-files", "-z"], cwd=REPO, capture_output=True, check=True).stdout
    names = {raw for raw in (history + b"\0" + tracked).split(b"\0") if raw.endswith(b".go")}
    hit = sorted(raw for raw in names if any(byte in QUOTED for byte in raw))
    print(f"센서스 HEAD {head} · 역사 전부 + 추적의 고유 `*.go` 이름 {len(names)} · git 이 인용하는 바이트가 든 것 {len(hit)}")
    for raw in hit[:10]:
        print("   ", raw)


if __name__ == "__main__":
    table()
    census()
