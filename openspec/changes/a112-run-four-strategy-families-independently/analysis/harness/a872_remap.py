#!/usr/bin/env python3
"""a112 8.7.2 — 본문이 바뀌지 않은 함수의 번들 좌표를 기계적으로 다시 맞춘다.

쓰는 법: a872_remap.py <번들 디렉터리>...

같은 파일의 다른 함수가 편집되면 그 뒤 함수들은 **본문이 바이트 동일한 채** 줄 좌표만 밀린다.
그런 번들은 새 증거가 아니라 좌표 이동이 필요하다. 이 스크립트는

1. 번들 ast.json 이 가리키는 파일의 HEAD blob 이 그 ast 의 source_sha256 과 같은지 확인하고(아니면 멈춤),
2. HEAD 와 워킹트리에서 그 함수의 본문 텍스트가 **바이트 동일**한지 확인하고(아니면 멈춤 — 그 함수는 편집된 것이다),
3. 새 ast 를 떠서 옛/새 ast 의 분기·반환·호출 좌표를 **순서대로 짝지어** md 파일의 `줄:열` 을 한 번에 바꾼다.

짝짓기는 개수가 같을 때만 한다(본문이 같으면 같아야 한다 — 다르면 멈춘다).
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = Path(__file__).resolve().parents[5]
COORD = re.compile(r"(?<![\d.])(\d+):(\d+)(?![\d])")


def logic_map(path: Path, function: str) -> dict:
    out = subprocess.run(["go", "run", "./tools/logic-map", "--file", str(path), "--func", function],
                         cwd=REPO, capture_output=True, text=True, check=True).stdout
    return json.loads(out)


def body(text: str, ast: dict) -> str:
    lines = text.splitlines()
    return "\n".join(lines[ast["start"]["line"] - 1: ast["end"]["line"]])


def positions(ast: dict) -> list[str]:
    values = []
    for key in ("branches", "returns", "calls"):
        for node in ast.get(key) or []:
            values.append(f"{node['at']['line']}:{node['at']['column']}")
    return values


def main() -> int:
    for bundle in map(Path, sys.argv[1:]):
        old = json.loads((bundle / "ast.json").read_text())
        source = old["file"]
        head = subprocess.run(["git", "show", f"HEAD:{source}"], cwd=REPO, capture_output=True, check=True).stdout
        import hashlib
        if hashlib.sha256(head).hexdigest() != old["source_sha256"]:
            print(f"STOP {bundle.name}: 번들 ast 가 HEAD blob 이 아니다")
            return 1
        name = f"{old['receiver']}.{old['function']}" if old.get("receiver") else old["function"]
        with tempfile.TemporaryDirectory() as tmp:
            copy = Path(tmp) / Path(source).name
            copy.write_bytes(head)
            head_ast = logic_map(copy, name)
        new = logic_map(REPO / source, name)
        if body(head.decode(), head_ast) != body((REPO / source).read_text(), new):
            print(f"STOP {bundle.name}: 본문이 바뀌었다 — 좌표 이동이 아니라 새 증거가 필요하다")
            return 1
        before, after = positions(old), positions(new)
        if len(before) != len(after):
            print(f"STOP {bundle.name}: 좌표 개수가 다르다")
            return 1
        mapping = dict(zip(before, after))
        mapping[f"{old['start']['line']}:{old['start']['column']}"] = f"{new['start']['line']}:{new['start']['column']}"
        mapping[f"{old['end']['line']}:{old['end']['column']}"] = f"{new['end']['line']}:{new['end']['column']}"
        delta = new["start"]["line"] - old["start"]["line"]
        new["file"] = source
        (bundle / "ast.json").write_text(json.dumps(new, indent=2, ensure_ascii=False) + "\n")
        for md in bundle.glob("*.md"):
            text = md.read_text()
            text = COORD.sub(lambda m: mapping.get(m.group(0), m.group(0)), text)
            text = text.replace(old["source_sha256"], new["source_sha256"])
            text = re.sub(rf"\({old['start']['line']}-{old['end']['line']}\)",
                          f"({new['start']['line']}-{new['end']['line']})", text)
            md.write_text(text)
        print(f"remapped {bundle.name}: Δ={delta:+d} lines, {len(before)} coordinates")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
