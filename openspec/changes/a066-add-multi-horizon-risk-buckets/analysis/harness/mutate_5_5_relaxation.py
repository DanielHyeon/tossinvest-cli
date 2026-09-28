#!/usr/bin/env python3
"""a066 5.5 완화(D8) 뮤테이션 — 사본에서만 변이하고 시험을 돌려 CAUGHT/SURVIVED 를 원장에 적음.

사용: python3 mutate_5_5_relaxation.py <ledger.tsv> [mutant-id ...]

M* 는 journal API(원장 시험), E* 는 엔진 endpoint(엔진 시험), C* 는 tossctl 명령(CLI 시험) 변이임. 대조군은 변이가 쓰는
시험 묶음마다 먼저 돎.

규율(기억: 뮤테이션은 대상에 닿아야 한다 · 원복은 baseline 이 맞아야 한다):
- 사본 디렉터리는 pid 를 달고 스크래치(TMPDIR)에 둠. 작업 트리는 절대 변이하지 않음.
- 사본 = 저장소 추적 파일(go.mod go.sum internal cmd tools)의 **작업 트리 내용** 복사.
- 무변이 대조군이 GREEN 이 아니면 멈춤(사본이 깨졌으면 모든 변이가 CAUGHT 로 찍힘).
- 변이마다: 치환 대상 문자열이 사본 파일에 정확히 한 번 있어야 함(아니면 NOT-APPLIED — 닿지 않은 변이는 판정 없음).
  변이 뒤 원본 바이트로 되돌리고, 되돌린 바이트가 원본과 같은지 단언함.
- 판정은 `go test` rc 로만(파이프 없음). 출력은 사본 옆 로그 파일.
"""
from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())

JOURNAL_TESTS = ("^(TestA066|TestEntryLossLock|TestMigrationV3[2-5]|TestJournalAuditActions|TestRefuseEntryUnderLossLock|"
                 "TestSchemaTablesAndColumns|TestTheReadOnlyHandleHasNoWriteMethods)")

ENGINE_TESTS = ("./internal/app/engine/", "^TestA066")
CLI_TESTS = ("./cmd/tossctl/", "^(TestEngineRisk|TestEngineEntryLock|TestEngineRelaxation|TestMutatingAnnotationOnTradeCommands)")
SUITES = {"M": ("./internal/journal/", JOURNAL_TESTS), "E": ENGINE_TESTS, "C": CLI_TESTS}

RELAX = "internal/journal/risk_bucket_relaxation.go"
ECMD = "internal/app/engine/risk_relaxation_command.go"
ETRANS = "internal/app/engine/risk_relaxation_transport.go"
ESTART = "internal/app/engine/position_policy_transport.go"
CLI = "cmd/tossctl/engine_risk_relaxation.go"
LOCK = "internal/journal/risk_bucket_entry_loss_lock.go"
V35 = "internal/journal/risk_bucket_relaxation_v35.sql"

OWN_UNTRACKED = [
    "internal/riskrelaxation/riskrelaxation.go",
    "internal/positionpolicyrpc/risk_relaxation_client.go",
    "internal/app/engine/risk_relaxation_command.go",
    "internal/app/engine/risk_relaxation_transport.go",
    "internal/app/engine/a066_risk_relaxation_test.go",
    "internal/journal/a066_relaxation_evidence_test.go",
]

# (id, file, old, new, what it breaks)
MUTANTS: list[tuple[str, str, str, str, str]] = [
    ("M01", LOCK, "INSERT INTO risk_bucket_entry_loss_lock_events(lock_seq,kind,cause,recorded_at) VALUES(?,'REAFFIRM',?,?)",
     "SELECT ?,?,?", "REAFFIRM event not written (a release approved before a new tightening would pass)"),
    ("M02", LOCK, "\t\tAND NOT EXISTS (SELECT 1 FROM risk_bucket_entry_loss_lock_releases r WHERE r.lock_seq=l.lock_seq)\n", "",
     "open lock ignores the release row (release has no effect on entry)"),
    ("M03", RELAX, "if !found || open.Seq != req.LockSeq {", "if !found {", "lock number binding dropped"),
    ("M04", RELAX, "if lastEvent != req.ExpectedLastEvent {", "if false {", "REAFFIRM staleness binding dropped"),
    ("M05", RELAX, "case strings.TrimSpace(actor) != RelaxationActorOperator:", "case false:", "non-operator actor accepted"),
    ("M06", RELAX, "case strings.TrimSpace(approval) == \"\":", "case false:", "blank approval accepted"),
    ("M07", RELAX, "case strings.TrimSpace(reason) == \"\":", "case false:", "blank reason accepted"),
    ("M08", RELAX, "case nilAuditor(auditor):", "case false:", "nil auditor accepted"),
    ("M09", RELAX, "if err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt, detail); err != nil {",
     "if err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt, detail); err != nil && false {",
     "entry release commits although the audit line failed"),
    ("M10", RELAX, "RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt,", "RecordAction(AuditActionOverageLatchRelease, setting, relaxationAuditAttempt,",
     "entry release audited under the wrong action"),
    ("M11", RELAX, "if overage == 0 {", "if false {", "latch release with no RISK_OVERAGE latch accepted"),
    ("M12", RELAX, "if persisted != strings.TrimSpace(req.ExpectedStateDigest) {", "if false {", "state digest binding dropped (ABA)"),
    ("M13", RELAX, "UPDATE risk_bucket_owners SET risk_overage_latched=0 WHERE", "UPDATE risk_bucket_owners SET risk_overage_latched=0,unknown_actual_latched=0 WHERE",
     "latch release also clears UNKNOWN_ACTUAL_RISK"),
    ("M14", RELAX, "UPDATE risk_bucket_reservations SET risk_overage_latched=0 WHERE", "UPDATE risk_bucket_reservations SET risk_overage_latched=risk_overage_latched WHERE",
     "reservation RISK_OVERAGE flags not cleared"),
    ("M15", RELAX, "if err := j.recordRiskBucketStateTx(ctx, tx, key, \"OVERAGE_LATCH_RELEASED\"", "if err := func(...any) error { return nil }(ctx, tx, key, \"OVERAGE_LATCH_RELEASED\"",
     "state not resealed after the latch release"),
    ("M16", RELAX, "if err := req.Auditor.RecordAction(AuditActionOverageLatchRelease, setting, relaxationAuditAttempt, detail); err != nil {",
     "if err := req.Auditor.RecordAction(AuditActionOverageLatchRelease, setting, relaxationAuditAttempt, detail); err != nil && false {",
     "latch release commits although the audit line failed"),
    ("M17", V35, "CREATE TRIGGER risk_bucket_entry_loss_lock_one_open BEFORE INSERT ON risk_bucket_entry_loss_locks\nWHEN EXISTS (",
     "CREATE TRIGGER risk_bucket_entry_loss_lock_one_open BEFORE INSERT ON risk_bucket_entry_loss_locks\nWHEN 0 AND EXISTS (",
     "one-open trigger never fires"),
    ("M18", RELAX, "COALESCE((SELECT MAX(e.event_seq) FROM risk_bucket_entry_loss_lock_events e WHERE e.lock_seq=l.lock_seq),0)", "0",
     "show reports no REAFFIRM event"),
    ("M19", RELAX, "if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {", "if err := error(nil); err != nil {",
     "latch release skips the ledger-vs-seal verification"),
    ("M20", RELAX, "if !validEntryLossLockScope(req.AccountRef, req.Market, req.Horizon) || req.LockSeq <= 0 || req.ExpectedLastEvent < 0 {",
     "if false {", "entry release request shape unchecked"),
    ("M21", RELAX, "case at.IsZero():", "case false:", "zero release time accepted"),
    ("M22", RELAX, "\tif errors.Is(err, sql.ErrNoRows) {\n\t\treturn RiskOverageLatchReleaseRecord{}, fmt.Errorf(\"%w: no active owner",
     "\tif errors.Is(err, sql.ErrNoRows) && false {\n\t\treturn RiskOverageLatchReleaseRecord{}, fmt.Errorf(\"%w: no active owner",
     "missing owner reported as storage error instead of stale"),
    ("M23", RELAX, "case nilAuditor(auditor):", "case auditor == nil:", "a typed-nil audit log (*audit.Log)(nil) accepted (CX-1)"),
       # --- 리뷰 R3(2026-09-29) 생존 변이와 Q6 · 어휘 수리의 변이 ---
    ("M24", RELAX, "\tif err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt, detail); err != nil {",
     "\tAuditActionEntryLockRelease := \"silent\"\n\tif err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt, detail); err != nil {",
     "census: a local variable shadows the audit action"),
    ("M25", RELAX, "\tif err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt, detail); err != nil {",
     "\trec := req.Auditor.RecordAction\n\t_ = AuditActionEntryLockRelease\n\tif err := rec(\"silent\", setting, relaxationAuditAttempt, detail); err != nil {",
     "census: RecordAction called through a method value"),
    ("M26", RELAX, "\tif err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt, detail); err != nil {",
     "\tconst AuditActionEntryLockRelease = \"silent\"\n\tif err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditAttempt, detail); err != nil {",
     "census: a function-local constant shadows the audit action"),
    ("M27", RELAX, "_ = req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, relaxationAuditNotCommitted,",
     "_ = func(...any) error { return nil }(AuditActionEntryLockRelease, setting, relaxationAuditNotCommitted,",
     "Q6: no compensating audit line after a failed commit"),
    ("M28", RELAX, "relaxationAuditAttempt      = \"release_attempt\"", "relaxationAuditAttempt      = \"released\"",
     "Q6: the pre-commit line claims completion"),
    ("M29", RELAX, "// relaxationSchemaVersion 은", "var forgedRecord = func(a RiskRelaxationAuditor, s, v, d string) error { return a.RecordAction(\"forged.action\", s, v, d) }\n\n// relaxationSchemaVersion 은",
     "census: a package-level var initializer calls RecordAction with a literal (re-review P2)"),
    ("M30", RELAX, "// relaxationSchemaVersion 은", "var forgedRecord = RiskRelaxationAuditor.RecordAction\n\n// relaxationSchemaVersion 은",
     "census: a package-level method expression (re-review P2)"),
    ("MX01", V35, "BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock events are immutable'); END;", "BEGIN SELECT 1; END;", "REAFFIRM events updatable"),
    ("MX02", V35, "BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock events cannot be deleted'); END;", "BEGIN SELECT 1; END;", "REAFFIRM events deletable"),
    ("MX03", V35, "BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock releases are immutable'); END;", "BEGIN SELECT 1; END;", "lock releases updatable"),
    ("MX04", V35, "BEGIN SELECT RAISE(ABORT,'risk bucket entry loss lock releases cannot be deleted'); END;", "BEGIN SELECT 1; END;", "lock releases deletable"),
    ("MX05", V35, "BEGIN SELECT RAISE(ABORT,'risk bucket latch releases are immutable'); END;", "BEGIN SELECT 1; END;", "latch releases updatable"),
    ("MX06", V35, "BEGIN SELECT RAISE(ABORT,'risk bucket latch releases cannot be deleted'); END;", "BEGIN SELECT 1; END;", "latch releases deletable"),
    ("MX07", V35, "lock_seq INTEGER NOT NULL UNIQUE REFERENCES", "lock_seq INTEGER NOT NULL REFERENCES", "two releases of one lock"),
    ("MX12", RELAX, "COALESCE((SELECT s.state_digest FROM risk_bucket_state_snapshots s", "COALESCE((SELECT '' FROM risk_bucket_state_snapshots s", "show reports no state digest"),
    ("MX17", RELAX, "UPDATE risk_bucket_reservations SET risk_overage_latched=0 WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?",
     "UPDATE risk_bucket_reservations SET risk_overage_latched=0 WHERE account_ref=? AND market=? AND symbol=? AND (owner_prospective_generation=? OR 1)",
     "release clears other generations of the symbol"),
    ("MX18", RELAX, "UPDATE risk_bucket_reservations SET risk_overage_latched=0 WHERE", "UPDATE risk_bucket_reservations SET risk_overage_latched=0,overage_minor='0' WHERE",
     "release erases the overage amounts"),
    ("MX20", RELAX, "prospective_generation=? AND released_at IS NULL`,\n\t\tkey.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&overage)",
     "prospective_generation=?`,\n\t\tkey.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&overage)",
     "latch release on a released owner"),
    ("E11", ECMD, "case errors.Is(err, journal.ErrRiskBucketReplayMismatch):", "case false:", "state mismatch reaches the CLI as outcome unknown"),
    ("E12", ECMD, "case errors.Is(err, journal.ErrRiskRelaxationAuditFailed):", "case false:", "audit write failure reaches the CLI as outcome unknown"),
    ("EX13", ETRANS, "mux.HandleFunc(RiskRelaxationEntryLockPath, server.auth(token,\n\t\triskRelaxationRequestHandler(commands.ReleaseEntryLossLock)))",
     "mux.HandleFunc(RiskRelaxationEntryLockPath, (\n\t\triskRelaxationRequestHandler(commands.ReleaseEntryLossLock)))", "entry route without bearer auth"),
    ("EX14", ETRANS, "mux.HandleFunc(RiskRelaxationLatchPath, server.auth(token,\n\t\triskRelaxationRequestHandler(commands.ReleaseRiskOverageLatch)))",
     "mux.HandleFunc(RiskRelaxationLatchPath, (\n\t\triskRelaxationRequestHandler(commands.ReleaseRiskOverageLatch)))", "latch route without bearer auth"),
    ("EX16", ECMD, "\tauditor, err := s.relaxationAuditor()\n\tif err != nil {\n\t\treturn riskrelaxation.Result{}, err\n\t}\n\toperator, reason, err := relaxationOperator(req.Operator, req.Reason)\n\tif err != nil {\n\t\treturn riskrelaxation.Result{}, err\n\t}\n\towner :=",
     "\tauditor := journal.RiskRelaxationAuditor(s.audit)\n\toperator, reason, err := relaxationOperator(req.Operator, req.Reason)\n\tif err != nil {\n\t\treturn riskrelaxation.Result{}, err\n\t}\n\towner :=",
     "latch path skips the engine audit-log check"),
    ("CX19", CLI, "errors.Is(err, riskrelaxation.ErrStateMismatch), errors.Is(err, riskrelaxation.ErrAuditUnavailable),", "errors.Is(err, riskrelaxation.ErrStateMismatch),",
     "CLI tells an audit refusal as an unknown outcome"),
    ("CX21", CLI, "return journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: path})", "return journal.Open(ctx, journal.Options{Path: path})",
     "risk-latch-show opens the journal as a writer"),
 ("E01", ECMD, "if s.audit == nil {", "if false {", "engine passes a nil audit log (audit line silently skipped)"),
    ("E02", ECMD, "_, err := repo.EnqueueAlert(context.WithoutCancel(ctx), journal.Alert{", "_, err := func(context.Context, journal.Alert) (int64, error) { return 0, nil }(context.WithoutCancel(ctx), journal.Alert{",
     "no notice is enqueued but the result says notified"),
    ("E03", ECMD, "\tif err != nil {\n\t\tresult.NotifyError = err.Error()", "\tif false {\n\t\tresult.NotifyError = err.Error()",
     "a failed notice is reported as notified"),
    ("E04", ECMD, "case errors.Is(err, journal.ErrRiskRelaxationStale):", "case false:", "stale refusal loses its type over the wire"),
    ("E05", ECMD, "return operator, strings.TrimSpace(reason) + \" (operator \" + operator + \")\", nil", "return operator, strings.TrimSpace(reason), nil",
     "operator name not recorded with the release"),
    ("E06", ECMD, "\tif operator == \"\" {\n\t\treturn \"\", \"\", fmt.Errorf(\"%w: the operator is named\"", "\tif false {\n\t\treturn \"\", \"\", fmt.Errorf(\"%w: the operator is named\"",
     "blank operator accepted by the engine"),
    ("E07", ESTART, "if relaxations, ok := commands.(riskRelaxationCommands); ok {", "if relaxations, ok := commands.(riskRelaxationCommands); ok && false {",
     "relaxation routes not registered"),
    ("E08", ETRANS, "case errors.Is(err, riskrelaxation.ErrStale):\n\t\tstatus, code = http.StatusPreconditionFailed, \"stale\"", "case false:\n\t\tstatus, code = http.StatusPreconditionFailed, \"stale\"",
     "stale mapped to internal on the wire"),
    ("E10", ECMD, "repo.EnqueueAlert(context.WithoutCancel(ctx),", "repo.EnqueueAlert(ctx,",
     "notice shares the request cancellation (client hang-up drops the notice)"),
    ("C01", CLI, "{\"--account\", c.account}, {\"--approval\", c.approval}, {\"--operator\", c.operator}, {\"--reason\", c.reason},",
     "{\"--account\", c.account}, {\"--operator\", c.operator}, {\"--reason\", c.reason},", "CLI does not refuse a blank approval before the engine"),
    ("C02", CLI, "\tif !result.Notified {", "\tif false {", "CLI hides a failed notice"),
    ("C03", CLI, "case errors.Is(err, riskrelaxation.ErrStale), errors.Is(err, riskrelaxation.ErrInvalidRequest),", "case errors.Is(err, riskrelaxation.ErrInvalidRequest),",
     "CLI tells a stale refusal as an unknown outcome"),
    ("C04", CLI, "\tif err != nil || client == nil {\n\t\treturn fmt.Errorf(\"engine %s: the engine is not running", "\tif client == nil {\n\t\treturn fmt.Errorf(\"engine %s: the engine is not running",
     "CLI ignores a dial error"),
    ("C05", CLI, "if lockSeq <= 0 || expectEvent < 0 {", "if false {", "CLI sends a release with no binding"),
    ("C06", CLI, "\"mutating\": \"true\"},\n\t\tSilenceUsage: true,\n\t\tArgs:         cobra.NoArgs,\n\t\tRunE: func(cmd *cobra.Command, _ []string) error {\n\t\t\tmarket, err := common.check(\"entry-lock-release\")",
     "\"mutating\": \"false\"},\n\t\tSilenceUsage: true,\n\t\tArgs:         cobra.NoArgs,\n\t\tRunE: func(cmd *cobra.Command, _ []string) error {\n\t\t\tmarket, err := common.check(\"entry-lock-release\")",
     "entry-lock-release not declared mutating"),
]


def copy_tree(dst: Path) -> None:
    files = subprocess.run(["git", "-C", str(ROOT), "ls-files", "--", "go.mod", "go.sum", "internal", "cmd", "tools"],
                           capture_output=True, text=True, check=True).stdout.split("\n")
    for rel in filter(None, files):
        src = ROOT / rel
        if not src.is_file():
            continue
        target = dst / rel
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(src, target)


def run_tests(copy: Path, log: Path, suite: str) -> int:
    package, pattern = SUITES[suite]
    env = dict(os.environ, GOFLAGS="-trimpath", GOCACHE=str(copy.parent / "gocache"))
    with log.open("w") as out:
        return subprocess.run(["go", "test", "-count=1", "-run", pattern, package],
                              cwd=copy, env=env, stdout=out, stderr=subprocess.STDOUT).returncode


def main() -> int:
    ledger = Path(sys.argv[1]).resolve()
    only = set(sys.argv[2:])
    base = Path(tempfile.mkdtemp(prefix=f"a066-relax-mut-{os.getpid()}-"))
    copy = base / "tree"
    copy_tree(copy)
    # 사본은 추적 파일만 복사함 — 이 로트의 아직 추적 안 된 새 파일도 넣음(없으면 조용히 건너뜀).
    for rel in OWN_UNTRACKED:
        src = ROOT / rel
        if src.is_file():
            (copy / rel).parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(src, copy / rel)
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short", "HEAD"], capture_output=True, text=True, check=True).stdout.strip()
    rows = [f"TREE\tHEAD {head} + working tree\tcopy {copy}"]
    chosen = [m for m in MUTANTS if not only or m[0] in only]
    for suite in sorted({m[0][0] for m in chosen}):
        rc = run_tests(copy, base / f"control-{suite}.log", suite)
        rows.append(f"CONTROL-{suite}\t{'GREEN' if rc == 0 else 'RED rc=' + str(rc)}\t{' '.join(SUITES[suite])}")
        if rc != 0:
            ledger.write_text("\n".join(rows) + "\n")
            print(f"control {suite} not GREEN (rc {rc}); see {base}", file=sys.stderr)
            return 2
    for mid, rel, old, new, what in chosen:
        path = copy / rel
        original = path.read_bytes()
        text = original.decode()
        if text.count(old) != 1:
            rows.append(f"{mid}\tNOT-APPLIED\t{rel}\t{what}\toccurrences={text.count(old)}")
            continue
        path.write_text(text.replace(old, new))
        try:
            rc = run_tests(copy, base / f"{mid}.log", mid[0])
        finally:
            path.write_bytes(original)
        assert path.read_bytes() == original, f"{mid}: revert failed"
        # 컴파일 실패는 시험이 잡은 것이 아님 — 변이가 빌드를 깨면 BUILD-FAILED 로 적고 판정에서 뺌(R3 지적: M22 거짓 CAUGHT).
        log_text = (base / f"{mid}.log").read_text(errors="replace")
        if rc != 0 and ("[build failed]" in log_text or "[setup failed]" in log_text):
            verdict = "BUILD-FAILED"
        else:
            verdict = "CAUGHT" if rc != 0 else "SURVIVED"
        rows.append(f"{mid}\t{verdict}\t{rel}\t{what}\trc={rc}")
        print(rows[-1], flush=True)
    ledger.write_text("\n".join(rows) + "\n")
    print(f"logs: {base}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
