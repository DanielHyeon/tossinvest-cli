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
TESTS = "^(TestA146.*|TestA137.*|TestAnExistingGuardianCapRefusalIsNotAScopeRefusal|TestTheEntrySkipTypeIsMadeOnlyWhereTheCensusSaysItIs)$"


def run(args):
    return subprocess.run(args, cwd=WT, capture_output=True, text=True, env=ENV)


G = "internal/app/engine/"
AD, DL, SV, SK = G + "strategy_first_leg_admission.go", G + "strategy_market_handoff_delivery.go", G + "strategy_entry_supervisor.go", G + "a146_entry_skip.go"
MUTANTS = [
    # 리뷰 2 R10: 전달 루프가 q_final 거절이면 무엇이든 건너뜀.
    ("R10 q_final 거절이면 건너뜀", DL, "if _, isSkip := strategyEntrySkipOf(err); isSkip {",
     "if _, isSkip := strategyEntrySkipOf(err); isSkip || errors.As(err, new(*execgw.QFinalRefusal)) {"),
    # 리뷰 2 R12: 원인과 무관하게 EXISTING_GUARDIAN_CAP 이면 건너뜀(전달 자리).
    ("R12 원인 무관 상한 코드면 건너뜀", DL, "if _, isSkip := strategyEntrySkipOf(err); isSkip {",
     "if _, isSkip := strategyEntrySkipOf(err); isSkip || func() bool { q := (*execgw.QFinalRefusal)(nil); return errors.As(err, &q) && q.Code == \"EXISTING_GUARDIAN_CAP\" }() {"),
    # 리뷰 1 M35/M36: 문구로 가름(타입 + 문구 / 문구만).
    ("M35 타입 또는 문구", SK, "\tif skip, ok := err.(*strategyEntrySkip); ok {\n\t\treturn skip, true\n\t}",
     "\tif skip, ok := err.(*strategyEntrySkip); ok {\n\t\treturn skip, true\n\t}\n\tif err != nil && strings.Contains(err.Error(), string(StrategyFirstLegGuardianCapZero)) {\n\t\treturn &strategyEntrySkip{}, true\n\t}"),
    ("M36 문구만", DL, "if _, isSkip := strategyEntrySkipOf(err); isSkip {",
     "if isSkip := err != nil && strings.Contains(err.Error(), string(StrategyFirstLegGuardianCapZero)); isSkip {"),
    # 리뷰 1 M31: 다른 시장 worker 에 기록.
    ("M31 다른 시장에 귀속", SV, "func(skip *strategyEntrySkip) { s.recordGuardianCapZeroSkip(worker, skip) }",
     "func(skip *strategyEntrySkip) { s.recordGuardianCapZeroSkip(s.workers[StrategyMarketKR], skip) }"),
    # 승격 P3-1 탐침: 판정이 다중 오류를 거르지 않음 / 트리 전체를 뒤짐 / 감싸개 문구 대조 없음.
    ("P1 다중 오류 거름 제거", SK, "if _, multi := err.(interface{ Unwrap() []error }); multi || err == nil {", "if err == nil {"),
    ("P2 트리 전체 errors.As", SK, "\tif skip, ok := err.(*strategyEntrySkip); ok {\n\t\treturn skip, true\n\t}",
     "\tif skip := (*strategyEntrySkip)(nil); errors.As(err, &skip) {\n\t\treturn skip, true\n\t}"),
    ("P3 감싸개 문구 대조 제거", SK, "if !ok || err.Error() != fmt.Sprintf(\"engine: first-leg admission %s: %s\", StrategyFirstLegGuardianCapZero, skip.Error()) {",
     "if !ok {"),
    ("P4 FX 권위를 알림에서 뺌", SK, "StrategyFirstLegGuardianCapZero, s.entry, s.stop, s.fx)", "StrategyFirstLegGuardianCapZero, s.entry, s.stop, \"\")"),
]


def test():
    return run(["go", "test", "-count=1", "-tags", "tossos_testseams", "-run", TESTS, PKG])


head = run(["git", "rev-parse", "HEAD"]).stdout.strip()
assert head == SHA, f"worktree HEAD {head} != expected {SHA}"
assert run(["git", "status", "--porcelain"]).stdout.strip() == "", "worktree not clean"
control = test()
assert control.returncode == 0 and "no tests to run" not in control.stdout, f"control not GREEN\n{control.stdout[-1500:]}"
lines = [f"# a146 변이 원장 round3 — 구현 리뷰 생존 자리\n\n- 사본: 저장소 밖 연결 워크트리 `{WT}` @ `{SHA[:8]}` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 GREEN)\n"
         f"- 시험 묶음(변이마다 같음): `go test -count=1 -tags tossos_testseams -run '{TESTS}' {PKG}`\n",
         "| 변이 | 파일 | 결과 | 처음 실패한 시험 |", "|---|---|---|---|"]
for label, path, old, new in MUTANTS:
    full = os.path.join(WT, path)
    src = open(full).read()
    if src.count(old) != 1:
        lines.append(f"| {label} | `{path}` | NOT-REACHED (문자열 {src.count(old)}회) | — |")
        print(label, "NOT-REACHED", flush=True)
        continue
    mutated = src.replace(old, new)
    if "strings.Contains" in new and '"strings"' not in mutated:
        mutated = mutated.replace('import (\n', 'import (\n\t"strings"\n', 1)
    open(full, "w").write(mutated)
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
