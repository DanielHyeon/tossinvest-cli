#!/usr/bin/env python3
"""a127 변이 하네스 (design 「반증 설계」 S1~S14 + Manager 필수 축: 버전 확인 생략 · prepare 생략 · tx 분리 · 0 수락).

규율: 저장소 루트 인자 · 대상 · 시험 파일 시작 sha 기록 · 무변이 대조군 GREEN 선행 · 변이마다 치환 대상 정확히 1회(아니면 UNREACHABLE 로 멈춤) ·
판마다 원문 바이트 복원과 sha 재단언 · 시험 파일 sha 판마다 불변 단언 · 한 번에 한 판. RUNS 는 빠른 구조 시험을 먼저 돌려(첫 실패에서 멈춤) tx 분리
변이가 연결 하나(`SetMaxOpenConns(1)`)를 기다리며 멈추는 시간을 줄임 — 멈춤도 실패로 셈(-timeout).

사용: python3 mutate.py <repo-root> [변이 id ...]
"""
import hashlib
import pathlib
import subprocess
import sys

RISK = "internal/riskbucket/production_snapshot_authority.go"
ROUTE = "internal/strategyrouter/production.go"
ERISK = "internal/app/engine/strategy_risk_authority.go"
EROUTE = "internal/app/engine/strategy_route_authority.go"

TESTS = [
    "internal/riskbucket/a127_schema_binding_test.go",
    "internal/riskbucket/production_snapshot_authority_test.go",
    "internal/strategyrouter/a127_schema_binding_test.go",
    "internal/strategyrouter/a127_real_journal_test.go",
    "internal/strategyrouter/a127_manifest_seam_parity_test.go",
    "internal/strategyrouter/production_route_manifest_testseam.go",
    "internal/strategyrouter/production_test.go",
    "internal/app/engine/a127_schema_injection_test.go",
    "internal/app/engine/a112_owner_scope_trading_test.go",
    "internal/app/engine/strategy_risk_authority_test.go",
]

RUNS = [
    (["-tags", "tossos_testseams"], "./internal/riskbucket/", "TestA127RiskLoaderReadsEverythingInOneReadOnlyTransaction", "60s"),
    ([], "./internal/strategyrouter/", "TestA127RouteLoaderPreparesBeforeAnyLedgerDataRead", "60s"),
    ([], "./internal/app/engine/", "TestA127", "120s"),
    (["-tags", "tossos_testseams"], "./internal/riskbucket/", "TestA127|TestProductionRisk|TestA126", "300s"),
    ([], "./internal/strategyrouter/", "TestA127|TestProductionRoute", "300s"),
    (["-tags", "tossos_testseams"], "./internal/strategyrouter/", "TestA127", "300s"),
    (["-tags", "tossos_testseams"], "./internal/app/engine/",
     "TestTheRiskLoaderReadsTheRealJournal|TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope|TestACorruptLedgerRow|TestAScopeLatch", "600s"),
]

NEWER_R = 'fmt.Errorf("risk bucket: journal schema %d is newer than this build\'s %d", version, config.JournalSchemaVersion)'
NEWER_Q = 'fmt.Errorf("%w: journal schema %d is newer than this build\'s %d", ErrProductionRouteUnavailable, version, journalSchema)'

MUTANTS = [
    ("S1", RISK, "\tif version > config.JournalSchemaVersion {", "\tconfig.JournalSchemaVersion = 27\n\tif version > config.JournalSchemaVersion {", "risk 동결 리터럴 27 복귀"),
    ("S2", ROUTE, "\tif version > journalSchema {", "\tjournalSchema = 27\n\tif version > journalSchema {", "route 동결 리터럴 27 복귀"),
    ("S3", RISK, "\tif version > config.JournalSchemaVersion {", "\tif false {", "risk 더 새 원장 검사 생략(버전 확인 생략 축)"),
    ("S4", ROUTE, "\tif version > journalSchema {", "\tif false {", "route 더 새 원장 검사 생략(버전 확인 생략 축)"),
    ("S5a", RISK, "\tif version < config.JournalSchemaVersion {", "\tif false {", "risk 더 옛 원장 수락((a) 로의 후퇴)"),
    ("S5b", ROUTE, "\tif version < journalSchema {", "\tif false {", "route 더 옛 원장 수락"),
    ("S6a", RISK, "\tif config.JournalSchemaVersion <= 0 {", "\tif false {", "risk 주입 0 수락(가드 삭제 — 0 수락 축)"),
    ("S6b", RISK, "\tif config.JournalSchemaVersion <= 0 {", "\tif config.JournalSchemaVersion == 0 {", "risk 음수 주입 수락"),
    ("S6c", ROUTE, "\tif config.JournalSchemaVersion <= 0 {", "\tif false {", "route 주입 0 수락"),
    ("S6d", ROUTE, "\tif config.JournalSchemaVersion <= 0 {", "\tif config.JournalSchemaVersion == 0 {", "route 음수 주입 수락"),
    ("S7", ERISK, "JournalSchemaVersion: journal.SchemaVersion,", "JournalSchemaVersion: journal.SchemaVersion + 0,", "engine risk 자리가 선택자가 아닌 식(런타임 값 대용)"),
    ("S8", ROUTE, "for _, statement := range []string{productionRouteOwnersSQL, productionRouteCampaignSQL} {",
     "for _, statement := range []string{} {", "route 판독 전 prepare 생략(prepare 생략 축)"),
    ("S9a", RISK, NEWER_R, NEWER_R.replace("newer", "older"), "risk 방향 문구 뒤바꿈"),
    ("S9b", ROUTE, NEWER_Q, NEWER_Q.replace("newer", "older"), "route 방향 문구 뒤바꿈"),
    ("S10a", ERISK, "JournalSchemaVersion: journal.SchemaVersion,", "JournalSchemaVersion: 35,", "engine risk 자리 리터럴 35"),
    ("S10b", EROUTE, "JournalSchemaVersion: journal.SchemaVersion}", "JournalSchemaVersion: 35}", "engine route 자리 리터럴 35"),
    ("S11", RISK, NEWER_R, 'fmt.Errorf("%w: journal schema %d is newer than this build\'s %d", ErrProductionRiskScopeRefused, version, config.JournalSchemaVersion)',
     "스키마 거절을 범위 국소 신원으로 재표식"),
    ("S12a", ROUTE, 'fmt.Errorf("%w: owner snapshot: %w", ErrProductionRouteUnavailable, err)', 'fmt.Errorf("%w: owner snapshot", ErrProductionRouteUnavailable)',
     "route Batch 감싸기가 원인을 다시 버림"),
    ("S12b", ROUTE, "return refuse(" + NEWER_Q + ")", "return refuse(ErrProductionRouteUnavailable)", "route opener 가 맨 sentinel 로 복귀"),
    ("S13a", RISK, "ReadJournalBucketUsage(ctx, tx, scope.AccountID", "ReadJournalBucketUsage(ctx, db, scope.AccountID", "risk 사용량 판독을 tx 밖으로(tx 분리 축)"),
    ("S13b", RISK, "if err := tx.QueryRowContext(ctx, productionRiskScopeLatchSQL", "if err := db.QueryRowContext(ctx, productionRiskScopeLatchSQL", "risk latch 판독을 tx 밖으로"),
    ("S13c", RISK, 'if err := tx.QueryRowContext(ctx, "PRAGMA user_version")', 'if err := db.QueryRowContext(ctx, "PRAGMA user_version")', "risk 버전 판독을 tx 밖으로"),
    ("S14", RISK, "for _, statement := range []string{productionRiskScopeLatchSQL, productionRiskUsageSQL} {",
     "for _, statement := range []string{} {", "risk 판독 전 prepare 생략(prepare 생략 축)"),
    # codex 구현 리뷰 P2 #1 · #2 보강(review 1.6.2): 트랜잭션 수명 · prepare 순서.
    ("S13d", RISK, "tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})", "tx, err := db.BeginTx(ctx, nil)", "risk BeginTx 의 ReadOnly 제거(보이스 X3 — 이제 구조 단언이 잡아야)"),
    ("S13e", RISK, "\t// a127 D7: 원장 데이터 질의 전부를 첫 판독 전에 prepare",
     "\ttx.Rollback()\n\ttx, _ = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})\n\t// a127 D7: 원장 데이터 질의 전부를 첫 판독 전에 prepare",
     "risk 버전 확인 뒤 트랜잭션을 닫고 같은 이름으로 다시 엶(수명 분리)"),
    ("S15", RISK, "\t// a127 D7: 원장 데이터 질의 전부를 첫 판독 전에 prepare",
     "\tvar a127Probe int\n\t_ = tx.QueryRowContext(ctx, productionRiskScopeLatchSQL, scope.AccountID, string(scope.Market), scope.Symbol).Scan(&a127Probe)\n\t// a127 D7: 원장 데이터 질의 전부를 첫 판독 전에 prepare",
     "risk 데이터 질의를 prepare 앞에 둠(순서 위반)"),
    ("S16", ROUTE, "\t// a127 D7: owner · campaign 질의를 판독 전에 prepare",
     "\tif rows, err := tx.QueryContext(ctx, productionRouteOwnersSQL, \"\", \"\", \"\"); err == nil {\n\t\trows.Close()\n\t}\n\t// a127 D7: owner · campaign 질의를 판독 전에 prepare",
     "route opener 가 prepare 앞에서 원장 데이터를 읽음(순서 위반)"),
]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def green(root):
    for tags, pkg, run, timeout in RUNS:
        proc = subprocess.run(["go", "test", *tags, pkg, "-run", run, "-count=1", "-timeout", timeout], cwd=root, capture_output=True, text=True)
        if proc.returncode != 0:
            return False
    return True


def main():
    root = pathlib.Path(sys.argv[1]).resolve()
    only = set(sys.argv[2:])
    files = {f: root / f for f in {RISK, ROUTE, ERISK, EROUTE}}
    originals = {f: p.read_bytes() for f, p in files.items()}
    start = {f: sha(p) for f, p in files.items()}
    tests = {f: sha(root / f) for f in TESTS}
    for f in sorted(start):
        print(f"start {start[f][:16]} {f}", flush=True)
    for f in TESTS:
        print(f"test {tests[f][:16]} {f}", flush=True)
    if not green(root):
        print("CONTROL RED — 멈춤", flush=True)
        return 2
    print("control GREEN", flush=True)
    caught = survived = 0
    for mid, f, old, new, why in MUTANTS:
        if only and mid not in only:
            continue
        text = originals[f].decode()
        if text.count(old) != 1:
            print(f"{mid} UNREACHABLE — 치환 대상 {text.count(old)}회 ({why})", flush=True)
            return 3
        files[f].write_text(text.replace(old, new, 1))
        try:
            ok = green(root)
        finally:
            files[f].write_bytes(originals[f])
            if sha(files[f]) != start[f]:
                print(f"{mid} RESTORE FAILED", flush=True)
                return 4
        if any(sha(root / t) != tests[t] for t in TESTS):
            print(f"{mid} TEST FILE CHANGED — 멈춤", flush=True)
            return 5
        if ok:
            survived += 1
            print(f"{mid} SURVIVED — {why}", flush=True)
        else:
            caught += 1
            print(f"{mid} CAUGHT — {why}", flush=True)
    print(f"caught {caught} · survived {survived}", flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
