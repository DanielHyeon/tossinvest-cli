package journal

import _ "embed"

// schemaV26 is an additive authority companion. Released v20-v25 tables,
// triggers and rows remain byte-for-byte untouched.
//
//go:embed strategy_first_leg_v26.sql
var schemaV26 string
