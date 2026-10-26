# Function Logic Map: `candidateFilterMatrixFromMarkets`

- Source: `internal/console/optimization_matrix_view.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시장별 후보 필터 registry | KR·US와 각 시장의 key·descriptor | `candidate.CandidateFilterMarkets()` | 누락된 시장/key도 화면에서 숨기지 않음 |
| descriptor 값 | registry 원문 값, mutation 금지 | `candidate.CandidateFilterDescriptor` | 행렬은 읽기 전용 projection만 만듦 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 시장 반복(94:2) | 시장별 key map·세션 구성 | 없음 | `TestCandidateFilterMatrixPreservesMarketOnlyKeys` |
| B2 | descriptor 반복(96:3) | key→descriptor 저장 | 없음 | 같은 시험 |
| B3 | 처음 본 key(98:4) | 시장 순서대로 합집합 행 순서 확정 | 없음 | 같은 시험 |
| B4 | 합집합 key 반복(108:2) | KR·US cell projection 생성 | 없음 | 같은 시험 |
| B5 | KR 미등록(115:3) | KR 시장·세션만 채운 unavailable cell 생성 | 없음 | 같은 시험 |
| B6 | US 미등록(118:3) | US 시장·세션만 채운 unavailable cell 생성 | 없음 | 같은 시험 |
| B7 | KR 미등록(122:3) | US descriptor label을 행 label로 사용 | 없음 | 같은 시험 |
| B8 | 두 시장 등록·모든 field 동일(127:3) | 공통 cell label만 변경 | 없음 | `TestCandidateFilterMatrixMergesEqualMarketDescriptors` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidateFilterMatrixCellFrom` | descriptor를 읽기 전용 template cell로 복사 | 외부 I/O 없음 | `ast.json` calls |
| `candidateFilterDescriptorEqual` | 두 시장 descriptor 공통 행 여부를 결정 | 외부 I/O 없음 | `ast.json` calls |

## State mutations and fallbacks

- 새 map·slice만 만듦. registry, handler, 설정, 주문 상태는 mutation하지 않음.

## Safety conclusion

- Safe edit boundary: 시장별 key의 합집합을 순서대로 표시하고, 한 시장에 없는 key를 명시적으로 fail-closed 표시함.
- High-risk impact: no — console projection만 바꾸며 descriptor 값·registry·주문 경로를 바꾸지 않음.
