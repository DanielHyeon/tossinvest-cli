# Function Logic Map: `TestAHeldRowIsNotWhispered`

- Source: `internal/obs/a099_round4_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :120–165, 분기 10, source_sha256 `8745c724f33b…`, 추출 커밋 `fbc6df5f`.
- Risk scan: `risk-pattern-report.md`
- 편집: **교차 change(a099) 시험 단언 변경**(Manager 판정 (나)) — WARN 단언 → INFO 단언 + 보유자 이름 단언 추가. 줄 실재(`line == nil`) · 만료(`claim_expires_at`) 단언은 그대로.
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(L-set)가 이 시험을 반증 도구로 씀. 단언 불약화 표는 `review.md` §24.7.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at :131 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B2 | if at :134 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B3 | else at :136 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B4 | if at :136 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B5 | if at :140 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B6 | if at :143 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B7 | if at :148 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B8 | if at :154 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B9 | if at :158 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B10 | if at :161 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 원장만.

## Safety conclusion

- Safe edit boundary: 주제 보존, 단언 모양 변경(§24.7 표).
- High-risk impact: no — 시험 코드.
