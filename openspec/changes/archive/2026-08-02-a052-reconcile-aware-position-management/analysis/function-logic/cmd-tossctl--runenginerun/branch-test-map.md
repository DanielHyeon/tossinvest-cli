# Branch Test Map: `runEngineRun`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 180 | `if ctx == nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `if` line 186 | `if err != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `if` line 192 | `if err != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `if` line 202 | `if err != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B5 | `if` line 203 | `if clauses := engine.UnmetInterlockClauses(err); clauses != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B6 | `range` line 205 | `for _, clause := range clauses {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B7 | `if` line 214 | `if !ectx.Automation.Verified {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B8 | `if` line 224 | `if lockPath, verr := engineVerifyLockPath(root); verr == nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B9 | `if` line 225 | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.StaleAfter); fresh {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B10 | `if` line 236 | `if merr != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B11 | `else` line 240 | `} else {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B12 | `if` line 247 | `if err != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B13 | `if` line 251 | `if err != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B14 | `if` line 255 | `if err != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B15 | `if` line 260 | `if err != nil {` true/entered and complementary path | go test ./cmd/tossctl | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
