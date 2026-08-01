# Function Logic Map: `TestAVerificationInProgressSuspendsTheRefresh`

- Source: `internal/console/portfolio_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and deterministic fixture/read state | values accepted by the typed function signature | current source plus OpenSpec a043 | tests fail explicitly; production reads degrade to typed unknown/unlinked evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if branch at source line 601 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B2 | if branch at source line 609 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B3 | if branch at source line 612 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B4 | if branch at source line 615 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B5 | if branch at source line 627 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B6 | if branch at source line 630 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B7 | if branch at source line 633 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B8 | if branch at source line 645 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B9 | if branch at source line 657 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |
| B10 | if branch at source line 660 | bounded test/read-model control flow only | TestAVerificationInProgressSuspendsTheRefresh coverage and focused package suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| typed callees listed in `ast.json` | Verification-held broker refresh regression assertion now checks typed unknown snapshot evidence. | no retry is introduced; read errors and assertions preserve their existing fail-closed behavior | current AST and focused tests |

## State mutations and fallbacks

- Verification-held broker refresh regression assertion now checks typed unknown snapshot evidence.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
