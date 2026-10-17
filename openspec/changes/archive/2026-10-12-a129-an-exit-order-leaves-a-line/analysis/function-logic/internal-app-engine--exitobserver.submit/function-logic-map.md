# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (**편집 뒤** `1416`–`1529`; 편집 전 `1402`–`1487`, base `381e73da`)
- Qualified: `ExitObserver.submit`
- AST evidence: `ast.json` (**편집 뒤** `source_sha256` fba676d31e44590c… — 편집 전 0733bd8641ed8c41…) — 편집 뒤 branches 15 · returns 10 · defers 0 · go 0 (편집 전 13 · 10)
- Risk scan: `risk-pattern-report.md` (파일 단위 — `:1925` `randomExitID` 의 panic, 이 함수 밖)
- 작성 시점: **편집 전**(tasks 0.5) — 아래 「Branches」 표는 편집 전 좌표. 편집 뒤 좌표·새 분기는 끝의 「편집 뒤 (a129)」 절. 이 문서의 분기 주장은 AST 열거에서 왔다.

**역할.** 무장된 청산 발의 하나를 브로커에 낸다: 확정 하한 캡 → Guardian 위험 감소 결정 → intent 부착 → 주문 모양 → `Place` → 결과별 처리.
호출자는 `ExitObserver.record` 하나(CodeGraph `callers ExitObserver.submit`: 생산 1 — 나머지 1 은 `internal/journal/durability_test.go:539` 의 이름 충돌).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m` | 이번 주기 작업 집합의 포지션 + exit state | `workingSet` | — |
| `proposal` | 무장된 발의(`Action`·`Level`) | `record` 의 `snapshot.ExecutableProposal()` | — |
| `quantity` | 투영 수량(decimal) | `snapshot.ProjectedQuantity` | 하한 캡에서 줄 수 있음 |
| `intentID` | 원장이 무장한 intent | `recorded.ArmedProposal.IntentID` | — |
| `observed` | 판정 관측가 | `judgement.ObservedPrice` | 비면 `sellIntent` 가 기준선 사용 |
| 불변식 | 브로커 호출 0~1 회(`Place`), 해제는 원장 판정 하나(`release`) | `exit_proposal_release.go` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:1406` `if err != nil` (`applyFloor`) | 없음 | `return err` `:1407` | a129: `order.not_dispatched` `reason=floor_unavailable` 줄 (base 시험 0 — 블록 count 0) |
| B2 | `:1409` `if isZeroQuantity(submitQuantity)` | `release(…, ProposalRefused)` | `:1414` 해제 결과 | a129: 해제 **뒤** `order.not_dispatched` `reason=confirmed_floor_zero` |
| B3 | `:1429` `if err != nil` (`IssueReduction`) | `alertProposalRefused` · `release` | `:1431` | a129: 해제 뒤 `reason=guardian_refused` (base 시험 0) |
| B4 | `:1438` `if err := AttachExitIntent…; err != nil` | 없음 | `:1439` wrap | a129: `reason=attach_failed` (base 시험 0) |
| B5 | `:1443` `if err != nil` (`sellIntent`) | `alertRefused` · `release` | `:1445` | a129: 해제 뒤 `reason=intent_invalid` (base 시험 0) |
| B6 | `:1453` `switch` (결과 분류) | — | — | 아래 case 들 |
| B7 | `:1454` `case out.State == StateConfirmed` | `o.log(EventOrderConfirmed…)` — 생산에서 `Log` nil 이라 무출력 | `return nil` `:1461` | a129: `order.confirmed` 줄 + 체결 따라가기 등록 |
| B8 | `:1462` `case InDoubt \|\| UnresolvedInDoubt` | 없음(발의 무장 유지) | `return nil` `:1466` | a129: `order.in_doubt` 줄 + 주문 번호 있으면 따라가기 |
| B9 | `:1467` `case out.Reason == ReasonSymbolInFlight` | `noteDelay` · `release(…, ProposalCancelled)` | `:1469` | a129: 해제 뒤 `order.not_dispatched` `reason=symbol_mutation_in_flight` (base 시험 0) |
| B10 | `:1470` `case out.AttemptID != "" && State ∉ {NOT_DISPATCHED, FAILED_CONFIRMED}` | 없음(무장 유지) | `:1477` wrap | a129: `order.in_doubt` 줄 |
| B11 | `:1474` `if err == nil` (B10 안) | `err = fmt.Errorf(…)` | — | B10 과 같은 줄(분기 무관) |
| B12 | `:1479` `default` | `alertProposalRefused` · `release` | `:1485` | a129: 해제 뒤 `FAILED_CONFIRMED` 면 `order.rejected`, 아니면 `order.not_dispatched` |
| B13 | `:1481` `if detail == "" && err != nil` (B12 안) | `detail = err.Error()` — 알림 본문에만 | — | 줄은 `detail` 을 싣지 않음(분기 무관) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.applyFloor` | RECONCILE 확정 하한 캡 | 오류 → B1 return; 0 → B2 | CodeGraph callees + AST `:1405` |
| `o.opts.Issuer.IssueReduction` | Guardian 위험 감소 결정 | 오류 → 알림 + 해제(B3) | AST `:1417` |
| `o.opts.Journal.AttachExitIntent` | 제출 전 intent 재확인 | 오류 → wrap 반환(B4), 해제 없음 | AST `:1438` |
| `o.sellIntent` | 주문 모양(LIMIT SELL, 관측가) | 오류 → 판정 거부 알림 + 해제(B5) | AST `:1442` |
| `o.opts.Submit.Place` | **유일한 브로커 호출** | 결과 분류 B7~B12 | AST `:1447` |
| `o.log` | 확정 줄(`Log` 싱크) | `Log` nil 이면 무동작(`:1842`) | AST `:1455` |
| `o.noteDelay` · `o.alertProposalRefused` · `o.alertRefused` | 지연 시계 · 알림 | 기록 전용 알림기 | AST |
| `o.release` | 비수용 발의 해제(원장 판정) | 오류 → wrap 반환 | AST `:1414` `:1431` `:1445` `:1469` `:1485` |

Live config binding: 없음(이 함수는 설정을 읽지 않는다). `ExitOrderLog` 는 a129 가 더하는 옵션.

## State mutations and fallbacks

- 원장: `AttachExitIntent`(B4 앞), `release`(B2 · B3 · B5 · B9 · B12). 메모리: `noteDelay` 의 지연 시계(B9).
- 폴백: `sellIntent` 의 지정가는 관측가, 비면 기준선.
- a129 가 더하는 것: 메모리의 체결 따라가기 항목(B7 · B8 · B10, `ExitOrderLog` 있을 때만)과 로그 줄. 원장 쓰기·브로커 호출 추가 0.

## Safety conclusion

- Safe edit boundary: 각 return 경로에 줄 하나를 **더하기만** 한다. 해제가 있는 갈래는 `err := o.release(…)` 뒤에 줄을 쓰고 그 `err` 를 그대로
  돌려준다. 줄은 `Place` 앞에 서지 않는다. 반환값·알림·지연 시계·`:1455` 는 바꾸지 않는다. 줄 쓰기는 recover 로 감싼다.
- High-risk impact: yes — 주문 제출 경로. 동작 변화 0 을 기존 시험(B2 · B7 · B8 · B10 · B12 진입)과 a129 시험(전 갈래 줄 + 패닉 로거)으로 잰다.

## 편집 뒤 (a129)

difflib 없이 손으로 맞춘 재번호이며 근거는 편집 뒤 `ast.json` 열거다. 새 분기 둘(B9 · B15)은 줄 이벤트를 고르는 `if` 뿐이고 반환값·해제에 닿지 않는다.

| 편집 뒤 | 편집 전 | 좌표 | a129 편집 |
|---|---|---|---|
| B1 | B1 | `:1420` 하한 산술 실패 | `noteExitOrder(not_dispatched, floor_arithmetic_failed)` 뒤 `return err` (해제 없음) |
| B2 | B2 | `:1427` 하한 0 | `releaseErr := o.release(…)` → `noteExitOrder(… confirmed_floor_zero)` → `return releaseErr` |
| B3 | B3 | `:1449` Guardian 거절 | 알림 → 해제 → 줄(`guardian_refused`) → 해제 결과 반환 |
| B4 | B4 | `:1460` intent 부착 실패 | 줄(`attach_failed`) → wrap 반환 |
| B5 | B5 | `:1466` 주문 모양 실패 | 알림 → 해제 → 줄(`intent_invalid`) → 해제 결과 반환 |
| B6 | B6 | `:1480` 결과 switch | `line = line.placed(out)` 가 switch 바로 앞(`Place` 뒤) |
| B7 | B7 | `:1481` 확정 | 기존 `o.log` 그대로 + `noteExitOrder(order.confirmed, follow)` |
| B8 | B8 | `:1490` 결과 미상 | `noteExitOrder(order.in_doubt \| order.unresolved_in_doubt, follow)` |
| B9 | (새) | `:1495` `if out.State == StateUnresolvedInDoubt` | 이벤트 선택만 |
| B10 | B9 | `:1500` 종목 진행 중 | 지연 시계 → 해제 → 줄 → 해제 결과 반환 |
| B11 | B10 | `:1505` attempt 있음 · 비수용 미입증 | 줄(`order.in_doubt`, follow) → wrap 반환 |
| B12 | B11 | `:1509` B11 안 `err == nil` | 무편집 |
| B13 | B12 | `:1515` 기본 | 알림 → 해제 → 줄 → 해제 결과 반환 |
| B14 | B13 | `:1517` 상세 비고 오류 있음 | 무편집(알림 본문에만, 줄에 안 실림) |
| B15 | (새) | `:1522` `if out.State == StateFailedConfirmed` | `order.rejected` / `order.not_dispatched` 선택만 |

`noteExitOrder` 는 메모리 append 와 따라가기 등록뿐이다(쓰기·I/O 없음, `exit_order_log.go`). 쓰기는 `ObserveOnce` 의 defer 가 주기 끝에 한다.
반환값은 모든 갈래에서 편집 전과 같다 — 해제 갈래는 `return o.release(…)` 를 `releaseErr := o.release(…); …; return releaseErr` 로 풀었을 뿐이다.
