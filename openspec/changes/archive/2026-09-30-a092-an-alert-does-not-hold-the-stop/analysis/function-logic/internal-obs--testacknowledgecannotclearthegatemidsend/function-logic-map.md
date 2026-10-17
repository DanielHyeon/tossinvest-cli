# Function Logic Map: `TestAcknowledgeCannotClearTheGateMidSend`

- Source: `internal/obs/a096_one_send_per_condition_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :378–422, 분기 7, source_sha256 `abda8f7e660f…`, 추출 커밋 `fbc6df5f`.
- Risk scan: `risk-pattern-report.md`
- 편집: **교차 change(a096) 시험 재작성** — 옛 단언 「전송 중 승인은 막힌다」는 frozen 24판이 금지한 잠금 형태(전송 위 잠금)와 허용한 선점(전송 중 정착)에 정면으로 반한다. 주제(「셈~해제 사이에 잠금 아래 기록 경로가 끼지 못한다」)는 구조 핀(`a096PinLockedBetween` — Acknowledge 의 셈 · 해제, claim, 기록 전용 입구가 각각 한 n.mu 임계 구역 안)으로 옮겼고, 행동 단언은 「승인이 전송을 기다리지 않는다 · 발송자가 승인을 되돌리지 않는다 · 선점은 래치가 아니다」.
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(L-set)가 이 시험을 반증 도구로 씀. 단언 불약화 표는 `review.md` §24.7.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | select at :385 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B2 | select at :394 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B3 | if at :396 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B4 | if at :402 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B5 | if at :407 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B6 | if at :410 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |
| B7 | if at :413 — 단언 갈래 | `t.Error`/`t.Fatal` 류 | — | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 원장만.

## Safety conclusion

- Safe edit boundary: 주제 보존, 단언 모양 변경(§24.7 표).
- High-risk impact: no — 시험 코드.
