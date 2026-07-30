package journal

// schemaV9 adds immutable policy identity to the two records that can establish
// exit-management t0. Both columns are nullable and historical rows are not
// rewritten.
const schemaV9 = `
ALTER TABLE exit_states ADD COLUMN policy_id TEXT;
ALTER TABLE position_adoptions ADD COLUMN exit_policy_id TEXT;
`
