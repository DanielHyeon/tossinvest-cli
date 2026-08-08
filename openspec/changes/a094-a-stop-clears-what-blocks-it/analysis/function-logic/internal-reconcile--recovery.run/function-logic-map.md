# Function Logic Map: `Recovery.Run`

- Source: `internal/reconcile/recovery.go` (`207`–`296`)
- Qualified: `Recovery.Run`
- AST evidence: `ast.json` (`source_sha256` 80ee029c47a7355d…)
- Risk scan: `risk-pattern-report.md`
- 분기 12 · return 8 · 호출 24

**역할.** 재시작 시 원장을 정리하고, 미종결 attempt를 해소하고, 계정을 읽어 상태를 재구성한다. **순서가 인과적이다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `PendingAttempts` | 미종결 attempt | 원장 | B3이 순회한다 |
| `rec.State` | attempt 상태 | 원장 | **B4가 IN_DOUBT만 해소로 보낸다** |
| `Replayer` | 재생 진입점 | `Context.Recovery` → `c.Gateway` | **attestation이 꺼져 있어 사양한다**(`replay.go:257`) |
| `Resolver` | 관측 해소기 | `Context.Recovery` → `c.Resolver` | `:253`. **mutator 없음** |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:214` `if err != nil {` | `fmt.Errorf`, `r.opts.Journal.PendingAttempts` | :215 | 아니오 |
| B2 | if | `:227` `if err != nil {` | `fmt.Errorf` | :228 | 아니오 |
| B3 | range | `:230` `for _, rec := range pending {` | — | — | 예 |
| B4 | if | `:231` `if rec.State != journal.StateInDoubt {` | `r.blockedSymbol` | — | 아니오 |
| B5 | if | `:237` `if berr != nil {` | `append`, `fmt.Errorf`, `r.replay` | :238 | 아니오 |
| B6 | if | `:245` `if rerr != nil {` | `fmt.Errorf` | :246 | 예 |
| B7 | if | `:249` `if settled {` | `r.opts.Resolver.Resolve` | — | 예 |
| B8 | if | `:254` `if rerr != nil {` | `append`, `fmt.Errorf` | :255 | 아니오 |
| B9 | if | `:259` `if res.State == journal.StateUnresolvedInDoubt {` | `append`, `r.stableSnapshot` | — | 아니오 |
| B10 | if | `:269` `if err != nil {` | `LocalStateFromJournal` | :270 | 예 |
| B11 | if | `:276` `if err != nil {` | `clk.Now`, `fmt.Errorf`, `r.opts.Comparer.Compare`, `r.opts.Gate.Clear` | :277 | 아니오 |
| B12 | if | `:291` `if report.Diff.BlocksEntry() {` | `r.opts.Gate.Block`, `report.Diff.BlocksEntry`, `report.Diff.Summary` | :295 | 예 |

## Calls and live bindings

`Journal.RecoverPending`(B1 앞) · `Journal.PendingAttempts`(B2 앞) · `blockedSymbol`(B4 안) · `r.replay`(B6 앞) · **`r.opts.Resolver.Resolve`(`:253`)** · `stableSnapshot`(B10 앞) · `LocalStateFromJournal`(B11 앞) · `Comparer.Compare` · `Gate.Clear`/`Gate.Block`(B12).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

원장의 attempt 상태 전이(해소 결과) · 진입 게이트 latch. **주문은 내지 않는다.**

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수의 본문을 바꾸지 않는다.** R3은 이 순회를 **다른 시점에도** 돌게 하는 배선을 더한다. 재시작 순회 자체는 그대로다(spec delta의 '재시작 복구는 계속 미종결 attempt를 순회한다(SHALL — 변경 없음)').
- **High-risk impact**: yes — 시작 시 상태 재구성 전체를 소유한다.
