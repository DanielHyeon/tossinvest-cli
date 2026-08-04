package journal

import _ "embed"

// schemaV28 freezes every dispatch identity and authority field across the
// lifecycle transitions introduced by v25.
//
//go:embed strategy_dispatch_immutable_v28.sql
var schemaV28 string
