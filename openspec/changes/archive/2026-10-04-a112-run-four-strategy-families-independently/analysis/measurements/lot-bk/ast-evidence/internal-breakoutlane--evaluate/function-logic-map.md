# Function Logic Map (편집 전): `Evaluate`

- Source: `internal/breakoutlane/machine.go`
- Source SHA-256: `3f2e013019481047008c2c7fbc6e365bab27467f4554f3e53ec0baef5e2488dc`
- Signature: `Evaluate(params=2, results=1)`
- Source range: `7:1`–`34:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 3be54204 — 편집 없음(BK2 는 시험만)).

## Inputs and invariants

- 편집 계획: 편집 없음 — breakout 덮개 2차 시험 · census 가 근거로 쓰는 분기 열거

## Branches and early returns

- Exact AST return nodes: `10:3`, `15:4`, `19:5`, `22:5`, `29:5`, `33:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 9:2 | `if !validSnapshot(snapshot) {` |
| B2 | if | 13:2 | `if prior != nil {` |
| B3 | if | 14:3 | `if !validDecision(*prior) {` |
| B4 | if | 17:3 | `if prior.setupID == setup {` |
| B5 | if | 18:4 | `if prior.snapshotDigest == snapshot.digest {` |
| B6 | if | 21:4 | `if !validCorrectionLineage(snapshot, *prior) {` |
| B7 | if | 24:4 | `if prior.phase == phaseProposed // terminalPhase(prior.phase) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validSnapshot` | 9:6 |
| `refused` | 10:10 |
| `setupID` | 12:11 |
| `validDecision` | 14:7 |
| `refused` | 15:11 |
| `validCorrectionLineage` | 21:8 |
| `refused` | 22:12 |
| `terminalPhase` | 24:39 |
| `append` | 26:25 |
| `(unnamed)` | 26:32 |
| `decisionSeal` | 28:22 |
| `evaluateFresh` | 33:9 |

## Safety conclusion

- 생산 코드 변경 0 — 근거 열거만
