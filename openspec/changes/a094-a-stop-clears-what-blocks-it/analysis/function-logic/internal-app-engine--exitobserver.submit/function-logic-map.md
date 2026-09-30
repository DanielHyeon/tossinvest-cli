# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (`1365`–`1449`)
- Qualified: `ExitObserver.submit`
- AST evidence: `ast.json` (`source_sha256` 9b7d9a800069b100…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 13

**역할.** 무장된 발의를 브로커로 보낸다. a094: 해제는 intent 를 넘겨 원장 판정이 하고(4.3c), attempt 가 기록됐는데 비수용 종결이 아니면 무장 유지.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `intentID` | 무장된 발의의 intent | record | 해제 판정의 기대 intent |
| `out` | 게이트웨이 결과 | `Submit.Place` | 갈래 B7~B12 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1368` `if err != nil {` | 아니오 |
| B2 | if | `:1371` `if isZeroQuantity(submitQuantity) {` | 예 |
| B3 | if | `:1391` `if err != nil {` | 아니오 |
| B4 | if | `:1400` `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | 예 |
| B5 | if | `:1405` `if err != nil {` | 아니오 |
| B6 | switch | `:1415` `switch {` | — |
| B7 | case | `:1416` `case out.State == journal.StateConfirmed:` | 예 |
| B8 | case | `:1424` `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | 예 |
| B9 | case | `:1429` `case out.Reason == execgw.ReasonSymbolInFlight:` | 아니오 |
| B10 | case | `:1432` `case out.AttemptID != "" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:` | 예 |
| B11 | if | `:1436` `if err == nil {` | 아니오 |
| B12 | case | `:1441` `default:` | 예 |
| B13 | if | `:1443` `if detail == "" && err != nil {` | 아니오 |

## Calls and live bindings

`applyFloor` · `IssueReduction` · `AttachExitIntent` · `sellIntent` · `Submit.Place` · `release`(= `ReleaseUnacceptedExitProposal`) · `noteDelay` · `alertProposalRefused`.

## State mutations and fallbacks

발의 해제(release — 원장 판정이 허락할 때만). 주문 1(Place).

## Safety conclusion

- B10(a094): `AttemptID != ""` 이고 상태가 NOT_DISPATCHED · FAILED_CONFIRMED 가 아니면 무장 유지 + 오류 — 결과를 못 쓴 제출 위에 두 번째 매도를 얹지 않는다(D−2.5). B12 default 는 R1 의 FAILED_CONFIRMED 가 가는 곳(2.9). High-risk: yes.
