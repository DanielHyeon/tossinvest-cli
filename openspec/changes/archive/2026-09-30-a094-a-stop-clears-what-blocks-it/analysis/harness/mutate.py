#!/usr/bin/env python3
"""a094 변이 원장 하네스 — 구현 조각을 하나씩 되돌리거나 뒤집고, 그 조각을 지키는 시험이 실패하는지(CAUGHT) 잰다.

규율(저장소 교훈):
  - 공유 워킹트리를 건드리지 않는다: 지정 커밋의 **분리된 git worktree** 에서만 변이한다(시작 sha 단언).
  - 무변이 대조군이 GREEN 이 아니면 멈춘다(대조군이 빨가면 모든 변이가 CAUGHT 로 찍힌다).
  - 변이가 대상에 **닿는지**(치환 문자열이 정확히 한 번 있는지) 먼저 단언한다.
  - 한 번에 한 판. 변이마다 원복 후 파일 바이트가 원본과 같은지 확인한다.

사용: python3 mutate.py <commit> <worktree-dir> <ledger.md>
"""
from __future__ import annotations

import hashlib
import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]

E = "./internal/app/engine/"
J = "./internal/journal/"
X = "./internal/execgw/"
R = "./internal/reconcile/"
C = "./cmd/tossctl/"

# (id, 설명, 파일, 원문, 변이, 패키지, -run 패턴)
MUTANTS = [
    ("M1", "R1 확정 거절 갈래 제거(모호로)", "internal/execgw/classify.go",
     "\t\treturn journal.DispatchOutcome{\n\t\t\tClass:      journal.DispatchRejected,\n\t\t\tReasonCode: string(reason),",
     "\t\treturn journal.DispatchOutcome{\n\t\t\tClass:      journal.DispatchAmbiguous,\n\t\t\tReasonCode: string(reason),",
     X, "TestA094Refusal"),
    ("M2", "R1 두 자리 모순 판정 끄기", "internal/execgw/refusal_code.go",
     "if top.present && nested.present && !strings.EqualFold(top.value, nested.value) {",
     "if false && top.present && nested.present && !strings.EqualFold(top.value, nested.value) {", X, "TestA094Refusal"),
    ("M3", "R1 대소문자 구별(ToLower 제거)", "internal/execgw/refusal_code.go",
     "definitiveRefusalCodes[strings.ToLower(code.value)]", "definitiveRefusalCodes[code.value]", X, "TestA094Refusal"),
    ("M4", "R1 문자열 아닌 code 를 부재로(모순 강제 해제)", "internal/execgw/refusal_code.go",
     "\t\treturn codeField{value: unreadableCodeMark + string(raw), present: true}",
     "\t\treturn codeField{}", X, "TestA094Refusal"),
    ("M5", "해제 판정이 park 도 풂", "internal/journal/exit_proposal_release.go",
     "return v == ExitIntentUnaccepted || v == ExitIntentNoAttempts",
     "return v == ExitIntentUnaccepted || v == ExitIntentNoAttempts || v == ExitIntentParked", J, "TestA094"),
    ("M6", "기대 intent 대조 제거", "internal/journal/apply_hook.go",
     "\tif strings.TrimSpace(pendingIntent.String) != strings.TrimSpace(expectedIntentID) {",
     "\tif false && strings.TrimSpace(pendingIntent.String) != strings.TrimSpace(expectedIntentID) {", J, "TestA094"),
    ("M7", "분류기가 park 를 비수용으로", "internal/journal/exit_proposal_release.go",
     "\t\tcase StateFailedConfirmed, StateNotDispatched:\n\t\tcase StateUnresolvedInDoubt:\n\t\t\tout.Parked = append(out.Parked, rec)",
     "\t\tcase StateFailedConfirmed, StateNotDispatched, StateUnresolvedInDoubt:\n\t\tcase \"never\":\n\t\t\tout.Parked = append(out.Parked, rec)",
     J, "TestA094"),
    ("M8", "종결 증거 술어 뒤집기", "internal/journal/exit_proposal_release.go",
     "WHERE c.order_id = '' OR NOT `+confirmedOrderTerminalEvidence+`",
     "WHERE c.order_id = '' OR `+confirmedOrderTerminalEvidence+`", J, "TestA094"),
    ("M9", "청소가 같은 종목 미종결을 안 봄", "internal/app/engine/exitloop.go",
     "\tif len(unsettled) > 0 {", "\tif false && len(unsettled) > 0 {", E, "TestA094"),
    ("M10", "매도 취소 접수를 치움으로", "internal/app/engine/exitloop.go",
     "\t\tif !buy {\n\t\t\t// 매도의 취소 접수는 치움의 증거가 아님",
     "\t\tif false && !buy {\n\t\t\t// 매도의 취소 접수는 치움의 증거가 아님", E, "TestA094|TestABreachDisplaces"),
    ("M11", "확정 취소된 매도를 다시 취소", "internal/app/engine/exitloop.go",
     "\t\t\tif waiting {", "\t\t\tif false && waiting {", E, "TestA094"),
    ("M12", "청소 해제 실패를 무시(치움 완료)", "internal/app/engine/exitloop.go",
     "\t\tif !released {\n\t\t\treturn clearResult{countable: true",
     "\t\tif false && !released {\n\t\t\treturn clearResult{countable: true", E, "TestA094"),
    ("M13", "결과 못 쓴 제출도 해제(4.3c)", "internal/app/engine/exitloop.go",
     "\tcase out.AttemptID != \"\" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:",
     "\tcase false && out.AttemptID != \"\" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:",
     E, "TestA094"),
    ("M14", "판정 진입 알림 호출 제거", "internal/app/engine/exitloop.go",
     "\to.noteHeldProposal(ctx, m)\n", "\n", E, "TestA094"),
    ("M15", "종결 대기 알림 한계 0", "internal/app/engine/exit_held_proposal.go",
     "if waited < o.delayBound() {", "if waited < 0 {", E, "TestA094"),
    ("M16", "연속 실패 제외 무시", "internal/app/engine/exit_held_proposal.go",
     "\tif !res.countable {\n\t\treturn\n\t}", "\tif false && !res.countable {\n\t\treturn\n\t}", E, "TestA094"),
    ("M17", "빈 가격을 실패로", "internal/app/engine/exit_held_proposal.go",
     "\tif strings.TrimSpace(price) == \"\" {\n\t\treturn 0, nil\n\t}", "", E, "TestA094"),
    ("M18", "Critical 배선 제거", "internal/app/engine/exitwiring.go",
     "\t\topts.Critical = c.Notifier\n", "\n", E, "TestA094"),
    ("M19", "미종결 판정 대상 뒤집기(3.R5a 두 경로)", "internal/execgw/a094_shared_judgements.go",
     "\t\tif same {\n\t\t\tout = append(out, rec)", "\t\tif !same {\n\t\t\tout = append(out, rec)", E, "TestA094AnotherIntentInFlight"),
    ("M19b", "미종결 판정 대상 뒤집기 — 주문 경로 쪽", "internal/execgw/a094_shared_judgements.go",
     "\t\tif same {\n\t\t\tout = append(out, rec)", "\t\tif !same {\n\t\t\tout = append(out, rec)", X, "TestGatewayRefuses|TestSymbol|InFlight|TestTransportOutcomeTable|TestA094"),
    ("M20", "응답 종목 공백 통과(D−5.1 되돌림)", "internal/execgw/a094_shared_judgements.go",
     "\tif facts.Symbol == \"\" {", "\tif false && facts.Symbol == \"\" {", X, "TestA094TheReadBack"),
    ("M20b", "응답 종목 공백 통과 — 기동 쪽", "internal/execgw/a094_shared_judgements.go",
     "\tif facts.Symbol == \"\" {", "\tif false && facts.Symbol == \"\" {", R, "TestA094"),
    ("M21", "기동 확정이 읽기 판정을 무시", "internal/reconcile/acked_boot.go",
     "\tif err := execgw.ConfirmPlacedOrder(ctx, r.opts.Resolver.Order, rec.BrokerOrderID, intent.Symbol); err != nil {",
     "\tif err := execgw.ConfirmPlacedOrder(ctx, r.opts.Resolver.Order, rec.BrokerOrderID, intent.Symbol); false && err != nil {",
     R, "TestA094"),
    ("M22", "기동 ACKED 확정 끄기", "internal/reconcile/recovery.go",
     "if rec.State == journal.StateAcked && r.confirmAcked(ctx, rec) {", "if false && r.confirmAcked(ctx, rec) {", R, "TestA094"),
    ("M23", "따라잡기가 해제하지 않음", "internal/app/engine/boot_catchup.go",
     "\t\tif ok {\n\t\t\treleased++", "\t\tif false && ok {\n\t\t\treleased++", E, "TestA094"),
    ("M24", "복구 뒤 따라잡기 호출 제거", "cmd/tossctl/engine.go",
     "\t\tr.CatchUpExitProposals(ctx)\n", "\n", C, "TestA094"),
    ("M25", "해동 audit 부재 검사 제거", "internal/app/engine/attempt_thaw_command.go",
     "\tif s.audit == nil {\n\t\treturn attemptthaw.Result{}, attemptthaw.ErrAuditUnavailable\n\t}", "", E, "TestA094AThaw|TestA094AnOperatorThaw"),
    ("M26", "해동 뒤 발의 해제 호출 제거", "internal/app/engine/attempt_thaw_command.go",
     "\ts.releaseThawedProposal(ctx, repo, rec, &result)\n", "\n", E, "TestA094AThaw|TestA094AnOperatorThaw"),
    ("M27", "해동 stale 사전 검사 제거(OperatorResolve 의 from 검사가 뒤에 있음)", "internal/app/engine/attempt_thaw_command.go",
     "\tif rec.State != journal.StateUnresolvedInDoubt {", "\tif false && rec.State != journal.StateUnresolvedInDoubt {", E, "TestA094AThaw"),
    ("M28", "3.E5 배제의 대상 주문 결속 제거", "internal/app/engine/exit_held_proposal.go",
     "\t\tif !containsString(targets, target) {\n\t\t\treturn false\n\t\t}", "", E, "TestA094AnInFlightEngineCancel"),
    ("M29", "3.E5 모든 대상 주문에 취소 요구 제거", "internal/app/engine/exit_held_proposal.go",
     "\tfor _, target := range targets {\n\t\tif !cancelled[target] {\n\t\t\treturn false\n\t\t}\n\t}", "", E, "TestA094AnInFlightEngineCancel"),
    ("M30", "판정 진입의 비수용 발의 해제 제거", "internal/app/engine/exit_held_proposal.go",
     "\tif facts.Verdict == journal.ExitIntentUnaccepted {", "\tif false && facts.Verdict == journal.ExitIntentUnaccepted {", E, "TestA094"),
    ("M31", "연속 실패 경보 래치를 기록 전에", "internal/app/engine/exit_held_proposal.go",
     "\tstreak.alerted = o.recordEpisodeCritical(ctx, key, obs.Event{", "\tstreak.alerted = true\n\t_ = o.recordEpisodeCritical(ctx, key, obs.Event{", E, "TestA094"),
    ("M32", "다른 intent 매도의 종결 대기 알림 제거", "internal/app/engine/exit_held_proposal.go",
     "\t\to.noteAwaitingClose(ctx, m, order.AccountRef, order.Market, order.Symbol, order.OrderID)\n\t}\n}",
     "\t\t_ = order\n\t}\n}", E, "TestA094"),
    ("M33", "intent 없는 발의 알림 제거", "internal/app/engine/exit_held_proposal.go",
     "\t\to.recordIntentlessProposal(ctx, m)\n", "\n", E, "TestA094"),
    ("M34", "ACKED 알림에 브로커 본문 노출", "internal/reconcile/acked_boot.go",
     "\tif errors.As(err, &apiErr) {", "\tif false && errors.As(err, &apiErr) {", R, "TestA094"),
    ("M35", "따라잡기 목록 실패 critical 제거", "internal/app/engine/boot_catchup.go",
     "\t\tif critical != nil {\n\t\t\tboot := randomExitID()", "\t\tif false {\n\t\t\tboot := randomExitID()", E, "TestA094"),
    ("M36", "해동 해제 실패 critical 제거", "internal/app/engine/attempt_thaw_command.go",
     "\t\t\ts.recordThawReleaseFailure(ctx, rec, a.PositionID, err)\n", "\n", E, "TestA094AThaw"),
]


def run(cmd: list[str], cwd: Path, timeout: int = 3600) -> tuple[int, str]:
    env = dict(os.environ)
    p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, timeout=timeout, env=env)
    return p.returncode, p.stdout + p.stderr


def main() -> int:
    commit, wt, ledger = sys.argv[1], Path(sys.argv[2]), Path(sys.argv[3])
    only = set(sys.argv[4:])
    if not wt.exists():
        rc, out = run(["git", "worktree", "add", "--detach", str(wt), commit], ROOT)
        if rc:
            print(out)
            return 2
    rc, head = run(["git", "rev-parse", "HEAD"], wt)
    assert head.strip().startswith(commit[:8]) or head.strip() == commit, f"worktree HEAD {head} != {commit}"
    lines = [f"# a094 변이 원장 — 커밋 `{head.strip()[:12]}` 의 분리 worktree", "",
             "| id | 변이 | 패키지 · -run | 결과 |", "|---|---|---|---|"]
    # 무변이 대조군 — 패키지별로 GREEN 이어야 함.
    controls = sorted({(m[5], m[6]) for m in MUTANTS if not only or m[0] in only})
    for pkg, pattern in controls:
        rc, out = run(["go", "test", pkg, "-run", pattern, "-count=1"], wt)
        if rc:
            print(f"CONTROL RED {pkg} {pattern}\n{out[-3000:]}")
            return 3
    print("controls green:", len(controls))
    for mid, desc, rel, old, new, pkg, pattern in MUTANTS:
        if only and mid not in only:
            continue
        path = wt / rel
        original = path.read_bytes()
        text = original.decode()
        count = text.count(old)
        if count != 1:
            lines.append(f"| {mid} | {desc} | `{pkg}` `{pattern}` | **안 닿음**(원문 {count}회) |")
            continue
        path.write_text(text.replace(old, new))
        rc, out = run(["go", "test", pkg, "-run", pattern, "-count=1"], wt)
        path.write_bytes(original)
        assert hashlib.sha256(path.read_bytes()).digest() == hashlib.sha256(original).digest()
        if rc == 0:
            verdict = "**SURVIVED**"
        elif "build failed" in out or "[build failed]" in out or "setup failed" in out:
            verdict = "BUILD FAIL(무효 변이)"
        else:
            failed = [l.split()[2] for l in out.splitlines() if l.startswith("--- FAIL")][:3]
            verdict = "CAUGHT — " + ", ".join(f"`{f}`" for f in failed)
        lines.append(f"| {mid} | {desc} | `{pkg}` `{pattern}` | {verdict} |")
        print(mid, verdict, flush=True)
    ledger.write_text("\n".join(lines) + "\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
