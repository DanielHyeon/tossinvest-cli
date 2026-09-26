#!/usr/bin/env python3
"""task 7.5.11 영수증 — `Path.is_dir()` 과 `Path.rglob()` 이 **이 인터프리터에서** 무엇을 삼키는가.

`_listing_outcome` 은 `child.is_dir()` 로 종류를 정하고 `_pattern_outcome` 은 `rglob` 으로 순회했다. 둘 다 `OSError`
처리가 **판본마다 다르다** ([[python-behaviour-differs-by-version]]) — 그래서 주장하기 전에 잰다. 스크래치 디렉터리에
다섯 모양을 만들고 각 모양에서 두 원시가 **돌려준 값**(또는 올린 예외)을 적는다. 저장소는 건드리지 않는다.

    python3 7511_isdir.py                 # 시스템 인터프리터
    uv run --python 3.14 7511_isdir.py    # 다른 판본과 대조할 때

root 는 권한을 무시하므로 권한 모양은 root 에서 뜻이 없다 — 그 줄에 그렇게 적는다.
"""
import os
import sys
import tempfile
from pathlib import Path


def outcome(action) -> str:
    try:
        return repr(action())
    except OSError as exc:
        return f"raises {type(exc).__name__}(errno {exc.errno})"


def main() -> None:
    print(f"python {sys.version.split()[0]} · euid {os.geteuid() if hasattr(os, 'geteuid') else '?'}")
    with tempfile.TemporaryDirectory() as raw:
        root = Path(raw)
        (root / "dangling").symlink_to(root / "nowhere")
        (root / "loop").symlink_to(root / "loop")
        locked = root / "locked"                 # r 만 있고 x 가 없다 — 목록은 되고 자식 stat 은 안 된다
        locked.mkdir()
        (locked / "inner").mkdir()
        (locked / "inner" / "x_test.go").write_text("package x\n")
        hidden = root / "hidden"                 # 목록(r)이 안 되는 하위 트리
        hidden.mkdir()
        (hidden / "y_test.go").write_text("package y\n")
        (root / "plain_test.go").write_text("package p\n")
        locked.chmod(0o444)
        hidden.chmod(0o000)
        try:
            print("is_dir · 끊긴 심링크          ", outcome(lambda: (root / "dangling").is_dir()))
            print("is_dir · 심링크 고리          ", outcome(lambda: (root / "loop").is_dir()))
            print("is_dir · x 없는 부모의 자식    ", outcome(lambda: (locked / "inner").is_dir()))
            print("is_dir · 사라진 이름(ENOENT)  ", outcome(lambda: (root / "vanished").is_dir()))
            print("os.stat · x 없는 부모의 자식   ", outcome(lambda: os.stat(locked / "inner").st_mode))
            print("rglob('*_test.go')           ", outcome(lambda: sorted(p.relative_to(root).as_posix()
                                                                           for p in root.rglob("*_test.go"))))
        finally:
            locked.chmod(0o755)
            hidden.chmod(0o755)


if __name__ == "__main__":
    main()
