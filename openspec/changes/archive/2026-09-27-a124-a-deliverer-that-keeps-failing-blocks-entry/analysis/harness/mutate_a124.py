#!/usr/bin/env python3
"""a124 tasks 4.1 — 뮤테이션 원장. 각 변이를 **사본**에 걸고 a124 시험이 빨개지는지 잰다.

규율(기억: 뮤테이션은 대상에 닿아야 한다 · 원복은 baseline 이 맞아야 한다 · 한 사본에 한 판):
- 사본은 인자로 받은 연결 워크트리다(병행 로트의 미커밋 편집이 섞이지 않게). 실행 전에 사본의 대상 파일
  sha256 을 적고, 변이마다 원래 바이트로 되돌린 뒤 sha 를 다시 대조한다 — 어긋나면 멈춘다.
- 변이를 걸기 전에 **무변이 대조군**이 GREEN 이어야 한다. 아니면 모든 변이가 「잡힘」으로 찍힌다.
- 변이 문자열이 파일에 정확히 한 번 있어야 건다(닿지 않는 변이는 결과가 아니다 → NOT-APPLIED 로 멈춤).
- 한 번에 한 판. 사본 경로에 pid 를 쓰는 것은 호출자(워크트리 이름) 몫.

쓰기: python3 mutate_a124.py --copy <워크트리> --out ledger.tsv [--only M03,M07]
"""

from __future__ import annotations

import argparse
import hashlib
import subprocess
import sys
import time
from pathlib import Path

ENGINE = "./internal/app/engine"
JOURNAL = "./internal/journal"
EXECGW = "./internal/execgw"

# a124 시험 이름(정규식) — 패키지별.
ENGINE_TESTS = ("^(TestTheExecutor|TestAnAcknowledgement|TestARestartRelatches|TestANewRow|TestOutcomesThatWrote|"
                "TestARecordThatFinds|TestAClear|TestAFull|TestAnUnrelated|TestALatchOnly|TestAFailedEscalation|"
                "TestAnErrorJudgement|TestConsecutiveRecord|TestClaimFailures|TestAnotherRows|TestLeaseLost|"
                "TestListingFailures|TestASuccessfulListing|TestACancelledEngine|TestNewLines|TestAPublishedBut|"
                "TestADeliveryRecord|TestTheProductionExecutor|TestTheModeProjector|TestTheLedgerModeRow|"
                "TestThePermitted|TestTheMandated|TestA124Fixtures|TestAStaleListing|TestAClaimThatFinds|"
                "TestATruncatedListing|TestAStopNotify|TestAClearJustBefore|TestADeliveryRecordError|TestADeliveryAndAnEmpty|"
                "TestAStaleClear|TestAnAcknowledgementThatFailed|TestARearmWithout|TestAPublisherThatTimesOut|"
                "TestARestartFromAReopened|TestARealLeaseLoss|TestNoLineCarries|TestALeaseCanLapse|TestAFailedAcknowledgementKeeps|"
                "TestACancelledDeliveryJudgement|TestTheLatchLineIsWritten|TestAnAppliedRecordOnTheRow|TestACompleteListingForgets|"
                "TestTheRunRestartsAfter)")
JOURNAL_TESTS = "^(TestAFailedAttempt|TestTheJudgementIgnores|TestARearmed|TestAResultThatWrote|TestDeliverySelection|TestOnlyTestsReassign)"
EXECGW_TESTS = ("^(TestEveryClearRequest|TestNothingButThatReasons|TestAConditionalBlock|TestTheEpochComparison|"
                "TestOnlyTheAcknowledgement|TestTheEpochMethodsCallNothing)")

AD = "internal/app/engine/alertdelivery.go"
RETRY = "internal/execgw/retry.go"
CLAIM = "internal/journal/alert_claim.go"
OUTBOX = "internal/journal/outbox.go"

# (id, 설명, 파일, 원문, 변이, 패키지, 시험)
MUTATIONS = [
    ("M01", "판정 제거 — 한도에 닿아도 판정하지 않음", AD,
     "\t\tif res.Attempts < alertAttemptLimit {\n\t\t\treturn\n\t\t}",
     "\t\tif true {\n\t\t\treturn\n\t\t}", ENGINE, ENGINE_TESTS),
    ("M02", "승격 제거", AD,
     "\tif !escalate {\n\t\treturn\n\t}",
     "\tif true {\n\t\treturn\n\t}", ENGINE, ENGINE_TESTS),
    ("M03", "「뒤」 해제에서 승격도 버림", AD,
     "\t\tif _, inserted := d.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, epoch, detail); inserted {\n\t\t\td.reportLatch(id, detail)\n\t\t}",
     "\t\tapplied, inserted := d.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, epoch, detail)\n\t\tif inserted {\n\t\t\td.reportLatch(id, detail)\n\t\t}\n\t\tif !applied {\n\t\t\treturn\n\t\t}",
     ENGINE, ENGINE_TESTS),
    ("M04", "선택 순서 제거 — 오래된 것 먼저만", OUTBOX,
     "ORDER BY (attempts >= ?), id`\n\targs := []any{AlertPending, attemptLimit}",
     "ORDER BY id`\n\targs := []any{AlertPending}", JOURNAL, JOURNAL_TESTS),
    ("M04e", "선택 순서 제거 — 실행자 쪽에서 본다", OUTBOX,
     "ORDER BY (attempts >= ?), id`\n\targs := []any{AlertPending, attemptLimit}",
     "ORDER BY id`\n\targs := []any{AlertPending}", ENGINE, ENGINE_TESTS),
    ("M05", "한도 행 버림", OUTBOX,
     "ORDER BY (attempts >= ?), id`\n\targs := []any{AlertPending, attemptLimit}",
     "AND attempts < ? ORDER BY id`\n\targs := []any{AlertPending, attemptLimit}", ENGINE, ENGINE_TESTS),
    ("M06", "나열 값으로 판정 (Alert.Attempts+1)", AD, [
        ("\t\td.recordFailedAttempt(ctx, alert.ID, claim.Token, cause)",
         "\t\ta124MutListed = alert.Attempts\n\t\td.recordFailedAttempt(ctx, alert.ID, claim.Token, cause)"),
        ("\t\tif res.Attempts < alertAttemptLimit {", "\t\tif a124MutListed+1 < alertAttemptLimit {"),
        ("func (d *alertDeliverer) claimant() string {", "var a124MutListed int\n\nfunc (d *alertDeliverer) claimant() string {"),
     ], None, ENGINE, ENGINE_TESTS),
    ("M07a", "Outcome 검사 제거 — 0 행 결과도 판정", AD,
     "\tcase journal.SettleAlreadySettled, journal.SettleLeaseLost:\n\t\t// 승인(또는 남의 발송)이 먼저였음",
     "\tcase journal.SettleOutcome(-1):\n\t\t// 승인(또는 남의 발송)이 먼저였음", ENGINE, ENGINE_TESTS),
    ("M07b", "err 먼저 보지 않음 — 오류 결과를 Applied 로 읽음", AD,
     "\tif err != nil {\n\t\t// 오류의 SettleResult{} 는",
     "\tif false {\n\t\t// 오류의 SettleResult{} 는", ENGINE, ENGINE_TESTS),
    ("M08", "세대 비교 제거", RETRY,
     "\tif g.clearEpochs[reason] != epoch {\n\t\treturn false, false\n\t}",
     "\tif false {\n\t\treturn false, false\n\t}", EXECGW, EXECGW_TESTS),
    ("M08e", "세대 비교 제거 — 실행자 쪽에서 본다", RETRY,
     "\tif g.clearEpochs[reason] != epoch {\n\t\treturn false, false\n\t}",
     "\tif false {\n\t\treturn false, false\n\t}", ENGINE, ENGINE_TESTS),
    ("M09", "Clear 세대 증가 제거", RETRY,
     "\tg.clearEpochs[reason]++\n",
     "\t_ = g.clearEpochs[reason]\n", ENGINE, ENGINE_TESTS),
    ("M10", "세대를 「실제로 지웠을 때만」 증가", RETRY,
     "\tg.clearEpochs[reason]++\n\tif _, exists := g.latches[reason]; exists {\n\t\tdelete(g.latches, reason)\n\t\tg.revision++\n\t}",
     "\tif _, exists := g.latches[reason]; exists {\n\t\tg.clearEpochs[reason]++\n\t\tdelete(g.latches, reason)\n\t\tg.revision++\n\t}",
     ENGINE, ENGINE_TESTS),
    ("M11", "다른 사유의 해제가 세대를 올림", RETRY,
     "\tg.clearEpochs[reason]++\n",
     "\tg.clearEpochs[ReasonAlertUndelivered]++\n", EXECGW, EXECGW_TESTS),
    ("M12", "세대 읽기를 정산 전으로", AD, [
        ("\tres, err := d.led().MarkAlertAttemptFailed(ctx, id, token, cause)",
         "\tearly := d.readEpoch(id)\n\tres, err := d.led().MarkAlertAttemptFailed(ctx, id, token, cause)"),
        ("\t\td.judge(ctx, id, d.readEpoch(id), true, fmt.Sprintf(alertLatchAttemptLimit, alertAttemptLimit))",
         "\t\td.judge(ctx, id, early, true, fmt.Sprintf(alertLatchAttemptLimit, alertAttemptLimit))"),
     ], None, ENGINE, ENGINE_TESTS),
    ("M13", "오류 판정에 원장 승인 상태 확인을 넣음", AD,
     "\t\td.judge(ctx, id, d.readEpoch(id), true, alertLatchUnrecorded)\n\t\treturn\n\t}",
     "\t\tif row, lerr := d.Journal.LookupAlert(ctx, id); lerr == nil && row.State == journal.AlertAcknowledged {\n\t\t\treturn\n\t\t}\n\t\td.judge(ctx, id, d.readEpoch(id), true, alertLatchUnrecorded)\n\t\treturn\n\t}",
     ENGINE, ENGINE_TESTS),
    ("M14", "기록 실패 계수 제거", AD,
     "\tepoch := d.readEpoch(id)\n\tif d.recordRuns == nil {",
     "\treturn\n\tepoch := d.readEpoch(id)\n\tif d.recordRuns == nil {", ENGINE, ENGINE_TESTS),
    ("M15", "ctx 취소 제외 제거", AD,
     "func (d *alertDeliverer) countRecordFailure(ctx context.Context, id int64) {\n\tif ctx.Err() != nil {\n\t\treturn\n\t}",
     "func (d *alertDeliverer) countRecordFailure(ctx context.Context, id int64) {\n\tif false {\n\t\treturn\n\t}",
     ENGINE, ENGINE_TESTS),
    ("M16", "LeaseLost/AlreadySettled 로 계수 지움", AD,
     "\tcase journal.SettleAlreadySettled, journal.SettleLeaseLost:\n\t\t// 승인(또는 남의 발송)이 먼저였음 — 잠그지도 승격하지도 않음. 0 행을 썼으므로 계수도 그대로(R3).",
     "\tcase journal.SettleAlreadySettled, journal.SettleLeaseLost:\n\t\td.forgetRecordRun(id)",
     ENGINE, ENGINE_TESTS),
    ("M16b", "반납 Applied 로 계수 지움", AD,
     "\tif _, err := d.led().ReleaseAlertClaim(relCtx, id, token); err != nil {",
     "\tif rr, err := d.led().ReleaseAlertClaim(relCtx, id, token); err == nil && rr.Outcome == journal.SettleApplied {\n\t\td.forgetRecordRun(id)\n\t} else if err != nil {",
     ENGINE, ENGINE_TESTS),
    ("M17", "「이미 정산됨」 거름 제거", AD,
     "\t\t// PENDING 을 떠난 것이 관측됨 — 이 행의 기록 실패 계수도 끝(D8).\n\t\td.forgetRecordRun(alert.ID)",
     "\t\t// PENDING 을 떠난 것이 관측됨 — 이 행의 기록 실패 계수도 끝(D8).", ENGINE, ENGINE_TESTS),
    ("M18", "전달 정산 실패 판정 제거", AD,
     "\t\td.judge(ctx, id, d.readEpoch(id), true, alertLatchUnrecorded)\n\t\treturn\n\t}",
     "\t\treturn\n\t}", ENGINE, ENGINE_TESTS),
    ("M19", "한도째 판정보다 리셋이 먼저 (AA2 순서 뒤집기)", AD,
     "\tif run.count+1 >= alertAttemptLimit {\n\t\treturn alertFailureRun{}, true, run.epoch\n\t}\n\tif epoch != run.epoch {\n\t\treturn alertFailureRun{count: 1, epoch: epoch}, false, 0\n\t}",
     "\tif epoch != run.epoch {\n\t\treturn alertFailureRun{count: 1, epoch: epoch}, false, 0\n\t}\n\tif run.count+1 >= alertAttemptLimit {\n\t\treturn alertFailureRun{}, true, run.epoch\n\t}",
     ENGINE, ENGINE_TESTS),
    ("M20", "나열 성공이 나열 계수를 안 지움", AD,
     "\td.listRun = alertFailureRun{}\n\tif len(pending) < d.batch() {",
     "\tif len(pending) < d.batch() {", ENGINE, ENGINE_TESTS),
    ("M21", "잘린 나열에서도 거름", AD,
     "\tif len(pending) < d.batch() {\n\t\t// 잘리지 않은(완전한) 나열에 없는 행은",
     "\tif true {\n\t\t// 잘리지 않은(완전한) 나열에 없는 행은", ENGINE, ENGINE_TESTS),
    ("M22", "승격 실패 무조건 차단 제거", AD,
     "\tif err := d.escalate(ctx, id); err != nil && d.Gate != nil {\n\t\td.Gate.Block(execgw.ReasonAlertUndelivered, alertLatchEscalationFailed)\n\t}",
     "\t_ = d.escalate(ctx, id)", ENGINE, ENGINE_TESTS),
    ("M23", "publisher 없음을 세지 않음 (D3)", AD,
     "\t\tperr, cause = errors.New(alertNoPublisherCause), alertNoPublisherCause",
     "\t\td.release(ctx, alert.ID, claim.Token)\n\t\treturn", ENGINE, ENGINE_TESTS),
    ("M24", "실패 기록 NotFound 가 승격", AD,
     "\t\td.judge(ctx, id, d.readEpoch(id), false, alertLatchUnaccounted)",
     "\t\td.judge(ctx, id, d.readEpoch(id), true, alertLatchUnaccounted)", ENGINE, ENGINE_TESTS),
    ("M25", "가산 읽기 없이 Attempts 0", CLAIM,
     "\t\treturn SettleResult{Outcome: SettleApplied, Attempts: attempts}, nil",
     "\t\t_ = attempts\n\t\treturn SettleResult{Outcome: SettleApplied}, nil", JOURNAL, JOURNAL_TESTS),
    ("M26", "가산 읽기를 커밋 뒤 j.db 로", CLAIM,
     "\t\tattempts, err := readSettledAttemptsTx(ctx, tx, id)\n\t\tif err != nil {\n\t\t\treturn SettleResult{}, fmt.Errorf(\"journal: reading the attempts %s alert %d committed: %w\", what, id, err)\n\t\t}\n\t\tif err := tx.Commit(); err != nil {\n\t\t\treturn SettleResult{}, fmt.Errorf(\"journal: committing %s of alert %d: %w\", what, id, err)\n\t\t}",
     "\t\tif err := tx.Commit(); err != nil {\n\t\t\treturn SettleResult{}, fmt.Errorf(\"journal: committing %s of alert %d: %w\", what, id, err)\n\t\t}\n\t\tvar attempts int\n\t\t_ = j.db.QueryRowContext(ctx, `SELECT attempts FROM alert_outbox WHERE id = ?`, id).Scan(&attempts)",
     JOURNAL, JOURNAL_TESTS),
    ("M27", "전달 정산 오류를 취소 중이면 판정하지 않음 (codex I1 이전 판)", AD,
     "\t\td.logf(obs.EventAlertUndelivered, err, \"settling a delivered alert failed\", \"alert_id\", id)\n",
     "\t\td.logf(obs.EventAlertUndelivered, err, \"settling a delivered alert failed\", \"alert_id\", id)\n\t\tif ctx.Err() != nil {\n\t\t\treturn\n\t\t}\n",
     ENGINE, ENGINE_TESTS),
    ("M28", "게이트 잠금 안에서 밖을 부름 (잠금 범위 — 승격을 잠금 안으로의 구조 대응)", RETRY,
     "\tif g.clearEpochs[reason] != epoch {\n\t\treturn false, false\n\t}",
     "\t_ = g.clk.Now()\n\tif g.clearEpochs[reason] != epoch {\n\t\treturn false, false\n\t}", EXECGW, EXECGW_TESTS),
    ("M29", "세대 비교와 삽입을 두 잠금으로 나눔 (split-lock)", RETRY,
     "\tif g.clearEpochs[reason] != epoch {\n\t\treturn false, false\n\t}\n\tif _, exists := g.latches[reason]; !exists {",
     "\tif g.clearEpochs[reason] != epoch {\n\t\treturn false, false\n\t}\n\tg.mu.Unlock()\n\ttime.Sleep(time.Microsecond)\n\tg.mu.Lock()\n\tif _, exists := g.latches[reason]; !exists {",
     EXECGW, EXECGW_TESTS),
    ("M30", "취소 중이면 승격을 건너뜀 (승격 실패 대체 차단도 사라짐)", AD,
     "\tif strings.TrimSpace(d.AccountRef) == \"\" {\n\t\treturn nil\n\t}",
     "\tif strings.TrimSpace(d.AccountRef) == \"\" || ctx.Err() != nil {\n\t\treturn nil\n\t}", ENGINE, ENGINE_TESTS),
    ("M31", "래치 줄을 이미 있던 래치에도 씀 (D9)", AD,
     "\t\tif _, inserted := d.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, epoch, detail); inserted {",
     "\t\tif applied, _ := d.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, epoch, detail); applied {",
     ENGINE, ENGINE_TESTS),
    ("M32", "그 행의 실패 기록 Applied 가 연속을 안 지움", AD,
     "\t\t// 원장이 이 행에 썼음 — 기록 실패의 연속은 끝.\n\t\td.forgetRecordRun(id)\n",
     "\t\t// 원장이 이 행에 썼음 — 기록 실패의 연속은 끝.\n", ENGINE, ENGINE_TESTS),
    ("M33", "완전한 나열이 거르지 않음", AD,
     "\t\td.pruneRecordRuns(pending)\n", "\t\t_ = pending\n", ENGINE, ENGINE_TESTS),
    ("M34", "기록 연속 판정 뒤 계수를 안 지움", AD,
     "\tdelete(d.recordRuns, id)\n\td.judge(ctx, id, blockEpoch, true, fmt.Sprintf(alertLatchRecordRun, alertAttemptLimit))",
     "\td.recordRuns[id] = alertFailureRun{count: alertAttemptLimit - 1, epoch: blockEpoch}\n\td.judge(ctx, id, blockEpoch, true, fmt.Sprintf(alertLatchRecordRun, alertAttemptLimit))",
     ENGINE, ENGINE_TESTS),
    ("M35", "나열 연속 판정 뒤 계수를 안 지움", AD,
     "\td.listRun = alertFailureRun{}\n\td.judge(ctx, alertNoRow,",
     "\td.listRun = alertFailureRun{count: alertAttemptLimit - 1, epoch: blockEpoch}\n\td.judge(ctx, alertNoRow,",
     ENGINE, ENGINE_TESTS),
]


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(copy: Path, pkg: str, tests: str, tags: str = "") -> tuple[bool, str]:
    cmd = ["go", "test", "-count=1", "-timeout", "15m", "-run", tests]
    if tags:
        cmd += ["-tags", tags]
    cmd.append(pkg)
    proc = subprocess.run(cmd, cwd=copy, capture_output=True, text=True)
    if "[build failed]" in proc.stdout + proc.stderr or "[setup failed]" in proc.stdout + proc.stderr:
        # 컴파일 실패는 「잡힘」이 아니다 — 변이가 시험에 닿지 않았다.
        raise SystemExit(f"BUILD-FAIL: {(proc.stdout + proc.stderr)[-800:]}")
    tail = "\n".join(l for l in (proc.stdout + proc.stderr).splitlines()
                     if l.startswith(("--- FAIL", "FAIL", "ok", "panic", "#")) or "cannot" in l or "undefined" in l)[-600:]
    return proc.returncode == 0, tail


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--copy", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--only", default="")
    args = ap.parse_args()
    copy = Path(args.copy).resolve()
    only = set(filter(None, args.only.split(",")))
    files = sorted({m[2] for m in MUTATIONS})
    originals = {f: (copy / f).read_bytes() for f in files}
    start_sha = {f: sha(copy / f) for f in files}

    out = open(args.out, "w", encoding="utf-8")
    out.write("id\tverdict\tsecs\tdescription\tevidence\n")

    # 무변이 대조군 — 세 패키지 모두 GREEN 이어야 한다(태그 시험 포함).
    for pkg, tests, tags in ((ENGINE, ENGINE_TESTS, "tossos_testseams"), (JOURNAL, JOURNAL_TESTS, ""), (EXECGW, EXECGW_TESTS, "")):
        ok, tail = run(copy, pkg, tests, tags)
        out.write(f"CONTROL\t{'GREEN' if ok else 'RED'}\t-\t{pkg}\t{tail.replace(chr(10), ' | ')}\n")
        out.flush()
        if not ok:
            print(f"control is RED for {pkg} — every mutant would read as caught; stopping", file=sys.stderr)
            return 2

    for mid, desc, f, old, new, pkg, tests in MUTATIONS:
        if only and mid not in only:
            continue
        path = copy / f
        text = originals[f].decode("utf-8")
        edits = old if isinstance(old, list) else [(old, new)]
        for anchor, _ in edits:
            if text.count(anchor) != 1:
                out.write(f"{mid}\tNOT-APPLIED\t-\t{desc}\tanchor found {text.count(anchor)} times\n")
                out.flush()
                print(f"{mid}: anchor count {text.count(anchor)} — the mutant does not reach its target; stopping", file=sys.stderr)
                return 3
        for anchor, replacement in edits:
            text = text.replace(anchor, replacement)
        path.write_text(text, encoding="utf-8")
        t0 = time.time()
        try:
            ok, tail = run(copy, pkg, tests, "tossos_testseams" if pkg == ENGINE else "")
        finally:
            path.write_bytes(originals[f])
        if sha(path) != start_sha[f]:
            print(f"{mid}: restore mismatch on {f}; stopping", file=sys.stderr)
            return 4
        verdict = "SURVIVED" if ok else "CAUGHT"
        out.write(f"{mid}\t{verdict}\t{time.time() - t0:.0f}\t{desc}\t{tail.replace(chr(10), ' | ')}\n")
        out.flush()
        print(f"{mid}\t{verdict}\t{desc}")
    out.close()
    return 0


if __name__ == "__main__":
    sys.exit(main())
