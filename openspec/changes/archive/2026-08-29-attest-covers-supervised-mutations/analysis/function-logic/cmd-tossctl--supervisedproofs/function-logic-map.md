# Function Logic Map: `supervisedProofs`

- Source: `cmd/tossctl/soak.go`
- Function: `cmd/tossctl/soak.go:supervisedProofs`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`

이 change가 추가한 함수. 시장별 검증 기록을 `verify`와 같은 해석 경로로 찾아 읽고 mutation 증거를 모은다. 없는 시장은 오류가 아니다 — 그것이 인터록이 보고할 '미증명'이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opts.verifyRecords | 비어 있으면 표준 경로 | --verify-record | 지정되면 그것만 읽는다 |
| 기록 파일 | KR·US 각각 | resolveVerifyRecordFor | 없으면 건너뜀, 다른 오류는 전파 |
| evidence.AccountRefs | 정확히 1개 | SucceededEndpoints | 2개 이상이면 모호하므로 거부 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 405) — `if len(opts.verifyRecords) > 0 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B2 | `else` (line 411) — `} else {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B3 | `range` (line 406) — `for _, p := range opts.verifyRecords {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B4 | `if` (line 407) — `if trimmed := strings.TrimSpace(p); trimmed != "" {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B5 | `range` (line 412) — `for _, market := range []string{verifylive.MarketKR, verifylive.MarketUS} {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B6 | `if` (line 414) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B7 | `range` (line 422) — `for _, s := range sources {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B8 | `if` (line 424) — `if err != nil {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B9 | `if` (line 425) — `if errors.Is(err, os.ErrNotExist) {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B10 | `if` (line 433) — `if len(evidence.AccountRefs) > 1 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B11 | `if` (line 440) — `if len(evidence.AccountRefs) == 1 {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B12 | `range` (line 443) — `for endpoint, at := range evidence.Endpoints {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B13 | `range` (line 457) — `for _, e := range soak.LiveOnlyEndpoints() {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B14 | `range` (line 461) — `for _, p := range proofs {` | 없음 | 루프/분기 계속 | 아래 참조 |
| B15 | `if` (line 462) — `if allowed[strings.ToUpper(p.Endpoint)] {` | 없음 | 루프/분기 계속 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `len` | ast.json calls (line 405) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.TrimSpace` | ast.json calls (line 407) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `append` | ast.json calls (line 408) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `resolveVerifyRecordFor` | ast.json calls (line 413) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `fmt.Errorf` | ast.json calls (line 415) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `verifylive.LoadEntries` | ast.json calls (line 423) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `errors.Is` | ast.json calls (line 425) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `verifylive.SucceededEndpoints` | ast.json calls (line 432) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.Join` | ast.json calls (line 437) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `filepath.Base` | ast.json calls (line 448) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `soak.LiveOnlyEndpoints` | ast.json calls (line 457) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `strings.ToUpper` | ast.json calls (line 458) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |
| `sort.Slice` | ast.json calls (line 466) | 순수 호출 또는 테스트 헬퍼 — error 계약은 호출부에서 처리 | ast.json |

라이브 바인딩 없음 — 브로커·계좌·네트워크를 직접 호출하지 않는다.

## State mutations and fallbacks

- 지역 slice만 쓴다. 기록 파일을 읽기만 하고 쓰지 않는다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 브로커 호출 0건.
- High-risk impact: yes — 게이트 증거를 나른다. 정책 거부는 BuildAttestation에 있다.
