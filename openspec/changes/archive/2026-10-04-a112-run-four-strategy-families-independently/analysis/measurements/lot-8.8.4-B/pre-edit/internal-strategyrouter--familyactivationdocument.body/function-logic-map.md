# Function Logic Map (편집 전): `body`

- Source: `internal/strategyrouter/production_family_activation.go`
- Source SHA-256: `c8e2efb9b1a8243bcec000f0c2fa9e96bd8576c97aa35a694a1e6c053407288f`
- Signature: `FamilyActivationDocument.body(params=0, results=2)`
- Source range: `225:1`–`265:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD f473d815).

## Inputs and invariants

- 편집 계획: a112 8.8.4 로트 B 항목 1(Q-B1=(c)): 맨 sentinel 반환마다 필드명 %w 래핑(errors.Is 보존). 복합 결속은 분기 그대로 필드-diff 목록으로 한 번 반환. :455 읽기 실패는 underlying err 를 %w 사슬로 보존하고 digest 불일치만 필드명(두 종류 분리).

## Branches and early returns

- Exact AST return nodes: `228:3`, `233:4`, `250:48`, `251:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 227:2 | `if len(want) == 0 {` |
| B2 | range | 231:2 | `for _, family := range document.On {` |
| B3 | if | 232:3 | `if !family.Known() {` |
| B4 | range | 238:2 | `for laneID, table := range want {` |
| B5 | if | 240:3 | `if on[table.Family] {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteDescriptors` | 226:10 |
| `len` | 227:5 |
| `family.Known` | 232:7 |
| `make` | 237:17 |
| `len` | 237:65 |
| `append` | 243:17 |
| `sort.Slice` | 250:2 |
| `Format` | 259:33 |
| `document.ApprovedAt.UTC` | 259:33 |
| `Format` | 260:33 |
| `document.IssuedAt.UTC` | 260:33 |
| `Format` | 261:33 |
| `document.ExpiresAt.UTC` | 261:33 |

## Safety conclusion

- 판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is). 메시지에 필드명이 더해질 뿐. 미선언(Undeclared)이 맨 앞인 순서 불변.
