#!/usr/bin/env python3
"""a168 tasks 0.3 ⑥ — 새 거절이 거부할 「정상 입력」 을 먼저 셈.

모집단(저장소 루트 기준, 실행 시각의 워킹트리):
  A. internal/official · internal/verifylive · cmd/tossctl 의 *_test.go 안 문자열 리터럴(백틱 원문 · 큰따옴표 해석)
     중 `{` 또는 `[` 로 시작하는 것.
  B. 같은 세 패키지 아래 testdata/ 의 모든 파일, 그리고 docs/migration/openapi.latest.json 의 example/examples 값
     (스펙 전체를 한 문서로도 한 번 검사).
  C. 같은 세 패키지 *_test.go 의 상태 코드 리터럴(WriteHeader(20x) · http.Status{Created,Accepted,NonAuthoritativeInfo,
     NoContent,ResetContent,PartialContent,MultiStatus,AlreadyReported,IMUsed}) — 200 아닌 2xx 가 응답으로 쓰이는 자리.

판정(새 거절 규칙과 같은 정의):
  - fold 중복: 한 객체 안의 두 키가 대소문자 접기로 같음(완전히 같은 키 포함). 파이썬 str.lower() 와 str.casefold()
    둘 다로 셈(키가 ASCII 면 Go bytes.EqualFold 와 같음 — 비 ASCII 키도 따로 셈).
  - 뒤따르는 데이터: raw_decode 가 값 하나를 읽은 뒤 공백 아닌 바이트가 남음.
  - 해독 불능 리터럴(fmt 서식 %q 등)은 「평가 못 함」 으로 따로 셈 — 0 이 「없음」 이 아니라 「못 봄」 일 수 있음.

사용: python3 -I openspec/changes/a168-verify-read-decode-hardening/analysis/measure/normal_input_census.py
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
PKGS = ["internal/official", "internal/verifylive", "cmd/tossctl"]

RAW = re.compile(r"`([^`]*)`", re.S)
DQ = re.compile(r'"((?:[^"\\\n]|\\.)*)"')
STATUS = re.compile(r"WriteHeader\(\s*(20[1-9]|2[1-9]\d)\s*\)|http\.Status(Created|Accepted|NonAuthoritativeInfo|NoContent|"
                    r"ResetContent|PartialContent|MultiStatus|AlreadyReported|IMUsed)\b")


class Dup(Exception):
    pass


class Variant(Exception):
    pass


# Manager 결정 4(2026-11-25) 술어 — 대사 리더가 읽는 계약 키와 접히지만 바이트가 다른 키. 목록은 구현 전 HEAD 의 리더에서 읽어 적음
# (봉투 apiEnvelope · decodeReconcileEnvelope 의 키 · apiConditionalOrder · apiConditionalCondition · 주문 행 · 종목 항목 · by-id 해독).
# 구현 뒤에는 생산 목록(`internal/official/reconcile_response.go` 의 reconcileContractKeys)과 같은지 아래에서 대조함.
CONTRACT_KEYS = sorted({
    "result", "conditionalOrders", "orders", "hasNext", "nextCursor",
    "conditionalOrderId", "clientOrderId", "type", "status", "symbol", "market", "quantity", "orderType", "expireDate",
    "first", "second", "createdAt",
    "orderSide", "triggerPrice", "targetProfitRate", "orderPrice", "triggeredOrderId",
    "orderId",
})
CONTRACT_LOWER = {c.lower(): c for c in CONTRACT_KEYS}


# 비 ASCII 키 수 — 파이썬 lower()/casefold() 와 Go bytes.EqualFold(단순 접기)·U+FFFD 대체는 비 ASCII 에서 갈린다
# (리뷰 1회차 Codex: `ss`~`ß` 과대 보고, 짝 없는 대리쌍 과소 보고). 평가한 키가 전부 ASCII 면 두 술어가 그 모집단에서 같다.
NON_ASCII_KEYS = []


def pairs_hook(pairs):
    NON_ASCII_KEYS.extend(k for k, _ in pairs if not k.isascii())
    for k, _ in pairs:
        c = CONTRACT_LOWER.get(k.lower())
        if c is not None and c != k:
            raise Variant(f"{k!r} ~ contract {c!r}")
    seen_lower, seen_fold = {}, {}
    for k, _ in pairs:
        for seen, key in ((seen_lower, k.lower()), (seen_fold, k.casefold())):
            if key in seen:
                raise Dup(f"{seen[key]!r} ~ {k!r}")
            seen[key] = k
    return dict(pairs)


def judge(text: str):
    """(verdict, detail) — verdict ∈ ok · dup · trailing · unparsed."""
    dec = json.JSONDecoder(object_pairs_hook=pairs_hook)
    s = text.strip()
    try:
        _, end = dec.raw_decode(s)
    except Dup as d:
        return "dup", str(d)
    except Variant as v:
        return "variant", str(v)
    except ValueError:
        return "unparsed", ""
    if s[end:].strip():
        return "trailing", s[end:end + 20]
    return "ok", ""


def go_literals(path: Path):
    src = path.read_text(encoding="utf-8")
    for m in RAW.finditer(src):
        yield src.count("\n", 0, m.start()) + 1, m.group(1)
    for m in DQ.finditer(src):
        try:
            val = json.loads('"' + m.group(1) + '"')  # Go 해석 문자열의 흔한 이스케이프는 JSON 과 같음
        except ValueError:
            continue
        yield src.count("\n", 0, m.start()) + 1, val


def examples(node, out):
    if isinstance(node, dict):
        for k, v in node.items():
            if k == "example":
                out.append(v)
            elif k == "examples" and isinstance(v, dict):
                for ex in v.values():
                    out.append(ex.get("value") if isinstance(ex, dict) and "value" in ex else ex)
            else:
                examples(v, out)
    elif isinstance(node, list):
        for v in node:
            examples(v, out)


def main() -> int:
    counts = {"ok": 0, "dup": 0, "variant": 0, "trailing": 0, "unparsed": 0}
    findings = []
    files = 0
    for pkg in PKGS:
        for path in sorted((ROOT / pkg).rglob("*_test.go")):
            files += 1
            for line, lit in go_literals(path):
                if not lit.strip().startswith(("{", "[")):
                    continue
                v, d = judge(lit)
                counts[v] += 1
                if v in ("dup", "variant", "trailing"):
                    findings.append(f"A {path.relative_to(ROOT)}:{line} {v} {d}")
    print(f"A. *_test.go 파일 {files} · JSON 후보 리터럴 {sum(counts.values())} → {counts}")

    tcounts = {"ok": 0, "dup": 0, "variant": 0, "trailing": 0, "unparsed": 0}
    tfiles = 0
    for pkg in PKGS:
        for path in sorted(p for p in (ROOT / pkg).rglob("*") if p.is_file() and "testdata" in p.parts):
            tfiles += 1
            try:
                text = path.read_text(encoding="utf-8")
            except UnicodeDecodeError:
                tcounts["unparsed"] += 1
                continue
            texts = [text]
            if path.suffix == ".jsonl":
                texts = [ln for ln in text.splitlines() if ln.strip()]
            for t in texts:
                if not t.strip().startswith(("{", "[")):
                    tcounts["unparsed"] += 1
                    continue
                v, d = judge(t)
                tcounts[v] += 1
                if v in ("dup", "variant", "trailing"):
                    findings.append(f"B {path.relative_to(ROOT)} {v} {d}")
    print(f"B1. testdata 파일 {tfiles} · 문서 {sum(tcounts.values())} → {tcounts}")

    spec = (ROOT / "docs/migration/openapi.latest.json").read_text(encoding="utf-8")
    v, d = judge(spec)
    print(f"B2. openapi.latest.json 문서 전체 → {v} {d}")
    # B4. 응답 스키마의 속성 이름(브로커 응답 키의 계약) — B2 의 문서 전체 검사는 components 이름(`parameters.Symbol` 등 문서 구조 키)까지
    # 세므로, 브로커가 실제로 보낼 수 있는 키는 `properties` 의 이름으로 따로 셈.
    props, pvariant = 0, []

    def walk_props(node):
        nonlocal props
        if isinstance(node, dict):
            for k, v in node.items():
                if k == "properties" and isinstance(v, dict):
                    for name in v:
                        props += 1
                        c = CONTRACT_LOWER.get(name.lower())
                        if c is not None and c != name:
                            pvariant.append(f"{name} ~ {c}")
                walk_props(v)
        elif isinstance(node, list):
            for v in node:
                walk_props(v)

    walk_props(json.loads(spec))
    print(f"B4. openapi 응답·요청 스키마 속성 이름 {props} 중 계약 키 변형 {len(pvariant)} {pvariant[:5]}")
    exs = []
    examples(json.loads(spec), exs)
    ecounts = {"ok": 0, "dup": 0, "variant": 0, "trailing": 0, "unparsed": 0}
    for ex in exs:
        text = ex if isinstance(ex, str) else json.dumps(ex)
        if not text.strip().startswith(("{", "[")):
            continue
        ev, ed = judge(text)
        ecounts[ev] += 1
        if ev in ("dup", "variant", "trailing"):
            findings.append(f"B2 example {ev} {ed}")
    print(f"B3. openapi example 값 {len(exs)} 중 JSON 객체·배열 → {ecounts}"
          " (주의: json.loads 로 한 번 읽은 뒤라 원문 중복 키는 B2 가 봄)")

    # D. 저장소 전체(.git·.sdd·node_modules 제외)의 *.json·*.jsonl 중 브로커 응답 모양 키를 담은 문서(영수증·골든).
    dcounts = {"ok": 0, "dup": 0, "variant": 0, "trailing": 0, "unparsed": 0}
    dfiles = 0
    shape = re.compile(r'"(conditionalOrderId|orderId|conditionalOrders|orders|result)"')
    for path in sorted(ROOT.rglob("*")):
        if path.suffix not in (".json", ".jsonl") or not path.is_file():
            continue
        if any(p in (".git", ".sdd", "node_modules", ".codegraph") for p in path.relative_to(ROOT).parts):
            continue
        try:
            text = path.read_text(encoding="utf-8")
        except (UnicodeDecodeError, OSError):
            continue
        if not shape.search(text):
            continue
        dfiles += 1
        texts = [ln for ln in text.splitlines() if ln.strip()] if path.suffix == ".jsonl" else [text]
        for t in texts:
            v, d = judge(t)
            dcounts[v] += 1
            if v in ("dup", "variant", "trailing"):
                findings.append(f"D {path.relative_to(ROOT)} {v} {d}")
    print(f"D. 저장소 *.json·*.jsonl 중 브로커 응답 모양 키를 담은 파일 {dfiles} · 문서 {sum(dcounts.values())} → {dcounts}")

    go_list = ROOT / "internal/official/reconcile_response.go"
    if go_list.is_file():
        m = re.search(r"reconcileContractKeys\s*=\s*\[\]string\{(.*?)\}", go_list.read_text(encoding="utf-8"), re.S)
        got = sorted(set(re.findall(r'"([^"]+)"', m.group(1)))) if m else None
        print(f"F. 계약 키 목록 대조(생산 reconcileContractKeys) → {'같음' if got == CONTRACT_KEYS else f'다름 {got}'}")
    else:
        print(f"F. 계약 키 목록 — 생산 파일 없음(구현 전), 측정 목록 {len(CONTRACT_KEYS)}개 사용")

    sfind = []
    for pkg in PKGS:
        for path in sorted((ROOT / pkg).rglob("*_test.go")):
            for i, ln in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
                if STATUS.search(ln):
                    sfind.append(f"C {path.relative_to(ROOT)}:{i} {ln.strip()[:140]}")
    print(f"C. 200 아닌 2xx 상태 리터럴 자리 {len(sfind)}")
    print(f"E. 평가한 객체의 비 ASCII 키 {len(NON_ASCII_KEYS)} {sorted(set(NON_ASCII_KEYS))[:10]}"
          " — 0 이면 파이썬 접기 술어와 Go EqualFold 가 이 모집단에서 같음")
    for f in findings + sfind:
        print("  " + f)
    return 0


if __name__ == "__main__":
    sys.exit(main())
