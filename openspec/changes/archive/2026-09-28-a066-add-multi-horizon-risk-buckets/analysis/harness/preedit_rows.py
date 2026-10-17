#!/usr/bin/env python3
"""a066 5.6.1 — 편집 **전** 리비전(git rev)의 소스로 FLM 분기 행을 만듦.

branch_coverage_rows.py 는 작업 트리 소스를 읽으므로 편집이 먼저 들어간 뒤에는 편집 전 AST 좌표와 어긋남.
이 스크립트는 ast.json 좌표를 `git show <rev>:<file>` 소스에 맞춰 읽고, 편집 전 소스로 돈 coverprofile 을 대조함.

사용: preedit_rows.py <bundle-dir> <rev> <coverprofile[,..]>  → FLM 분기 표 행(markdown) 출력
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from branch_coverage_rows import blocks  # noqa: E402


def main() -> None:
    bundle, rev, profiles = Path(sys.argv[1]), sys.argv[2], sys.argv[3]
    ast = json.loads((bundle / "ast.json").read_text())
    source = subprocess.run(["git", "show", f"{rev}:{ast['file']}"], capture_output=True, text=True, check=True).stdout.splitlines()
    profile_blocks = blocks(profiles, ast["file"])
    for branch in ast.get("branches") or []:
        line, col = branch["at"]["line"], branch["at"]["column"]
        text = source[line - 1].strip().replace("|", "\\|")[:170]
        after = [b for b in profile_blocks if (b[0], b[1]) > (line, col) and b[0] <= line + 40]
        covered = "no block" if not after else ("covered" if after[0][3] > 0 else "NOT covered")
        print(f"| {branch['id']} | {branch['kind']} at {line}:{col} | `{text}` | {covered} |")


if __name__ == "__main__":
    main()
