# Function Logic Map: `withoutSymbol`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `list` | nil 가능 | 호출자의 정규화된 목록 | nil은 빈 슬라이스를 돌려준다 |
| `symbol` | 이미 upper+trim | 호출자 | 정규화되지 않은 값을 주면 일치하지 않는다 — 두 호출자 모두 정규화 후 호출한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range list` | 없음 | 필터된 새 슬라이스 | `TestReleasingAnExclusion` |
| B2 | `s != symbol` | 없음 | 유지 | `TestRemovingADesignationOnlyAffectsTheFuture` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | 호출하지 않는다(순수 함수) | — | CodeGraph + AST |

## State mutations and fallbacks

- 입력 슬라이스를 변형하지 않는다 — 새 슬라이스를 만든다. 호출자가 `current`를 그대로 들고 있어도 오염되지 않는다.
- 중복이 있으면 전부 제거한다. 저장 경로의 `normaliseSymbols`가 중복을 만들지 않으므로 실제로는 최대 1건이다.

## Safety conclusion

- Safe edit boundary: 신규 순수 함수. `handleSettingsInclude`의 인라인 루프를 그대로 옮긴 것이라 기존 동작과 동일하다.
- High-risk impact: no — 부작용이 없다.
