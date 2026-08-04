package journal

import _ "embed"

// schemaV25 is additive and leaves released v24 bytes and rows untouched.
//
//go:embed strategy_dispatch_runtime_v25.sql
var schemaV25 string
