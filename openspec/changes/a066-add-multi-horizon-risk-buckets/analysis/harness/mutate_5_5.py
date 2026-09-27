#!/usr/bin/env python3
"""a066 5.5 변이 하네스 — 진입 손실 잠금의 세 호출 자리·판정 규칙·원장 규율을 사본에서 변이함.

사용: python3 mutate_5_5.py <scratch-dir> [mutant-id-regex]

- 사본: <scratch-dir>/mut-<pid>/ 에 go.mod·go.sum·internal/·cmd/ 만 복사(.env 등은 복사하지 않음).
- 무변이 대조군이 먼저 GREEN 이어야 함. 아니면 멈춤(형제 파일이 빠진 사본은 모든 변이를 CAUGHT 로 찍음).
- 변이마다 원본 파일을 되돌린 뒤 정확히 한 곳(count==1)을 바꾸고, 대상 시험 두 묶음을 돌려
  하나라도 실패하면 CAUGHT. 변이가 사본에 실제로 앉았는지 바꾼 문자열로 다시 확인함.
- 한 번에 한 판(병렬 금지), 결과는 <scratch-dir>/mut-<pid>/ledger.tsv.
"""
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
J = "internal/journal/"
LOCK = J + "risk_bucket_entry_loss_lock.go"
SQL = J + "risk_bucket_entry_loss_lock_v33.sql"
CRA = J + "risk_bucket.go"
ISS = J + "risk_bucket_issuance.go"
CRA_CALL = "refuseEntryUnderLossLock(ctx, tx, plan.Owner.Key.AccountID, plan.Owner.Key.Market, admissionHorizon(decision)); err != nil {\n\t\treturn RiskBucketAdmissionReceipt{}, err\n\t}\n\n\townerReused := false\n\tvar prospective, lane, campaign string\n\terr = "
FRESH_CALL = "refuseEntryUnderLossLock(ctx, tx, plan.Owner.Key.AccountID, plan.Owner.Key.Market, admissionHorizon(decision)); err != nil {\n\t\treturn RiskBucketAdmissionReceipt{}, err\n\t}\n\n\townerReused := false\n\tvar prospective, lane, campaign string\n\terr := "
GW = "internal/execgw/gateway.go"
GW_MAP = "if errors.Is(err, journal.ErrRiskBucketEntryLossLocked) {\n\t\t\t\t\treturn reject(ReasonEntryLossLockActive,"
REVAL_CALL = "refuseEntryUnderLossLock(ctx, j.db, account, riskbucket.Market(market), horizon); err != nil {"

# (id, file, old, new) — old 는 파일 안에서 정확히 한 번 나와야 함.
MUTANTS = [
    ("M01 CRA call removed", CRA, CRA_CALL, CRA_CALL.replace("refuseEntryUnderLossLock(", "func(...any) error { return nil }(", 1)),
    ("M02 CRA account<-LaneID", CRA, CRA_CALL, CRA_CALL.replace("plan.Owner.Key.AccountID,", "plan.Owner.LaneID,", 1)),
    ("M03 CRA market<-KR", CRA, CRA_CALL, CRA_CALL.replace("plan.Owner.Key.Market,", "riskbucket.MarketKR,", 1)),
    ("M04 CRA horizon<-SHORT", CRA, CRA_CALL, CRA_CALL.replace("admissionHorizon(decision))", "riskbucket.HorizonShort)", 1)),
    ("M05 fresh call removed", ISS, FRESH_CALL, FRESH_CALL.replace("refuseEntryUnderLossLock(", "func(...any) error { return nil }(", 1)),
    ("M06 fresh account<-LaneID", ISS, FRESH_CALL, FRESH_CALL.replace("plan.Owner.Key.AccountID,", "plan.Owner.LaneID,", 1)),
    ("M07 fresh market<-KR", ISS, FRESH_CALL, FRESH_CALL.replace("plan.Owner.Key.Market,", "riskbucket.MarketKR,", 1)),
    ("M08 fresh horizon<-SHORT", ISS, FRESH_CALL, FRESH_CALL.replace("admissionHorizon(decision))", "riskbucket.HorizonShort)", 1)),
    ("M09 reval call removed", ISS, REVAL_CALL, "func(...any) error { return nil }(ctx, j.db, account, riskbucket.Market(market), horizon); err != nil {"),
    ("M10 reval account<-lane", ISS, REVAL_CALL, REVAL_CALL.replace("j.db, account,", "j.db, lane,", 1)),
    ("M11 reval market<-KR", ISS, REVAL_CALL, REVAL_CALL.replace("riskbucket.Market(market)", "riskbucket.MarketKR", 1)),
    # 컴파일되는 변이여야 함 — 1차 판(run1)은 `horizon` 미사용 컴파일 오류로 "CAUGHT" 가 찍혔음(닿지 않음).
    ("M12 reval horizon<-SHORT", ISS, REVAL_CALL, REVAL_CALL.replace(", horizon);", ", riskbucket.Horizon(strings.Repeat(string(horizon), 0)+\"SHORT\"));", 1)),
    ("M13 decisionHorizon reads market dimension", LOCK, "string(riskbucket.DimensionHorizon)).Scan(&value)", "string(riskbucket.DimensionMarket)).Scan(&value)"),
    ("M14 admissionHorizon reads market cap", LOCK, "if cap.Key.Dimension == riskbucket.DimensionHorizon {", "if cap.Key.Dimension == riskbucket.DimensionMarket {"),
    ("M15 rule ignores found lock", LOCK, "\tif found {\n\t\t// 타입 있는 거절", "\tif false && found {\n\t\t// 타입 있는 거절"),
    ("M16 lock query ignores horizon", LOCK, "WHERE account_ref=? AND market=? AND horizon=?\n\t\tORDER BY lock_seq LIMIT 1`, account, string(market), string(horizon))", "WHERE account_ref=? AND market=?\n\t\tORDER BY lock_seq LIMIT 1`, account, string(market))"),
    ("M17 lock query ignores market", LOCK, "WHERE account_ref=? AND market=? AND horizon=?\n\t\tORDER BY lock_seq LIMIT 1`, account, string(market), string(horizon))", "WHERE account_ref=? AND horizon=?\n\t\tORDER BY lock_seq LIMIT 1`, account, string(horizon))"),
    ("M18 lock query ignores account", LOCK, "WHERE account_ref=? AND market=? AND horizon=?\n\t\tORDER BY lock_seq LIMIT 1`, account, string(market), string(horizon))", "WHERE market=? AND horizon=?\n\t\tORDER BY lock_seq LIMIT 1`, string(market), string(horizon))"),
    ("M19 rule read error admits", LOCK, "\tif err != nil {\n\t\treturn fmt.Errorf(\"%w: %v\", ErrRiskBucketEntryBlocked, err)\n\t}", "\tif err != nil {\n\t\treturn nil\n\t}"),
    ("M20 rule unknown scope admits", LOCK, "\tif !validEntryLossLockScope(account, market, horizon) {\n\t\treturn fmt.Errorf(\"%w: entry loss lock scope", "\tif false && !validEntryLossLockScope(account, market, horizon) {\n\t\treturn fmt.Errorf(\"%w: entry loss lock scope"),
    ("M21 locked error not wrapping blocked", LOCK, "fmt.Errorf(\"%w: entry loss lock active\", ErrRiskBucketEntryBlocked)", "errors.New(\"journal: entry loss lock active\")"),
    ("M22 API skips existing check", LOCK, "\tif found {\n\t\treturn existing, false, nil\n\t}", "\tif false && found {\n\t\treturn existing, false, nil\n\t}"),
    ("M23 API accepts empty cause", LOCK, "strings.TrimSpace(lock.Cause) == \"\" || lock.ActivatedAt.IsZero()", "lock.ActivatedAt.IsZero()"),
    ("M24 API accepts zero time", LOCK, "strings.TrimSpace(lock.Cause) == \"\" || lock.ActivatedAt.IsZero()", "strings.TrimSpace(lock.Cause) == \"\""),
    ("M25 no first-cause trigger", SQL, "CREATE TRIGGER risk_bucket_entry_loss_lock_first_cause_wins BEFORE INSERT ON risk_bucket_entry_loss_locks\nWHEN EXISTS (", "CREATE TRIGGER risk_bucket_entry_loss_lock_first_cause_wins BEFORE INSERT ON risk_bucket_entry_loss_locks\nWHEN 0 AND EXISTS ("),
    ("M26 first-cause trigger ignores horizon", SQL, "   AND existing.horizon=NEW.horizon)", ")"),
    ("M27 no update trigger", SQL, "BEGIN SELECT RAISE(ABORT,'risk bucket entry loss locks are immutable'); END;", "BEGIN SELECT 1; END;"),
    ("M28 no delete trigger", SQL, "BEGIN SELECT RAISE(ABORT,'risk bucket entry loss locks cannot be deleted'); END;", "BEGIN SELECT 1; END;"),
    ("M29 schema version not bumped", J + "schema.go", "const SchemaVersion = 33", "const SchemaVersion = 32"),
    # reason code 양층(Manager 승인 2026-09-27).
    ("M30 gateway typed mapping removed", GW, GW_MAP, GW_MAP.replace("errors.Is(err, journal.ErrRiskBucketEntryLossLocked)", "false && errors.Is(err, journal.ErrRiskBucketEntryLossLocked)", 1)),
    ("M31 gateway maps lock to mismatch", GW, GW_MAP, GW_MAP.replace("return reject(ReasonEntryLossLockActive,", "return reject(ReasonGuardianRiskBucketMismatch,", 1)),
    ("M32 refusal code wrong", LOCK, "Code: riskbucket.RefusalEntryLossLockActive,", "Code: riskbucket.RefusalBucketCapExhausted,"),
    ("M33 reason code not registered", "internal/execgw/failclosed.go", "\t\tReasonEntryLossLockActive,\n", ""),
    # legacy 잔여 census 양성 대조군 — 표본이 채워지면 빨개져야 함.
    ("M34 production NewTracer call", "internal/app/engine/guardian_wiring.go", "func defaultProductionGuardianFactory(", "func a066CensusProbeTracer() { _, _ = NewTracer(TracerOptions{}) }\n\nfunc defaultProductionGuardianFactory("),
    ("M35 production GuardianAdapter", "internal/strategydispatch/adapters.go", "func (a *GuardianAdapter) IssueAndPlan(", "var a066CensusProbeAdapter = &GuardianAdapter{}\n\nfunc (a *GuardianAdapter) IssueAndPlan("),
    ("M36 new exposure-raising producer", "internal/execgw/issue.go", "func (i *Issuer) record(", "func a066CensusProbeProducer() journal.DecisionRequest {\n\treturn journal.DecisionRequest{SafetyClass: journal.SafetyClassExposureRaising}\n}\n\nfunc (i *Issuer) record("),
]

# 1차 판(run1)의 journal 정규식은 `EntryLossLock` 이라 TestRefuseEntryUnderLossLock… 를 못 골랐고
# M19·M20 이 SURVIVED 로 찍혔음(닿지 않음). `LossLock` 으로 넓힘.
TESTS = [
    ["go", "test", "-count=1", "-run", "LossLock|MigrationV32ToV33|TestSchemaTablesAndColumns", "./internal/journal"],
    ["go", "test", "-count=1", "-run", "TestA066|TestReasonCodeEnumIsStable", "./internal/execgw"],
]


def run_tests(copy: Path, env: dict) -> tuple[bool, str]:
    # 두 묶음을 모두 돌리고 실패한 (하위)시험 이름을 적음 — 자리별 변이가 **자기 자리의 행**에서
    # 빨개지는지 보려면 첫 실패가 아니라 실패 목록 전체가 필요함.
    failed, notes = [], []
    for command in TESTS:
        result = subprocess.run(command, cwd=copy, env=env, capture_output=True, text=True)
        if result.returncode != 0:
            out = result.stdout + result.stderr
            names = [line.strip()[len("--- FAIL: "):].split(" ")[0] for line in out.splitlines() if line.strip().startswith("--- FAIL: ")]
            leaves = [n for n in names if not any(other != n and other.startswith(n + "/") for other in names)]
            failed.extend(leaves)
            if not names:
                notes.append(out.strip().splitlines()[-1][:160] if out.strip() else "no output")
    if failed or notes:
        return False, f"{len(failed)} failing: " + ", ".join(failed) + (" | " + " | ".join(notes) if notes else "")
    return True, ""


def main() -> None:
    scratch = Path(sys.argv[1])
    copy = scratch / f"mut-{os.getpid()}"
    copy.mkdir(parents=True)
    for name in ("go.mod", "go.sum"):
        shutil.copy2(ROOT / name, copy / name)
    # cmd/ 도 복사함 — legacy census 시험이 internal/ 과 cmd/ 의 비시험 소스를 전수로 읽음.
    for tree in ("internal", "cmd"):
        shutil.copytree(ROOT / tree, copy / tree, symlinks=True)
    env = dict(os.environ, GOFLAGS="-trimpath")
    ledger = open(copy / "ledger.tsv", "w", encoding="utf-8")
    ok, why = run_tests(copy, env)
    ledger.write(f"CONTROL\t{'GREEN' if ok else 'RED'}\t{why}\n")
    ledger.flush()
    if not ok:
        print("control RED — stop:", why)
        sys.exit(2)
    only = re.compile(sys.argv[2]) if len(sys.argv) > 2 else None
    for ident, rel, old, new in MUTANTS:
        if only and not only.search(ident):
            continue
        target = copy / rel
        pristine = (ROOT / rel).read_text(encoding="utf-8")
        if pristine.count(old) != 1:
            ledger.write(f"{ident}\tNOT-APPLIED\told occurs {pristine.count(old)} times\n")
            continue
        target.write_text(pristine.replace(old, new, 1), encoding="utf-8")
        if new not in target.read_text(encoding="utf-8"):
            ledger.write(f"{ident}\tNOT-APPLIED\tmutant text absent after write\n")
            continue
        green, why = run_tests(copy, env)
        ledger.write(f"{ident}\t{'SURVIVED' if green else 'CAUGHT'}\t{why}\n")
        ledger.flush()
        target.write_text(pristine, encoding="utf-8")
    ledger.close()
    print((copy / "ledger.tsv").read_text(encoding="utf-8"))


if __name__ == "__main__":
    main()
