# Function Logic Map: `runEngineRun`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `cmd/tossctl/engine.go` (196-384)
- AST evidence: `ast.json` — branches **26** / returns **10** / calls **69**, source SHA-256 `27513e682d3bee3ea304ba884e8d66345577b51c3ff7160424742f27fd0cd7cd`(파일 단위).
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
| B7 (if@232) | `if !ectx.Automation.Verified {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@242) | `if lockPath, verr := engineVerifyLockPath(root); verr == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@243) | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.StaleAfter); fresh {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@259) | `if token, terr := engineProcInstance(os.Getpid()); terr == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@262) | `if merr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (else@266) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (if@278) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B14 (if@282) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B15 (if@305) | `if policyControl != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (if@308) | `if policyControlErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (if@313) | `if policyRuntime != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B18 (if@316) | `if policyRuntimeErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B19 (if@331) | `if strategyRuntime != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B20 (if@341) | `if projErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B21 (if@348) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B22 (if@352) | `if alertControl != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B23 (if@355) | `if alertControlErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B24 (if@364) | `if modeControlErr == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B25 (if@367) | `if modeControl != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B26 (if@371) | `if modeControlErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 69개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 잠금 · 마커 · endpoint 기동 · 출력.

## Safety conclusion

- a199 편집(예정): B7@232 를 청산 쪽 표지로 + 청산 전용 기동 출력 한 줄(D8). 새 defer 없음.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
