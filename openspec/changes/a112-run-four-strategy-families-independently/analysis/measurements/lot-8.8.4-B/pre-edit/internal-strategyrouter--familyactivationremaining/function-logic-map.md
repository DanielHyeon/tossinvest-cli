# Function Logic Map (편집 전): `familyActivationRemaining`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`
- Signature: `familyActivationRemaining(params=2, results=2)`
- Source range: `383:1`–`388:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD f473d815).

## Inputs and invariants

- 편집 계획: a112 8.8.4 로트 B 항목 1(Q-B1=(c)): 맨 sentinel 반환마다 필드명 %w 래핑(errors.Is 보존). 복합 결속은 분기 그대로 필드-diff 목록으로 한 번 반환. :455 읽기 실패는 underlying err 를 %w 사슬로 보존하고 digest 불일치만 필드명(두 종류 분리).

## Branches and early returns

- Exact AST return nodes: `385:3`, `387:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 384:2 | `if !now.Before(expires) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `now.Before` | 384:6 |
| `expires.Sub` | 387:9 |

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is). 메시지에 필드명이 더해질 뿐. 미선언(Undeclared)이 맨 앞인 순서 불변.
