# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (**편집 전** `1509`–`1634`, base `064f59fd9` → 재고정 `b93ed133a`)
- Qualified: `ExitObserver.submit`
- AST evidence: `ast.json` (편집 전 `source_sha256` 8d29c9484a5a…) — branches 18 · returns 10 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md` (파일 단위)
- **편집 뒤 갱신(a197 3.9, 2026-11-25)**: `ast.json` · 위험 보고 재생성, 이 문서의 이 파일 좌표는 base `b93ed133a` → 작업 트리로 difflib 줄 정렬해 옮김(다른 파일 좌표 무변), 분기 번호는 위치라 새 분기 뒤를 재번호함. B17 `:1623` 새 분기(대기열 넣기), 옛 B17 → B18. 호출 자리 2 의 `alertProposalRefused` 인자 둘 추가.
- 작성 시점: **편집 전**(a197 tasks 1.6, 2026-11-23). 아래 분기 주장은 전부 `ast.json` 열거에서 왔다. 편집 뒤 좌표는 tasks 3.9 에서 이 문서를 갱신한다.
- **2026-11-25 재추출**: 현재 HEAD(`f18de89ef`, base `b93ed133a`) AST 로 `ast.json` 을 다시 뽑음. 본문 구조는 옛 base 와 같음(AST 정규화 비교 동일 — 분기 · 호출 · 반환 동일, 줄만 이동: a198 이 `ExitObserver` 구조체에 필드 2줄+주석 1줄, `clearTheSymbol` 에 8줄을 더함). 이 문서의 `exitloop.go` 좌표는 그 이동량(구조체 뒤 +3, `clearTheSymbol` 뒤 +11)으로 기계적으로 옮겼고 이동 뒤 4 자리를 원문과 대조함.

**역할.** 무장된 청산 발의 하나를 브로커에 낸다: 확정 하한 캡 → Guardian 위험 감소 결정 → intent 부착 → 주문 모양 → `Place` → 결과별 처리.
호출자는 `ExitObserver.record` 하나(`exitloop.go:1467`, CodeGraph 1.6.0 `callers submit` 의 비시험 4건 중 엔진 쪽 1 — 나머지 셋은 `Gateway.place/Cancel/Amend` 가 부르는 동명 `Gateway.submit`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m` | 이번 주기 작업 집합의 포지션 + exit state | `workingSet` | — |
| `proposal` | 무장된 발의(`Action`·`Level`). 보호 여부는 `isProtective` = `ActionBaselineBreach ∨ ActionLadderStop`(`:1489-1491`) | `record` 의 `snapshot.ExecutableProposal()` | — |
| `quantity` | 투영 수량(decimal) | `snapshot.ProjectedQuantity` | 하한 캡에서 줄 수 있음 |
| `intentID` | 원장이 무장한 intent | `recorded.ArmedProposal.IntentID` | — |
| `observed` | 판정 관측가 | `judgement.ObservedPrice` | 비면 `sellIntent` 가 기준선 사용 |
| 불변식 | 브로커 mutation 0~1 회(`Place`), 해제는 원장 판정 하나(`release`) | `exit_proposal_release.go` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:1514` `if err != nil` (`applyFloor`) | `noteExitOrder(floor_arithmetic_failed)` | `return err` `:1518` (해제 없음) | a129 B1 |
| B2 | `:1521` `if isZeroQuantity(submitQuantity)` | `release(ProposalRefused)` · 줄 | `:1528` 해제 결과 | a129 B2 · zero floor |
| B3 | `:1543` `if err != nil` (`IssueReduction`) | **`alertProposalRefused` 자리 1** · `release` · 줄 | `:1547` | a129 B3 |
| B4 | `:1554` `if err := AttachExitIntent…` | 줄 | `:1556` wrap | a129 B4 |
| B5 | `:1560` `if err != nil` (`sellIntent`) | `alertRefused` · `release` · 줄 | `:1564` | 도달 시험 없음(관측가가 늘 수) |
| B6 | `:1567` `if isProtective(proposal)` | 제출 context 우선순위 Critical | — | a166 우선순위(무편집) |
| B7 | `:1576` `if isProtective(proposal)` | `markProtectiveAttempt` (Place 반환 뒤) | — | a087 구조 시험 |
| B8 | `:1581` `switch` (결과 분류) | `line = line.placed(out)` `:1580` | — | B9~B18 |
| B9 | `:1582` `case out.State == StateConfirmed` | `o.log` · 줄 | `return nil` `:1590` | 확정 줄 |
| B10 | `:1591` `case InDoubt ∨ UnresolvedInDoubt` | 줄(발의 무장 유지) | `return nil` `:1600` | 무장 유지 |
| B11 | `:1596` `if out.State == StateUnresolvedInDoubt` | 이벤트 선택만 | — | a129 unresolved |
| B12 | `:1601` `case out.Reason == ReasonSymbolInFlight` | `noteDelay` · `release(ProposalCancelled)` · 줄 | `:1605` | a129 symbol in flight |
| B13 | `:1606` `case out.AttemptID != "" ∧ State ∉ {NOT_DISPATCHED, FAILED_CONFIRMED}` | 줄(무장 유지) | `:1614` wrap | a094 unrecorded outcome |
| B14 | `:1610` `if err == nil` (B13 안) | `err = fmt.Errorf(…)` | — | B13 과 같음 |
| B15 | `:1616` `default` — **a197 편집 자리.** `FAILED_CONFIRMED`(사유 `ReasonOppositePendingOrder` 포함)와 `NOT_DISPATCHED` 가 여기로 옴 | **`alertProposalRefused` 자리 2** `:1610` · `release(ProposalRefused)` `:1630` · 줄 | `:1632` 해제 결과 | 거절 해제 · 재발의 |
| B16 | `:1618` `if detail == "" && err != nil` (B15 안) | `detail = err.Error()` — 알림 본문에만 | — | 미측정 |
| B17 | `:1623` `if externalBuyOpens(proposal, out)` (B15 안, **a197 새 분기**) | `noteExternalBuyBlock(m)` — 메모리 대기열에 넣기만(토글 꺼짐이면 그 함수가 안 넣음) · 브로커 호출 0 | — | `TestA197AnExternalBuyIsCancelledAtTheCycleEndAndTheStopFollows` · `TestA197TheToggleOffReadsAndCancelsNothing` · `TestA197TheStageOpensOnlyOnAProtectiveOppositePendingRefusal` |
| B18 | `:1627` `if out.State == StateFailedConfirmed` (B15 안) | `order.rejected` / `order.not_dispatched` 선택 | — | a129 broker_rejected / not_dispatched |

`ReasonOppositePendingOrder` 는 게이트웨이에서 `out.Reason = ReasonCode(res.ReasonCode)`(`gateway.go:756`)로 오고, 그 값은 `classifyMutation` B4 `case refusalCodeDefinitive`(`classify.go:45`) → `DispatchRejected` → `Settle(StateFailedConfirmed, …)`(`journal/dispatch.go:158-160`)이다. B12 가 B15 보다 앞이고 B13 이 `FAILED_CONFIRMED` 를 빼므로 그 결과는 B15 에만 닿는다(AST 순서 `:1601` → `:1606` → `:1616`).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitLimitPrice` | 지정가 결정(a087) | 오류 없음 | AST `:1512` |
| `o.applyFloor` | RECONCILE 확정 하한 캡 | 오류 → B1; 0 → B2 | AST `:1513` |
| `o.opts.Issuer.IssueReduction` | Guardian 위험 감소 결정(생산 = `*execgw.RiskGuardian`, Side SELL · Kind PLACE 고정) | 오류 → B3 | AST `:1531` |
| `o.opts.Journal.AttachExitIntent` | 제출 전 intent 재확인 | 오류 → B4 | AST `:1554` |
| `o.sellIntent` | 주문 모양 | 오류 → B5 | AST `:1559` |
| `o.opts.Submit.Place` | **유일한 브로커 호출** | 결과 분류 B9~B18, 재시도 0 | AST `:1570` |
| `o.opts.lowerLimitBoard.markProtectiveAttempt` | a087 첫 시도 표시 | — | AST `:1577` |
| `o.alertProposalRefused` | 거절 알림(자리 2: `:1533` · `:1610`) | 기록 전용 알림기(`Alerts` = `obs.RecordOnly`) | AST |
| `o.alertRefused` · `o.noteDelay` · `o.log` · `o.noteExitOrder` · `line.because` · `line.placed` · `exitOrderLineFor` | 알림 · 지연 시계 · 줄 | 메모리/기록 전용 | AST |
| `o.release` | 비수용 발의 해제(원장 판정) | 오류 → wrap 반환 | AST `:1526` `:1545` `:1562` `:1603` `:1630` |
| `accountbudget.WithPriority` | a166 계좌 예산 우선순위 | — | AST `:1566` `:1568` |

Live config binding: 없음(이 함수는 설정을 읽지 않는다). a197 의 토글은 `ExitObserverOptions` 의 새 칸으로 들어오며 기동 때 고정된다(precedent `SourceAgeReport`).

## State mutations and fallbacks

- 원장: `AttachExitIntent`(B4 앞), `release`(B2 · B3 · B5 · B12 · B15), Guardian 결정 행(`IssueReduction`).
- 메모리: `noteDelay`(B12), `markProtectiveAttempt`, a129 줄 대기열.
- 폴백: `sellIntent` 의 지정가는 관측가, 비면 기준선.

## Safety conclusion

- Safe edit boundary (a197 계약 기준): B15 안에서만 — `alertProposalRefused` 의 인자에 결과 상태·사유를 넘기고(D9), 토글 켜짐 ∧ `isProtective(proposal)` ∧ `out.State == StateFailedConfirmed` ∧ `out.Reason == ReasonOppositePendingOrder` 일 때 새 단계(새 파일)를 부른다(D1). **`release` 호출과 그 반환값은 그대로**(D7). B12 · B13 · 그 밖의 갈래와 `Place` 앞은 무편집. 자리 1(`:1533`, Guardian 거절)은 「브로커에 닿기 전」이 참인 자리라 D9 의 문장 선택이 그대로 두어야 한다.
- 구조 고정: `a087_submit_structure_test.go` 는 `exitLimitPrice < applyFloor`, `exitOrderLineFor` 6번째 인자 `choice`, `sellIntent` 3번째 인자 `choice.price`, `Place < markProtectiveAttempt` 를 본다 — B15 안의 새 호출이 이 다섯 이름을 쓰지 않으면 영향 없음. a163 렌더 해시는 `record` 를 굳히고 `submit` 은 굳히지 않는다.
- **차단 사실(1.3 probe)**: B3 의 Guardian 결정(`RiskGuardian.IssueReduction`)으로는 취소를 낼 수 없다 — 새 단계가 같은 `Issuer` 로 취소 결정을 만들면 게이트웨이 `Prepare` 가 거절한다(`analysis/probe/probe-output.txt`). 새 단계의 결정 발행 경로는 Manager 결정 대기. → **2026-11-25: a198 착지로 해소** — 길 치우기는 `execgw.Issuer`(Kind CANCEL, Side = 주문 방향)로 발행하고, 같은 발행으로 외부 매수 번호를 엔진 게이트웨이에 내면 접수 확정(`analysis/probe/resume-probe-output.txt` Q1).
- High-risk impact: yes — 주문 제출·취소 경로.

## 재개 개정(2026-11-25) — 편집 확정

- B15 `case@1616` 안: (1) `alertProposalRefused` 에 결과 상태 · 사유를 넘김(D9, 자리 2 — B3 `:1533` 도 같은 시그니처로), (2) 토글 켜짐 ∧ `isProtective` ∧ `FAILED_CONFIRMED` ∧ `ReasonOppositePendingOrder` 이면 관측기 메모리 대기열에 「열 조건」을 넣기만 한다 — **브로커 호출 · 원장 쓰기 0**(실행은 주기 끝 `endExitOrderCycle`, design D5). `release` 호출과 반환값은 그대로.
- 2026-11-23 판의 「B15 안에서 새 단계(새 파일)를 부른다」는 폐기 — 판정 경로 안 실행은 같은 주기의 뒤 포지션 판정을 늦춘다(r1 #4).
- a087 구조 핀의 다섯 이름(`exitLimitPrice` · `applyFloor` · `exitOrderLineFor` · `sellIntent` · `Place`/`markProtectiveAttempt`)을 B15 안에 새로 부르지 않는다.
