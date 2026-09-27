#!/usr/bin/env python3
"""a066 5.6.1 변이 하네스 — 공유 bucket 정합(F2 policy record 결속, F1 원장 사용량 대조)을 사본에서 변이함.

사용: python3 mutate_5_6_1.py <scratch-dir> <own-untracked-file>... [--only REGEX]

- 사본: git 추적 파일(작업 트리 내용) + 이 로트의 미추적 파일만 복사함(남의 미추적 RED 파일이 사본에 들어오지 않게).
- 원장 첫 줄: HEAD sha 와 사본에 딸려 간 미커밋 추적 파일 목록.
- 무변이 대조군이 GREEN 이 아니면 멈춤. 변이마다 정확히 한 곳(count==1)을 바꾸고, 실패한 시험 이름을 적음.
- 한 번에 한 판.
"""
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
J = "internal/journal/"
RB = "internal/riskbucket/production_snapshot_authority.go"
USAGE = J + "risk_bucket_usage.go"
REC = J + "risk_bucket_policy_records.go"
SQL34 = J + "risk_bucket_policy_records_v34.sql"
FILL = J + "risk_bucket_fill.go"
CRA = J + "risk_bucket.go"
ISS = J + "risk_bucket_issuance.go"
F1_CALL = "if err := refuseStaleBucketUsage(ctx, tx, plan.Owner.Key.AccountID, plan.Admission.Buckets); err != nil {"

MUTANTS = [
    ("N01 CRA stale check removed", CRA, F1_CALL, "if err := func(...any) error { return nil }(ctx, tx, plan.Owner.Key.AccountID, plan.Admission.Buckets); err != nil {"),
    ("N02 fresh stale check removed", ISS, F1_CALL, "if err := func(...any) error { return nil }(ctx, tx, plan.Owner.Key.AccountID, plan.Admission.Buckets); err != nil {"),
    ("N03 CRA stale check account<-LaneID", CRA, F1_CALL, F1_CALL.replace("plan.Owner.Key.AccountID,", "plan.Owner.LaneID,")),
    ("N04 fresh stale check account<-LaneID", ISS, F1_CALL, F1_CALL.replace("plan.Owner.Key.AccountID,", "plan.Owner.LaneID,")),
    ("N05 comparison accepts understated usage", USAGE, "if claimed.Cmp(ledger) < 0 {", "if claimed.Cmp(ledger) < -1000000 {"),
    ("N06 comparison ignores held", USAGE, "claimed, ok := sumMinor(bucket.FilledMinor, bucket.HeldMinor)", "claimed, ok := sumMinor(bucket.FilledMinor, bucket.HeldMinor+\"0\")"),
    ("N07 ledger read error admits", USAGE, "\t\tif err != nil {\n\t\t\treturn fmt.Errorf(\"%w: %s bucket %q ledger usage unreadable", "\t\tif err != nil {\n\t\t\treturn nil\n\t\t\treturn fmt.Errorf(\"%w: %s bucket %q ledger usage unreadable"),
    ("N08 stale error not wrapping ErrSnapshotStale", CRA, "ErrRiskBucketUsageStale = fmt.Errorf(\"%w: risk bucket usage snapshot is behind the ledger\", ErrSnapshotStale)", "ErrRiskBucketUsageStale = fmt.Errorf(\"%w: risk bucket usage snapshot is behind the ledger\", ErrSnapshotSuperseded)"),
    ("N09 refusal code wrong", USAGE, "Code: riskbucket.RefusalBucketUsageStale,", "Code: riskbucket.RefusalStaleBucket,"),
    ("N10 shared sum drops filled", RB, "\t\tfilled.Add(filled, rowFilled)\n", "\t\t_ = rowFilled\n"),
    ("N11 shared sum drops held", RB, "\t\theld.Add(held, rowHeld)\n", "\t\t_ = rowHeld\n"),
    ("N12 production reader ignores latch", RB, "\t\tif usage.Latched {", "\t\tif false && usage.Latched {"),
    ("N13 aggregate forgets latched rows", RB, "latched = latched || row.OverageLatched != 0 || row.UnknownLatched != 0", "latched = false"),
    # 1회차 N14 는 컴파일 오류로 "CAUGHT" 가 찍혔음(닿지 않음) — record 표 대신 부모 표에 다시 쓰는 컴파일되는 변이로 바꿈.
    ("N14 policy record not stored", REC, "`INSERT OR IGNORE INTO risk_bucket_policy_records(`+columns", "`INSERT OR IGNORE INTO risk_bucket_policies(`+columns"),
    ("N15 key-only collision reinstated", REC, "\treturn recordDigest, nil\n}", "\tvar parent string\n\tif err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_policies WHERE bucket_dimension=? AND bucket_value=? AND policy_version=?`, string(key.Dimension), key.Value, key.PolicyVersion).Scan(&parent); err != nil || parent != recordDigest {\n\t\treturn \"\", fmt.Errorf(\"%w: immutable policy collision\", ErrRiskBucketSnapshotMismatch)\n\t}\n\treturn recordDigest, nil\n}"),
    ("N16 fill loader joins by key only", FILL, "AND p.record_digest=r.policy_record_digest WHERE r.decision_id=? AND r.policy_record_digest IS NOT NULL", "WHERE r.decision_id=? AND r.policy_record_digest IS NOT NULL"),
    ("N17 v34 required-record trigger removed", SQL34, "WHEN NEW.policy_record_digest IS NULL OR NOT EXISTS (", "WHEN 0 AND (NEW.policy_record_digest IS NULL OR NOT EXISTS ("),
    ("N18 v34 immutable binding trigger removed", SQL34, "WHEN OLD.policy_record_digest IS NOT NEW.policy_record_digest", "WHEN 0"),
    ("N19 schema version not bumped", J + "schema.go", "const SchemaVersion = 34", "const SchemaVersion = 33"),
    ("N20 legacy branch of fill loader dropped", FILL, "WHERE r.decision_id=? AND r.policy_record_digest IS NULL`, decisionID, decisionID)", "WHERE r.decision_id=? AND 0`, decisionID, decisionID)"),
]

TESTS = [
    ["go", "test", "-count=1", "-run", "SharedBucket|StaleUsage|MigrationV33ToV34|TestSchemaTablesAndColumns|TestRiskBucketAdmission|TestFirstLegAtomic|TestRiskBucketActualAndRelease|TestStrategyDispatch|LossLock", "./internal/journal"],
    ["go", "test", "-count=1", "-tags", "tossos_testseams", "./internal/riskbucket"],
    ["go", "test", "-count=1", "-run", "TestA066LedgerUsageHasOneComputation", "./internal/execgw"],
]


def run_tests(copy: Path, env: dict) -> tuple[bool, str]:
    failed, notes = [], []
    for command in TESTS:
        result = subprocess.run(command, cwd=copy, env=env, capture_output=True, text=True)
        if result.returncode != 0:
            out = result.stdout + result.stderr
            names = [line.strip()[len("--- FAIL: "):].split(" ")[0] for line in out.splitlines() if line.strip().startswith("--- FAIL: ")]
            leaves = [n for n in names if not any(o != n and o.startswith(n + "/") for o in names)]
            failed.extend(leaves)
            if not names:
                notes.append(out.strip().splitlines()[-1][:160] if out.strip() else "no output")
    if failed or notes:
        return False, f"{len(failed)} failing: " + ", ".join(failed[:8]) + (" …" if len(failed) > 8 else "") + (" | " + " | ".join(notes) if notes else "")
    return True, ""


def main() -> None:
    args = sys.argv[1:]
    only = None
    if "--only" in args:
        i = args.index("--only")
        only = re.compile(args[i + 1])
        args = args[:i] + args[i + 2:]
    scratch, own = Path(args[0]), args[1:]
    copy = scratch / f"mut561-{os.getpid()}"
    copy.mkdir(parents=True)
    tracked = subprocess.run(["git", "ls-files", "--", "go.mod", "go.sum", "internal", "cmd", "tools"], cwd=ROOT,
                             capture_output=True, text=True, check=True).stdout.split()
    for rel in tracked + own:
        source = ROOT / rel
        if source.exists():
            (copy / rel).parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, copy / rel)
    env = dict(os.environ, GOFLAGS="-trimpath")
    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    dirty = subprocess.run(["git", "diff", "--name-only", "HEAD", "--", "go.mod", "go.sum", "internal", "cmd", "tools"],
                           cwd=ROOT, capture_output=True, text=True, check=True).stdout.split()
    ledger = open(copy / "ledger.tsv", "w", encoding="utf-8")
    ledger.write(f"TREE\tHEAD {head}\tuncommitted tracked: {','.join(dirty) or 'none'}\town untracked: {','.join(own) or 'none'}\n")
    ok, why = run_tests(copy, env)
    ledger.write(f"CONTROL\t{'GREEN' if ok else 'RED'}\t{why}\n")
    ledger.flush()
    if not ok:
        print("control RED — stop:", why)
        sys.exit(2)
    for ident, rel, old, new in MUTANTS:
        if only and not only.search(ident):
            continue
        target = copy / rel
        pristine = target.read_text(encoding="utf-8")
        if pristine.count(old) != 1:
            ledger.write(f"{ident}\tNOT-APPLIED\told occurs {pristine.count(old)} times\n")
            continue
        target.write_text(pristine.replace(old, new, 1), encoding="utf-8")
        green, why = run_tests(copy, env)
        ledger.write(f"{ident}\t{'SURVIVED' if green else 'CAUGHT'}\t{why}\n")
        ledger.flush()
        target.write_text(pristine, encoding="utf-8")
    ledger.close()
    print((copy / "ledger.tsv").read_text(encoding="utf-8"))


if __name__ == "__main__":
    main()
