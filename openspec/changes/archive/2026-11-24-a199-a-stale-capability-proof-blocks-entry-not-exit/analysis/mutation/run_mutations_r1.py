#!/usr/bin/env python3
"""a199 변이 원장 추가분(구현 리뷰 r1 처분으로 바뀐 줄과 새 시험) — run_mutations.py 와 같은 규율(사본 · 시작 sha · 대조군 · 닿음 · 원복 · 제거).

사용(저장소 루트): python3 <이 파일> <scratch 폴더>
"""

from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

ENGINE = "./internal/app/engine"
# (이름, 파일, 옛 문자열, 새 문자열, 패키지, -run, tags)
MUTANTS = [
    ("M21 nil 진입 게이트 가드 삭제", "internal/app/engine/exit_only_start.go",
     "\tif entry == nil {\n", "\tif false {\n", ENGINE, "A199", ""),
    ("M22 청산 전용 감사 실패 무시", "internal/app/engine/exit_only_start.go",
     "\t\treturn status, fmt.Errorf(\"engine: recording the exit-only start: %w\", err)\n",
     "\t\treturn status, nil\n", ENGINE, "A199", ""),
    ("M23 gate_exit_only 를 늘 싣기", "internal/app/engine/interlock.go",
     "\t\t\"entry_permitted\", status.EntryPermitted,\n",
     "\t\t\"entry_permitted\", status.EntryPermitted,\n\t\t\"gate_exit_only\", status.ExitOnly,\n", ENGINE, "A199", ""),
    ("M24 래치 세우기 삭제 — 전략 dispatch 시험만으로", "internal/app/engine/exit_only_start.go",
     "\tentry.Block(execgw.ReasonCapabilityAttestationStale, exitOnlyLatchDetail(status))\n",
     "\t_ = execgw.ReasonCapabilityAttestationStale\n", ENGINE, "TestA199TheStaleLatchStopsTheRealStrategyDispatch", "tossos_testseams"),
    ("M25 래치 세우기 삭제 — 런타임 시험 전체", "internal/app/engine/exit_only_start.go",
     "\tentry.Block(execgw.ReasonCapabilityAttestationStale, exitOnlyLatchDetail(status))\n",
     "\t_ = execgw.ReasonCapabilityAttestationStale\n", ENGINE, "A199", ""),
    ("M26 도달 불가 조항 줄을 되살려도 시험 무변(대조 — 삭제가 행동을 안 바꿈)", "internal/app/engine/runtime_wiring.go",
     "\t{attest.ErrAccountMismatch, \"4. capability attestation — 다른 계좌의 것이다\"},\n",
     "\t{attest.ErrAccountMismatch, \"4. capability attestation — 다른 계좌의 것이다\"},\n"
     "\t{attest.ErrExpired, \"4. capability attestation — 만료되었다\"},\n", ENGINE, "A199|TestEachInterlockClauseHasALine", ""),
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
    sha = run(["git", "rev-parse", "HEAD"], repo).stdout.strip()
    copy = scratch / f"a199-mut-r1-{os.getpid()}"
    print(f"start sha {sha}  copy {copy}")
    if run(["git", "worktree", "add", "--detach", str(copy), sha], repo).returncode != 0:
        return 2
    try:
        assert run(["git", "rev-parse", "HEAD"], copy).stdout.strip() == sha
        for pattern, tags in sorted({(m[5], m[6]) for m in MUTANTS}):
            rc, out = go_test(copy, ENGINE, pattern, tags)
            print(f"control {pattern} tags={tags or '-'}: rc={rc}")
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
                print(f"{name}: BUILD/OTHER rc={rc}\n{out[-1500:]}")
            else:
                print(f"{name}: SURVIVED")
        print(f"copy clean after all restores: {run(['git', 'status', '--porcelain'], copy).stdout.strip() == ''}")
    finally:
        run(["git", "worktree", "remove", "--force", str(copy)], repo)
        print(f"copy removed: {not copy.exists()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
