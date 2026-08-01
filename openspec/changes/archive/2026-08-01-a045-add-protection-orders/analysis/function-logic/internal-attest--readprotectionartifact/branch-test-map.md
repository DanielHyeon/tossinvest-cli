# Branch Test Map: `readProtectionArtifact`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | pre-read invalid parent/file metadata | existing filesystem table | pass | pass regression |
| B2 | file changes during read | post-read replacement | pass | pass regression |
| B3 | parent changes/mode changes during read | new parent TOCTOU table | absent | pass after remediation |
| B4 | platform lacks owner/link metadata | Windows compile plus fail-closed platform helper | `os.Geteuid` not portable | pass after remediation |
| B5 | pre-read file mode differs from exact policy | mode table | partial metadata risk | rejected |
| B6 | pre-read file owner/type/link check fails | filesystem table | pass | pass regression |
| B7 | no-follow-safe open fails | symlink/replacement table | pass | rejected |
| B8 | opened descriptor cannot be statted | filesystem error path | pass | rejected |
| B9 | opened descriptor differs from lstat snapshot/mode | replacement race test | pass | rejected |
| B10 | opened descriptor metadata check fails | owner/link/type table | pass | rejected |
| B11 | bounded read fails | read error path | pass | rejected |
| B12 | artifact exceeds maximum bytes | oversize test | pass | rejected |
| B13 | deterministic post-read race hook exists | parent/file replacement tests | parent race untested | hook invoked only in same-package tests |
| B14 | opened descriptor restat fails | post-read error path | incomplete mediation | rejected |
| B15 | path disappears or file snapshot changes | replacement/mode tests | pass | rejected |
| B16 | parent disappears or snapshot changes | parent replacement/mode tests | parent only checked before read | rejected |
| B17 | opened/path mode changes during read | post-read mode test | partial snapshot risk | rejected |
| B18 | opened descriptor owner/type/link changes | post-read metadata test | incomplete mediation | rejected |
| B19 | path owner/type/link changes | post-read metadata test | incomplete mediation | rejected |
| B20 | parent permissions/owner/type fail after read | parent post-read table | absent | rejected |
