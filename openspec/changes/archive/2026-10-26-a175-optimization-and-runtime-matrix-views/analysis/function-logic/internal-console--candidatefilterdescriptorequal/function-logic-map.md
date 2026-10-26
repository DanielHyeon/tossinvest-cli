# Function Logic Map: `candidateFilterDescriptorEqual`

- Source: `internal/console/optimization_matrix_view.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| KR·US descriptor | Market 이외 필드를 포함한 registry 원문 | `candidate.CandidateFilterDescriptor` | 차이가 있으면 공통 행 병합 금지 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 분기 없음(146:1–149:2) | 두 사본의 `Market`만 중립화 후 전체 descriptor deep equal | `false` 또는 `true` | `TestCandidateFilterMatrixKeepsCategoryDifferenceSeparate` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reflect.DeepEqual` | Market을 제외한 현재·미래 descriptor field 전체 비교 | 외부 I/O 없음 | `ast.json` |

## State mutations and fallbacks

- 값 전달 사본의 `Market`만 중립화함. 호출자 descriptor는 mutation하지 않음.

## Safety conclusion

- Safe edit boundary: Market만 중립화한 뒤 모든 현재·미래 descriptor field를 비교하도록 축약함.
- High-risk impact: no — 표시 병합 여부만 바꾸며 registry 값을 바꾸지 않음.
