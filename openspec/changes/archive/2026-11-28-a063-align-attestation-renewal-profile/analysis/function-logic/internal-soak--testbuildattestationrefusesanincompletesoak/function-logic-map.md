# Function Logic Map: `TestBuildAttestationRefusesAnIncompleteSoak`

- Source: `internal/soak/attest_test.go`; E-based current S snapshot, SHA-256 `7a68e756e4aa71e2c09054e1611b4c0c61abfa4d51c869a00a94b15db5e4dddc` (retrospective, not pre-edit evidence).
- AST evidence: `ast.json`.

## Inputs and invariants

The fixture summarizes only the first two of `threeCleanDays`, evaluates at day one with `criteria()`, and expects `BuildAttestation` to refuse. It is an in-process test fixture: it has no profile path, broker, timeout, retry, or runtime-gate binding.

## Branches and early returns

| B-id | Source condition and effect | Test mapping |
|---|---|---|
| B1 | `err == nil` calls `t.Fatal`; normal intended path keeps a nonnil error. | The test itself establishes the nonnil-error scenario. |
| B2 | `!errors.Is(err, ErrIncomplete)` records `t.Errorf`. | Establishes wrapped sentinel compatibility. |
| B3 | `!errors.As(err, &incomplete)` calls `t.Fatalf`, ending the test. | Establishes typed `*IncompleteError`. |
| B4 | Missing established refresh text records `t.Errorf`. | Establishes diagnostic text. |
| B5 | Empty/misordered `ReasonCodes` records `t.Errorf`. | Establishes a nonempty first `ReasonStreak` code. |

## Calls and live bindings

Calls construct fixture/criteria/time, invoke `BuildAttestation`, and inspect `errors.Is/As`, text, and codes. This test does not execute a live request or write an attestation.

## State mutations and fallbacks

Mutations are test-local variables only; assertion failure is the test fallback.

## Safety conclusion

The in-process fixture has no profile path, broker, timeout, retry, or runtime-gate binding.
