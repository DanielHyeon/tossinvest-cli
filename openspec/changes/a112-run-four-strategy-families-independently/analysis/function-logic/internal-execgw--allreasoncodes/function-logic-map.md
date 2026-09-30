# Function Logic Map: `AllReasonCodes`

- Source: `internal/execgw/failclosed.go`
- AST evidence: `ast.json` — **편집 뒤**, :254–311, 분기 0, source_sha256 `fbe97b79db74…`.
- Risk scan: `risk-pattern-report.md`
- 편집: (5.6.2.1, 커밋 `3260f4eb`) 열거에 `ReasonStrategyCentralIntegrity` 한 줄 — 분기 없음. 골든 재생성 · a098 census 한 줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 편집이 더한 입력 | 진입 게이트(`*execgw.EntryGate`) 또는 없음 | 엔진 조립(`Context.Entry`) | 위 편집 설명 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path(분기 없음) | 열거 | 정렬된 목록 | `TestReasonCodeEnumIsStable` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `sort.Slice` | 309:2 |

## State mutations and fallbacks

- 위 편집 설명 외 없음.

## Safety conclusion

- Safe edit boundary: 어휘 추가(이름 바꾸기 아님) — 기존 기록 불변.
- High-risk impact: yes(진입 게이트 경로) — 편집은 진입을 닫는 방향만 더함.
