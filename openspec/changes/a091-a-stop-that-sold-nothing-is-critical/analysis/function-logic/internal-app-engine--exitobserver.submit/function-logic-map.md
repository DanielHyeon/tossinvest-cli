# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (`1395`–`1479`)
- Qualified: `ExitObserver.submit`
- AST evidence: `ast.json` (`source_sha256` 0f943813a3efa423…) — base `b30318d6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 13 · 반환 10

**편집.** **a091 편집 대상**(tasks 3.6) — `applyFloor` 호출에 `isProtective(proposal)` 를 넘긴다. 분기 · 반환 무변경.

**역할.** 무장된 발의를 브로커로 보낸다. 0주(B2)면 제출 없이 해제한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `proposal` | `Action.Orderable()` 5종 중 하나 | `record` | `isProtective` = BaselineBreach · LadderStop(`exitloop.go:1375-1377`) |
| `quantity` | 양의 정수 정규형 | `record` ← `snapshot.ProjectedQuantity` | — |
| `submitQuantity` | `applyFloor` 반환 | — | `isZeroQuantity` 가 0 판정(수치 비교 — 아래) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1398` `if err != nil {` | 아니오 |
| B2 | if | `:1401` `if isZeroQuantity(submitQuantity) {` | 예 |
| B3 | if | `:1421` `if err != nil {` | 아니오 |
| B4 | if | `:1430` `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | 예 |
| B5 | if | `:1435` `if err != nil {` | 아니오 |
| B6 | switch | `:1445` `switch {` | — |
| B7 | case | `:1446` `case out.State == journal.StateConfirmed:` | 예 |
| B8 | case | `:1454` `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | 예 |
| B9 | case | `:1459` `case out.Reason == execgw.ReasonSymbolInFlight:` | 아니오 |
| B10 | case | `:1462` `case out.AttemptID != "" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:` | 예 |
| B11 | if | `:1466` `if err == nil {` | 아니오 |
| B12 | case | `:1471` `default:` | 예 |
| B13 | if | `:1473` `if detail == "" && err != nil {` | 아니오 |

Exact AST return positions: `1399:3`, `1406:3`, `1423:3`, `1431:3`, `1437:3`, `1453:3`, `1458:3`, `1461:3`, `1469:3`, `1477:3`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `o.applyFloor` | `:1397` | 확정 하한 | 위 applyFloor 번들의 계약 — RECONCILE 에서 브로커 읽기 ≤2 Query |
| `isZeroQuantity` | `:1401` | 0주 판정 | 순수. **수치 비교**(`CompareDecimal(q,"0") <= 0`, 파싱 실패 · 빈 문자열도 0) |
| `o.release` | `:1406` | 0주 → 발의 해제(`ProposalRefused`) | 원장 트랜잭션(`ReleaseUnacceptedExitProposal`), busy_timeout 5s |
| `o.opts.Issuer.IssueReduction` | `:1409` | Guardian 축소 발행 | 로컬 판정(브로커 0) |
| `costs.Market` | `:1412` | 시장 변환 | 순수 |
| `strings.ToLower` | `:1412` | — | 순수 |
| `strings.TrimSpace` | `:1412` | — | 순수 |
| `fmt.Sprintf` | `:1418` | 사유 문구 | 순수 |
| `o.alertProposalRefused` | `:1422` | 발행 거절 알림 | critical → 기록 전용(ALERT 계약) |
| `err.Error` | `:1422` | — | 순수 |
| `o.release` | `:1423` | 해제 | 원장 |
| `o.opts.Journal.AttachExitIntent` | `:1430` | intent 부착 | 원장 |
| `fmt.Errorf` | `:1431` | — | 순수 |
| `o.sellIntent` | `:1434` | 주문 모양 | 순수 |
| `o.alertRefused` | `:1436` | 판정 거절 알림 | critical → 기록 전용 |
| `o.release` | `:1437` | 해제 | 원장 |
| `o.opts.Submit.Place` | `:1439` | **유일한 브로커 mutation**(매도 1) | 게이트웨이 계약(a094 · order-execution) — a091 무변경 |
| `o.log` | `:1447` | 체결 확정 로그 | 로그 한 줄 |
| `string` | `:1450` | — | 순수 |
| `o.noteDelay` | `:1460` | 심볼 점유 지연 | critical 지연 알림(기록 전용) |
| `o.release` | `:1461` | 해제(취소) | 원장 |
| `fmt.Errorf` | `:1467` | — | 순수 |
| `fmt.Errorf` | `:1469` | — | 순수 |
| `err.Error` | `:1474` | — | 순수 |
| `o.alertProposalRefused` | `:1476` | 거절 알림 | critical → 기록 전용 |
| `o.release` | `:1477` | 해제 | 원장 |

## State mutations and fallbacks

발의 해제(원장) · 주문 1(Place). 0주 경로(B2)는 제출 0 · 해제 1.

## Safety conclusion

- a091 은 `applyFloor` 호출의 인자 하나만 바꾼다 — B1~B13 무변경. High-risk: yes.
