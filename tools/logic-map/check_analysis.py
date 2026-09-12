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


class AmbiguousChange(ValueError):
    """같은 id 가 활성과 아카이브에 **동시에** 있다.

    호출자가 이 실패를 문구가 아니라 **타입**으로 가릴 수 있도록 따로 둔다.
    문구로 가르면 메시지를 고치는 순간 조용히 뚫린다. `ValueError` 를 상속하는
    것은 이 함수의 다른 실패를 이미 `ValueError` 로 받고 있는 자리를 깨지 않기
    위해서다.
    """


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
        if path.is_dir()
        and (matched := ARCHIVED_CHANGE.fullmatch(path.name)) is not None
        and matched.group("change") == change
    )
    found = ([direct] if open_here else []) + archived
    if not found:
        raise ValueError(f"reference change is neither open nor archived: {change}")
    if open_here and archived:
        raise AmbiguousChange(
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


LANDING_FILE = "landed-commit.txt"
FULL_SHA = re.compile(r"[0-9a-f]{40}")


def _committed_bytes(root: Path, ref: str, relative: str) -> bytes | None:
    """`ref` 시점의 파일 내용. 워킹트리를 보지 않는다."""
    process = subprocess.run(
        ["git", "show", f"{ref}:{relative}"],
        cwd=root, capture_output=True, timeout=30, check=False,
    )
    return None if process.returncode else process.stdout


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
        return "working tree (no landed-commit.txt)"
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


def _pinning_at(root: Path, candidate: str, analysis: Path) -> tuple[int, list[str]]:
    """`candidate` 에서 고정 번들이 몇 개이고 그중 어느 소스가 안 맞는가."""
    mismatched: list[str] = []
    bundles = _pinning_bundles(root, analysis)
    for _, source, digest in bundles:
        blob = _committed_bytes(root, candidate, source)
        if blob is None or hashlib.sha256(blob).hexdigest() != digest:
            mismatched.append(source)
    return len(bundles), mismatched


def _pre_archive_path(relative: str) -> str:
    """아카이브된 경로의 **옮기기 전** 이름. 아카이브가 아니면 빈 문자열이다."""
    prefix = "openspec/changes/archive/"
    if not relative.startswith(prefix):
        return ""
    dated, _, tail = relative[len(prefix):].partition("/")
    match = ARCHIVED_CHANGE.fullmatch(dated)
    return f"openspec/changes/{match.group('change')}/{tail}" if match and tail else ""


def _evidence_floor(root: Path, analysis: Path) -> str:
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
    for ast_path, _, _ in _pinning_bundles(root, analysis):
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


def _unheld_bundles(root: Path, candidate: str, analysis: Path) -> list[str]:
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
    unheld: list[str] = []
    for ast_path, _, _ in _pinning_bundles(root, analysis):
        try:
            relative = ast_path.relative_to(root).as_posix()
        except ValueError:
            continue
        try:
            judged = ast_path.read_bytes()  # 판정과 **같은 읽기** — 심링크면 따라간다
        except OSError:
            unheld.append(ast_path.parent.name)
            continue
        committed = _committed_bytes(root, candidate, relative)
        if committed is None:
            before = _pre_archive_path(relative)
            if before:
                committed = _committed_bytes(root, candidate, before)
        if committed is None or committed != judged:
            unheld.append(ast_path.parent.name)
    return sorted(set(unheld))


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
    # base 는 조상이면 되고 같아도 된다. a074·a079·a075 의 정답이 바로 같은 경우다 —
    # 2026-08-04 재기준화가 base 를 그 change 들의 작업 뒤로 옮겼다. 느슨해지지 않는
    # 이유는 아래 증거 판정이 값을 고정하기 때문이다.
    if not _is_ancestor(root, base, candidate):
        raise ValueError(
            f"landing point precedes the comparison base {base[:12]}: {candidate}"
        )
    pinning, mismatched = _pinning_at(root, candidate, analysis)
    if not pinning:
        # 위의 판정들은 전부 "이것이 어느 커밋인가"만 묻는다. 어느 커밋인지를
        # **고르지 못하게** 하는 것은 이 순회 하나뿐이고, 순회가 0회 돌면 저자가
        # 구간의 바닥을 골라 요구 집합을 ∅ 로 만들 수 있다. 그 상태는 면제와
        # 구분되지 않는다 — 고정할 증거가 없으면 선언도 없다.
        raise ValueError(
            f"landing point {candidate[:12]} is pinned by no `revision: current` "
            "evidence: a declared landing must be the revision some bundle describes"
        )
    if mismatched:
        raise ValueError(
            f"landing point {candidate[:12]} is not the revision this evidence describes: "
            + ", ".join(sorted(set(mismatched)))
        )
    # 해시가 맞는다는 것만으로는 값이 안 정해진다. a122 6.1.1 이 활성 13건을 전수로
    # 재서 얻은 것: 고정을 통과하는 커밋이 13건 중 12건에서 2~59개이고, 그 수는 번들
    # 수와 무관하다(a112 번들 132 → 후보 39, a091 번들 2 → 후보 27). 저자가 그중
    # **가장 낮은 것**을 고르면 창이 비어 요구 집합이 ∅ 이 된다 — 8건이 오늘 그
    # 상태이고, 증거를 base 상태로 써 두면 누구나 그 상태를 만들 수 있다.
    # 하한 하나만 저자가 못 고른다: 자기 증거가 역사에 들어온 지점.
    floor = _evidence_floor(root, analysis)
    if not floor:
        raise ValueError(
            f"landing point {candidate[:12]} is pinned by evidence that never entered "
            "this history: commit the `revision: current` bundles that pin it"
        )
    if not _is_ancestor(root, floor, candidate):
        raise ValueError(
            f"landing point {candidate[:12]} precedes the evidence that pins it "
            f"({floor[:12]}): a declared landing cannot be older than the commit that "
            "put this change's evidence into the history"
        )
    # 하한 **뒤**에 선다. 앞에 두면 "증거가 역사에 없다"·"착지가 증거보다 앞선다"를
    # 재던 시험들의 거절 지점을 이 등식이 가로채서, 그 가드를 지워도 스위트가 초록으로
    # 남는다([[first-failure-is-not-the-fix-scope]] 가 같은 파일에서 실측한 모양).
    unheld = _unheld_bundles(root, candidate, analysis)
    if unheld:
        raise ValueError(
            f"landing point {candidate[:12]} does not hold the evidence this verdict read: "
            + ", ".join(unheld)
            + " — the bundle judged here is not the bundle committed there"
        )
    return candidate


def normalized_source(value: str, root: Path) -> tuple[Path, str]:
    raw = Path(value)
    path = raw if raw.is_absolute() else root / raw
    resolved = path.resolve()
    if not resolved.is_relative_to(root.resolve()):
        raise ValueError("AST source escapes repository")
    return resolved, resolved.relative_to(root.resolve()).as_posix()


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
    try:
        change_dir = resolve_referenced_change(root, change)
    except AmbiguousChange as exc:
        # 아래 fallback 으로 흘려보내면 활성이 아카이브를 조용히 이긴다 —
        # 그 침묵이 이 자리에서 고치려는 것 자체다. 타입으로 가른다.
        return [str(exc)]
    except ValueError:
        change_dir = root / "openspec" / "changes" / change
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
        # 창의 **끝**도 같아야 한다. 빌린 번들이 착지를 고정하긴 하지만, 그 번들이
        # 안 바뀌는 구간 안에서는 저자가 여전히 고를 수 있다 — a072 실측으로 326
        # 커밋 중 2개가 그 구간이다. 선언 파일을 고른 근거가 "번들이 고정하므로
        # 저자가 고를 수 없다" 하나이므로, 남의 번들로 고정할 때 남는 그 선택을
        # 없앤다. 착지는 그것을 고정하는 증거가 사는 자리에 선언하고 빌리는 쪽은
        # 값을 **복사**한다.
        try:
            shared = _declared_landing(change_dir, root) == _declared_landing(referenced_dir, root)
        except ValueError as exc:
            return [f"cannot derive modified Go functions: {exc}"]
        if not shared:
            return ["function-logic reference must share the exact landing point"]
        analysis = referenced_dir / "analysis" / "function-logic"
    adopted = bool(facts.get("execution_baseline_adoption"))
    if adopted and _landing_record(change_dir, root) is not None:
        # 이관 예외의 정당성은 "판정에 들어가는 입력을 하나도 빠짐없이 열거하고
        # digest 로 묶었다"이다. `landed-commit.txt` 는 `openspec/` 아래라 drift 검사가
        # 통과시키고, 추적 파일이라 untracked 감사도 못 보고, 닫힌 키 집합에도 없다.
        # 그런데 비교 대상을 고른다 — 손잡이는 하나여야 하고 그것은 감사된 쪽이다.
        return [
            "execution-baseline adoption does not accept a "
            f"`{LANDING_FILE}` record: the window ends at the audited source commit"
        ]
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


def compute_landing(root: Path, base: str, analysis: Path) -> tuple[str, str]:
    """게이트가 기록할 착지 지점. **저자가 고르지 않는다** (task 6.1.2).

    `(값, 못 정한 사유)` 를 돌려준다. 값이 있으면 사유는 빈 문자열이다.

    고르는 규칙은 하나다: 바닥(그 change 의 고정 번들이 역사에 들어온 지점) 이후
    이면서 고정을 통과하는 **가장 낮은** 커밋. 가장 낮은 것을 고르는 이유는 창을
    좁히는 것이 이 기능의 목적이기 때문이고, 그것이 안전한 이유는 바닥 아래로는
    못 내려가기 때문이다 — 저자는 오늘 만든 번들을 과거 커밋에 넣을 수 없다.
    """
    if not _pinning_bundles(root, analysis):
        return "", "no `revision: current` evidence pins a landing for this change"
    floor = _evidence_floor(root, analysis)
    if not floor:
        return "", "the pinning evidence never entered this history (commit the bundles)"
    start = floor if _is_ancestor(root, base, floor) else base
    process = subprocess.run(
        ["git", "rev-list", "--reverse", f"{start}..HEAD"],
        cwd=root, capture_output=True, text=True, timeout=60, check=False,
    )
    if process.returncode:
        return "", f"cannot walk the history after {start[:12]}"
    unheld: list[str] = []
    for candidate in [start, *process.stdout.split()]:
        if not _is_ancestor(root, base, candidate) or not _is_ancestor(root, floor, candidate):
            continue
        if _pinning_at(root, candidate, analysis)[1]:
            continue
        # 고정을 통과해도, 판정이 읽은 번들을 그 커밋이 들고 있지 않으면 착지가 아니다
        # (task 7.2, 리뷰 C3). `resolve_landing` 과 **같은** 조건이다.
        unheld = _unheld_bundles(root, candidate, analysis)
        if not unheld:
            return candidate, ""
    if unheld:
        return "", (
            "no commit holds the evidence this verdict read (" + ", ".join(unheld) + ") "
            "— commit the bundles as they are on disk"
        )
    return "", (
        f"no commit at or after the evidence ({floor[:12]}) matches every pinning bundle "
        "— the evidence does not describe any revision on this history"
    )


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
        change_dir = resolve_referenced_change(root, change)
    except AmbiguousChange as exc:
        return 1, [str(exc)]
    except ValueError:
        change_dir = root / "openspec" / "changes" / change
    landing_file = change_dir / LANDING_FILE
    if landing_file.exists() or _landing_record(change_dir, root) is not None:
        return 1, [f"{change}: `{LANDING_FILE}` already exists — not overwritten"]
    if (change_dir / "analysis" / "function-logic-reference.txt").exists():
        return 1, [
            f"{change}: this change borrows its evidence — copy the landing recorded on "
            "the change that owns the bundles instead of computing a second one"
        ]
    facts: dict[str, object] = {}
    try:
        base = resolve_base(change_dir, root, facts, change_id=change)
    except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
        return 1, [f"{change}: cannot resolve the comparison base: {exc}"]
    if facts.get("execution_baseline_adoption"):
        return 1, [
            f"{change}: execution-baseline adoption does not accept a `{LANDING_FILE}` "
            "record: the window already ends at the audited source commit"
        ]
    dirty = subprocess.run(
        ["git", "diff", "--quiet", "HEAD"], cwd=root, capture_output=True,
        timeout=30, check=False,
    )
    if dirty.returncode:
        return 1, [
            f"{change}: the working tree has uncommitted changes to tracked files — commit "
            "them first, because a recorded landing points at a commit and step 5 would "
            "then never compare those edits"
        ]
    landing, why = compute_landing(root, base, change_dir / "analysis" / "function-logic")
    if not landing:
        return 1, [f"{change}: no landing recorded — {why}"]
    landing_file.write_text(landing + "\n", encoding="utf-8")
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
        code, lines = record_landing(args.change, root)
        for line in lines:
            print(f"[logic-map] {line}")
        return code
    errors = check(args.change, root, context)
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
                print(
                    f"{window} — run "
                    f"`python3 tools/logic-map/check_analysis.py --change {args.change} "
                    f"--record-landing` to let the gate compute and record "
                    f"`{LANDING_FILE}`, which narrows it to this change's own work"
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
