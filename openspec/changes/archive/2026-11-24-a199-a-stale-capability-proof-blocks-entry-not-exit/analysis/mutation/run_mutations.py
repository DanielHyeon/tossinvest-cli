#!/usr/bin/env python3
"""a199 변이 원장(design §3-12) — 사본 worktree 에서 변이 하나씩 적용 → 범위 안 시험 → 원복.

규율(기억: 「뮤테이션은 대상에 닿아야 한다」·「원복은 baseline 이 맞아야 한다」):
  - 사본: `git worktree add --detach <scratch>/a199-mut-<pid> <sha>` — 작업 트리가 아니라 고정 커밋의 사본. 시작 sha 를 단언.
  - 무변이 대조군: 변이 없이 같은 시험 묶음이 GREEN 이어야 시작함(아니면 멈춤).
  - 닿음: 변이 문자열은 대상 파일에서 정확히 1회 적중해야 적용(아니면 INVALID), 적용 뒤 바이트가 달라졌는지 단언.
  - 판정: 시험 rc != 0 이고 출력에 「--- FAIL」 이 있으면 CAUGHT, 빌드 실패면 BUILD(무효 — 다시 설계), rc 0 이면 SURVIVED.
  - 원복: 원본 바이트로 되쓰고 같음을 단언. 끝나면 사본 worktree 를 지우고 없어졌음을 단언.

사용(저장소 루트): python3 <이 파일> <scratch 폴더>
"""

from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

ENGINE = "./internal/app/engine"
MUTANTS = [
    ("M1 만료 시 계좌 확인 삭제", "internal/app/engine/exit_only_start.go",
     "if expired && !att.NamesAccount(accountRef) {", "if false && !att.NamesAccount(accountRef) {", ENGINE, "A199"),
    ("M2 닫힌 집합 넓히기(불완전도 낡음으로)", "internal/app/engine/exit_only_start.go",
     "if !expired && !errors.Is(verr, attest.ErrEndpointNotAttested) {",
     "if !expired && !errors.Is(verr, attest.ErrEndpointNotAttested) && !errors.Is(verr, attest.ErrIncomplete) {", ENGINE, "A199"),
    ("M3 빈 endpoint 항목 가드 삭제", "internal/app/engine/exit_only_start.go",
     'if strings.TrimSpace(e) == "" {', 'if false && strings.TrimSpace(e) == "" {', ENGINE, "A199"),
    ("M4 형식 하한 삭제", "internal/app/engine/exit_only_start.go",
     "if att.FormatVersion < 1 || att.FormatVersion > attest.FormatVersion {",
     "if att.FormatVersion > attest.FormatVersion {", ENGINE, "A199"),
    ("M5 뒤 조항 검사 생략(낡음이면 바로 통과)", "internal/app/engine/interlock.go",
     "\tstale, err := classifyAttestation(att, facts.now, accountRef, RequiredEndpoints())\n\tif err != nil {\n\t\treturn fmt.Errorf(\"%w: %w\", ErrAutomationGateRefused, err)\n\t}\n",
     "\tstale, err := classifyAttestation(att, facts.now, accountRef, RequiredEndpoints())\n\tif err != nil {\n\t\treturn fmt.Errorf(\"%w: %w\", ErrAutomationGateRefused, err)\n\t}\n"
     "\tif stale != nil {\n\t\tstatus.ExitOnly = true\n\t\tstatus.AttestationStale = stale.code()\n\t\treturn nil\n\t}\n", ENGINE, "A199"),
    ("M6 래치 세우기 삭제", "internal/app/engine/exit_only_start.go",
     "\tentry.Block(execgw.ReasonCapabilityAttestationStale, exitOnlyLatchDetail(status))\n",
     "\t_ = execgw.ReasonCapabilityAttestationStale\n", ENGINE, "A199"),
    ("M7 Verified 를 true 로", "internal/app/engine/exit_only_start.go",
     "\tstatus.Verified = false\n", "\tstatus.Verified = true\n", ENGINE, "A199"),
    ("M8 Guardian nil 유지(청산 쪽 판정 대신 Verified)", "internal/app/engine/engine.go",
     "if !automation.RunsExitSide() {", "if !automation.Verified {", ENGINE, "A199"),
    ("M9 감사 action 을 수락으로", "internal/app/engine/exit_only_start.go",
     "Action:  audit.ActionGateExitOnly,", "Action:  audit.ActionGateAccepted,", ENGINE, "A199"),
    ("M10 감시 판정을 따로 두기(Verify 첫 실패 그대로)", "internal/app/engine/attestation_watch.go",
     "\tstale, err := classifyAttestation(att, now, accountRef, required)\n",
     "\tstale, err := func() (*attestationStaleness, error) {\n"
     "\t\tverr := att.Verify(now, accountRef, required)\n"
     "\t\tif errors.Is(verr, attest.ErrExpired) {\n\t\t\treturn &attestationStaleness{expired: true}, nil\n\t\t}\n"
     "\t\tif errors.Is(verr, attest.ErrEndpointNotAttested) {\n\t\t\treturn &attestationStaleness{missing: att.MissingEndpoints(required)}, nil\n\t\t}\n"
     "\t\treturn nil, verr\n\t}()\n", ENGINE, "A199"),
    ("M11 worker 조건에서 Verified 항 삭제", "internal/app/engine/strategy_entry_supervisor.go",
     "c.Journal != nil && c.Gateway != nil && c.Guardian != nil && c.Automation.Verified,",
     "c.Journal != nil && c.Gateway != nil && c.Guardian != nil,", ENGINE, "A199"),
    ("M12 감시 알림 조건을 파일 판정으로", "internal/app/engine/attestation_watch.go",
     "if v.Stage == stageHealthy && !w.exitOnly {", "if v.Stage == stageHealthy {", ENGINE, "A199"),
    ("M13 보호 수명 문장 조건에서 청산 전용 삭제", "internal/app/engine/interlock.go",
     "if status.RunsExitSide() && status.Protection != ProtectionWired {",
     "if status.Verified && status.Protection != ProtectionWired {", ENGINE, "A199"),
    ("M14 exit 관측 생성을 Verified 로", "internal/app/engine/exitwiring.go",
     "if !c.Automation.RunsExitSide() {", "if !c.Automation.Verified {", ENGINE, "A199"),
    ("M15 대사 생성을 Verified 로", "internal/app/engine/reconcileloop.go",
     "if !c.Automation.RunsExitSide() {", "if !c.Automation.Verified {", ENGINE, "A199"),
    ("M16 회복 안내 키 보존 삭제(R2-3)", "internal/app/engine/attestation_watch.go",
     "if strings.HasPrefix(k, string(stageHealthy)+\":\") {", "if false && strings.HasPrefix(k, string(stageHealthy)+\":\") {", ENGINE, "A199"),
    ("M17 낡음 표지 안 세움(verifyGate 끝)", "internal/app/engine/interlock.go",
     "\t\tstatus.ExitOnly = true\n\t\tstatus.AttestationStale = stale.code()\n",
     "\t\tstatus.AttestationStale = stale.code()\n", ENGINE, "A199"),
    ("M18 runEngineRun 이 청산 전용을 게이트 OFF 로", "cmd/tossctl/engine.go",
     "if !ectx.Automation.RunsExitSide() {", "if !ectx.Automation.Verified {", "./cmd/tossctl", "A199"),
    ("M19 latchOrder 에서 새 사유 빼기", "internal/execgw/retry.go",
     "\tReasonCapabilityAttestationStale,\n\t// Appended, per the rule above. The operating mode",
     "\t// Appended, per the rule above. The operating mode", "./internal/execgw", "A199"),
    ("M20 NamesAccount 를 가림 허용 규칙으로", "internal/attest/account_match.go",
     "return sameAccount(a.AccountRef, accountRef)", "return SameAccountMasked(accountRef, a.AccountRef)", "./internal/attest", "A199"),
]


def run(cmd: list[str], cwd: Path) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)


def go_test(copy: Path, pkg: str, pattern: str) -> tuple[int, str]:
    p = run(["go", "test", "-count=1", "-timeout", "30m", "-run", pattern, pkg], copy)
    return p.returncode, p.stdout + p.stderr


def main() -> int:
    repo = Path.cwd()
    scratch = Path(sys.argv[1])
    sha = run(["git", "rev-parse", "HEAD"], repo).stdout.strip()
    copy = scratch / f"a199-mut-{os.getpid()}"
    print(f"start sha {sha}  copy {copy}")
    p = run(["git", "worktree", "add", "--detach", str(copy), sha], repo)
    if p.returncode != 0:
        print(p.stderr)
        return 2
    try:
        got = run(["git", "rev-parse", "HEAD"], copy).stdout.strip()
        assert got == sha, f"copy HEAD {got} != start {sha}"
        for pkg in sorted({m[4] for m in MUTANTS}):
            rc, out = go_test(copy, pkg, "A199")
            print(f"control {pkg}: rc={rc}")
            if rc != 0:
                print(out[-3000:])
                print("STOP: the unmutated control is not GREEN")
                return 3
        for name, rel, old, new, pkg, pattern in MUTANTS:
            path = copy / rel
            original = path.read_bytes()
            text = original.decode("utf-8")
            hits = text.count(old)
            if hits != 1:
                print(f"{name}: INVALID (target hit {hits})")
                continue
            path.write_text(text.replace(old, new, 1), encoding="utf-8")
            assert path.read_bytes() != original, "mutation did not change the file"
            rc, out = go_test(copy, pkg, pattern)
            path.write_bytes(original)
            assert path.read_bytes() == original, "restore failed"
            if rc != 0 and "--- FAIL" in out:
                failed = sorted({ln.split()[2] for ln in out.splitlines() if ln.startswith("--- FAIL")})
                print(f"{name}: CAUGHT by {', '.join(failed)}")
            elif rc != 0:
                print(f"{name}: BUILD/OTHER rc={rc}\n{out[-1500:]}")
            else:
                print(f"{name}: SURVIVED")
        clean = run(["git", "status", "--porcelain"], copy).stdout.strip()
        print(f"copy clean after all restores: {clean == ''}")
    finally:
        run(["git", "worktree", "remove", "--force", str(copy)], repo)
        print(f"copy removed: {not copy.exists()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
