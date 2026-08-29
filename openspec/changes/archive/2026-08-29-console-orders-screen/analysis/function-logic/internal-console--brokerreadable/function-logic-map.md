# Function Logic Map: `brokerReadable`

- Source: `internal/console/overview.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**본문 무변경** — 이 change의 base 대비 함수 본문이 byte 동일하다(base·HEAD 두 판본의 선언 범위 텍스트를 직접 비교해 확인했다). 인접 hunk 교차로 evidence가 요구됐고, `ast.json`은 base revision에서 뜬 것이다.

브로커 캐시가 쓸 만한 판독을 갖고 있는지, 아니면 왜 없는지를 답한다. **본문 무변경** — console-orders-screen의 base 대비 byte 동일이며, 바로 아래 `openOrdersPanelFrom` 삽입의 diff hunk가 교차해 evidence가 요구됐다. AST는 base revision이다.

현재 모양은 console-operator-overview의 issues.md I-5가 만든 것이다: 코드가 콜드 캐시에 `verify_suspended`를 먼저 답했고 주석은 `never_fetched`라고 적혀 있었다. 그 화면에서 '검증 중 — 갱신 보류 … 끝나면 다시 읽힌다'는 **양쪽으로 다 거짓**이다 — 개요는 계약상 브로커를 부르지 않으므로 보류되는 것이 없고, 검증이 끝나도 그 값은 읽히지 않는다(이 캐시를 채우는 것은 `/positions`를 여는 것뿐이다). 코드를 주석에 맞췄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snap.Wired` | seam이 주입됐는가 | `holdingsCache.peek` | false면 `seam_unwired` |
| `snap.Present` | 판독이 있는가(오래됐어도) | 같은 곳 | true면 측정됨 |
| `snap.Error` | 마지막 갱신 실패 문자열 | 같은 곳 | 비어 있지 않으면 `broker_read_failed` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch {` (사유 선택) | 없음 | 아래 넷 중 하나 | `TestAColdBrokerCacheSaysNotYetReadAndLinksToTheScreenThatFillsIt` 외 |
| B2 | `!snap.Wired` | 없음 | `unmeasuredFor(reasonSeamUnwired)` | 브로커 미배선 콘솔의 개요 렌더 |
| B3 | `snap.Present` | 없음 | `measured("")` | `TestPeekServesWhatTheLastRefreshFoundAndStillCallsNothing` 이후의 개요 렌더 |
| B4 | `snap.Error != ""` | 없음 | `unmeasuredFor(reasonBrokerReadFailed)` | `TestAFailedBrokerReadIsNotTheSameAsAnEmptyCache` |
| B5 | `default` | 없음 | `unmeasuredFor(reasonNeverFetched)` | `TestAColdBrokerCacheSaysNotYetReadAndLinksToTheScreenThatFillsIt`, `TestARunningVerificationIsANoteAndNotTheReasonTheCacheIsCold` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `unmeasuredFor` / `measured` | `reading` 삼중항 생성 | 순수 | overview.go:178, 181 |
| (금지 바인딩) | 계좌·원장 무접촉 — 스냅샷 필드 셋만 읽는다 | 순수 함수 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수·본문 무변경).
- `verify_suspended`는 이 화면에서 **산출되지 않는다**. 사유 열거에서 지우지는 않았다 — 다른 seam이 배선되면 다시 산출될 수 있는 어휘이고 spec이 일곱 중 하나로 고정한다(issues.md I-5).
- 판정 순서가 계약이다: `Wired → Present → Error → never_fetched`. `Present`가 `Error`보다 앞이므로 **실패했지만 이전 판독이 남아 있는** 캐시는 측정됨으로 읽히고, 화면은 나이와 오류 문장을 따로 적는다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입(`openOrdersPanelFrom`)만 존재한다.
- High-risk impact: no (계좌·원장·주문 무접촉, 사유 코드 선택뿐). 이 함수의 산출물은 값이 아니라 **운영자의 다음 행동을 결정하는 문장**이며, I-5가 고친 것은 정확히 그 문장이 없는 사건을 근거로 들고 있었다는 것이다.
