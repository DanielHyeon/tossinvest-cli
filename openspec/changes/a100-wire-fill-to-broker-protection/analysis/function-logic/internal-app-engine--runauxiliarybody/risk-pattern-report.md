# Risk Pattern Report: `runAuxiliaryBody`

Source: `internal/app/engine/auxiliary.go`.

Panic containment prevents total process loss but also routes the resulting error into `runAuxiliary`'s optional callback. The causal chain must remain visible in A100 tests: panic → typed stop → notification, never panic → implicit authority mutation → startup deadlock.
