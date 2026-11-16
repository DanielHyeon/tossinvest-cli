#!/usr/bin/env python3
"""a187 tasks 3.1 사본 변이 하네스.

HEAD 를 git archive 로 pid 붙은 임시 디렉터리에 풀고 그 사본에서만 변이함(작업 트리는 건드리지 않음).
무변이 대조군이 GREEN 이 아니면 멈춤. 변이마다 대상 문자열이 정확히 1회인지 단언 → 적용 → 시험 → 원복 후 SHA-256 단언.
환경: GOFLAGS=-trimpath, 전용 GOCACHE. 무변이 대조군 1차(history-1-control-trimpath.txt)에서 실패한 정적 가드 3개를
-skip 으로 뺌. 둘(TestStrictMinuteCandlesFilesCarryNoSeamIdentifier·TestStrictMinutePageLiteralLivesOnlyInItsOwnFile)은
-trimpath 에서 runtime.Caller 가 모듈 경로를 돌려줘 소스 파일을 못 찾는 것이고, 하나
(TestA112MBUSStaticNoProductCallerOrForbiddenPath)는 base 6dbb752f6 에서도 실패하는 기존 결함임(analysis/verify/base-failures.txt).
셋 다 종목 문법과 무관함. 앞의 둘은 -trimpath 없는 3.2 패키지 시험에서 통과함.
사용: python3 -I mutation.py [ID…]
"""
from __future__ import annotations

import hashlib
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

TARGET = "internal/official/strict_minute_candles.go"
PACKAGES = ["./internal/official/", "./internal/orderintent/"]
SKIP = "^(TestStrictMinuteCandlesFilesCarryNoSeamIdentifier|TestStrictMinutePageLiteralLivesOnlyInItsOwnFile|TestA112MBUSStaticNoProductCallerOrForbiddenPath)$"
WRAPPER = "func strictMinuteCheckCandleSymbol(market, symbol string) error {\n\treturn strictMinuteCheckSymbol(market, symbol)\n}"
QUOTES = "internal/official/strict_quote_reads.go"
# (ID, 변이할 파일, 원문, 변이, 설명)
MUTANTS = [
    ("numeric-only", TARGET, "`^[A-Z0-9]{6}$`", "`^[0-9]{6}$`", "숫자 전용으로 되돌림"),
    ("no-length", TARGET, "`^[A-Z0-9]{6}$`", "`^[A-Z0-9]+$`", "길이 고정 제거"),
    ("no-start-anchor", TARGET, "`^[A-Z0-9]{6}$`", "`[A-Z0-9]{6}$`", "앞 고정 제거"),
    ("no-end-anchor", TARGET, "`^[A-Z0-9]{6}$`", "`^[A-Z0-9]{6}`", "뒤 고정 제거"),
    ("lowercase", TARGET, "`^[A-Z0-9]{6}$`", "`^[A-Za-z0-9]{6}$`", "소문자 허용"),
    ("wrapper-skips-check", TARGET, WRAPPER, WRAPPER.replace("return strictMinuteCheckSymbol(market, symbol)", "return nil"), "래퍼가 검사를 건너뜀"),
    # 구현 리뷰 r1(gpt-6-astra)이 「생존 예상」으로 짚은 셋
    ("r1-leading-digits", TARGET, "`^[A-Z0-9]{6}$`", "`^[0-9]{4}[A-Z0-9]{2}$`", "r1 앞 네 자리를 숫자로 축소"),
    ("r1-multiline", TARGET, "`^[A-Z0-9]{6}$`", "`(?m)^[A-Z0-9]{6}$`", "r1 여러 줄 모드(개행 허용)"),
    ("r1-echo-skip-krw", QUOTES, "\tif echoed != symbol {\n", "\tif echoed != symbol && currency != \"KRW\" {\n", "r1 KR 현재가 메아리 검사 건너뜀"),
    # gstack 리뷰 발견 2·7
    ("g-wrapper-own-regex", TARGET, WRAPPER,
     WRAPPER.replace("\treturn strictMinuteCheckSymbol(market, symbol)",
                     "\tif market == \"KR\" && regexp.MustCompile(`^[A-Z0-9]{6}$|^[a-z0-9]{6}$`).MatchString(symbol) {\n\t\treturn nil\n\t}\n\treturn strictMinuteCheckSymbol(market, symbol)"),
     "gstack 래퍼가 자기 정규식을 다시 들여 공용 검사와 갈라짐"),
    ("g-order-inference-widened", "internal/orderintent/intent.go",
     "\tif intent.Market == \"us\" && looksLikeKRSymbol(intent.Symbol) {\n",
     "\tif intent.Market == \"us\" && regexp.MustCompile(`^A?[A-Z0-9]{6}$`).MatchString(intent.Symbol) {\n",
     "gstack 주문 시장 추론 호출 자리를 영숫자 6자리로 넓힘"),
]


def sha(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


def run_tests(base: Path, verbose: bool = False) -> tuple[int, str]:
    # 변이 판에서는 -v 를 끔: go test 는 그때 실패한 시험의 출력만 찍으므로, 통과한 시험의 t.Log 가
    # 「실패 원문」에 섞이지 않음(gstack 발견 16). 대조군만 -v 로 PASS 줄을 셈.
    env = dict(os.environ, GOFLAGS="-trimpath", GOCACHE="/tmp/a187-mut-gocache")
    flags = ["-v"] if verbose else []
    proc = subprocess.run(["go", "test", "-count=1", *flags, "-skip", SKIP, *PACKAGES], cwd=base, env=env,
                          stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    return proc.returncode, proc.stdout


def summary(out: str) -> str:
    fails = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL")]
    msgs = []
    for l in out.splitlines():
        if re.match(r"^\s+\S+_test\.go:\d+: ", l) and len(msgs) < 6:
            msgs.append(l.strip())
    build = [l for l in out.splitlines() if l.startswith("#") or ": undefined" in l or "declared and not used" in l][:4]
    return "\n".join(["  잡은 시험:"] + ["    " + f for f in fails] + ["  실패 원문(앞 6줄):"] + ["    " + m for m in msgs] + (["  빌드:"] + ["    " + b for b in build] if build else []))


def main() -> int:
    root = Path(subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip())
    head = subprocess.check_output(["git", "-C", str(root), "rev-parse", "HEAD"], text=True).strip()
    wanted = set(sys.argv[1:])
    base = Path(tempfile.mkdtemp(prefix=f"a187-mut-{os.getpid()}-"))
    try:
        archive = subprocess.run(["git", "-C", str(root), "archive", "HEAD", "go.mod", "go.sum", "internal", "cmd", "tools", "docs/migration"],
                                 check=True, stdout=subprocess.PIPE).stdout
        subprocess.run(["tar", "-x", "-C", str(base)], input=archive, check=True)
        assert (base / "go.mod").is_file() and str(base).startswith(tempfile.gettempdir())
        files = sorted({m[1] for m in MUTANTS})
        originals = {f: (base / f).read_text() for f in files}
        starts = {f: sha(base / f) for f in files}
        for f in files:
            assert starts[f] == sha(root / f), f"작업 트리 {f} 가 HEAD 와 다름 — 커밋 뒤 돌릴 것"
        print(f"기준 HEAD {head}, 사본 {base}")
        for f in files:
            print(f"  {f} sha256={starts[f]}")
        status, out = run_tests(base, verbose=True)
        if status != 0:
            print(out)
            print("control: FAIL — 변이 판정 중단")
            return 1
        print(f"control: GREEN ({out.count('--- PASS')} PASS 줄, -skip {SKIP})")
        issues = 0
        for mid, fname, old, new, label in MUTANTS:
            if wanted and mid not in wanted:
                continue
            target, original, start = base / fname, originals[fname], starts[fname]
            n = original.count(old)
            if n != 1:
                print(f"{mid}: TARGET-COUNT {n} — {label}")
                issues += 1
                continue
            target.write_text(original.replace(old, new, 1))
            status, out = run_tests(base)
            target.write_text(original)
            assert sha(target) == start
            if status == 0:
                print(f"{mid}: SURVIVED — {label}")
                issues += 1
            elif "[build failed]" in out or "setup failed" in out:
                print(f"{mid}: INVALID — {label}")
                print(summary(out))
                issues += 1
            else:
                print(f"{mid}: CAUGHT — {label}")
                print(summary(out))
        return 1 if issues else 0
    finally:
        shutil.rmtree(base, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
