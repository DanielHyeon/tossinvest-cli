#!/usr/bin/env python3
"""a146 변이 원장 하네스 — 저장소 밖에 고정한 연결 워크트리(사본) 하나에서 변이를 하나씩 넣고 a146 시험 묶음을 돌린 뒤 되돌림.

사용: python3 a146_mutate.py <worktree> <expected-sha> <out.md>
- 시작 sha 단언 · 사본 깨끗함 확인 · 무변이 대조군(같은 시험 묶음) GREEN 이어야 시작(GREEN 이 아니면 멈춤).
- 변이마다: 정확히 한 번 나타나는 문자열을 바꿈(아니면 NOT-REACHED) → 시험 → git checkout 으로 그 파일만 되돌림 → 사본이 다시 깨끗한지 확인.
- 사본 경로는 호출자가 pid 접미로 만든다(두 판이 한 사본을 쓰지 않게 — 한 번에 한 판).
"""
import os, subprocess, sys

WT, SHA, OUT = sys.argv[1], sys.argv[2], sys.argv[3]
ENV = dict(os.environ, GOFLAGS="-trimpath")
PKG = "./internal/app/engine"
TESTS = "^(TestA146.*|TestAnExistingGuardianCapRefusalIsNotAScopeRefusal|TestTheEntrySkipTypeIsMadeOnlyWhereTheCensusSaysItIs)$"


def run(args):
    return subprocess.run(args, cwd=WT, capture_output=True, text=True, env=ENV)


G = "internal/app/engine/"
AD, DL, SV, SK = G + "strategy_first_leg_admission.go", G + "strategy_market_handoff_delivery.go", G + "strategy_entry_supervisor.go", G + "a146_entry_skip.go"
MUTANTS = [
    # M1 은 동등 변이로 선언(원장 round2 머리말): 원인 사슬에 sentinel 을 싣는 q_final 거절은 EXISTING_GUARDIAN_CAP 하나뿐이라 코드 조건은 방어용.
    ("M1 코드 조건 제거(동등 — 선언)", AD, "errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalExistingGuardianCap &&", "errors.As(err, &qFinal) && true &&"),
    ("M2 sentinel 조건 제거", AD, "\t\t\terrors.Is(qFinal, risk.ErrStrategyQuantityZero) {", "\t\t\ttrue {"),
    ("M3 다른 sentinel", AD, "errors.Is(qFinal, risk.ErrStrategyQuantityZero)", "errors.Is(qFinal, risk.ErrAccountBaseFXUnavailable)"),
    ("M4 sentinel 을 오류 전체에서 찾음", AD, "errors.Is(qFinal, risk.ErrStrategyQuantityZero)", "errors.Is(err, risk.ErrStrategyQuantityZero)"),
    ("M5 원인이 스킵 타입을 잃음", AD, "\t\t\trefusal.cause = skip\n", "\t\t\trefusal.cause = err\n"),
    ("M6 코드를 AUTHORITY_MISMATCH 로", AD, "strategyFirstLegRefusal(StrategyFirstLegGuardianCapZero, accepted.market, err.Error())",
     "strategyFirstLegRefusal(StrategyFirstLegAuthorityMismatch, accepted.market, err.Error())"),
    ("M7 기록 호출 제거", AD, "\t\t\trecordStrategyEntrySkip(ctx, skip)\n", ""),
    ("M8 스킵을 주기 오류에 모음", DL, "\t\tif errors.As(err, &skip) {\n\t\t\tcontinue", "\t\tif errors.As(err, &skip) {\n\t\t\tskipped = append(skipped, err)\n\t\t\tcontinue"),
    ("M9 스킵 판정 제거", DL, "\t\tif errors.As(err, &skip) {", "\t\tif false && errors.As(err, &skip) {"),
    ("M10 스킵에서 주기를 끝냄(뒤 범위 굶음)", DL, "\t\tif errors.As(err, &skip) {\n\t\t\tcontinue", "\t\tif errors.As(err, &skip) {\n\t\t\treturn errors.Join(skipped...)"),
    ("M11 감독자 기록자 설치 무력화", SV, "func(skip *strategyEntrySkip) { s.recordGuardianCapZeroSkip(worker, skip) }", "func(skip *strategyEntrySkip) {}"),
    ("M12 계수 증가 제거", SK, "\t\tworker.guardianCapZeroSkips++\n", ""),
    ("M13 첫 원인 덮어쓰기", SK, "\tif worker.firstGuardianCapZeroSkip == \"\" {", "\tif true {"),
    ("M14 종목 dedup 제거", SK, "\tannounce := !noticed\n", "\tannounce := noticed || !noticed\n"),
    ("M15 일 키 바뀜에 기억 비우기 제거", SK, "if worker.entrySkipNoticed == nil || worker.entrySkipNoticeDay != day {", "if worker.entrySkipNoticed == nil {"),
    ("M16 빈 일 키에서 침묵(a137 모양)", SK, "\tannounce := !noticed\n", "\tannounce := day != \"\" && !noticed\n"),
    ("M17 생산 알림 배선 제거", SV, " EntrySkipNotice: c.strategyEntrySkipNotice(),", ""),
    ("M18 감독자가 알림 옵션을 버림", SV, " entrySkipNotice: opts.EntrySkipNotice,", ""),
    ("M19 스냅숏 계수 칸 제거", SV, "GuardianCapZeroSkips: worker.guardianCapZeroSkips, ", ""),
    ("M20 알림 줄에서 가격 제거", SK, "StrategyFirstLegGuardianCapZero, s.entry, s.stop)", "StrategyFirstLegGuardianCapZero, \"\", \"\")"),
    ("M21 알림 이벤트 이름 바꿈", SK, "log.Warn(obs.EventStrategyEntrySkipped,", "log.Warn(obs.EventDailyLossUnmeasurable,"),
]


def test():
    return run(["go", "test", "-count=1", "-tags", "tossos_testseams", "-run", TESTS, PKG])


head = run(["git", "rev-parse", "HEAD"]).stdout.strip()
assert head == SHA, f"worktree HEAD {head} != expected {SHA}"
assert run(["git", "status", "--porcelain"]).stdout.strip() == "", "worktree not clean"
control = test()
assert control.returncode == 0 and "no tests to run" not in control.stdout, f"control not GREEN\n{control.stdout[-1500:]}"
lines = [f"# a146 변이 원장\n\n- 사본: 저장소 밖 연결 워크트리 `{WT}` @ `{SHA[:8]}` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 GREEN)\n"
         f"- 시험 묶음(변이마다 같음): `go test -count=1 -tags tossos_testseams -run '{TESTS}' {PKG}`\n",
         "| 변이 | 파일 | 결과 | 처음 실패한 시험 |", "|---|---|---|---|"]
for label, path, old, new in MUTANTS:
    full = os.path.join(WT, path)
    src = open(full).read()
    if src.count(old) != 1:
        lines.append(f"| {label} | `{path}` | NOT-REACHED (문자열 {src.count(old)}회) | — |")
        print(label, "NOT-REACHED", flush=True)
        continue
    open(full, "w").write(src.replace(old, new))
    r = test()
    out = r.stdout + r.stderr
    if r.returncode == 0:
        verdict, first = "SURVIVED", "—"
    elif "build failed" in out or "[build failed]" in out or "setup failed" in out:
        verdict, first = "CAUGHT(build)", "—"
    else:
        verdict = "CAUGHT"
        fails = [l.split()[2] for l in out.splitlines() if l.startswith("--- FAIL:")]
        first = "`" + fails[0] + "`" if fails else "(패키지 실패)"
    run(["git", "checkout", "--", path])
    assert run(["git", "status", "--porcelain"]).stdout.strip() == "", f"revert failed after {label}"
    lines.append(f"| {label} | `{path}` | {verdict} | {first} |")
    print(label, verdict, first, flush=True)
open(OUT, "w").write("\n".join(lines) + "\n")
