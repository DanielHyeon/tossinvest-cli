# Risk Pattern Report: `ExitObserver.record`

Source: `internal/app/engine/exitloop.go`.

Conditional protection cancellation is observation/reconciliation work. Folding it into the existing clear failure branches would delay or prevent a stop-loss submission.
