#!/usr/bin/env python3
"""a126 변이 하네스 (design 「반증 설계」 M1~M8 + a066 결함 수리 + 보강).

규율: 저장소 루트를 인자로 받음 · 대상 파일 시작 sha 기록 · 무변이 대조군 GREEN 선행 · 변이마다 치환 대상 정확히 1회(0 이면
UNREACHABLE 로 멈춤) · 판마다 원문 바이트로 복원하고 sha 재단언 · 한 번에 한 판 · 시험 파일 sha 도 시작에 기록하고 판마다 재단언
(1.5 R14 — 시험이 도중에 바뀌면 판정이 다른 시험의 것이 됨).

사용: python3 mutate.py <repo-root> [변이 id ...]
"""
import hashlib
import pathlib
import subprocess
import sys

AGG = "internal/riskbucket/production_snapshot_authority.go"
OWNER = "internal/journal/risk_bucket_owner.go"
FILL = "internal/journal/risk_bucket_fill.go"
USAGE = "internal/journal/risk_bucket_usage.go"
STRAT = "internal/journal/strategy_dispatch_runtime.go"
HOOK = "internal/journal/apply_hook.go"
ISSUE = "internal/journal/risk_bucket_issuance.go"

TESTS = [
    "internal/journal/a126_filled_exposure_leaves_the_bucket_test.go",
    "internal/journal/risk_bucket_fill_test.go",
    "internal/journal/risk_bucket_owner_test.go",
    "internal/riskbucket/a126_departed_rows_test.go",
    "internal/riskbucket/a126_snapshot_digest_replay_test.go",
    "internal/riskbucket/production_snapshot_authority_test.go",
]

RUNS = [
    ("./internal/riskbucket/", "TestA126|TestProductionRisk", ["-tags", "tossos_testseams"]),
    ("./internal/journal/", "TestA126|TestRiskBucketOwner|TestReleasedOwner|TestA066|TestRiskBucketUnsafe", []),
]

MUTANTS = [
    ("M1", AGG, "departed := row.Receipted != 0 && row.ScopeLatched == 0", "departed := true", "떠남을 항상 참(활성 owner 도 빠짐)"),
    ("M1b", AGG, "		if !departed {\n			filled.Add(filled, rowFilled)\n			held.Add(held, rowHeld)\n		}",
     "		filled.Add(filled, rowFilled)\n		held.Add(held, rowHeld)", "떠남을 아예 적용하지 않음(수리 되돌림 — 양성 대조가 잡아야)"),
    ("M2", AGG, "departed := row.Receipted != 0 && row.ScopeLatched == 0",
     "departed := row.OwnerReleasedAt != \"\" && row.ScopeLatched == 0", "영수증 대신 owner released_at 만 봄"),
    ("M3", AGG, "departed := row.Receipted != 0 && row.ScopeLatched == 0", "departed := row.Receipted != 0",
     "scope latch 되돌림 제거(D4 fail-open)"),
    ("M4", USAGE, "AND (r.state IN ('HELD','FILLED') OR r.filled_minor<>'0')",
     "AND (r.state IN ('HELD','FILLED') OR r.filled_minor<>'0') AND NOT EXISTS(SELECT 1 FROM risk_bucket_owner_release_receipts x WHERE x.account_ref=r.account_ref AND x.market=r.market AND x.symbol=r.symbol AND x.prospective_generation=r.owner_prospective_generation)",
     "떠난 행을 한도 모집단에서 뺌(D3 완화)"),
    ("M6", AGG, "row.OwnerReleasedAt != row.ReceiptReleasedAt || row.State == \"HELD\" || rowHeld.Sign() != 0)",
     "row.OwnerReleasedAt != row.ReceiptReleasedAt)", "영수증 행의 HELD · held≠0 을 받음"),
    ("M6b-1", AGG, "row.OwnerReleasedAt == \"\" || row.OwnerReleasedAt != row.ReceiptReleasedAt || ",
     "", "released_at 불일치를 받음"),
    ("M6b-2", AGG, "if (row.Receipted != 0 || row.DecisionReceipted != 0) && row.OwnerKeyMatches == 0 {",
     "if false && row.OwnerKeyMatches == 0 {", "owner 키 사본 불일치를 받음"),
    ("M6b-3", AGG, "if (row.Receipted != 0 || row.DecisionReceipted != 0) && row.OwnerKeyMatches == 0 {",
     "if row.Receipted != 0 && row.OwnerKeyMatches == 0 {", "결정 사본 쪽 영수증을 보지 않음"),
    ("M6c", AGG, "		if row.Receipted != 0 && (row.OwnerReleasedAt == \"\"",
     "		if row.Receipted != 0 && row.ScopeLatched == 0 && (row.OwnerReleasedAt == \"\"",
     "손상 검사를 scope latch 판정 뒤로(되돌림이 손상 검사 우회 — codex #4)"),
    ("M7", AGG, "		latched = latched || row.OverageLatched != 0 || row.UnknownLatched != 0\n		overage = overage || row.OverageLatched != 0\n		unknown = unknown || row.UnknownLatched != 0\n		if !departed {",
     "		if !departed {\n			latched = latched || row.OverageLatched != 0 || row.UnknownLatched != 0\n			overage = overage || row.OverageLatched != 0\n			unknown = unknown || row.UnknownLatched != 0\n		}\n		if !departed {",
     "떠난 행의 latch 플래그를 집계에서 뺌"),
    ("M9", AGG, "CASE WHEN EXISTS(SELECT 1 FROM risk_bucket_scope_latches l WHERE l.account_ref=r.account_ref AND l.market=r.market AND l.symbol=r.symbol\n\t\t\tAND l.prospective_generation=r.owner_prospective_generation) THEN 1 ELSE 0 END",
     "0", "SQL 이 scope latch 를 싣지 않음(되돌림 입력 제거)"),
    ("M10", AGG, "CASE WHEN rc.released_at IS NULL THEN 0 ELSE 1 END", "CASE WHEN ow.released_at IS NULL THEN 0 ELSE 1 END",
     "SQL 의 영수증 존재를 owner released_at 으로 바꿔치기"),
    ("F1", OWNER, "AND NOT ` + riskBucketFillActualResolvedSQL",
     "AND (f.actual_known=0 OR NOT EXISTS(SELECT 1 FROM risk_bucket_fill_actual_evidence a WHERE a.fill_id=f.fill_id))`",
     "a066 결함 수리 되돌림(해제 검사의 옛 철자)"),
    ("F2", OWNER, "AND NOT ` + riskBucketFillActualResolvedSQL", "AND 0`", "해제 검사가 미해소 체결을 보지 않음(fail-open)"),
    ("F3", FILL, "const riskBucketFillActualResolvedSQL = `(f.actual_known=1 OR EXISTS(",
     "const riskBucketFillActualResolvedSQL = `(f.actual_known=1 AND EXISTS(", "해소 조각 OR→AND(두 자리 공통)"),
    ("P1", STRAT, "	handle.invalidate()\n	return j.applyRiskBucketOwnerBindingInTx(ctx, tx, fill)\n}", "	handle.invalidate()\n	return nil\n}",
     "전략 정산 경로(campaign 미결선)의 binding 제거 — 그 경로의 되돌림 시험이 잡아야"),
    ("P2", HOOK, "		if err := j.applyRiskBucketOwnerBindingInTx(ctx, tx, fill); err != nil {", "		if err := error(nil); err != nil {",
     "RecordFill 경로(campaign 결선)의 binding 제거 — 그 경로의 되돌림 시험이 잡아야"),
    # 1.5 보강(증거 보이스 생존 변이 · 적대 보이스 tripwire · codex #2 재작성 · R8).
    ("X1", AGG, " AND rc.prospective_generation=r.owner_prospective_generation", "", "영수증 조인이 generation 을 안 봄(R6)"),
    ("X2", AGG, " AND rc.symbol=r.symbol", "", "영수증 조인이 symbol 을 안 봄(R6)"),
    ("X3", AGG, "row.State == \"HELD\" || ", "", "HELD 상태 검사 제거, held≠0 은 남김(R4)"),
    ("X7", AGG, " AND d.symbol=r.symbol", "", "사본 일치가 symbol 을 안 봄(R5)"),
    ("X7m", AGG, " AND d.market=r.market", "", "사본 일치가 market 을 안 봄(R5)"),
    ("X9", AGG, "ON rc.account_ref=r.account_ref AND ", "ON ", "영수증 조인이 account 를 안 봄(R6)"),
    ("B3", USAGE, "return fmt.Errorf(\"%w: %s bucket %q ledger usage unreadable: %v\", ErrRiskBucketSnapshotMismatch, bucket.Key.Dimension, bucket.Key.Value, err)",
     "_ = err", "admission 이 판독 불가 사용량을 거절하지 않음(R8)"),
    ("C2close", ISSUE, "		if err := latchedUsageRefusal(ref.dimension, ref.value, usage); err != nil {",
     "		if usage.FilledMinor == \"50\" && usage.HeldMinor == \"60\" {\n			return true, fmt.Errorf(\"a126 C2 closed by a recompute\")\n		}\n		if err := latchedUsageRefusal(ref.dimension, ref.value, usage); err != nil {",
     "제출 재검증이 latch 아닌 재계산으로 C2 를 닫음 — 잔여 핀이 깨져야(R2)"),
    ("T1", OWNER, "func (j *Journal) releaseRiskBucketOwner(ctx context.Context, key riskbucket.OwnerKey) (RiskBucketOwnerReleaseResult, error) {",
     "func a126Tripwire(j *Journal) { _, _ = j.releaseRiskBucketOwner(context.Background(), riskbucket.OwnerKey{}) }\n\nfunc (j *Journal) releaseRiskBucketOwner(ctx context.Context, key riskbucket.OwnerKey) (RiskBucketOwnerReleaseResult, error) {",
     "owner 해제 생산 호출 추가 — tasks 3.1 tripwire 가 잡아야(R7)"),
]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def green(root):
    for pkg, run, tags in RUNS:
        proc = subprocess.run(["go", "test", *tags, pkg, "-run", run, "-count=1", "-timeout", "1200s"], cwd=root, capture_output=True, text=True)
        if proc.returncode != 0:
            return False
    return True


def main():
    root = pathlib.Path(sys.argv[1]).resolve()
    only = set(sys.argv[2:])
    files = {f: root / f for f in {AGG, OWNER, FILL, USAGE, STRAT, HOOK, ISSUE}}
    tests = {f: sha(root / f) for f in TESTS}
    for f in TESTS:
        print(f"test {tests[f][:16]} {f}")
    originals = {f: p.read_bytes() for f, p in files.items()}
    start = {f: sha(p) for f, p in files.items()}
    for f in sorted(start):
        print(f"start {start[f][:16]} {f}")
    if not green(root):
        print("CONTROL RED — 멈춤")
        return 2
    print("control GREEN")
    caught = survived = 0
    for mid, f, old, new, why in MUTANTS:
        if only and mid not in only:
            continue
        text = originals[f].decode()
        if text.count(old) != 1:
            print(f"{mid} UNREACHABLE — 치환 대상 {text.count(old)}회 ({why})")
            return 3
        files[f].write_text(text.replace(old, new, 1))
        try:
            ok = green(root)
        finally:
            files[f].write_bytes(originals[f])
            if sha(files[f]) != start[f]:
                print(f"{mid} RESTORE FAILED")
                return 4
            if any(sha(root / t) != tests[t] for t in TESTS):
                print(f"{mid} TEST FILE CHANGED — 멈춤")
                return 5
        if ok:
            survived += 1
            print(f"{mid} SURVIVED — {why}")
        else:
            caught += 1
            print(f"{mid} CAUGHT — {why}")
    print(f"caught {caught} · survived {survived}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
