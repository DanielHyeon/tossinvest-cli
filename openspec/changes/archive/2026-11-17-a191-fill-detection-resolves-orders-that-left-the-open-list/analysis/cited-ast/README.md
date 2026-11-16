# 인용 함수 AST (a191 1단계)

이 change 가 **고치지 않지만** proposal·design 에서 분기·early return·side effect 를 근거로 쓴 함수들의 `tools/logic-map` 산출물이다.
편집 대상(`Detector.collect`)의 번들은 `../function-logic/internal-filldetect--detector.collect/` 에 있다.

- 기준: worktree `~/tossos-work/a191-filldetect-closed-list`, HEAD `a392308e8`(base 와 같음), 2026-11-16.
- 명령: `GOFLAGS=-trimpath go run ./tools/logic-map --file <파일> --func <함수>` → `<패키지-경로>--<함수>.ast.json`. 19개 전부 rc 0.
- 요약: `summary.txt`(생성기 `astsum.py.txt` — 분기 `B<n>:kind@line`, return 줄, call `text@line`).
- 좌표 표기: design·proposal 의 `B<n>(kind@line)` 은 이 파일들의 `branches[].id`·`at.line` 이고, `call@line` 은 `calls[].at.line` 이다.

## AST 가 말하지 못해서 손으로 읽은 것 (design 에 「손으로 읽음」 표시)

| 무엇 | 자리 | 왜 AST 밖인가 |
|---|---|---|
| 추적 SQL 의 `f.terminal = 0`·셋째 갈래 `NOT EXISTS` | fills.go:1553~1654 | SQL 문자열 리터럴 |
| legacy 미러의 범위 일치 `WHERE` | fills.go:778~805 | SQL 문자열 |
| missing-order 블록의 `Cause`·`Release` 값 | mismatch.go:1419~1430 | 복합 리터럴의 필드 값 |
| 영구 블록의 `Cause` 값 | mismatch.go:729~737 | 복합 리터럴 |
| `Diff.BlocksEntry` 의 식 | compare.go:313 | 분기 0, 식 본문 |
| 오버셀 → QUANTITY_MISMATCH | position_projection.go:428~433 | AST 안 만듦(근거로 판정하지 않음 — 설계가 그 경로를 막음) |
| `ReconcileDriver.logCycle` | reconcileloop.go:535~556 | 범위 밖 후속, AST 안 만듦 |
| 공식 어댑터의 필터 전달 | orders_source.go:47~63, orders_raw.go:62~90 | 분기 없는 필드 복사(AST 안 만듦) |
| exit 루프의 `Critical` 우선순위 | exitloop.go:872·1393·1554 | grep 으로 자리만 확인 |

## 목록

| 파일 | 함수 | 근거로 쓴 곳 |
|---|---|---|
| `internal-brokerstate--derive.ast.json` | `brokerstate.Derive` | design §1-3, D3, N4 |
| `internal-journal--journal.recordfill.ast.json` | `Journal.RecordFill` | design §1-4, Q4, Q5 부수 효과 |
| `internal-journal--classifyfillrefusal.ast.json` | `classifyFillRefusal` | refusal 경로(새 refusal 을 안 만든다는 근거) |
| `internal-journal--journal.trackedfillorders.ast.json` | `Journal.TrackedFillOrders` | Q5(a) — Go 부분만, SQL 은 손으로 읽음 |
| `internal-reconcile--localstatefromjournal.ast.json` | `reconcile.LocalStateFromJournal` | proposal 표, Q5(b) |
| `internal-reconcile--comparer.compare.ast.json` | `reconcile.Comparer.Compare` | proposal 표, Q3 통과 (a) |
| `internal-reconcile--diff.blocksentry.ast.json` | `reconcile.Diff.BlocksEntry` | Q5(d) |
| `internal-reconcile--tracker.observe.ast.json` | `reconcile.Tracker.Observe` | Q5(c), Q5(d) |
| `internal-reconcile--blocksfor.ast.json` | `reconcile.blocksFor` | Q5(c) |
| `internal-execgw--scanorders.ast.json` | `execgw.ScanOrders` | D2, Q1 |
| `internal-execgw--resolver.scanboth.ast.json` | `execgw.Resolver.scanBoth` | §2 선례 |
| `internal-filldetect--detector.polllocked.ast.json` | `filldetect.Detector.pollLocked` | §1-4, §2, D7 |
| `internal-filldetect--detector.run.ast.json` | `filldetect.Detector.Run` | §1-5 |
| `internal-filldetect--detector.blocksymbol.ast.json` | `filldetect.Detector.blockSymbol` | §2(새 refusal 이 만들 래치) |
| `internal-filldetect--journalledger.apply.ast.json` | `filldetect.JournalLedger.Apply` | Q5 부수 효과(`Refresh`) |
| `internal-journal--applyexitfill.ast.json` | `journal.ApplyExitFill` | proposal Why, D11, Q5 |
| `cmd-tossctl--runenginereconcileresolve.ast.json` | `runEngineReconcileResolve` | proposal 표, Q5(d) |
| `cmd-tossctl--buildreconcileresolveruntime.ast.json` | `buildReconcileResolveRuntime`(validate 클로저 포함) | Q5(d) |
| `cmd-tossctl--enginefilldetector.ast.json` | `engineFillDetector` | D8, M1-A 편집 후보(편집 전 AST) |
