#!/usr/bin/env python3
"""a199 변이 원장 추가분(G2 · G3 · G9) — run_mutations.py · run_mutations_r1.py 와 같은 규율
(사본 worktree · 시작 sha 단언 · 무변이 대조군 GREEN · 대상 1회 적중 · 원복 바이트 단언 · 사본 제거).
빌드 실패는 CAUGHT 로 세지 않음(BUILD/OTHER 로 따로 적음).

사용(저장소 루트): python3 <이 파일> <scratch 폴더> [변이 이름 접두]
"""

from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

ENGINE = "./internal/app/engine"
JOURNAL = "./internal/journal"
EXIT_ONLY = "internal/app/engine/exit_only_start.go"
EVIDENCE = "internal/journal/a199_order_path_evidence.go"
SQL_WHERE = "WHERE a.state = ? AND a.kind = ? AND i.source = ? AND TRIM(i.account_ref) = ? LIMIT 1"

# (이름, 파일, 옛 문자열, 새 문자열, 패키지, -run, tags)
MUTANTS = [
    ("M27 증거 검사 삭제(낡으면 늘 내림)", EXIT_ONLY,
     # 첫 판의 "if false" 는 proven 미사용으로 빌드 실패(BUILD/OTHER — CAUGHT 아님, run-g2.txt) → 컴파일되는 같은 뜻으로 재실행(run-g2-m27.txt).
     "\tif readErr != nil || !proven {\n", "\tif (readErr != nil || !proven) && false {\n", ENGINE, "A199", ""),
    ("M28 계좌 조건 삭제", EVIDENCE, SQL_WHERE,
     "WHERE a.state = ? AND a.kind = ? AND i.source = ? AND (TRIM(i.account_ref) = ? OR 1 = 1) LIMIT 1", JOURNAL, "A199", ""),
    ("M29 source 조건 삭제", EVIDENCE, SQL_WHERE,
     "WHERE a.state = ? AND a.kind = ? AND (i.source = ? OR 1 = 1) AND TRIM(i.account_ref) = ? LIMIT 1", JOURNAL, "A199", ""),
    ("M30 상태 조건을 CONFIRMED 밖으로 넓힘", EVIDENCE, SQL_WHERE,
     "WHERE (a.state = ? OR a.state <> '') AND a.kind = ? AND i.source = ? AND TRIM(i.account_ref) = ? LIMIT 1", JOURNAL, "A199", ""),
    ("M31 kind 조건 삭제(CANCEL 도 셈)", EVIDENCE, SQL_WHERE,
     "WHERE a.state = ? AND (a.kind = ? OR 1 = 1) AND i.source = ? AND TRIM(i.account_ref) = ? LIMIT 1", JOURNAL, "A199", ""),
    ("M32 읽기 실패 시 내림", EXIT_ONLY,
     "\tif readErr != nil || !proven {\n", "\tif readErr == nil && !proven {\n", ENGINE, "A199", ""),
    ("M33 유효 경로에서도 원장 읽기", EXIT_ONLY,
     "\tverr := att.Verify(now, accountRef, required)\n",
     "\tif proof != nil {\n\t\t_, _ = proof(accountRef)\n\t}\n\tverr := att.Verify(now, accountRef, required)\n", ENGINE, "A199", ""),
    ("M34 감시의 기한 뒤 예측을 증거와 무관하게(늘 청산 전용)", "internal/app/engine/attestation_watch.go",
     "\tif v.Stage != stageHealthy {\n\t\t// 기한 뒤 재시작의 귀결 예측", "\tif false {\n\t\t// 기한 뒤 재시작의 귀결 예측", ENGINE, "A199", ""),
    ("M35 G9 거절 줄의 청산 전용 표지 지우기 삭제", EXIT_ONLY,
     "\t\tstatus.ExitOnly, status.AttestationStale, status.AttestationMissing = false, \"\", \"\"\n", "", ENGINE, "A199", ""),
    ("M24' 래치 세우기 삭제 — 전략 dispatch 시험만(G3 로 exitOnlyStart 를 거침)", EXIT_ONLY,
     "\tentry.Block(execgw.ReasonCapabilityAttestationStale, exitOnlyLatchDetail(status))\n",
     "\t_ = execgw.ReasonCapabilityAttestationStale\n", ENGINE, "TestA199TheStaleLatchStopsTheRealStrategyDispatch", "tossos_testseams"),
    ("M36 래치 세우기 삭제 — 조립된 게이트웨이 시험만(G3)", EXIT_ONLY,
     "\tentry.Block(execgw.ReasonCapabilityAttestationStale, exitOnlyLatchDetail(status))\n",
     "\t_ = execgw.ReasonCapabilityAttestationStale\n", ENGINE, "TestA199TheBootLatchIsTheGateTheAssembledGatewayChecks", ""),
]


def run(cmd: list[str], cwd: Path) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)


def go_test(copy: Path, pkg: str, pattern: str, tags: str) -> tuple[int, str]:
    cmd = ["go", "test", "-count=1", "-timeout", "30m"]
    if tags:
        cmd += ["-tags", tags]
    p = run(cmd + ["-run", pattern, pkg], copy)
    return p.returncode, p.stdout + p.stderr


def main() -> int:
    repo = Path.cwd()
    scratch = Path(sys.argv[1])
    if len(sys.argv) > 2:  # 이름 접두로 골라 다시 돌림(예: M27)
        MUTANTS[:] = [m for m in MUTANTS if m[0].startswith(sys.argv[2])]
    sha = run(["git", "rev-parse", "HEAD"], repo).stdout.strip()
    copy = scratch / f"a199-mut-g2-{os.getpid()}"
    print(f"start sha {sha}  copy {copy}")
    if run(["git", "worktree", "add", "--detach", str(copy), sha], repo).returncode != 0:
        return 2
    try:
        assert run(["git", "rev-parse", "HEAD"], copy).stdout.strip() == sha
        for pkg, pattern, tags in sorted({(m[4], m[5], m[6]) for m in MUTANTS}):
            rc, out = go_test(copy, pkg, pattern, tags)
            print(f"control {pkg} {pattern} tags={tags or '-'}: rc={rc}")
            if rc != 0:
                print(out[-3000:])
                print("STOP: the unmutated control is not GREEN")
                return 3
        for name, rel, old, new, pkg, pattern, tags in MUTANTS:
            path = copy / rel
            original = path.read_bytes()
            text = original.decode("utf-8")
            if text.count(old) != 1:
                print(f"{name}: INVALID (target hit {text.count(old)})")
                continue
            path.write_text(text.replace(old, new, 1), encoding="utf-8")
            assert path.read_bytes() != original
            rc, out = go_test(copy, pkg, pattern, tags)
            path.write_bytes(original)
            assert path.read_bytes() == original
            if rc != 0 and "--- FAIL" in out:
                failed = sorted({ln.split()[2] for ln in out.splitlines() if ln.startswith("--- FAIL")})
                print(f"{name}: CAUGHT by {', '.join(failed)}")
            elif rc != 0:
                print(f"{name}: BUILD/OTHER rc={rc} (not counted as caught)\n{out[-1500:]}")
            else:
                print(f"{name}: SURVIVED")
        print(f"copy clean after all restores: {run(['git', 'status', '--porcelain'], copy).stdout.strip() == ''}")
    finally:
        run(["git", "worktree", "remove", "--force", str(copy)], repo)
        print(f"copy removed: {not copy.exists()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
