#!/usr/bin/env python3
"""Bind Function Logic Map evidence to every modified existing Go function."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import tempfile
from pathlib import Path

from role_check import call_enumeration_in_use, role_errors
from execution_baseline import AdoptionError, validate as validate_execution_baseline

ROOT = Path(__file__).resolve().parents[2]
REQUIRED = (
    "ast.json",
    "function-logic-map.md",
    "branch-test-map.md",
    "risk-pattern-report.md",
)
EXEMPTION = "Function Logic Map: not-applicable"
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


def _safe_changed_go_paths(root: Path, base: str, target: str) -> None:
    """Reject names that the unified-diff header grammar cannot represent losslessly."""
    process = subprocess.run(
        ["git", "diff", "--no-ext-diff", "--name-only", "-z", base,
         *([target] if target else []), "--", "*.go"],
        cwd=root,
        capture_output=True,
        check=False,
    )
    if process.returncode:
        stderr = process.stderr.decode("utf-8", "replace") if isinstance(process.stderr, bytes) else process.stderr
        raise RuntimeError(stderr.strip() or f"git diff failed for base {base}")
    raw_output = process.stdout if isinstance(process.stdout, bytes) else process.stdout.encode("utf-8")
    for raw in (item for item in raw_output.split(b"\0") if item):
        try:
            path = raw.decode("utf-8", "strict")
        except UnicodeDecodeError as error:
            raise RuntimeError("modified Go path is not UTF-8") from error
        # Git quotes tab/newline headers; do not silently parse that quoted form
        # as a different path. Ordinary Unicode names remain supported.
        if "\n" in path or "\r" in path or "\t" in path:
            raise RuntimeError("modified Go path cannot be represented losslessly in unified diff")


def changed_existing_functions(
    root: Path = ROOT,
    base: str = "",
    target: str = "",
) -> dict[tuple[str, str], dict]:
    if not base:
        raise ValueError("Function Logic Map comparison base is required")
    _safe_changed_go_paths(root, base, target)
    process = subprocess.run(
        [
            "git",
            "-c",
            "core.quotePath=false",
            "diff",
            "--no-ext-diff",
            "--find-renames",
            "--unified=0",
            base,
            *( [target] if target else [] ),
            "--",
            "*.go",
        ],
        cwd=root,
        capture_output=True,
        text=True,
        timeout=30,
        check=False,
    )
    if process.returncode:
        raise RuntimeError(process.stderr.strip() or f"git diff failed for base {base}")
    required: dict[tuple[str, str], dict] = {}
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
            current = base_file(root, target, new_source) if target and new_source else (root / new_source if new_source else None)
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
            if target and current is not None:
                current.unlink(missing_ok=True)
        hunks = []

    for line in process.stdout.splitlines():
        if line.startswith("diff --git "):
            flush()
            old_source = ""
            new_source = ""
        elif line.startswith("--- "):
            value = line[4:]
            old_source = "" if value == "/dev/null" else value.removeprefix("a/")
        elif line.startswith("+++ "):
            value = line[4:]
            new_source = "" if value == "/dev/null" else value.removeprefix("b/")
        else:
            match = HUNK.match(line)
            if match:
                hunks.append(
                    (
                        int(match.group(1)),
                        int(match.group(2) or 1),
                        int(match.group(3)),
                        int(match.group(4) or 1),
                    )
                )
    flush()
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
    changes = root / "openspec" / "changes"
    direct = changes / change
    open_here = direct.is_dir()
    archive = changes / "archive"
    archived = sorted(
        path
        for path in (archive.iterdir() if archive.is_dir() else ())
        if path.is_dir() and _archived_change_id(path.name) == change
    )
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
        candidate = path.read_text(encoding="utf-8").strip()
    except OSError as exc:
        raise ValueError(
            "missing base-commit.txt; run "
            "`python3 tools/sdd/capture_change_base.py --change <change-id>` "
            "at proposal freeze"
        ) from exc

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
# 이보다 낮은 git 에서는 그 옵션이 거절되어 blob 읽기가 전부 실패하고, 게이트는 막히는
# 쪽으로 틀리지만(fail-closed 실측: required 32 → 269, rc 120) **사유를 저자의 증거 탓으로
# 오진한다**. 값을 여기 한 곳에 적어 두고 README·WORKFLOW 가 같은 수를 인용한다.
GIT_BATCH_MINIMUM = "2.42"
LANDING_FILE = "landed-commit.txt"
FULL_SHA = re.compile(r"[0-9a-f]{40}")
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
    안에 있는 개행이 응답을 쪼갠다. `-Z` 는 양쪽을 NUL 로 끊는다(git 2.34+ 가 `-z`,
    출력까지는 2.42+ 의 `-Z`; 이 저장소는 2.43.0).

    **blob 만 내용으로 친다.** `git show <ref>:<디렉터리>` 는 트리 **목록**을 찍는데, 그
    바이트가 판정에 들어오면 파일 내용인 척한다. 여기서는 `None` 이다 — 막는 쪽으로 엄해진다.
    오늘 실물 입력(번들의 `file` · 번들 `ast.json` 경로)에 디렉터리는 0 건이다
    ([[fail-closed-must-name-what-it-rejects]]).

    프로세스가 실패하면 전부 `None` 이다 — 옛 `git show` 의 rc≠0 과 **같은 방향**이고,
    부르는 쪽은 `None` 을 불일치로 세므로 판정이 느슨해지지 않는다.
    """
    wanted = list(dict.fromkeys(relatives))     # 순서 유지 + 중복 제거(같은 소스를 적은 번들 여럿)
    found: dict[str, bytes | None] = {relative: None for relative in wanted}
    if not wanted:
        return found
    for relative in wanted:
        if "\0" in relative:
            # 요청 자체가 NUL 로 끊기므로 경로 안의 NUL 은 레코드를 쪼개고, 그 뒤 자리들이
            # **다른 파일의 바이트**를 받는다 (독립 리뷰 2026-09-18). 물어보지 않는다.
            raise RuntimeError(f"path contains a NUL byte, cannot be asked for: {relative!r}")
    process = subprocess.run(
        ["git", "cat-file", "--batch", "-Z"],
        cwd=root,
        # `surrogateescape` 인 이유: 파일 이름은 바이트다. 옛 판본은 경로를 argv 로 넘겨
        # `os.fsencode` 를 탔으므로 디코딩 불가능한 이름도 그대로 갔다. 엄격한 `utf-8` 로
        # 인코딩하면 그런 이름 앞에서 **판정 대신 traceback** 이 된다.
        input=b"".join(f"{ref}:{relative}\0".encode("utf-8", "surrogateescape")
                       for relative in wanted),
        capture_output=True, timeout=60, check=False,
    )
    if process.returncode:
        # 전부 `None` 이다 — 옛 `git show` 의 rc≠0 과 **같은 방향**이고, 부르는 쪽은 `None` 을
        # 불일치로 세므로 판정이 느슨해지지 않는다.
        #
        # **여기를 결함으로 올리려다 되돌렸다** (독립 리뷰 2026-09-18, P1). 올리면 `-Z` 를
        # 모르는 git(2.42 미만) 아래의 오진("저자의 증거가 낡았다")이 사라지지만, 거부할
        # 정상 입력을 세어 보니 **저장소가 아닌 루트**가 이 함수의 정상 호출 모양이었다 —
        # 번들 검증만 보는 시험 21개가 임시 디렉터리에서 돈다
        # ([[fail-closed-must-name-what-it-rejects]]: 열거가 설계를 죽였다).
        # 최소 git 버전은 `GIT_BATCH_MINIMUM` 과 README·WORKFLOW 에 적어 둔다.
        return found
    data, position = process.stdout, 0
    answered: dict[str, bytes] = {}
    for relative in wanted:
        end = data.find(b"\0", position)
        if end < 0:
            # **부분 답을 쓰지 않는다** (독립 리뷰 2026-09-18, P0). 예전에는 여기서 `break`
            # 했고 "나머지는 `None` 이라 답이 같다"고 적었는데 **거짓이었다**: 마지막 머리가
            # 멀쩡하고 내용만 잘린 응답에서는 그 자리가 `None` 이 아니라 `b""` 가 되고,
            # `None` 과 `b""` 는 `_unheld_bundles` 의 아카이브 대체 갈래를 여닫아 판정을 바꾼다.
            # 프레이밍을 끝까지 못 읽으면 그것은 결함이다.
            raise RuntimeError(
                f"`git cat-file --batch -Z` answered {len(answered)} of {len(wanted)} "
                f"request(s) at {ref[:12]}: the response is truncated"
            )
        header = data[position:end]
        position = end + 1
        fields = header.rsplit(b" ", 2)
        # 찾은 객체의 머리는 `<oid> <type> <size>` 이고 내용이 뒤따른다. `missing`·
        # `ambiguous` 는 크기가 없으므로 내용도 없다 — 크기 자리가 숫자인지로 가른다.
        if len(fields) != 3 or not fields[2].isdigit():
            continue
        size = int(fields[2])
        if fields[1] == b"blob":
            answered[relative] = data[position:position + size]
        position += size + 1                    # 내용 뒤의 NUL 하나
    if position != len(data):
        # 정상 응답은 **정확히** 소진된다(실측). 남은 바이트는 프레이밍을 잘못 읽었다는 뜻이고,
        # 잘못 읽은 프레이밍은 자리를 밀어 다른 파일의 바이트를 답에 넣는다.
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


def _is_ancestor(root: Path, older: str, newer: str) -> bool:
    return subprocess.run(
        ["git", "merge-base", "--is-ancestor", older, newer],
        cwd=root, capture_output=True, timeout=10, check=False,
    ).returncode == 0


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


def _commits_after(root: Path, base: str) -> str:
    """base 뒤에 착지한 커밋 수. 숫자를 **재서** 쓴다 — 못 재면 빈 문자열이다."""
    process = subprocess.run(
        ["git", "rev-list", "--count", f"{base}..HEAD"],
        cwd=root, capture_output=True, text=True, timeout=10, check=False,
    )
    return "" if process.returncode else process.stdout.strip()


def _landing_record(change_dir: Path, root: Path) -> tuple[str, bytes] | None:
    """HEAD 커밋에 착지 기록이 **있는가**. 있으면 `(경로, 바이트)`, 없으면 `None`.

    해독하지 않는다. "기록이 있는가"와 "기록에 무엇이 적혔나"는 다른 질문이고, 앞의
    것에 해독하는 함수를 부르면 못 읽는 기록 앞에서 질문 자체가 터진다 — 이관 경로의
    probe 와 `record_landing` 의 덮어쓰기 거절이 그랬다(task 6.2.1). 못 읽는 기록도
    기록이다.
    """
    try:
        relative = (change_dir / LANDING_FILE).relative_to(root).as_posix()
    except ValueError:
        return None
    raw = _committed_bytes(root, "HEAD", relative)
    return None if raw is None else (relative, raw)


def _declared_landing(change_dir: Path, root: Path) -> str | None:
    """HEAD 커밋에 적힌 착지 선언. 선언 자체가 없으면 `None` 이다.

    값을 **판정하지 않는다** — 40자리인지, 커밋인지, 조상인지는 `resolve_landing`
    이 묻는다. 여기가 답하는 것은 "선언이 있는가, 있다면 무엇이라고 적혀 있는가"
    하나뿐이다. 빈 파일은 `""` 이고 `None` 과 **다르다**: 빈 선언도 선언이므로
    없는 것으로 읽으면 그 change 가 조용히 워킹트리를 대상으로 삼게 된다.
    """
    record = _landing_record(change_dir, root)
    if record is None:
        return None
    relative, raw = record
    try:
        return raw.decode("utf-8").strip()
    except UnicodeDecodeError as exc:
        raise ValueError(f"landing point is not UTF-8: {relative}") from exc


def _pinning_bundles(root: Path, analysis: Path) -> list[tuple[Path, str, str]]:
    """착지를 고정하는 번들 = `revision: current` 이고 `file`·`source_sha256` 둘 다 있는 것.

    이 선별은 **한 곳에만** 산다. `resolve_landing` 과 `compute_landing` 이 각자
    순회를 가지면 한쪽만 고쳐도 양쪽 시험이 초록이 된다
    ([[two-judgements-cover-for-each-other]] 가 a064 에서 실측한 모양).
    """
    found: list[tuple[Path, str, str]] = []
    for ast_path in sorted(analysis.glob("*/ast.json")) if analysis.is_dir() else ():
        value = _ast_value(ast_path)
        if not isinstance(value, dict) or value.get("revision", "current") != "current":
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


def _base_shaped_bundles(root: Path, base: str, analysis: Path) -> list[str]:
    """`revision: current` 인데 **base 의 소스**를 적은 고정 번들의 이름들.

    이 모양이 §6 독립 리뷰의 V1 이다. 저장소 규칙대로 FLM 을 **먼저** 커밋하고 코드를
    편집한 뒤 번들을 갱신하지 않으면, 번들은 base 의 소스를 적은 채로 남는다.

    이것을 세는 이유는 **조언 한 줄** 때문이다. 그런 번들이 있으면 도구가 받아들일 수
    있는 착지는 전부 그 파일이 아직 base 와 같은 지점이다. 그러므로 거기서 좁힌 창은
    그 함수를 요구할 수 없고, `--record-landing` 을 권하면 게이트가 자기 조언 줄로
    자기 판정을 지우는 길을 가리키게 된다.

    판정은 신원이 아니라 **blob 등식**으로 한다 — 누구의 편집인지는 묻지 않는다
    (1.12 가 신원 판정을 이미 배제했다).
    """
    names: list[str] = []
    for ast_path, source, digest in _pinning_bundles(root, analysis):
        path = root / source
        if path.is_file() and hashlib.sha256(path.read_bytes()).hexdigest() == digest:
            continue  # 오늘의 소스를 적은 번들 — 세탁할 것이 없다
        at_base = _committed_bytes(root, base, source)
        if at_base is not None and hashlib.sha256(at_base).hexdigest() == digest:
            names.append(ast_path.parent.name)
    return sorted(names)


def _pinning_at(
    root: Path, candidate: str, bundles: list[tuple[Path, str, str]],
) -> tuple[int, list[str]]:
    """`candidate` 에서 고정 번들이 몇 개이고 그중 어느 소스가 안 맞는가.

    번들 목록은 **호출자가 한 번 재서 넘긴다** (task 7.5) — `floor` · `repairs` 와 같다.
    디렉터리를 받아 스스로 순회하던 판본은 후보마다 glob + JSON 파싱 + `Path.resolve()`
    를 다시 했다(2026-09-18 프로파일: a071 의 walk 하나에서 `_pinning_bundles` 가 341회 ·
    13.07s 중 9.30s). 번들은 걷는 동안 안 변한다 — 도구가 쓰지 않기 때문이다.
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


def _evidence_floor(root: Path, bundles: list[tuple[Path, str, str]]) -> str:
    """고정 번들이 역사에 들어온 **마지막** 커밋. 못 찾으면 빈 문자열이다.

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
         "HEAD", "--", *paths],
        cwd=root, capture_output=True, text=True, timeout=60, check=False,
    )
    return "" if process.returncode else process.stdout.strip()


def _unheld_bundles(
    root: Path, candidate: str, bundles: list[tuple[Path, str, str]],
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
        try:
            judged = ast_path.read_bytes()  # 판정과 **같은 읽기** — 심링크면 따라간다
        except OSError:
            unheld.append(ast_path.parent.name)
            continue
        committed = blobs[relative]
        if committed is None and before:
            committed = blobs[before]
        if committed is None or committed != judged:
            unheld.append(ast_path.parent.name)
    return sorted(set(unheld))


def _self_repair_commits(root: Path, analysis: Path) -> list[str]:
    """이 change 자신의 **나중 Go 작업** 커밋들. 오래된 것부터. 없으면 빈 목록이다.

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
        ["git", "log", "--full-history", "--no-merges", "--format=%H", "HEAD", "--", *paths],
        cwd=root, capture_output=True, text=True, timeout=120, check=False,
    )
    if touching.returncode:
        # **실패는 거절이 아니라 판정이다.** 빈 목록으로 물러나면 git 이 멎은 것과 "만진 커밋이
        # 없다"가 같은 말이 되고, 가드가 조용히 꺼진다(적대 리뷰 F2 — 주입 실험에서 rc 1 하나로
        # 거절돼야 할 입력이 초록이었다). 옆의 `_evidence_floor` 는 실패하면 거절로 간다.
        raise RuntimeError(
            f"cannot list the commits that touched {change_dir}: "
            f"{touching.stderr.strip().splitlines()[0] if touching.stderr.strip() else 'git log failed'}"
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
            f"{listing.stderr.strip().splitlines()[0] if listing.stderr.strip() else 'git log --no-walk failed'}"
        )
    flagged: set[str] = set()
    for block in listing.stdout.split("\0"):
        lines = [line for line in block.splitlines() if line]
        if lines and any(name.endswith(".go") for name in lines[1:]):
            flagged.add(lines[0])
    # `git log` 는 새 것부터 준다. 거절 문장이 **가장 오래된** 것을 이름으로 대야 저자가
    # 고칠 첫 자리를 가리킨다 — 뒤의 것들은 그 뒤에 쌓인 작업이다.
    return [commit for commit in reversed(hashes) if commit in flagged]


def _repairs_after(root: Path, candidate: str, repairs: list[str]) -> list[str]:
    """`candidate` **뒤에** 서는 수리 커밋들. 순서는 `repairs` 의 순서(오래된 것부터).

    `candidate..HEAD` 한 번으로 묻는다 — 후보마다 `merge-base` 를 수리 개수만큼 돌면
    a112 처럼 수리가 스물여섯인 change 에서 후보 하나에 프로세스가 스물여섯이다. 집합은
    같다: `rev-list A..HEAD` 가 곧 "HEAD 에서 닿고 A 의 조상이 아닌" 커밋이고, 후보 자신은
    자기 조상이므로 빠진다(그래서 수리 커밋 **자신**은 착지가 될 수 있다 — 복구 경로).
    """
    if not repairs:
        return []
    process = subprocess.run(
        ["git", "rev-list", f"{candidate}..HEAD"],
        cwd=root, capture_output=True, text=True, timeout=120, check=False,
    )
    if process.returncode:
        # 실패는 판정이다 — 빈 목록이면 가드가 조용히 꺼진다 (적대 리뷰 F2).
        raise RuntimeError(f"cannot walk the history after {candidate[:12]}")
    after = set(process.stdout.split())
    return [commit for commit in repairs if commit in after]


def _landing_refusal(
    root: Path, base: str, candidate: str, bundles: list[tuple[Path, str, str]], floor: str,
    repairs: list[str],
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
    느린 아카이브 change 일곱의 몫이고 활성 change 에서는 둘뿐이다. 그 walk 는 7.5 가 맡는다.

    `floor` 와 `repairs` 는 호출자가 한 번 재서 넘긴다 — 후보마다 `git log` 를 다시 돌리지
    않는다. 둘째 값은 판정이 읽은 번들을 그 커밋이 **안 들고 있을 때만** 그 번들 이름이다.
    계산 경로가 "왜 못 찾았나"를 그 이름으로 말한다.
    """
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
    unheld = _unheld_bundles(root, candidate, bundles)
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
    if all(
        _committed_bytes(root, base, source) == _committed_bytes(root, candidate, source)
        for source in sources
    ):
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
    later = _repairs_after(root, candidate, repairs)
    if later:
        return (
            f"landing point {candidate[:12]} is followed by {len(later)} later commit(s) of "
            f"this change's own Go work (first {later[0][:12]}): a non-merge commit that edits "
            "Go while touching this change's directory landed after it, and the window "
            f"{base[:12]}..{candidate[:12]} would not compare that edit"
        ), []
    return "", []


def resolve_landing(change_dir: Path, root: Path, base: str, analysis: Path) -> str:
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
    돌고 저자가 구간의 바닥을 고를 수 있다. `analysis` 는 그래서 이 change 가 실제로
    딛는 증거 디렉터리여야 한다 — 빌린 증거를 쓰는 change 는 지역 번들이 0 이므로
    호출자가 **빌린 쪽을 먼저 풀어서** 넘긴다.

    마지막 판정은 "유효한가"가 아니라 "**그 값인가**"다 (task 7.3). 유효 조건을 통과하는
    값은 여럿이므로(실측: 착지를 얻는 76건 중 65건) 그것만으로는 저자의 선택이 안 없어진다.
    기록은 `compute_landing` 이 내는 값과 같아야 한다.
    """
    candidate = _declared_landing(change_dir, root)
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
    if not _is_ancestor(root, candidate, "HEAD"):
        raise ValueError(f"landing point never landed on this history: {candidate}")
    # 유효한가는 **한 함수**가 판정한다 — 아래 `compute_landing` 이 후보마다 묻는 것과
    # 같은 함수다 (task 7.6, 리뷰 I2).
    # 번들 목록은 **한 번** 잰다 (task 7.5) — 하한과 규칙이 같은 목록을 본다. 두 번
    # 재면 그 사이의 디스크 변화가 두 판정을 갈라 놓을 수 있다.
    bundles = _pinning_bundles(root, analysis)
    refusal, _ = _landing_refusal(
        root, base, candidate, bundles, _evidence_floor(root, bundles),
        _self_repair_commits(root, analysis),
    )
    if refusal:
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
    computed, why = compute_landing(root, base, analysis)
    if candidate != computed:
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
    anchor = root.resolve()
    raw = Path(value)
    path = raw if raw.is_absolute() else root / raw
    resolved = path.resolve()
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
        lines = path.read_text(encoding="utf-8", errors="replace").splitlines()
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
    for path in root.rglob("*_test.go"):
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
    if "/" in cited:
        qualified = root / cited
        return qualified if qualified.is_file() else None
    local = package_dir / cited
    if local.is_file():
        return local
    matches = [path for path in root.rglob(cited) if ".git" not in path.parts]
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
            total = len(path.read_text(encoding="utf-8", errors="replace").splitlines())
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


def _bundle_text(map_path):
    """번들 디렉터리에서 **읽히는 파일 전부**를 이어 붙인다.

    강제 판정은 "이 change 가 열거를 쓰는가"이고, 그 답은 열거가 번들 안 어느
    파일에 있든 같다. 파일을 **열거해서** 읽으면 그 목록 밖으로 옮기는 것으로
    판정이 꺼진다.

    다섯 라운드 동안 "남는 회피는 X 뿐"이라는 문장이 매번 코드가 실제로 보는
    범위보다 한 칸 넓었다: 절 → 표지 철자 → `##` 절 → 파일 이름 둘 →
    확장자 `.md` 하나. 마지막 것은 9차 적대 리뷰가 `notes.txt` 로 보였다.
    이제 이름으로도 확장자로도 거르지 않고 **읽히는가**로만 거른다 — 목록이
    없으므로 새 파일 이름이나 새 확장자가 이 판정을 비켜 갈 수 없다."""
    texts = []
    for path in sorted(map_path.parent.glob("*")):
        if not path.is_file():
            continue
        try:
            texts.append(path.read_text(encoding="utf-8"))
        except (OSError, UnicodeDecodeError):
            # 텍스트가 아니면 표가 들어 있을 수 없다. 건너뛰는 것과 목록을
            # 만드는 것은 다르다 — 여기서 거르는 기준은 파일 이름이 아니라
            # "읽히는가"이고, 그래서 새 확장자가 생겨도 저절로 포함된다.
            continue
    return "\n".join(texts)


def _ast_value(path):
    """ast.json 을 읽는다. 없거나 깨졌으면 판정에서 뺀다(모르는 것은 근거가 아니다)."""
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return {}


def validate_target(
    target: Path, root: Path, index: dict | None = None, require_calls: bool = False,
    revision_ref: str = "",
) -> tuple[list[str], tuple[str, str] | None]:
    """`revision_ref` 가 있으면 `revision: current` 해싱을 그 커밋에서 한다.

    비교 대상과 증거 대조 대상이 갈리면 하나는 착지를, 하나는 오늘을 기술하는 두
    정본이 된다. 병합 뒤에는 그 둘이 반드시 어긋난다."""
    errors: list[str] = []
    texts: dict[str, str] = {}
    for name in REQUIRED:
        path = target / name
        if not path.exists():
            errors.append(f"{target.name}: missing {name}")
            continue
        texts[name] = path.read_text(encoding="utf-8")
        if "TODO" in texts[name]:
            errors.append(f"{target.name}: {name} still contains TODO")
    if "ast.json" not in texts:
        return errors, None
    try:
        value = json.loads(texts["ast.json"])
    except ValueError:
        return errors + [f"{target.name}: ast.json is invalid"], None
    keys = ("file", "source_sha256", "package", "function", "signature", "start", "end")
    if not isinstance(value, dict) or any(not value.get(key) for key in keys):
        return errors + [f"{target.name}: ast.json is placeholder evidence"], None
    try:
        source, relative = normalized_source(str(value["file"]), root)
    except ValueError as exc:
        return errors + [f"{target.name}: {exc}"], None
    revision = value.get("revision", "current")
    if revision == "current":
        if revision_ref:
            blob = _committed_bytes(root, revision_ref, relative)
            if blob is None:
                errors.append(f"{target.name}: AST source is missing: {relative}")
            elif hashlib.sha256(blob).hexdigest() != value["source_sha256"]:
                errors.append(f"{target.name}: AST source hash is stale: {relative}")
        elif not source.is_file():
            errors.append(f"{target.name}: AST source is missing: {relative}")
        elif hashlib.sha256(source.read_bytes()).hexdigest() != value["source_sha256"]:
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
    # 아카이브된 change 도 자기 id 로 재검사할 수 있어야 한다. 아카이브 문법을 아는
    # 해소기가 이미 있으므로 세 번째 사본을 만들지 않는다.
    # 해소기가 못 찾은 id 는 **그 문장으로** 멈춘다 (task 7.6, 리뷰 I4). 예전에는 없는
    # `openspec/changes/<id>` 로 바꿔 넘겨서 "base 를 capture 하라"는 조언이 나갔다 —
    # 오타 난 id 에 새 change 를 만들라는 말이다. 2026-09-13 전수: 게이트가 받을 수 있는
    # id 126개 중 그 갈래로 떨어지는 것 0.
    try:
        change_dir = resolve_referenced_change(root, change)
    except ValueError as exc:
        return [str(exc)]
    analysis = change_dir / "analysis" / "function-logic"
    reference_file = change_dir / "analysis" / "function-logic-reference.txt"
    review = change_dir / "review.md"
    review_text = review.read_text(encoding="utf-8") if review.exists() else ""
    # 문맥은 **항상** 채운다. 호출자가 안 줘도 `check` 자신이 읽어야 하는 사실이 여기
    # 들어온다(이관인가, 감사된 source 는 무엇인가). 호출자가 문맥을 줬으면 같은
    # 사전이므로 밖에서 보이는 것은 그대로다.
    facts: dict[str, object] = {} if context is None else context
    try:
        base = resolve_base(change_dir, root, facts, change_id=change)
    except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
        return [f"cannot derive modified Go functions: {exc}"]
    # 빌린 증거는 착지 판정 **앞에서** 푼다. 착지가 유효한지는 그 change 가 실제로
    # 딛는 증거로 판정하는데, 빌린 change 는 지역 번들이 0 이라 뒤에서 풀면 고정할
    # 것이 하나도 없는 채로 판정이 끝난다(a073 이 그 모양이다).
    if reference_file.exists():
        if analysis.exists() and any(path.is_dir() and any(path.iterdir()) for path in analysis.iterdir()):
            return ["function-logic reference cannot coexist with local function-logic evidence"]
        referenced_change = reference_file.read_text(encoding="utf-8").strip()
        if not re.fullmatch(r"[a-z0-9][a-z0-9-]*", referenced_change) or referenced_change == change:
            return ["function-logic reference names an invalid or recursive change"]
        try:
            referenced_dir = resolve_referenced_change(root, referenced_change)
            referenced_base = resolve_base(referenced_dir, root, change_id=referenced_change)
        except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
            return [f"function-logic reference base is invalid: {exc}"]
        if referenced_base != base:
            return ["function-logic reference must share the exact comparison base"]
        # 빌리는 change 는 창을 **좁히지 않는다** (task 7.2.3, 리뷰 C4 — 사람이 2026-09-14 에
        # 골랐다). 1.8 은 빌려주는 쪽의 착지를 복사하게 했는데, 그 착지는 빌려주는 쪽 증거의
        # 가장 낮은 값이라 빌리는 쪽이 그 **뒤에** 한 Go 작업이 창 밖으로 나갔다. 빌리는 쪽은
        # 자기 번들이 0 이라 자기 작업이 어디 착지했는지 말할 증거를 소유하지 않는다.
        # 그래서 빌리는 쪽의 기록은 이름으로 거절하고, 빌려주는 쪽의 기록은 이 창에 쓰지 않는다
        # (아래 `resolve_landing` 은 빌리는 쪽 디렉터리를 보므로 기록이 없으면 빈 값 = 워킹트리).
        # 있는가만 묻는다 — 해독하면 못 읽는 기록 앞에서 질문이 터진다(6.2.1).
        # 오늘 저장소의 빌리는 change 는 a073 하나이고 기록이 없다(2026-09-14 확인).
        if _landing_record(change_dir, root) is not None:
            return [BORROWED_REFUSES_A_LANDING]
        analysis = referenced_dir / "analysis" / "function-logic"
    adopted = bool(facts.get("execution_baseline_adoption"))
    if adopted and _landing_record(change_dir, root) is not None:
        # 이관 예외의 정당성은 "판정에 들어가는 입력을 하나도 빠짐없이 열거하고
        # digest 로 묶었다"이다. `landed-commit.txt` 는 `openspec/` 아래라 drift 검사가
        # 통과시키고, 추적 파일이라 untracked 감사도 못 보고, 닫힌 키 집합에도 없다.
        # 그런데 비교 대상을 고른다 — 손잡이는 하나여야 하고 그것은 감사된 쪽이다.
        return [ADOPTION_REFUSES_A_LANDING]
    try:
        # 조상 판정은 `base-commit.txt` 의 글자가 아니라 `resolve_base` 가 **반환한**
        # 값에 건다. a063 은 그 둘이 다르다(P → E).
        # 이관이면 착지를 **해소하지 않는다**. `resolve_landing` 의 판정들은 저자가
        # 고른 값을 위한 것이고, 감사된 source 는 고른 값이 아니다 — `validate` 가
        # ancestry(P,E)·ancestry(E,source)·ancestry(source,head,strict)·tree 대조·
        # digest 셋으로 이미 묶는다. 같은 판정을 두 번 하지 않는다.
        landing = str(facts.get("adoption_source", "")) if adopted \
            else resolve_landing(change_dir, root, base, analysis)
        required = changed_existing_functions(root, base, landing)
    except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
        return [f"cannot derive modified Go functions: {exc}"]
    facts["landing"] = landing
    facts["required_count"] = len(required)
    if not landing:
        # 조언 줄은 대상이 워킹트리일 때만 나가므로 그때만 잰다 (task 7.1).
        facts["base_shaped_bundles"] = _base_shaped_bundles(root, base, analysis)
    if not analysis.exists():
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
    targets = sorted(path for path in analysis.iterdir() if path.is_dir())
    if not targets:
        return ["function-logic analysis directory has no targets"]
    covered: dict[tuple[str, str], Path] = {}
    # Built once: the tree has thousands of test functions and every target
    # would otherwise rescan them.
    index = test_index(root)
    # 열거형 호출 표를 **어디서든** 쓰는 change 는 모든 번들에서 써야 한다.
    # 번들마다 표지를 고를 수 있으면 감사 여부를 저자가 정하게 되고, 그 문으로
    # a112 의 39개가 빠져나갔다(4차 적대 리뷰). 판정은 표지 철자가 아니라
    # 표의 **내용**으로 한다 — 철자로 보던 판본이 공백 하나에 뚫렸다(6차).
    require_calls = call_enumeration_in_use(
        (_bundle_text(path), _ast_value(path.parent / "ast.json"))
        for path in sorted(analysis.glob("*/function-logic-map.md"))
    )
    for target in targets:
        target_errors, binding = validate_target(target, root, index, require_calls, landing)
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
        try:
            ast_value = json.loads((target / "ast.json").read_text(encoding="utf-8"))
        except (OSError, ValueError):
            continue
        expected_hash = expected.get("current_hash") or expected.get("base_hash")
        if ast_value.get("source_sha256") != expected_hash:
            errors.append(f"{target.name}: AST hash does not match modified function revision")
        expected_revision = "current" if expected.get("current_hash") else "base"
        if ast_value.get("revision", "current") != expected_revision:
            errors.append(f"{target.name}: AST revision must be {expected_revision}")
    return errors


def _walk_floor(
    root: Path, analysis: Path,
) -> tuple[str, str, list[tuple[Path, str, str]]]:
    """걷기 **전에** 정해지는 것 — 후보 순회가 설 하한. `(하한, 못 서는 사유)`.

    `compute_landing` 과 `_recording_refusal` 이 **같이** 묻는다 (task 7.7, 리뷰 I8). 조언
    줄은 걷지 않고 이것까지만 묻는데, 여기 두 사유를 조언 쪽이 따로 들고 있으면 계산이
    문장을 바꿀 때 조언만 옛 문장으로 남는다([[two-judgements-cover-for-each-other]]).
    """
    bundles = _pinning_bundles(root, analysis)
    if not bundles:
        return "", "no `revision: current` evidence pins a landing for this change", []
    floor = _evidence_floor(root, bundles)
    if not floor:
        # 걸은 것만 말한다 (task 7.2.4, 리뷰 H7). 하한을 찾는 `git log` 는 `-m` 이 없어서 병합 커밋
        # 자신의 변경을 읽지 않는다 — 병합을 마치며 번들을 처음 커밋하면 여기로 온다. 옛 문장
        # "never entered this history (commit the bundles)" 는 그 경우 거짓이었고, 이미 한 커밋을
        # 또 하라고 권했다. 동작은 사람이 2026-09-14 에 한계로 두기로 골랐다(막는 쪽으로 틀린다).
        return "", (
            "no ordinary commit on this history adds the pinning evidence — commit the bundles "
            "in an ordinary commit (a merge commit's own changes are not read)"
        ), bundles
    return floor, "", bundles


def compute_landing(root: Path, base: str, analysis: Path) -> tuple[str, str]:
    """게이트가 기록할 착지 지점. **저자가 고르지 않는다** (task 6.1.2).

    `(값, 못 정한 사유)` 를 돌려준다. 값이 있으면 사유는 빈 문자열이다.

    고르는 규칙은 하나다: 바닥(그 change 의 고정 번들이 역사에 들어온 지점) 이후
    이면서 고정을 통과하는 **가장 낮은** 커밋. 가장 낮은 것을 고르는 이유는 창을
    좁히는 것이 이 기능의 목적이기 때문이고, 그것이 안전한 이유는 바닥 아래로는
    못 내려가기 때문이다 — 저자는 오늘 만든 번들을 과거 커밋에 넣을 수 없다.
    """
    floor, why, bundles = _walk_floor(root, analysis)
    if why:
        return "", why
    start = floor if _is_ancestor(root, base, floor) else base
    process = subprocess.run(
        ["git", "rev-list", "--reverse", f"{start}..HEAD"],
        cwd=root, capture_output=True, text=True, timeout=60, check=False,
    )
    if process.returncode:
        return "", f"cannot walk the history after {start[:12]}"
    # 후보마다 다시 재지 않는다 — `floor` 와 같은 모양으로 **한 번** 잰다 (task 7.2.6).
    repairs = _self_repair_commits(root, analysis)
    first = ""
    unheld: list[str] = []
    for candidate in [start, *process.stdout.split()]:
        # 선언 경로와 **같은 함수**에 묻는다 (task 7.6, 리뷰 I2). 받는 가장 낮은 후보가 착지다.
        refusal, names = _landing_refusal(root, base, candidate, bundles, floor, repairs)
        if not refusal:
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


def _recording_refusal(change: str, change_dir: Path, root: Path) -> tuple[str, str]:
    """`--record-landing` 이 **걷기 전에** 멈추는 사유. `(사유, base)` — 멈추면 base 는 빈칸이다.

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
    if _landing_record(change_dir, root) is not None:
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
    except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
        return f"cannot resolve the comparison base: {exc}", ""
    if facts.get("execution_baseline_adoption"):
        return ADOPTION_REFUSES_A_LANDING, ""
    _, why, _bundles = _walk_floor(root, change_dir / "analysis" / "function-logic")
    if why:
        return f"no landing recorded — {why}", ""
    # 추적 파일 수정은 **맨 뒤**다 — 커밋하면 사라지는 유일한 사유라서다. 앞에 두면 영원히
    # 기록할 수 없는 change(번들 0 · 빌리는 쪽)가 "먼저 커밋하라"를 듣고, 커밋한 뒤에야
    # 진짜 사유를 듣는다. 이 저장소의 활성 change 일곱이 번들 0 이고, tasks.md 한 줄만
    # 고쳐도 트리는 dirty 다 (task 7.7 — 변이 R10 이 이 순서를 재는 시험이 0 임을 보였다).
    dirty = subprocess.run(
        ["git", "diff", "--quiet", "HEAD"], cwd=root, capture_output=True,
        timeout=30, check=False,
    )
    if dirty.returncode:
        return (
            "the working tree has uncommitted changes to tracked files — commit "
            "them first, because a recorded landing points at a commit and step 5 would "
            "then never compare those edits"
        ), ""
    return "", base


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
    try:
        # 걷기 전에 멈추는 사유는 5단계의 조언 줄이 묻는 **그 함수**에 묻는다 (task 7.7, I8).
        # 경계 **안에서** 묻는다: 걷기 전 부분도 번들을 읽으므로 저장소 밖을 가리키는 번들이
        # 거기서 터진다 — 이 판정이 경계 밖에 있던 첫 판본을 7.4 의 시험이 잡았다.
        refusal, base = _recording_refusal(change, change_dir, root)
        if refusal:
            return 1, [f"{change}: {refusal}"]
        landing, why = compute_landing(root, base, change_dir / "analysis" / "function-logic")
    except GATE_FAULTS as exc:
        # 저장소 밖을 가리키는 번들·읽을 수 없는 ast.json·멎은 git 은 전부 이 함수가
        # 답할 수 있는 것이다. 스택으로 죽으면 "왜 기록이 안 됐나"가 아무 데도 안 남는다.
        return 1, [f"{change}: no landing recorded — {exc}"]
    if not landing:
        return 1, [f"{change}: no landing recorded — {why}"]
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
        print(
            f"[logic-map] {args.change}: base {base[:12]} → {_target_text(landing, audited)} "
            f"required {context.get('required_count', 0)} function(s)"
        )
        if not landing:
            landed_after = _commits_after(root, base)
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
            if base_shaped:
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
                    refusal, _ = _recording_refusal(
                        args.change, resolve_referenced_change(root, args.change), root
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
