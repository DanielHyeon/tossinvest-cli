#!/usr/bin/env python3
"""a112 8.7.2 — 사본 뮤테이션 하네스.

쓰는 법: a872_mutate.py <작업 디렉터리>

- 저장소의 `go.mod`·`go.sum`·`internal/`·`cmd/`·`tools/` 를 `<작업>/sandbox.<pid>` 로 복사한다.
  실제 저장소는 한 바이트도 건드리지 않는다(시작·끝에 대상 파일 해시를 대조한다).
- **무변이 대조군을 먼저 돈다.** 초록이 아니면 멈춘다 — 사본이 깨져 있으면 모든 변이가
  "CAUGHT" 로 찍힌다.
- 변이마다 (파일, 옛 문자열, 새 문자열) 목록을 순서대로 적용한다. 옛 문자열은 정확히 한 번
  나와야 한다(안 닿는 변이는 아무것도 재지 않는다). 적용 뒤 원본을 되돌려 놓는다.
- 판정: `go vet` 가 실패하면 BUILD(변이가 컴파일되지 않음 = 잰 것이 아님), 시험이 실패하면
  CAUGHT, 통과하면 SURVIVED.
- 환경: `GOFLAGS=-trimpath`, 전용 `GOCACHE`(작업 디렉터리 아래).
"""

from __future__ import annotations

import hashlib
import os
import shutil
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parents[5]
R = "internal/strategyrouter/production_family_activation.go"
E = "internal/app/engine/strategy_family_activation.go"
D = "internal/app/engine/strategy_dispatch_cycle.go"
PACKAGES = ["./internal/strategyrouter/", "./internal/app/engine/"]
# -trimpath 아래에서 runtime.Caller 로 자기 소스를 읽는 a111 시험 둘은 경로를 못 찾아 실패한다(측정).
# 이 로트의 대상과 무관하므로 건너뛴다 — 건너뛰지 않으면 무변이 대조군이 빨갛다.
TRIMPATH_BLIND = ("^(TestA111ObserverUsesClockLeaseHelpersForTheUseLease|"
                  "TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration)$")

UNDECLARED = '''	if strings.TrimSpace(config.ManifestDigest) == "" {
		return FamilyActivation{}, ErrProductionFamilyActivationUndeclared
	}
'''
CTX_ERR = '''	if err := ctx.Err(); err != nil {
		return FamilyActivation{}, err
	}
	config.ConfigDir'''
CEILING_CALL = '''	leaseCeiling, err := family.LeaseCeiling(cycle.clockNow(), 30*time.Second)
	if err != nil {
		return execgw.Outcome{}, err
	}
'''
FINAL_CHECK = '''			_, err := family.LeaseCeiling(cycle.clockNow(), 30*time.Second)
			return err
'''
REVALIDATE = '''			if err := cycle.revalidateSchedule(checkCtx, market, schedule); err != nil {
				return err
			}
'''
S = "internal/app/engine/strategy_entry_supervisor.go"

MUTANTS: dict[str, list[tuple[str, str, str]]] = {
    "M1 미선언 판정 삭제": [(R, 'if strings.TrimSpace(config.ManifestDigest) == "" {', 'if false && strings.TrimSpace(config.ManifestDigest) == "" {')],
    "M2 미선언 판정을 ctx 검사 뒤로": [(R, UNDECLARED, ""), (R, CTX_ERR, CTX_ERR.replace("\tconfig.ConfigDir", UNDECLARED + "\tconfig.ConfigDir"))],
    "M3 LeaseCeiling 이 상한을 그대로": [(R, "return min(ceiling, remaining), nil", "return ceiling + 0*remaining, nil")],
    "M4 LeaseCeiling 이 수명을 그대로": [(R, "return min(ceiling, remaining), nil", "return remaining + 0*ceiling, nil")],
    "M5 미검증 활성화가 상한을 0 으로": [(R, "if !activation.Verified() {\n\t\treturn ceiling, nil", "if !activation.Verified() {\n\t\treturn 0, nil")],
    "M6 만료 경계를 1ns 뒤로": [(R, "if !now.Before(expires) {\n\t\treturn 0, ErrProductionFamilyActivationExpired", "if now.After(expires) {\n\t\treturn 0, ErrProductionFamilyActivationExpired")],
    "M7 적재가 만료를 따로 판정(동치 사본)": [(R, "if _, err := familyActivationRemaining(expires, now); err != nil {\n\t\treturn nil, err\n\t}", "if !now.Before(expires) {\n\t\treturn nil, ErrProductionFamilyActivationExpired\n\t}")],
    "M8 모든 적재 오류를 기존 경로로(편집 전 동작)": [(E, "if errors.Is(err, strategyrouter.ErrProductionFamilyActivationUndeclared) {", "if err != nil || errors.Is(err, strategyrouter.ErrProductionFamilyActivationUndeclared) {")],
    "M9 되돌림 표식 빠짐": [(E, "return strategyFamilyGate{lanes: lanes, rolledBack: true}", "return strategyFamilyGate{lanes: lanes}")],
    "M10 installed 가 되돌림을 안 봄": [(E, "return gate.rolledBack || gate.activation.Verified()\n", "return gate.activation.Verified()\n")],
    "M11 검증된 관문에 레인 조건 복원(8.7.1 모양)": [(E, "return gate.rolledBack || gate.activation.Verified()\n", "return gate.rolledBack || gate.activation.Verified() && len(gate.lanes) != 0\n")],
    "M11b 되돌림에도 레인 조건": [(E, "return gate.rolledBack || gate.activation.Verified()\n", "return (gate.rolledBack || gate.activation.Verified()) && len(gate.lanes) != 0\n")],
    "M12 보정 조기 반환 복원": [(E, "calibration, _ := strategyMarketCalibrationDigest(routes)", "calibration, agreed := strategyMarketCalibrationDigest(routes)\n\tif !agreed {\n\t\treturn strategyrouter.FamilyActivation{}, strategyrouter.ErrProductionFamilyActivationUnavailable\n\t}")],
    "M13 dispatch 가 상한을 안 씀": [(D, "ttl := min(leaseCeiling, activationExpiresAt.Sub(now))", "ttl := min(30*time.Second+0*leaseCeiling, activationExpiresAt.Sub(now))")],
    "M14 dispatch 가 만료 오류를 버림": [(D, CEILING_CALL, CEILING_CALL.replace("leaseCeiling, err :=", "leaseCeiling, _ :=").replace("\tif err != nil {\n\t\treturn execgw.Outcome{}, err\n\t}\n", ""))],
    "M15 admission 앞 검사가 벽시계를 넘김": [(D, "leaseCeiling, err := family.LeaseCeiling(cycle.clockNow(),", "leaseCeiling, err := family.LeaseCeiling(time.Now(),")],
    "M16 dispatch 가 admission 뒤에서 판정": [(D, CEILING_CALL, ""), (D, "\tttl := min(leaseCeiling,", CEILING_CALL + "\tttl := min(leaseCeiling,")],
    "M17 무오류·미검증을 기존 경로로": [(E, "if err != nil || !activation.Verified() {\n\t\treturn strategyFamilyGate{lanes: lanes, rolledBack: true}", "if err != nil {\n\t\treturn strategyFamilyGate{lanes: lanes, rolledBack: true}")],
    "M18 보정 결속의 형식 검사 삭제(조기 반환 제거 뒤 그 입력을 막는 문)": [(R, "\t\t!productionRouteIdentity(config.CalibrationDigest) ||\n", "")],
    "M19 최종 검사에서 가족 검사 제거": [(D, FINAL_CHECK, "\t\t\treturn nil\n")],
    "M20 최종 검사가 파도 시각을 넘김": [(D, FINAL_CHECK, FINAL_CHECK.replace("cycle.clockNow()", "cycle.schedule.observedAt"))],
    "M21 admission 앞 검사가 파도 시각을 넘김": [(D, "leaseCeiling, err := family.LeaseCeiling(cycle.clockNow(),", "leaseCeiling, err := family.LeaseCeiling(cycle.schedule.observedAt,")],
    "M22 시계 없음 거절 삭제": [(D, "if family.Verified() && cycle.now == nil {", "if false && family.Verified() && cycle.now == nil {")],
    "M23 생산 조립이 시계를 안 넣음": [(S, "\t\tdispatchCycle.now = clk.Now\n", "")],
    "M24 가족 검사를 스케줄 재검증 앞으로(Codex 재리뷰 P1 이전 모양)": [(D, REVALIDATE + FINAL_CHECK, "\t\t\tif _, err := family.LeaseCeiling(cycle.clockNow(), 30*time.Second); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n\t\t\treturn cycle.revalidateSchedule(checkCtx, market, schedule)\n")],
    # 적대 리뷰 B 가 연 변이 넷 (그때 셋이 살아남았다).
    "RB1 admission 앞 상한 30초→60초": [(D, "leaseCeiling, err := family.LeaseCeiling(cycle.clockNow(), 30*time.Second)", "leaseCeiling, err := family.LeaseCeiling(cycle.clockNow(), 60*time.Second)")],
    "RB2 가족 활성화를 KR 로 고정": [(D, "family := cycle.proposals.forMarket(market).familyActivation()", "family := cycle.proposals.forMarket(StrategyMarketKR).familyActivation()")],
    "RB3 취소된 주기를 기존 경로로": [(E, "if errors.Is(err, strategyrouter.ErrProductionFamilyActivationUndeclared) {", "if errors.Is(err, strategyrouter.ErrProductionFamilyActivationUndeclared) || errors.Is(err, context.Canceled) {")],
    "RB7 되돌림 경로가 레인 상태를 씀": [(E, "\t\tcycle := lane.Run(gate.activation, input)\n", "\t\tlane.Succeed()\n\t\tcycle := lane.Run(gate.activation, input)\n")],
}


def run(command: list[str], cwd: Path, env: dict[str, str]) -> subprocess.CompletedProcess:
    return subprocess.run(command, cwd=cwd, env=env, capture_output=True, text=True, timeout=1800, check=False)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> int:
    work = Path(sys.argv[1]).resolve()
    sandbox = work / f"sandbox.{os.getpid()}"
    targets = [REPO / R, REPO / E, REPO / D, REPO / S]
    before = [digest(path) for path in targets]
    if sandbox.exists():
        shutil.rmtree(sandbox)
    sandbox.mkdir(parents=True)
    for name in ("go.mod", "go.sum"):
        shutil.copy2(REPO / name, sandbox / name)
    for name in ("internal", "cmd", "tools"):
        shutil.copytree(REPO / name, sandbox / name, symlinks=True)
    env = dict(os.environ, GOFLAGS="-trimpath", GOCACHE=str(work / "gocache-mutate"))
    test = ["systemd-run", "--user", "--scope", "-q", "-p", "MemoryMax=16G", "-p", "MemorySwapMax=0",
            "go", "test", "-count=1", "-tags", "tossos_testseams", "-skip", TRIMPATH_BLIND, *PACKAGES]
    control = run(test, sandbox, env)
    print(f"CONTROL (무변이) exit={control.returncode}")
    if control.returncode != 0:
        print(control.stdout[-3000:], control.stderr[-3000:])
        return 2
    verdicts = {}
    for name, edits in MUTANTS.items():
        originals = {}
        for relative, old, new in edits:
            path = sandbox / relative
            text = path.read_text()
            originals.setdefault(relative, text)
            if text.count(old) != 1:
                verdicts[name] = f"NOT REACHED ({relative}: 옛 문자열 {text.count(old)} 회)"
                break
            path.write_text(text.replace(old, new))
        else:
            vet = run(["go", "vet", "-tags", "tossos_testseams", *PACKAGES], sandbox, env)
            if vet.returncode != 0:
                verdicts[name] = "BUILD " + vet.stderr.strip().splitlines()[-1][:200]
            else:
                result = run(test, sandbox, env)
                failed = [line.strip() for line in result.stdout.splitlines() if line.startswith("--- FAIL")]
                verdicts[name] = ("CAUGHT " + "; ".join(failed[:4])) if result.returncode else "SURVIVED"
        for relative, text in originals.items():
            (sandbox / relative).write_text(text)
        print(f"{name}\t{verdicts[name]}", flush=True)
    after = [digest(path) for path in targets]
    print("저장소 대상 파일 불변:", before == after)
    shutil.rmtree(sandbox)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
