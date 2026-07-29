# Function Logic Map: `acceptSupervised`

- Source: `internal/soak/attest.go`
- Function: `internal/soak/attest.go:acceptSupervised`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`

이 change가 추가한 leaf. 감독 검증 증거 중 attestation에 실을 수 있는 것을 고른다. 허용 집합은 `LiveOnlyEndpoints()`로 닫혀 있고, 계좌 불일치는 거부, 창 밖은 건너례다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| accountRef string | soak의 계좌(비마스킹) | soak 기록 | 빈 값이면 `SameAccountMasked`가 false → 거부 |
| now / validity | 발급 시각과 유효 기간 | 호출자 / Criteria | 창 밖·미래는 건너뜀 |
| supervised []attest.Proof | nil 허용 | 검증 기록 | nil이면 빈 결과 — 읽기만 담아 발급 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` (line 259) — `for _, e := range LiveOnlyEndpoints() {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `range` (line 265) — `for _, p := range supervised {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `if` (line 267) — `if endpoint == "" {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `if` (line 271) — `if !allowed[key] {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `if` (line 278) — `if !attest.SameAccountMasked(accountRef, p.AccountRef) {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B6 | `if` (line 286) — `if age < 0 \|\| age >= validity {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B7 | `if` (line 289) — `if seen[key] {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `LiveOnlyEndpoints` | ast.json calls (line 259) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `normaliseEndpoint` | ast.json calls (line 260) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Join` | ast.json calls (line 266) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Fields` | ast.json calls (line 266) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.TrimSpace` | ast.json calls (line 266) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `fmt.Errorf` | ast.json calls (line 272) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attest.SameAccountMasked` | ast.json calls (line 278) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `attest.Mask` | ast.json calls (line 283) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `now.Sub` | ast.json calls (line 285) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `append` | ast.json calls (line 294) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 지역 slice·map만 쓴다. 입력 slice를 변형하지 않고 Proof 복사본의 철자만 정규화해 append한다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 거부가 기본 방향이고, 허용은 세 조건을 모두 통과해야 한다.
- High-risk impact: yes — 게이트 5절이 읽는 집합을 결정한다. 단, 9절 상수가 기동 자체를 막는다.
