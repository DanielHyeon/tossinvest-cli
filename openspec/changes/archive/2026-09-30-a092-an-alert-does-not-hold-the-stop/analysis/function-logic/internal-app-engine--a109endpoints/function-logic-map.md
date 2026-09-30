# Function Logic Map: `a109Endpoints`

- Source: `internal/app/engine/a109_the_sibling_endpoints_recover_too_test.go`
- AST evidence: `ast.json` — **편집 뒤**, :55–91, 분기 0, source_sha256 `4ed962de949a…`, 추출 커밋 `2714e393`.
- Risk scan: `risk-pattern-report.md`
- 편집: 형제 엔드포인트 표에 모드 제어 엔드포인트를 더함 — 회수 · 권한 · 낯선 엔트리 거부 · staging 잔재 회수 시험 여덟이 모드 제어에도 돈다(`-v` 로 `/mode_control` 부분시험 확인).
- 비례 원칙: 시험 코드 → 변이 원장은 생산 변이(`--set 25.9`)가 반증 도구로 씀. `not-applicable`: 이 함수 자체의 다중 리뷰.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | 시험 | go test | `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 분기 없음 | — | 표 반환 | 호출 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST 의 호출 목록 | 픽스처 · 단언 | 시험 | AST |

## State mutations and fallbacks

- 임시 자원만.

## Safety conclusion

- Safe edit boundary: 표 · 단언 편집.
- High-risk impact: no — 시험 코드.
