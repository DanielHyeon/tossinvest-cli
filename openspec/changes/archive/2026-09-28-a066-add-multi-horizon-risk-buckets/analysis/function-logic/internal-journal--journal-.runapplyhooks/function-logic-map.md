# Function Logic Map: `Journal.runApplyHooks`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | live `RecordFill` BEGIN IMMEDIATE transaction | caller in `RecordFill` | hook error rolls the entire fill transaction back |
| `fill` | authoritative, non-refused changed fill observation | `RecordFill` | never reinterprets broker state |
| `applyHooks` | zero or one immutable Project/Campaign/Exit binding | `SetApplyHooks` under `applyMu` | absent hooks are a no-op |
| risk owner bind | only after Project and Campaign have persisted exact Position/campaign generation | journal rows in the same `tx` | semantic absence/mismatch latches entry without rejecting fill; storage error rolls back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | all hooks nil | none | nil | existing unbound-hook fill tests |
| B2/B3 | Project configured / Project error | Position projection may move | wrapped error; outer rollback | existing apply-hook atomic rollback tests |
| B4/B5 | Campaign configured / Campaign error | campaign projection may move | wrapped error; outer rollback | existing campaign apply tests |
| B6 (new) | post-Campaign owner binding returns an error | semantic authority mismatch latches in the callee; storage error has no safe fallback | wrapped storage error; outer rollback | `TestRiskBucketFillHookLatchesBindGapWithoutReturningError` plus existing hook rollback tests |
| B7/B8 | Exit configured / Exit error | exit projection may move | wrapped error; outer rollback | existing exit apply tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `hooks.Project` | persist authoritative Position before lineage | error aborts all fill writes | current AST and apply-hook contract |
| `hooks.Campaign` | bind prospective campaign to Project's actual generation | error aborts all fill writes | current AST and campaign tests |
| `applyRiskBucketOwnerBindingInTx` (new) | cross-bind a066 owner to exact campaign/Position generation | semantic risk error becomes full-scope latch; transport error aborts | `TestRunApplyHooksBindsRiskBucketOwnerAfterCampaignInSameTransaction`, `TestRiskBucketFillHookLatchesBindGapWithoutReturningError` |
| `hooks.Exit` | apply exit state after Position/campaign | error aborts all fill writes | current AST and apply-hook contract |

## State mutations and fallbacks

- Mutations are owned by injected hooks and share one transaction.
- The new risk binding runs after Campaign and before Exit, never before Position generation exists.
- No broker request, wait, toggle, stop or risk-reducing gate is introduced.

## Safety conclusion

- Safe edit boundary: add one local journal derivation call after the successful Campaign hook; preserve all existing ordering and errors.
- High-risk impact: yes — fill/Position atomicity. Semantic a066 evidence failure must be converted to a latch and must not roll back the authoritative fill/Position.
