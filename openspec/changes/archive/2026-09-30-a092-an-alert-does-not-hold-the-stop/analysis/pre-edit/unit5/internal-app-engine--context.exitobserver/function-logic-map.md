# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go`
- AST evidence: `ast.json` — **편집 뒤**, :319–356, 분기 7 · 반환 4 · 호출 5, source_sha256 `d0a5a24ce1a2…`, 추출 커밋 `c6e2e3ac`.
  편집 전 번들(:319–348, 분기 6, `43dcca8f37b7…`, `8c390aa6`)은 `analysis/pre-edit/unit2/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집(22판 C1, 착지 단위 ②): exit goroutine에서 닿는 알림 입구를 **주입 지점별로** 기록 전용에 묶음.
  도우미 둘(`exitSideRetrier` · `exitSideFloor`)은 새 파일 `internal/app/engine/exit_record_only.go`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | nil 거절 | 엔진 조립 | B1 → `ErrExitObserverUnavailable` |
| `c.Automation.Verified` | true여야 함 | 자동화 게이트 | B2 → 거절 |
| `c.Guardian` | `ReductionIssuer`여야 함 | Guardian 조립 | B3 → 거절 |
| `c.Retrier` | 공유 포인터(대사 · tracer · 운영 명령 · 편입과 같음) | `buildGateway` | 값 복사본만 exit에 준다 — 공유본 불변 |
| `c.Notifier` | nil 허용 | 조립 | nil이면 `RecordOnly{N:nil}` → 무동작, `Alerts`/`Announcer` 기본값 미설정(B5 · B6) |
| `c.exitFloor` | nil 허용(시험) | `buildGateway` | nil이면 오늘처럼 그 nil을 그대로(B7 → `exitSideFloor` nil 갈래) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c == nil` (:320) | — | `ErrExitObserverUnavailable` | (미실행 — 측정) |
| B2 | `!c.Automation.Verified` (:323) | — | 거절 | `TestTheExitObserverIsUnavailableWithoutAVerifiedGate` |
| B3 | `c.Guardian`가 `ReductionIssuer`가 아님 (:327) | — | 거절 | (미실행 — 측정) |
| B4 | `opts.Names == nil` (:343) | `opts.Names = c.Names` | — | `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` |
| B5 | `opts.Alerts == nil && c.Notifier != nil` (:346) | `opts.Alerts = recordOnly` — **기록 전용** | — | `TestA092ExitObserverGetsRecordOnlyAlertPaths` |
| B6 | `opts.Announcer == nil && c.Notifier != nil` (:349) — **새 분기** | `opts.Announcer = recordOnly` — 관측 두절 강화 통지 | — | 같음 |
| B7 | `opts.Floor == nil` (:352) | `opts.Floor = exitSideFloor(c.exitFloor, exitRetrier)` — retrier만 exit 복사본 | — | 같음 |
| 종단 | — | `NewExitObserver(opts)` (:355) | 그 반환 | 같음 |

무조건 대입: `recordOnly := obs.RecordOnly{N: c.Notifier}` (:333) · `exitRetrier := exitSideRetrier(c.Retrier, recordOnly)` (:334) · `opts.Retrier = exitRetrier`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitSideRetrier` :334 | 공유 Retrier의 값 복사본 + `Announcer`만 기록 전용 — 401/403 강화 통지(`retry.go` `escalateCredentialFailure`)가 exit에서 기록만 | nil 공유본이면 nil | AST · `exit_record_only.go` |
| `exitSideFloor` :353 | floor 복사본이 exit Retrier로 조회(`ConfirmedFloor` → `f.retrier.Query`) | nil 공유본이면 nil | AST |
| `NewExitObserver(opts)` :355 | 관측기 생성 · 옵션 검증 | 옵션 누락은 오류 | AST |

## State mutations and fallbacks

- `opts`(값)만 채운다. `c`의 필드는 바꾸지 않는다 — 공유 Retrier · floor를 **복사**한다. `TestA092ExitObserverGetsRecordOnlyAlertPaths`가 공유본 불변(Announcer 가 여전히 `*obs.Notifier`)을 단언하고,
  변이 U24(복사 대신 공유본 변경)가 잡힌다.
- `Retrier`가 설정값만 가진 구조체라 복사가 안전하다(상태는 공유 `Gate` 포인터). 복사본과 공유본은 Announcer 외 모든 필드가 같다(시험이 필드별로 대조).
- `engineRuntime`은 더 이상 `Announcer`를 넘기지 않는다(`TestA092EngineRuntimeDoesNotHandTheExitLoopASyncAlertPath` 구조 핀).

## Safety conclusion

- Safe edit boundary: 분기 B1~B3(거절)과 `NewExitObserver` 호출은 그대로. 분기 하나 추가(B6), 대입 넷이 기록 전용 쪽으로.
  공유 Retrier의 다른 소비자(대사 · 운영 명령 · 편입)는 동기 통지 그대로 — Q1 문자 해석.
- High-risk impact: yes — 알림 · 모드 통지 입구. 손절 판정 · 청산 주문 경로는 건드리지 않음(Retrier 정책 · Gate · 공급자 값 동일).
