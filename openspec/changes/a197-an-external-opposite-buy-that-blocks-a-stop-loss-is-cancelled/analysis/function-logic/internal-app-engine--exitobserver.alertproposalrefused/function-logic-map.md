# Function Logic Map: `ExitObserver.alertProposalRefused`

- Source: `internal/app/engine/exitloop.go` (**편집 전** `1901`–`1918`, base `064f59fd9`)
- Qualified: `ExitObserver.alertProposalRefused`
- AST evidence: `ast.json` (편집 전 `source_sha256` bc81b5fddef6b143…) — branches 0 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(a197 tasks 1.6). 분기 없는 함수 — 행동 행 B1 하나.

**역할.** 청산 발의가 제출되지 않았음을 운영자 알림 하나로 기록한다. 호출 자리는 `ExitObserver.submit` 안 **둘**(AST calls: `:1530` B3 Guardian 거절, `:1607` B15 default). CodeGraph 1.6.0 `callers alertProposalRefused` = 함수 1(`submit`) — 자리 수는 AST 로 셌다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m` | 포지션 + exit state | `submit` | — |
| `proposal` | `Action` · `Level` | `submit` | — |
| `detail` | 원인 문자열(게이트웨이 `out.Detail` 또는 오류 문장) | `submit` | 본문 · `Fields[detail]` 에 그대로 실림 |
| 불변식 | key = `exit.proposal_refused|<position>|<action>|<level>` — 같은 발의 레벨의 반복 거절은 한 key(재알림 창 `obs.DefaultRemindAfter` = 1h, `notifier.go:60`) | `obs` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 항상 `o.alert(obs.Event{…})` `:1903` | 제목 `「<종목> 청산 주문이 제출되지 않았다」` `:1907`, 본문 `「%s (단계 %s)가 브로커에 닿기 전에 거부됐다.\n원인: %s」` `:1908` | 없음 | 거절 알림 유형 · 등급 |

본문 문장의 참·거짓: 자리 1(Guardian 거절)과 B15 의 `NOT_DISPATCHED` 에서는 참, B15 의 `FAILED_CONFIRMED`(브로커 거절, `ReasonOppositePendingOrder` 포함)에서는 거짓. 함수는 결과 상태를 인자로 받지 않으므로 지금은 가를 수 없다(D9 편집은 시그니처 변경 → 두 호출 자리 모두 편집).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.alert` | 알림 기록 | `Alerts` nil 이면 무동작; `Notify` 오류는 `logErr` 로 남고 삼킴(`:1953-1960`) | AST `:1903` |
| `o.label` | 종목 표시(이름(코드)) | — | AST `:1907` |
| `fmt.Sprintf` | 본문 | — | AST `:1908` |
| `string` | 키 · 필드 변환 | — | AST `:1905` `:1906` `:1913` |

Live config binding: 없음. 생산의 `Alerts` 는 `obs.RecordOnly`(기록 전용, `exitwiring.go:350`), `EventExitProposalRefused` 등급은 critical(`obs/event.go:381`).

## State mutations and fallbacks

- 알림 원장 기록 1(critical 이면 `recordCritical`, 재알림 창 1h). 그 밖의 상태 변화 없음.

## Safety conclusion

- Safe edit boundary (D9): 본문 문장을 결과(Guardian 거절 · 미발송 · 브로커 확정 거절 · 그중 반대 미체결)에 따라 고르는 것만. **key · 유형 · 등급 · Fields 의 키 집합은 그대로**(D9 「알림 수는 그대로」).
- 기존 `detail` 필드는 게이트웨이 `out.Detail`(브로커 오류 본문 포함 가능)을 이미 싣는다 — a197 범위 밖의 기존 사실로 기록만 한다.
- High-risk impact: no(알림 문장) — 단 호출자 `submit` 은 High-risk.
