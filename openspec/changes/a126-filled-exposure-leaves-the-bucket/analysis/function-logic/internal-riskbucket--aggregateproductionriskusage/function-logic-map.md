# Function Logic Map: `aggregateProductionRiskUsage`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`475`–`502`)
- Qualified: `aggregateProductionRiskUsage`
- AST evidence: `ast.json` (`source_sha256` 5f26bf28bbf8207f…) — **편집 전**(freeze census 와 sha · 분기 일치)
- Risk scan: `risk-pattern-report.md`
- AST branches 3 · return 3

**역할.** 예약 행을 검증하고 filled · held 를 합하고 latch 를 모으며 RowDigest 를 만든다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rows` | `readProductionRiskUsage` 결과 | 원장 | 행 계약 위반은 B2 → `ErrJournalUsageInvalid` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:479` `for _, row := range rows {` | 합 · latch · parts 누적 | — | 진입 |
| B2 | `:482` 금액 파싱 · 음수 · 256bit · 상태 enum · RELEASED held≠0 · snapshot/policy 빈 값 · 식별자 | — | `:486` `ErrJournalUsageInvalid` | 진입(기존 시험) |
| B3 | `:494` 합이 256bit 초과 | — | `:495` overflow `ErrJournalUsageInvalid` | 미진입 — 기존 |

## Calls and live bindings

`big.Int.SetString` · `canonicalIdentity` · `productionRiskDigest`. 브로커 · 원장 호출 없음(순수 계산). 결과는 `JournalBucketUsage`.

## State mutations and fallbacks

없다 — 순수. latch 는 합에서 빼지 않고 플래그로만 알린다(`:488` 주석).

## Safety conclusion

- **Safe edit boundary (a126 D1 · D2)**: 떠남 규칙의 **유일한** 적용 자리. 영수증 있는 행의 손상(HELD · held≠0) · released_at 불일치 · owner 키 사본
  불일치를 scope latch 판정 **앞**에서 거절하고, 떠난 행은 합에서만 빼며 latch 플래그 · 검증 · RowDigest parts 형식은 그대로 둔다.
- **High-risk impact**: yes — 감소는 진입을 여는 방향이다.
