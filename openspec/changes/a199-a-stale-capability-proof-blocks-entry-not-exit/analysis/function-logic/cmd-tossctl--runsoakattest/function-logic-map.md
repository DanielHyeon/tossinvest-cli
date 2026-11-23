# Function Logic Map: `runSoakAttest`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `cmd/tossctl/soak.go` (476-581)
- AST evidence: `ast.json` — branches **20** / returns **10** / calls **49**, source SHA-256 `3c1263d5d690d192265d0b255e91840620dae6c25e1d1b7fa8c3080b6e50162c`(파일 단위).
  명령 `go run ./tools/logic-map --file cmd/tossctl/soak.go --func runSoakAttest`.
- Risk scan: `risk-pattern-report.md`.
- 의도: `tossctl soak attest` — attestation 발급 · 저장 · 범위 상실 안내 출력.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 발급 결과 · 빠진 endpoint | soak.BuildAttestation | 출력 문구 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@479) | `if opts.recordRenewalStatus {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@480) | `if strings.TrimSpace(opts.record) != "" \|\| strings.TrimSpace(opts.out) != "" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@484) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@492) | `if result == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@495) | `if a, err := attest.Load(attestationPath); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (else@497) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@501) | `if err := saveRenewalStatus(statusPath, status); err != nil && result == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@507) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@510) | `if opts.validity > 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@515) | `if base := soakSurveyedBase(cmd.Context(), root); base != "" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@521) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (if@526) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (if@532) | `if !opts.recordRenewalStatus {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B14 (if@535) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B15 (if@539) | `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (if@542) | `if err := attest.Save(path, attestation); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (range@554) | `for _, p := range attestation.SupervisedBy {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B18 (if@556) | `if market == "" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B19 (if@564) | `if len(missing) > 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B20 (range@567) | `for _, e := range missing {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 49개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- attestation 파일 저장 · 출력.

## Safety conclusion

- a199 편집(예정): 결정 11: 565 「The automation gate will refuse to start」 · 577 「clause 9 … still refuse」 문구를 새 귀결로(문구만).
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
