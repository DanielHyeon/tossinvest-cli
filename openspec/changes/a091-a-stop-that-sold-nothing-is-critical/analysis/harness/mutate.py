#!/usr/bin/env python3
"""a091 변이 원장 하네스 — 구현 조각을 하나씩 되돌리거나 뒤집고, 그 조각을 지키는 시험이 실패하는지(CAUGHT) 잰다.

규율(a094 하네스와 같음): 지정 커밋의 분리된 git worktree 에서만 변이(시작 sha 단언) · 무변이 대조군 GREEN 먼저 · 치환 원문이 정확히
한 번 있는지(닿는지) 단언 · 한 번에 한 판 · 원복 뒤 바이트 대조.

사용: python3 mutate.py <commit> <worktree-dir> <ledger.md> [M…]
"""
from __future__ import annotations

import hashlib
import os
import subprocess
import sys
from pathlib import Path

E = "./internal/app/engine/"
O = "./internal/obs/"
ENG = "internal/app/engine/exit_stop_sold_nothing.go"
LOOP = "internal/app/engine/exitloop.go"
T = "TestA091"

MUTANTS = [
    ("M1", "새 종류 등록 제거", "internal/obs/event.go", "\tEventExitStopSoldNothing: true,\n", "\n", O, T),
    ("M1e", "새 종류 등록 제거(엔진 쪽 관측)", "internal/obs/event.go", "\tEventExitStopSoldNothing: true,\n", "\n", E, "TestA091AProtectiveZero"),
    ("M2", "submit 이 보호 여부를 거짓으로", LOOP, "o.applyFloor(ctx, m, quantity, isProtective(proposal))",
     "o.applyFloor(ctx, m, quantity, false)", E, T),
    ("M3", "submit 이 보호 여부를 늘 참으로", LOOP, "o.applyFloor(ctx, m, quantity, isProtective(proposal))",
     "o.applyFloor(ctx, m, quantity, true)", E, T),
    ("M4", "알림 켜짐 게이트 제거", ENG, "return z.protective && o.opts.NotificationsEnabled && (",
     "return z.protective && (", E, T),
    ("M5", "생산 배선의 설정 덮기 제거", "internal/app/engine/exitwiring.go",
     "\topts.NotificationsEnabled = c.Config.Engine.Notifications.Enabled\n", "\n", E, "TestA091TheProductionAssembly"),
    ("M6", "보유 0 배제 제거", ENG, "\tcase floor.Bound == riskcalc.FloorBoundHoldings:\n\t\treturn zeroNoHolding\n", "", E, T),
    ("M7", "합쳐진 오류에서 잎 하나만 취소여도 억제(errors.Is 의미)", ENG,
     "\t\t\tif !cancellationOnly(e) {\n\t\t\t\treturn false\n\t\t\t}",
     "\t\t\tif cancellationOnly(e) {\n\t\t\t\treturn true\n\t\t\t}", E, T),
    ("M7b", "취소 판정에서 ctx 확인 제거", ENG,
     "case err != nil && ctx.Err() != nil && cancellationOnly(err):", "case err != nil && cancellationOnly(err):", E, T),
    ("M8", "취소 판정 늘 거짓", ENG, "\treturn err == context.Canceled\n}", "\treturn false\n}", E, T),
    ("M9", "기록에서 WithoutCancel 제거", ENG, "o.opts.Alerts.Notify(context.WithoutCancel(ctx), e)",
     "o.opts.Alerts.Notify(ctx, e)", E, T),
    ("M10", "B2 critical 알림 빠뜨림", ENG, "\t\tif kind != obs.EventExitStopSoldNothing {\n\t\t\treturn\n\t\t}",
     "\t\treturn", E, T),
    ("M11", "B2 가 늘 알림(익절 · 꺼짐도)", ENG, "\t\tif kind != obs.EventExitStopSoldNothing {\n\t\t\treturn\n\t\t}",
     "", E, T),
    ("M12", "끝 경로 0주 분기 제거(옛 부분 캡 알림으로)", LOOP, "\tif isZeroQuantity(floor.Quantity) {", "\tif false {", E, T),
    ("M13", "escalate 실패 줄에 계좌 복원", "internal/obs/notifier.go",
     "\t\tn.Log.Error(EventOperatingMode, MaskAccount(err, n.AccountRef),\n\t\t\tFieldTriggerEvent, string(e.Type),",
     "\t\tn.Log.Error(EventOperatingMode, MaskAccount(err, n.AccountRef),\n\t\t\tFieldAccount, n.AccountRef,\n\t\t\tFieldTriggerEvent, string(e.Type),",
     O, T),
    ("M13b", "escalate 승격 줄에 계좌 복원", "internal/obs/notifier.go",
     "\t\tn.Log.Warn(EventOperatingMode,\n\t\t\tFieldToState, journal.ModeEntryBlocked,",
     "\t\tn.Log.Warn(EventOperatingMode,\n\t\t\tFieldAccount, n.AccountRef,\n\t\t\tFieldToState, journal.ModeEntryBlocked,", O, T),
    ("M13e", "escalate 실패 줄에 계좌 복원(엔진 카나리)", "internal/obs/notifier.go",
     "\t\tn.Log.Error(EventOperatingMode, MaskAccount(err, n.AccountRef),\n\t\t\tFieldTriggerEvent, string(e.Type),",
     "\t\tn.Log.Error(EventOperatingMode, MaskAccount(err, n.AccountRef),\n\t\t\tFieldAccount, n.AccountRef,\n\t\t\tFieldTriggerEvent, string(e.Type),",
     E, "TestA091NoLineOrRow"),
    ("M14", "a091 로그 줄에 계좌 필드 · 원문 오류", ENG, "o.opts.ZeroFloorLog.Error(kind, obs.MaskAccount(err, o.opts.AccountRef),",
     "o.opts.ZeroFloorLog.Error(kind, err, obs.FieldAccount, o.opts.AccountRef,", E, T),
    ("M15", "보고가 o.alert 경유(실패 로그 원문 계좌)", ENG,
     "\tif err := o.opts.Alerts.Notify(context.WithoutCancel(ctx), e); err != nil {\n\t\to.logZeroFloor(kind, m, z, err, \"the report that the stop sold nothing could not be made durable\")\n\t}",
     "\to.alert(context.WithoutCancel(ctx), e)", E, T),
    ("M16", "키에 원인 포함", ENG, "Key:   string(kind) + \"|\" + m.position.ID,",
     "Key:   string(kind) + \"|\" + m.position.ID + \"|\" + zeroCauseCode(z),", E, T),
    ("M17", "본문에서 시각 제거", ENG, "label, o.clk.Now().UTC().Format(time.RFC3339), zeroCauseText(z), z.proposed, what),",
     "label, time.Time{}.Format(\"\"), zeroCauseText(z), z.proposed, what),", E, T),
    ("M18", "본문에 원문 오류", ENG, "label, o.clk.Now().UTC().Format(time.RFC3339), zeroCauseText(z), z.proposed, what),",
     "label, o.clk.Now().UTC().Format(time.RFC3339), zeroCauseText(z)+fmt.Sprint(z.err), z.proposed, what),", E, T),
    ("M19", "B2 반환값 변경(오류로)", LOOP, "\t\treturn \"0\", true, nil", "\t\treturn \"\", false, err", E, T),
    ("M20", "끝 0주가 원안을 돌려줌(전량 제출)", LOOP, "\t\treturn floor.Quantity, true, nil\n\t}\n\to.alert(ctx, obs.Event{",
     "\t\treturn quantity, true, nil\n\t}\n\to.alert(ctx, obs.Event{", E, T),
    ("M21", "잎 없는 다중 오류를 취소뿐으로", ENG, "\t\treturn seen", "\t\treturn seen || len(leaves) > 0", E, T),
    # --- 구현 리뷰 i1 이 낸 생존 · 미변이 자리(보이스 A · B) ---
    ("M22", "보유 0 배제를 로컬 매도까지 넓힘", ENG, "\tcase floor.Bound == riskcalc.FloorBoundHoldings:",
     "\tcase floor.Bound == riskcalc.FloorBoundHoldings || floor.Bound == riskcalc.FloorBoundLocalSells:", E, T),
    ("M23", "보유 0 배제를 스냅숏 부재 · 낡음까지 넓힘", ENG, "\tcase floor.Bound == riskcalc.FloorBoundHoldings:",
     "\tcase floor.Bound == riskcalc.FloorBoundHoldings || floor.Bound == riskcalc.FloorBoundStaleSnapshot || floor.Bound == riskcalc.FloorBoundNoSnapshot:", E, T),
    ("M24", "부재 · 낡음 문구 맞바꿈", ENG, "\t\treturn \"계좌 스냅숏을 읽지 못했다\"", "\t\treturn \"계좌 스냅숏이 낡았다\"", E, T),
    ("M25", "계산 실패 원인 표지 바꿈", ENG, "\t\treturn \"floor_unknown\"", "\t\treturn \"floor_zero\"", E, T),
    ("M26", "payload 에서 position_id 제거", ENG, "\t\t\"position_id\":     m.position.ID,\n\t\t\"cause\"", "\t\t\"cause\"", E, T),
    ("M27", "payload 에서 floor_bound 제거", ENG, "\t\tfields[\"floor_bound\"] = z.floor.Bound", "\t\t_ = z.floor.Bound", E, T),
    ("M28", "critical B2 의 가린 오류 줄 생략", ENG,
     "\t\to.logZeroFloor(kind, m, z, z.err, \"the confirmed floor of \"+m.position.Symbol+\" could not be computed; nothing is submitted\")\n",
     "\t\tif kind != obs.EventExitStopSoldNothing { o.logZeroFloor(kind, m, z, z.err, \"the confirmed floor of \"+m.position.Symbol+\" could not be computed; nothing is submitted\") }\n", E, T),
    ("M29", "기록 실패 줄 생략", ENG,
     "\t\to.logZeroFloor(kind, m, z, err, \"the report that the stop sold nothing could not be made durable\")", "\t\t_ = err", E, T),
    ("M30", "익절 본문도 「손절」", ENG, "\tif z.protective {\n\t\twhat = \"손절\"", "\tif true {\n\t\twhat = \"손절\"", E, T),
    ("M31", "본문에서 종목 표지 제거", ENG, "Body: fmt.Sprintf(\"%s · %s 관측", "Body: fmt.Sprintf(\"%.0s%s 관측", E, T),
    ("M32", "로그 줄에서 포지션 표지 제거", ENG, "\"position_id\", m.position.ID, \"cause\"", "\"cause\"", E, T),
    ("M33", "Unwrap nil 을 취소뿐으로", ENG, "\t\tif inner := u.Unwrap(); inner != nil {\n\t\t\treturn cancellationOnly(inner)\n\t\t}\n\t\treturn false",
     "\t\tif inner := u.Unwrap(); inner != nil {\n\t\t\treturn cancellationOnly(inner)\n\t\t}\n\t\treturn true", E, T),
    ("M34", "재시도 대기 중 취소를 취소로 반환(execgw)", "internal/execgw/retry.go",
     "\t\t\tif !r.sleepWithin(ctx, clk, policy.BackoffFor(attempt, policy.BaseBackoff), deadline) {\n\t\t\t\treturn err\n\t\t\t}",
     "\t\t\tif !r.sleepWithin(ctx, clk, policy.BackoffFor(attempt, policy.BaseBackoff), deadline) {\n\t\t\t\tif ctx.Err() != nil {\n\t\t\t\t\treturn ctx.Err()\n\t\t\t\t}\n\t\t\t\treturn err\n\t\t\t}", E, T),
    ("M35", "옛 종류 B2 줄에 계좌(카나리 범위)", ENG, "o.opts.ZeroFloorLog.Error(kind, obs.MaskAccount(err, o.opts.AccountRef),",
     "o.opts.ZeroFloorLog.Error(kind, obs.MaskAccount(err, o.opts.AccountRef), obs.FieldAccount, o.opts.AccountRef,", E, "TestA091NoLineOrRow"),
    ("M36", "생산 배선의 0주 로그 싱크 제거", "internal/app/engine/exitwiring.go", "\topts.ZeroFloorLog = c.Log\n", "\n", E, "TestA091TheProductionAssembly"),

]


def run(cmd, cwd, timeout=3600):
    p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, timeout=timeout, env=dict(os.environ))
    return p.returncode, p.stdout + p.stderr


def main() -> int:
    commit, wt, ledger = sys.argv[1], Path(sys.argv[2]), Path(sys.argv[3])
    only = set(sys.argv[4:])
    rc, head = run(["git", "rev-parse", "HEAD"], wt)
    assert head.strip().startswith(commit[:8]), f"worktree HEAD {head} != {commit}"
    rc, dirty = run(["git", "status", "--porcelain", "--untracked-files=no"], wt)
    assert not dirty.strip(), f"worktree is dirty:\n{dirty}"
    lines = [f"# a091 변이 원장 — 커밋 `{head.strip()[:12]}` 의 분리 worktree", "",
             "| id | 변이 | 패키지 · -run | 결과 |", "|---|---|---|---|"]
    controls = sorted({(m[5], m[6]) for m in MUTANTS if not only or m[0] in only})
    for pkg, pattern in controls:
        rc, out = run(["go", "test", pkg, "-run", pattern, "-count=1", "-timeout", "30m"], wt)
        if rc:
            print(f"CONTROL RED {pkg} {pattern}\n{out[-3000:]}")
            return 3
    print("controls green:", len(controls), flush=True)
    lines.append(f"| C | 무변이 대조군 {len(controls)} 조합 | — | GREEN |")
    for mid, desc, rel, old, new, pkg, pattern in MUTANTS:
        if only and mid not in only:
            continue
        path = wt / rel
        original = path.read_bytes()
        text = original.decode()
        count = text.count(old)
        if count != 1:
            lines.append(f"| {mid} | {desc} | `{pkg}` `{pattern}` | **안 닿음**(원문 {count}회) |")
            print(mid, "NOT REACHED", count, flush=True)
            continue
        path.write_text(text.replace(old, new))
        rc, out = run(["go", "test", pkg, "-run", pattern, "-count=1", "-timeout", "30m"], wt)
        path.write_bytes(original)
        assert hashlib.sha256(path.read_bytes()).digest() == hashlib.sha256(original).digest()
        if rc == 0:
            verdict = "**SURVIVED**"
        elif "build failed" in out or "setup failed" in out:
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
