# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go` (**편집 전** `319`–`365`, base `064f59fd9` → 재고정 `b93ed133a`)
- Qualified: `Context.ExitObserver`
- AST evidence: `ast.json` (편집 전 `source_sha256` 7001a47267ea…) — branches 6 · returns 4
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(a197 tasks 1.6). **2026-11-25 재추출: 본문 변경 1곳** — a199(`3445403ee`)가 B2 조건을 `Verified` → `RunsExitSide()` 로 바꿈(분기 수 · 좌표 무변). 귀결: 이 관측기는 청산 전용 기동에서도 만들어진다 — a197 단계도 그 기동에서 돈다. 편집 여부는 조건부 — 브로커 미체결 판독기(와, Manager 결정에 따라 취소 결정 발행기)를 주입하는 자리 후보다.

**역할.** 생산 exit 관측 루프의 배선. 호출자가 넘긴 옵션 위에 Context 의 원장 · 가격 · Retrier · Guardian · 게이트웨이 · 알림기 · 설정 값을 **호출자 값과 무관하게** 덮어쓴 뒤 `NewExitObserver` 를 부른다. 생산 호출자는 `cmd/tossctl/engine.go:671` `engineRuntime` 하나(CodeGraph 1.6.0 `callers ExitObserver` 비시험: `engineRuntime` · `NewExitObserver` · `ExitObserver` 자신 — 뒤 둘은 동명 이름 충돌).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | 조립된 엔진 Context | `NewContext` | nil → `ErrExitObserverUnavailable` |
| `c.Automation.RunsExitSide()` | true(검증 기동 또는 a199 청산 전용 기동) | 기동 interlock | false → 거절 |
| `c.Guardian` | `ReductionIssuer` 를 만족(생산 = `*execgw.RiskGuardian`) | `guardian_wiring.go:100` | 아니면 거절 |
| `opts` | 호출자 값(시계 · 비용 · SLO · 로거) | `engineRuntime` | — |
| 불변식 | 안전 필드(Retrier · Issuer · Submit · Alerts · Critical · 토글)는 호출자 값과 무관하게 덮는다 | 이 함수 주석 `:347-348` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:320` `if c == nil` | 없음 | `ErrExitObserverUnavailable` | 배선 거절 |
| B2 | `:323` `if !c.Automation.RunsExitSide()` (a199 — 청산 전용 기동도 통과) | 없음 | 거절 | 배선 거절 |
| B3 | `:327` `if !ok` (Guardian 이 `ReductionIssuer` 아님) | 없음 | 거절 | 배선 거절 |
| B4 | `:344` `if opts.Names == nil` | `opts.Names = c.Names` | — | 이름 배선 |
| B5 | `:349` `if c.Notifier != nil` | `Alerts`/`Announcer` = 기록 전용, `Critical` = `c.Notifier` | — | 기록 전용 배선 |
| B6 | `:361` `if opts.Floor == nil` | `opts.Floor = exitSideFloor(…)` | — | 하한 배선 |

무조건 덮는 줄(분기 밖): `Journal` · `Prices` · `Retrier` · `Issuer` · `Submit = c.Gateway` · `lowerLimitBoard` · `AccountRef` · `CommonPolicy` · `NotificationsEnabled` `:356` · `SourceAgeReport = c.Config.Engine.ExitSourceAgeReport` `:358`(a163 토글 선례) · `ZeroFloorLog`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Errorf` | 거절 문장 | — | AST |
| `obs.RecordOnly` (리터럴) · `c.NormalAlertRelay` | 기록 전용 알림기 | — | AST `:333` |
| `exitSideRetrier` | exit 쪽 Retrier | — | AST `:334` |
| `exitSideFloor` | 확정 하한 | — | AST `:362` |
| `NewExitObserver` | 검증 · 생성 | 오류 그대로 반환 | AST `:364` |

Live config binding: `c.Config.Engine.ExitPolicy.CommonPolicy` `:343`, `c.Config.Engine.Notifications.Enabled` `:356`, `c.Config.Engine.ExitSourceAgeReport` `:358` — 전부 **기동 때 로드된 값**(주기 재읽기 없음). 브로커 주문 판독기 주입은 **0**(`ExitObserverOptions` 에 그런 필드 없음, `exitloop.go:171-254`).

## State mutations and fallbacks

- `opts` 값 복사본만 바꾼다. 원장 · 브로커 접촉 없음.

## Safety conclusion

- Safe edit boundary (a197): 무조건 덮는 줄을 **덧붙이기만** — 예: 새 토글 `opts.<Toggle> = c.Config.Engine.<Field>`, 미체결 판독기 `opts.<Reader> = execgw.OfficialOrders{Client: c.Official}`(선례: `engineFillDetector` `cmd/tossctl/engine.go:770`, `SnapshotCollector` `runtime_wiring.go:133`). 기존 줄 · 분기 순서 무변경.
- a166 census: 이 함수와 `exitwiring.go` 는 `shapes.json` 의 고정 함수 90개에 **없다**(파일명 grep 0). 새 줄이 `Do` · `RoundTrip` · `NewClient` · `Budget` · `official.New*` 선택자를 쓰지 않으면 `sites.json` 재고에도 안 걸린다.
- a163 census: `SourceAgeReport` 식별자의 허용 자리 중 하나가 이 함수(`:358`) — 같은 식별자를 새로 쓰지 않으면 무관.
- High-risk impact: yes(배선이 취소 능력을 넘김) — 토글 꺼짐이면 새 판독기는 호출되지 않아야 함(design §4-2).

## 재개 개정(2026-11-25) — 편집 확정

- 편집 확정(design D8 · D4): 무조건 덮는 줄에 (1) 미체결 판독기 `opts.<Reader>`(선례 `execgw.OfficialOrders{Client: c.Official}`), (2) 토글 `opts.<Toggle> = c.Config.Engine.<Field>`(선례 `:358`), (3) 토글 audit — `c.Audit.RecordChange(<새 동작 상수>, …)`(`engine.go:227` 의 `Audit`, `audit.go:225`; `NewContext` 무편집 — r1 #10). (3) 의 실패 처리는 design O-1(Manager 결정 대기) — 결정에 따라 이 함수에 갈래가 하나 늘 수 있다(그때 BTM 행 추가).
- 기존 분기 B1~B6 · 순서 무변경. a166 census 밖(rg). 청산 전용 기동에서도 이 함수가 관측기를 만든다(B2 `RunsExitSide()`) — 단계도 그 기동에서 돈다.
