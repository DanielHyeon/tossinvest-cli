# Function Logic Map: `Recovery.Run`

- Source: `internal/reconcile/recovery.go` (`247`–`345`)
- Qualified: `Recovery.Run`
- AST evidence: `ast.json` (`source_sha256` c8054430ffbeb761…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 13

**역할.** 재시작 복구 순서(재시작 규칙 → 미종결 정산 → 계좌 읽기 → 재구성). a094: ACKED 갈래에서 발주를 기록 번호로 한 번 읽어 바이트 일치면 확정(confirmAcked), 아니면 종전대로 StillPending + 명명 critical.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pending` | 미종결 attempt | `PendingAttempts` | 읽기 실패는 ErrRecoveryIncomplete(종전) |
| `rec.State == ACKED` | 접수 뒤 확정 전 죽은 attempt | 원장 | confirmAcked — 실패는 복구를 실패시키지 않음 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:254` `if err != nil {` | 아니오 |
| B2 | if | `:267` `if err != nil {` | 아니오 |
| B3 | range | `:270` `for _, rec := range pending {` | 예 |
| B4 | if | `:271` `if rec.State != journal.StateInDoubt {` | 예 |
| B5 | if | `:279` `if rec.State == journal.StateAcked && r.confirmAcked(ctx, rec) {` | 예 |
| B6 | if | `:284` `if berr != nil {` | 아니오 |
| B7 | if | `:292` `if rerr != nil {` | 예 |
| B8 | if | `:296` `if settled {` | 예 |
| B9 | if | `:301` `if rerr != nil {` | 아니오 |
| B10 | if | `:306` `if res.State == journal.StateUnresolvedInDoubt {` | 아니오 |
| B11 | if | `:318` `if err != nil {` | 예 |
| B12 | if | `:325` `if err != nil {` | 아니오 |
| B13 | if | `:340` `if report.Diff.BlocksEntry() {` | 예 |

## Calls and live bindings

`RecoverPending` · `PendingAttempts` · **`confirmAcked`**(LookupIntent · `execgw.ConfirmPlacedOrder` · Resume · ResolveConfirmed · RecordCritical 창 0) · `blockedSymbol` · `replay` · `Resolver.Resolve` · `stableSnapshot` · `LocalStateFromJournal` · Comparer · Gate.

## State mutations and fallbacks

attempt ACKED→CONFIRMED(바이트 일치일 때만) · 그 밖 종전.

## Safety conclusion

- ACKED 는 목록 대조 해소로 보내지 않는다(matcher 번호 판별자 부재). 새 오류 반환 경로 0 — 관측 루프가 뜨지 않는 경로를 만들지 않음(D−4.2-2). 브로커 호출은 ACKED PLACE 행마다 읽기 1(§0.4 D−5.5). High-risk: yes.
