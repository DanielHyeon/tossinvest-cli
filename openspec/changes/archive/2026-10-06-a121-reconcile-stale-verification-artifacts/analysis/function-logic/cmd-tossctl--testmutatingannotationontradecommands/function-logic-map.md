# Function Logic Map: `TestMutatingAnnotationOnTradeCommands`

- Source: `cmd/tossctl/help_convention_test.go` (96-160)
- Qualified function: `TestMutatingAnnotationOnTradeCommands` (시험 함수)
- Revision: `current` (GREEN 편집 뒤, `source_sha256` a1540d91…)
- AST evidence: `ast.json` — AST branches 3, 반환 0
- Risk scan: `risk-pattern-report.md`
- 편집: 고정 집합 `wantMutating` 에 `"tossctl verify reconcile": true` 한 줄과 주석 2줄(Manager 허용 — design 「로트 1 처분」 S2).
  비례 원칙상 시험 전용 편집이지만 게이트 5단계가 수정된 기존 함수로 세므로 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `wantMutating` | 잎 명령 경로 → true | 이 시험의 고정 집합 | 집합 밖 mutating=true 또는 집합 안 미표지면 `t.Errorf` |
| `leafCommands(newRootCmd())` | 등록된 잎 명령 전부 | `newRootCmd` | — |

불변식: mutating=true 인 잎 명령의 집합이 이 고정 집합과 **정확히** 같다.

## Branches and early returns

| Branch | AST kind | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range at 150:2` | 잎 명령 순회 | — | 없음 | 이 시험 자체 |
| B2 | `if at 153:3` | 집합 안인데 표지 없음 | `t.Errorf` | 없음 | 변이 C-annotation(원장) |
| B3 | `if at 156:3` | 집합 밖인데 mutating=true | `t.Errorf` | 없음 | 변이 C-register 의 짝(등록 누락 시 B2 미발화 — 다른 시험이 잡음) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `leafCommands` · `newRootCmd` | 명령 트리 | 순수 | AST |
| `c.CommandPath` · `t.Errorf` | 경로·실패 보고 | 없음 | AST |

## State mutations and fallbacks

- 없음(시험).

## Safety conclusion

- Safe edit boundary: 집합에 한 줄 추가뿐 — 다른 명령의 판정 불변.
- High-risk impact: no(시험). 이 집합이 `verify reconcile` 의 자동 실행 금지 표지를 고정한다.
