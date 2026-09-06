# Function Logic Map: `Project`

- Source: `internal/strategyevidence/projection.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot` | a sealed snapshot with a market-matched scope and a non-zero evaluation time | `SealSnapshot` / `Replay` | a scope error before any requirement is read |
| `policy.Fatal` | the fatal facts this market declares, each with an authority priority | the caller's versioned policy | unavailability is recorded; only `Required` blocks |
| `policy.Required` / `policy.Optional` | lane scoring requirements | the caller's versioned policy | unavailability is recorded; only `Required` makes the lane ineligible |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the policy version, market or snapshot scope is invalid | none | a scope error | `TestProjectRefusesAnInvalidPolicyOrSnapshotScope` |
| B2 | walk the declared fatal facts | fills the fatal assessment | falls through | `TestProjectKeepsFatalAndLaneEvidenceSeparate` |
| B3 | a fatal fact cannot be chosen from this snapshot | records the refusal in `Fatal.Unavailable` | continues to the next fact | `TestProjectRecordsMissingFatalEvidenceEvenWhenItIsOptional` |
| B4 | that unavailable fatal fact was declared required | blocks and appends the reason | continues | `TestProjectRecordsMissingFatalEvidenceEvenWhenItIsOptional` |
| B5 | the fatal payload does not decode, or claims blocked with no code | blocks with `EVIDENCE_PAYLOAD_INVALID` | continues | `TestProjectRejectsFatalPayloadWithoutExplicitBlockedState` |
| B6 | the fatal payload says blocked | blocks with `FATAL_EVIDENCE` and the source's code | continues | `TestProjectKeepsFatalAndLaneEvidenceSeparate` |
| B7 | walk the required and optional lane requirements | fills the lane evidence set | falls through | `TestProjectRequiredMissingAndStaleFailClosedWhileOptionalIsTypedUnavailable` |
| B8 | a lane requirement cannot be chosen | records it in `Lane.Unavailable` | continues | `TestProjectRequiredMissingAndStaleFailClosedWhileOptionalIsTypedUnavailable` |
| B9 | that unavailable lane requirement was declared required | makes the lane ineligible and appends the refusal | continues | `TestProjectRequiredEvidenceConflictsAndScopeMismatchesFailClosed` |
| B10 | order the fatal reasons deterministically | sorts in place | falls through | `TestProjectOrdersFatalReasonsAndLaneRefusalsDeterministically` |
| B11 | order the lane refusals deterministically | sorts in place | the projection result | `TestProjectOrdersFatalReasonsAndLaneRefusalsDeterministically` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `choose` | pick the highest-priority authoritative envelope for one requirement | returns a typed refusal instead of a substitute | AST + `TestProjectRequiredEvidenceConflictsAndScopeMismatchesFailClosed` |
| `json.Unmarshal` | read the fatal payload's `blocked`/`code` | a decode failure blocks with an invalid-payload reason | AST + `TestProjectRejectsFatalPayloadWithoutExplicitBlockedState` |
| `sort.Slice` | make both reason lists deterministic | pure | AST + `TestProjectOrdersFatalReasonsAndLaneRefusalsDeterministically` |

## State mutations and fallbacks

- Builds a result value; the snapshot is never modified and no I/O occurs.
- No neutral score, no previous market day and no cross-market substitution is ever produced — every gap becomes a typed refusal.
- The completion pass added one recording, not one decision: `Fatal.Unavailable` is now filled for every fatal refusal. `Blocked` and `Reasons` still move only for `Required` facts, so a caller reading the old two fields sees exactly what it saw before.

## Safety conclusion

- High-risk: this is the function that decides whether a fatal veto exists, and a fatal fact that vanishes silently is indistinguishable from no fatal fact at all.
- Before B3 recorded it, a declared-but-optional fatal fact whose evidence was missing left no trace anywhere in the result — the lane loop had an `Unavailable` map, the fatal loop did not (issues.md I18).
- The addition is purely additive; `TestProjectLeavesFatalUnavailableEmptyWhenTheFactIsPresent` is the positive control that the map is not filled unconditionally.
- Two branches had never executed before this pass: B1's true arm (every existing test supplies a valid scope and fails the test on error) and the B10/B11 comparators (no test ever produced two reasons). Both are now driven directly, and flipping either comparator fails `TestProjectOrdersFatalReasonsAndLaneRefusalsDeterministically` — measured under `go test -overlay` on 2026-09-07.
