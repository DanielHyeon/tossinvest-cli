#!/usr/bin/env python3
"""a066 Wave 2A — ast.json 분기마다 소스 조건·도입 커밋·문장 커버리지를 측정해 표 행으로 출력함.

사용: python3 branch_coverage_rows.py <bundle-dir> <coverprofile[,coverprofile...]> [annotations.json] [per-test-dir]

- per-test-dir: `<TestName>.out` 프로필 모음(시험 하나씩 `-test.run '^Name$' -test.coverprofile` 로 만든 것).
  주면 분기 본문 블록을 실행한 시험 이름을 측정으로 붙임(최대 3개 + 총수). annotations 의 시험 이름이 우선함.

- 프로필을 쉼표로 여럿 주면 같은 블록의 실행수를 합침(예: 태그 없는 스위트 + `-tags tossos_testseams` 스위트).
- 도입 커밋이 아래 A066_COMMITS 에 있으면 a066 소유 분기로 표시함(2026-08-03~04 a066 구현 커밋 일곱: Wave 1A 8b9821de 포함).

- 조건 텍스트: ast.json 좌표의 실제 소스 줄(손으로 옮기지 않음).
- 도입 커밋: `git blame -L n,n` 의 커밋(분기 줄을 마지막으로 바꾼 커밋).
- 커버리지: `go test -coverprofile` 의 블록 중 분기 좌표 **뒤에서 처음 시작하는** 블록(6줄 이내)의
  실행 여부(40줄 이내 — 여러 줄 조건) — `if`/`for`/`range`/`case` 에서는 참 갈래(본문)가 한 번이라도 실행됐는가임. `else` 는 같은
  줄에서 else 토큰 ±2칸에 시작하는 블록을 봄(`} else if` 이면 else 갈래의 조건 평가 블록).
  문장 커버리지이므로 "그 블록의 문장이 한 번이라도 실행됨"만 말하고, 어느 시험이 실행했는지는 말하지 않음.
- annotations.json: {"B7": {"scenario": "...", "tests": ["TestX"], "red": "..."}} — 이 change 가 소유한
  분기에 손으로 붙이는 시나리오·시험 이름. 없으면 측정값만 씀.
"""
import json
import re
import subprocess
import sys
from pathlib import Path


A066_COMMITS = {"8b9821de", "4aee6853", "c60fee07", "4a364caf", "9bf1a3a6", "a37d97f5", "ebb87d3d"}


def blocks(profiles: str, relative: str):
    """커버리지 프로필(들)에서 한 파일의 블록 (시작줄, 시작칸, 끝줄, 실행수) 목록을 읽음."""
    out = []
    lines = []
    for profile in profiles.split(","):
        lines.extend(Path(profile).read_text().splitlines()[1:])
    for line in lines:
        match = re.match(r"^(\S+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$", line)
        if match and match.group(1).endswith("/" + relative):
            out.append((int(match.group(2)), int(match.group(3)), int(match.group(4)), int(match.group(7))))
    # 같은 블록이 여러 번 나오면(패키지 여러 번 측정) 실행수를 합침
    merged = {}
    for start_line, start_col, end_line, count in out:
        key = (start_line, start_col, end_line)
        merged[key] = merged.get(key, 0) + count
    return sorted((k[0], k[1], k[2], v) for k, v in merged.items())


def main() -> None:
    bundle, profile = Path(sys.argv[1]), sys.argv[2]
    notes = json.loads(Path(sys.argv[3]).read_text()) if len(sys.argv) > 3 and sys.argv[3] != "-" else {}
    ast = json.loads((bundle / "ast.json").read_text())
    relative = ast["file"]
    # 시험별 프로필 — 분기 본문을 실행한 시험 이름을 측정으로 고르기 위함
    per_test = {}
    if len(sys.argv) > 4:
        for path in sorted(Path(sys.argv[4]).glob("Test*.out")):
            per_test[path.stem] = blocks(str(path), relative)
    source = Path(relative).read_text().split("\n")
    profile_blocks = blocks(profile, relative)
    for branch in ast.get("branches") or []:
        line, col = branch["at"]["line"], branch["at"]["column"]
        text = source[line - 1].strip().replace("|", "\\|")
        blame = subprocess.run(["git", "blame", "-L", f"{line},{line}", "--porcelain", relative],
                               capture_output=True, text=True, check=True).stdout.split()[0][:8]
        if branch["kind"] == "else":
            # `} else {` 의 본문 블록은 else 토큰 바로 앞 칸에서 시작함(Go 커버리지 블록 경계)
            after = [b for b in profile_blocks if b[0] == line and abs(b[1] - col) <= 2]
        else:
            after = [b for b in profile_blocks if (b[0], b[1]) > (line, col) and b[0] <= line + 40]
        covered = "no block" if not after else ("covered" if after[0][3] > 0 else "NOT covered")
        # 분기 본문의 첫 문장 — 본문 블록이 시작하는 줄(`{` 줄) 다음의 첫 코드 줄. 블록이 없으면 분기 줄 다음 줄.
        opener = (after[0][0] - 1) if after else (line - 1)
        then = next((s.strip() for s in source[opener + 1:opener + 6] if s.strip() and not s.strip().startswith("//")), "")
        then = then.replace("|", "\\|")[:160]
        note = notes.get(branch["id"], {})
        tests = ", ".join(f"`{name}`" for name in note.get("tests", []))
        if not tests and per_test and after:
            target = after[0][:3]
            hits = [name for name, test_blocks in per_test.items()
                    if any(b[:3] == target and b[3] > 0 for b in test_blocks)]
            if hits:
                tests = ", ".join(f"`{name}`" for name in hits[:3]) + (f" (+{len(hits) - 3} more)" if len(hits) > 3 else "") + " — per-test coverprofile"
        print(json.dumps({"id": branch["id"], "kind": branch["kind"], "at": f"{line}:{col}", "text": text,
                          "commit": blame, "coverage": covered, "scenario": note.get("scenario", ""),
                          "tests": tests, "red": note.get("red", ""), "a066": blame in A066_COMMITS, "then": then}, ensure_ascii=False))


if __name__ == "__main__":
    main()
