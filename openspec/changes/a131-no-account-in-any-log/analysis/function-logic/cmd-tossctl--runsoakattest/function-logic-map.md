# Function Logic Map: `runSoakAttest`

- Source: `cmd/tossctl/soak.go` (476-582)
- Qualified: `runSoakAttest`
- AST evidence: `ast.json` (편집 전(base), `source_sha256` 5602b691caadfdf4…) — AST branches 20 · returns 10
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** `soak attest` — 요약·감독 증거로 attestation 을 만들고 저장, 결과를 출력. 출력은 tossos-attest 타이머의 로그(attest-autorun.log · journald)로 간다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts` | 플래그 | cobra | 충돌 플래그는 거절 |
| 기록 파일들 | — | 디스크 | 읽기 오류 반환 |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 479:2 | `if opts.recordRenewalStatus {` | 무편집 |
| B2 | if at 480:3 | `if strings.TrimSpace(opts.record) != "" \|\| strings.TrimSpace(opts.out) != "" {` | 무편집 |
| B3 | if at 484:3 | `if err != nil {` | 무편집 |
| B4 | if at 492:4 | `if result == nil {` | 무편집 |
| B5 | if at 495:5 | `if a, err := attest.Load(attestationPath); err != nil {` | 무편집 |
| B6 | else at 497:12 | `} else {` | 무편집 |
| B7 | if at 501:4 | `if err := saveRenewalStatus(statusPath, status); err != nil && result == nil {` | 무편집 |
| B8 | if at 507:2 | `if err != nil {` | 무편집 |
| B9 | if at 510:2 | `if opts.validity > 0 {` | 무편집 |
| B10 | if at 515:2 | `if base := soakSurveyedBase(root); base != "" {` | 무편집 |
| B11 | if at 521:2 | `if err != nil {` | 무편집 |
| B12 | if at 526:2 | `if err != nil {` | 무편집 |
| B13 | if at 532:2 | `if !opts.recordRenewalStatus {` | 무편집 |
| B14 | if at 535:3 | `if err != nil {` | 무편집 |
| B15 | if at 539:2 | `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {` | 무편집 |
| B16 | if at 542:2 | `if err := attest.Save(path, attestation); err != nil {` | 무편집 |
| B17 | range at 555:2 | `for _, p := range attestation.SupervisedBy {` | 무편집 |
| B18 | if at 557:3 | `if market == "" {` | 무편집 |
| B19 | if at 565:2 | `if len(missing) > 0 {` | 무편집 |
| B20 | range at 568:3 | `for _, e := range missing {` | 무편집 |

Exact AST return positions: 481:4, 485:4, 508:3, 522:3, 528:3, 536:4, 540:3, 543:3, 571:3, 581:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loadSoakSummary`·`supervisedProofs`·`soak.BuildAttestation`·`attest.Save` | 발급 | 오류 반환 | AST calls |
| `fmt.Fprintf` | 결과 출력 | — | AST calls |

## State mutations and fallbacks

- attestation 파일·갱신 상태 파일 쓰기(무편집). a131: 출력의 `account` 줄 하나 삭제.

## Safety conclusion

- Safe edit boundary: 출력 한 줄 삭제 — 분기·반환·파일 쓰기 불변.
- High-risk impact: no(직접) — 출력 텍스트만.
