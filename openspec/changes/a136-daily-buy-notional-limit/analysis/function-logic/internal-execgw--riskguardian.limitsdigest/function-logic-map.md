# Function Logic Map: `LimitsDigest`

- Source: `internal/execgw/riskguardian.go` (321-327, 편집 뒤)
- Qualified: `RiskGuardian.LimitsDigest`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 88cf7c183018aae8…, 편집 전 분기 수 0) — 현재 branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 바인딩이 결속하는 Guardian 한도 digest.

**a136 편집.** freeze 리뷰 P2-4 수리: digest 입력에 하루 매수 상한을 더함(`limitsJSON + "\ndaily_buy=" + amount + currency`) — 이 한도를 바꿔도 다른 다섯처럼 바인딩이 무효화됨. 생산 바인딩은 Guardian 에서 자기 유도라 재서명 불요.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `g.limitsJSON` · `g.policy.MaxDailyBuyNotional` | — | 생성 시 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `TestA136TheLimitsDigestBindsTheDailyBuyCeiling` |

Exact AST return positions: 326

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sha256.Sum256` | digest | — | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: digest 입력 확장.
- High-risk impact: yes(바인딩 결속)
