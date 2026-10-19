# Function Logic Map: `TestRecordQFinalCampaignFirstLegRejectsStrategyflowTamperWithoutRows`

Source: `internal/journal/strategyflow_projection_test.go`; current and pre-edit AST retained.

## Inputs and invariants

Test-only mutations must alter the payload or lineage before verifying strict refusal. First-leg test checks no database rows remain after any refusal.

## Branches and early returns

Existing error and no-write assertions remain. New no-op checks fail tests early.

## Calls and live bindings

Calls production canonical verifier and local journal first-leg writer; no live operations.

## State mutations and fallbacks

Only local test requests and temporary journals are mutated.

## Safety conclusion

Corrects test fixture v4 schema mutation without changing the verifier; all tagged tests in source passed as `/tmp/tossos-a150-gate2-tamper-repair.log`. Aggregate gate repeat pending.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
