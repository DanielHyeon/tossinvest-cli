#!/usr/bin/env python3
"""B2 근거용 AST 열거 — 각 ast.json 의 분기 · return · 호출을 소스 줄과 함께 찍음.

목적: 「root 경로 문자열이 참여자 사이에 같아야 한다」는 요구가 코드에서 강제되는지를
손 읽기가 아니라 AST 열거로 판정함. 분기 조건 줄에 root 문자열(root · m.Root · b.root ·
c.Root · path)이 비교 대상으로 나오는지 표시함.

실행: 저장소 루트에서 `python3 -I <이 파일> <ast 디렉터리>`.
"""
import json
import pathlib
import re
import sys

ROOT_TOKENS = re.compile(r"\b(m\.Root|b\.root|c\.Root|cfg\.Manifest|root|path)\b")


def main() -> int:
    ast_dir = pathlib.Path(sys.argv[1])
    repo = pathlib.Path.cwd()
    for f in sorted(ast_dir.glob("*.ast.json")):
        d = json.loads(f.read_text())
        src = (repo / d["file"]).read_text().splitlines()
        print(f"## {d['file']} :: {d['function']}  (lines {d['start']['line']}-{d['end']['line']}, source_sha256 {d['source_sha256'][:16]})")
        print(f"branches={len(d['branches'] or [])} returns={len(d['returns'] or [])} calls={len(d['calls'] or [])}")
        for b in d["branches"] or []:
            line = b["at"]["line"]
            text = src[line - 1].strip()
            mark = "  <-- root/path 토큰" if ROOT_TOKENS.search(text) else ""
            print(f"- {b['id']} {b['kind']} L{line}: {text}{mark}")
        for c in d["calls"] or []:
            if ROOT_TOKENS.search(src[c["at"]["line"] - 1]):
                print(f"  call L{c['at']['line']}: {c['text']}  | {src[c['at']['line'] - 1].strip()}")
        print()
    return 0


if __name__ == "__main__":
    sys.exit(main())
