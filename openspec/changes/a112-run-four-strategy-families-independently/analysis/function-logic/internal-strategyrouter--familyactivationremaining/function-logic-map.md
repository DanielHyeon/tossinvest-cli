# Function Logic Map: `familyActivationRemaining`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `7842842bc1b1bf9d2be542624169502fcc846a4e60080841887f5072f4fb6507`
- Signature: `familyActivationRemaining(params=2, results=2)`
- Source range: `384:1`–`390:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 이 패키지의 유일한 만료 판정 — 적재와 lease 상한이 같이 부른다(판정 불변).

## Branches and early returns

- Exact AST return nodes: `386:3, 389:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 385:2 | 만료(지금이 expires_at 이상) → `%w: expires_at … is not after …` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `now.Before` | 385:6 |
| `fmt.Errorf` | 386:13 |
| `Format` | 387:4 |
| `expires.UTC` | 387:4 |
| `Format` | 387:44 |
| `now.UTC` | 387:44 |
| `expires.Sub` | 389:9 |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.
