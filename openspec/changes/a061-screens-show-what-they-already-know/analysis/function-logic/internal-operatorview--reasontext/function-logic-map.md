# Function Logic Map: `reasonText`

- Source: `internal/operatorview/exit_line.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a061-screens-show-what-they-already-know/base-commit.txt`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reason string` | a typed reason produced by the journal or by a caller's freshness policy | `journal`, `console` | unrecognised -> the default arm's generic sentence, never a raw identifier |

**Invariant**: this is a total function from reason code to one Korean sentence.
It has no default that leaks the code itself to the screen, and it never decides
whether a value is shown -- only how the refusal is worded.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the switch over the trimmed reason | returns one operator sentence | falls out of the switch | every case below |
| B2 | `engine_not_running` (a061) | returns one operator sentence | falls out of the switch | `TestAStoppedEngineClosesTheProtectionLine` |
| B3 | `snapshot_quarantined` (a061) | returns one operator sentence | falls out of the switch | `TestAQuarantinedPositionIsNotDrawnAsProtected` |
| B4 | `observation_older_than_limit` | returns one operator sentence | falls out of the switch | `TestAConsoleWithNoEngineMarkerKeepsTheObservationAgeBound` |
| B5 | `observation_in_future` | returns one operator sentence | falls out of the switch | `TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted` |
| B6 | `invalid_observed_at` | returns one operator sentence | falls out of the switch | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B7 | `not_evaluated_yet` | returns one operator sentence | falls out of the switch | existing SEED fixture test |
| B8 | `no_saved_evaluation` | returns one operator sentence | falls out of the switch | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B9 | `legacy_snapshot_absent`, `legacy_event` | returns one operator sentence | falls out of the switch | `TestPositionsRenderCanonicalExitLineFixtures` |
| B10 | `invalid_stored_snapshot`, `invalid_effective_snapshot` | returns one operator sentence | falls out of the switch | `TestExitLinesStayClosedWhenTheEvidenceCannotBeTrusted` |
| B11 | `legacy_policy_identity_unknown`, `legacy_adoption_context_required` | returns one operator sentence | falls out of the switch | existing legacy tests |
| B12 | `partial_snapshot_tuple` and the other partial tuples | returns one operator sentence | falls out of the switch | existing corruption tests |
| B13 | `partial_policy_tuple`, `invalid_policy_identity` | returns one operator sentence | falls out of the switch | existing corruption tests |
| B14 | `flattened_snapshot_mismatch` | returns one operator sentence | falls out of the switch | existing corruption tests |
| B15 | `ambiguous_exit_evidence` | returns one operator sentence | falls out of the switch | existing orders evidence tests |
| B16 | `exit_evidence_unlinked` | returns one operator sentence | falls out of the switch | existing orders evidence tests |
| B17 | `lineage_cycle` | returns one operator sentence | falls out of the switch | existing lineage tests |
| B18 | `lineage_ambiguous` | returns one operator sentence | falls out of the switch | existing lineage tests |
| B19 | `lineage_depth_exceeded` | returns one operator sentence | falls out of the switch | existing lineage tests |
| B20 | `lineage_scope_mismatch` | returns one operator sentence | falls out of the switch | existing lineage tests |
| B21 | `invalid_event_evidence` | returns one operator sentence | falls out of the switch | existing orders evidence tests |
| B22 | the empty reason | returns one operator sentence | falls out of the switch | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |
| B23 | the default arm | returns one operator sentence | falls out of the switch | `TestReasonTextMapsKnownCorruptionAndNeverLeaksRawCodes` |

**a061 adds B2 and B3** at the top of the switch and changes nothing else. Cases
are mutually exclusive string matches, so prepending two new ones cannot alter
the answer for any existing reason.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | tolerate padded reason codes | pure | AST |

No I/O, no clock, no package state.

## State mutations and fallbacks

- Returns a constant string. Mutates nothing.
- The default arm is the fallback and is unchanged, so an unknown reason still
  renders the safe generic sentence.

## Safety conclusion

- Safe edit boundary: two new switch cases with new literals.
- High-risk impact: no. Wording only; the decision to hide a value is made
  before this function is called.
- Safety invariant 0.3 untouched.
