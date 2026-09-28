# Function Logic Map: `NewPositionPolicyCommandService`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json` (post-edit, 90–104, 2 branches). Pre-edit map (HEAD `09465697`, 86–99): `analysis/pre-edit/5.5-socket/internal-app-engine--newpositionpolicycommandservice/`.
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| ectx | the assembled engine context with its owned journal | `engine.Open` | nil or no journal → error |
| ectx.Audit | the engine's operational audit log (`openAuditLog`; `engine.Open` fails if it cannot open) | `engine.Open` | stored as is; nil → the release methods refuse (`ErrAuditUnavailable`) |
| clk | any clock | caller | nil → system clock |

## Branches and early returns (post-edit, from `ast.json`)

| Branch | Line | Condition | Effect |
|---|---|---|---|
| B1 | 91 | `ectx == nil \|\| ectx.Journal == nil` | error |
| B2 | 94 | `clk == nil` | system clock |

The edit adds one field (`audit: ectx.Audit`) to the returned literal. No branch changed.

## Calls and live bindings

- None beyond `clock.System()` (B2). The audit field is read by `relaxationAuditor` in `risk_relaxation_command.go`.

## State mutations and fallbacks

- Pure constructor. A nil audit log is **not** defaulted: `relaxationAuditor` refuses it, because `(*audit.Log)(nil).RecordAction` returns nil and would make an unaudited release commit.

## Safety conclusion

- Wiring only. The policy and quarantine paths do not read the new field.
