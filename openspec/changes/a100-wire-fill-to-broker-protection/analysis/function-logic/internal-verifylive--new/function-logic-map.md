# Function Logic Map: `New`

- Source: `internal/verifylive/runner.go:320-425`
- Qualified function: `New`
- Revision: `current`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Inputs and durable state | Exact typed/current values | `internal/verifylive/runner.go` plus A100 tasks 0.2a.1–0.2a.9 | Reject or terminal HOLD; never infer evidence |
| Receipt/official evidence | Same-client raw result and attempts, active exclusive lease | Sealed official source and causal receipt | Any read/decode/identity/write/sync gap remains HOLD |
| Mutation authority | Exact M0 prerequisites and existing six methods | CLI/New gates and `MutationMethods()` | No factory/mutation outside the authorized trigger-only mode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/runner.go:321` — `if o.Broker == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B2 | `if` at `internal/verifylive/runner.go:324` — `if o.Recorder == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B3 | `if` at `internal/verifylive/runner.go:327` — `if o.Confirm == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B4 | `if` at `internal/verifylive/runner.go:330` — `if o.ConfirmBatch == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B5 | `if` at `internal/verifylive/runner.go:334` — `if strings.TrimSpace(o.AccountRef) == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B6 | `if` at `internal/verifylive/runner.go:338` — `if err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B7 | `if` at `internal/verifylive/runner.go:341` — `if o.IncludeTrigger {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B8 | `if` at `internal/verifylive/runner.go:342` — `if !o.ConfirmEach \|\| !o.Resume \|\| o.IncludeTTLEdge \|\| o.Receipt == nil \|\| !o.Receipt.usable() \|\|` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B9 | `if` at `internal/verifylive/runner.go:346` — `if len(Outstanding(o.Prior)) != 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B10 | `if` at `internal/verifylive/runner.go:349` — `if checkpoint, ok, checkpointErr := M0Unsettled(o.Prior); checkpointErr != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B11 | `else` at `internal/verifylive/runner.go:351` — `} else if ok && checkpoint.Kind != "pending-create" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B12 | `if` at `internal/verifylive/runner.go:351` — `} else if ok && checkpoint.Kind != "pending-create" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B13 | `else` at `internal/verifylive/runner.go:353` — `} else if !ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B14 | `if` at `internal/verifylive/runner.go:353` — `} else if !ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B15 | `if` at `internal/verifylive/runner.go:358` — `if err := M0ExactPrerequisites(o.Prior); err != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B16 | `if` at `internal/verifylive/runner.go:362` — `if _, ok := official.M0ReadSourceFor(o.Broker); !ok {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B17 | `if` at `internal/verifylive/runner.go:392` — `if r.approvalChannel == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B18 | `if` at `internal/verifylive/runner.go:395` — `if r.out == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B19 | `if` at `internal/verifylive/runner.go:398` — `if r.now == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B20 | `if` at `internal/verifylive/runner.go:401` — `if r.sleep == nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B21 | `if` at `internal/verifylive/runner.go:404` — `if r.maxSellQuantity <= 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B22 | `if` at `internal/verifylive/runner.go:407` — `if r.ttlWait <= 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B23 | `if` at `internal/verifylive/runner.go:410` — `if r.triggerWindow <= 0 {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B24 | `if` at `internal/verifylive/runner.go:413` — `if r.process.InstanceID == "" {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B25 | `range` at `internal/verifylive/runner.go:416` — `for _, id := range o.Redo {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |
| B26 | `if` at `internal/verifylive/runner.go:420` — `if r.m0Receipt != nil {` | Preserve source ordering; missing causal authority must HOLD | `TestM0ForgedBrokerCannotMintOfficialTransportEvidence` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `strings.TrimSpace` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `ValidateOffset` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `o.Receipt.usable` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `len` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `Outstanding` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `M0Unsettled` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `fmt.Errorf` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `M0ExactPrerequisites` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `official.M0ReadSourceFor` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `NormalizeMarket` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `make` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `append` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `call` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `UTC` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `time.Now` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `NewProcess` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.now` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `newToken` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |
| `r.m0Receipt.RunID` | Source dependency | Propagate errors; M0 critical gaps are irreversible | AST |

## State mutations and fallbacks

- Receipt/checkpoint persistence precedes the broker action or read it authorizes.
- Pending recovery is read-only; parent/child unresolved states are manual-only and never cleanup targets.
- Retry success cannot erase an earlier critical attempt failure.

## Safety conclusion

- Safe edit boundary: exact same-client authority, exclusive receipt lease, causal fsync order, terminal HOLD, and six-method mutation surface.
- High-risk impact: yes; every AST branch is linked to the named M0 or preservation test.
