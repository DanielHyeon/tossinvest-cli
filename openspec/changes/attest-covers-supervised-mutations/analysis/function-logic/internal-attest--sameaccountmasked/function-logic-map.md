# Function Logic Map: `SameAccountMasked`

- Source: `internal/attest/attest.go`
- Function: `internal/attest/attest.go:SameAccountMasked`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`

이 change가 추가한 leaf. 전체로 쓴 계좌와 마스킹된 계좌가 같은 것을 가리키는지 답한다. soak 기록은 비마스킹, 검증 기록은 항상 마스킹이라 둘을 결속하려면 이 비교가 필요하고, 한 곳에만 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| full string | 비마스킹 계좌 | soak 기록 | 빈 값이면 false(fail-closed) |
| other string | 마스킹될 수 있는 계좌 | 검증 기록 | 빈 값이면 false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 149) — `if strings.TrimSpace(full) == "" \|\| strings.TrimSpace(other) == "" {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | ast.json calls (line 149) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `sameAccount` | ast.json calls (line 152) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `Mask` | ast.json calls (line 152) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 없음 — 순수 함수.

## Safety conclusion

- Safe edit boundary: 신규 함수. 빈 값 둘 다 false가 기본이다.
- High-risk impact: yes — 계좌 결속이 틀리면 다른 계좌의 증거로 이 계좌의 게이트가 채워진다.
