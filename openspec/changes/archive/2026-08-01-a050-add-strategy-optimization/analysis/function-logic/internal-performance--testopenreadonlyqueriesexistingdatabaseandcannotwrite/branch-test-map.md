# Branch Test Map: `TestOpenReadOnlyQueriesExistingDatabaseAndCannotWrite`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | writer fixture opens | this test | read-only seam absent at `948e721` | PASS |
| B2 | one measured fixture trade is collected | this test | read-only seam absent at `948e721` | PASS |
| B3 | writer closes/checkpoints | this test | read-only seam absent at `948e721` | PASS |
| B4 | baseline main-file identity is readable | this test | read-only seam absent at `948e721` | PASS |
| B5 | both WAL and SHM baseline paths are inspected | this test | read-only seam absent at `948e721` | PASS |
| B6 | no writer sidecar remains before immutable open | this test | read-only seam absent at `948e721` | PASS |
| B7 | public read-only capability opens | this test | read-only seam absent at `948e721` | PASS |
| B8 | internal immutable/query-only handle opens for authority assertions | this test | read-only seam absent at `948e721` | PASS |
| B9 | dashboard query returns exactly one complete row | this test | read-only seam absent at `948e721` | PASS |
| B10 | schema DDL is rejected | this test | read-only seam absent at `948e721` | PASS |
| B11 | `BEGIN IMMEDIATE` cannot acquire writer transaction | this test | read-only seam absent at `948e721` | PASS |
| B12 | `PRAGMA query_only` is exactly enabled | this test | read-only seam absent at `948e721` | PASS |
| B13 | forbidden table remains absent | this test | read-only seam absent at `948e721` | PASS |
| B14 | ephemeral test handle closes successfully | this test | read-only seam absent at `948e721` | PASS |
| B15 | all exported reader methods are enumerated | this test | read-only seam absent at `948e721` | PASS |
| B16 | all forbidden writer method names are inspected | this test | read-only seam absent at `948e721` | PASS |
| B17 | no forbidden writer method is exposed | this test | read-only seam absent at `948e721` | PASS |
| B18 | public reader lifecycle closes successfully | this test | read-only seam absent at `948e721` | PASS |
| B19 | final main-file identity is readable | this test | read-only seam absent at `948e721` | PASS |
| B20 | main DB size and mtime are unchanged | this test | read-only seam absent at `948e721` | PASS |
| B21 | both final sidecar paths are inspected | this test | read-only seam absent at `948e721` | PASS |
| B22 | immutable read created neither WAL nor SHM | this test | read-only seam absent at `948e721` | PASS |
