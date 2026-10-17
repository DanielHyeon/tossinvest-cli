#!/usr/bin/env python3
"""통합 diff 본문 줄이 파일 헤더로 읽히는지 **진짜 git 으로** 잰다 (a122 task 7.5.2.4).

`git diff --unified=0` 의 본문에는 문맥 줄이 없고 지워진 줄은 `-`+내용, 더한 줄은
`+`+내용이다. 그래서 `-- ` 로 시작하는 **소스 줄**은 `--- …` 로, `++ ` 로 시작하는
소스 줄은 `+++ …` 로 나온다 — 파일 헤더와 글자가 같다.

이 하네스는 지어낸 diff 문자열을 쓰지 않는다. 네 벌의 **진짜 저장소**를 만들고 진짜
`git diff` 를 흘려서 `changed_existing_functions` 가 무엇을 필수로 세는지 찍는다.
Go 추출기만 세우고(고정 저장소에 `./tools/logic-map` 이 없다) 나머지는 전부 생산 코드다.

    python3 7524_diff.py

수리 전후로 같은 명령을 돌려 표를 비교한다 — 이것은 이름의 증거가 아니라 그 소스의 증거다.
"""
import os
import re
import subprocess
import sys
from pathlib import Path
from unittest import mock

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
WORK = Path(__file__).resolve().parent / "_work" / f"7524.{os.getpid()}"
sys.path.insert(0, str(ROOT / "tools" / "logic-map"))
import check_analysis  # noqa: E402

FUNC = re.compile(r"^func\s+(?:\([^)]*\)\s*)?(\w+)", re.MULTILINE)


def stub_go_functions(path: Path, root: Path) -> list[dict]:
    """`go run` 대신 줄 번호만 센다 — 이 하네스가 재는 것은 파서이지 추출기가 아니다."""
    lines = Path(path).read_text(encoding="utf-8").splitlines()
    found: list[dict] = []
    for index, line in enumerate(lines, start=1):
        match = FUNC.match(line)
        if not match:
            continue
        end = index
        while end < len(lines) and lines[end - 1] != "}" and not lines[end - 1].endswith("}"):
            end += 1
        found.append({
            "function": match.group(1),
            "start": {"line": index},
            "end": {"line": end},
            "source_sha256": f"sha-of-{Path(path).name}-{match.group(1)}",
        })
    return found


def build(name: str, before: str, after: str) -> tuple[Path, str]:
    repo = WORK / name
    repo.mkdir(parents=True)
    run = lambda *args: subprocess.run(args, cwd=repo, check=True, capture_output=True)
    run("git", "init", "-q", ".")
    run("git", "config", "user.email", "harness@example.com")
    run("git", "config", "user.name", "harness")
    (repo / "x.go").write_text(before, encoding="utf-8")
    run("git", "add", ".")
    run("git", "commit", "-qm", "base")
    base = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repo, text=True).strip()
    (repo / "x.go").write_text(after, encoding="utf-8")
    run("git", "add", ".")
    run("git", "commit", "-qm", "edit")
    return repo, base


# 소스 줄 하나만 다르다. F 의 편집은 네 벌 모두 같다.
BODY = {
    "control": ("keep", "keep"),
    "deleted-dev-null": ("-- /dev/null\nkeep", "keep"),
    "added-dev-null": ("keep", "keep\n++ /dev/null"),
    "accidental-sql-comment": ("-- name of the table\nkeep", "keep"),
}
BEFORE = "package pkg\n\nconst q = `\n{body}\n`\n\nfunc F() int {{\n\treturn 1\n}}\n"
AFTER = "package pkg\n\nconst q = `\n{body}\n`\n\nfunc F() int {{\n\treturn 2\n}}\n"


def main() -> int:
    WORK.mkdir(parents=True, exist_ok=True)
    print(f"# 7524_diff — 진짜 git · 진짜 파서 (work={WORK})\n")
    print("| 벌 | git 이 낸 헤더 모양 줄 | required | 바뀐 자리 |")
    print("|---|---|---|---|")
    worst = 0
    for name, (before_body, after_body) in BODY.items():
        repo, base = build(name, BEFORE.format(body=before_body), AFTER.format(body=after_body))
        text = subprocess.check_output(
            ["git", "-c", "core.quotePath=false", "diff", "--no-ext-diff", "--find-renames",
             "--unified=0", base, "HEAD", "--", "*.go"],
            cwd=repo, text=True)
        shaped = [line for line in text.splitlines()[4:]
                  if line.startswith("--- ") or line.startswith("+++ ")]
        try:
            with mock.patch("check_analysis.go_functions", side_effect=stub_go_functions):
                required = check_analysis.changed_existing_functions(repo, base, "HEAD")
            cell = ", ".join(sorted(f"{file}:{function}" for file, function in required)) or "**0 건**"
            keys = ", ".join(sorted({key for value in required.values() for key in value
                                     if key.endswith("_hash")})) or "-"
        except Exception as error:  # noqa: BLE001 — 무엇이 올라오는지 그대로 찍는 것이 측정이다
            cell = f"**{type(error).__name__}**: {error}"
            keys = "-"
            worst = 1
        print(f"| `{name}` | {' / '.join(f'`{s}`' for s in shaped) or '없음'} | {cell} | {keys} |")
    # 같은 키를 현재 쪽 통과가 **덮는다**(`required[key] = …`) — 그래서 정상 결과의 해시 칸은
    # `current_hash` 하나다. `base_hash` 만 남는 것은 현재 쪽이 안 돈다는 뜻이고, 그것이
    # `revision: base` 를 요구하는 갈래다.
    print("\n기대(수리 뒤): 네 벌 모두 `x.go:F` 1건 · 해시 칸은 `current_hash` · 예외 0.")
    return worst


if __name__ == "__main__":
    raise SystemExit(main())
