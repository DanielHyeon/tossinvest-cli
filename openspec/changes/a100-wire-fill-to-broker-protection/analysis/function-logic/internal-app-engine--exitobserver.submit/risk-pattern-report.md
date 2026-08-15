# Risk Pattern Report: `ExitObserver.submit`

Source: `internal/app/engine/exitloop.go`.

Protective reduction submission intentionally does not wait for broker protection convergence. New waiting/gating here would violate immediate-stop behavior.
