#!/usr/bin/env python3
"""a090 변이 하네스 — 편집한 생산 코드에 변이를 하나씩 넣고 a090 시험이 빨강이 되는지 잼.

사용: a090_mutate.py <대상 워크트리(사본)> <출력 로그>
- 대상은 반드시 사본(연결 워크트리)이어야 함 — 공유 트리에서 돌리지 않음.
- 무변이 대조군이 GREEN 이 아니면 멈춤(계측기가 눈먼 경우를 가름).
- 변이마다 원본 바이트를 되쓰고 sha256 로 복원을 확인함.
- 한 번에 한 판(병렬 금지).
"""
from __future__ import annotations

import hashlib
import pathlib
import subprocess
import sys

ENGINE = "internal/app/engine/exitloop.go"
UNOBS = "internal/app/engine/exit_unobserved.go"
RUNTIME = "cmd/tossctl/engine.go"

ENGINE_TEST = ["go", "test", "./internal/app/engine/", "-run", "TestA090", "-count=1"]
CMD_TEST = ["go", "test", "./cmd/tossctl/", "-run", "TestA090", "-count=1"]

# (id, 축, 파일, 원문, 변이, 시험)
MUTATIONS = [
    ("M01", "기록 삭제", ENGINE,
     "\t\t\to.noteUnobservedCause(&cycle, state.position.ID, unobservedNoQuote)\n", "", ENGINE_TEST),
    ("M02", "기록 삭제", ENGINE,
     "\t\t\to.noteUnobservedCause(&cycle, state.position.ID, unobservedQuoteExpired)\n", "", ENGINE_TEST),
    ("M03", "수리 되돌림(판정 표시)", ENGINE, "\t\to.noteJudged(&cycle, state)\n", "", ENGINE_TEST),
    ("M04", "수리 되돌림(순회 뒤 처리)", ENGINE,
     "\t// a090: 임계 판정 · 알림 · 강화는 모든 포지션 판정이 끝난 뒤 한 번 — 뒤 포지션의 손절 판정 앞에 서지 않음.\n"
     "\to.settleUnobserved(ctx, &cycle)\n", "", ENGINE_TEST),
    ("M05", "수리 되돌림(B3 처리)", ENGINE,
     "\t\to.settleUnobserved(ctx, &cycle)\n\t\t// Nothing is held", "\t\t// Nothing is held", ENGINE_TEST),
    ("M06", "수리 되돌림(표시)", ENGINE, "\t\to.markHeld(cycle, p)\n", "", ENGINE_TEST),
    ("M07", "수리 되돌림(해제)", ENGINE, "\t\t\t\to.unmarkHeld(cycle, p.ID)\n", "", ENGINE_TEST),
    ("M08", "무조건 실행(B4 에서도 처리)", ENGINE,
     "\t\tcycle.Err = err\n\t\to.checkOutage(ctx, &cycle)\n",
     "\t\tcycle.Err = err\n\t\to.settleUnobserved(ctx, &cycle)\n\t\to.checkOutage(ctx, &cycle)\n", ENGINE_TEST),
    ("M09", "무조건 실행(B1 에서도 처리)", ENGINE,
     "\t\tcycle.Deferred = true\n", "\t\tcycle.Deferred = true\n\t\to.settleUnobserved(ctx, &cycle)\n", ENGINE_TEST),
    ("M10", "순서(순회 앞 처리)", ENGINE,
     "\tfor _, state := range states {\n\t\tquote, ok := quotes",
     "\to.settleUnobserved(ctx, &cycle)\n\tfor _, state := range states {\n\t\tquote, ok := quotes", ENGINE_TEST),
    ("M11", "경계(< → <=)", UNOBS, "\t\tif elapsed < o.outageAfter() {", "\t\tif elapsed <= o.outageAfter() {", ENGINE_TEST),
    ("M12", "기점 갱신 삭제", UNOBS, "\t\t\trec.base = judged\n", "", ENGINE_TEST),
    ("M13", "에피소드 key 고정", UNOBS, "rec.streak = &unobservedStreak{id: o.opts.NewID()}",
     "rec.streak = &unobservedStreak{id: \"streak\"}", ENGINE_TEST),
    ("M14", "재강화 래치 삭제", UNOBS, "\ts.tightened, s.modeFail = true, false\n", "\ts.modeFail = false\n", ENGINE_TEST),
    ("M15", "계좌 필드 남김", UNOBS, "\tdelete(e.Fields, obs.FieldAccount)\n", "", ENGINE_TEST),
    ("M16", "공지 key 에 계좌", UNOBS, "\te.Key = unobservedModeNoticeHead + rec.Mode + \":\" + rec.ID\n",
     "\te.Key = obs.OperatingModeEventKey(rec)\n", ENGINE_TEST),
    ("M17", "동기 알림 경로(Notify)", UNOBS,
     "\t\tif err := recorder.RecordCritical(ctx, o.unobservedEvent(id, rec, elapsed), 0); err != nil {",
     "\t\tif err := o.opts.Alerts.Notify(ctx, o.unobservedEvent(id, rec, elapsed)); err != nil {", ENGINE_TEST),
    ("M18", "미적재 공지 대기열 삭제", UNOBS, "\t\to.pendingModeNotices = append(o.pendingModeNotices, e)\n", "",
     ENGINE_TEST),
    ("M19", "벽시계 경과", UNOBS, "\t\telapsed := clock.LeaseElapsed(o.clk, rec.base.anchor)\n",
     "\t\telapsed := o.clk.Now().Sub(rec.base.wall)\n", ENGINE_TEST),
    ("M20", "알림 없는 조임", UNOBS,
     "\t\t\ts.alertFail = true\n\t\t\treturn\n", "\t\t\ts.alertFail = true\n", ENGINE_TEST),
    ("M21", "원문 오류 로그", UNOBS, "\targs := []any{\"failure\", failure, obs.FieldDetail, unobservedFailureDetail}\n",
     "\targs := []any{\"failure\", failure, obs.FieldDetail, fmt.Sprint(o.opts.AccountRef, \" down disk full\")}\n",
     ENGINE_TEST),
    ("M22", "주기 신원 무시(B4 표시 누수)", UNOBS,
     "\tif tally != nil && tally.owner != cycle {\n\t\ttally = nil // 이 주기에는 표시가 없었음(옛 주기의 집계)\n\t}\n",
     "", ENGINE_TEST),
    ("M23", "적재 래치 삭제(같은 연속 재적재)", UNOBS, "\t\ts.recorded, s.alertFail = true, false\n",
     "\t\ts.alertFail = false\n", ENGINE_TEST),
    ("M24", "정리 삭제", UNOBS, "\t\tif _, ok := tally.held[id]; !ok {\n\t\t\tdelete(o.unobserved, id)\n\t\t}\n",
     "\t\tif _, ok := tally.held[id]; !ok {\n\t\t}\n", ENGINE_TEST),
    ("M25", "공지 재시도 삭제", UNOBS, "\to.retryModeNotices(ctx)\n\n\ttally := o.cycleUnobserved",
     "\n\ttally := o.cycleUnobserved", ENGINE_TEST),
    # 리뷰(분리 컨텍스트, 2026-09-30) 가 찾은 생존 변이 넷 — 보강 뒤 재측정
    ("M28", "무조건 실행(B2 에서도 처리)", ENGINE,
     "\tstates, err := o.workingSet(ctx, &cycle)\n\tif err != nil {\n\t\tcycle.Err = err\n\t\treturn cycle\n",
     "\tstates, err := o.workingSet(ctx, &cycle)\n\tif err != nil {\n\t\tcycle.Err = err\n\t\to.settleUnobserved(ctx, &cycle)\n\t\treturn cycle\n",
     ENGINE_TEST),
    ("M29", "단조 앵커 → 벽시계", UNOBS, "\tanchor := clock.LeaseAnchor(o.clk)\n", "\tanchor := o.clk.Now()\n", ENGINE_TEST),
    ("M30", "공지 실패를 커밋 실패로", UNOBS, "!errors.Is(err, journal.ErrModeAnnouncementFailed)",
     "!errors.Is(err, context.Canceled)", ENGINE_TEST),
    ("M31", "적재된 공지를 대기열에 남김", UNOBS,
     "\t\tif err := recorder.RecordCritical(ctx, e, 0); err != nil {\n\t\t\tkept = append(kept, e)\n\t\t}\n",
     "\t\t_ = recorder.RecordCritical(ctx, e, 0)\n\t\tkept = append(kept, e)\n", ENGINE_TEST),
    ("M26", "생산 배선 삭제", RUNTIME, "\t\tUnobservedLog: logger,\n", "", CMD_TEST),
    ("M27", "생산 배선에 전체 Log", RUNTIME, "\t\tUnobservedLog: logger,\n", "\t\tUnobservedLog: logger,\n\t\tLog: logger,\n",
     CMD_TEST),
]


def sha(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(root: pathlib.Path, cmd: list[str]) -> tuple[int, str]:
    p = subprocess.run(cmd, cwd=root, capture_output=True, text=True, timeout=900)
    tail = [l for l in (p.stdout + p.stderr).splitlines() if l.startswith(("--- FAIL", "ok", "FAIL"))]
    return p.returncode, " | ".join(tail[:4])


def main() -> int:
    root = pathlib.Path(sys.argv[1]).resolve()
    out = pathlib.Path(sys.argv[2])
    top = subprocess.run(["git", "-C", str(root), "rev-parse", "--show-toplevel"], capture_output=True,
                         text=True, check=True).stdout.strip()
    if pathlib.Path(top).resolve() != root or root.name == "TossOS":
        print("대상은 연결 워크트리 사본이어야 함:", root)
        return 2
    lines = [f"# a090 mutation — root {root} · HEAD {subprocess.run(['git', '-C', str(root), 'rev-parse', '--short', 'HEAD'], capture_output=True, text=True).stdout.strip()}"]
    originals = {f: (root / f).read_bytes() for f in {ENGINE, UNOBS, RUNTIME}}
    shas = {f: sha(root / f) for f in originals}
    for f, s in sorted(shas.items()):
        lines.append(f"# {f} sha256 {s}")

    for name, cmd in (("C0-engine", ENGINE_TEST), ("C0-cmd", CMD_TEST)):
        rc, tail = run(root, cmd)
        lines.append(f"{name} 무변이 대조군 rc={rc} {tail}")
        if rc != 0:
            out.write_text("\n".join(lines) + "\n무변이 대조군이 GREEN 이 아님 — 중단\n")
            return 1

    caught = 0
    for mid, axis, f, old, new, cmd in MUTATIONS:
        path = root / f
        src = originals[f].decode()
        if src.count(old) != 1:
            lines.append(f"{mid} {axis}: 변이 앵커가 유일하지 않음(count {src.count(old)}) — 측정 안 함")
            continue
        path.write_text(src.replace(old, new, 1))
        try:
            rc, tail = run(root, cmd)
        finally:
            path.write_bytes(originals[f])
        assert sha(path) == shas[f], f"{f} 복원 실패"
        verdict = "CAUGHT" if rc != 0 else "SURVIVED"
        caught += rc != 0
        lines.append(f"{mid} {axis} [{f}]: {verdict} rc={rc} {tail}")
    lines.append(f"# 합계 CAUGHT {caught}/{len(MUTATIONS)}")
    out.write_text("\n".join(lines) + "\n")
    print("\n".join(lines))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
