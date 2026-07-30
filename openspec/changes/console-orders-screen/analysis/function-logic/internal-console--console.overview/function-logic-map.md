# Function Logic Map: `Console.overview`

- Source: `internal/console/overview.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

개요 화면 조립. 패널은 서로 독립이고 **어느 것도 에러를 돌려주지 않는다** — 여기서 실패는 전부 화면이 말로 가진 상태다. console-orders-screen이 바꾼 것은 한 줄이다: `v.Open = openOrdersPanel{Count: unmeasuredFor(reasonSeamUnwired)}` → `v.Open = c.openOrdersPanelFrom(now)`. `seam_unwired`는 '이 seam을 배선하라'는 지시였고 이 change가 그것을 배선했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.now()` | 주입 시계 | `Options.Now` | 해당 없음 |
| `c.opts.Binary()` | 설치 바이너리 지문 | `binstamp` | 실패하면 zero stamp — '변경 없음' |
| 원장 | `OpenReadOnly` 1회 | `overviewLedger` | 한 번 열고 하나의 `journalView`를 낸다 — 두 판독이 서로 다른 말을 하는 상태를 없앤 것이 이 함수의 설계 |
| 브로커 캐시 / 주문 캐시 | `peek`만 | `holdingsCache` / `ordersCache` | 브로커 호출 0콜이 이 화면의 계약(D4) |
| `c.verifyHold(now)` | (보류 여부, 사유) | in-process run + runlock | 값의 사유가 아니라 **패널 옆 별도 안내**로 렌더된다(issues.md I-5) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 본문에 분기가 없다 | 단일 경로 | `TestTheOverviewMakesNoBrokerCall`, `TestAnUnreadableJournalEmptiesOnlyItsOwnPanels`, `TestTheLedgerIsOpenedOncePerRenderAndTheNoticePrintsOnce`, `TestTheOverviewHasNoFormAndNoConfirmationInput` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.readEngine(now, installed)` | 엔진 실행 여부·거부 사유 | `enginelock` 마커만 읽는다 — flock을 엔진과 다투지 않는다. 콘솔은 게이트를 **판정하지 않는다** | engineproc.go, `TestTheConsoleDecidesNothingAboutTheGate` |
| `c.overviewLedger(ctx)` | 원장 1회 open, 1개 view | 부분 답 + 사유 | overview.go:540 |
| `c.accountPanelFrom` / `todayPanelFrom` / `recentPanelFrom` | 패널별 조립 | 시장을 가로지르는 합계를 만들지 않는다(D6) | overview.go:648, 800, 914 |
| `c.openOrdersPanelFrom(now)` | 미체결 건수 | `peek` — 0콜. 한쪽 목록이 미측정이면 합산하지 않는다 | overview.go:770 |
| `c.safetyPanelFrom(v.Today)` | 잔여물 패널 | KR·US 두 시장 모두 읽는다(리뷰 P1-9) | overview.go:928 |
| (금지 바인딩) | 브로커 호출 0, 원장은 read-only, config 타입을 쥐지 않는다 — Guardian 한도는 `GateLimits() (GateLimits, error)` 값 타입 seam으로 온다(issues.md I-1) | `TestTheOverviewMakesNoBrokerCall`, `TestTheConsoleOpensTheJournalReadOnly`, `TestTheConsoleDecidesNothingAboutTheGate` | overview_test.go:236 외 |

## State mutations and fallbacks

- `overviewView` 값 하나를 만든다. 계좌·원장·config 쓰기 없음.
- 패널 단위 격리: 한 패널의 실패가 다른 패널을 비우지 않는다(`TestAnUnreadableJournalEmptiesOnlyItsOwnPanels`, `TestAnUnwiredGateLimitsSeamOnlyDarkensItsOwnPanel`).
- 미측정은 값이 아니라 사유 코드를 갖는다 — 일곱 코드 + `signals.go`의 여덟째(`discovery_unreadable`, 개요의 열거에는 넣지 않았다).

## Safety conclusion

- Safe edit boundary: `v.Open` 한 줄. 나머지 여섯 패널의 조립 순서와 인자는 무변경.
- High-risk impact: yes (원장 + Guardian 한도 표시 + 잔여물 건수 — 운영자가 '다음 검증을 시작해도 되는가'를 이 화면에서 판단한다)
