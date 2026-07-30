# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- Qualified function: `runConsole`
- AST evidence: `ast.json` (`ef133dc61d797dff9fadf273cb2ac7bd66c9f6c0c404fcdfe9a93195838e60a6`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `cmd/tossctl/console.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `cmd/tossctl/console.go:168` — if ctx == nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B2 | `if` at `cmd/tossctl/console.go:177` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B3 | `if` at `cmd/tossctl/console.go:181` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B4 | `if` at `cmd/tossctl/console.go:185` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B5 | `if` at `cmd/tossctl/console.go:189` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B6 | `if` at `cmd/tossctl/console.go:194` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B7 | `if` at `cmd/tossctl/console.go:207` — if dir, derr := engineJournalDir(root); derr == nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B8 | `else` at `cmd/tossctl/console.go:210` — } else { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B9 | `if` at `cmd/tossctl/console.go:218` — if self, serr := binstamp.SelfPath(); serr != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B10 | `else` at `cmd/tossctl/console.go:220` — } else { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B11 | `if` at `cmd/tossctl/console.go:222` — if cerr != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B12 | `else` at `cmd/tossctl/console.go:230` — } else { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B13 | `if` at `cmd/tossctl/console.go:224` — if updater, uerr := localupdate.New(self); uerr != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B14 | `else` at `cmd/tossctl/console.go:226` — } else { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B15 | `if` at `cmd/tossctl/console.go:233` — if updater != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B16 | `if` at `cmd/tossctl/console.go:237` — if uerr != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B17 | `else` at `cmd/tossctl/console.go:239` — } else { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B18 | `if` at `cmd/tossctl/console.go:246` — if engineDir != "" { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |
| B19 | `if` at `cmd/tossctl/console.go:249` — if err != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.Context`, `context.Background`, `signal.NotifyContext`, `stop`, `resolveVerifyRecord`, `resolveVerifyRecordFor`, `resolveSoakRecord`, `resolveSoakAttestationPath`, `journal.DefaultPath`, `fmt.Fprintf`, `cmd.ErrOrStderr`, `engineJournalDir` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 27 assignment(s) and 11 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: no; current AST hash and affected-package tests are required.
