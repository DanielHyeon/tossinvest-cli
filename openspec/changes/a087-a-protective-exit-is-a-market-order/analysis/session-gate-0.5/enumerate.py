#!/usr/bin/env python3
"""a087 tasks 0.5 — 청산 판정 상류 함수의 AST 산출물에서 세션·시각 입력을 기계적으로 열거함.

손으로 읽은 증거는 볼 곳을 고르므로, 분기 줄과 호출 텍스트를 AST(`tools/logic-map`)가 준 좌표
전부에서 뽑아 정규식으로 분류함. 분류 정규식은 아래 두 개뿐이고 결과는 표준 출력의 Markdown 임.

사용:
    python3 enumerate.py            # 이 디렉터리의 *.ast.json 전부
재생성(같은 디렉터리의 AST 갱신):
    go build -o /tmp/logicmap ./tools/logic-map
    /tmp/logicmap --file internal/app/engine/exitloop.go --func ExitObserver.<F> > engine--exitobserver.<f>.ast.json
"""

from __future__ import annotations

import hashlib
import json
import re
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]

# 세션·달력 입력: 정규장 판정·장 시간·시장 현지 시각을 읽는 이름. 하나라도 걸리면 「세션 입력」임.
SESSION = re.compile(r"InRegularSession|RegularSession|TradingHours|TradingDay|Session|Location|Weekday|Hour\(|Calendar|Holiday",
                     re.IGNORECASE)
# 시각 입력: 벽시계·경과 시간·시세 신선도. 세션이 아니라 「얼마나 지났나」를 읽는 이름.
CLOCK = re.compile(r"Now|Since|Sub\(|\.Sub\b|After|Before|IsZero|Sleep|Interval|outageAfter|delayBound|Lease|FetchedAt|"
                   r"QueryPriceEvidenceDuration|FillDetectionBehind|quoteUsable|lastObserved|startedAt|observation\.at")


def classify(text: str) -> str:
    if SESSION.search(text):
        return "SESSION"
    if CLOCK.search(text):
        return "CLOCK"
    return ""


def main() -> int:
    paths = sorted(HERE.glob("*.ast.json"))
    if not paths:
        print("AST 산출물 0건 — 열거 불가", file=sys.stderr)
        return 1
    totals = {"SESSION": 0, "CLOCK": 0}
    out: list[str] = []
    for path in paths:
        ast = json.loads(path.read_text(encoding="utf-8"))
        source_path = ROOT / ast["file"]
        source_bytes = source_path.read_bytes()
        # AST 가 판정한 바이트와 지금 읽는 바이트가 같아야 줄 좌표가 유효함 — 다르면 멈춤.
        if hashlib.sha256(source_bytes).hexdigest() != ast["source_sha256"]:
            print(f"{path.name}: source_sha256 불일치 — AST 재생성 필요", file=sys.stderr)
            return 2
        lines = source_bytes.decode("utf-8").splitlines()
        # 패키지 함수는 receiver 키가 없음 — 메서드만 「수신자.이름」으로 표기함.
        name = f"{ast['receiver']}.{ast['function']}" if ast.get("receiver") else ast["function"]
        out.append(f"### `{name}` ({ast['file']}:{ast['start']['line']}-{ast['end']['line']})")
        out.append("")
        out.append(f"분기 {len(ast['branches'])} · 반환 {len(ast['returns'])} · 호출 {len(ast['calls'])}")
        out.append("")
        out.append("| 종류 | 좌표 | 원문 | 분류 |")
        out.append("|---|---|---|---|")
        for branch in ast["branches"]:
            line = branch["at"]["line"]
            text = lines[line - 1].strip()
            kind = classify(text)
            if kind:
                totals[kind] += 1
            cell = text.replace("|", "\\|")
            out.append(f"| 분기 {branch['id']} ({branch['kind']}) | :{line} | `{cell}` | {kind or '-'} |")
        for call in ast["calls"]:
            kind = classify(call["text"])
            if not kind:
                continue
            totals[kind] += 1
            out.append(f"| 호출 | :{call['at']['line']} | `{call['text']}` | {kind} |")
        out.append("")
    print("\n".join(out))
    print(f"합계: SESSION {totals['SESSION']} · CLOCK {totals['CLOCK']} (AST {len(paths)}개)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
