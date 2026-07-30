# Function Logic Map: `verifyGate`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

조항 9(ProtectionReady)의 거부가 제거됐다. 조항 1~8은 무수정.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| status.Protection | WIRED | UNWIRED | execgw.ProfileProtection | 더 이상 이 함수의 반환에 영향을 주지 않는다 |

## Branches and early returns

제거된 분기는 `status.Protection != ProtectionWired` 하나다. 남은 분기는 전부 조항 1~8이며 각각 기존 테스트가 덮는다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/interlock.go:475 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/interlock.go:483 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/app/engine/interlock.go:490 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/app/engine/interlock.go:496 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/app/engine/interlock.go:508 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/app/engine/interlock.go:513 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `if` @ internal/app/engine/interlock.go:519 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `if` @ internal/app/engine/interlock.go:524 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B9 | `if` @ internal/app/engine/interlock.go:527 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (제거) status.Protection 비교 | 조항 9였다 | 해당 없음 | diff |

## State mutations and fallbacks

- 없음 — 이 함수는 status의 AttestationExpiresAt만 쓰고 나머지는 읽는다.

## Safety conclusion

- Safe edit boundary: 마지막 절의 제거. 앞 여덟 절의 순서·조건·오류는 한 글자도 바뀌지 않았다.
- High-risk impact: yes
