# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go:319-365`
- AST evidence: `ast.json` (Pre-Edit HEAD `e1eaf720c`, source_sha256 `e8f131a3…` — a087 의 current 번들과 같은 값, 분기 6)
- Risk scan: `risk-pattern-report.md` (적중 0)
- 판: **pre-edit**. 편집 범위는 design D2 의 안에 달림 — (나)·(가-A): B3 뒤 취소 결정 발행기 단언·대입 1~2줄(D3: 실패면 ErrExitObserverUnavailable). (가-B): 이 함수 무편집(edit-surface.md §3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | non-nil | 엔진 기동 | ErrExitObserverUnavailable(B1) |
| `c.Automation.Verified` | true | interlock | 거절(B2) |
| `c.Guardian` | `ReductionIssuer` 구현(운영 = `*execgw.RiskGuardian`) | 엔진 조립 | 거절(B3) |
| `c.Notifier` | 있으면 기록 전용 경로로 덮음 | 조립 | 없으면 호출자 값 |
| 불변식 | 운영 부품(Journal·Prices·Retrier·Issuer·Submit·AccountRef·정책)은 호출자 값과 무관하게 덮음 | a092 C1 · a087 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L320 | `c == nil` | 없음 | ErrExitObserverUnavailable | 기존 |
| B2 L323 | 게이트 미검증 | 없음 | ErrExitObserverUnavailable | tracer_test.go:338 |
| B3 L327 | Guardian 이 ReductionIssuer 아님 | 없음 | ErrExitObserverUnavailable | cmd/tossctl/engine_runtime_branch_test.go:64-68 |
| B4 L344 | Names nil | 기본 이름표 | 계속 | 기존 |
| B5 L349 | Notifier 있음 | Alerts·Announcer·Critical 덮음 | 계속 | a092·a094 wiring |
| B6 L361 | Floor nil | 기본 상한 | 계속 | 기존 |
| (편집 예정) | 취소 결정 발행기 단언 실패 | 없음 | ErrExitObserverUnavailable(D3) | a198 §4-7 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.Guardian.(ReductionIssuer)` | 청산 발행기 | 실패 → B3 | AST L326 |
| `exitSideRetrier` | 기록 전용 Retrier | 무오류 | AST L332 |
| `c.NormalAlertRelay`·`obs.RecordOnly` | 기록 전용 알림 | — | AST L331 |
| `exitSideFloor` | RECONCILE 상한 | — | AST L362 |
| `NewExitObserver` | 최종 검증 | 그 오류 반환 | AST L364 |
| live config | `Engine.ExitPolicy.CommonPolicy` · `Notifications.Enabled` · `ExitSourceAgeReport`(기동 1회) | 재읽기 없음 | AST 대입 |

## State mutations and fallbacks

- `opts` 값 복사본의 필드를 덮음. `opts.Log` 는 덮지 않음 — 그래서 운영 관측기의 `warnEpisode`·`log` 는 무기록(design D4 에 영향, claim-check D4).

## Safety conclusion

- Safe edit boundary: B3 직후 단언·대입만. 기존 대입·덮기 순서 무편집.
- High-risk impact: yes — 손절 루프의 유일한 운영 조립 자리. 단언 실패는 기동 거부(손절 루프 전체가 안 뜸)이므로, 운영 Guardian 이 새 인터페이스를 반드시 구현해야 함 — 조립 시험(a198 §4-7)이 그것을 고정해야 함.
- a087 겹침: a087 의 current 번들이 이 함수의 지금 sha 를 고정 — 이 함수를 편집하면 a087 번들이 낡음(edit-surface.md §4).
