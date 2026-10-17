# Function Logic Map (pre-edit): `NewPositionPolicyCommandService`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json` (pre-edit, extracted 2026-09-29 at HEAD `09465697`, 86–99, 2 branches)

## Branches and early returns (pre-edit, from `ast.json`)

| Branch | Line | Condition | Effect |
|---|---|---|---|
| B1 | 87 | `ectx == nil || ectx.Journal == nil` | error: the service requires the owned journal |
| B2 | 90 | `clk == nil` | system clock |

## Planned edit (a066 5.5)

The composite literal gains one field, `audit: ectx.Audit` (the engine's operational audit log, opened by `engine.Open` → `openAuditLog`). No branch is added. A nil log is not defaulted here: the release methods refuse with `riskrelaxation.ErrAuditUnavailable` when the field is nil. That matters because `(*audit.Log)(nil).RecordAction` returns nil, so a nil log passed as an auditor would silently skip the audit line.
