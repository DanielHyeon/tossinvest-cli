# Function Logic Map: `runEngineRun`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/cmd-tossctl--runenginerun.ast.json`.

- Source: `cmd/tossctl/engine.go` (196-388)
- AST evidence: `ast.json` — branches **27** / returns **10** / calls **72**, source SHA-256 `cf8e80ad009ec7b312cbab6306980f0fd2c8ffbb5043fe5f63143292376d6d92`(파일 단위).
  명령 `go run ./tools/logic-map --file cmd/tossctl/engine.go --func runEngineRun`.
- Risk scan: `risk-pattern-report.md`.
- 의도: `tossctl engine run` — 잠금 → 조립/인터록 → 게이트 OFF 거절 → verify 잠금 → 마커 → 루프 → 실행.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx.Automation.Verified` | NewContext | B7@232 | 아래 분기 표 |
| 불변식 | defer 등록 순서(a109 정적 핀) | a109_the_lock_outlives_every_endpoint_close_test.go | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@198) | `if ctx == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@204) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@210) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@220) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@221) | `if clauses := engine.UnmetInterlockClauses(err); clauses != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (range@223) | `for _, clause := range clauses {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@233) | `if !ectx.Automation.RunsExitSide() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@236) | `if ectx.Automation.ExitOnly {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@246) | `if lockPath, verr := engineVerifyLockPath(root); verr == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@247) | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.StaleAfter); fresh {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@263) | `if token, terr := engineProcInstance(os.Getpid()); terr == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (if@266) | `if merr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (else@270) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B14 (if@282) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B15 (if@286) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (if@309) | `if policyControl != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (if@312) | `if policyControlErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B18 (if@317) | `if policyRuntime != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B19 (if@320) | `if policyRuntimeErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B20 (if@335) | `if strategyRuntime != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B21 (if@345) | `if projErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B22 (if@352) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B23 (if@356) | `if alertControl != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B24 (if@359) | `if alertControlErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B25 (if@368) | `if modeControlErr == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B26 (if@371) | `if modeControl != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B27 (if@375) | `if modeControlErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 72개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 잠금 · 마커 · endpoint 기동 · 출력.

## Safety conclusion

- a199 편집(적용됨): B7@232 를 청산 쪽 표지로 + 청산 전용 기동 출력 한 줄(D8). 새 defer 없음.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
