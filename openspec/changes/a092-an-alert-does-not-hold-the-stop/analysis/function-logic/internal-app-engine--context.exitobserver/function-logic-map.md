# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go`
- AST evidence: `ast.json` — **편집 전**, :319–348, 분기 6 · 반환 4 · 호출 3, source_sha256 `43dcca8f37b7…`, 추출 HEAD `8c390aa6`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(22판 C1 · design D0.3g 1): exit goroutine에서 닿는 알림 입구를 **주입 지점별로** 기록 전용에 묶는다.
  - `Alerts` 기본값 → 기록 전용 어댑터.
  - `Announcer` 기본값 → 기록 전용 announcer.
  - `Retrier` → `Announcer`만 바꾼 값 복사본.
  - `Floor` → retrier만 그 복사본으로 바꾼 floor 복사본.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | nil 거절 | 엔진 조립 | B1 → `ErrExitObserverUnavailable` |
| `c.Automation.Verified` | true여야 함 | 자동화 게이트 | B2 → 거절 |
| `c.Guardian` | `ReductionIssuer`여야 함 | Guardian 조립 | B3 → 거절 |
| `c.Retrier` | 공유 포인터(대사 · tracer · 운영 명령 · 편입과 같음) | `buildGateway` :324 | 편집 뒤 복사본만 exit에 준다 |
| `c.Notifier` | nil 허용 | 조립 | B5 — nil이면 `Alerts` 미설정 |
| `c.exitFloor` | nil 허용(시험) | `buildGateway` :350-353 | B6 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c == nil` (:320) | — | `ErrExitObserverUnavailable` (:321) | (미실행 — 측정) |
| B2 | `!c.Automation.Verified` (:323) | — | 거절 (:324) | `TestTheExitObserverIsUnavailableWithoutAVerifiedGate` |
| B3 | `c.Guardian`가 `ReductionIssuer`가 아님 (:327) | — | 거절 (:328) | (미실행 — 측정) |
| B4 | `opts.Names == nil` (:338) | `opts.Names = c.Names` | — | `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` |
| B5 | `opts.Alerts == nil && c.Notifier != nil` (:341) | `opts.Alerts = c.Notifier` — **동기 `Notify`** | — | 같음 |
| B6 | `opts.Floor == nil` (:344) | `opts.Floor = c.exitFloor` — **공유 Retrier를 든 floor** | — | 같음 |
| 종단 | — | `NewExitObserver(opts)` (:347) | 그 반환 | 같음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewExitObserver(opts)` :347 | 관측기 생성 · 옵션 검증 | 옵션 누락은 오류 | AST |
| (대입) `opts.Retrier = c.Retrier` :333 | 가격 조회 재시도 — 401/403이면 `escalateCredentialFailure` → `r.Announcer`(동기 `Notify`) | 공유 포인터 | AST · `retry.go:357-365,413-414` |

## State mutations and fallbacks

- `opts`(값)만 채운다. `c`의 필드는 바꾸지 않는다 — 공유 Retrier · floor를 **복사**해서 exit에만 준다(편집 뒤).
- 조립 바깥에서 `Announcer`를 넘기는 호출자(`cmd/tossctl/engine.go` `engineRuntime` :634-640 — `Announcer: ectx.Notifier`)는 편집 뒤 그 값을 넘기지 않는다. 기본값이 기록 전용이다.

## Safety conclusion

- Safe edit boundary: 분기 B1~B3(거절)과 `NewExitObserver` 호출은 그대로다. 대입 넷(Alerts · Announcer · Retrier · Floor)만 기록 전용 쪽으로 바꾼다.
  공유 Retrier의 다른 소비자(대사 `reconcileloop.go:360` 등)는 바뀌지 않는다 — Q1 문자 해석.
- High-risk impact: yes — 알림 · 모드 통지 입구. 손절 판정 · 청산 경로는 건드리지 않는다.
