# Function Logic Map: `NewPositionPolicyCommandService`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json` — **편집 뒤**, :93–112, 분기 3 · 반환 2 · 호출 3, source_sha256 `ddc4d6776125…`, 추출 커밋 `0e4f26af`. 편집 전 번들은 `analysis/pre-edit/25.6/`에 보존(:90–104, 분기 2).
- Risk scan: `risk-pattern-report.md`
- 편집(a092 25.6): 서비스에 통지 기록자 `notices` — 엔진 알림기가 있을 때만(새 분기 B3). 타입 있는 nil 포인터를 인터페이스에 넣지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` · `ectx.Journal` | nil 거절 | 엔진 조립 | B1 → 오류 |
| `clk` | nil 이면 시스템 시계 | 호출자 | B2 |
| `ectx.Notifier` | nil 허용 | `newNotifier`(생산 조립은 항상 있음) | B3 — 있을 때만 `notices` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ectx == nil || ectx.Journal == nil` (:94) | — | 오류 | `TestPositionPolicyCommandServiceRequiresEngineOwnedJournal` |
| B2 | `clk == nil` (:97) | `clk = clock.System()` | — | (미실행) |
| B3 | `ectx.Notifier != nil` (:108) — **새 분기** | `service.notices = ectx.Notifier` | — | `TestA092CommandServiceTakesTheEngineNotifier`(있음 · 없음 둘 다) |
| 종단 | — | — | 서비스, nil | `TestA066*` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New` · `clock.System` · `strings.TrimSpace` | 오류 · 기본 시계 · 정책 이름 | 순수 | AST |

## State mutations and fallbacks

- 새 서비스 값만 만든다. 알림기가 없으면 `notices` 는 nil 인터페이스 → 해제 결과가 「통지되지 않음」(거짓 「통지됨」 없음). 변이 R08(타입 있는 nil) · R09(누락) CAUGHT.

## Safety conclusion

- Safe edit boundary: B1 · B2 불변, B3 추가.
- High-risk impact: 낮음(조립) — 통지 기록자 배선.
