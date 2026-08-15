# Risk Pattern Report: `buildGateway`

Source: `internal/app/engine/gateway.go`.

Gateway construction restores authority before creating execution objects. Starting an A100 worker here would bypass the existing runtime recovery/supervision boundary.
