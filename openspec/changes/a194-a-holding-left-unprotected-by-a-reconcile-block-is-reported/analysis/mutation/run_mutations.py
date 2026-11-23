#!/usr/bin/env python3
"""a194 변이 원장 하네스(tasks 4.1).

worktree 가 아니라 사본에서 돈다: 저장소 HEAD 를 `git archive` 로 스크래치 아래 `a194-mut-<pid>` 에 풀고(시작 sha 단언),
무변이 대조군이 GREEN 인지 먼저 보고, 각 변이 자리가 무변이 커버(set 모드)에서 실행되는지(닿음)를 본 뒤, 한 번에 한 판씩
변이 → 선택 시험 → 원복(원문 문자열을 다시 세어 1 회인지 확인)한다. 결과는 표준 출력의 JSON 줄과 ledger.md 로 남긴다.

실행: 저장소 루트에서 `python3 openspec/changes/<id>/analysis/mutation/run_mutations.py <scratch-dir>`.
"""
import json
import os
import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parents[5]
HERE = pathlib.Path(__file__).resolve().parent
RUN = ("A194|Adoption|ReconcileLoop|Unmanaged|TestA095NotificationsOff|TestUSIncludedSymbolWaits|"
       "TestActiveForeignCause")
NEW = "internal/app/engine/reconcile_block_report.go"
ADOPT = "internal/app/engine/adoption.go"

# (id, 설명, 파일, 원문, 변이) — 원문은 파일에 정확히 1 회 나와야 함. 닿음은 원문이 있는 줄의 커버 블록으로 잼.
MUTANTS = [
    ("M1", "운영자 전용 판정을 항상 거짓", NEW,
     "earliestCoveringBlock(blocks, h.market, h.symbol, true); ok {",
     "earliestCoveringBlock(blocks, h.market, h.symbol, true); ok && false {"),
    ("M2a", "유예 비교 경계(< → <=)", NEW, "now.Sub(first) < blockedHoldingGrace", "now.Sub(first) <= blockedHoldingGrace"),
    ("M2b", "유예 비교 방향 뒤집기(< → >)", NEW, "now.Sub(first) < blockedHoldingGrace", "now.Sub(first) > blockedHoldingGrace"),
    ("M3", "시계를 지우지 않기", NEW, "\t\t\tdelete(d.blockedSince, id)\n", "\t\t\t_ = id\n"),
    ("M4", "D4 조건 삭제", NEW,
     "if d.opts.Adoption.Excludes(h.symbol) || (!d.opts.Adoption.Enabled && !d.opts.Adoption.Included(h.symbol)) {",
     "if false {"),
    ("M5", "차단 보고 key 에서 날짜 삭제", NEW, "symbol, b.Cause, day}, \"|\")", "symbol, b.Cause}, \"|\"); _ = day"),
    ("M6", "보유 보고 critical 앞에 메모리 래치", NEW,
     "if !d.opts.NotificationsEnabled && d.unmanaged[r.position.ID][r.fact] {",
     "if d.unmanaged[r.position.ID][r.fact] {"),
    ("M7", "차단 보고 래치를 Notify 앞에(결과 무시)", NEW,
     "|| d.recordOperatorOnlyBlock(ctx, b, symbol, key, now) {",
     "|| (d.recordOperatorOnlyBlock(ctx, b, symbol, key, now) || true) {"),
    ("M8", "KST 날짜를 UTC 날짜로", NEW, "now.In(kst).Format(", "now.UTC().Format("),
    ("M9", "Blocks().Release 대신 EntryAllowed 사유(Permanent)로 판정", NEW,
     "(operatorOnly && b.Release != reconcile.ReleaseOperatorOnly)", "(operatorOnly && !b.Permanent)"),
    ("M10", "주기 상한 삭제", NEW, "if attempted == blockedReportsPerCycle {", "if false {"),
    ("M11", "새 보고를 adopt 앞으로 옮기기", ADOPT,
     "\tresults := d.adopt(ctx, candidates, cycle)\n",
     "\td.reportBlocked(ctx, held)\n\tresults := d.adopt(ctx, candidates, cycle)\n"),
    ("M12", "순서를 포지션 id 만으로", NEW,
     "\t\tif !due[i].first.Equal(due[j].first) {\n\t\t\treturn due[i].first.Before(due[j].first)\n\t\t}\n", ""),
    # 구현 적대 리뷰 r1 의 생존 변이를 죽이는 시험을 더한 뒤(2026-11-23)
    ("M3r", "덮는 블록 중 첫 것(가장 이른 Since 아님)", NEW,
     "if !found || b.Since.Before(best.Since) {", "if !found {"),
    ("M7a", "운영자 전용 본문에서 할 일 문장 삭제", NEW,
     "body += \"이 차단은 운영자만 풀 수 있다 — 운영자 해제(`tossctl engine reconcile-resolve`, 엔진 정지 필요) 전까지 앱에서 직접 관리하라.\"",
     "body += \"이 차단은 운영자만 풀 수 있다.\""),
    ("M7b", "stale 본문에서 할 일 문장 삭제", NEW,
     "\"필요할 수 있다. 엔진 로그의 대사 상태를 확인하고 그 전까지 앱에서 직접 관리하라.\"", "\"필요할 수 있다.\""),
    ("M12r", "알림기 없음을 보낸 것으로 셈", NEW,
     "return false // 기록할 곳이 없으면 알린 것으로 세지 않음", "return true"),
    ("M19", "제목에서 label 대신 코드", NEW,
     "Title: d.label(p.Symbol) + \" 대사 차단으로", "Title: p.Symbol + \" 대사 차단으로"),
    ("M23", "풀린 차단의 래치 key 유지", NEW,
     "\td.blockReported = next\n",
     "\tif d.blockReported == nil {\n\t\td.blockReported = next\n\t} else {\n\t\tfor k := range next {\n\t\t\td.blockReported[k] = true\n\t\t}\n\t}\n"),
]
# M6 는 latch 를 꺼짐에서만 세우므로 켜짐에서도 세우는 짝 자리를 함께 바꿈(두 자리 변이).
M6_PAIR = ("\tif !d.opts.NotificationsEnabled {\n\t\tt = obs.EventExitPositionUnmanaged\n",
           "\tif d.unmanaged[p.ID] == nil {\n\t\td.unmanaged[p.ID] = map[string]bool{}\n\t}\n\td.unmanaged[p.ID][r.fact] = true\n"
           "\tif !d.opts.NotificationsEnabled {\n\t\tt = obs.EventExitPositionUnmanaged\n")
# M11 은 원래 자리의 호출을 지워야 옮긴 것이 됨.
M11_PAIR = ("\td.reportBlocked(ctx, held)\n}", "}")


def sh(cmd, cwd, env=None):
    return subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True)


def go_test(copy, env, cover=None):
    cmd = ["go", "test", "./internal/app/engine", "-count=1", "-run", RUN, "-v"]
    if cover:
        cmd += ["-coverprofile", cover]
    r = sh(cmd, copy, env)
    failed = sorted(set(re.findall(r"^\s*--- FAIL: (\S+)", r.stdout, re.M)))
    return r.returncode, failed, r.stdout + r.stderr


def hit_lines(profile, rel):
    hits = {}
    for line in open(profile):
        m = re.match(r".*/" + re.escape(rel) + r":(\d+)\.\d+,(\d+)\.\d+ \d+ (\d+)", line)
        if m:
            for n in range(int(m.group(1)), int(m.group(2)) + 1):
                hits[n] = max(hits.get(n, 0), int(m.group(3)))
    return hits


def main():
    scratch = pathlib.Path(sys.argv[1]).resolve()
    copy = scratch / f"a194-mut-{os.getpid()}"
    copy.mkdir(parents=True)
    head = sh(["git", "rev-parse", "HEAD"], ROOT).stdout.strip()
    assert sh(["git", "status", "--porcelain", "--", "internal"], ROOT).stdout == "", "internal/ 에 미커밋 변경 — 사본이 HEAD 와 다름"
    subprocess.run(f"git -C {ROOT} archive HEAD | tar -x -C {copy}", shell=True, check=True)
    env = dict(os.environ, GOFLAGS="-trimpath")
    print(json.dumps({"copy": str(copy), "head": head}), flush=True)

    cover = str(copy / "control.cover")
    rc, failed, out = go_test(copy, env, cover)
    print(json.dumps({"control": {"rc": rc, "failed": failed}}), flush=True)
    assert rc == 0 and not failed, "무변이 대조군이 GREEN 이 아님:\n" + out[-3000:]

    results = []
    for mid, desc, rel, old, new in MUTANTS:
        path = copy / rel
        src = path.read_text()
        assert src.count(old) == 1, f"{mid}: 원문 {src.count(old)} 회"
        line = src[: src.index(old)].count("\n") + 1
        reached = hit_lines(cover, rel).get(line, 0) > 0
        mutated = src.replace(old, new)
        if mid == "M6":
            assert mutated.count(M6_PAIR[0]) == 1
            mutated = mutated.replace(M6_PAIR[0], M6_PAIR[1])
        if mid == "M11":
            assert mutated.count(M11_PAIR[0]) == 1
            mutated = mutated.replace(M11_PAIR[0], M11_PAIR[1])
        path.write_text(mutated)
        try:
            rc, failed, out = go_test(copy, env)
        finally:
            path.write_text(src)
        assert path.read_text().count(old) == 1, f"{mid}: 원복 실패"
        build_failed = "build failed" in out or "[setup failed]" in out
        verdict = "BUILD-FAIL" if build_failed else ("CAUGHT" if rc != 0 else "SURVIVED")
        row = {"id": mid, "desc": desc, "file": rel, "line": line, "reached": reached, "verdict": verdict,
               "killed_by": failed}
        results.append(row)
        print(json.dumps(row, ensure_ascii=False), flush=True)
    (HERE / "results.json").write_text(json.dumps({"head": head, "control_green": True, "rows": results},
                                                  ensure_ascii=False, indent=1) + "\n")


if __name__ == "__main__":
    main()
