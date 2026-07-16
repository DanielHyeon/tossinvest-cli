#!/usr/bin/env python3
"""계좌 fixture 의 값을 합성 데이터로 치환한다.

fetch_auth_fixtures.py 의 sanitize() 는 식별자 키(accountNo/name/...)만 가린다.
금액·수량·종목코드까지 합성으로 바꿔야 fixture 가 온전히 더미가 되므로, 커밋 전에 이 단계를 거친다.

- 시드 고정 → 같은 입력이면 같은 출력 (재실행해도 diff 가 안 튄다)
- 구조·타입 보존: 0 과 null 은 그대로 둔다 (구조적 의미가 있음)
- 종목코드는 형식을 유지한 합성 코드로 매핑 (같은 코드는 같은 값으로)

    python3 tools/scrub_fixture_values.py fixtures/responses/auth-sanitized
"""

from __future__ import annotations

import json
import random
import re
import sys
from pathlib import Path
from typing import Any

MONEY_KEY = re.compile(
    r"(amount|price|krw|usd|balance|principal|profit|cash|won)", re.IGNORECASE
)
RATE_KEY = re.compile(r"rate$", re.IGNORECASE)
QTY_KEY = re.compile(r"quantity$", re.IGNORECASE)
# 종목 식별자는 키 이름이 제각각이고(productCode/stockCode/stockSymbol/stockIsin/
# productName/logoImageUrl…) 복합키·URL 안에도 박혀 있다. 키 목록으로 쫓지 않고,
# 먼저 문서 전체에서 식별자를 수집한 뒤 모든 문자열에서 치환한다.
ID_KEY = re.compile(
    r"(productCode|stockCode|isinCode|stockIsin|symbol|stockSymbol|ticker)$",
    re.IGNORECASE,
)
NAME_KEY = re.compile(r"(stockName|productName)$", re.IGNORECASE)

# 결과가 한눈에 '합성'으로 보이도록 이익 나는 소액 포트폴리오로 고정한다.
HEADLINE = {
    "totalAssetAmount": 12_450_000,
    "principalAmount": 11_000_000,
    "evaluatedProfitAmount": 1_450_000,
    "evaluatedAmount": 12_450_000,
    "profitRate": 0.1318,
}


class Scrubber:
    def __init__(self, seed: int = 20260716) -> None:
        self.rng = random.Random(seed)
        self.codes: dict[str, str] = {}

    def code_for(self, original: str) -> str:
        """식별자를 형식 유지한 합성 값으로. 같은 입력 → 같은 출력."""
        if original in self.codes:
            return self.codes[original]

        n = len(self.codes) + 1
        if re.fullmatch(r"[A-Z]{1,5}", original):  # 티커: MSTY
            synthetic = f"SYM{n}"
        elif re.fullmatch(r"[A-Z]{2}[A-Z0-9]{9}\d", original):  # ISIN: US88636X7324
            synthetic = f"US{n:09d}0"
        elif re.match(r"^[A-Z]{2,3}\d", original):  # 종목코드: AMX0240221001
            prefix = re.match(r"^([A-Z]{2,3})", original).group(1)
            synthetic = f"{prefix}20200101{n:03d}"
        elif re.fullmatch(r"\d+", original):  # 숫자 코드
            synthetic = f"{n:0{len(original)}d}"
        else:  # 종목명 등
            synthetic = f"Demo Holding {n}"
        self.codes[original] = synthetic
        return synthetic

    def collect_ids(self, node: Any) -> None:
        """식별자 키의 값을 모아 치환표를 만든다 (키 이름과 무관하게 쓰기 위해)."""
        if isinstance(node, dict):
            for k, v in node.items():
                if isinstance(v, str) and v and v != "<REDACTED>":
                    if ID_KEY.search(k) and len(v) > 1:
                        self.code_for(v)
                    elif NAME_KEY.search(k) and len(v) > 1:
                        self.code_for(v)
                self.collect_ids(v)
        elif isinstance(node, list):
            for item in node:
                self.collect_ids(item)

    def replace_ids(self, text: str) -> str:
        """수집한 식별자를 문자열 어디에 있든 치환 (복합키·URL 포함).
        긴 것부터 바꿔야 짧은 코드가 긴 코드의 일부를 깨뜨리지 않는다."""
        for original in sorted(self.codes, key=len, reverse=True):
            if original in text:
                text = text.replace(original, self.codes[original])
        return text

    def money(self, value: int | float) -> int | float:
        if isinstance(value, float):
            return round(self.rng.uniform(1_000, 500_000), 2)
        return self.rng.randrange(1_000, 5_000_000)

    def walk(self, node: Any) -> Any:
        if isinstance(node, dict):
            return {k: self.field(k, v) for k, v in node.items()}
        if isinstance(node, list):
            return [self.walk(item) for item in node]
        return node

    def field(self, key: str, value: Any) -> Any:
        # 0 / null / bool 은 구조적 의미가 있으므로 보존
        if value is None or isinstance(value, bool) or value == 0:
            return self.walk(value) if isinstance(value, (dict, list)) else value

        if key in HEADLINE and isinstance(value, (int, float)):
            return HEADLINE[key]
        if isinstance(value, str):
            return self.replace_ids(value)
        if isinstance(value, (int, float)):
            if RATE_KEY.search(key):
                return round(self.rng.uniform(-0.15, 0.25), 4)
            if QTY_KEY.search(key):
                return self.rng.randrange(1, 50)
            if MONEY_KEY.search(key):
                return self.money(value)
        return self.walk(value)


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print(__doc__)
        return 2

    target = Path(argv[1])
    files = sorted(p for p in target.glob("*.json") if p.name != "manifest.json")
    if not files:
        print(f"no fixtures in {target}", file=sys.stderr)
        return 1

    for path in files:
        original = json.loads(path.read_text())
        s = Scrubber()
        s.collect_ids(original)  # 1패스: 식별자 수집
        scrubbed = s.walk(original)  # 2패스: 값 치환
        path.write_text(json.dumps(scrubbed, indent=2, ensure_ascii=False) + "\n")
        print(f"scrubbed {path.name}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
