# Function Logic Map: `NewPositionPolicyCommandService`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json` — **편집 전**, :90–104, 분기 2 · 반환 2 · 호출 3, source_sha256 `fbb9d29d47f9…`, 추출 HEAD `81934b46`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(a092 25.6): 서비스 리터럴에 통지 기록자(`notices`)를 더한다 — 엔진 알림기(`ectx.Notifier`)가 있으면 그것, 없으면 nil 인터페이스
  (타입 있는 nil 포인터를 인터페이스에 넣지 않는다 — nil 판정이 거짓이 되므로).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` · `ectx.Journal` | nil 거절 | 엔진 조립 | B1 → 오류 |
| `clk` | nil 이면 시스템 시계 | 호출자 | B2 |
| `ectx.Notifier` | nil 허용 | `newNotifier` | 편집 뒤 nil 이면 `notices` nil → 통지 시 `Notified=false` + 사유 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ectx == nil || ectx.Journal == nil` (:91) | — | 오류 (:92) | `TestPositionPolicyCommandServiceRequiresEngineOwnedJournal` |
| B2 | `clk == nil` (:94) | `clk = clock.System()` | — | (미실행) |
| 종단 | — | 서비스 리터럴 (:97) | 서비스, nil | `TestA066*` 전부 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New` · `clock.System` · `strings.TrimSpace` | 오류 · 기본 시계 · 정책 이름 | 순수 | AST |

## State mutations and fallbacks

- 새 서비스 값만 만든다.

## Safety conclusion

- Safe edit boundary: 분기 둘 불변, 리터럴 필드 하나 추가(편집 뒤 분기 하나가 늘 수 있음 — nil 알림기 판정).
- High-risk impact: 낮음(조립) — 통지 기록자 배선.
