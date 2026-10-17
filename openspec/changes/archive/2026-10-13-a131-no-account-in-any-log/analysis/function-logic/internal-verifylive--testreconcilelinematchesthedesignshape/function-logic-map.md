# Function Logic Map: `TestReconcileLineMatchesTheDesignShape`

- Source: `internal/verifylive/reconcile_g2_test.go` (215-301)
- Qualified: `TestReconcileLineMatchesTheDesignShape`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 69d35b7c393e8b03…) — AST branches 24 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 시험 함수(a131 이 기대값만 바꿈). 대사 줄의 계좌(마스크)가 평문이 아니라 봉인되어 있고 열면 기록의 마스크임을 단언 · 손으로 만든 옛 모양 줄의 `account_ref` 는 실제 줄의 `account_ref_sealed` 에 대응. 편집 뒤 작성 — 생산 코드가 아니라 기대값의 정정이며 시험의 분기·흐름은 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 하네스 | — | 같은 파일의 픽스처 | 실패는 t.Fatal |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 219:2 | `if v, _ := line["format_version"].(float64); int(v) != RecordFormatVersion {` | 무편집 |
| B2 | if at 222:2 | `if line["step_id"] != string(StepReconcile) {` | 무편집 |
| B3 | if at 226:2 | `if _, plain := line["account_ref"]; plain {` | 무편집 |
| B4 | if at 230:2 | `if got, ok := redact.OpenAccount(sealed, filepath.Dir(h.path)); !ok \|\| got != rcMask() {` | 무편집 |
| B5 | if at 233:2 | `if v, ok := line["verdict"]; ok && v != "" {` | 무편집 |
| B6 | if at 236:2 | `if line["mutating"] != false {` | 무편집 |
| B7 | if at 239:2 | `if _, ok := line["calls"]; ok {` | 무편집 |
| B8 | if at 243:2 | `if len(obs) != 1 {` | 무편집 |
| B9 | if at 247:2 | `if o["key"] != ObservationReconcileBasis \|\| !strings.HasPrefix(o["value"].(string), ReconcileBasisDomain) {` | 무편집 |
| B10 | if at 253:2 | `if o["value"] != ReconcileBasisDigest(nil) {` | 무편집 |
| B11 | if at 257:2 | `if len(arts) != 1 {` | 무편집 |
| B12 | range at 262:2 | `for field, value := range map[string]any{"kind": want.Kind, "id": want.ID, "symbol": want.Symbol, "chain_id": ` | 무편집 |
| B13 | if at 263:3 | `if a[field] != value {` | 무편집 |
| B14 | if at 267:2 | `if a["reconciled_absent"] != true {` | 무편집 |
| B15 | if at 270:2 | `if c, _ := time.Parse(time.RFC3339Nano, a["created_at"].(string)); !c.Equal(want.CreatedAt) {` | 무편집 |
| B16 | if at 273:2 | `if r, _ := time.Parse(time.RFC3339Nano, a["reconciled_at"].(string)); !r.Equal(rcNow) {` | 무편집 |
| B17 | range at 276:2 | `for _, forbidden := range []string{"cancelled_at", "filled", "filled_at", "held_until"} {` | 무편집 |
| B18 | if at 277:3 | `if _, ok := a[forbidden]; ok {` | 무편집 |
| B19 | if at 281:2 | `if a["cancelled"] == true {` | 무편집 |
| B20 | if at 286:2 | `if err := json.Unmarshal([]byte(rcReconcileLineJSON(rcTarget(), rcNow, "x")), &hand); err != nil {` | 무편집 |
| B21 | range at 289:2 | `for k := range hand {` | 무편집 |
| B22 | if at 290:3 | `if k == "run_id" \|\| k == "process" \|\| k == "title" \|\| k == "started_at" \|\| k == "finished_at" {` | 무편집 |
| B23 | if at 293:3 | `if k == "account_ref" {` | 무편집 |
| B24 | if at 296:3 | `if _, ok := line[k]; !ok && k != "verdict" {` | 무편집 |

반환문 없음(AST returns 0).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 하네스·대상 함수 | 시험 | — | AST calls |
| `redact.String`/`redact.OpenAccount`(일부) | 치환·봉인 기대값 계산 | 순수 | a131 |

## State mutations and fallbacks

- 시험 상태만.

## Safety conclusion

- Safe edit boundary: 기대값 식만 바꿈 — 분기·흐름 불변.
- High-risk impact: no — 시험 기대값.
