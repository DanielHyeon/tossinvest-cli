# Function Logic Map: `ExitObserver.alertProposalRefused`

- Source: `internal/app/engine/exitloop.go` (**편집 전** `1927`–`1952`, base `064f59fd9` → 재고정 `b93ed133a`)
- Qualified: `ExitObserver.alertProposalRefused`
- AST evidence: `ast.json` (편집 전 `source_sha256` 8d29c9484a5a…) — branches 3 · returns 0
- Risk scan: `risk-pattern-report.md`
- **편집 뒤 갱신(a197 3.9, 2026-11-25)**: `ast.json` · 위험 보고 재생성, 이 문서의 이 파일 좌표는 base `b93ed133a` → 작업 트리로 difflib 줄 정렬해 옮김(다른 파일 좌표 무변), 분기 번호는 위치라 새 분기 뒤를 재번호함. 분기 0 → 3(`switch` · 두 `case`, D9).
- 작성 시점: **편집 전**(a197 tasks 1.6). 분기 없는 함수 — 행동 행 B1 하나.
- **2026-11-25 재추출**: 현재 HEAD(`f18de89ef`, base `b93ed133a`) AST 로 `ast.json` 을 다시 뽑음. 본문 구조는 옛 base 와 같음(AST 정규화 비교 동일 — 분기 · 호출 · 반환 동일, 줄만 이동: a198 이 `ExitObserver` 구조체에 필드 2줄+주석 1줄, `clearTheSymbol` 에 8줄을 더함). 이 문서의 `exitloop.go` 좌표는 그 이동량(구조체 뒤 +3, `clearTheSymbol` 뒤 +11)으로 기계적으로 옮겼고 이동 뒤 4 자리를 원문과 대조함.

**역할.** 청산 발의가 제출되지 않았음을 운영자 알림 하나로 기록한다. 호출 자리는 `ExitObserver.submit` 안 **둘**(AST calls: `:1533` B3 Guardian 거절, `:1610` B15 default). CodeGraph 1.6.0 `callers alertProposalRefused` = 함수 1(`submit`) — 자리 수는 AST 로 셌다.

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
| B1 | `:1931` `switch` (a197 D9 — 결과 상태 · 사유로 본문 선택) | 기본 본문 「…브로커에 닿기 전에 거부됐다」(Guardian 거절 · `NOT_DISPATCHED`) | 없음 | `TestARefusedProposalReleasesTheLevelAndAlerts` |
| B2 | `:1932` `case state == FAILED_CONFIRMED && reason == ReasonOppositePendingOrder` | 본문 「이 종목에 매수 미체결 주문이 있어 브로커가 이 청산 주문을 거절했다. 앱에서 그 매수 주문을 취소해야 청산이 나간다」 | 없음 | `TestA197TheToggleOffReadsAndCancelsNothing` |
| B3 | `:1935` `case state == FAILED_CONFIRMED` | 본문 「…를 브로커가 거절했다」 | 없음 | 미측정(이 갈래를 단언하는 시험 없음) |

본문 문장의 참·거짓: 자리 1(Guardian 거절)과 B15 의 `NOT_DISPATCHED` 에서는 참, B15 의 `FAILED_CONFIRMED`(브로커 거절, `ReasonOppositePendingOrder` 포함)에서는 거짓. 함수는 결과 상태를 인자로 받지 않으므로 지금은 가를 수 없다(D9 편집은 시그니처 변경 → 두 호출 자리 모두 편집).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.alert` | 알림 기록 | `Alerts` nil 이면 무동작; `Notify` 오류는 `logErr` 로 남고 삼킴(`:1987-1994`) | AST `:1938` |
| `o.label` | 종목 표시(이름(코드)) | — | AST `:1942` |
| `fmt.Sprintf` | 본문 | — | AST `:1919` |
| `string` | 키 · 필드 변환 | — | AST `:1940` `:1941` `:1947` |

Live config binding: 없음. 생산의 `Alerts` 는 `obs.RecordOnly`(기록 전용, `exitwiring.go:350`), `EventExitProposalRefused` 등급은 critical(`obs/event.go:381`).

## State mutations and fallbacks

- 알림 원장 기록 1(critical 이면 `recordCritical`, 재알림 창 1h). 그 밖의 상태 변화 없음.

## Safety conclusion

- Safe edit boundary (D9): 본문 문장을 결과(Guardian 거절 · 미발송 · 브로커 확정 거절 · 그중 반대 미체결)에 따라 고르는 것만. **key · 유형 · 등급 · Fields 의 키 집합은 그대로**(D9 「알림 수는 그대로」).
- 기존 `detail` 필드는 게이트웨이 `out.Detail`(브로커 오류 본문 포함 가능)을 이미 싣는다 — a197 범위 밖의 기존 사실로 기록만 한다.
- High-risk impact: no(알림 문장) — 단 호출자 `submit` 은 High-risk.

## 재개 개정(2026-11-25)

- 편집 내용은 2026-11-23 판과 같다(design D9): 결과 상태 · 사유를 인자로 받아 문장을 가른다. key · 등급 · 재알림 창 무변. 토글과 무관.
