#!/usr/bin/env python3
"""Bind Function Logic Map evidence to every modified existing Go function."""

from __future__ import annotations

import argparse
import contextlib
import errno
import hashlib
import io
import json
import os
import re
import stat
import subprocess
import sys
import tempfile
import zlib
from pathlib import Path
from typing import Iterator, NamedTuple

from role_check import call_enumeration_in_use, role_errors
from execution_baseline import AdoptionError, validate as validate_execution_baseline

ROOT = Path(__file__).resolve().parents[2]
# 게이트가 띄우는 **모든** git 은 교체 참조(`refs/replace/*`)를 따르지 않는다 (task 7.5.31). 참조 하나로 어떤 blob 이든
# 다른 blob 으로 읽히게 할 수 있다 — 편집 blob 을 base blob 으로 돌리면 두 diff 가 **함께** 속아 요구가 빈다(7.5.25
# 적대 재리뷰 F1 · F4 실측, `git status` 는 `M` 인데 `[]`). 자식이 열여섯 자리 넘게 있으므로 자리마다 적지 않고 이
# 프로세스의 환경에 한 번 둔다 — 새 호출 자리가 저절로 따른다. 판정이 교체 참조를 쓸 까닭은 없다.
os.environ["GIT_NO_REPLACE_OBJECTS"] = "1"
REQUIRED = (
    "ast.json",
    "function-logic-map.md",
    "branch-test-map.md",
    "risk-pattern-report.md",
)
EXEMPTION = "Function Logic Map: not-applicable"
# 읽기의 상한과 실패 문구 (task 7.5.2.3). 상한은 **거부하는 정상 입력을 먼저 세어** 골랐다. 열거표는
# `analysis/harness/7524_census.py` 가 다시 찍는다 — 값은 저장소와 함께 움직이므로 날짜를 적는다.
# 결정에 쓰이는 사실은 **크기**다: 가장 큰 `*.go` 97,231 B · 가장 큰 번들 파일 39,327 B(아카이브
# a047) · 활성 번들 중 최대 26,694 B · **16 MiB 를 넘는 번들 파일 0** (2026-09-23 실측).
# **번들의 총 개수는 여기 안 적는다 (7.5.23).** 그 수는 번들을 더하는 **모든 커밋마다** 바뀌므로,
# 어느 값을 적어도 적는 그 커밋에서 이미 낡는다 — 두 로트 연속으로 그렇게 틀렸다(12,411 → 12,412 →
# 실제 12,414, 매번 그 로트 자신이 더한 번들이 빠진 값이었다). 세려면 `7524_census.py` 를 돌린다.
# (7.5.2.4 정정: 옛 문장은 한 문장에 모집단이 둘이었다 — 25,466 B 는 활성 번들만의 최대였는데 전수를
#  세는 0/12,193 과 나란히 적혀 전수의 최대처럼 읽혔다.)
READ_CAP = 16 << 20
NOT_REGULAR = "not a regular file"
TOO_LARGE = f"larger than the {READ_CAP} byte limit"
NOT_UTF8 = "not UTF-8 text"
UNREADABLE = "{what} could not be read: {why}"
HUNK = re.compile(r"^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@")
MAP_SECTIONS = (
    "## Inputs and invariants",
    "## Branches and early returns",
    "## Calls and live bindings",
    "## State mutations and fallbacks",
    "## Safety conclusion",
)


def qualified(value: dict) -> str:
    receiver = str(value.get("receiver", "")).strip()
    function = str(value.get("function", "")).strip()
    return f"{receiver}.{function}" if receiver else function


def go_functions(path: Path, root: Path) -> list[dict]:
    process = subprocess.run(
        [
            "go",
            "run",
            "./tools/logic-map",
            "--file",
            str(path),
            "--list",
            "--pretty=false",
        ],
        cwd=root,
        capture_output=True,
        text=True,
        timeout=60,
        check=False,
    )
    if process.returncode:
        raise RuntimeError(process.stderr.strip() or f"cannot parse {path}")
    value = json.loads(process.stdout)
    return value if isinstance(value, list) else []


def intersects(start: int, end: int, line: int, count: int) -> bool:
    if count == 0:
        return start <= line <= end + 1
    return start <= line + count - 1 and line <= end


def base_file(root: Path, base: str, source: str) -> Path | None:
    process = subprocess.run(
        ["git", "show", f"{base}:{source}"],
        cwd=root,
        capture_output=True,
        check=False,
    )
    if process.returncode:
        return None
    descriptor, name = tempfile.mkstemp(suffix=".go")
    os.close(descriptor)
    path = Path(name)
    path.write_bytes(process.stdout)
    return path


# 인덱스의 일반 파일 모드 둘. 그 밖(심링크 `120000` · gitlink `160000`)은 Go 소스가 아니다 — 저장소 전수 0 / 1,762.
GO_FILE_MODES = (b"100644", b"100755")


def _temporary_go(data: bytes) -> Path:
    """바이트를 `go_functions` 가 읽을 임시 `.go` 로 쓴다. 호출자가 지운다."""
    descriptor, name = tempfile.mkstemp(suffix=".go")
    with os.fdopen(descriptor, "wb") as handle:
        handle.write(data)
    return Path(name)


def _write_loose_blob(store: Path, digest: str, data: bytes) -> None:
    """blob 하나를 **임시** 객체 저장소에 loose 객체로 쓴다 — `"blob <크기>\\0" + 바이트` 를 zlib 로 압축한 것.

    `git hash-object -w` 를 안 쓰는 까닭은 둘이다: git 이 파일을 **다시** 읽지 않아야 판정한 바이트가 git 에
    건넨 바이트이고(지문 = 판정한 바이트), 편집 파일 수만큼 프로세스를 띄우지 않는다.
    """
    folder = store / digest[:2]
    folder.mkdir(parents=True, exist_ok=True)
    with open(folder / digest[2:], "wb") as handle:
        handle.write(zlib.compress(b"blob %d\0" % len(data) + data))


def _c_quoted(text: str) -> str:
    """git 이 대체 저장소 목록에서 읽는 C 인용 (task 7.5.31). `"` 로 시작하는 항목은 인용을 풀어 읽는다 — 그래서
    경로에 목록 구분자(`:`)나 `"` 가 있어도 한 항목으로 적힌다(git 2.43 실측)."""
    escaped = text.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n").replace("\t", "\\t")
    return f'"{escaped}"'


@contextlib.contextmanager
def _worktree_snapshot(root: Path) -> Iterator[tuple[dict[str, str], dict[str, bytes], dict[bytes, str]]]:
    """워킹트리 대상 판정이 견줄 **바이트**를 한 번 읽어 임시 인덱스로 git 에 건넨다 (task 7.5.25).

    `git diff <base>` 는 워킹트리를 **git 의 투영**으로 본다 — 비교 전에 clean·process 필터 · `ident` ·
    `working-tree-encoding` 이 바이트를 다시 쓰고, stat 캐시 · fsmonitor · 인덱스 플래그는 파일을 아예 안 읽게
    한다. 그 문들은 `--numstat` 과 판정 diff 를 **함께** 속이므로 두 시야의 대조가 원리상 못 본다. 7.5.27 · 7.5.28
    은 문을 하나씩 닫았고 재리뷰는 매번 하나를 더 찾았다(마지막은 UTF-7 의 이중 표현). 설정이 전혀 없어도
    stat 캐시만으로 편집이 감춰졌다(같은 inode · 같은 크기 · mtime 되돌림 · 같은 초의 ctime, 10/10 실측).

    그래서 투영을 **안 쓴다**: 인덱스가 추적하는 `*.go` 마다 디스크 바이트를 깔때기(`_read_regular`)로 읽어
    — 원장에 남으므로 끝의 재확인이 다시 읽는다 — 그 바이트의 blob 을 **임시** 객체 저장소에 쓰고, 그 blob 을
    가리키는 **임시** 인덱스를 만든다. 두 diff 는 `--cached` 로 그 인덱스를 base 와 견준다 — 인덱스와 트리를
    견주는 diff 는 워킹트리도 필터도 stat 도 안 본다. 실제 인덱스와 객체 저장소에는 **아무것도 쓰지 않는다**
    (실제 저장소는 대체 저장소로 읽기만 한다).

    돌려주는 것: 두 diff 에 줄 환경, 경로 → 읽은 바이트(판정의 현재 쪽이 **이 바이트**를 읽는다), 그리고 임시 인덱스에
    실은 날 경로 → oid(git 밖의 대조가 쓴다 — `_snapshot_disagreement`, task 7.5.31).

    git 의 시야를 그대로 두는 자리는 하나다 — `skip-worktree` 인데 파일이 없으면 sparse-checkout 으로 **안 꺼낸**
    것이므로 인덱스의 blob 을 쓴다. `assume-unchanged` 인데 없으면 지운 것이다(플래그는 편집을 감추는 데만 쓰인다).
    """
    described = subprocess.run(
        ["git", "rev-parse", "--show-object-format", "--git-path", "objects"],
        cwd=root, capture_output=True, timeout=10, check=False,
    )
    if described.returncode:
        raise RuntimeError(_first_line(described.stderr.decode("utf-8", "replace"), "git rev-parse failed"))
    answer = described.stdout.decode("utf-8", "strict").split("\n")
    if len(answer) < 2 or not answer[0] or not answer[1]:
        raise RuntimeError("cannot read git rev-parse output")
    algorithm, objects = answer[0], (root / answer[1]).resolve()
    listed = subprocess.run(
        ["git", *SNAPSHOT_PINS, "ls-files", "-s", "-v", "-z", "--", "*.go"],
        cwd=root, capture_output=True, timeout=30, check=False,
    )
    if listed.returncode:
        raise RuntimeError(_first_line(listed.stderr.decode("utf-8", "replace"), "git ls-files failed"))
    with tempfile.TemporaryDirectory(prefix="check-analysis-") as scratch:
        store = Path(scratch) / "objects"
        store.mkdir()
        lines: list[bytes] = []
        contents: dict[str, bytes] = {}
        placed: dict[bytes, str] = {}
        for entry in listed.stdout.split(b"\0"):
            if not entry:
                continue
            head, _, raw_path = entry.partition(b"\t")
            fields = head.split()
            if len(fields) != 4 or not raw_path:
                raise RuntimeError("cannot read git ls-files -s -v record")
            tag, mode, oid, stage = fields
            # 파일 시스템의 해독(`surrogateescape`)으로 읽는다 — 이름이 UTF-8 이 아닌 **안 바뀐** 파일로 판정이 멈추지
            # 않게(7.5.25 재리뷰 F9). 바뀐 파일의 이름은 가드가 전처럼 엄격히 해독해 거절한다.
            path = os.fsdecode(raw_path)
            if stage != b"0":
                raise RuntimeError(f"Go file is unmerged: {path} — resolve the conflict before the gate")
            if mode not in GO_FILE_MODES:
                raise RuntimeError(f"tracked Go path is not a regular file in the index (mode {mode.decode()}): {path}")
            try:
                data = _read_regular(root / path)
            except (FileNotFoundError, NotADirectoryError):
                # `-v` 는 skip-worktree 를 `S` 로, assume-unchanged 를 **소문자**로 적는다(둘 다면 `s`).
                if tag.upper() == b"S":
                    lines.append(b"%s %s\t%s\0" % (mode, oid, raw_path))
                    placed[raw_path] = oid.decode("ascii")
                continue
            digest = hashlib.new(algorithm, b"blob %d\0" % len(data) + data).hexdigest()
            # 인덱스와 같은 blob 은 실제 저장소에 이미 있다(대체 저장소로 읽는다) — 다른 것만 쓴다.
            if digest != oid.decode("ascii"):
                _write_loose_blob(store, digest, data)
            contents[path] = data
            placed[raw_path] = digest
            lines.append(b"%s %s\t%s\0" % (mode, digest.encode("ascii"), raw_path))
        inherited = os.environ.get("GIT_ALTERNATE_OBJECT_DIRECTORIES", "")
        environment = {
            **os.environ,
            "GIT_INDEX_FILE": str(Path(scratch) / "index"),
            "GIT_OBJECT_DIRECTORY": str(store),
            # 인용해 적는다 (task 7.5.31) — 7.5.25 는 경로에 `:`·`"` 가 있으면 거절했다(헛거절, 재리뷰 F10).
            "GIT_ALTERNATE_OBJECT_DIRECTORIES": os.pathsep.join(filter(None, (_c_quoted(str(objects)), inherited))),
        }
        # 임시 인덱스를 쓸 때 실제 저장소로 새는 것을 끈다 — split index 는 `sharedindex.*` 를 `.git` 에 쓰고,
        # 인덱스 쓰기는 `post-index-change` 훅을, 인덱스 읽기는 fsmonitor 명령을 띄운다.
        built = subprocess.run(
            ["git", *SNAPSHOT_PINS, "-c", "core.splitIndex=false", "-c", "core.hooksPath=/dev/null",
             "update-index", "-z", "--index-info"],
            cwd=root, env=environment, input=b"".join(lines), capture_output=True, timeout=30, check=False,
        )
        if built.returncode:
            raise RuntimeError(_first_line(built.stderr.decode("utf-8", "replace"), "git update-index failed"))
        yield environment, contents, placed


# 인덱스를 읽는 git 호출(스냅숏의 `ls-files` · 두 diff)이 모두 받는 고정. `--cached` diff 는 워킹트리를 안
# 보지만, 인덱스를 **읽는** 순간 설정된 fsmonitor 명령이 뜬다(실제 인덱스를 읽는 `ls-files` 에서 실측) —
# 판정이 저장소가 설정한 프로그램을 띄울 까닭이 없다.
SNAPSHOT_PINS = ("-c", "core.fsmonitor=false")


def _compared(base: str, target: str) -> list[str]:
    """두 diff 가 견주는 쪽. 대상이 커밋이면 트리 둘, 워킹트리면 base 대 스냅숏 인덱스다 (task 7.5.25).

    두 호출이 **이 함수 하나**에서 받는다 — 가드와 판정이 다른 쌍을 견주면 두 시야의 대조가 성립하지 않는다.
    """
    return [base, target] if target else ["--cached", base]


def _header_name(line: str) -> str:
    """`--- a/x.go` · `+++ b/x.go` 머리 줄의 이름 (task 7.5.28).

    git 은 이름에 **공백이 있으면** 머리 줄 끝에 탭 하나를 붙인다(GNU diff 의 관례). 떼지 않으면
    `my file.go\\t` 를 base 에서 찾다가 `cannot load existing base file` 로 **거짓 차단**했다 — 이 change
    이전부터 있던 결함이고 재리뷰가 유니코드 줄 구분자를 재현하다 드러났다. 이름 가드가 `\\t` 든 이름을
    이미 거절하므로 끝의 탭 하나를 떼는 것은 모호하지 않다.
    """
    value = line[4:]
    return value[:-1] if value.endswith("\t") else value


def _numstat_records(raw_output: bytes) -> list[tuple[bytes, bytes, list[bytes]]]:
    """`git diff --numstat -z` 의 레코드를 (더함, 지움, 경로들)로 읽는다 (task 7.5.22).

    평범한 레코드는 `<더함>\\t<지움>\\t<경로>\\0` 이고, **rename 은 경로 칸이 비고 다음 두 칸**이
    옛 이름과 새 이름이다. 경로는 `-z` 라 인용되지 않으므로 탭이 든 이름도 `maxsplit=2` 로 온전히 나온다.
    모양이 다르면 지어내지 않고 결함으로 올린다 — 못 읽은 표를 "바뀐 파일 없음" 으로 읽으면 그것이
    이 task 가 닫는 바로 그 구멍이다.
    """
    chunks = raw_output.split(b"\0")
    records: list[tuple[bytes, bytes, list[bytes]]] = []
    index = 0
    while index < len(chunks):
        if not chunks[index]:
            index += 1
            continue
        parts = chunks[index].split(b"\t", 2)
        if len(parts) != 3:
            raise RuntimeError("cannot read git diff --numstat record")
        added, deleted, first = parts
        if first:
            records.append((added, deleted, [first]))
            index += 1
            continue
        pair = [item for item in chunks[index + 1:index + 3] if item]
        if len(pair) != 2:
            raise RuntimeError("cannot read git diff --numstat rename record")
        records.append((added, deleted, pair))
        index += 3
    return records


def _safe_changed_go_paths(
    root: Path, base: str, target: str, environment: dict[str, str] | None = None
) -> list[tuple[bytes, bytes, list[bytes]]]:
    """바뀐 `*.go` 를 **이름**과 **본문 유무** 둘로 거른다.

    이름: 통합 diff 헤더 문법이 무손실로 표현하지 못하는 것을 거절한다.

    본문: git 이 **본문을 안 낸** 파일을 거절한다 (task 7.5.22). `changed_existing_functions` 는
    훅(`@@`)으로만 "바뀐 기존 함수" 를 세므로, `.gitattributes` 한 줄(`*.go binary` 또는 `*.go -diff`)
    이면 git 이 `Binary files … differ` 를 내고 훅이 **0 개**가 되어 그 파일의 요구가 **조용히** 사라진다.
    그 `.gitattributes` 는 **추적될 필요조차 없다**. "훅이 0 개면 거절" 로는 못 가른다 — 정상인
    mode-only 변경도 훅이 0 개다. `--numstat` 이 정확히 가른다: 억제는 `-`/`-`, mode-only 는 `0`/`0`.

    `--name-only` 이 아니라 `--numstat` 을 읽으므로 rename 의 **양쪽 이름**을 다 본다. 파서가
    `base_file` 에 넘기는 것은 **옛** 이름인데 `--name-only` 은 그것을 내지 않았다.

    새 거절은 **가장 뒤에 선다** — 이름 검사를 전부 마친 뒤에 본문을 묻는다. 앞의 가드를 가리면
    그 가드의 시험이 남의 가드를 재게 된다.

    **레코드를 돌려준다 (task 7.5.23).** `--numstat` 이 가르는 것은 "git 이 본문을 냈는가" 가 아니라
    "git 이 이것을 이진으로 다루는가" 다 — 7.5.22 가 "정확히 가른다" 고 적은 것은 **거짓**이었다.
    `textconv` 필터는 `--numstat` 에 `1`/`1` 을 내면서 판정 diff 의 본문을 **통째로 지운다**.
    그래서 호출자가 이 레코드를 판정이 실제로 낸 훅과 **대조**한다(`changed_existing_functions`).
    `--find-renames` 를 여기도 준다 — 판정과 같은 짝을 봐야 대조가 성립한다.

    **워킹트리 대상은 스냅숏 환경이 있어야 한다 (task 7.5.25).** 없으면 `--cached` 가 **실제** 인덱스를
    견주게 된다 — 워킹트리도 스냅숏도 아닌 셋째 시야다. 그래서 받지 않는다.
    """
    if not target and environment is None:
        raise ValueError("a worktree comparison needs the worktree snapshot")
    process = subprocess.run(
        ["git", *SNAPSHOT_PINS, "diff", "--no-ext-diff", "--no-textconv", "--find-renames",
         "--numstat", "-z", *_compared(base, target), "--", "*.go"],
        cwd=root,
        env=environment,
        capture_output=True,
        check=False,
    )
    if process.returncode:
        stderr = process.stderr.decode("utf-8", "replace") if isinstance(process.stderr, bytes) else process.stderr
        raise RuntimeError(stderr.strip() or f"git diff failed for base {base}")
    raw_output = process.stdout if isinstance(process.stdout, bytes) else process.stdout.encode("utf-8")
    records = _numstat_records(raw_output)
    for _, _, paths in records:
        for raw in paths:
            try:
                path = raw.decode("utf-8", "strict")
            except UnicodeDecodeError as error:
                raise RuntimeError("modified Go path is not UTF-8") from error
            # Git quotes tab/newline headers; do not silently parse that quoted form
            # as a different path. Ordinary Unicode names remain supported.
            if "\n" in path or "\r" in path or "\t" in path:
                raise RuntimeError("modified Go path cannot be represented losslessly in unified diff")
    for added, deleted, paths in records:
        if added == b"-" and deleted == b"-":
            raise RuntimeError(
                "modified Go file has no textual diff (binary or -diff attribute): "
                + paths[-1].decode("utf-8", "strict")
            )
    return records


def changed_existing_functions(
    root: Path = ROOT,
    base: str = "",
    target: str = "",
) -> dict[tuple[str, str], dict]:
    if not base:
        raise ValueError("Function Logic Map comparison base is required")
    # 워킹트리 대상이면 git 의 워킹트리 투영을 **안 믿는다** — 디스크 바이트를 한 번 읽어 만든 스냅숏
    # 인덱스를 두 diff 가 같이 본다 (task 7.5.25). 대상이 커밋이면 트리 둘을 견주므로 스냅숏이 없다.
    if target:
        return _changed_existing_functions(root, base, target, None, {}, None)
    with _worktree_snapshot(root) as (environment, contents, placed):
        return _changed_existing_functions(root, base, target, environment, contents, placed)


def _snapshot_disagreement(
    root: Path, base: str, placed: dict[bytes, str], records: list[tuple[bytes, bytes, list[bytes]]]
) -> None:
    """게이트가 **스스로 해시한** oid 와 base 트리의 oid 가 다른 경로는 diff 에 내용 변경으로 나와야 한다 (task 7.5.31).

    스냅숏은 디스크 바이트를 바르게 읽지만, 두 diff 는 그 blob 을 **oid 로** 다시 찾는다 — 실제 저장소에 그 oid 로
    base 의 바이트를 담은 객체(loose · pack)가 있으면 git 은 그것을 읽는다. git 은 pack 을 먼저 보므로 스냅숏이 객체를
    늘 써도 못 막는다(7.5.25 적대 재리뷰 F2 · F3 실측 — 7.5.25 가 연 회귀다). 두 diff 가 **함께** 속으므로 대조의
    한쪽은 git 이 아니어야 한다: 스냅숏의 oid 는 Python 이 바이트에서 계산했다.

    규칙 셋. (1) 스냅숏에 있고 oid 가 base 와 다르거나 base 에 없는 경로는 어떤 레코드의 **새** 이름이어야 한다.
    (2) base 에 있고 스냅숏에 없는 경로는 어떤 레코드의 **옛** 이름이어야 한다. (3) 이름이 안 바뀐 레코드인데 oid 가
    다르면 `0`/`0` 이면 안 된다 — git 이 두 쪽 내용이 같다고 본 것이다. base 트리 자체는 믿는다(그 위조는 7.5.33).
    """
    listed = subprocess.run(
        ["git", "ls-tree", "-r", "-z", "--full-tree", base],
        cwd=root, capture_output=True, timeout=30, check=False,
    )
    if listed.returncode:
        raise RuntimeError(_first_line(listed.stderr.decode("utf-8", "replace"), "git ls-tree failed"))
    before: dict[bytes, str] = {}
    for entry in listed.stdout.split(b"\0"):
        if not entry:
            continue
        head, _, raw_path = entry.partition(b"\t")
        fields = head.split()
        if len(fields) != 3 or not raw_path:
            raise RuntimeError("cannot read git ls-tree record")
        # 종류(blob · commit)로 거르지 않는다 — base 의 `*.go` 가 gitlink · 심링크면 워킹트리에 남아 있을 때는 스냅숏이
        # 모드로 먼저 거절하고, 지웠을 때는 diff 가 그 삭제를 보여 준다. 거르는 줄을 빼는 변이가 살아남아 지웠다 (7.5.31).
        if raw_path.endswith(b".go"):
            before[raw_path] = fields[2].decode("ascii")
    new_names = {paths[-1] for _, _, paths in records}
    old_names = {paths[0] for _, _, paths in records}
    unmoved = {paths[0]: (added, deleted) for added, deleted, paths in records if len(paths) == 1}
    for raw_path, oid in sorted(placed.items()):
        if before.get(raw_path) == oid:
            continue
        if raw_path not in new_names or unmoved.get(raw_path) == (b"0", b"0"):
            raise RuntimeError(
                "the object store answers for " + os.fsdecode(raw_path)
                + " with bytes that are not on disk — git's diff does not show the edit the gate hashed"
            )
    for raw_path in sorted(set(before) - set(placed)):
        if raw_path not in old_names:
            raise RuntimeError(
                "the object store answers for " + os.fsdecode(raw_path)
                + " as if it were still there — git's diff does not show its deletion"
            )


def _changed_existing_functions(
    root: Path, base: str, target: str, environment: dict[str, str] | None, contents: dict[str, bytes],
    placed: dict[bytes, str] | None,
) -> dict[tuple[str, str], dict]:
    """`changed_existing_functions` 의 몸통. 워킹트리 대상이면 스냅숏이 열려 있는 동안 돈다 — 임시 저장소가 두 diff
    보다 오래 산다. 커밋 대상이면 스냅숏 없이(`environment`·`placed` 가 `None`) 돈다."""
    records = _safe_changed_go_paths(root, base, target, environment)
    process = subprocess.run(
        [
            "git",
            *SNAPSHOT_PINS,
            "-c",
            "core.quotePath=false",
            "diff",
            # 이 셋이 **본문을 지우는 문**을 닫는다 (task 7.5.23). `--no-ext-diff` 는 외부 diff
            # 명령을, `--no-textconv` 는 `diff=<드라이버>` 의 textconv 를 끈다 — 둘 다 없으면
            # git 이 훅을 **0 개** 내고 그 파일의 요구가 조용히 사라진다. 깃발이 사라져도
            # 아래 교차 검사가 런타임에 잡는다(깃발과 검사 둘 다 시험이 못 박는다).
            "--no-ext-diff",
            "--no-textconv",
            "--find-renames",
            "--unified=0",
            *_compared(base, target),
            "--",
            "*.go",
        ],
        cwd=root,
        env=environment,
        capture_output=True,
        timeout=30,
        check=False,
    )
    if process.returncode:
        raise RuntimeError(
            process.stderr.decode("utf-8", "replace").strip() or f"git diff failed for base {base}"
        )
    # **바이트로 받아 `\n` 에서만 자른다** (task 7.5.28). `text=True` 는 `\r` 을 줄바꿈으로 바꾸고,
    # `str.splitlines()` 는 U+2028 · U+2029 · U+0085 에서도 자른다 — `core.quotePath=false` 라 git 은 그 글자를
    # 인용하지 않으므로, 경로에 그 글자가 있으면 `diff --git` 머리가 잘려 뒷조각이 훅으로 읽히고 그 파일의
    # 요구가 통째로 사라졌다(재리뷰 재현). 통합 diff 의 줄 경계는 `\n` 하나뿐이다. 해독은 엄격하다 —
    # UTF-8 이 아닌 diff 는 전처럼 결함이다.
    diff_lines = process.stdout.decode("utf-8", "strict").split("\n")
    required: dict[tuple[str, str], dict] = {}
    # 판정 diff 의 **구역마다** 본문(훅)이 있었는지. numstat 레코드와 **순서로** 짝짓는다 —
    # 이름으로 짝지으면 안 된다 (task 7.5.24): 이 파서가 아는 이름은 `removeprefix` 를 거친
    # **유도된** 것이고 git 이 인용한 것(`"a/we\"ird.go"`)일 수도 있는데, numstat 의 이름은
    # `-z` 라 날 바이트다. 두 이름 공간을 교집합으로 견주면 정상 입력이 거절된다(실측).
    bodied: list[bool] = []
    old_source = ""
    new_source = ""
    hunks: list[tuple[int, int, int, int]] = []

    def flush() -> None:
        nonlocal hunks
        if not old_source or not hunks:
            hunks = []
            return
        temporary = base_file(root, base, old_source)
        if temporary is None:
            raise RuntimeError(
                f"cannot load existing base file {base}:{old_source}"
            )
        current: Path | None = None
        try:
            old_functions = go_functions(temporary, root)
            old_qualified = {qualified(function) for function in old_functions}
            for function in old_functions:
                start = int(function["start"]["line"])
                end = int(function["end"]["line"])
                if any(intersects(start, end, old, old_count) for old, old_count, _, _ in hunks):
                    key = (new_source or old_source, qualified(function))
                    required[key] = {
                        "file": new_source or old_source,
                        "function": qualified(function),
                        "base_hash": function.get("source_sha256"),
                    }
            # 현재 쪽은 **스냅숏의 바이트**를 읽는다 (task 7.5.25) — 디스크를 다시 읽으면 git 에 건넨 바이트와
            # 판정한 바이트가 갈릴 수 있다([[a-fingerprint-must-be-the-bytes-judged]]). 스냅숏에 없으면(지웠거나
            # sparse 로 안 꺼냈으면) 현재 논리가 없다.
            if target:
                current = base_file(root, target, new_source) if new_source else None
            else:
                current = _temporary_go(contents[new_source]) if new_source in contents else None
            if current and current.exists():
                for function in go_functions(current, root):
                    # Function Logic Maps are required for functions that existed
                    # at the frozen comparison base and whose logic changed. A
                    # newly added function in an existing file has no base logic
                    # to map; treating it as "modified existing" contradicts the
                    # workflow contract and makes added test helpers look like
                    # high-risk edits.
                    if qualified(function) not in old_qualified:
                        continue
                    start = int(function["start"]["line"])
                    end = int(function["end"]["line"])
                    if any(
                        intersects(start, end, new, new_count)
                        for _, _, new, new_count in hunks
                    ):
                        key = (new_source, qualified(function))
                        required[key] = {
                            "file": new_source,
                            "function": qualified(function),
                            "current_hash": function.get("source_sha256"),
                        }
        finally:
            temporary.unlink(missing_ok=True)
            if current is not None:
                current.unlink(missing_ok=True)
        hunks = []

    # 통합 diff 는 문법이고 그 문법에는 **상태**가 있다 (task 7.5.2.4, 7.5.2.3 재리뷰 보안).
    # `--unified=0` 이라 문맥 줄이 없고 본문은 전부 `-`·`+`·`\` 로 시작한다. 그래서 열 0 의
    # `diff --git ` 과 `@@` 는 모호하지 않지만 `--- `·`+++ ` 는 **모호하다** — 지워진 소스 줄
    # `-- x` 가 `--- x` 로, 더한 소스 줄 `++ x` 가 `+++ x` 로 나오기 때문이다(진짜 git 실측).
    # 상태 없이 읽던 판본은 파일 **중간에서** 이름을 바꿔, 그 파일의 요구를 통째로 지우거나
    # (`/dev/null` 모양) 편집 전 논리의 지도로 내려앉혔다(`revision: base` 모양).
    # 이름은 **첫 훅 앞에서만** 읽는다. 새로 거절하는 입력은 없다 — 본문에서 이름을 안 읽을 뿐이다.
    in_body = False

    def hunk(line: str) -> bool:
        match = HUNK.match(line)
        if match is None:
            return False
        hunks.append(
            (
                int(match.group(1)),
                int(match.group(2) or 1),
                int(match.group(3)),
                int(match.group(4) or 1),
            )
        )
        return True

    for line in diff_lines:
        if line.startswith("diff --git "):
            flush()
            old_source = ""
            new_source = ""
            in_body = False
            bodied.append(False)
        elif in_body:
            # 본문이다. 여기서 `--- `·`+++ ` 는 소스 줄이지 파일 이름이 아니다.
            hunk(line)
        elif hunk(line):
            # 파일의 **첫** 훅이다. 본문을 냈다는 표시는 여기 한 번이면 된다 — 뒤의 훅은 같은
            # 구역에 속하므로 더 말할 것이 없다 (task 7.5.23).
            in_body = True
            if bodied:
                bodied[-1] = True
        elif line.startswith("--- "):
            value = _header_name(line)
            old_source = "" if value == "/dev/null" else value.removeprefix("a/")
        elif line.startswith("+++ "):
            value = _header_name(line)
            new_source = "" if value == "/dev/null" else value.removeprefix("b/")
    flush()
    # **가드가 센 것과 판정이 읽은 것을 맞춰 본다** (task 7.5.23). numstat 이 내용이 바뀌었다고
    # (`0`/`0` 도 `-`/`-` 도 아니라고) 말한 파일은 본문을 내야 한다. 안 냈으면 무언가가 본문을
    # 지운 것이고 — textconv · 외부 diff · 아직 모르는 git 기능 — 그 파일의 요구는 **조용히**
    # 사라진다. 문을 하나씩 세는 대신 **두 투영이 어긋났다**는 것을 본다.
    #
    # 짝은 **순서**로 짓는다 (task 7.5.24). 두 호출은 같은 `git diff` 에 형식만 다르므로 파일을
    # 같은 순서로 낸다(섞인 픽스처 여섯 모양으로 실측). 이름으로 짝지으려던 앞 판본은 정상
    # 입력을 거절했다 — 이 파서의 이름은 `removeprefix` 를 거친 유도값이고 git 이 인용한 것일
    # 수도 있는데 numstat 의 이름은 날 바이트라, 두 이름 공간이 안 만난다.
    if len(bodied) != len(records):
        raise RuntimeError(
            f"git listed {len(records)} changed Go file(s) but the judged diff has "
            f"{len(bodied)} — the two views of the same diff disagree"
        )
    for (added, deleted, paths), had_body in zip(records, bodied):
        if added in (b"-", b"0") and deleted in (b"-", b"0"):
            continue
        if not had_body:
            raise RuntimeError(
                "git reported content changes but emitted no diff body for "
                + paths[-1].decode("utf-8", "strict")
                + " — an external diff or textconv filter is hiding it"
            )
    # git **밖의** 대조는 가장 뒤에 선다 (task 7.5.31) — 앞의 가드들이 먼저 말해야 그 시험들이 이것을 안 잰다.
    # 가드가 읽은 **같은** 레코드를 쓴다: 다시 물으면 두 번째 대답과 대조하게 된다.
    if placed is not None:
        _snapshot_disagreement(root, base, placed, records)
    return required


# `openspec archive` 는 끝난 change 를 `archive/<YYYY-MM-DD>-<id>` 로 옮긴다.
# 접미사로 고르면 `2026-08-29-other-reference` 가 `reference` 로 통과하므로
# 날짜 접두사를 벗긴 나머지를 **전부** 맞춘다.
ARCHIVED_CHANGE = re.compile(r"\d{4}-\d{2}-\d{2}-(?P<change>.+)")
ARCHIVE_PREFIX = "openspec/changes/archive/"


def _archived_change_id(name: str) -> str:
    """아카이브 디렉터리 이름 `<YYYY-MM-DD>-<id>` 이면 `<id>`, 아니면 빈 문자열.

    이 해독은 **여기 한 곳**에 산다 (task 7.6, 리뷰 I6). 해소기(id → 디렉터리)와
    `_pre_archive_path`(아카이브 경로 → 옮기기 전 경로)가 각자 정규식을 들고 있었고,
    6.2 가 `validate` 에 "셋째를 만들지 않는다"고 적은 바로 그 모양이었다. shell 쪽
    사본(`tools/gate.sh`)은 언어가 달라 합칠 수 없어 자기 시험이 따로 못 박는다.
    """
    matched = ARCHIVED_CHANGE.fullmatch(name)
    return matched.group("change") if matched else ""


def resolve_referenced_change(root: Path, change: str) -> Path:
    """증거를 빌려주는 change 의 디렉터리를 찾는다.

    빌려주는 쪽이 먼저 아카이브되면 `changes/<id>` 는 사라진다. 거기만 보면
    빌리는 쪽의 게이트가 영원히 막히므로 아카이브도 본다 — a073 이 a072 의
    번들을 빌려 쓰는데 a072 가 먼저 아카이브되어 실제로 그렇게 됐다.

    **활성을 찾아도 아카이브를 마저 센다.** 예전에는 활성이 있으면 거기서
    바로 돌려줘서 아카이브를 열어 보지도 않았고, 그래서 중복 판정의 범위가
    아카이브 안으로 좁아져 있었다. 아카이브 뒤 같은 id 로 디렉터리를 다시 만들면
    조용히 그쪽이 이겼다는 뜻이다 — 고르면 어느 증거로 게이트가 열렸는지
    기록에 남지 않는다. `tools/gate.sh` 의 해소기는 처음부터 이렇게 셌다.
    """
    # 고르기도 깔때기로 (task 7.5.2.3) — 어느 디렉터리를 판정하는지가 판정의 첫 입력이고, `is_dir()` 은
    # `OSError` 를 삼켜 "못 물었다" 를 "없다" 로 만든다. 디렉터리가 아닌 것만 "없다" 이고 못 여는 것은 결함이다.
    changes = root / "openspec" / "changes"
    direct = changes / change
    open_here = _kind(direct) == "dir"
    archive = changes / "archive"
    try:
        entries = _listed(archive)
    except (FileNotFoundError, NotADirectoryError):
        entries = []
    archived = [archive / name for name, is_dir in entries
                if is_dir and _archived_change_id(name) == change]
    found = ([direct] if open_here else []) + archived
    if not found:
        # 이 해소기는 빌린 증거만이 아니라 **게이트 대상**도 찾는다 (task 7.6, I4).
        # 그래서 문장에 "reference" 를 안 붙인다 — 오타 난 대상 id 에도 이 문장이 나간다.
        raise ValueError(f"change is neither open nor archived: {change}")
    if open_here and archived:
        # 따로 타입(`AmbiguousChange`)을 두던 이유는 호출자 둘이 "못 찾음"은 fallback 으로
        # 흘리고 "모호함"만 멈춰야 했기 때문이다. 그 fallback 이 없어져서(task 7.6, I4)
        # 가를 호출자가 0 이 됐다 — 이제 모든 실패가 같은 방식으로 멈춘다.
        raise ValueError(
            f"{change} is open and archived at once: "
            + ", ".join(path.relative_to(root).as_posix() for path in found)
        )
    if len(archived) > 1:
        # 고르면 어느 증거로 통과했는지 기록에 안 남는다. 세어서 멈춘다.
        raise ValueError(
            f"archive holds {len(archived)} copies of {change}: "
            + ", ".join(path.name for path in archived)
        )
    return found[0]


def resolve_base(
    change_dir: Path, root: Path, context: dict[str, object] | None = None,
    *, change_id: str,
) -> str:
    path = change_dir / "base-commit.txt"
    try:
        # 깔때기로 (task 7.5.2.3) — 옛 `read_text` 는 그 자리의 FIFO 에 영원히 멎었고, 창의 **시작**을 정하는
        # 이 읽기가 원장에 없으면 실행 중 바뀌어도 재확인이 못 본다. 없는 것과 못 읽는 것을 가른다.
        candidate = _decoded(_read_regular(path)).strip()
    except FileNotFoundError as exc:
        raise ValueError(
            "missing base-commit.txt; run "
            "`python3 tools/sdd/capture_change_base.py --change <change-id>` "
            "at proposal freeze"
        ) from exc
    except (OSError, UnicodeDecodeError) as exc:
        raise ValueError(UNREADABLE.format(what="base-commit.txt", why=_why(exc))) from exc

    def resolve(value: str) -> str:
        process = subprocess.run(
            ["git", "rev-parse", "--verify", f"{value}^{{commit}}"],
            cwd=root,
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )
        if process.returncode:
            raise ValueError(
                process.stderr.strip() or f"invalid Function Logic Map base: {value}"
            )
        return process.stdout.strip()

    persisted = resolve(candidate)
    try:
        # 이관 신원은 디렉터리 이름이 아니라 요청받은 id 로 가른다 — 아카이브가 이름을
        # 바꾼다(task 6.2). 필수 인자인 이유: 기본값이 있으면 id 를 잊은 호출자가 조용히
        # 옛 판정(이름)으로 떨어지고, 그것이 아카이브된 a063 을 막던 바로 그 판정이다.
        adoption = validate_execution_baseline(change_dir, root, persisted, change_id)
    except AdoptionError as exc:
        raise ValueError(f"invalid execution-baseline adoption: {exc}") from exc
    effective = str(adoption["effective_base"]) if adoption else persisted
    if context is not None:
        context["execution_baseline_adoption"] = adoption is not None
        context["effective_base"] = effective
        if adoption:
            # `validate` 는 감사된 창의 **끝**을 이미 돌려준다. 옛 판본은 그 값을
            # 버리고 `landed-commit.txt` 를 찾았고, 없으니 대상이 워킹트리가 됐다.
            context["adoption_source"] = str(adoption["source"])
    override = os.environ.get("SDD_BASE_REF", "").strip()
    if override and resolve(override) != effective:
        raise ValueError("SDD_BASE_REF must resolve to the selected effective comparison base")
    return effective


# `git cat-file --batch` 의 `-Z`(입력·출력 **둘 다** NUL 로 끊는다)가 들어온 버전.
# 이보다 낮은 git 은 그 옵션을 모른다며 rc 129 로 멈추고, 게이트는 그 **git 의 말**을 담은
# 결함으로 멈춘다 — 착지 기록을 읽는 자리가 모든 change 에 있으므로 모든 change 가 멈춘다
# (task 7.5.1 이전에는 실패를 "파일 없음" 과 섞어 사유를 저자의 증거 탓으로 오진했다).
# 값을 여기 한 곳에 적어 두고 README·WORKFLOW 가 같은 수를 인용한다.
# 영수증: git 자신의 릴리스 노트 `RelNotes/2.42.0.txt` — *"git cat-file --batch" and friends
# learned "-Z" that uses NUL delimiter for both input and output.* (처음엔 기억으로 적었다가
# 2026-09-19 gstack 리뷰 F9 가 근거를 물어 이 줄로 확인했다.)
GIT_BATCH_MINIMUM = "2.42"
# `git cat-file --batch` 가 **찾은 객체**에 대해 내는 머리. 없는 것은 `<물은 spec> missing` 이다
# (2026-09-19 실측: blob · tree · 심링크 · gitlink · 없는 경로 · HEAD · 짧은 ref 전부 이 둘 중 하나).
_OBJECT_HEADER = re.compile(
    rb"(?P<oid>[0-9a-f]{40}|[0-9a-f]{64}) (?P<type>blob|tree|commit|tag) (?P<size>[0-9]+)"
)
# 새 git 이 커밋이 없는 gitlink 에 내는 머리 — 내용이 없고 물은 spec 대신 oid 를 적는다
# (task 7.5.2, 재리뷰 Codex P2). 영수증: git 소스 `builtin/cat-file.c` 의
# `if (data->mode == S_IFGITLINK) report_object_status(opt, NULL, &data->oid, "submodule");` 와
# `printf("%s %s%c", obj_name ? obj_name : oid_to_hex(oid), status, …)`. 이 기계의 git 2.43.0 은
# 같은 경우에 `<spec> missing` 을 낸다 — 두 모양 모두 blob 이 아니다(`None`).
_SUBMODULE_HEADER = re.compile(rb"(?:[0-9a-f]{40}|[0-9a-f]{64}) submodule")
# 응답이 끝까지 안 온 두 자리(머리 · 내용)가 **한 문장**을 쓴다 — 두 벌이면 한쪽만 고쳐진다
# (F4 가 이미 한 번 고쳤다). 센 것은 **읽은 레코드**다 — `missing` 도 레코드다.
_TRUNCATED = (
    "`git cat-file --batch -Z` answered {count} of {total} request(s) at {ref}: "
    "the response is truncated"
)
# 착지를 판정하는 동안 증거(존재 · 번들 목록 · 바이트 · 고정 목록)가 움직였다는 문장 (task 7.5.1 · 7.5.2).
# 수락 직전 재확인과 선언 경로의 거절 앞 재확인이 **같은 문장**을 쓴다. 역사는 여기 없다 — 명령마다
# 한 번 푼 sha 로 고정하므로 움직일 것이 없다 (task 7.5.2.1).
INPUTS_MOVED = (
    "the evidence changed while the landing was being judged — a bundle was added, removed or "
    "rewritten mid-run; run it again"
)
LANDING_FILE = "landed-commit.txt"
FULL_SHA = re.compile(r"[0-9a-f]{40}")
# 판정을 내놓기 직전에 판정한 역사나 증거가 달라졌다는 문장 (task 7.5.2.2). `check` 의 출구와 기록 명령의 쓰기
# 직전이 **같은 문장**을 쓴다. `{what}` 은 무엇이 움직였는지다(`HEAD` 의 두 sha · 증거).
JUDGED_STATE_MOVED = (
    "{what} while this change was being judged — the verdict would describe a state that is no longer "
    "the one checked out; run it again"
)
# 이관 경로가 착지 기록을 거절하는 문장. 판정 경로(`check`)와 기록 경로(`record_landing`)가
# 한 벌씩 들고 있던 동안 이미 "ends" / "already ends" 로 갈렸다 (task 7.6, 리뷰 I5).
ADOPTION_REFUSES_A_LANDING = (
    f"execution-baseline adoption does not accept a `{LANDING_FILE}` record: "
    "the window ends at the audited source commit"
)
# 증거를 빌리는 change 가 착지 기록을 거절하는 문장 (task 7.2.3, 리뷰 C4). 판정 경로와 기록 경로가
# **같은 문장**을 쓴다 — 두 벌이면 갈린다(위 이관 문장이 그랬다).
BORROWED_REFUSES_A_LANDING = (
    f"a change that borrows its evidence does not accept a `{LANDING_FILE}` record: "
    "a borrowed window is never narrowed, because the lender's evidence cannot pin where "
    "this change's own Go work landed"
)
# 적힌 기록이 거절당한 저자가 돌아가는 길 (task 7.2.6). **한 문장**을 판정 경로와 기록 명령이
# 같이 쓴다 — 두 벌이면 갈린다. 이 길은 원래도 열려 있었고(기록을 지우면 다시 기록된다) 아무
# 규칙도 느슨하게 하지 않는다: 다시 계산한 착지는 갱신된 번들이 세우는 새 하한 뒤에 선다.
# 없던 것은 **문장**뿐이었다 — 성실하게 번들을 갱신한 저자가 "이 증거가 기술하는 리비전이
# 아니다"와 "이미 있다 — 덮어쓰지 않았다" 사이에 갇혔다.
LANDING_RECOVERY = (
    f"to move a record the evidence has outgrown: refresh the bundles, remove `{LANDING_FILE}` "
    "in a commit, then run `--record-landing` again (a record is never overwritten, so the "
    "value it had stays in the history)"
)
# 판정을 못 내는 입력이 **판정 대신 traceback** 이 되지 않게 하는 목록 (task 7.4, H5·H6).
# `subprocess.SubprocessError` 가 여기 있는 이유: `TimeoutExpired` 는 `OSError` 가 아니라서
# 자리마다 적혀 있던 목록 넷(`OSError, RuntimeError, ValueError, JSONDecodeError`)을 그대로
# 빠져나갔다 — git 이 한 번 멎으면 5단계는 창 줄도 안 찍고 스택만 남겼다.
# `json.JSONDecodeError` 는 `ValueError` 의 하위형이라 따로 안 적는다.
# 목록은 **한 곳에만** 둔다 — 자리마다 베껴 두면 한 자리를 고쳐도 다른 자리의 시험이
# 초록으로 남는다([[two-judgements-cover-for-each-other]]).
GATE_FAULTS = (OSError, RuntimeError, ValueError, subprocess.SubprocessError)


def _committed_many(root: Path, ref: str, relatives: list[str]) -> dict[str, bytes | None]:
    """`ref` 시점의 여러 파일 내용을 **한 프로세스**로 읽는다. 없으면 그 자리가 `None` 이다.

    이 함수가 있는 이유는 fetch **단위** 때문이다 (task 7.5, 리뷰 I1). 도구는 blob 을
    `git show` 로 한 프로세스에 하나씩 읽었고, 후보 하나를 판정하는 데 번들 수만큼
    프로세스를 썼다. 2026-09-18 실측: 아무 후보도 안 받는 walk 에서 spawn 의 **97.1~97.3%**
    가 blob fetch 였다(a071 은 12,155 중 11,799 · 73.19s). 한 후보당 83.4ms 가 8.5ms 가 된다.

    `-Z` 인 이유는 **경로가 문자열이기 때문**이다. 입력을 줄로 끊으면 개행이 든 경로가
    둘로 쪼개져 엉뚱한 blob 이나 `missing` 이 되고, 출력을 줄로 끊으면 tree 의 raw 바이트
    안에 있는 개행이 응답을 쪼갠다. `-Z` 는 양쪽을 NUL 로 끊는다 — 요구 버전은
    `GIT_BATCH_MINIMUM` 한 곳에 산다(여기 숫자를 다시 적으면 두 벌이 된다,
    [[two-judgements-cover-for-each-other]]).

    **blob 만 내용으로 친다.** `git show <ref>:<디렉터리>` 는 트리 **목록**을 찍는데, 그
    바이트가 판정에 들어오면 파일 내용인 척한다. 여기서는 `None` 이다 — 막는 쪽으로 엄해진다.
    오늘 실물 입력(번들의 `file` · 번들 `ast.json` 경로)에 디렉터리는 0 건이다
    ([[fail-closed-must-name-what-it-rejects]]).

    **`None` 은 git 의 답이다 — 그 spec 에 `missing`(또는 커밋 없는 gitlink 의 `submodule`)이라고
    답했거나 blob 이 아니다.** 트리 안에 있는데 객체를 못 읽는 경우(깨진 객체 · 가져오기에 실패한
    partial clone)도 git 은 rc 0 으로 `missing` 이라 답한다 — 이 함수가 그것을 가를 수는 없고,
    `None` 의 소비자는 전부 거절 · 누락 · 가장 넓은 창 쪽으로 읽는다(재리뷰 보안 전문가 실측).
    "못 물었다" 는 `None` 이 아니라 `GATE_FAULTS` 안의 예외다 — 아래 결함 여섯은 `RuntimeError`,
    프로세스를 못 띄우거나 멎으면 `OSError` · `SubprocessError`. 둘을 섞던 판본에서 가드 7 이
    **편집 전 커밋을 착지로 기록했고**(permissive), `_landing_record` 는 거절 하나를 건너뛰고
    가장 넓은 창으로 갔다 (task 7.5.1, 2026-09-19 gstack 리뷰). 결함 여섯: 프로세스 실패(rc≠0,
    git 의 stderr 를 담는다) · 요청에 NUL · 응답 잘림 · 모르는 머리 · 내용 뒤 NUL 종단 없음 ·
    안 읽힌 바이트. 전부 조용히 넘기면 **다른 파일의 바이트**나 부분 답이나 부재가 판정에
    들어간다. 경계의 `GATE_FAULTS` 가 오류 줄로 바꾼다 ([[a-fault-must-become-a-verdict]]).
    """
    wanted = list(dict.fromkeys(relatives))     # 순서 유지 + 중복 제거(같은 소스를 적은 번들 여럿)
    found: dict[str, bytes | None] = {relative: None for relative in wanted}
    if not wanted:
        return found
    specs = [f"{ref}:{relative}" for relative in wanted]
    for spec in specs:
        if "\0" in spec:
            # 요청 자체가 NUL 로 끊기므로 NUL 은 레코드를 쪼개고, 그 뒤 자리들이 **다른 파일의
            # 바이트**를 받는다. 경로만 보던 판본은 `ref` 를 안 봤다 (gstack 리뷰 F7).
            raise RuntimeError(f"request contains a NUL byte, cannot be asked for: {spec!r}")
    # `surrogateescape` 인 이유: 파일 이름은 바이트다. 옛 판본은 경로를 argv 로 넘겨
    # `os.fsencode` 를 탔으므로 디코딩 불가능한 이름도 그대로 갔다. 엄격한 `utf-8` 로
    # 인코딩하면 그런 이름 앞에서 **판정 대신 traceback** 이 된다.
    asked = [spec.encode("utf-8", "surrogateescape") for spec in specs]
    process = subprocess.run(
        ["git", "cat-file", "--batch", "-Z"],
        cwd=root, input=b"".join(spec + b"\0" for spec in asked),
        capture_output=True, timeout=60, check=False,
    )
    if process.returncode:
        # **rc≠0 은 "없다" 가 아니라 결함이다** (task 7.5.1 — gstack 리뷰 F1 · Codex P1-1).
        # `cat-file --batch` 는 없는 것을 `missing` 으로 **답 안에** 적고 rc 는 0 이다. 그러니
        # 여기 rc≠0 은 명령이 죽은 것이다. 예전에는 전부 `None` 을 돌려줬고 "부르는 쪽이 `None` 을
        # 불일치로 센다" 고 적었는데 **틀렸다**: 가드 7 은 한쪽만 실패한 `None != bytes` 를
        # "소스가 바뀌었다" 로 읽어 편집 전 커밋을 착지로 **기록했고**, `_landing_record` 는
        # `None` 을 "기록 없음" 으로 읽어 착지 검증을 **건너뛰었다**. 1.4 수리 때 이것을 결함으로
        # 올리려다 "시험 21개가 저장소 아닌 곳에서 돈다" 는 이유로 되돌렸는데, 다시 재니 그것은
        # git 을 mock 하려는 픽스처가 `_landing_record` 만 빠뜨린 것이었다.
        # 이름은 **git 의 말**로 댄다 — 버렸던 stderr 가 오진을 없앤다(F6). 버전 조언은 git 이
        # **옵션을 모른다**고 할 때(rc 129 — git 의 사용법 오류)만 붙인다. 저장소가 아닌 루트
        # (rc 128)에도 붙이던 판본은 git 의 말 옆에 추측한 진단을 다시 달았다(task 7.5.2, 재리뷰).
        said = _first_line(process.stderr, "")   # git 의 말 첫 줄은 한 벌이다 (task 7.5.2.2)
        raise RuntimeError(
            f"cannot read blobs at {ref[:12]}: `git cat-file --batch -Z` failed "
            f"(rc {process.returncode}" + (f": {said}" if said else "") + ")"
            + (f" — `-Z` needs git {GIT_BATCH_MINIMUM} or newer" if process.returncode == 129 else "")
        )
    data, position = process.stdout, 0
    answered: dict[str, bytes] = {}
    for index, (relative, request) in enumerate(zip(wanted, asked)):
        end = data.find(b"\0", position)
        if end < 0:
            # **부분 답을 쓰지 않는다** (독립 리뷰 2026-09-18, P0).
            raise RuntimeError(_TRUNCATED.format(count=index, total=len(wanted), ref=ref[:12]))
        header = data[position:end]
        position = end + 1
        if header == request + b" missing" or _SUBMODULE_HEADER.fullmatch(header):
            # 없는 것 둘: 물은 spec **그대로**의 `missing`, 그리고 커밋 없는 gitlink — 이쪽은 spec 대신
            # **oid** 가 돌아오므로 모양 전체로 가른다(`fullmatch` — 앞에 무엇이 붙으면 프레이밍이 밀린 것이다).
            continue
        match = _OBJECT_HEADER.fullmatch(header)
        if match is None:
            # 모르는 머리를 "없음" 으로 치면 결함이 부재로 둔갑한다 (Codex P2). 다른 spec 의
            # `missing` 도 여기 온다 — 프레이밍이 밀렸다는 뜻이다.
            raise RuntimeError(
                f"`git cat-file --batch -Z` gave an unrecognised answer to request "
                f"{index + 1} of {len(wanted)} ({relative}) at {ref[:12]}: {header[:80]!r}"
            )
        size = int(match.group("size"))
        if position + size >= len(data):
            # 내용과 그 뒤의 NUL 이 들어갈 자리가 없다 — 잘렸다. 옛 판본은 여기서 넘쳐
            # `left -6 byte(s) unread` 라고 **거꾸로** 말했다 (gstack 리뷰 F5). 경계(`==`)가
            # "내용은 다 왔고 NUL 만 없다" 이고, `>` 로 좁히면 아래 색인이 `IndexError` 로 샌다.
            raise RuntimeError(_TRUNCATED.format(count=index, total=len(wanted), ref=ref[:12]))
        if data[position + size] != 0:
            # 총량이 맞는다고 프레이밍이 맞는 것은 아니다 — `abcX` 가 `abc` 로 받아들여졌다 (Codex P2).
            raise RuntimeError(
                f"`git cat-file --batch -Z` payload {index + 1} of {len(wanted)} ({relative}) at "
                f"{ref[:12]} is not NUL-terminated: the response framing was not understood"
            )
        if match.group("type") == b"blob":
            answered[relative] = data[position:position + size]
        position += size + 1                    # 내용 뒤의 NUL 하나
    if position != len(data):
        # 정상 응답은 **정확히** 소진된다(실측). 남은 바이트는 프레이밍을 잘못 읽었다는 뜻이다.
        raise RuntimeError(
            f"`git cat-file --batch -Z` left {len(data) - position} byte(s) unread at "
            f"{ref[:12]}: the response framing was not understood"
        )
    found.update(answered)
    return found


def _committed_bytes(root: Path, ref: str, relative: str) -> bytes | None:
    """`ref` 시점의 파일 내용. 워킹트리를 보지 않는다.

    "그 커밋의 blob 을 읽는다"는 철자는 **한 곳**에 산다 (task 7.5). 두 벌이면 한쪽만
    고쳐도 양쪽 시험이 초록이다([[two-judgements-cover-for-each-other]]). 단건도 배치가
    더 싸다 — 실측 `git show` 4.05~4.80 ms 대 `cat-file --batch -Z` 2.64~3.39 ms.
    """
    return _committed_many(root, ref, [relative])[relative]


def _first_line(said: str | bytes | None, fallback: str) -> str:
    """git 이 한 말의 첫 줄. 결함 문장이 **git 의 말**로 이름을 대게 한다 (7.5.1 F6)."""
    text = said.decode("utf-8", "replace") if isinstance(said, bytes) else (said or "")
    lines = text.strip().splitlines()
    return lines[0][:160] if lines else fallback


def _is_ancestor(root: Path, older: str, newer: str) -> bool:
    """`older` 가 `newer` 의 조상인가. git 이 "예"(0) · "아니오"(1) 말고 답하면 **결함**이다.

    rc 128(객체를 못 읽음 · 저장소가 아님)을 "조상이 아니다" 로 읽던 판본은 결함을 착지 거절 사유로 만들고
    "기록을 지우고 다시 기록하라" 는 복구 조언까지 붙였다 (task 7.5.2.1, 재리뷰 레드팀 repro_a3) — 7.5.1 이
    `_committed_many` 에서 고친 것과 같은 섞임이다([[a-fault-must-become-a-verdict]]).
    """
    process = subprocess.run(
        ["git", "merge-base", "--is-ancestor", older, newer],
        cwd=root, capture_output=True, text=True, timeout=10, check=False,
    )
    if process.returncode in (0, 1):
        return process.returncode == 0
    raise RuntimeError(
        f"cannot tell whether {older[:12]} precedes {newer[:12]}: "
        + _first_line(process.stderr, "git merge-base failed")
    )


def _target_text(landing: str, audited: bool = False) -> str:
    """비교 대상 쪽 끝을 사람이 읽는 말로. **무엇이** 그 끝을 고정했는지까지 말한다.

    이관 예외의 끝은 저자가 선언한 값이 아니라 `execution-baseline.json` 이 감사한
    `source_commit` 이다. 둘을 같은 말로 적으면 있지도 않은 파일을 가리키게 된다.
    """
    if not landing:
        # "in HEAD" 인 이유 (task 7.7, 리뷰 I8): 게이트는 기록을 **커밋에서** 읽는다. 옛 문장
        # "(no landed-commit.txt)" 는 기록이 디스크에만 있을 때 거짓이었고, 바로 옆 줄이 권한
        # `--record-landing` 은 그 파일이 "이미 있다"고 거절했다.
        return f"working tree (no {LANDING_FILE} in HEAD)"
    return f"audited source-commit {landing}" if audited else f"landed-commit {landing}"


def _commits_after(root: Path, base: str, head: str) -> str:
    """base 뒤에 착지한 커밋 수. 숫자를 **재서** 쓴다 — 못 재면 빈 문자열이다.

    끝은 판정이 푼 `head` 다 (task 7.5.2.1) — 창 줄의 숫자와 판정이 같은 역사를 말해야 한다.
    """
    process = subprocess.run(
        ["git", "rev-list", "--count", f"{base}..{head}"],
        cwd=root, capture_output=True, text=True, timeout=10, check=False,
    )
    return "" if process.returncode else process.stdout.strip()


def _landing_record(change_dir: Path, root: Path, head: str) -> tuple[str, bytes] | None:
    """`head` 커밋에 착지 기록이 **있는가**. 있으면 `(경로, 바이트)`, 없으면 `None`.

    해독하지 않는다. "기록이 있는가"와 "기록에 무엇이 적혔나"는 다른 질문이고, 앞의
    것에 해독하는 함수를 부르면 못 읽는 기록 앞에서 질문 자체가 터진다 — 이관 경로의
    probe 와 `record_landing` 의 덮어쓰기 거절이 그랬다(task 6.2.1). 못 읽는 기록도
    기록이다.

    `head` 는 명령이 **한 번** 푼 sha 다 (task 7.5.2.1). 상징 `HEAD` 를 여기서 따로 읽던 판본은 기록을
    한 역사에서, 그 기록의 판정을 다른 역사에서 했다(재리뷰 레드팀 repro_b — 가지 전환 한 번에 rc 0).
    """
    try:
        relative = (change_dir / LANDING_FILE).relative_to(root).as_posix()
    except ValueError:
        return None
    raw = _committed_bytes(root, head, relative)
    return None if raw is None else (relative, raw)


def _declared_landing(change_dir: Path, root: Path, head: str) -> str | None:
    """`head` 커밋에 적힌 착지 선언. 선언 자체가 없으면 `None` 이다.

    값을 **판정하지 않는다** — 40자리인지, 커밋인지, 조상인지는 `resolve_landing`
    이 묻는다. 여기가 답하는 것은 "선언이 있는가, 있다면 무엇이라고 적혀 있는가"
    하나뿐이다. 빈 파일은 `""` 이고 `None` 과 **다르다**: 빈 선언도 선언이므로
    없는 것으로 읽으면 그 change 가 조용히 워킹트리를 대상으로 삼게 된다.
    """
    record = _landing_record(change_dir, root, head)
    if record is None:
        return None
    relative, raw = record
    try:
        return raw.decode("utf-8").strip()
    except UnicodeDecodeError as exc:
        raise ValueError(f"landing point is not UTF-8: {relative}") from exc


class Evidence(NamedTuple):
    """증거 디렉터리를 **한 번** 읽은 것 (task 7.5.2 · 7.5.2.1). 판정은 이것만 본다.

    셋 다 입력이다 — 디렉터리가 **있었나**, 어떤 번들이 **있었나**, 각 `ast.json` 의 **바이트**. 7.5.2 는
    바이트만 한 번 읽고 나머지 둘은 판정 도중 다시 물었다: 착지를 판정한 뒤 디렉터리를 지우면 조기 반환이
    "증거 없음" 으로 읽었고(재리뷰 적대), 번들 안 파일 목록을 다시 만들면 한 번 읽은 `ast.json` 이 빠졌다
    (Codex P1). 이 값 **하나**가 착지 판정 · 대상 판정 · base 모양 조언에 같이 들어간다(`--record-landing` 을
    권하는 조언 줄은 **다음에 부를 명령**을 예측하므로 스스로 다시 읽는다 — 판정 줄은 그 읽기에 안 기댄다).
    번들 **산문**은 여기 없다 — 착지가 판정하지 않는 범위라 대상 판정이 읽을 때 읽는다. 같은지 볼 때는 바이트를 그대로
    비교한다 — 해시를 따로 둘 이유가 없다.

    번들 **안의 이름 목록**도 여기 있다 (task 7.5.2.3). 7.5.2.2 는 그 목록을 세 곳에서 따로 만들었다: 빌림
    공존 판정 · 감사 참여 판정(`lexists`) · `_bundle_text`. 목록이 셋이면 같은 디렉터리에 대해 서로 다른
    대답이 판정에 들어갈 수 있고(중간에 파일이 생기면 실제로 갈린다), 세 자리 모두 syscall 을 따로 낸다.
    """

    directory: Path
    present: bool                               # 증거 디렉터리가 디렉터리로 있었나
    targets: tuple[Path, ...]                   # 번들 디렉터리 (`iterdir` 의 디렉터리)
    held: dict[Path, bytes | None]              # `target/ast.json` → 바이트 · 못 읽음 `None` · 없으면 키가 없다
    listings: dict[Path, tuple[str, ...]]       # 번들 → 그 안의 이름들 (목록을 못 열면 키가 없다)
    unlistable: dict[Path, str]                 # 목록을 못 연 번들 → 그 사유 (예외 객체를 담으면 값 비교가 깨진다)


class NotRegularFile(OSError):
    """정규 파일이 아닌 것을 열었다 — FIFO · 장치 · 소켓. 판정 줄이 **이름을 댄다** (task 7.5.2.3).

    7.5.2.2 는 이 모양을 `None` 으로 돌려 **조용히 건너뛰었다**. 보안 리뷰가 그 문으로 들어왔다: 커밋된 표
    파일을 게이트가 **여는 그 순간에만** FIFO 로 바꿨다 되돌리면 열거형 호출 감사가 꺼지고 `[]` 가 찍힌다
    (타이밍으로 맞출 수 있다 — 그 판의 하네스는 안 남겼으므로 여기 횟수를 적지 않는다, 7.5.2.4 정정).
    디스크에는 표가 든 정규 파일이 그대로다. 목록과 열기는 다른 syscall 이라 "at rest 에 이런
    파일이 없다"(전수 0/12,193)는 **열기 순간**에 대한 진술이 아니다 ([[a-silent-skip-is-a-door]]).
    """


class FileTooLarge(OSError):
    """정규 파일인데 상한보다 크다 (task 7.5.2.3).

    종류만 막고 크기는 안 막아서, 큰 정규 파일 하나가 `MemoryError`(GATE_FAULTS 밖)로 판정 줄을 0 으로
    만들었다(재리뷰 보안 · 적대). 상한의 근거: 저장소에서 가장 큰 `*.go` 97 KB · 가장 큰 번들 파일 26 KB ·
    16 MiB 넘는 번들 파일 **0 / 12,193**.
    """


class NotUtf8Text(OSError):
    """읽었는데 UTF-8 이 아니다 (task 7.5.2.3). 조용히 건너뛰면 그 파일의 표가 판정에서 빠진다.

    `OSError` 로 만드는 까닭은 하나다: 번들 파일의 실패를 **이름 대는 자리가 하나**여야 한다. `_verdict` 가
    `except OSError` 한 곳에서 "못 읽은 것"을 전부 이름 대므로, 새 실패 모양이 어느 목록에도 안 걸려
    traceback 이 되는 길이 없다 ([[a-fault-must-become-a-verdict]]).
    """


class ReadLedger:
    """판정이 디스크에서 읽은 것의 **원장**. 끝의 재확인이 이 목록을 그대로 다시 읽는다 (task 7.5.2.3).

    7.5.2.1 은 손으로 적은 키 목록(`RUN_FACTS`)이 낡아서 깨졌고, 7.5.2.2 는 손으로 고른 재확인 집합
    (`HEAD` + `Evidence`)이 판정 집합보다 좁아서 깨졌다 — **같은 실패가 두 번**이다(a112 실측: 판정이 읽은
    1,739 경로 중 재확인이 본 것 149). 그래서 목록을 만들지 않는다: 디스크를 읽는 깔때기가 읽은 것의
    **결과**를 여기 적고, `_reads_moved` 가 그것을 다시 읽어 견준다. 새 읽기 자리는 깔때기를 쓰는 순간
    저절로 참여하고, 깔때기를 비켜 가는 것은 AST 구조 시험이 막는다.

    **실패도 적는다**(없음 · 정규 파일 아님 · 권한 · 너무 큼). 그래야 "여는 순간에만 FIFO" · "이름을 빼고
    되돌리기" 가 끝에서 **다르게** 보인다 — 실패를 안 적으면 견줄 것이 없다.
    """

    def __init__(self) -> None:
        self.seen: dict[tuple[str, str], str] = {}
        self.recording = True
        self.diverged = ""                      # 한 판에서 같은 경로가 갈렸으면 그 경로


_LEDGER: ReadLedger | None = None


@contextlib.contextmanager
def _ledger():
    """이 명령이 읽는 것을 적는 원장을 연다. 열려 있지 않으면 깔때기는 아무것도 적지 않는다(시험이 직접 부를 때)."""
    global _LEDGER
    previous, book = _LEDGER, ReadLedger()
    _LEDGER = book
    try:
        yield book
    finally:
        _LEDGER = previous


def _remember(kind: str, key: str, outcome: str) -> None:
    """읽기 결과를 원장에 적는다. 재확인 중에는 적지 않는다 — 그러면 자기 읽기를 자기가 견주게 된다."""
    book = _LEDGER
    if book is None or not book.recording:
        return
    previous = book.seen.get((kind, key))
    if previous is not None and previous != outcome and not book.diverged:
        # 한 판에서 같은 경로를 두 번 읽었고 결과가 갈렸다 — 끝까지 기다릴 것 없이 그 자리가 움직임이다.
        book.diverged = key
    book.seen[(kind, key)] = outcome


def _failed(exc: OSError) -> str:
    """실패의 지문. 종류와 errno 로 — 권한이 풀린 것도 파일이 생긴 것도 "달라졌다" 다."""
    return f"{type(exc).__name__}:{exc.errno}"


def _opened_bytes(path: Path) -> bytes:
    """파일을 **여는 유일한 자리**. 정규 파일의 바이트, 그 밖은 이름을 담은 `OSError`.

    `O_NONBLOCK` 으로 열고 **연 것의** 종류를 `fstat` 으로 본다 — 확인과 읽기 사이에 파일이 바뀔 틈이 없다
    (쓰는 쪽이 없는 FIFO 도 그렇게 바로 열리고 곧 닫힌다). 심링크는 따라가고, 따라간 끝의 종류를 본다.

    **디렉터리도 정규 파일이 아닌 것도 건너뛰지 않는다** (task 7.5.2.2 · 7.5.2.3). 건너뛰면 번들 안에 폴더를
    만들거나 표 파일을 FIFO 로 바꿔 두는 것으로 감사가 꺼진다. 종류 검사는 파일 객체로 감싸기 **전에** 한다:
    먼저 감싸면 `open(fd)` 가 디렉터리에서 스스로 터지는데 그 예외의 `filename` 은 경로가 아니라 **정수 fd**
    라서 이름을 대려던 `_verdict` 가 `TypeError`(GATE_FAULTS 밖)로 죽었다(7.5.2.2 가 만든 회귀).

    상한보다 한 바이트 더 읽어 넘치는지 본다 — `st_size` 를 믿지 않는다(희소 파일 · `/proc`).
    어느 갈래로 나가도 서술자를 닫는다(`closefd=False` 라 `with` 는 닫지 않는다).
    """
    descriptor = os.open(path, os.O_RDONLY | os.O_NONBLOCK)
    try:
        mode = os.fstat(descriptor).st_mode
        if stat.S_ISDIR(mode):
            raise IsADirectoryError(errno.EISDIR, os.strerror(errno.EISDIR), str(path))
        if not stat.S_ISREG(mode):
            raise NotRegularFile(errno.EINVAL, NOT_REGULAR, str(path))
        with open(descriptor, "rb", closefd=False) as handle:
            raw = handle.read(READ_CAP + 1)
    finally:
        os.close(descriptor)
    if len(raw) > READ_CAP:
        raise FileTooLarge(errno.EFBIG, TOO_LARGE, str(path))
    return raw


def _file_outcome(path: Path) -> tuple[str, bytes | OSError]:
    """`(지문, 바이트 또는 실패)`. 지문을 만드는 **한 자리** — 판정과 재확인이 같은 함수를 부른다.

    둘이 각자 지문을 만들면 한쪽만 고쳐도 양쪽 시험이 초록이 된다
    ([[two-judgements-cover-for-each-other]] 가 a064 에서 실측한 모양).
    """
    try:
        raw = _opened_bytes(path)
    except OSError as exc:
        return _failed(exc), exc
    return "bytes:" + hashlib.sha256(raw).hexdigest(), raw


def _read_regular(path: Path) -> bytes:
    """정규 파일의 바이트. 없거나 · 종류가 틀리거나 · 못 읽거나 · 너무 크면 이름을 담은 `OSError`.

    판정이 **파일을 읽는 유일한 자리**다 (task 7.5.2.3). 결과를 원장에 적으므로, 이 함수로 읽은 것은
    끝의 재확인이 반드시 다시 읽는다.
    """
    outcome, value = _file_outcome(path)
    _remember("file", str(path), outcome)
    if isinstance(value, OSError):
        raise value
    return value


def _listing_outcome(path: Path) -> tuple[str, list[tuple[str, bool]] | OSError]:
    """`(지문, [(이름, 디렉터리인가)] 또는 실패)`. 종류까지 지문에 넣는다 — 같은 이름의 파일↔폴더 교체도 변화다."""
    try:
        entries = sorted((child.name, child.is_dir()) for child in path.iterdir())
    except OSError as exc:
        return _failed(exc), exc
    joined = "\n".join(f"{name}\t{'d' if is_dir else 'f'}" for name, is_dir in entries)
    return "list:" + hashlib.sha256(joined.encode("utf-8")).hexdigest(), entries


def _listed(path: Path) -> list[tuple[str, bool]]:
    """디렉터리를 **여는 유일한 자리**. 이름과 종류를 정렬해 돌려주고 결과를 원장에 적는다."""
    outcome, value = _listing_outcome(path)
    _remember("dir", str(path), outcome)
    if isinstance(value, OSError):
        raise value
    return value


def _pattern_outcome(root: Path, pattern: str) -> tuple[str, list[Path]]:
    """`(지문, 맞은 경로들)`. 트리 순회는 **집합**이 판정의 입력이다 — 파일 하나가 생겨도 달라진다."""
    matches = sorted(root.rglob(pattern))
    joined = "\n".join(str(path) for path in matches)
    return "glob:" + hashlib.sha256(joined.encode("utf-8")).hexdigest(), matches


def _globbed(root: Path, pattern: str) -> list[Path]:
    """트리를 **순회하는 유일한 자리**(시험 함수 색인 · 인용 해소). 결과를 원장에 적는다."""
    outcome, matches = _pattern_outcome(root, pattern)
    _remember("glob", f"{root}\n{pattern}", outcome)
    return matches


def _kind_outcome(path: Path) -> tuple[str, str]:
    """`(지문, 종류)` — `dir` · `reg` · `other` · 못 물으면 빈 글자."""
    try:
        mode = os.stat(path).st_mode
    except OSError as exc:
        return _failed(exc), ""
    kind = "dir" if stat.S_ISDIR(mode) else "reg" if stat.S_ISREG(mode) else "other"
    return "kind:" + kind, kind


def _kind(path: Path) -> str:
    """이름만 재는 **유일한 자리**. 읽지 않고 고르는 판정(디렉터리인가 · 정규 파일인가)도 원장에 남는다."""
    outcome, kind = _kind_outcome(path)
    _remember("kind", str(path), outcome)
    return kind


def _worktree_digest(path: Path) -> str | None:
    """워킹트리 파일의 sha256 — 없거나 못 읽으면 `None`. 깔때기로 읽으므로 원장에 남는다."""
    try:
        return hashlib.sha256(_read_regular(path)).hexdigest()
    except OSError:
        return None


def _shown(root: Path, path: str) -> str:
    """판정 줄에 적을 이름 — 저장소 안이면 상대경로, 밖이면 그대로."""
    try:
        return Path(path).relative_to(root).as_posix()
    except ValueError:
        return path


def _why(exc: BaseException) -> str:
    """실패를 사람의 말로. `strerror` 가 있으면 그것 — 깔때기가 담은 사유가 그 자리에 있다."""
    if isinstance(exc, UnicodeDecodeError):
        # 해독 실패의 기본 문장은 바이트 위치까지 담아 길다 — 판정 줄에는 무엇이 틀렸는지만 적는다.
        return NOT_UTF8
    return str(getattr(exc, "strerror", "") or exc)


_PROBE_NOW = {
    "file": lambda key: _file_outcome(Path(key))[0],
    "dir": lambda key: _listing_outcome(Path(key))[0],
    "glob": lambda key: _pattern_outcome(Path(key.split("\n")[0]), key.split("\n")[1])[0],
    "kind": lambda key: _kind_outcome(Path(key))[0],
}


def _reads_moved(root: Path, book: ReadLedger) -> str:
    """판정 중 읽은 것을 **전부** 다시 읽어 지문을 견준다. 처음 다른 것의 이름, 같으면 빈 글자.

    이 읽기의 바이트는 판정에 **들어가지 않는다** — 견주기만 한다. 재확인 중에는 원장에 적지 않는다.
    """
    if book.diverged:
        return f"{_shown(root, book.diverged)} changed"
    book.recording = False
    try:
        for (kind, key), before in sorted(book.seen.items()):
            if _PROBE_NOW[kind](key) != before:
                return f"{key.split(chr(10))[1] if kind == 'glob' else _shown(root, key)} changed"
    finally:
        book.recording = True
    return ""


def _read_evidence(analysis: Path) -> Evidence:
    """증거 디렉터리를 **한 번** 읽는다 — `ast.json` 을 읽는 **유일한** 자리다 (task 7.5.2).

    7.5.1 의 지문과 판정 목록은 같은 디렉터리를 **따로** 읽었다 — 둘째 읽기에서만 번들 B 가 한 번 실패하면
    B 가 판정에서 빠졌는데 두 지문은 같다고 답했다(Codex 재현). 한 번 읽은 값을 모두가 쓰면 갈릴 자리가 없다.

    번들은 `iterdir` 로 세고 `ast.json` 은 `target/ast.json` 을 **직접** 연다 (task 7.5.2.1, Codex P2).
    `glob("*/ast.json")` 은 목록(r)이 안 되는 번들 디렉터리를 조용히 건너뛰어, 검색(x)만으로 멀쩡히 읽히는
    `ast.json` 을 "없다" 로 만들었다. 없는 것(`FileNotFoundError` — 끊긴 심링크 포함)은 키가 없고, 있는데
    못 읽는 것(권한)과 정규 파일이 아닌 것(디렉터리 · FIFO · 장치 — 7.5.2.2, 열면 멎거나 끝없이 읽는다)은
    `None` 이다 — 판정 줄이 다르다(`missing` / `could not be read`). 못 읽은 것도 이 값에 남으므로 다음 읽기에서
    읽히면 재확인이 본다.

    번들 **안의 이름 목록**도 여기서 한 번 읽는다 (task 7.5.2.3). 목록을 못 여는 번들은 이름과 사유를 들고
    가서 대상마다 **이름 댄 판정 줄**이 된다 — 조용히 빈 목록으로 두면 그 번들의 표가 감사에서 빠진다.
    디렉터리가 **아닌 것**(없음 · 파일 · 끊긴 링크)만 "증거 없음" 이고, 있는데 **못 여는 것**은 결함으로
    올린다 — 7.5.2.2 는 `is_dir()` 이 `OSError` 를 삼켜서 읽을 수 없는 증거 디렉터리를 면제로 통과시켰다.
    """
    try:
        entries = _listed(analysis)
    except (FileNotFoundError, NotADirectoryError):
        return Evidence(analysis, False, (), {}, {}, {})
    targets = tuple(analysis / name for name, is_dir in entries if is_dir)
    held: dict[Path, bytes | None] = {}
    listings: dict[Path, tuple[str, ...]] = {}
    unlistable: dict[Path, str] = {}
    for target in targets:
        ast_path = target / "ast.json"
        try:
            held[ast_path] = _read_regular(ast_path)
        except FileNotFoundError:
            pass
        except OSError:
            held[ast_path] = None
        try:
            listings[target] = tuple(name for name, _ in _listed(target))
        except OSError as exc:
            unlistable[target] = _why(exc)
    return Evidence(analysis, True, targets, held, listings, unlistable)


def _select_pinning(root: Path, evidence: Evidence) -> list[tuple[Path, str, str]]:
    """착지를 고정하는 번들 = `revision: current` 이고 `file`·`source_sha256` 둘 다 있는 것.

    이 선별은 **한 곳에만** 산다. `resolve_landing` 과 `compute_landing` 이 각자
    순회를 가지면 한쪽만 고쳐도 양쪽 시험이 초록이 된다
    ([[two-judgements-cover-for-each-other]] 가 a064 에서 실측한 모양). 한 번 읽은 증거를
    받는다 — 착지 판정과 대상 판정이 같은 읽기에서 나오게 (task 7.5.2 · 7.5.2.1).
    """
    found: list[tuple[Path, str, str]] = []
    for ast_path, raw in evidence.held.items():
        value = _parsed(raw)
        if value.get("revision", "current") != "current":
            continue
        raw_source = str(value.get("file", ""))
        digest = str(value.get("source_sha256", ""))
        if not raw_source or not digest:
            continue
        # 번들의 `file` 은 절대경로일 수 있다 — `validate_target` 이 정규화하는 이유가
        # 그것이다. 여기서 안 하면 `git show <sha>:/abs/path` 가 언제나 실패해서
        # **정상 입력이 위조로 몰린다**. 저장소 번들 3048개는 오늘 전부 상대경로라
        # 실물 영향은 0 이고, a063 이관 픽스처가 절대경로를 만들면서 드러났다.
        _, source = normalized_source(raw_source, root)
        found.append((ast_path, source, digest))
    return found


def _base_shaped_bundles(root: Path, base: str, evidence: Evidence) -> list[str]:
    """`revision: current` 인데 **base 의 소스**를 적은 고정 번들의 이름들.

    이 모양이 §6 독립 리뷰의 V1 이다. 저장소 규칙대로 FLM 을 **먼저** 커밋하고 코드를
    편집한 뒤 번들을 갱신하지 않으면, 번들은 base 의 소스를 적은 채로 남는다.

    이것을 세는 이유는 **조언 한 줄** 때문이다. 그런 번들이 있으면 도구가 받아들일 수
    있는 착지는 전부 그 파일이 아직 base 와 같은 지점이다. 그러므로 거기서 좁힌 창은
    그 함수를 요구할 수 없고, `--record-landing` 을 권하면 게이트가 자기 조언 줄로
    자기 판정을 지우는 길을 가리키게 된다.

    판정은 신원이 아니라 **blob 등식**으로 한다 — 누구의 편집인지는 묻지 않는다
    (1.12 가 신원 판정을 이미 배제했다).

    base 는 **한 프로세스**로 읽는다 (task 7.5.2, 재리뷰 적대 F6) — 7.5 가 남긴 per-bundle 루프였다.
    증거는 판정이 한 번 읽은 것이다 (task 7.5.2.1) — 조언이 따로 읽으면 판정과 다른 번들을 말할 수 있다.
    """
    stale = [
        (ast_path, source, digest)
        for ast_path, source, digest in _select_pinning(root, evidence)
        # 오늘의 소스를 적은 번들은 세탁할 것이 없다. 워킹트리 소스도 깔때기로 읽는다 (task 7.5.2.3) —
        # `is_file()` + `read_bytes()` 두 syscall 이 하나가 되고, 그 읽기가 원장에 남아 재확인이 본다.
        if _worktree_digest(root / source) != digest
    ]
    at_base = _committed_many(root, base, [source for _, source, _ in stale])
    return sorted(
        ast_path.parent.name for ast_path, source, digest in stale
        if at_base[source] is not None and hashlib.sha256(at_base[source]).hexdigest() == digest
    )


def _pinning_at(
    root: Path, candidate: str, bundles: list[tuple[Path, str, str]],
) -> tuple[int, list[str]]:
    """`candidate` 에서 고정 번들이 몇 개이고 그중 어느 소스가 안 맞는가.

    번들 목록은 **호출자가 한 번 재서 넘긴다** (task 7.5) — `floor` · `repairs` 와 같다.
    디렉터리를 받아 스스로 순회하던 판본은 후보마다 glob + JSON 파싱 + `Path.resolve()`
    를 다시 했다(2026-09-18 프로파일: a071 의 walk 하나에서 `_pinning_bundles` 가 341회 ·
    13.07s 중 9.30s). 도구는 번들을 쓰지 않지만 **다른 쓰는 이**(병행 세션 · 편집기)는 걷는
    동안에도 쓴다 — 그래서 `compute_landing` 이 수락 직전에 증거를 다시 읽어 대조한다(task 7.5.1 ·
    7.5.2 · 7.5.2.1 — 그 읽기는 대조만 하고 판정에 안 들어간다).
    """
    mismatched: list[str] = []
    # 번들마다 프로세스를 띄우지 않는다 (task 7.5) — 이 함수가 실패하는 walk 비용의
    # 거의 전부였다. 같은 소스를 적은 번들 여럿은 `_committed_many` 가 한 번만 묻는다.
    blobs = _committed_many(root, candidate, [source for _, source, _ in bundles])
    for _, source, digest in bundles:
        blob = blobs[source]
        if blob is None or hashlib.sha256(blob).hexdigest() != digest:
            mismatched.append(source)
    return len(bundles), mismatched


def _pre_archive_path(relative: str) -> str:
    """아카이브된 경로의 **옮기기 전** 이름. 아카이브가 아니면 빈 문자열이다."""
    if not relative.startswith(ARCHIVE_PREFIX):
        return ""
    dated, _, tail = relative[len(ARCHIVE_PREFIX):].partition("/")
    change = _archived_change_id(dated)
    return f"openspec/changes/{change}/{tail}" if change and tail else ""


def _evidence_floor(root: Path, bundles: list[tuple[Path, str, str]], head: str) -> str:
    """고정 번들이 `head` 의 역사에 들어온 **마지막** 커밋. 걸었는데 없으면 빈 문자열이다.

    git 이 못 걸으면 **결함**이다 (task 7.5.2.1, 재리뷰 레드팀 repro_a3). 빈 문자열로 돌려주던 판본은
    "평범한 커밋이 이 증거를 들인 적이 없다" 는 거절을 만들고 복구 조언까지 붙였다 — 결함을 판정으로.

    저자가 고를 수 없는 유일한 하한이다 — 오늘 만든 번들을 과거 커밋에 넣을 수
    없기 때문이다. a122 6.1.1 이 활성 13건을 전수로 재서 13건 **전부** 고정 번들이
    자기 base 뒤에 커밋됐음을 확인했으므로, 이 하한은 13건 전부에서 실재한다.

    `--diff-filter=MAT` 인 이유: 아카이브는 번들을 통째로 옮긴다. rename 을 세면
    아카이브하는 순간 그 커밋이 바닥이 되어 **이미 유효했던 기록이 무효가 된다**.
    옮기기 전 경로를 같이 줘야 rename 이 rename 으로 보이고 건너뛰어진다 —
    a099(유일한 실물 기록)의 바닥이 `21a315d1` 로 나와 기록된 착지 `e6c4636a` 이하다.
    경로마다 `--follow` 를 도는 판본은 a112 에서 61초였고 이 형태는 0.02초다.

    건너뛰는 것은 **그대로 옮긴 것뿐**이다 (task 7.2, H1). `-M` 의 기본 유사도 50% 는
    내용을 고치면서 옮긴 것도 같은 rename 으로 보고 건너뛰었다 — 이동에 재작성을 숨기면
    하한이 안 움직였다. `-M100%` 은 바이트가 같을 때만 rename 으로 본다: 아카이브의 순수
    이동은 그대로 건너뛰고 재작성만 센다. `T`(정규 파일↔심링크)도 내용 교체인데 `MA` 가
    안 셌다. 2026-09-12 전수 실측 — 하한이 움직이는 change 1건, 착지를 잃는 change 0건.

    `--no-follow` 인 이유 (task 7.2, H2): `log.follow = true` 는 흔한 개인 설정이고,
    경로가 **하나**일 때 이 호출을 `--follow` 로 만든다. 그러면 rename 이 짝지어져
    필터에서 빠지고 하한이 이동 이전으로 내려간다 — 같은 저장소·같은 선언이 기계마다
    다른 판정을 받는다. 하한은 저장소의 함수여야 한다. `diff.renames` 는 따로 안 지운다:
    `-M100%` 이 명령줄에서 이미 탐지 방식을 정한다(그 절을 지워 봐도 죽는 시험이 없다 —
    2026-09-12 뮤테이션 M-C3). 둘 다 기본값이라 실측 기준선(설정 없는 기계)의 답은 같다.
    """
    paths: list[str] = []
    for ast_path, _, _ in bundles:
        try:
            relative = ast_path.relative_to(root).as_posix()
        except ValueError:
            continue
        paths.append(relative)
        before = _pre_archive_path(relative)
        if before:
            paths.append(before)
    if not paths:
        return ""
    process = subprocess.run(
        ["git", "log", "-M100%", "--no-follow", "--diff-filter=MAT", "-1", "--format=%H",
         head, "--", *paths],
        cwd=root, capture_output=True, text=True, timeout=60, check=False,
    )
    if process.returncode:
        raise RuntimeError(
            "cannot find the commit that put this change's evidence into the history: "
            + _first_line(process.stderr, "git log failed")
        )
    return process.stdout.strip()


def _unheld_bundles(
    root: Path, candidate: str, bundles: list[tuple[Path, str, str]],
    held: dict[Path, bytes | None],
) -> list[str]:
    """`candidate` 가 **들고 있지 않은** 번들. 판정이 읽은 바이트를 기준으로 센다.

    하한은 번들의 **경로**를 보고 판정은 **워킹트리의 내용**을 읽는다. 그 둘이 갈리는
    자리마다 저자는 하한을 안 움직이고 증거를 바꿀 수 있다 (task 7.2, 리뷰 C3) —
    자리표시자를 일찍 커밋해 두고 워킹트리에서 갈아 끼우거나, `ast.json` 을 감시 경로
    밖으로 향하는 심링크로 두거나, 아예 커밋하지 않거나.

    그래서 등식은 **워킹트리와** 세운다. HEAD 와 세우면 갈아 끼운 자리를 못 본다 —
    바뀐 쪽이 워킹트리이기 때문이다(2026-09-12 실측: K1_head 는 자리표시자 구멍에 눈이
    멀고 K1_worktree 는 안 멀다). 심링크는 blob 이 **가리키는 경로 문자열**이라 이
    등식에서 저절로 갈린다.

    범위가 `ast.json` 하나인 것도 실측이다. 번들의 산문 파일까지 넓히면 활성·아카이브
    76건 중 3건이 착지를 잃는데(a092·a043·a096), 셋 다 **이웃 change 가 나중에 단
    무효화 배너**라 정상 입력이다. `ast.json` 만으로는 거부 0 이다.

    `held` 는 명령이 **한 번** 읽은 바이트다 (task 7.5.2). 디스크를 다시 읽지 않는다 — 후보마다 다시
    읽으면 바뀌었다 돌아온 바이트(ABA)로 판정한 착지를 수락 직전 재확인이 못 본다. 필수 인자다 (task
    7.5.2.1): 시험만 쓰던 "안 주면 여기서 읽는다" 는 디스크를 읽는 둘째 길이었다(재리뷰 maintainability).
    """
    # 물어볼 경로를 **먼저 다 모으고** 한 프로세스로 읽는다 (task 7.5). 받는 후보는
    # `_pinning_at` 과 이 함수를 **둘 다** 통과하므로, 한쪽만 고치면 성공하는 walk 의
    # 비용은 절반만 준다(2026-09-18 실측: a112 에서 47.5% · 47.5%).
    watched: list[tuple[Path, str, str]] = []
    for ast_path, _, _ in bundles:
        try:
            relative = ast_path.relative_to(root).as_posix()
        except ValueError:
            continue
        watched.append((ast_path, relative, _pre_archive_path(relative)))
    blobs = _committed_many(
        root, candidate, [path for _, relative, before in watched
                          for path in ((relative, before) if before else (relative,))],
    )
    unheld: list[str] = []
    for ast_path, relative, before in watched:
        judged = held.get(ast_path)             # 판정과 **같은 읽기** — 심링크면 따라간 바이트다
        if judged is None:
            unheld.append(ast_path.parent.name)
            continue
        committed = blobs[relative]
        if committed is None and before:
            committed = blobs[before]
        if committed is None or committed != judged:
            unheld.append(ast_path.parent.name)
    return sorted(set(unheld))


def _self_repair_commits(root: Path, analysis: Path, head: str) -> list[str]:
    """이 change 자신의 **나중 Go 작업** 커밋들 — `head` 의 역사에서, 오래된 것부터. 없으면 빈 목록이다.

    세는 것은 하나다: 그 change 의 디렉터리를 만지면서 Go 파일도 고친 **비병합** 커밋
    (task 7.2.6 — 사람이 2026-09-16 에 고른 변형 B-ANYGO).

    **왜 내용이 아니라 이 신호인가.** 착지를 기록한 뒤 같은 고정 파일을 리뷰 수리로
    고치고 번들을 안 갱신하면, 5단계는 그 수리를 **창 밖**에 두고 초록이 된다(H4). 내용으로
    가르려던 후보 규칙(고정 소스가 착지와 지금 같아야 한다)은 착지 있는 68건 중 파일 단위
    61건 · 함수 단위 55건을 거절했고 원인은 거의 전부 **이웃**의 커밋이었다 — 자기 수리와
    남의 편집이 내용으로는 안 갈린다(2026-09-16 전수, review.md `## MEASURE — task 7.2.6`).
    갈리는 신호는 "같은 커밋이 이 change 의 디렉터리도 만졌는가" 하나뿐이었다.

    **신원을 거절에만 쓴다.** 1.12 가 막은 것은 신원으로 착지를 **받는** 판정이다
    (`base-commit.txt` 가 freeze 뒤 모든 커밋에서 참이라 아무것도 못 갈랐다). 이 집합은
    후보를 거절만 하므로 창을 좁히는 데 못 쓰이고, 틀리면 창이 **넓어지는** 쪽으로 틀린다.

    **한계 둘을 적는다** [[fail-closed-must-name-what-it-rejects]]:
    - Go 수리와 문서를 **다른 커밋**으로 쪼개면 이 집합에 안 들어온다. 망각을 막는
      가드이지 위조를 막는 가드가 아니다.
    - `--no-merges` 라 병합 커밋 자신의 변경은 안 읽는다 — 7.2.4 의 H7 과 같은 한계 부류.

    파일 목록은 커밋마다 따로 묻지 않는다. 경로 제한을 건 `git log` 는 **그 경로만** 적어
    주므로("이 커밋이 Go 도 고쳤나"를 못 본다), 디렉터리를 만진 커밋을 먼저 고른 뒤
    `--no-walk` 한 번으로 그 커밋들의 **전체** 목록을 읽는다. 측정 때 쓴 커밋별
    `git diff-tree` 판본과 착지 있는 68건 전수에서 깃발 224개 · 불일치 0 이다.
    """
    try:
        change_dir = analysis.parent.parent.relative_to(root).as_posix()
    except ValueError as exc:
        # 오늘 도달하지 않는다 — 호출자는 전부 `root` 에서 조립한 경로를 넘기고, 루트 밖
        # 번들만 있는 change 는 하한이 비어 이 함수 **앞에서** 거절된다. 그래도 빈 목록으로
        # 물러나지 않는다: 빈 목록은 "거절할 것이 없다"라서, 판정 못 한 것을 위반 0 으로
        # 읽게 된다([[missing-tool-reports-clean]]). 판정 못 하면 판정이 되어야 한다.
        raise RuntimeError(f"cannot name this change's directory under the repository: {exc}") from exc
    paths = [f"{change_dir}/"]
    if change_dir.startswith(ARCHIVE_PREFIX):
        # 아카이브 이동 **전**에 그 디렉터리를 만진 커밋이 대부분이다. 옮긴 뒤 이름만 보면
        # 활성 시절의 자기 수리가 전부 안 보인다.
        before = _archived_change_id(change_dir[len(ARCHIVE_PREFIX):].partition("/")[0])
        if before:
            paths.append(f"openspec/changes/{before}/")
    touching = subprocess.run(
        # `--full-history` 가 **핵심**이다 (2026-09-16 적대 리뷰 F1, 실측으로 재현). 경로 제한을
        # 건 `git log` 는 기본으로 역사를 단순화한다 — 병합이 **그 경로에 대해** 한 부모와
        # TREESAME 이면 반대편 가지를 통째로 버린다. 그러면 곁가지에서 한 자기 수리가 목록에서
        # 사라지고, 이 규칙이 닫으려는 H4 가 바로 그 모양(병합)에서 다시 열린다. 픽스처 셋이
        # 전부 초록이었다: 곁가지 수리 뒤 디렉터리 편집을 되돌린 경우 · 양쪽 가지가 같은 편집을
        # 한 경우 · 디렉터리 충돌을 main 것으로 해소한 병합. 오늘 이 저장소에서 `--full-history`
        # 가 더 세는 커밋은 활성 · 아카이브 전수에서 0 이다 — 지금 넣으면 공짜다.
        ["git", "log", "--full-history", "--no-merges", "--format=%H", head, "--", *paths],
        cwd=root, capture_output=True, text=True, timeout=120, check=False,
    )
    if touching.returncode:
        # **실패는 거절이 아니라 판정이다.** 빈 목록으로 물러나면 git 이 멎은 것과 "만진 커밋이
        # 없다"가 같은 말이 되고, 가드가 조용히 꺼진다(적대 리뷰 F2 — 주입 실험에서 rc 1 하나로
        # 거절돼야 할 입력이 초록이었다). 옆의 `_evidence_floor` 도 7.5.2.1 부터 실패하면 결함이다 — 두 자리가 같은
        # 방향이다(그전에는 거절로 가서 복구 조언까지 붙었다).
        raise RuntimeError(
            f"cannot list the commits that touched {change_dir}: {_first_line(touching.stderr, 'git log failed')}"
        )
    hashes = touching.stdout.split()
    if not hashes:
        return []
    listing = subprocess.run(
        # `core.quotePath` 의 기본값은 이름의 비ASCII 바이트를 따옴표로 감싸 인용한다 —
        # 그러면 그런 이름의 `.go` 가 `.go` 로 안 끝나서 안 보인다(안 보이면 **거절을 안 해서**
        # 창이 좁아지는 쪽으로 틀린다). 오늘 이 저장소의 추적 경로 15,974개 중 비ASCII 는 0 이라
        # 측정값은 그대로다(A/B 깃발 224 · 불일치 0). 지어낸 규칙이 아니라 읽기를 사실대로 만드는 것이다.
        # `diff.renames` 도 명령줄에서 못 박는다 (적대 리뷰 F4). 기본값이 켜져 있으면 `.go` 를
        # 비-`.go` 이름으로 옮긴 커밋의 옛 이름이 목록에서 사라져 **깃발이 안 선다** — 판정이
        # 사람의 git 설정의 함수가 된다. `_evidence_floor` 가 `-M100% --no-follow` 로 지킨
        # 원칙과 같다: 하한도 이 신호도 저장소의 함수여야 한다. 끄는 쪽이 더 세는 방향이다.
        ["git", "-c", "core.quotePath=false", "-c", "diff.renames=false",
         "log", "--no-walk", "--no-merges", "--stdin", "--format=%x00%H", "--name-only"],
        cwd=root, capture_output=True, text=True, timeout=120, check=False,
        input="\n".join(hashes) + "\n",
    )
    if listing.returncode:
        raise RuntimeError(
            "cannot read the files those commits changed: "
            + _first_line(listing.stderr, "git log --no-walk failed")
        )
    flagged: set[str] = set()
    for block in listing.stdout.split("\0"):
        lines = [line for line in block.splitlines() if line]
        if lines and any(name.endswith(".go") for name in lines[1:]):
            flagged.add(lines[0])
    # `git log` 는 새 것부터 준다. 거절 문장이 **가장 오래된** 것을 이름으로 대야 저자가
    # 고칠 첫 자리를 가리킨다 — 뒤의 것들은 그 뒤에 쌓인 작업이다.
    return [commit for commit in reversed(hashes) if commit in flagged]


def _repairs_after(root: Path, candidate: str, repairs: list[str] | None, head: str) -> list[str]:
    """`candidate` **뒤에** 서는 수리 커밋들. 순서는 `repairs` 의 순서(오래된 것부터).

    끝은 명령이 푼 `head` 다 (task 7.5.2.1). 상징 `HEAD` 를 여기서 다시 읽던 판본은 수리 신호를 잰 역사와
    다른 역사에서 "뒤에 있나" 를 물었다 — 그 순간만 `HEAD` 가 물러났다 돌아오면 rc 0 이었다(이 세션이 재현).

    `candidate..head` 한 번으로 묻는다 — 후보마다 `merge-base` 를 수리 개수만큼 돌면
    a112 처럼 수리가 스물여섯인 change 에서 후보 하나에 프로세스가 스물여섯이다. 집합은
    같다: `rev-list A..head` 가 곧 "head 에서 닿고 A 의 조상이 아닌" 커밋이고, 후보 자신은
    자기 조상이므로 빠진다(그래서 수리 커밋 **자신**은 착지가 될 수 있다 — 복구 경로).

    `None` 은 "잰 적 없다" 다 — 하한이 못 서면 `_measure_landing_inputs` 가 재지 않는다. 그것을
    `[]`("수리 커밋이 없다")와 섞으면 `_self_repair_commits` 가 빈 목록을 거절한 이유가 여기서
    되살아난다 (task 7.5.2, 재리뷰). 오늘은 하한 가드가 이 앞에서 거절해서 안 닿는다.
    """
    if repairs is None:
        raise RuntimeError(
            "this change's own later Go work was never measured — a landing cannot be judged without it"
        )
    if not repairs:
        return []
    process = subprocess.run(
        ["git", "rev-list", f"{candidate}..{head}"],
        cwd=root, capture_output=True, text=True, timeout=120, check=False,
    )
    if process.returncode:
        # 실패는 판정이다 — 빈 목록이면 가드가 조용히 꺼진다 (적대 리뷰 F2).
        raise RuntimeError(
            f"cannot walk the history after {candidate[:12]}: "
            + _first_line(process.stderr, "git rev-list failed")
        )
    after = set(process.stdout.split())
    return [commit for commit in repairs if commit in after]


def _landing_refusal(
    root: Path, base: str, candidate: str, inputs: LandingInputs,
) -> tuple[str, list[str]]:
    """`candidate` 를 착지로 **받지 않는** 사유. 받으면 `("", [])`.

    착지의 수락 규칙은 **여기 한 곳**에 산다 (task 7.6, 리뷰 I2). 선언된 값을 판정하는
    `resolve_landing` 과 값을 계산하는 `compute_landing` 이 각자 한 벌씩 들고 있었다 —
    같은 여섯 조건을 다른 순서로. 두 벌이면 한쪽만 고쳐도 양쪽 시험이 초록이고, 갈리는
    순간 도구는 자기 게이트가 거절할 값을 쓴다([[two-judgements-cover-for-each-other]]).

    **순서가 거절 지점이다.** 선언 경로의 시험들은 가드마다 **그 가드의 문장**을 못
    박는다(6.3). 순서를 바꾸면 한 입력이 다른 가드에 먼저 걸려 그 문장이 못에서 빠진다
    ([[a-new-guard-unpins-the-guards-behind-it]]). 그래서 순서는 `resolve_landing` 이 쓰던
    것을 그대로 쓴다. 계산 경로는 이 순서 때문에 곁가지 후보에서 번들 대조를 더 한다 —
    2026-09-13 전수 실측으로 93건 합계 +328회(오늘 6,491회의 +5%), 그 거의 전부가 이미
    느린 아카이브 change 일곱의 몫이고 활성 change 에서는 둘뿐이다(7.5 가 그 walk 의 fetch
    단위를 고쳤다).

    입력은 호출자가 **한 벌**로 재서 넘긴다(`LandingInputs` — 고정한 `head` · 한 번 읽은 증거 · 고정
    목록 · 하한 · 수리 신호). 후보마다 아무것도 다시 재거나 읽지 않는다 — 불변 커밋과 이 한 벌만의
    함수라서, 같은 입력의 선언 경로와 계산 경로가 같은 답을 낸다 (task 7.5.2.1). 둘째 값은 판정이
    읽은 번들을 그 커밋이 **안 들고 있을 때만** 그 번들 이름이다. 계산 경로가 "왜 못 찾았나"를 그
    이름으로 말한다.
    """
    bundles, floor = inputs.bundles, inputs.floor
    # base 는 조상이면 된다. 같은 커밋도 이 판정은 통과하지만 맨 뒤의 조건(고정 소스 중
    # 하나 이상이 base 와 달라야 한다, task 7.2.2)이 반드시 거절한다. 예전 주석은 "같아도
    # 된다 — a074·a079·a075 의 정답"이라 적었는데, 그 모양은 증거가 base 의 소스를 적은
    # V1 과 가를 수 없어서 사람이 2026-09-14 에 둘 다 막았다.
    if not _is_ancestor(root, base, candidate):
        return f"landing point precedes the comparison base {base[:12]}: {candidate}", []
    pinning, mismatched = _pinning_at(root, candidate, bundles)
    if not pinning:
        # 위의 판정은 "이것이 어느 커밋인가"만 묻는다. 어느 커밋인지를 **고르지 못하게**
        # 하는 것은 번들 순회이고, 순회가 0회 돌면 저자가 구간의 바닥을 골라 요구 집합을
        # ∅ 로 만들 수 있다. 그 상태는 면제와 구분되지 않는다 — 고정할 증거가 없으면
        # 선언도 없다.
        return (
            f"landing point {candidate[:12]} is pinned by no `revision: current` "
            "evidence: a declared landing must be the revision some bundle describes"
        ), []
    if mismatched:
        return (
            f"landing point {candidate[:12]} is not the revision this evidence describes: "
            + ", ".join(sorted(set(mismatched)))
        ), []
    # 해시가 맞는다는 것만으로는 값이 안 정해진다. a122 6.1.1 이 활성 13건을 전수로
    # 재서 얻은 것: 고정을 통과하는 커밋이 13건 중 12건에서 2~59개이고, 그 수는 번들
    # 수와 무관하다(a112 번들 132 → 후보 39, a091 번들 2 → 후보 27). 저자가 그중
    # **가장 낮은 것**을 고르면 창이 비어 요구 집합이 ∅ 이 된다 — 8건이 오늘 그
    # 상태이고, 증거를 base 상태로 써 두면 누구나 그 상태를 만들 수 있다.
    # 하한 하나만 저자가 못 고른다: 자기 증거가 역사에 들어온 지점.
    if not floor:
        # 위 `_walk_floor` 와 같은 한계를 같은 말로 (task 7.2.4, 리뷰 H7).
        return (
            f"landing point {candidate[:12]} is pinned by evidence that no ordinary commit on "
            "this history adds: commit the `revision: current` bundles that pin it in an "
            "ordinary (non-merge) commit"
        ), []
    if not _is_ancestor(root, floor, candidate):
        return (
            f"landing point {candidate[:12]} precedes the evidence that pins it "
            f"({floor[:12]}): a declared landing cannot be older than the commit that "
            "put this change's evidence into the history"
        ), []
    # 하한 **뒤**에 선다. 앞에 두면 "증거가 역사에 없다"·"착지가 증거보다 앞선다"를
    # 재던 시험들의 거절 지점을 이 등식이 가로채서, 그 가드를 지워도 스위트가 초록으로
    # 남는다([[first-failure-is-not-the-fix-scope]] 가 같은 파일에서 실측한 모양).
    unheld = _unheld_bundles(root, candidate, bundles, inputs.evidence.held)
    if unheld:
        return (
            f"landing point {candidate[:12]} does not hold the evidence this verdict read: "
            + ", ".join(unheld)
            + " — the bundle judged here is not the bundle committed there"
        ), unheld
    # 착지는 고정 소스 중 **하나 이상**을 base 와 다르게 가진 커밋이어야 한다 (task 7.2.2,
    # 리뷰 C1 · C2 — 사람이 2026-09-14 에 고른 규칙). 증거가 base 의 소스를 적은 채로 남으면
    # 그 증거가 맞는 커밋은 전부 그 소스가 아직 base 와 같은 자리이고, 거기서 좁힌 창은 이
    # change 의 작업을 하나도 담지 못한다. FLM 을 먼저 커밋하고 번들을 안 갱신한 change(V1)도,
    # 작업 이전의 곁가지에 증거를 둔 병합(V2)도 그 모양이다.
    # 재기준화가 base 를 작업 **뒤로** 옮긴 change(a074 · a077 · a079 …)도 증거 내용으로는 같은
    # 모양이라 같이 거절된다. 가를 정보가 저장소에 없고, 사람이 그 대가를 알고 골랐다
    # (7.2 실측 8건, 편집 전 재측정은 review.md `## Pre-Edit Gate — task 7.2.2`).
    # 파일 단위로 본다 — 함수 단위 변화는 파일 단위 변화의 부분집합이라 여기서 받는 착지가
    # 함수 단위에서 더 많이 받지는 않는다. 고정 번들이 0 이면 위에서 이미 돌아갔으므로 `all` 이
    # 빈 표본 위에서 참이 되는 자리는 없다.
    # **맨 뒤**에 선다 — 앞의 가드들은 각자 자기 문장으로 못 박혀 있다
    # ([[a-new-guard-unpins-the-guards-behind-it]]).
    sources = sorted({source for _, source, _ in bundles})
    # 양쪽을 **한 프로세스씩** 읽는다 (task 7.5.1, gstack 리뷰 F3 — 소스마다 둘이었다).
    # 그리고 이제 한쪽이 **못 읽히면** 여기까지 오지 않는다: `_committed_many` 가 결함으로
    # 올린다. 예전에는 못 읽힌 쪽이 `None` 이라 `None != bytes` 가 "바뀌었다" 로 읽혀 이
    # 가드가 **편집 전 커밋을 통과시켰다**(F1, 재현됨).
    at_base = _committed_many(root, base, sources)
    at_candidate = _committed_many(root, candidate, sources)
    if all(at_base[source] == at_candidate[source] for source in sources):
        listed = ", ".join(sources[:3]) + (f", and {len(sources) - 3} more" if len(sources) > 3 else "")
        return (
            f"landing point {candidate[:12]} changes none of the sources its evidence pins "
            f"since the comparison base {base[:12]}: {listed} — evidence that still describes "
            "the base cannot tell where this change's Go work landed"
        ), []
    # 착지 **뒤에** 이 change 자신의 Go 작업이 더 서 있으면 그 후보는 착지가 아니다
    # (task 7.2.6 — 사람이 2026-09-16 에 고른 규칙). 기록이 한 번 쓰이고 나면 5단계는 그
    # 기록까지만 대조하므로, 기록 뒤의 리뷰 수리는 고정 파일을 고쳐도 창 밖에 남는다(H4).
    # 신호는 `_self_repair_commits` 가 한 곳에서 만든다(디렉터리 + Go, 비병합).
    # **맨 뒤**에 선다 — 앞의 일곱 가드는 각자 자기 문장으로 못 박혀 있고, 앞에 세우면 그
    # 못이 빠진다([[a-new-guard-unpins-the-guards-behind-it]]).
    # 자기 자신은 세지 않는다: `_is_ancestor` 는 같은 커밋에서 참이므로 수리 커밋 **자신**은
    # 착지가 될 수 있다. 그것이 복구 경로다(번들을 갱신하고 다시 기록하면 착지가 앞으로 온다).
    later = _repairs_after(root, candidate, inputs.repairs, inputs.head)
    if later:
        return (
            f"landing point {candidate[:12]} is followed by {len(later)} later commit(s) of "
            f"this change's own Go work (first {later[0][:12]}): a non-merge commit that edits "
            "Go while touching this change's directory landed after it, and the window "
            f"{base[:12]}..{candidate[:12]} would not compare that edit"
        ), []
    return "", []


def resolve_landing(
    change_dir: Path, root: Path, base: str, head: str, evidence: Evidence,
) -> str:
    """이 change 의 작업이 착지한 지점. 기록이 없으면 빈 문자열이다.

    유효성은 그 change 의 **증거**로 판정한다. 신원으로 판정하면 안 된다 —
    `base-commit.txt` 는 proposal freeze 에 쓰이고 spec 이 불변을 SHALL 로 요구하므로
    "착지 커밋에 그 파일이 같은 내용으로 있다"는 freeze 이후 **모든** 커밋에서 참이고
    아무것도 가르지 못한다. 그 판정으로는 a112 에 freeze 직후 커밋을 주어 base→착지를
    135→12 파일로 줄이고 비테스트 Go 48개를 숨길 수 있었다(2026-09-09 재현).

    기록은 워킹트리가 아니라 **HEAD 커밋**에서 읽는다. 워킹트리에서 읽으면 untracked
    파일로 게이트를 통과한 뒤 지울 수 있고, 그러면 어떤 대상으로 통과했는지가 아무
    데도 안 남는다.

    그 증거 판정에는 전제가 있다: **고정할 번들이 있어야 한다.** 없으면 순회가 0회
    돌고 저자가 구간의 바닥을 고를 수 있다. `evidence` 는 그래서 이 change 가 실제로
    딛는 증거 디렉터리를 읽은 것이어야 한다 — 빌린 증거를 쓰는 change 는 지역 번들이 0
    이므로 호출자가 **빌린 쪽을 먼저 풀어서** 읽는다.

    마지막 판정은 "유효한가"가 아니라 "**그 값인가**"다 (task 7.3). 유효 조건을 통과하는
    값은 여럿이므로(실측: 착지를 얻는 76건 중 65건) 그것만으로는 저자의 선택이 안 없어진다.
    기록은 `compute_landing` 이 내는 값과 같아야 한다.

    입력 둘은 호출자가 **한 번** 푼 것이다 (task 7.5.2.1): 역사의 끝 `head`(sha)와 증거 `evidence`.
    호출자(`check`)는 같은 `evidence` 로 대상을 판정한다 — 착지가 판정한 바이트와 판정이 읽는 바이트가
    **같은 값**이라 대조할 것이 없다. 7.5.2 는 여기서 지문을 넘기고 `check` 가 다시 읽어 대조했는데,
    대조 앞의 조기 반환과 대조 밖의 목록 열거가 디스크를 다시 물었다(재리뷰: 셋 다 재현).
    """
    candidate = _declared_landing(change_dir, root, head)
    if candidate is None:
        return ""
    if not FULL_SHA.fullmatch(candidate):
        # `rev-parse` 는 `HEAD`·브랜치·태그를 받는다. 이름은 리뷰 시점과 게이트 시점
        # 사이에 뜻이 바뀌고, `HEAD` 한 단어면 커밋 안 된 Go 편집이 통째로 요구에서
        # 빠진다.
        raise ValueError(
            f"landing point must be a full 40-hex commit id, not {candidate!r}"
        )
    process = subprocess.run(
        ["git", "rev-parse", "--verify", f"{candidate}^{{commit}}"],
        cwd=root, capture_output=True, text=True, timeout=10, check=False,
    )
    if process.returncode or process.stdout.strip() != candidate:
        raise ValueError(f"landing point is not a commit in this repository: {candidate}")
    if not _is_ancestor(root, candidate, head):
        raise ValueError(f"landing point never landed on this history: {candidate}")
    # 유효한가는 **한 함수**가 판정한다 — 아래 `compute_landing` 이 후보마다 묻는 것과
    # 같은 함수다 (task 7.6, 리뷰 I2).
    # 입력은 **한 번** 잰다 — 아래 `compute_landing` 도 이 한 벌을 쓴다 (task 7.5.1, F2).
    # 7.5 의 주석은 "한 번 잰다" 였는데 실제로는 여기서 한 번, `compute_landing` 이 또 한 번이었다.
    inputs = _measure_landing_inputs(root, head, evidence)
    refusal, _ = _landing_refusal(root, base, candidate, inputs)
    if refusal:
        # 거절이 **움직이던 입력**을 잰 것이면 복구 조언을 하지 않는다 (task 7.5.2, 재리뷰 적대 F5):
        # 다시 돌리면 통과할 기록을 지우라고 권하게 된다. 거절 자체는 한 번 읽은 증거와 불변 커밋만의
        # 함수다 — 그러니 여기서 달라졌다면 **읽을 때** 저자가 쓰던 중이었다.
        _raise_if_inputs_moved(root, inputs)
        # **복구 경로를 말한다** (task 7.2.6). 여기 오는 모든 거절은 "적힌 기록이 지금
        # 규칙으로는 착지가 아니다"이고, 돌아가는 길은 언제나 같다 — 증거를 갱신하고,
        # 기록을 지우는 커밋을 하고, 다시 기록한다. 예전 문장들은 무엇이 틀렸는지만 말해서,
        # 번들을 성실하게 갱신한 저자가 빨간 문장과 "이미 있다 — 덮어쓰지 않았다" 사이에
        # 갇혔다(2026-09-16 픽스처). 기록을 덮어쓰지 않는 규칙은 그대로다.
        raise ValueError(f"{refusal} — {LANDING_RECOVERY}")
    # **맨 뒤**에 선다 (task 7.3, 리뷰 H3). 위의 판정들은 전부 "이 값이 유효한가"를
    # 묻고, 이것 하나가 "이 값이 **그 값인가**"를 묻는다. 앞에 두면 위 가드들의 거절
    # 지점을 이 등식이 가로채서 그것들을 지워도 스위트가 초록으로 남는다
    # ([[a-new-guard-unpins-the-guards-behind-it]] — 7.2.1 이 같은 파일에서 실측했다).
    #
    # 넣는 이유는 유효 조건이 값을 **정하지 못하기** 때문이다. 2026-09-13 전수 실측:
    # 착지를 얻는 76건 중 65건이 유효한 값을 둘 이상 갖고(최대 516개), 44건에서는 그
    # 선택이 판정 입력을 바꾼다. 그 선택이 오늘 이 역사에서 창을 **넓히는** 방향뿐이라는
    # 것도 실측이지만(자손이 아닌 수락값 0/76 · 파일이 빠지는 자리 0/44), 되돌려진 편집이
    # 하나만 있어도 넓힌 창은 느슨해질 수 있다. 값을 만드는 주체를 도구로 옮긴 규칙은
    # 게이트가 그 값을 확인할 때만 규칙이다.
    computed, why = compute_landing(root, base, inputs)
    if candidate != computed:
        # 여기 오는 것은 거의 언제나 **진짜 불일치**다 — 같은 역사 · 같은 증거에서 더 낮은 수락 후보가 있다.
        # 선언값은 계산의 순회 안에 있고(하한 · base 의 자손이고 `head` 의 조상이다) 같은 입력에서 위 거절을
        # 통과했으므로, 역사를 고정하고(task 7.5.2.1) 결함을 예외로 올린 뒤로는 계산이 그 값 이하에서 받는다.
        # **"뿐이다" 는 아니다** (task 7.5.2.2, 재리뷰 Codex): sha 를 고정해도 git 이 그 역사를 **읽는 방식**
        # (얕은 복제의 경계 · 대체 ref · graft)이 실행 중 바뀌면 순회가 다른 답을 내고, 빈 값도 온다. 7.5.2.1 은
        # "빈 값은 못 온다" 며 그 사유 문장을 지워 `computes ()` 를 찍었다 — 사유를 다시 말한다.
        named = computed[:12] if computed else f"none — {why}"
        raise ValueError(
            f"landing point {candidate[:12]} is not the landing this change's evidence "
            f"computes ({named}): the record must be the gate's own value — "
            f"`--record-landing` writes it — not one of the later commits that also match "
            f"— {LANDING_RECOVERY}"
        )
    return candidate


def normalized_source(value: str, root: Path) -> tuple[Path, str]:
    # `root.resolve()` 는 이 호출 안에서 **불변**인데 두 번 돌고 있었다 (task 7.5). 한 번
    # 재서 두 자리가 같이 쓴다 — 판정은 그대로이고 파일시스템 왕복만 셋에서 둘로 준다
    # (2026-09-18 프로파일: 이 함수가 `_pinning_bundles` 안에서 resolve 를 호출당 3회,
    # a071 의 walk 하나에 35,805회 · lstat 216,876회).
    # 기준점도 `realpath` 다 (task 7.5.2.2) — 두 경로를 같은 도구로 풀어야 비교가 같은 규칙이다.
    anchor = Path(os.path.realpath(root))
    raw = Path(value)
    path = raw if raw.is_absolute() else root / raw
    # `Path.resolve()` 가 아니라 `os.path.realpath` 다 (task 7.5.2.1, 재리뷰 보안 전문가). 둘은 같은 풀이인데
    # 3.12 의 `resolve()` 만 심링크 고리에서 `RuntimeError` 를 **더** 낸다(2026-09-20 실측: 3.12.3 은 내고 3.13.13 ·
    # 3.14.5 는 안 낸다, `realpath` 는 셋 다 같은 경로를 돌려준다) — 판정 쪽은
    # `ValueError` 만 대상 이름을 붙여 받으므로, 고리 하나가 판정 전체를 한 줄로 바꾸고 다른 대상의 오류를
    # 지웠다. 판정이 파이썬 판본의 함수여서는 안 된다. 고리는 풀리지 않은 채 남고, 그 경로에는 읽을
    # 소스가 없으므로 "AST source is missing" 이 된다 — 사실 그대로다.
    resolved = Path(os.path.realpath(path))
    if not resolved.is_relative_to(anchor):
        raise ValueError("AST source escapes repository")
    return resolved, resolved.relative_to(anchor).as_posix()


def branch_ids(text: str) -> list[str]:
    return [
        match.group(1)
        for line in text.splitlines()
        if (match := re.match(r"^\|\s*(B\d+)\s*\|", line))
    ]


# Prose coordinates drift silently: the bundle keeps its files, its hash stays
# current and every AST branch stays covered while the map underneath describes
# a revision the function no longer has. Both claims below are optional in the
# corpus — most maps cite neither — so they are checked only where stated.
SOURCE_RANGE = re.compile(r"Source:\s*`[^`]+`\s*\((\d+)\s*[-–]\s*(\d+)\)")
# Anchored to AST on purpose: "미테스트 분기 5개" is prose about coverage, not a
# claim about what the extractor found, and must not be read as one.
AST_BRANCH_COUNT = (
    re.compile(r"AST[^\n]*?\bbranches\s+(\d+)"),
    re.compile(r"AST[^\n]*?분기\s+(\d+)"),
)


# B-T1: until 19판 the checker never opened the file a row pointed at. It
# confirmed the bundle's four files existed, that the AST hash was current and
# that every branch ID had a row -- every one of which a fabricated test name
# satisfies. Ten of a092's rows carried false coverage claims through eighteen
# rounds on exactly that gap.
CITED_TEST = re.compile(r"`(Test[A-Za-z0-9_]+)`")
CITED_TEST_LINE = re.compile(r"`([A-Za-z0-9_./-]*_test\.go):(\d+)`")
GO_TEST_DECL = re.compile(r"^func\s+(Test[A-Za-z0-9_]+)\s*\(")


def test_spans(path: Path) -> dict[str, tuple[int, int]]:
    """Map each Test function in one file to the line range of its body."""
    spans: dict[str, tuple[int, int]] = {}
    try:
        # 깔때기로 읽는다 (task 7.5.2.3) — 옛 `read_text` 는 그 자리의 FIFO 에 영원히 멎었고, 이 읽기가
        # 원장에 없으면 시험 색인(판정의 입력)이 재확인 밖에 남는다. 글자가 아닌 바이트는 옛 판본과 같이
        # `errors="replace"` 로 견딘다 — 이 판정은 Go 시험 선언만 찾고, 색인에 없는 이름은 인용 판정이 댄다.
        lines = _read_regular(path).decode("utf-8", "replace").splitlines()
    except OSError:
        return spans
    for index, line in enumerate(lines):
        declared = GO_TEST_DECL.match(line)
        if not declared:
            continue
        depth = 0
        for offset in range(index, len(lines)):
            depth += lines[offset].count("{") - lines[offset].count("}")
            if depth <= 0 and offset > index:
                spans[declared.group(1)] = (index + 1, offset + 1)
                break
        else:
            spans[declared.group(1)] = (index + 1, len(lines))
    return spans


def test_index(root: Path) -> dict[str, list[tuple[Path, int, int]]]:
    """Every Test function in the tree, by name. Built once per check() run."""
    index: dict[str, list[tuple[Path, int, int]]] = {}
    for path in _globbed(root, "*_test.go"):
        if ".git" in path.parts:
            continue
        for name, (start, end) in test_spans(path).items():
            index.setdefault(name, []).append((path, start, end))
    return index


def resolve_test_file(cited: str, package_dir: Path, root: Path) -> Path | None:
    """Resolve a cited test file: a qualified path wins, then the package, then the tree.

    A citation that carries a directory means it, and honouring that is the
    whole point -- `replay_test.go` exists in both internal/execgw and
    internal/journal, and resolving the bare name against the package under test
    silently answered for the wrong one. Telling an author to qualify the path
    is useless if the qualification is then discarded.

    An ambiguous bare name resolves to nothing rather than to a guess.
    """
    # 고르기도 깔때기로 (task 7.5.2.3) — 어느 파일을 읽을지 고르는 것이 판정의 입력이다. 그 고르기가 원장에
    # 없으면 "그때 없던 파일이 지금 있다" 를 재확인이 못 본다.
    if "/" in cited:
        qualified = root / cited
        return qualified if _kind(qualified) == "reg" else None
    local = package_dir / cited
    if _kind(local) == "reg":
        return local
    matches = [path for path in _globbed(root, cited) if ".git" not in path.parts]
    return matches[0] if len(matches) == 1 else None


def test_citation_errors(
    target: str, text: str, package_dir: Path, root: Path, index: dict
) -> list[str]:
    errors: list[str] = []
    for name in sorted(set(CITED_TEST.findall(text))):
        if name not in index:
            errors.append(
                f"{target}: branch test map cites {name}, which is not a Go test "
                f"function anywhere in the tree"
            )
    # What is NOT checked, and why -- because a check nobody can satisfy gets
    # worked around, and a check that fires on true rows is worse than none.
    #
    # 19판 built the stronger version first: the cited line must belong to the
    # test named beside it. It was measured against the corpus and withdrawn.
    # Rows legitimately cite a test's doc comment (a092 cites two), the shared
    # harness a test is built from (`obs_test.go:338` is `newNotifier`, a helper),
    # and -- decisively -- a named test in ONE file alongside independent call
    # sites in ANOTHER. a091's `severityof` row does exactly that and is correct:
    #
    #   `TestAQuarantineCreationIsCritical` `a074_quarantine_event_test.go:16`
    #     · `exitloop_test.go:910`, `:1013`
    #
    # Which coordinate answers for which claim is carried by the prose, not by
    # co-occurrence on a line. So the line check stays at what a coordinate can
    # be wrong about on its own: pointing past the end of the file it names.
    for line in text.splitlines():
        for basename, raw in CITED_TEST_LINE.findall(line):
            path = resolve_test_file(basename, package_dir, root)
            if path is None:
                continue
            number = int(raw)
            try:
                total = len(_read_regular(path).decode("utf-8", "replace").splitlines())
            except OSError:
                # 고를 때는 정규 파일이었는데 읽을 때 사라졌거나 못 읽는다 — 줄 수를 모르면 이 줄에 대해
                # 아무 주장도 하지 않는다. 그 변화는 원장에 남아 끝의 재확인이 댄다 (task 7.5.2.3).
                continue
            if number > total:
                errors.append(
                    f"{target}: branch test map cites {basename}:{number}, "
                    f"past the end of a {total}-line file"
                )
    return errors


def coordinate_errors(target: str, texts: dict[str, str], value: dict, branches: list) -> list[str]:
    errors: list[str] = []
    start = (value.get("start") or {}).get("line")
    end = (value.get("end") or {}).get("line")
    for name in ("function-logic-map.md", "branch-test-map.md"):
        text = texts.get(name, "")
        cited = SOURCE_RANGE.search(text)
        if cited and (int(cited.group(1)), int(cited.group(2))) != (start, end):
            errors.append(
                f"{target}: {name} cites line range "
                f"{cited.group(1)}-{cited.group(2)} but ast.json is {start}-{end}"
            )
        for pattern in AST_BRANCH_COUNT:
            for claim in pattern.finditer(text):
                if int(claim.group(1)) != len(branches):
                    errors.append(
                        f"{target}: {name} claims AST branch count "
                        f"{claim.group(1)} but ast.json has {len(branches)}"
                    )
    return errors


def _bundle_text(target: Path, ast_raw: bytes | None, *, names: tuple[str, ...] | None = None) -> str:
    """번들의 **읽히는 파일 전부**를 이어 붙인다. `ast.json` 은 한 번 읽은 바이트다.

    이름 목록은 여기서 다시 만들지 않고 증거 읽기가 만든 것을 `names` 로 **넘겨받는다** (task 7.5.2.3 ·
    7.5.2.4 정정: 옛 docstring 은 이 함수가 디렉터리를 연다고 읽혔다).

    강제 판정은 "이 change 가 열거를 쓰는가"이고, 그 답은 열거가 번들 안 어느
    파일에 있든 같다. 파일을 **열거해서** 읽으면 그 목록 밖으로 옮기는 것으로
    판정이 꺼진다.

    다섯 라운드 동안 "남는 회피는 X 뿐"이라는 문장이 매번 코드가 실제로 보는
    범위보다 한 칸 넓었다: 절 → 표지 철자 → `##` 절 → 파일 이름 둘 →
    확장자 `.md` 하나. 마지막 것은 9차 적대 리뷰가 `notes.txt` 로 보였다.
    이제 이름으로도 확장자로도 거르지 않고 **읽히는가**로만 거른다 — 목록이
    없으므로 새 파일 이름이나 새 확장자가 이 판정을 비켜 갈 수 없다.

    `ast.json` 은 목록과 **무관하게** 명령이 한 번 읽은 바이트를 쓴다 (task 7.5.2.1, Codex P1 재현).
    7.5.2 는 살아 있는 목록에 `ast.json` 이 있을 때만 그 바이트를 썼다 — 한 번 읽은 뒤 지워지면 판정한
    바이트가 판정에서 빠졌다. 목록은 `iterdir()` 로 만든다: 못 열면 `OSError` 가 올라가 `check` 가
    **이름 댄 판정 줄**로 바꾼다. `glob` 은 그 실패를 삼켜 빈 목록을 돌려줬다 — 그 번들의 표가 열거형
    호출 판정에서 조용히 빠졌다(permissive).

    파일은 `_read_regular` 로만 연다 (task 7.5.2.2). 7.5.2.1 은 이 함수를 다시 쓰며 옛 `is_file()` 거름을 빠뜨려
    FIFO 에 멎고 `/dev/zero` 에 메모리를 다 썼다.

    **조용히 건너뛰는 모양이 없다** (task 7.5.2.3, 재리뷰 보안). 7.5.2.2 는 정규 파일이 아닌 것과 못 푸는
    바이트를 조용히 넘겼는데, 그것이 공격 경로였다: 커밋된 표 파일을 게이트가 **여는 그 순간에만** FIFO 로
    바꿨다 되돌리거나(타이밍 — 재현 하네스 없음, 7.5.2.4 정정) 못 푸는 바이트를 한 개 심으면
    그 파일의 표가 판정에서 빠져 감사가 꺼진다.
    이제 사라진 것(`FileNotFoundError` — 끊긴 심링크 · 목록을 만든 뒤 지워짐)만 건너뛰고 나머지는 전부 올라가
    **이름 댄 판정 줄**이 된다. 건너뛴 그 실패도 원장에 남으므로, 이름을 빼고 되돌리는 공격은 끝의 재확인이 댄다.

    목록은 **호출자가 한 번 읽은 것**을 받는다 (task 7.5.2.3) — 증거 읽기가 이미 만든 목록이다. 안 주면
    스스로 읽는다(단위 시험 · 번들 하나만 볼 때)."""
    listed = names if names is not None else tuple(name for name, _ in _listed(target))
    paths = {name: target / name for name in listed if name != "ast.json"}
    if ast_raw is not None:
        paths["ast.json"] = target / "ast.json"
    texts = []
    for name in sorted(paths):
        if name == "ast.json":
            raw = ast_raw
        else:
            try:
                raw = _read_regular(paths[name])
            except FileNotFoundError:
                continue                        # 끊긴 링크 · 목록을 만든 뒤 사라진 파일 — 없는 것은 없는 것이다
        try:
            texts.append(_decoded(raw))
        except UnicodeDecodeError as exc:
            # 글자가 아니면 표가 들어 있을 수 없다 — 그래서 **이름을 대고 멈춘다**. 조용히 넘기면 그 파일에
            # 있던 표가 감사에서 빠지고, 그것이 저장소에 커밋된 채로 남을 수 있다(task 7.5.6 의 나머지 절반).
            raise NotUtf8Text(errno.EILSEQ, NOT_UTF8, str(paths[name])) from exc
    return "\n".join(texts)


def _decoded(raw: bytes) -> str:
    """`Path.read_text(encoding="utf-8")` 와 **같은** 글자 — 줄바꿈 변환(`newline=None`)까지 같다.

    `check` 가 한 번 읽은 `ast.json` 바이트를 여러 판정이 나눠 쓰려면 글자로 바꾸는 규칙도
    디스크에서 읽던 것과 같아야 한다 (task 7.5.2). 못 푸는 바이트는 `read_text` 처럼 터진다.
    """
    return io.TextIOWrapper(io.BytesIO(raw), encoding="utf-8").read()


# 판정이 **꺼내 쓰는** 칸의 모양 (task 7.5.2.1). 이 모양이 아니면 뒤의 판정 어딘가가 `AttributeError` ·
# `TypeError` 로 판정 줄 없이 끝난다. 7.5.2 는 이 검사를 대상 판정(`validate_target`) 안에만 둬서, 대상
# 판정보다 **먼저** 같은 값을 쓰는 열거형 호출 판정(`role_check.call_enumeration_in_use`)이 열 가지 모양에서
# 터졌다(`calls`/`returns` 가 수 · 참거짓 · 최상위가 목록 · 글자 · 수 · 참거짓 — 재리뷰, 이 세션이 재현).
# 비었거나 없는 칸은 모양이 맞는 것으로 친다 — 추출기는 빈 목록을 `null` 로 적는다.
_AST_OBJECTS = ("start", "end")
_AST_LISTS = ("branches", "calls", "returns")


def _parse_ast(raw: bytes) -> tuple[dict, str]:
    """읽은 `ast.json` 바이트 → `(값, 못 쓰는 사유)`. 사유가 있으면 값은 `{}` 다.

    모양 검사가 사는 **유일한** 자리다 — 바이트를 값으로 푸는 곳. 사유는 둘이다: `invalid`(글자나 JSON 이
    아니거나 꺼내 쓰는 칸의 모양이 틀림) · `placeholder`(사전이 아님 — 옛 대상 판정의 말 그대로). 필수 칸이
    **비었는지**는 대상 판정이 따로 본다 — 다른 소비자는 빈 칸을 `.get` 으로 읽으므로 터지지 않는다.
    저장소 3,048 번들 전수에서 이 둘에 걸리는 것 0 (review.md `## MEASURE · Pre-Edit Gate — task 7.5.2.1`).
    """
    try:
        value = json.loads(_decoded(raw))
    except (ValueError, RecursionError):
        # `UnicodeDecodeError` 도 `ValueError` 다. `RecursionError` 는 아주 깊은 배열 · 사전이다 — `RuntimeError` 의
        # 하위형이라 결함으로 올라가 대상 이름 없이 판정 전체를 한 줄로 바꿨다 (task 7.5.2.2, 재리뷰 적대).
        return {}, "invalid"
    if not isinstance(value, dict):
        return {}, "placeholder"
    if any(value.get(key) and not isinstance(value[key], dict) for key in _AST_OBJECTS) \
            or any(value.get(key) and not isinstance(value[key], list) for key in _AST_LISTS):
        return {}, "invalid"
    return value, ""


def _parsed(raw: bytes | None) -> dict:
    """판정이 꺼내 쓸 수 있는 값. 없거나 · 깨졌거나 · 모양이 틀리면 `{}` — 모르는 것은 근거가 아니다.

    7.5.1 까지는 `_ast_value(path)` 가 디스크를 직접 읽었다. 이제 증거는 `_read_evidence` 가 한 번
    읽고 모든 판정이 그 바이트를 이 함수로 푼다 (task 7.5.2) — 디스크를 읽는 철자가 둘이면 둘째가
    판정하지 않은 바이트를 들여온다.
    """
    return {} if raw is None else _parse_ast(raw)[0]


def validate_target(
    target: Path, root: Path, index: dict | None = None, require_calls: bool = False,
    revision_ref: str = "", prefetched: dict[str, bytes | None] | None = None,
    *, held: dict[Path, bytes | None],
) -> tuple[list[str], tuple[str, str] | None]:
    """`revision_ref` 가 있으면 `revision: current` 해싱을 그 커밋에서 한다.

    비교 대상과 증거 대조 대상이 갈리면 하나는 착지를, 하나는 오늘을 기술하는 두
    정본이 된다. 병합 뒤에는 그 둘이 반드시 어긋난다.

    `held` 는 명령이 **한 번** 읽은 `ast.json` 들이다(`Evidence.held` — 착지가 판정한 바로 그 값,
    task 7.5.2). `ast.json` 을 디스크에서 다시 읽지 않는다 — 그 사이 바뀌거나 **새로 생긴** 증거로 판정하면
    착지가 판정하지 않은 증거가 판정의 입력이 된다. 그 읽기에 없던 것은 없는 것이다. 필수 인자다
    (task 7.5.2.1) — 시험만 쓰던 "안 주면 디스크에서" 는 둘째 읽기 길이었다."""
    errors: list[str] = []
    texts: dict[str, str] = {}
    for name in REQUIRED:
        path = target / name
        if name == "ast.json":
            if path not in held:
                errors.append(f"{target.name}: missing {name}")
                continue
            if held[path] is None:
                errors.append(f"{target.name}: {name} could not be read")
                continue
            try:
                texts[name] = _decoded(held[path])
            except ValueError:
                texts[name] = ""                # 글자가 아니다 — 아래 `_parse_ast` 가 invalid 라고 말한다
        else:
            # 산문도 `_read_regular` 로 연다 (task 7.5.2.2). `exists()` 뒤 `read_text` 는 그 자리의 FIFO 에 영원히
            # 멎었다(앞 로트부터). 정규 파일이 아니거나 못 읽으면 `ast.json` 과 같은 "could not be read" 다.
            # 글자가 아니면 옛 판본처럼 올라간다(`_decoded` 가 `read_text` 와 같은 글자를 낸다).
            try:
                raw = _read_regular(path)
            except FileNotFoundError:
                errors.append(f"{target.name}: missing {name}")
                continue
            except OSError:
                errors.append(f"{target.name}: {name} could not be read")
                continue
            try:
                texts[name] = _decoded(raw)
            except UnicodeDecodeError:
                # 못 푸는 필수 산문은 **이 대상의** 판정 줄이다 (task 7.5.2.3, 재리뷰 적대). 올려 보내면
                # 판정 전체가 대상 이름 없는 한 줄이 되어, 저자는 어느 번들인지 알 수 없다.
                errors.append(f"{target.name}: {name} is not UTF-8 text")
                continue
        if "TODO" in texts[name]:
            errors.append(f"{target.name}: {name} still contains TODO")
    if "ast.json" not in texts:
        return errors, None
    # 모양은 바이트를 값으로 푸는 **한 곳**(`_parse_ast`)이 본다 (task 7.5.2.1). 7.5.2 는 여기서만 봐서
    # 이 함수보다 먼저 같은 값을 쓰는 열거형 호출 판정이 열 가지 모양에서 터졌다.
    value, problem = _parse_ast(held[target / "ast.json"])
    if problem == "invalid":
        return errors + [f"{target.name}: ast.json is invalid"], None
    keys = ("file", "source_sha256", "package", "function", "signature", "start", "end")
    if problem or any(not value.get(key) for key in keys):
        return errors + [f"{target.name}: ast.json is placeholder evidence"], None
    try:
        source, relative = normalized_source(str(value["file"]), root)
    except ValueError as exc:
        return errors + [f"{target.name}: {exc}"], None
    revision = value.get("revision", "current")
    if revision == "current":
        if revision_ref:
            # `check` 가 착지에서 한 번 미리 읽은 것을 쓴다 (task 7.5.1, F3). 빠진 경로면 예전처럼
            # 여기서 읽는다 — 미리 읽기는 **최적화**이고 판정은 그것에 기대지 않는다.
            blob = (prefetched[relative] if prefetched is not None and relative in prefetched
                    else _committed_bytes(root, revision_ref, relative))
            if blob is None:
                errors.append(f"{target.name}: AST source is missing: {relative}")
            elif hashlib.sha256(blob).hexdigest() != value["source_sha256"]:
                errors.append(f"{target.name}: AST source hash is stale: {relative}")
        else:
            # 워킹트리 소스도 깔때기로 (task 7.5.2.3): `is_file()` + `read_bytes()` 가 하나가 되고, 그 읽기가
            # 원장에 남아 재확인이 본다. **없는 것과 못 읽는 것을 가른다** — 한 말로 하면 저자가 있는 파일을
            # 다시 만들러 간다(7.5.1 의 "git 이 못 돌았다 ≠ 없다" 와 같은 구분).
            try:
                blob = _read_regular(source)
            except (FileNotFoundError, NotADirectoryError):
                errors.append(f"{target.name}: AST source is missing: {relative}")
            except OSError as exc:
                errors.append(f"{target.name}: AST source could not be read: {relative} ({_why(exc)})")
            else:
                if hashlib.sha256(blob).hexdigest() != value["source_sha256"]:
                    errors.append(f"{target.name}: AST source hash is stale: {relative}")
    elif revision != "base":
        errors.append(f"{target.name}: unsupported AST revision {revision!r}")
    function = qualified(value)
    logic = texts.get("function-logic-map.md", "")
    for section in MAP_SECTIONS:
        if section not in logic:
            errors.append(f"{target.name}: function map missing section {section}")
    if relative not in logic or function not in logic:
        errors.append(f"{target.name}: function map is not source/function-bound")
    # The Go extractor marshals a nil slice as JSON null, so a branchless
    # function arrives as "branches": null rather than []. Treat both as empty.
    branches = value.get("branches") or []
    branch_map = texts.get("branch-test-map.md", "")
    if "# Branch Test Map" not in branch_map:
        errors.append(f"{target.name}: branch test map header missing")
    mapped = branch_ids(branch_map)
    if len(mapped) != len(set(mapped)):
        errors.append(f"{target.name}: branch test map contains duplicate branch IDs")
    expected = {
        str(branch.get("id"))
        for branch in branches
        if isinstance(branch, dict) and branch.get("id")
    }
    if branches and len(expected) != len(branches):
        errors.append(f"{target.name}: AST branches do not have stable unique IDs")
    missing = expected - set(mapped)
    if missing:
        errors.append(
            f"{target.name}: branch test map is missing AST branches {sorted(missing)}"
        )
    # Covering every AST branch says nothing about rows the AST never produced.
    # A map left over from a wider revision keeps naming branches that are gone,
    # and each one carries a coverage claim no source line answers for.
    unexpected = set(mapped) - (expected if branches else {"B1"})
    if unexpected:
        errors.append(
            f"{target.name}: branch test map has rows for branch IDs absent "
            f"from the AST {sorted(unexpected)}"
        )
    errors.extend(coordinate_errors(target.name, texts, value, branches))
    # 좌표의 **역할**은 위의 검사가 보지 않는다. 범위와 개수만 맞으면 분기 표에
    # 분기가 아닌 좌표를 적어도, 호출 표를 40행에서 잘라도 통과했다 — a112 3라운드
    # 적대 리뷰가 그 구멍에 오류 넷을 심어 전부 통과시켰고, 저장소의 열거형 호출 표
    # 세 개가 실제로 잘려 있었다.
    errors.extend(role_errors(target.name, logic, value, branch_map, require_calls))
    errors.extend(
        test_citation_errors(
            target.name,
            branch_map,
            (root / relative).parent,
            root,
            test_index(root) if index is None else index,
        )
    )
    if not branches and "B1" not in mapped:
        errors.append(f"{target.name}: branchless function still needs one happy-path row")
    risk = texts.get("risk-pattern-report.md", "")
    if "# Risk Pattern Report" not in risk or relative not in risk:
        errors.append(f"{target.name}: risk report is not source-bound")
    return errors, (relative, function)


def check(
    change: str, root: Path = ROOT, context: dict[str, object] | None = None
) -> list[str]:
    """이 change 의 판정 줄들. 빈 목록이면 증거가 완전하다.

    **이 함수가 원장을 열고 닫는다** (task 7.5.2.3). 판정이 **이 모듈의 파이썬 코드로** 읽은 것은 전부 원장에
    남고, 끝의 재확인이 다시 읽는 집합은 손으로 고른 것이 아니라 그 원장이다.
    7.5.2.2 는 그 집합을 손으로 골라(`HEAD` + `Evidence`) 판정이 읽는 1,739 중 149 만 봤다(a112 실측).

    **범위를 정확히 적는다 (task 7.5.2.4 · 7.5.22 정정).** 원장은 자식 프로세스가 읽는 것을 **하나도**
    못 본다. 이 파일의 `subprocess.run` 은 **열여섯 자리**이고 원장에 남는 것은 **0** 이다 — 앞 로트가
    넷(`git diff` · `git show` · `go run` · `execution_baseline`)만 대고 그친 것은 열거가 아니라 예시였다.
    판정 입력을 읽는 것만 꼽아도 `changed_existing_functions` 의 `git diff` · `base_file` 의 `git show` ·
    `go_functions` 의 `go run`(워킹트리 Go 바이트) · `_safe_changed_go_paths` 의 `git diff --numstat` ·
    `_committed_many` 의 `git cat-file` · `_recording_refusal` 의 `git diff --quiet` · 역사를 걷는 **열**
    자리 · 그리고 `execution_baseline.validate` 다 (7.5.23 정정: 앞 판본은 "아홉" 이라 적었다 —
    6 + 9 = 15 라 자기가 바로 앞에 적은 16 과 안 맞았다. 열거를 고치면서 열거를 틀렸다).
    그것들이 판정 도중 바뀌면 재확인이 통과한다.
    a112 실측(7.5.22): `required` 를 정하는 Go 파일 **32** 중 **13** 은 원장에 이름조차 없고, 나머지 19 도
    **인용·시험 색인 때문에** 있는 것이지 `go run` 이 읽어서가 아니다. 7.5.2.3 의 README·VERIFY 가
    "디스크를 읽는 자리가 깔때기 넷뿐" 이라고 적은 것은 **거짓**이었다 — 그 열거의 모집단이
    이 모듈의 AST 였고 자식 프로세스는 애초에 모집단에 없었다. 닫는 것은 7.5.8 이다.

    판정 **앞에서** 돌아가는 결함 · 거절은 대조하지 않는다 — 이미 빨갛다. 그래서 `_judged` 가 끝까지 갔는지를
    같이 돌려준다.
    """
    facts: dict[str, object] = {} if context is None else context
    # 이 사전은 **이번 실행의 사실만** 담는다 — 맨 앞에서, 어느 반환보다 먼저 비운다 (task 7.5.2.2). 7.5.2 는 둘만,
    # 7.5.2.1 은 손으로 적은 목록(`RUN_FACTS`)을 id 해소 **뒤에서** 지워서, 오타 난 id 에는 앞선 실행의 창
    # (`landing` · `head` · …)이 남았다(재리뷰 레드팀 · Codex). 문맥은 출력 통로다 — 호출자가 미리 넣어 두고 읽는
    # 키는 없다(`main` · 하네스 · 시험 전부 빈 사전). 목록이 없으니 새 사실을 더해도 지우는 쪽을 잊을 수 없다.
    facts.clear()
    with _ledger() as book:
        verdict, judged = _judged(change, root, facts)
        if not judged:
            return verdict
        # 판정을 내놓기 **직전**, 판정이 읽은 것이 아직 그대로인지 본다 (task 7.5.2.2 · 7.5.2.3).
        moved = _judged_state_moved(root, str(facts["head"]), book)
        return [moved] if moved else verdict


def _judged(
    change: str, root: Path, facts: dict[str, object]
) -> tuple[list[str], bool]:
    """판정 줄들과 "끝까지 갔는가". **이 모듈이 파이썬으로** 디스크를 읽는 것은 전부 깔때기로 간다 —
    열린 원장에 남는다 (task 7.5.2.3). 자식 프로세스(`git` · `go run`)가 읽는 것은 원장 밖이다 (7.5.2.4 정정).

    `check` 에서 떼어 낸 까닭은 하나다: 원장을 여는 자리와 판정하는 자리를 가르면 조기 반환마다 재확인을
    적을 필요가 없다(적는 자리가 여럿이면 하나를 잊는다 — 7.5.2.1 의 실패 모양).
    """
    # 아카이브된 change 도 자기 id 로 재검사할 수 있어야 한다. 아카이브 문법을 아는
    # 해소기가 이미 있으므로 세 번째 사본을 만들지 않는다.
    # 해소기가 못 찾은 id 는 **그 문장으로** 멈춘다 (task 7.6, 리뷰 I4). 예전에는 없는
    # `openspec/changes/<id>` 로 바꿔 넘겨서 "base 를 capture 하라"는 조언이 나갔다 —
    # 오타 난 id 에 새 change 를 만들라는 말이다. 2026-09-13 전수: 게이트가 받을 수 있는
    # id 126개 중 그 갈래로 떨어지는 것 0.
    # 문맥은 **항상** 채운다. 호출자가 안 줘도 판정 자신이 읽어야 하는 사실이 여기
    # 들어온다(이관인가, 감사된 source 는 무엇인가). 호출자가 문맥을 줬으면 같은
    # 사전이므로 밖에서 보이는 것은 그대로다.
    try:
        change_dir = resolve_referenced_change(root, change)
    except ValueError as exc:
        return [str(exc)], False
    analysis = change_dir / "analysis" / "function-logic"
    reference_file = change_dir / "analysis" / "function-logic-reference.txt"
    review = change_dir / "review.md"
    # 면제 표지를 읽는 자리다 — 깔때기로 (task 7.5.2.3). 옛 `exists()` 뒤 `read_text` 는 그 자리의 FIFO 에
    # 영원히 멎었고(재리뷰 보안), 이 읽기가 원장에 없으면 면제 경로의 재확인 표본이 **0** 이었다
    # (증거가 없으면 비교할 바이트가 없다 — 전칭이 공허하게 참).
    try:
        review_text = _decoded(_read_regular(review))
    except FileNotFoundError:
        review_text = ""
    except (OSError, UnicodeDecodeError) as exc:
        return [UNREADABLE.format(what="review.md", why=_why(exc))], False
    try:
        base = resolve_base(change_dir, root, facts, change_id=change)
        # 역사는 **여기서 한 번** 푼다 (task 7.5.2.1 — 재리뷰 적대 · 레드팀 · 이 세션이 재현). 7.5.2 는 상징
        # `HEAD` 를 열 자리에서 따로 읽고 지문에 표본 하나를 넣었다: 기록은 표본 **전에** 읽혔고 뒤의 git
        # 호출은 살아 있는 `HEAD` 를 다시 읽어서, 가지 전환 한 번도 떠났다 돌아온 `HEAD` 도 rc 0 이었다.
        # 이 뒤의 역사 읽기는 전부 이 sha 다 — 비교할 표본이 없으면 섞일 것도 없다. 태어나지 않은
        # `HEAD`(첫 커밋 전 고아 가지)는 여기서 결함이 된다(옛 판본은 워킹트리 창으로 판정했다 — 막는 쪽).
        head = _head_commit(root)
    except GATE_FAULTS as exc:
        return [f"cannot derive modified Go functions: {exc}"], False
    facts["head"] = head
    # 빌린 증거는 착지 판정 **앞에서** 푼다. 착지가 유효한지는 그 change 가 실제로
    # 딛는 증거로 판정하는데, 빌린 change 는 지역 번들이 0 이라 뒤에서 풀면 고정할
    # 것이 하나도 없는 채로 판정이 끝난다(a073 이 그 모양이다).
    # 참조 파일도 깔때기로 (task 7.5.2.3) — `exists()` 뒤 `read_text` 는 그 자리의 FIFO 에 영원히 멎었다.
    try:
        reference_raw: bytes | None = _read_regular(reference_file)
    except FileNotFoundError:
        reference_raw = None
    except OSError as exc:
        return [UNREADABLE.format(what="function-logic-reference.txt", why=_why(exc))], False
    if reference_raw is not None:
        # 지역 증거가 있으면 빌릴 수 없다. 여기서는 번들 **목록**만 본다 — 판정이 딛는 증거는 빌려주는
        # 쪽이므로 지역 바이트를 읽을 이유가 없고, 증거를 읽는 자리는 아래 **하나**로 남는다.
        try:
            local = [analysis / name for name, is_dir in _listed(analysis) if is_dir]
        except (FileNotFoundError, NotADirectoryError):
            local = []
        if any(_listed(bundle) for bundle in local):
            return ["function-logic reference cannot coexist with local function-logic evidence"], False
        # 글자가 아닌 바이트는 아래 이름 판정이 거절한다 — 해독 실패로 판정 줄을 잃지 않는다.
        referenced_change = reference_raw.decode("utf-8", "replace").strip()
        if not re.fullmatch(r"[a-z0-9][a-z0-9-]*", referenced_change) or referenced_change == change:
            return ["function-logic reference names an invalid or recursive change"], False
        try:
            referenced_dir = resolve_referenced_change(root, referenced_change)
            referenced_base = resolve_base(referenced_dir, root, change_id=referenced_change)
        except GATE_FAULTS as exc:
            return [f"function-logic reference base is invalid: {exc}"], False
        if referenced_base != base:
            return ["function-logic reference must share the exact comparison base"], False
        # 빌리는 change 는 창을 **좁히지 않는다** (task 7.2.3, 리뷰 C4 — 사람이 2026-09-14 에
        # 골랐다). 1.8 은 빌려주는 쪽의 착지를 복사하게 했는데, 그 착지는 빌려주는 쪽 증거의
        # 가장 낮은 값이라 빌리는 쪽이 그 **뒤에** 한 Go 작업이 창 밖으로 나갔다. 빌리는 쪽은
        # 자기 번들이 0 이라 자기 작업이 어디 착지했는지 말할 증거를 소유하지 않는다.
        # 그래서 빌리는 쪽의 기록은 이름으로 거절하고, 빌려주는 쪽의 기록은 이 창에 쓰지 않는다
        # (아래 `resolve_landing` 은 빌리는 쪽 디렉터리를 보므로 기록이 없으면 빈 값 = 워킹트리).
        # 있는가만 묻는다 — 해독하면 못 읽는 기록 앞에서 질문이 터진다(6.2.1).
        # 오늘 저장소의 빌리는 change 는 a073 하나이고 기록이 없다(2026-09-14 확인).
        if _landing_record(change_dir, root, head) is not None:
            return [BORROWED_REFUSES_A_LANDING], False
        analysis = referenced_dir / "analysis" / "function-logic"
    # 증거는 **여기 한 자리에서 한 번** 읽는다 (task 7.5.2 · 7.5.2.1 · 7.5.2.3) — 디렉터리가 있었나 · 어떤
    # 번들이 있었나 · 번들 안의 이름 목록 · 각 `ast.json` 의 바이트. 빌리는 change 면 위에서 `analysis` 가
    # 빌려주는 쪽으로 바뀌었다 — 판정이 딛는 증거는 그쪽이다. 읽는 자리가 둘이면 어느 쪽이 판정에 들어갔는지
    # 시험이 못 세고, 7.5.2.2 는 공존 판정이 같은 디렉터리를 **따로** 물어서 두 대답을 만들었다.
    try:
        evidence = _read_evidence(analysis)
    except GATE_FAULTS as exc:
        return [f"cannot derive modified Go functions: {exc}"], False
    adopted = bool(facts.get("execution_baseline_adoption"))
    if adopted and _landing_record(change_dir, root, head) is not None:
        # 이관 예외의 정당성은 "판정에 들어가는 입력을 하나도 빠짐없이 열거하고
        # digest 로 묶었다"이다. `landed-commit.txt` 는 `openspec/` 아래라 drift 검사가
        # 통과시키고, 추적 파일이라 untracked 감사도 못 보고, 닫힌 키 집합에도 없다.
        # 그런데 비교 대상을 고른다 — 손잡이는 하나여야 하고 그것은 감사된 쪽이다.
        return [ADOPTION_REFUSES_A_LANDING], False
    try:
        # 조상 판정은 `base-commit.txt` 의 글자가 아니라 `resolve_base` 가 **반환한**
        # 값에 건다. a063 은 그 둘이 다르다(P → E).
        # 이관이면 착지를 **해소하지 않는다**. `resolve_landing` 의 판정들은 저자가
        # 고른 값을 위한 것이고, 감사된 source 는 고른 값이 아니다 — `validate` 가
        # ancestry(P,E)·ancestry(E,source)·ancestry(source,head,strict)·tree 대조·
        # digest 셋으로 이미 묶는다. 같은 판정을 두 번 하지 않는다.
        # 착지 판정 · 대상 판정 · base 모양 조언은 위에서 **한 번 읽은** 증거를 쓴다 (task 7.5.2 · 7.5.2.1).
        landing = str(facts.get("adoption_source", "")) if adopted \
            else resolve_landing(change_dir, root, base, head, evidence)
        required = changed_existing_functions(root, base, landing)
    except GATE_FAULTS as exc:
        return [f"cannot derive modified Go functions: {exc}"], False
    facts["landing"] = landing
    facts["required_count"] = len(required)
    if not landing:
        # 조언 줄은 대상이 워킹트리일 때만 나가므로 그때만 잰다 (task 7.1).
        # **조언은 판정이 아니다** (task 7.5.2, 재리뷰 적대 F6): 여기서 git 이 죽으면 옛 판본은 판정
        # 전체를 결함 한 줄로 바꿨다 — 워킹트리가 대상인 판정은 git 없이 디스크에서 서는데도.
        # 결함은 조언 줄이 말한다(`main`).
        try:
            facts["base_shaped_bundles"] = _base_shaped_bundles(root, base, evidence)
        except GATE_FAULTS as exc:
            facts["base_shaped_fault"] = str(exc)
    # 스냅숏은 판정을 **내적으로** 일관되게 하고, `check` 의 재확인은 그 판정이 **지금 거기 있는 것**의
    # 판정임을 보인다 — 둘 다 있어야 한다 (task 7.5.2.2 · 7.5.2.3).
    return _verdict(root, base, landing, adopted, required, evidence, review_text), True


def _verdict(
    root: Path, base: str, landing: str, adopted: bool, required: dict[tuple[str, str], dict],
    evidence: Evidence, review_text: str,
) -> list[str]:
    """한 번 읽은 증거로 대상들을 판정한다 — `check` 가 창(base · 착지 · 요구 집합)을 정한 **뒤**의 전부.

    `check` 에서 떼어 낸 까닭은 하나다 (task 7.5.2.2): 판정이 어느 갈래로 끝나든 `check` 의 **한 출구**에서
    역사와 증거가 아직 그대로인지 대조하려고. 안에서는 디스크의 증거를 다시 읽지 않는다 — `evidence` 만 본다.
    """
    if not evidence.present:
        if required:
            names = ", ".join(f"{source}:{function}" for source, function in required)
            # 이름만 쏟아내면 "왜 이것들이 요구되는가"가 안 남는다. 2026-09-10 실측으로
            # a076 은 이름 316개를 이은 21,838자짜리 한 줄이었고 창을 말하는 줄이 0 이었다.
            return [
                f"missing Function Logic Map for {len(required)} function(s) modified "
                f"between base {base[:12]} and {_target_text(landing, adopted)}: {names}"
            ]
        return [] if EXEMPTION in review_text else [f"missing analysis or `{EXEMPTION}` review marker"]
    errors: list[str] = []
    if not evidence.targets:
        return ["function-logic analysis directory has no targets"]
    covered: dict[tuple[str, str], Path] = {}
    # Built once: the tree has thousands of test functions and every target
    # would otherwise rescan them.
    index = test_index(root)
    # 열거형 호출 표를 **어디서든** 쓰는 change 는 모든 번들에서 써야 한다.
    # 번들마다 표지를 고를 수 있으면 감사 여부를 저자가 정하게 되고, 그 문으로
    # a112 의 39개가 빠져나갔다(4차 적대 리뷰). 판정은 표지 철자가 아니라
    # 표의 **내용**으로 한다 — 철자로 보던 판본이 공백 하나에 뚫렸다(6차).
    # 참여하는 번들은 한 번 읽은 목록에서 고른다 — 이름이 있으면 든다(끊긴 링크 · 디렉터리도).
    # (`lexists` 로 여기서 다시 묻던 판본은 7.5.2.3 에서 없어졌다 — 7.5.2.4 정정.)
    # 그 번들의 파일 목록을 못 열면 **이름 댄 판정 줄**이다 (task 7.5.2.1) —
    # 그 번들의 산문 없이 판정하면 그 번들의 표가 조용히 빠진다.
    bundle_texts: dict[Path, str] = {}
    for target in evidence.targets:
        # 목록은 증거 읽기가 한 번 만든 것이다 (task 7.5.2.3) — 여기서 `lexists` 로 다시 묻던 판본은 감사
        # 참여 여부를 판정과 **다른 읽기**로 정했고, 그 자리는 원장에도 안 남았다.
        if target in evidence.unlistable:
            errors.append(f"{target.name}: cannot read every file in the bundle "
                          f"({evidence.unlistable[target]}: {target.name})")
            continue
        names = evidence.listings.get(target, ())
        if "function-logic-map.md" not in names:
            continue
        try:
            bundle_texts[target] = _bundle_text(
                target, evidence.held.get(target / "ast.json"), names=names)
        except OSError as exc:
            # 번들 파일 하나를 못 읽었다 (task 7.5.2.1 · 7.5.2.2 · 7.5.2.3) — 사유와 이름을 댄다: 권한 ·
            # 디렉터리 · 정규 파일 아님 · 너무 큼 · UTF-8 아님이 모두 이 한 자리로 온다. 이름이 글자가 아니면
            # (`open(fd)` 의 예외는 `filename` 에 **정수 fd** 를 담는다) 번들 이름으로 떨어진다 — 이름을 대려다
            # `TypeError` 로 판정 줄 없이 죽지 않게. 경로를 담는 쪽은 `_opened_bytes` 가 못 박는다.
            named = exc.filename if isinstance(exc.filename, str) else ""
            where = Path(named).name if named else target.name
            errors.append(f"{target.name}: cannot read every file in the bundle ({_why(exc)}: {where})")
    require_calls = call_enumeration_in_use(
        (text, _parsed(evidence.held.get(target / "ast.json"))) for target, text in bundle_texts.items()
    )
    # 착지가 있으면 고정 소스를 그 커밋에서 **한 번** 읽는다 (task 7.5.1, gstack 리뷰 F3).
    # `index` 를 한 번 만드는 것과 같은 모양이다. 7.5 는 walk 만 배치로 바꾸고 이 루프 —
    # **매 게이트 실행의 본 판정 경로**(a112 번들 147 → 프로세스 147) — 를 남겼다.
    # 고르는 것은 착지 판정과 **같은 읽기**(`evidence`)다. 저장소 밖 소스처럼 **고를 수 없는** 번들이
    # 있으면 미리 읽기를 건너뛴다 — 그 대상의 오류는 `validate_target` 이 대상 이름과 함께 낸다 (task 7.5.2,
    # 재리뷰 Codex P2 · 적대 F3). 닿는 것은 착지 판정을 안 거치는 이관 change 다 — 착지 경로는 같은 바이트를
    # 이미 골랐고(못 고르면 거기서 결함), 여기서 달라질 수 있는 것은 소스 경로의 심링크 풀이뿐이다.
    # git 의 결함은 건너뛰지 않는다 — 대상마다 다시 물으면 멎은 git 을 대상 수만큼 기다린다.
    # 미리 읽기는 **최적화**다 — 빠진 경로는 `validate_target` 이 착지에서 읽는다.
    prefetched: dict[str, bytes | None] | None = None
    if landing:
        try:
            sources = sorted({source for _, source, _ in _select_pinning(root, evidence)})
        except ValueError:
            sources = []
        prefetched = _committed_many(root, landing, sources)
    for target in evidence.targets:
        target_errors, binding = validate_target(
            target, root, index, require_calls, landing, prefetched, held=evidence.held,
        )
        errors.extend(target_errors)
        if binding:
            if binding in covered:
                errors.append(f"{target.name}: duplicate evidence for {binding[0]}:{binding[1]}")
            covered[binding] = target
    for binding, expected in required.items():
        target = covered.get(binding)
        if target is None:
            errors.append(f"missing evidence for modified function {binding[0]}:{binding[1]}")
            continue
        # 묶인 대상은 `validate_target` 이 같은 바이트를 풀어 통과시킨 것이다 — 값이 있다.
        ast_value = _parsed(evidence.held.get(target / "ast.json"))
        expected_hash = expected.get("current_hash") or expected.get("base_hash")
        if ast_value.get("source_sha256") != expected_hash:
            errors.append(f"{target.name}: AST hash does not match modified function revision")
        expected_revision = "current" if expected.get("current_hash") else "base"
        if ast_value.get("revision", "current") != expected_revision:
            errors.append(f"{target.name}: AST revision must be {expected_revision}")
    return errors


class LandingInputs(NamedTuple):
    """착지 판정에 들어가는 **한 번 잰** 입력 (task 7.5.1 · 7.5.2 · 7.5.2.1).

    착지를 판정하는 함수는 전부 이 한 벌만 받는다 — 스스로 읽거나 재는 기본값이 없다(7.5.2 재리뷰
    maintainability: 시험만 쓰던 `None` 기본값이 디스크를 읽는 둘째 길로 남아 있었다). 위치 튜플이던
    판본은 `floor` 와 `why`(둘 다 `str`)를 바꿔 넣어도 아무도 못 봤다. `repairs` 는 하한이 못 서면
    **재지 않으므로** `None` 이다 — `[]` 는 "수리 커밋이 없다" 이고 둘을 섞으면 안 된다
    (`_repairs_after` 가 `None` 을 결함으로 친다).

    7.5.2 의 `Fingerprint`(HEAD 표본 · 바이트 해시 · 고정 목록)는 없어졌다: 역사는 표본이 아니라 고정한
    `head` 이고, 바이트는 해시 대신 `evidence` 그 자체를 비교한다.
    """

    head: str                                   # 명령이 한 번 푼 `HEAD` — 모든 역사 읽기의 끝
    evidence: Evidence                          # 한 번 읽은 증거 — 판정은 이 바이트만 본다
    bundles: list[tuple[Path, str, str]]        # 그 읽기에서 고른 고정 번들
    floor: str
    why: str
    repairs: list[str] | None


def _head_commit(root: Path) -> str:
    """지금 `HEAD` 의 커밋. 판정이 쓰는 역사 sha 는 `_judged` 가 **한 번** 푼 그것이고 그 뒤의 역사 읽기는
    전부 그 sha 다 (task 7.5.2.1). **판정이 끝까지 간 경로**에서 이 함수는 `check` 한 번에 셋 불린다:
    판정 앞에 한 번, 재확인의 **앞뒤**로 한 번씩(`_head_moved`). 더 앞에서 거절이 나면 그보다 적다 —
    활성 change 27건 실측(7.5.22): **3회 24건 · 0회 3건**(head 를 묻기 전에 거절). 앞 로트가 적은
    "셋 불린다" 는 조건 없이 쓰여 있어서 거짓이었다. 기록 명령은 쓰기 직전에 자기 몫을 또 묻는다.

    못 읽으면 결함이다 — 빈 값은 "못 물었다" 를 "같다" 로 만든다.
    """
    process = subprocess.run(
        ["git", "rev-parse", "--verify", "HEAD^{commit}"],
        cwd=root, capture_output=True, text=True, timeout=10, check=False,
    )
    if process.returncode:
        raise RuntimeError(f"cannot read HEAD: {_first_line(process.stderr, 'git rev-parse failed')}")
    return process.stdout.strip()


def _head_moved(root: Path, head: str) -> str:
    """판정한 역사가 아직 `HEAD` 인가. 역사를 다시 묻는 **한 자리**다 — 못 읽으면 결함이다(`_head_commit`)."""
    now = _head_commit(root)
    if now != head:
        return JUDGED_STATE_MOVED.format(what=f"HEAD moved from {head[:12]} to {now[:12]}")
    return ""


def _judged_state_moved(root: Path, head: str, book: ReadLedger) -> str:
    """판정한 역사와 **판정이 읽은 것 전부**가 지금도 그대로인가. 달라졌으면 그 문장, 같으면 빈 문자열.

    `check` 의 출구와 기록 명령의 쓰기 직전이 묻는다 (task 7.5.2.2). 판정 · 계산은 명령이 한 번 푼 sha 와 한 번
    읽은 바이트로만 서므로(7.5.2.1) 안에서는 섞이지 않는다 — 이 대조는 **그 판정을 내놓아도 되는가**를 본다. 이
    읽기의 바이트는 판정에 들어가지 않는다.

    대조 집합은 **원장**이다 (task 7.5.2.3). 7.5.2.2 는 `HEAD` + `Evidence` 만 골라서, 판정이 읽는 1,739 경로
    중 149 만 봤다(a112 실측) — 번들 산문 · 워킹트리 소스 · `review.md` · `base-commit.txt` · 시험 색인이 밖에
    있었고, 실행 중 그것을 건드리면 `[]` PASS 였다.

    `HEAD` 를 **앞뒤로** 묻는다 (재리뷰 Codex 적대): 7.5.2.2 는 증거 스캔 앞에서만 물어서 스캔 도중의 커밋이
    통과했다. 원장을 다시 읽는 데 a112 에서 ~0.7s 가 걸리고 그 사이에도 역사는 설 수 있다.

    거절하는 정상 입력: 실행 중 병행 세션의 커밋 · 편집. 값은 재실행 한 번이다 — 7.5.2.1 은 그것을 받아
    들였는데, 그러면 판정은 옛 상태의 것이고 게이트의 나머지는 새 트리로 가서 무엇이 PASS 했는지 아무 데도 안
    남았다(재리뷰 적대 · Codex P1). **재확인 뒤의 창**은 어느 설계로도 못 없앤다 — rc 를 읽는 `tools/gate.sh`
    까지가 그 창이다.
    """
    moved = _head_moved(root, head)
    if moved:
        return moved
    changed = _reads_moved(root, book)
    if changed:
        return JUDGED_STATE_MOVED.format(what=changed)
    return _head_moved(root, head)


def _raise_if_inputs_moved(root: Path, inputs: LandingInputs) -> None:
    """판정한 증거가 지금도 디스크에 그대로인가. 아니면 결함이다 — 그 위의 판정은 어느 쪽이든 믿을 수 없다.

    수락 직전(`compute_landing`)과 선언 경로의 거절 앞(`resolve_landing`)이 **같은 함수**에 묻는다. 이
    읽기의 바이트는 판정에 **들어가지 않는다** — 대조만 한다. 존재 · 번들 목록 · 바이트를 먼저 대조하고,
    같을 때만 고정 목록(소스 경로의 심링크 풀이)을 본다. 그 고르기에서 난 `ValueError`(저장소 밖으로
    바뀜 · 심링크 고리)도 움직임이다 — 7.5.2 는 그것을 "움직였다" 보다 **먼저** 결함 문장으로 냈다(재리뷰
    레드팀). 역사는 비교하지 않는다 — 고정했으므로 움직일 것이 없다 (task 7.5.2.1).
    """
    now = _read_evidence(inputs.evidence.directory)
    if now != inputs.evidence:
        raise RuntimeError(INPUTS_MOVED)
    try:
        bundles = _select_pinning(root, now)
    except ValueError as exc:
        raise RuntimeError(INPUTS_MOVED) from exc
    if bundles != inputs.bundles:
        raise RuntimeError(INPUTS_MOVED)


def _measure_landing_inputs(root: Path, head: str, evidence: Evidence) -> LandingInputs:
    """착지 판정의 입력을 **한 번** 잰다. 선언 경로와 계산 경로가 같은 한 벌을 쓴다.

    7.5 는 "번들 목록을 한 번 잰다" 고 주석을 달았는데 `resolve_landing` 이 재고 나서
    `compute_landing` 이 **또** 쟀다 (task 7.5.1, 서브에이전트 F2 — 실측 `_evidence_floor` ·
    `_self_repair_commits` 각 두 번). 증거는 여기서 읽지 않는다 — 호출자가 한 번 읽은 것을 받는다
    (task 7.5.2.1): 여기서 읽으면 착지가 판정한 바이트와 대상 판정이 읽는 바이트가 다른 읽기가 된다.
    """
    bundles = _select_pinning(root, evidence)
    floor, why = _walk_floor(root, bundles, head)
    repairs = None if why else _self_repair_commits(root, evidence.directory, head)
    return LandingInputs(head, evidence, bundles, floor, why, repairs)


def _walk_floor(root: Path, bundles: list[tuple[Path, str, str]], head: str) -> tuple[str, str]:
    """걷기 **전에** 정해지는 것 — 후보 순회가 설 하한. `(하한, 못 서는 사유)`.

    `bundles` 는 한 번 읽은 증거에서 고른 고정 목록이다 — 여기서 다시 읽지 않는다 (task 7.5.2 · 7.5.2.1).

    `_measure_landing_inputs`(선언 · 계산 경로가 같이 쓴다)와 `_recording_refusal` 이 **같이**
    묻는다 (task 7.7, 리뷰 I8). 조언 줄은 걷지 않고 이것까지만 묻는데, 여기 두 사유를 조언
    쪽이 따로 들고 있으면 계산이 문장을 바꿀 때 조언만 옛 문장으로 남는다
    ([[two-judgements-cover-for-each-other]]).
    """
    if not bundles:
        return "", "no `revision: current` evidence pins a landing for this change"
    floor = _evidence_floor(root, bundles, head)
    if not floor:
        # 걸은 것만 말한다 (task 7.2.4, 리뷰 H7). 하한을 찾는 `git log` 는 `-m` 이 없어서 병합 커밋
        # 자신의 변경을 읽지 않는다 — 병합을 마치며 번들을 처음 커밋하면 여기로 온다. 옛 문장
        # "never entered this history (commit the bundles)" 는 그 경우 거짓이었고, 이미 한 커밋을
        # 또 하라고 권했다. 동작은 사람이 2026-09-14 에 한계로 두기로 골랐다(막는 쪽으로 틀린다).
        return "", (
            "no ordinary commit on this history adds the pinning evidence — commit the bundles "
            "in an ordinary commit (a merge commit's own changes are not read)"
        )
    return floor, ""


def compute_landing(root: Path, base: str, inputs: LandingInputs) -> tuple[str, str]:
    """게이트가 기록할 착지 지점. **저자가 고르지 않는다** (task 6.1.2).

    `(값, 못 정한 사유)` 를 돌려준다. 값이 있으면 사유는 빈 문자열이다. git 이 못 걸으면 사유가
    아니라 **결함**이다 (task 7.5.2.1) — 사유로 돌려주던 판본에서 선언 경로가 그것을 "계산값과 다르다"
    로 읽고 유효한 기록을 지우라고 권했다(재리뷰 레드팀 repro_a).

    고르는 규칙은 하나다: 바닥(그 change 의 고정 번들이 역사에 들어온 지점) 이후
    이면서 고정을 통과하는 **가장 낮은** 커밋. 가장 낮은 것을 고르는 이유는 창을
    좁히는 것이 이 기능의 목적이기 때문이고, 그것이 안전한 이유는 바닥 아래로는
    못 내려가기 때문이다 — 저자는 오늘 만든 번들을 과거 커밋에 넣을 수 없다.

    입력은 호출자가 **한 벌** 잰 것이다(선언 경로는 `resolve_landing`, 기록 경로는 `record_landing`) —
    하한 · 수리 신호 · 고정 목록 · 증거 · 역사의 끝이 전부 그 한 벌에서 온다 (task 7.2.6 · 7.5.1 · 7.5.2.1).
    """
    if inputs.why:
        return "", inputs.why
    floor = inputs.floor
    start = floor if _is_ancestor(root, base, floor) else base
    process = subprocess.run(
        ["git", "rev-list", "--reverse", f"{start}..{inputs.head}"],
        cwd=root, capture_output=True, text=True, timeout=60, check=False,
    )
    if process.returncode:
        raise RuntimeError(
            f"cannot walk the history after {start[:12]}: "
            + _first_line(process.stderr, "git rev-list failed")
        )
    first = ""
    unheld: list[str] = []
    for candidate in [start, *process.stdout.split()]:
        # 선언 경로와 **같은 함수**에 묻는다 (task 7.6, 리뷰 I2). 받는 가장 낮은 후보가 착지다.
        refusal, names = _landing_refusal(root, base, candidate, inputs)
        if not refusal:
            # **수락 직전에** 판정이 딛은 증거가 디스크에 그대로인지 본다 (task 7.5.1 · 7.5.2). 받는
            # 쪽에만 선다 — 거절은 이미 보수적이고, 어느 가드가 거절하는지는 안 바뀐다
            # ([[a-new-guard-unpins-the-guards-behind-it]]).
            _raise_if_inputs_moved(root, inputs)
            return candidate, ""
        # 첫 후보의 거절을 남긴다 (task 6.5). 대개 증거가 역사에 들어온 바로 그 커밋이고,
        # 거기서 이미 틀린 소스가 저자가 고칠 번들이다 — a089 · a095 는 증거를 뽑은 뒤 같은
        # 커밋에서 Go 를 한 번 더 고쳤다. 뒤 후보의 문장에는 남이 나중에 고친 파일까지 붙는다.
        first = first or refusal
        if names:
            # 다른 조건은 다 통과했는데 판정이 읽은 번들을 안 들고 있다 (task 7.2, 리뷰 C3).
            unheld = names
    if unheld:
        return "", (
            "no commit holds the evidence this verdict read (" + ", ".join(unheld) + ") "
            "— commit the bundles as they are on disk"
        )
    # 걸은 것만 말한다 (task 6.5). 옛 꼬리 "the evidence does not describe any revision on
    # this history" 는 순회가 안 걷는 하한 아래까지 주장했고, 2026-09-14 걷기 실패 17건
    # 전수에서 4건(a089 · a095 · 아카이브 둘)은 증거가 하한 아래 커밋을 전부 맞게 기술했다.
    # 머리도 "맞는 커밋이 없다"가 아니다 (task 7.2.2) — 증거와 맞는데 규칙이 거절한 후보가
    # 있다(고정 소스가 base 와 같은 착지). 무엇이 거절했는지는 인용한 문장이 말한다.
    return "", (
        f"no commit at or after the evidence ({floor[:12]}) is accepted as the landing "
        f"— at the first commit walked, {first}"
    )


def _recording_refusal(
    change: str, change_dir: Path, root: Path, head: str, evidence: Evidence,
) -> tuple[str, str]:
    """`--record-landing` 이 **걷기 전에** 멈추는 사유. `(사유, base)` — 멈추면 base 는 빈칸이다.

    역사의 끝(`head`)과 증거(`evidence`)는 호출자가 **한 번** 푼 · 읽은 것이다 (task 7.5.2.1) — 기록
    명령에서는 걷기 전 판정과 걷기가 같은 값을 쓴다.

    이 판정은 **한 곳에** 산다 (task 7.7, 리뷰 I8). 기록 명령이 거절 조건을 들고 있는
    동안 5단계의 조언 줄은 그중 아무것도 묻지 않고 그 명령을 권했다. 그래서 기록이
    디스크에만 있는 change 에서 게이트는 "기록이 없다, 기록하라"고 하고 명령은 "이미
    있다"고 했다 — 두 문장이 서로를 가리키며 돈다. 리뷰가 그런 모양을 넷 셌고 수리 전에
    다시 재니 일곱이었다(디스크에만 있는 기록 · staged 아카이브 이동 · 빌리는 쪽 · 번들 0 ·
    더러운 트리 · 커밋 전 번들 · 끊긴 심링크 기록).

    **순서는 영구적인 사유가 먼저다.** 기록 명령이 쓰던 순서를 그대로 두되 하나만 옮겼다 —
    추적 파일 수정(커밋하면 사라진다)을 걷기 전 하한 사유 뒤로. 조언 줄은 사유 **하나**만
    말하므로 순서가 곧 조언이다. 순서마다 시험이 그 조합의 문장을 못 박는다
    ([[a-new-guard-unpins-the-guards-behind-it]]).

    **걷는 것은 여기 없다.** 후보 순회는 change 하나에 133초까지 걸리고(리뷰 I1) 조언 줄은
    워킹트리가 대상인 모든 실행에서 나간다. 그래서 걷고 나서야 아는 거절(어느 커밋도 번들과
    안 맞음)은 조언이 예측하지 않고, 대신 약속하지도 않는 문장으로 권한다.
    """
    landing_file = change_dir / LANDING_FILE
    if landing_file.is_symlink():
        # `exists()` 는 링크를 **따라가서** 답한다. 끊긴 링크면 거짓이므로 존재 확인을
        # 그대로 통과하고, 그 뒤의 쓰기가 링크를 따라 저장소 **밖에** 기록을 만든다
        # (2026-09-13 실측: rc 0 으로 "recorded" 라고 말하면서 임시 디렉터리에 썼다).
        # 기록은 커밋되는 파일이어야 하므로 링크 자리는 기록 자리가 아니다.
        return (
            f"`{LANDING_FILE}` is a symlink — not followed: the record must be a "
            "regular file in the change directory, or the value the gate reads back is not "
            "the value it wrote"
        ), ""
    if _landing_record(change_dir, root, head) is not None:
        # 복구 경로를 **여기서도** 말한다 (task 7.2.6). 번들을 갱신한 저자가 5단계에서 빨간
        # 문장을 받고 이 명령을 부르면, 예전에는 "이미 있다"만 듣고 두 문장 사이에 갇혔다.
        return f"`{LANDING_FILE}` already exists — not overwritten; {LANDING_RECOVERY}", ""
    if landing_file.exists():
        # HEAD 에 없는데 디스크에 있다 — 커밋 전이거나 아카이브 이동이 아직 staged 다.
        # "이미 있다"만 말하면 게이트가 "기록이 없다"고 말하는 것과 고리를 이룬다. 둘이
        # 갈리는 이유(게이트는 커밋에서 읽는다)가 곧 할 일이다. 경로를 적는 이유: staged
        # 아카이브 이동에서는 HEAD 의 **옛 자리**에 기록이 있어서 "HEAD 에 없다"만으로는 틀린다.
        relative = landing_file.relative_to(root).as_posix()
        return (
            f"`{relative}` already exists on disk but not in HEAD — not overwritten: "
            "the gate reads the record from the commit, so commit it"
        ), ""
    if (change_dir / "analysis" / "function-logic-reference.txt").exists():
        # 5단계와 같은 문장이다 (task 7.2.3). 예전 문장은 "빌려주는 쪽의 착지를 복사하라"였고
        # 그 복사가 리뷰 C4 의 구멍이었다.
        return BORROWED_REFUSES_A_LANDING, ""
    facts: dict[str, object] = {}
    try:
        base = resolve_base(change_dir, root, facts, change_id=change)
    except GATE_FAULTS as exc:
        return f"cannot resolve the comparison base: {exc}", ""
    if facts.get("execution_baseline_adoption"):
        return ADOPTION_REFUSES_A_LANDING, ""
    _, why = _walk_floor(root, _select_pinning(root, evidence), head)
    if why:
        return f"no landing recorded — {why}", ""
    # 추적 파일 수정은 **맨 뒤**다 — 커밋하면 사라지는 유일한 사유라서다. 앞에 두면 영원히
    # 기록할 수 없는 change(번들 0 · 빌리는 쪽)가 "먼저 커밋하라"를 듣고, 커밋한 뒤에야
    # 진짜 사유를 듣는다. 이 저장소의 활성 change 일곱이 번들 0 이고, tasks.md 한 줄만
    # 고쳐도 트리는 dirty 다 (task 7.7 — 변이 R10 이 이 순서를 재는 시험이 0 임을 보였다).
    dirty = subprocess.run(
        ["git", "diff", "--quiet", head], cwd=root, capture_output=True,
        timeout=30, check=False,
    )
    if dirty.returncode not in (0, 1):
        # 0(같다) · 1(다르다) 말고는 **답이 아니다** (task 7.5.2.1). rc 128 을 "바뀌었다" 로 읽던 판본은
        # 커밋할 것이 없는 저자에게 "먼저 커밋하라" 고 했다.
        raise RuntimeError(
            f"cannot tell whether the working tree matches {head[:12]}: "
            + _first_line(dirty.stderr, "git diff failed")
        )
    if dirty.returncode:
        return (
            "the working tree has uncommitted changes to tracked files — commit "
            "them first, because a recorded landing points at a commit and step 5 would "
            "then never compare those edits"
        ), ""
    return "", base


def _recording_moved(
    change: str, change_dir: Path, root: Path, head: str, book: ReadLedger,
) -> str:
    """쓰기 **직전**: 이 기록을 만들어도 되는가. 안 되면 그 사유, 되면 빈 글자 (task 7.5.2.3).

    7.5.2.2 는 역사와 증거만 다시 봤다. 그런데 거절 사유 일곱 중 **더러운 워킹트리**는 둘 중 어느 것도
    아니다(기록 명령은 Go 소스를 읽지 않으므로 원장에도 없다): 후보 순회는 change 하나에 133~219초이고
    (리뷰 I1 실측) 그 사이에 추적 Go 파일을 고치면 기록이 `open("xb")` 로 **영구히** 만들어져 그 편집이 든
    분기가 창 밖에 남는다 — 보수는 사람 손이다(재리뷰 보안 · 적대).

    **순서가 곧 사유다.** 역사를 먼저 묻는다: 실행 중 커밋이 서면 워킹트리는 그 커밋과 견주어 "더럽다" 로도
    보이는데, 그때 할 말은 "커밋하라" 가 아니라 "역사가 움직였다" 다. 그다음 거절 집합, 마지막으로 원장
    전체 — 뒤로 갈수록 넓게 본다.
    """
    moved = _head_moved(root, head)
    if moved:
        return moved
    # 거절은 **쓰는 순간의** 디스크에 대한 질문이므로 증거를 다시 읽어서 묻는다. 두 읽기가 갈리면 원장이 본다.
    refusal, _ = _recording_refusal(
        change, change_dir, root, head, _read_evidence(change_dir / "analysis" / "function-logic"))
    return refusal or _judged_state_moved(root, head, book)


def record_landing(change: str, root: Path = ROOT) -> tuple[int, list[str]]:
    """계산한 착지를 `landed-commit.txt` 에 쓴다. `(rc, 줄들)`.

    **거부하는 정상 입력을 먼저 적는다** [[fail-closed-must-name-what-it-rejects]]:

    - 고정 번들이 없는 change(a122 자신·a067·a113 …). spec 의 "고정할 증거가 없는
      착지 지점" 시나리오가 그 기록을 거절하므로, 쓰면 반드시 실패할 값을 쓰는 것이다.
    - 증거를 빌리는 change. spec 이 "착지는 그것을 고정하는 증거가 있는 change 에
      기록하고 빌리는 쪽은 값을 복사한다(SHALL)"로 자리를 정했다.
    - 실행 기준선 이관 change(a063). 그 경로의 창 끝은 감사된 source commit 이고
      spec 이 이 기록을 받지 않는다(SHALL NOT).
    - 기록이 이미 있는 change. 덮어쓰면 그 값이 무엇이었는지가 아무 데도 안 남는다.
    - 추적 파일이 수정된 워킹트리. 기록은 **커밋된** 지점을 가리키는데 그 상태의
      Go 편집은 어느 커밋에도 없다. 그대로 쓰면 5단계가 못 보는 편집이 생긴다.
    """
    try:
        # `check` 와 **같은** 해소다 — 못 찾거나 모호하면 해소기의 문장으로 멈춘다 (I4).
        change_dir = resolve_referenced_change(root, change)
    except ValueError as exc:
        return 1, [str(exc)]
    landing_file = change_dir / LANDING_FILE
    analysis = change_dir / "analysis" / "function-logic"
    try:
        with _ledger() as book:
            # 역사는 **한 번** 풀고 증거는 **한 번** 읽는다 (task 7.5.2.1) — 걷기 전 판정과 걷기가 같은
            # 값을 쓴다. 이 뒤에 커밋이 서도 기록은 시작할 때의 역사에서 계산한 값이다: 게이트가 기록을
            # 그때의 역사에서 **다시 판정한다**(기록 뒤의 자기 수리면 거절한다).
            head = _head_commit(root)
            evidence = _read_evidence(analysis)
            # 걷기 전에 멈추는 사유는 5단계의 조언 줄이 묻는 **그 함수**에 묻는다 (task 7.7, I8).
            # 경계 **안에서** 묻는다: 걷기 전 부분도 번들을 읽으므로 저장소 밖을 가리키는 번들이
            # 거기서 터진다 — 이 판정이 경계 밖에 있던 첫 판본을 7.4 의 시험이 잡았다.
            refusal, base = _recording_refusal(change, change_dir, root, head, evidence)
            if refusal:
                return 1, [f"{change}: {refusal}"]
            landing, why = compute_landing(root, base, _measure_landing_inputs(root, head, evidence))
            # 쓰기 **직전**에 거절 집합 **전체**를 다시 묻는다 (task 7.5.2.3).
            moved = _recording_moved(change, change_dir, root, head, book) if landing else ""
    except GATE_FAULTS as exc:
        # 저장소 밖을 가리키는 번들·읽을 수 없는 ast.json·멎은 git 은 전부 이 함수가
        # 답할 수 있는 것이다. 스택으로 죽으면 "왜 기록이 안 됐나"가 아무 데도 안 남는다.
        return 1, [f"{change}: no landing recorded — {exc}"]
    if not landing:
        return 1, [f"{change}: no landing recorded — {why}"]
    if moved:
        return 1, [f"{change}: no landing recorded — {moved}"]
    try:
        with landing_file.open("xb") as output:
            output.write((landing + "\n").encode("utf-8"))
    except FileExistsError:
        # 존재 확인과 이 쓰기 사이에 walk 하나가 통째로 들어간다(리뷰 I1 실측 133.7s·219.6s).
        # 그 사이에 생긴 기록을 덮으면 그것이 무엇이었는지가 아무 데도 안 남는다. `x` 는
        # 심링크에도 걸리므로 위의 거절이 뚫려도 저장소 밖으로는 못 쓴다 — 확인이 아니라
        # **쓰기 자체**가 한 번만 만들어지는 것이 이 보장의 자리다.
        return 1, [
            f"{change}: `{LANDING_FILE}` appeared while the landing was being computed "
            "— not overwritten"
        ]
    return 0, [
        f"{change}: recorded `{LANDING_FILE}` = {landing} (base {base[:12]}) — computed "
        "from this change's own evidence, not chosen; commit it with the change"
    ]


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--change", required=True)
    parser.add_argument("--root", default=str(ROOT))
    parser.add_argument(
        "--record-landing", action="store_true",
        help="이 change 의 증거로 착지 지점을 계산해 기록한다 (저자가 고르지 않는다)",
    )
    args = parser.parse_args()
    reconfigure = getattr(sys.stdout, "reconfigure", None)
    if reconfigure is not None:
        # 판정 줄은 경로를 이름으로 댄다. JSON 의 `\udcff` 같은 홀로 선 서로게이트는 합법이고, 출력이
        # 엄격한 UTF-8 이면(`en_US.UTF-8` 의 기본) `print` 가 `UnicodeEncodeError` 로 **판정 줄 없이** 끝났다 —
        # 결과가 로캘의 함수였다 (task 7.5.2.1, 재리뷰 레드팀). 못 쓰는 글자는 `\udcff` 로 적는다.
        reconfigure(errors="backslashreplace")
    context: dict[str, object] = {}
    root = Path(args.root)
    if args.record_landing:
        try:
            code, lines = record_landing(args.change, root)
        except GATE_FAULTS as exc:
            code, lines = 1, [f"{args.change}: no landing recorded — {exc}"]
        for line in lines:
            print(f"[logic-map] {line}")
        return code
    # 판정의 **경계**가 여기다 (task 7.4, H6). 안쪽 자리마다 목록을 베끼면 새로 부르는
    # git 하나가 그중 아무 목록에도 안 걸려 다시 스택이 된다 — 이 도구가 실제로 불리는
    # 곳은 `tools/gate.sh` 의 CLI 하나뿐이므로 그 하나에 세운다. 창 줄은 아래에서 계속
    # 찍힌다: `context` 는 참조로 채워지므로 착지를 **잰 뒤에** 터진 결함이면 그 창이 남아 있다.
    try:
        errors = check(args.change, root, context)
    except GATE_FAULTS as exc:
        errors = [f"cannot judge this change: {exc}"]
    # 창을 **먼저** 찍는다. 그리고 실패해도 찍는다 — 옛 판본은 `if errors: return 1`
    # 이 이 자리를 건너뛰어서, 요구된 함수 이름 316개가 어느 두 지점 사이에서 나온
    # 것인지 출력 어디에도 없었다(2026-09-10 a074·a076 실측: 그런 줄 0개).
    base = str(context.get("effective_base", ""))
    # 창은 **잰 것만** 찍는다. 착지 해소가 실패하면 `landing` 과 `required_count` 는
    # 아예 안 채워지고, 그 빈칸을 기본값으로 찍으면 `working tree … required 0` 이라는
    # 거짓말이 된다 — 대상도 아니고 세지도 않은 값이다. 3.3 이 창을 찍게 만든 이유가
    # "이름만 있고 이유가 없다"였는데 지어낸 이유는 그보다 나쁘다. 그 경우 사유는
    # 아래 `cannot derive modified Go functions: …` 줄이 말한다.
    if base and "landing" in context:
        landing = str(context.get("landing", ""))
        audited = bool(context.get("execution_baseline_adoption"))
        # 창 줄 · 조언은 판정이 푼 **그** 역사를 말한다 (task 7.5.2.1). `landing` 이 채워졌으면 `head` 도
        # 채워졌다 — `check` 가 둘보다 먼저 푼다. 그 sha 를 창 줄 끝에 적는다 (task 7.5.2.2, 재리뷰 적대): 게이트의
        # PASS 가 **어느 역사의** 판정인지 출력에 남아야 한다. 맨 뒤에 붙인다 — 앞의 문구를 읽는 사람 · 시험이 그대로다.
        head = str(context["head"])
        print(
            f"[logic-map] {args.change}: base {base[:12]} → {_target_text(landing, audited)} "
            f"required {context.get('required_count', 0)} function(s) — judged at HEAD {head[:12]}"
        )
        if not landing:
            landed_after = _commits_after(root, base, head)
            # 창의 크기는 **사실**이라 두 갈래가 공유한다. 갈리는 것은 조언뿐이다.
            window = (
                f"[logic-map] {args.change}: the target is the working tree, so this window "
                f"also holds {landed_after or '?'} commit(s) that landed after the base and "
                f"every existing function they changed is required here too"
            )
            # 번들이 base 의 소스를 적고 있으면 `--record-landing` 을 **권하지 않는다**
            # (task 7.1). 권하면 게이트가 자기 조언 줄로 자기 판정을 지운다: 그 명령이
            # 계산할 수 있는 착지는 전부 그 함수가 아직 base 와 같은 지점이다.
            base_shaped = [str(name) for name in (context.get("base_shaped_bundles") or [])]
            fault = context.get("base_shaped_fault")
            if fault:
                # 그것을 **못 쟀으면** 권하지도 거절하지도 않는다 — 모르는 채로 명령을 권하면
                # 위의 이유가 그대로 되살아난다 (task 7.5.2, 재리뷰 적대 F6).
                print(f"{window} — cannot tell whether `--record-landing` could narrow it: {fault}")
            elif base_shaped:
                # 이름은 세 개까지만 적고 나머지는 **세어서** 말한다. a076 의 21,838자
                # 한 줄이 이 저장소가 이름을 다 쏟아내지 않는 이유다.
                named = ", ".join(base_shaped[:3])
                if len(base_shaped) > 3:
                    named += f", and {len(base_shaped) - 3} more"
                print(
                    f"{window} — but {len(base_shaped)} of this change's `revision: current` "
                    f"bundle(s) ({named}) record the source as it stood at the base, not at "
                    f"HEAD, so every landing the gate could compute is one where those "
                    f"functions still equal the base and the narrowed window would require "
                    f"none of them; refresh those bundles against the current source instead"
                )
            else:
                # 명령을 권하기 전에 **그 명령이 묻는 함수**에 묻는다 (task 7.7, 리뷰 I8).
                # 옛 판본은 아무것도 안 묻고 권해서, 명령이 거절하는 일곱 모양에서 두 문장이
                # 서로 모순됐다. 조언은 판정이 아니므로 여기서 난 결함이 판정 줄을 바꾸거나
                # 모르는 채로 명령을 권하게 두지 않는다.
                try:
                    # 증거는 여기서 다시 읽는다 — 이 줄은 **다음에 부를 명령**을 예측하고, 그 명령은 부르는
                    # 때에 스스로 읽는다. 판정 줄은 이 읽기에 기대지 않는다.
                    change_dir = resolve_referenced_change(root, args.change)
                    refusal, _ = _recording_refusal(
                        args.change, change_dir, root, head,
                        _read_evidence(change_dir / "analysis" / "function-logic"),
                    )
                except GATE_FAULTS as exc:
                    refusal = f"cannot tell whether it would record: {exc}"
                if refusal:
                    print(f"{window} — `--record-landing` cannot narrow it: {refusal}")
                else:
                    # 걷고 나서야 아는 거절은 예측하지 않는다 — 그래서 "기록된다"고 약속하지 않는다.
                    print(
                        f"{window} — run "
                        f"`python3 tools/logic-map/check_analysis.py --change {args.change} "
                        f"--record-landing` to let the gate compute `{LANDING_FILE}` from this "
                        f"change's evidence and record it, which narrows it to this change's own "
                        f"work; if no commit on this history is accepted as the landing, the "
                        f"command says so instead of recording"
                    )
    if errors:
        for error in errors:
            print(f"[logic-map] {error}")
        return 1
    if context.get("execution_baseline_adoption"):
        print(f"[logic-map] {args.change}: execution-baseline adoption exception evidence complete")
    else:
        print(f"[logic-map] {args.change}: evidence complete or diff-proven exempt")
    # 어떤 대상으로 몇 개를 요구해서 통과했는지는 위의 창 줄이 말한다 — 성공·실패
    # 양쪽에서 같은 한 줄이다. 두 줄로 나누면 실패 경로만 조용해진다(task 3.3).
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
