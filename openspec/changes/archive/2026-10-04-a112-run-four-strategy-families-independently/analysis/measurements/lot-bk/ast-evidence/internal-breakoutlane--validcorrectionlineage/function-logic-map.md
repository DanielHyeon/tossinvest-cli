# Function Logic Map (편집 전): `validCorrectionLineage`

- Source: `internal/breakoutlane/machine.go`
- Source SHA-256: `3f2e013019481047008c2c7fbc6e365bab27467f4554f3e53ec0baef5e2488dc`
- Signature: `validCorrectionLineage(params=2, results=1)`
- Source range: `169:1`–`185:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 3be54204 — 편집 없음(BK2 는 시험만)).

## Inputs and invariants

- 편집 계획: 편집 없음 — breakout 덮개 2차 시험 · census 가 근거로 쓰는 분기 열거

## Branches and early returns

- Exact AST return nodes: `172:3`, `178:4`, `184:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 171:2 | `if len(now) < len(prior.lineage) {` |
| B2 | range | 175:2 | `for i, old := range prior.lineage {` |
| B3 | if | 177:3 | `if old.sequence != next.sequence // old.id != next.id // old.sessionID != next.sessionID // next.revision < old.revision // next.revision == old.revision && next.contentDigest != old.contentDigest {` |
| B4 | if | 180:3 | `if next.revision > old.revision {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lineageFrom` | 170:9 |
| `len` | 171:5 |
| `len` | 171:16 |
| `len` | 174:13 |
| `len` | 174:24 |

## Safety conclusion

- 생산 코드 변경 0 — 근거 열거만
