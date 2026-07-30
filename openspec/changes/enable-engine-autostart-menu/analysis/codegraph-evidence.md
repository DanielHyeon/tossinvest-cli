# CodeGraph evidence

Captured before production edits on 2026-07-31.

## Definition and impact

| Symbol | Definition | Impact evidence | Edit boundary |
|---|---|---|---|
| `Console.handleSettingsGate` | `internal/console/settings_operating.go:210` | route `POST /settings/gate` through `Console.routes` | left unchanged; new autostart handler is separate |
| `runConsole` | `cmd/tossctl/console.go:191` | called by `newConsoleCmd`; owns all concrete console seams | inject autostart seam and run the one-shot startup helper |
| `startEngine` | `cmd/tossctl/engineproc.go:116` | six focused process tests plus `runConsole` wiring | reuse unchanged as the only engine-start implementation |
| `SaveEngineGateEnabled` | `internal/config/operating_io.go:125` | `consoleGateSwitch.Save` and three surgical-write tests | use its one-key splice pattern; do not widen it |

## Structural conclusions

- The engine-start capability already crosses into `internal/console` only as
  `StartEngine func() (string, error)`.
- `Console.routes` is the single route table and every state-changing route is
  wrapped by both `session0` and `mutating` (CSRF).
- Settings readers are independent seams; a new autostart reader does not need
  the automation gate or Guardian values.
- `startEngine` already checks the advisory marker/process list and delegates the
  final decision to `tossctl engine run`, whose journal flock and startup
  interlock remain authoritative.
- The safe implementation is therefore a new one-key seam plus calls to the
  existing start seam, not a second process runner or a second interlock.
