# Function Logic Map: `Tracker.Observe`

- Source: `internal/reconcile/mismatch.go`
- Frozen source span: lines 486–654 at base `3615f793c4a9dbe027fdbe88b3ed01e140b05cc9`
- AST evidence: `ast.json` (`source_sha256` binds this enumeration to the file)
- Risk scan: `risk-pattern-report.md` (no configured pattern matched)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff` | one authoritative comparison; `BlocksEntry` iff quantity or locally-owned missing-order disagreement exists | `reconcile.Diff` / `Comparer.Compare` | a blocking diff must latch before persistence; an unreadable comparison never reaches this function |
| `t.failures` | current scalar count of consecutive blocking comparisons | in-memory `Tracker` | current defect: different disputes share this scalar and can jointly reach permanent promotion |
| `t.blocks` | active and pending durable reconcile blocks | journal plus pending in-memory writes | journal failure keeps the conservative in-memory block and returns an error |
| `t.adjusted` | symbol → comparison as-of that earned a future release | `Converger` credit | missing, stale or unordered credit cannot release a block |
| `MaxFailures` | positive; zero maps to `DefaultMaxFailures=3` | tracker options | reaching the threshold creates an account-wide operator-only permanent block |

## Branches and early returns

| AST branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `t.blocks == nil` | initializes the block map | continues | ordinary first mismatch |
| B2 | `!diff.BlocksEntry()` | resets scalar `failures`; evaluates adjusted releases | continues to durable write | clean comparison reset/release tests |
| B3 | blocking diff | increments the single scalar `failures`; enters symbol blocks; may promote account-wide | continues to durable write | same-dispute and changing-dispute threshold tests |
| B4–B5 | iterate active blocks; cause/release mode is not the quantity auto-release contract | no release proposed | continue block loop | foreign-cause preservation |
| B6 | no strictly earlier adjustment credit | appends `AwaitingAdjustment` | continue block loop | coincidental agreement does not release |
| B7 | the later comparison still disputes the symbol | appends `AwaitingAdjustment` | continue block loop | reclassified disagreement |
| B8–B9 | iterate `blocksFor(diff)`; exact block key is absent | mark pending, add to memory and `Outcome.Added` | continues | new mismatch latches before I/O |
| B10 | scalar `failures >= maxFailures && !t.permanent` | constructs account-wide permanent proposal | continues | threshold promotion |
| B11 | permanent key not already present | marks permanent pending, sets `t.permanent`, adds outcome | continues | foreign account cause is not overwritten |
| B12 | sort outcome slices | deterministic output only | continues | deterministic assertions |
| B13–B14 | build current-added set; an older pending block was not re-added | retries pending durable enter | continues | failed enter retry |
| B15 | persisted durable additions | clears pending and replaces memory row | continues | write-through evidence |
| B16 | journal reports a different authoritative cause | replaces proposal with durable authority | continues with returned persistence error | cause-conflict test |
| B17 | exact releases committed | deletes only committed block keys | continues | partial durable release |
| B18 | build committed-symbol set | records which credits were spent by a committed release | continues | credit accounting |
| B19–B20 | iterate credits; comparison is not strictly later | preserves credit | continue credit loop | same/undated/earlier comparison tests |
| B21 | later comparison disputes symbol, release committed, or no block remains | deletes that symbol credit | continues | refuted/spent/orphan credit tests |
| Return | all paths | fills outcome, re-syncs gate, unlocks | returns outcome plus persistence error | all focused tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Diff.BlocksEntry` | separates clean from entry-blocking comparison | pure; quantity and missing local orders only | CodeGraph callee + `compare.go` |
| `blocksFor` | projects current blocking items to symbol-scoped blocks | pure; quantity and missing-order branches | AST B8 + source |
| `Tracker.persist` | makes additions/releases durable before final visibility | returns error; additions stay conservatively pending, releases publish only when committed | AST call at base line 594 |
| `Tracker.syncGate` | mirrors the current active block set to entry control | called before and after persistence | AST calls at base lines 583/649 |
| `creditUsableBy` | proves a re-read is strictly later than an adjustment | false on absent/unparseable/unordered times | AST B6/B20 |

## State mutations and fallbacks

- Current promotion state is `Tracker.failures int`, one account-level scalar. It has no dispute identity.
- The 2026-08-07 incident supplied different symbol disputes in consecutive blocking comparisons. B3 incremented the same scalar to 3, and B10/B11 created an account-wide permanent block even though the first two symbol blocks later committed `ADJUSTMENT_APPLIED` releases.
- Ordinary symbol blocks remain fail-closed and must not wait for permanent-streak classification.
- A persistence error remains conservative: no entry gate reopening and no silent release.
- The edit must split pending retry policy: ordinary symbol pending entries remain fail-closed and retry;
  a non-durable account permanent proposal retries only while its earning canonical dispute is present in
  the immediately next blocking comparison, and is withdrawn when that continuity breaks.
- Existing permanent blocks remain operator-only; this change alters promotion evidence, not release evidence.

## Safety conclusion

- Safe edit boundary: replace only the identity-free promotion streak accounting, its failed-permanent-write
  retry identity, and reset/restore bookkeeping; keep ordinary block entry/retry, durable ordering, credit
  accounting, exit allowance and operator release unchanged.
- High-risk impact: **yes**, reconciliation and automatic-adoption gating. RED-first incident fixture, mutation checks, race tests, full gate and independent adversarial review are mandatory.
