-- a084: stamp each quarantine with the recovery-selector revision that produced it.
--
-- Additive and nullable, and deliberately not backfilled (§0.6). NULL means "written
-- before this column existed, so which selector decided it is unknown" — which is a
-- different fact from "the current selector decided it", and the difference is the
-- whole point. A backfill would tell the engine that the three live ambiguous_recovery
-- rows were judged by the comparator that a078 fixed, which is exactly false: they were
-- judged by the one it replaced, and they are the rows this change exists to re-judge.
ALTER TABLE exit_snapshot_quarantines ADD COLUMN selector_revision INTEGER;
