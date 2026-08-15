# Risk Pattern Report: `Runtime.runAuxiliary`

Source: `internal/app/engine/auxiliary.go`.

Auxiliaries are intentionally non-supervised: their failure does not stop the engine, but an `OnStop` callback can mutate a latch. Using this stop callback for protection-worker authority creates a circular availability dependency. A100 must retain current alert-delivery semantics and prove worker recovery is not gated by its own failure notification.
