package journal

import _ "embed"

// schemaV27 adds durable weekly-value reservation uniqueness for KR and US.
//
//go:embed strategy_weekly_reservation_v27.sql
var schemaV27 string
