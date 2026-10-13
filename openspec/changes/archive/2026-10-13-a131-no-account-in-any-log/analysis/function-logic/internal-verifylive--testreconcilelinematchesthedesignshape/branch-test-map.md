# Branch Test Map: `TestReconcileLineMatchesTheDesignShape`

- Source: `internal/verifylive/reconcile_g2_test.go` (215-301) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 219:2 | `if v, _ := line["format_version"].(float64); int(v) != RecordFormatVersion {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | if at 222:2 | `if line["step_id"] != string(StepReconcile) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 226:2 | `if _, plain := line["account_ref"]; plain {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 230:2 | `if got, ok := redact.OpenAccount(sealed, filepath.Dir(h.path)); !ok \|\| got != rcMask() {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 233:2 | `if v, ok := line["verdict"]; ok && v != "" {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B6 | if at 236:2 | `if line["mutating"] != false {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B7 | if at 239:2 | `if _, ok := line["calls"]; ok {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B8 | if at 243:2 | `if len(obs) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B9 | if at 247:2 | `if o["key"] != ObservationReconcileBasis \|\| !strings.HasPrefix(o["value"].(string), ReconcileBasisDomain) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B10 | if at 253:2 | `if o["value"] != ReconcileBasisDigest(nil) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B11 | if at 257:2 | `if len(arts) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B12 | range at 262:2 | `for field, value := range map[string]any{"kind": want.Kind, "id": want.ID, "symbol": want.Symbol, "chain_id": ` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B13 | if at 263:3 | `if a[field] != value {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B14 | if at 267:2 | `if a["reconciled_absent"] != true {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B15 | if at 270:2 | `if c, _ := time.Parse(time.RFC3339Nano, a["created_at"].(string)); !c.Equal(want.CreatedAt) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B16 | if at 273:2 | `if r, _ := time.Parse(time.RFC3339Nano, a["reconciled_at"].(string)); !r.Equal(rcNow) {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B17 | range at 276:2 | `for _, forbidden := range []string{"cancelled_at", "filled", "filled_at", "held_until"} {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B18 | if at 277:3 | `if _, ok := a[forbidden]; ok {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B19 | if at 281:2 | `if a["cancelled"] == true {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B20 | if at 286:2 | `if err := json.Unmarshal([]byte(rcReconcileLineJSON(rcTarget(), rcNow, "x")), &hand); err != nil {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B21 | range at 289:2 | `for k := range hand {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B22 | if at 290:3 | `if k == "run_id" \|\| k == "process" \|\| k == "title" \|\| k == "started_at" \|\| k == "finished_at" {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B23 | if at 293:3 | `if k == "account_ref" {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B24 | if at 296:3 | `if _, ok := line[k]; !ok && k != "verdict" {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c
